package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	Prefix = "/media/"

	maxBytes     = 8 << 20
	maxFullBytes = 32 << 20
	maxTotal     = 160 << 20
	maxItems     = 64
	fetchTimeout = 30 * time.Second
)

type Picture struct {
	Data   []byte
	Type   string
	Width  int
	Height int
}

var (
	ErrUnknown    = errors.New("media: unknown picture")
	errNotImage   = errors.New("media: not a picture")
	errTooBig     = errors.New("media: picture too big")
	errBlocked    = errors.New("media: address not allowed")
	allowedImages = map[string]bool{"image/png": true, "image/jpeg": true, "image/gif": true, "image/webp": true}
)

type item struct {
	url     string
	limit   int64
	started bool
	done    chan struct{}
	pic     Picture
	err     error
}

type Store struct {
	client *http.Client

	mu    sync.Mutex
	items map[string]*item
	order []string
	total int
}

func NewStore() *Store {
	dialer := &net.Dialer{Timeout: 10 * time.Second, Control: publicOnly}
	return &Store{
		items: map[string]*item{},
		client: &http.Client{
			Timeout: fetchTimeout,
			Transport: &http.Transport{
				Proxy:                 nil,
				DialContext:           dialer.DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				MaxIdleConns:          8,
				IdleConnTimeout:       90 * time.Second,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return errors.New("media: too many redirects")
				}
				if req.URL.Scheme != "https" && req.URL.Scheme != "http" {
					return errBlocked
				}
				if trusted(via[0].URL) && !trusted(req.URL) {
					return errBlocked
				}
				return nil
			},
		},
	}
}

func publicOnly(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return errBlocked
	}
	return nil
}

func ID(rawURL string) string {
	h := sha256.Sum256([]byte(rawURL))
	return hex.EncodeToString(h[:12])
}

func (s *Store) Start(rawURL string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, it := s.addLocked(rawURL, maxBytes)
	s.startLocked(it)
	return id
}

func (s *Store) Later(rawURL string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, _ := s.addLocked(rawURL, maxFullBytes)
	return id
}

func (s *Store) addLocked(rawURL string, limit int64) (string, *item) {
	id := ID(rawURL)
	if it, ok := s.items[id]; ok {
		return id, it
	}
	it := &item{url: rawURL, limit: limit, done: make(chan struct{})}
	s.items[id] = it
	s.order = append(s.order, id)
	s.evictLocked()
	return id, it
}

func (s *Store) startLocked(it *item) {
	if it.started {
		return
	}
	it.started = true
	go func() {
		pic, err := s.fetch(it.url, it.limit)
		s.mu.Lock()
		it.pic, it.err = pic, err
		s.total += len(pic.Data)
		s.evictLocked()
		s.mu.Unlock()
		close(it.done)
	}()
}

func (s *Store) Wait(ctx context.Context, id string) (Picture, error) {
	s.mu.Lock()
	it, ok := s.items[id]
	if ok {
		s.startLocked(it)
	}
	s.mu.Unlock()
	if !ok {
		return Picture{}, ErrUnknown
	}
	select {
	case <-it.done:
		return it.pic, it.err
	case <-ctx.Done():
		return Picture{}, ctx.Err()
	}
}

func (s *Store) evictLocked() {
	for i := 0; i < len(s.order)-1 && (s.total > maxTotal || len(s.order) > maxItems); {
		id := s.order[i]
		it := s.items[id]
		if it.started {
			select {
			case <-it.done:
			default:
				i++
				continue
			}
		}
		s.total -= len(it.pic.Data)
		delete(s.items, id)
		s.order = append(s.order[:i], s.order[i+1:]...)
	}
}

func (s *Store) fetch(rawURL string, limit int64) (Picture, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return Picture{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) myChat")
	req.Header.Set("Accept", "image/webp,image/png,image/jpeg,image/gif;q=0.9,*/*;q=0.5")
	resp, err := s.client.Do(req)
	if err != nil {
		return Picture{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Picture{}, fmt.Errorf("media: HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > limit {
		return Picture{}, errTooBig
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return Picture{}, err
	}
	if int64(len(data)) > limit {
		return Picture{}, errTooBig
	}
	typ := http.DetectContentType(data)
	if !allowedImages[typ] {
		return Picture{}, errNotImage
	}
	pic := Picture{Data: data, Type: typ}
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
		pic.Width, pic.Height = cfg.Width, cfg.Height
	}
	return pic, nil
}

func (s *Store) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, Prefix) {
			next.ServeHTTP(w, r)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), fetchTimeout+5*time.Second)
		defer cancel()
		pic, err := s.Wait(ctx, strings.TrimPrefix(r.URL.Path, Prefix))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		h := w.Header()
		h.Set("Content-Type", pic.Type)
		h.Set("Content-Length", fmt.Sprint(len(pic.Data)))
		h.Set("Cache-Control", "private, max-age=86400, immutable")
		h.Set("X-Content-Type-Options", "nosniff")
		_, _ = w.Write(pic.Data)
	})
}
