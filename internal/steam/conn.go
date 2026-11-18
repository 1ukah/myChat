package steam

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

type conn struct {
	ws     *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc

	incoming chan packet

	jobsMu  sync.Mutex
	jobs    map[uint64]chan packet
	nextJob atomic.Uint64

	steamID   atomic.Uint64
	sessionID atomic.Int32

	writeMu sync.Mutex
}

var errClosed = errors.New("steam: connection closed")

func dial(parent context.Context, endpoint string) (*conn, error) {
	dctx, dcancel := context.WithTimeout(parent, 10*time.Second)
	defer dcancel()
	ws, _, err := websocket.Dial(dctx, "wss://"+endpoint+"/cmsocket/", nil)
	if err != nil {
		return nil, err
	}
	ws.SetReadLimit(64 << 20)
	ctx, cancel := context.WithCancel(parent)
	c := &conn{
		ws:       ws,
		ctx:      ctx,
		cancel:   cancel,
		incoming: make(chan packet, 64),
		jobs:     map[uint64]chan packet{},
	}
	c.nextJob.Store(uint64(rand.Uint32()))
	go c.readLoop()
	return c, nil
}

func (c *conn) close() {
	c.cancel()
	_ = c.ws.Close(websocket.StatusNormalClosure, "")
}

func (c *conn) done() <-chan struct{} { return c.ctx.Done() }

func (c *conn) readLoop() {
	defer c.cancel()
	defer close(c.incoming)
	for {
		_, data, err := c.ws.Read(c.ctx)
		if err != nil {
			return
		}
		if err := c.dispatch(data); err != nil {
			return
		}
	}
}

func (c *conn) dispatch(data []byte) error {
	p, err := parsePacket(data)
	if err != nil {
		return nil
	}
	if p.emsg == emsgMulti {
		return c.handleMulti(p.body)
	}
	if p.hdr.jobTarget != jobNone {
		c.jobsMu.Lock()
		ch, ok := c.jobs[p.hdr.jobTarget]
		if ok {
			delete(c.jobs, p.hdr.jobTarget)
		}
		c.jobsMu.Unlock()
		if ok {
			ch <- p
			return nil
		}
	}
	select {
	case c.incoming <- p:
		return nil
	case <-c.ctx.Done():
		return errClosed
	}
}

func (c *conn) handleMulti(body []byte) error {
	var sizeUnzipped uint32
	var payload []byte
	if err := walk(body, func(f field) {
		switch f.Num {
		case 1:
			sizeUnzipped = f.u32()
		case 2:
			payload = f.Bytes
		}
	}); err != nil {
		return nil
	}
	if sizeUnzipped > 0 {
		zr, err := gzip.NewReader(bytes.NewReader(payload))
		if err != nil {
			return nil
		}
		buf := make([]byte, 0, sizeUnzipped)
		w := bytes.NewBuffer(buf)
		if _, err := io.Copy(w, io.LimitReader(zr, int64(sizeUnzipped)+1)); err != nil {
			return nil
		}
		payload = w.Bytes()
	}
	for len(payload) >= 4 {
		n := int(binary.LittleEndian.Uint32(payload))
		payload = payload[4:]
		if n > len(payload) {
			return nil
		}
		if err := c.dispatch(payload[:n]); err != nil {
			return err
		}
		payload = payload[n:]
	}
	return nil
}

func (c *conn) write(b []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	wctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()
	return c.ws.Write(wctx, websocket.MessageBinary, b)
}

func (c *conn) send(emsg uint32, h header, body []byte) error {
	if emsg != emsgClientHello && emsg != emsgServiceMethodCallFromClientNonAuthed {
		h.steamID = c.steamID.Load()
		h.sessionID = c.sessionID.Load()
	}
	return c.write(buildPacket(emsg, h, body))
}

func (c *conn) call(ctx context.Context, emsg uint32, method string, body []byte) (packet, error) {
	job := c.nextJob.Add(1) & 0x7fffffffffffffff
	ch := make(chan packet, 1)
	c.jobsMu.Lock()
	c.jobs[job] = ch
	c.jobsMu.Unlock()
	defer func() {
		c.jobsMu.Lock()
		delete(c.jobs, job)
		c.jobsMu.Unlock()
	}()

	h := newHeader()
	h.jobSource = job
	h.targetJobName = method
	if emsg == emsgServiceMethodCallFromClientNonAuthed {
		h.realm = 1
	}
	if err := c.send(emsg, h, body); err != nil {
		return packet{}, err
	}
	tctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	select {
	case p := <-ch:
		if p.hdr.eresult != EResultOK {
			return p, &Error{EResult: p.hdr.eresult, Message: p.hdr.errorMessage}
		}
		return p, nil
	case <-tctx.Done():
		return packet{}, fmt.Errorf("steam: %s timed out", method)
	case <-c.ctx.Done():
		return packet{}, errClosed
	}
}

func (c *conn) notify(method string, body []byte) error {
	h := newHeader()
	h.targetJobName = method
	return c.send(emsgServiceMethodCallFromClient, h, body)
}

type cmServer struct {
	Endpoint string `json:"endpoint"`
	Type     string `json:"type"`
}

var (
	cmCacheMu sync.Mutex
	cmCache   []string
	cmCacheAt time.Time
)

func cmEndpoints(ctx context.Context) ([]string, error) {
	cmCacheMu.Lock()
	defer cmCacheMu.Unlock()
	if len(cmCache) > 0 && time.Since(cmCacheAt) < 30*time.Minute {
		return cmCache, nil
	}
	rctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(rctx, http.MethodGet,
		"https://api.steampowered.com/ISteamDirectory/GetCMListForConnect/v1/?cellid=0&cmtype=websockets", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Response struct {
			ServerList []cmServer `json:"serverlist"`
		} `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	var list []string
	for _, s := range out.Response.ServerList {
		if s.Type == "websockets" && s.Endpoint != "" {
			list = append(list, s.Endpoint)
		}
	}
	if len(list) == 0 {
		return nil, errors.New("steam: no CM servers")
	}
	cmCache, cmCacheAt = list, time.Now()
	return list, nil
}

func dialAny(ctx context.Context) (*conn, error) {
	list, err := cmEndpoints(ctx)
	if err != nil {
		return nil, err
	}
	n := min(len(list), 5)
	var lastErr error
	for _, i := range rand.Perm(n) {
		c, err := dial(ctx, list[i])
		if err == nil {
			return c, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
