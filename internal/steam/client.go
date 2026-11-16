package steam

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"math/rand/v2"
	"sort"
	"strings"
	"sync"
	"time"
)

const chatEntryChatMsg = 1

// Friend names and avatars are fetched from the web API this often.
const profileRefresh = time.Minute

// State is the connection state of the Steam session.
type State string

const (
	StateOffline    State = "offline"
	StateConnecting State = "connecting"
	StateOnline     State = "online"
	StateNeedsLogin State = "needs_login"
)

// Friend is one friend's latest known profile data.
type Friend struct {
	SteamID      uint64
	Name         string
	Nickname     string // the nickname the user gave this friend, if any
	PersonaState uint32 // 0 offline, 1 online, 2 busy, 3 away, 4 snooze, 5 trade, 6 play
	GameID       uint64
	AvatarHash   string // hex
}

const defaultAvatar = "fef49e7fa7e1997310d705b2a6158ff8dc1cdfeb"

// AvatarURL is the 64 px avatar on Steam's public image server.
func (f Friend) AvatarURL() string {
	h := f.AvatarHash
	if h == "" || strings.Trim(h, "0") == "" {
		h = defaultAvatar
	}
	return "https://avatars.steamstatic.com/" + h + "_medium.jpg"
}

// Event types sent on Client.Events().
type (
	// StatusEvent reports a connection state change.
	StatusEvent struct {
		State  State
		Detail string
	}
	// FriendsEvent reports that the friends list or profile data changed.
	FriendsEvent struct{}
	// MessageEvent is a friend chat message. When Echo is true the message
	// was sent by this account from another session (for example the Steam
	// client); Partner is then the receiver.
	MessageEvent struct {
		Partner uint64
		Text    string // Steam's plain-text version of the message
		BBCode  string // the message as sent, with BBCode such as [img]
		Time    time.Time
		Echo    bool
	}
)

// Client is a background Steam session that only does friends chat.
//
// It stays a silent, offline session: it never sets a status
// (ClientChangeStatus) and never asks for persona data
// (ClientRequestFriendData). Steam copies the answer to a persona request to
// the user's other sessions, and an offline session is answered with "offline"
// for everyone, so the Steam app would show itself signed out or all friends
// offline. Names and avatars come from the web API instead (profiles.go).
// Steam still delivers friend messages to an offline session.
type Client struct {
	machineID []byte
	logonID   uint32

	events chan any

	mu       sync.RWMutex
	token    string
	self     uint64
	friends  map[uint64]*Friend // includes self
	relation map[uint64]uint32
	state    State
	conn     *conn
	refresh  chan struct{} // asks the profile loop for a fetch now

	access        string // web API access token
	accessExpires time.Time

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a client. machineSeed should be stable per install.
func New(machineSeed string) *Client {
	return &Client{
		machineID: machineID(machineSeed),
		// The obfuscated private IP acts as the logon ID. A value different
		// from the Steam client's keeps both sessions logged on.
		logonID:  rand.Uint32() | 1,
		events:   make(chan any, 256),
		friends:  map[uint64]*Friend{},
		relation: map[uint64]uint32{},
		state:    StateOffline,
		refresh:  make(chan struct{}, 1),
	}
}

// Events returns the event stream. Read it continuously.
func (c *Client) Events() <-chan any { return c.events }

func (c *Client) emit(e any) {
	select {
	case c.events <- e:
	default:
	}
}

func (c *Client) setState(s State, detail string) {
	c.mu.Lock()
	c.state = s
	c.mu.Unlock()
	c.emit(StatusEvent{State: s, Detail: detail})
}

// State returns the current connection state.
func (c *Client) State() State {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// SelfID returns the logged-on SteamID, or 0.
func (c *Client) SelfID() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.self
}

// Start logs on with a refresh token and keeps the session alive until Stop.
func (c *Client) Start(token string) error {
	info, err := ParseToken(token)
	if err != nil {
		return err
	}
	c.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	c.mu.Lock()
	c.token = token
	c.self = info.SteamID
	c.access = ""
	c.cancel = cancel
	c.mu.Unlock()
	c.wg.Add(1)
	go c.run(ctx)
	return nil
}

// Stop logs off and ends the session.
func (c *Client) Stop() {
	c.mu.Lock()
	cancel := c.cancel
	cn := c.conn
	c.cancel = nil
	c.mu.Unlock()
	if cancel == nil {
		return
	}
	if cn != nil {
		_ = cn.send(emsgClientLogOff, newHeader(), nil)
	}
	cancel()
	c.wg.Wait()
	c.setState(StateOffline, "")
}

func (c *Client) run(ctx context.Context) {
	defer c.wg.Done()
	backoff := time.Second
	for ctx.Err() == nil {
		c.setState(StateConnecting, "")
		err := c.session(ctx)
		if ctx.Err() != nil {
			return
		}
		var se *Error
		if errors.As(err, &se) {
			switch se.EResult {
			case EResultInvalidPassword, EResultAccessDenied, EResultRevoked, EResultExpired:
				c.setState(StateNeedsLogin, "Steam login expired. Log in again.")
				return
			}
		}
		detail := ""
		if err != nil {
			detail = err.Error()
		}
		c.setState(StateOffline, detail)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff + time.Duration(rand.IntN(1000))*time.Millisecond):
		}
		backoff = min(backoff*2, 60*time.Second)
		if err == nil {
			backoff = time.Second
		}
	}
}

// session connects, logs on and processes messages until the connection ends.
func (c *Client) session(ctx context.Context) error {
	cn, err := dialAny(ctx)
	if err != nil {
		return err
	}
	defer cn.close()

	c.mu.RLock()
	token, self := c.token, c.self
	c.mu.RUnlock()
	cn.steamID.Store(self)

	var ip enc
	ip.fixed32(1, c.logonID)
	var logon enc
	logon.varint(1, protocolVersion)
	logon.string(6, "english")
	logon.varint(7, 20) // EOSType Win11
	logon.bool(8, true) // should_remember_password
	logon.bytes(11, ip.b)
	logon.bytes(30, c.machineID)
	logon.varint(33, 2) // chat_mode: new Steam chat
	logon.string(96, "myChat")
	logon.bool(102, true) // supports_rate_limit_response
	logon.string(108, token)
	if err := cn.send(emsgClientLogon, newHeader(), logon.b); err != nil {
		return err
	}

	heartbeat := 9 * time.Second
	timeout := time.NewTimer(15 * time.Second)
	defer timeout.Stop()
	loggedOn := false
	for !loggedOn {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return errors.New("steam: logon timed out")
		case p, ok := <-cn.incoming:
			if !ok {
				return errClosed
			}
			if p.emsg != emsgClientLogOnResponse {
				continue
			}
			eresult := int32(EResultFail)
			_ = walk(p.body, func(f field) {
				switch f.Num {
				case 1:
					eresult = f.i32()
				case 3:
					if f.Value > 0 {
						heartbeat = time.Duration(f.Value) * time.Second
					}
				}
			})
			if eresult != EResultOK {
				return &Error{EResult: eresult}
			}
			cn.sessionID.Store(p.hdr.sessionID)
			if p.hdr.steamID != 0 {
				cn.steamID.Store(p.hdr.steamID)
				c.mu.Lock()
				c.self = p.hdr.steamID
				c.mu.Unlock()
			}
			loggedOn = true
		}
	}

	c.mu.Lock()
	c.conn = cn
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		if c.conn == cn {
			c.conn = nil
		}
		c.mu.Unlock()
	}()
	c.setState(StateOnline, "")

	pctx, stopProfiles := context.WithCancel(ctx)
	defer stopProfiles()
	go c.profileLoop(pctx, cn)

	tick := time.NewTicker(heartbeat)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
			if err := cn.send(emsgClientHeartBeat, newHeader(), nil); err != nil {
				return err
			}
		case p, ok := <-cn.incoming:
			if !ok {
				return errClosed
			}
			if err := c.handle(cn, p); err != nil {
				return err
			}
		}
	}
}

func (c *Client) handle(cn *conn, p packet) error {
	switch p.emsg {
	case emsgClientLoggedOff:
		eresult := int32(EResultFail)
		_ = walk(p.body, func(f field) {
			if f.Num == 1 {
				eresult = f.i32()
			}
		})
		return &Error{EResult: eresult, Message: "logged off by Steam"}
	case emsgClientFriendsList:
		c.handleFriendsList(p.body)
	case emsgServiceMethod:
		if p.hdr.targetJobName == "FriendMessagesClient.IncomingMessage#1" {
			c.handleIncoming(p.body)
		}
	}
	return nil
}

func (c *Client) handleFriendsList(body []byte) {
	incremental := false
	type rel struct {
		id  uint64
		rel uint32
	}
	var list []rel
	_ = walk(body, func(f field) {
		switch f.Num {
		case 1:
			incremental = f.boolean()
		case 2:
			var r rel
			_ = walk(f.Bytes, func(g field) {
				switch g.Num {
				case 1:
					r.id = g.Value
				case 2:
					r.rel = g.u32()
				}
			})
			list = append(list, r)
		}
	})
	c.mu.Lock()
	if !incremental {
		c.relation = map[uint64]uint32{}
	}
	for _, r := range list {
		if r.rel == 0 {
			delete(c.relation, r.id)
			continue
		}
		c.relation[r.id] = r.rel
	}
	c.mu.Unlock()
	c.RefreshProfiles()
	c.emit(FriendsEvent{})
}

func (c *Client) handleIncoming(body []byte) {
	var m MessageEvent
	var entryType int32
	var text, plain string
	var ts uint32
	_ = walk(body, func(f field) {
		switch f.Num {
		case 1:
			m.Partner = f.Value
		case 2:
			entryType = f.i32()
		case 4:
			text = f.str()
		case 5:
			ts = f.u32()
		case 7:
			m.Echo = f.boolean()
		case 8:
			plain = f.str()
		}
	})
	if entryType != chatEntryChatMsg || m.Partner == 0 {
		return
	}
	m.BBCode = text
	if plain != "" {
		text = plain
	}
	m.Text = text
	m.Time = time.Now()
	if ts != 0 {
		m.Time = time.Unix(int64(ts), 0)
	}
	c.emit(m)
}

func (c *Client) currentConn() (*conn, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conn == nil {
		return nil, errors.New("not connected to Steam")
	}
	return c.conn, nil
}

// SendMessage sends a friend chat message.
func (c *Client) SendMessage(ctx context.Context, to uint64, text string) error {
	cn, err := c.currentConn()
	if err != nil {
		return err
	}
	var e enc
	e.fixed64(1, to)
	e.int32(2, chatEntryChatMsg)
	e.string(3, text)
	e.bool(4, false)
	e.bool(5, true) // echo_to_sender: show it in the Steam client too
	_, err = cn.call(ctx, emsgServiceMethodCallFromClient, "FriendMessages.SendMessage#1", e.b)
	return err
}

// Ack marks a conversation as read up to ts.
func (c *Client) Ack(partner uint64, ts time.Time) {
	cn, err := c.currentConn()
	if err != nil {
		return
	}
	var e enc
	e.fixed64(1, partner)
	e.varint(2, uint64(ts.Unix()))
	_ = cn.notify("FriendMessages.AckMessage#1", e.b)
}

// Friends returns all friends (relationship "friend"), sorted by name.
func (c *Client) Friends() []Friend {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Friend, 0, len(c.relation))
	for id, r := range c.relation {
		if r != 3 {
			continue
		}
		f := Friend{SteamID: id}
		if p := c.friends[id]; p != nil {
			f = *p
		}
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out
}

// Friend returns one user's persona data, if known.
func (c *Client) Friend(id uint64) (Friend, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	p := c.friends[id]
	if p == nil {
		return Friend{SteamID: id}, false
	}
	return *p, true
}

// machineID builds the binary KeyValues machine ID that the Steam client sends.
func machineID(seed string) []byte {
	h := func(s string) string {
		sum := sha1.Sum([]byte(seed + s))
		return hex.EncodeToString(sum[:])
	}
	var b []byte
	str := func(s string) { b = append(append(b, s...), 0) }
	b = append(b, 0)
	str("MessageObject")
	for _, k := range []string{"BB3", "FF2", "3B3"} {
		b = append(b, 1)
		str(k)
		str(h(k))
	}
	return append(b, 8, 8)
}
