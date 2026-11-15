//go:build windows

// Package config stores settings and the encrypted Steam session token in
// %APPDATA%\myChat.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Hotkey is stored as a KeyboardEvent.code (physical key) plus modifiers.
type Hotkey struct {
	Code  string `json:"code"`
	Ctrl  bool   `json:"ctrl"`
	Shift bool   `json:"shift"`
	Alt   bool   `json:"alt"`
}

// Config is the user's settings.
type Config struct {
	Hotkey         Hotkey  `json:"hotkey"`
	CloseAfterSend bool    `json:"closeAfterSend"`
	Animations     bool    `json:"animations"`
	Blur           bool    `json:"blur"`
	ShowToasts     bool    `json:"showToasts"`
	ToastSeconds   int     `json:"toastSeconds"`
	BottomOffset   int     `json:"bottomOffset"`
	Scale          float64 `json:"scale"`
	MachineSeed    string  `json:"machineSeed"`
}

// Default returns the default settings.
func Default() Config {
	return Config{
		Hotkey:         Hotkey{Code: "KeyI"},
		CloseAfterSend: true,
		Animations:     true,
		Blur:           true,
		ShowToasts:     true,
		ToastSeconds:   6,
		BottomOffset:   150,
		Scale:          1,
	}
}

// Sanitize clamps values to safe ranges.
func (c *Config) Sanitize() {
	d := Default()
	if c.Hotkey.Code == "" {
		c.Hotkey = d.Hotkey
	}
	if c.ToastSeconds < 2 || c.ToastSeconds > 30 {
		c.ToastSeconds = d.ToastSeconds
	}
	if c.BottomOffset < 0 || c.BottomOffset > 1000 {
		c.BottomOffset = d.BottomOffset
	}
	if c.Scale < 0.75 || c.Scale > 1.5 {
		c.Scale = d.Scale
	}
}

// Store reads and writes the files.
type Store struct {
	mu  sync.Mutex
	dir string
}

// Open prepares %APPDATA%\myChat.
func Open() (*Store, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(base, "myChat")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

// Dir returns the data folder.
func (s *Store) Dir() string { return s.dir }

// Load returns the saved settings, or defaults.
func (s *Store) Load() Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := Default()
	if b, err := os.ReadFile(filepath.Join(s.dir, "config.json")); err == nil {
		_ = json.Unmarshal(b, &c)
	}
	c.Sanitize()
	if c.MachineSeed == "" {
		var r [16]byte
		_, _ = rand.Read(r[:])
		c.MachineSeed = hex.EncodeToString(r[:])
		_ = s.saveLocked(c)
	}
	return c
}

// Save writes the settings.
func (s *Store) Save(c Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(c)
}

func (s *Store) saveLocked(c Config) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(s.dir, "config.json.tmp")
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(s.dir, "config.json"))
}

// Session is the saved Steam login.
type Session struct {
	RefreshToken string `json:"refreshToken"`
	AccountName  string `json:"accountName"`
}

// LoadSession decrypts the saved login. Only this Windows user can decrypt it.
func (s *Store) LoadSession() (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := os.ReadFile(filepath.Join(s.dir, "session.dat"))
	if err != nil {
		return Session{}, err
	}
	plain, err := unprotect(b)
	if err != nil {
		return Session{}, err
	}
	var sess Session
	if err := json.Unmarshal(plain, &sess); err != nil {
		return Session{}, err
	}
	if sess.RefreshToken == "" {
		return Session{}, errors.New("config: empty session")
	}
	return sess, nil
}

// SaveSession encrypts and saves the login.
func (s *Store) SaveSession(sess Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	plain, _ := json.Marshal(sess)
	enc, err := protect(plain)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, "session.dat"), enc, 0o600)
}

// ClearSession deletes the saved login.
func (s *Store) ClearSession() {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = os.Remove(filepath.Join(s.dir, "session.dat"))
}

var entropy = []byte("myChat.steam.session.v1")

func blob(b []byte) *windows.DataBlob {
	if len(b) == 0 {
		return &windows.DataBlob{}
	}
	return &windows.DataBlob{Size: uint32(len(b)), Data: &b[0]}
}

func protect(plain []byte) ([]byte, error) {
	var out windows.DataBlob
	if err := windows.CryptProtectData(blob(plain), nil, blob(entropy), 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return append([]byte(nil), unsafe.Slice(out.Data, out.Size)...), nil
}

func unprotect(enc []byte) ([]byte, error) {
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(blob(enc), nil, blob(entropy), 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return append([]byte(nil), unsafe.Slice(out.Data, out.Size)...), nil
}
