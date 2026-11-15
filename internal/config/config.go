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
	"strings"
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
	Hotkey         Hotkey `json:"hotkey"`
	CloseAfterSend bool   `json:"closeAfterSend"`
	// DoubleEnterClose applies when CloseAfterSend is off: Enter on an empty
	// line closes the box only when it follows another Enter quickly.
	DoubleEnterClose bool   `json:"doubleEnterClose"`
	Animations       bool   `json:"animations"`
	Blur             bool   `json:"blur"`
	AnimationStyle   string `json:"animationStyle"`
	AnimationSpeed   string `json:"animationSpeed"`
	ShowAvatars      bool   `json:"showAvatars"`
	ShowToasts       bool   `json:"showToasts"`
	ToastSeconds     int    `json:"toastSeconds"`
	// ToastLayout is how several new messages show: "stack" (one box per
	// message) or "group" (consecutive messages from one person share a
	// box). ToastMax is how many messages show at once (1..5).
	ToastLayout  string  `json:"toastLayout"`
	ToastMax     int     `json:"toastMax"`
	BottomOffset int     `json:"bottomOffset"`
	Scale        float64 `json:"scale"`
	// OverlayStyle is the look of the chat box and pop-ups: "legacy" (CS2
	// chat) or "minimal" (light, compact).
	OverlayStyle string `json:"overlayStyle"`
	// ShowAccountName shows the Steam account (login) name in the app; the
	// persona name is always shown.
	ShowAccountName bool   `json:"showAccountName"`
	MachineSeed     string `json:"machineSeed"`
	// LoginSource is how myChat logs in: "steam" (the login saved by the
	// Steam app on this PC), "qr" or "password" (myChat's own login, kept in
	// session.dat) or "" (not chosen yet). SteamAccount is the Steam account
	// name for "steam".
	LoginSource  string `json:"loginSource"`
	SteamAccount string `json:"steamAccount"`
}

// Default returns the default settings.
func Default() Config {
	return Config{
		Hotkey:           Hotkey{Code: "KeyI"},
		CloseAfterSend:   true,
		DoubleEnterClose: true,
		Animations:       true,
		Blur:             false, // unused: DWM blur-behind turns the overlay opaque
		AnimationStyle:   "slide-up",
		AnimationSpeed:   "normal",
		ShowAvatars:      true,
		ShowToasts:       true,
		ToastSeconds:     6,
		ToastLayout:      "group",
		ToastMax:         3,
		BottomOffset:     150,
		Scale:            1,
		OverlayStyle:     "legacy",
	}
}

// AnimationStyles are the overlay animation styles.
var AnimationStyles = map[string]bool{
	"fade": true, "slide-up": true, "slide-down": true,
	"slide-left": true, "slide-right": true, "typewriter": true,
}

// AnimationSpeeds maps a speed to the time Go waits before hiding the window
// (the UI's exit animation must end before it).
var AnimationSpeeds = map[string]int{"fast": 120, "normal": 200, "slow": 320}

// Sanitize clamps values to safe ranges.
func (c *Config) Sanitize() {
	d := Default()
	// Windows blur-behind made the overlay opaque and was removed; clear it
	// from configs saved before that.
	c.Blur = false
	if c.Hotkey.Code == "" {
		c.Hotkey = d.Hotkey
	}
	if !AnimationStyles[c.AnimationStyle] {
		c.AnimationStyle = d.AnimationStyle
	}
	if _, ok := AnimationSpeeds[c.AnimationSpeed]; !ok {
		c.AnimationSpeed = d.AnimationSpeed
	}
	if c.ToastSeconds < 2 || c.ToastSeconds > 30 {
		c.ToastSeconds = d.ToastSeconds
	}
	if c.ToastLayout != "stack" && c.ToastLayout != "group" {
		c.ToastLayout = d.ToastLayout
	}
	if c.ToastMax < 1 || c.ToastMax > 5 {
		c.ToastMax = d.ToastMax
	}
	if c.BottomOffset < 0 || c.BottomOffset > 1000 {
		c.BottomOffset = d.BottomOffset
	}
	if c.Scale < 0.75 || c.Scale > 1.5 {
		c.Scale = d.Scale
	}
	if c.OverlayStyle != "legacy" && c.OverlayStyle != "minimal" {
		c.OverlayStyle = d.OverlayStyle
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

// Session is myChat's own saved Steam login.
type Session struct {
	RefreshToken string `json:"refreshToken"`
	AccountName  string `json:"accountName"`
	Method       string `json:"method"` // "qr" | "password"
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

// LoadGuardData returns Steam's machine token for an account (email Steam
// Guard remembers this PC with it). It survives logout.
func (s *Store) LoadGuardData(account string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readGuardLocked()[strings.ToLower(account)]
}

// SaveGuardData stores Steam's machine token for an account.
func (s *Store) SaveGuardData(account, data string) {
	if data == "" || account == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.readGuardLocked()
	m[strings.ToLower(account)] = data
	plain, _ := json.Marshal(m)
	if enc, err := protect(plain); err == nil {
		_ = os.WriteFile(filepath.Join(s.dir, "guard.dat"), enc, 0o600)
	}
}

func (s *Store) readGuardLocked() map[string]string {
	m := map[string]string{}
	b, err := os.ReadFile(filepath.Join(s.dir, "guard.dat"))
	if err != nil {
		return m
	}
	if plain, err := unprotect(b); err == nil {
		_ = json.Unmarshal(plain, &m)
	}
	return m
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
