package main

import (
	"context"
	"errors"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/wailsapp/wails/v3/pkg/application"

	"mychat/internal/chat"
	"mychat/internal/config"
	"mychat/internal/input"
	"mychat/internal/steam"
	"mychat/internal/win"
)

// Overlay modes.
const (
	ModeHidden   = "hidden"
	ModeToast    = "toast"
	ModeComposer = "composer"
)

// Event names sent to the UI.
const (
	EvMode         = "mychat:mode"
	EvComposer     = "mychat:composer"
	EvEntry        = "mychat:entry"
	EvEntryUpdated = "mychat:entry-updated"
	EvLogCleared   = "mychat:log-cleared"
	EvToasts       = "mychat:toasts"
	EvStatus       = "mychat:status"
	EvConfig       = "mychat:config"
	EvLogin        = "mychat:login"
)

// ModeInfo tells the overlay what to draw.
type ModeInfo struct {
	Mode           string `json:"mode"`
	Animations     bool   `json:"animations"`
	AnimationStyle string `json:"animationStyle"`
	AnimationSpeed string `json:"animationSpeed"`
	HideDelayMs    int    `json:"hideDelayMs"`
	Blur           bool   `json:"blur"`
	ShowAvatars    bool   `json:"showAvatars"`
}

// Status is the Steam connection summary.
type Status struct {
	State       string `json:"state"` // offline | connecting | online | needs_login
	Detail      string `json:"detail"`
	AccountName string `json:"accountName"`
	Source      string `json:"source"` // "steam" | "qr" | "password" | ""
	PersonaName string `json:"personaName"`
}

// ConfigView is the config plus derived display values.
type ConfigView struct {
	config.Config
	HotkeyLabel string `json:"hotkeyLabel"`
}

// Layout gives the overlay sizes in CSS pixels.
type Layout struct {
	Width          int `json:"width"`
	ComposerHeight int `json:"composerHeight"`
	ToastHeight    int `json:"toastHeight"`
}

// Toasts are the messages shown as pop-ups, oldest first. The UI lays them
// out per Config.ToastLayout.
type Toasts struct {
	Entries []chat.Entry `json:"entries"`
}

// Snapshot is the full state for a newly loaded page.
type Snapshot struct {
	Mode     ModeInfo           `json:"mode"`
	Toasts   Toasts             `json:"toasts"`
	Log      []chat.Entry       `json:"log"`
	Composer chat.ComposerState `json:"composer"`
	Status   Status             `json:"status"`
	Config   ConfigView         `json:"config"`
	Login    LoginState         `json:"login"`
	Local    LocalLogins        `json:"local"`
	Layout   Layout             `json:"layout"`
}

func init() {
	application.RegisterEvent[ModeInfo](EvMode)
	application.RegisterEvent[chat.ComposerState](EvComposer)
	application.RegisterEvent[chat.Entry](EvEntry)
	application.RegisterEvent[chat.Entry](EvEntryUpdated)
	application.RegisterEvent[application.Void](EvLogCleared)
	application.RegisterEvent[Toasts](EvToasts)
	application.RegisterEvent[Status](EvStatus)
	application.RegisterEvent[ConfigView](EvConfig)
	application.RegisterEvent[LoginState](EvLogin)
}

// App wires input, Steam, chat logic and windows together.
type App struct {
	app      *application.App
	store    *config.Store
	steam    *steam.Client
	engine   *chat.Engine
	inp      *input.Manager
	ov       overlay
	settings *application.WebviewWindow

	mu          sync.Mutex
	cfg         config.Config
	session     config.Session
	mode        string
	hideTimer   *time.Timer
	toastTimer  *time.Timer
	toasts      []chat.Entry // shown as pop-ups, oldest first
	openFG      uintptr
	lastEnter   time.Time // for double-tap Enter to close
	lastRefresh time.Time
	login       LoginState
	loginCancel context.CancelFunc
	credLogin   *steam.CredentialLogin // password login waiting for Steam Guard
	watchCancel context.CancelFunc     // waiting for a Steam app sign-in
	steamToken  string                 // token in use with the Steam source
	status      Status
	ready       bool

	settingsReady   bool
	settingsPending bool
}

func newApp(store *config.Store) *App {
	a := &App{store: store, mode: ModeHidden, login: LoginState{Phase: "idle"}}
	a.cfg = store.Load()
	a.steam = steam.New(a.cfg.MachineSeed)
	a.engine = chat.New(backend{a.steam}, notifier{a})
	a.status = Status{State: string(steam.StateOffline)}
	return a
}

func (a *App) emit(name string, data any) {
	if a.app != nil {
		a.app.Event.Emit(name, data)
	}
}

// start runs after the Wails app has started.
func (a *App) start() {
	hk, err := a.hotkey(a.cfg.Hotkey)
	if err == nil {
		a.inp, err = input.Start(hk)
	}
	if err != nil {
		log.Printf("hotkey: %v", err)
		a.engine.System("Could not register the hotkey. Pick another one in Settings.")
		if a.inp == nil {
			a.inp, _ = input.Start(input.Hotkey{})
		}
	}

	a.autoLogin()

	go a.pump()
}

func (a *App) hotkey(h config.Hotkey) (input.Hotkey, error) {
	vk, err := input.CodeToVK(h.Code)
	if err != nil {
		return input.Hotkey{}, err
	}
	return input.Hotkey{VK: vk, Ctrl: h.Ctrl, Shift: h.Shift, Alt: h.Alt}, nil
}

// pump is the single event loop for hotkeys, typed keys and Steam events.
func (a *App) pump() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("pump panic: %v\n%s", r, debug.Stack())
		}
	}()
	fg := time.NewTicker(250 * time.Millisecond)
	defer fg.Stop()
	for {
		select {
		case <-a.inp.Hotkeys:
			a.openComposer()
		case ev := <-a.inp.Events:
			a.onKey(ev)
		case ev := <-a.steam.Events():
			a.onSteam(ev)
		case <-fg.C:
			a.mu.Lock()
			open, want := a.mode == ModeComposer, a.openFG
			a.mu.Unlock()
			// Close if the user left the game window some other way.
			if open && want != 0 && win.ForegroundWindow() != want {
				a.closeComposer("game window lost focus")
			}
		}
	}
}

// doubleEnterWindow is how quickly a second Enter must follow the first to
// count as a double tap.
const doubleEnterWindow = 500 * time.Millisecond

func (a *App) onKey(ev input.Event) {
	now := time.Now()
	a.mu.Lock()
	open := a.mode == ModeComposer
	doubleTap := !a.cfg.CloseAfterSend && a.cfg.DoubleEnterClose
	quick := now.Sub(a.lastEnter) <= doubleEnterWindow
	if ev.Key == input.KeyEnter {
		a.lastEnter = now
	}
	a.mu.Unlock()
	if !open {
		return
	}
	switch ev.Key {
	case input.KeyFocusLost:
		a.closeComposer("alt+tab or windows key")
		return
	case input.KeyPaste:
		if text, ok := a.app.Clipboard.Text(); ok {
			a.engine.Insert(text)
		}
		return
	}
	switch a.engine.HandleKey(ev.Text, string(ev.Key), ev.Shift) {
	case chat.ActClose:
		if ev.Key == input.KeyEnter && doubleTap && !quick {
			return // a single Enter on an empty line keeps the box open
		}
		a.closeComposer("esc or enter on empty line")
	case chat.ActSent:
		a.mu.Lock()
		closeAfter := a.cfg.CloseAfterSend
		a.mu.Unlock()
		if closeAfter {
			a.closeComposer("sent")
		}
	}
}

func (a *App) onSteam(ev any) {
	switch e := ev.(type) {
	case steam.StatusEvent:
		log.Printf("steam: %s %s", e.State, e.Detail)
		a.setStatus(func(s *Status) {
			s.State = string(e.State)
			s.Detail = e.Detail
		})
		if e.State == steam.StateNeedsLogin {
			a.onLoginExpired()
		}
		if e.State == steam.StateOnline {
			a.updatePersonaName()
		}
	case steam.FriendsEvent:
		a.updatePersonaName()
		a.engine.Refresh()
	case steam.MessageEvent:
		ent, isNew := a.engine.Incoming(e.Partner, e.Text, e.Time, e.Echo)
		if isNew {
			a.toast(ent)
		}
	}
}

func (a *App) updatePersonaName() {
	self, _ := a.steam.Friend(a.steam.SelfID())
	a.setStatus(func(s *Status) { s.PersonaName = self.Name })
}

func (a *App) setStatus(f func(*Status)) {
	a.mu.Lock()
	f(&a.status)
	s := a.status
	a.mu.Unlock()
	a.emit(EvStatus, s)
}

func (a *App) modeInfoLocked() ModeInfo {
	return ModeInfo{Mode: a.mode, Animations: a.cfg.Animations, AnimationStyle: a.cfg.AnimationStyle,
		AnimationSpeed: a.cfg.AnimationSpeed, HideDelayMs: int(a.animDelay() / time.Millisecond),
		Blur: a.cfg.Blur, ShowAvatars: a.cfg.ShowAvatars}
}

func (a *App) animDelay() time.Duration {
	if a.cfg.Animations {
		return time.Duration(config.AnimationSpeeds[a.cfg.AnimationSpeed]) * time.Millisecond
	}
	return 0
}

func (a *App) stopTimersLocked() {
	if a.hideTimer != nil {
		a.hideTimer.Stop()
		a.hideTimer = nil
	}
	if a.toastTimer != nil {
		a.toastTimer.Stop()
		a.toastTimer = nil
	}
}

// openComposer shows the chat box and starts catching keys.
func (a *App) openComposer() {
	a.mu.Lock()
	if !a.ready || a.mode == ModeComposer {
		a.mu.Unlock()
		return
	}
	a.stopTimersLocked()
	a.toasts = nil
	a.mode = ModeComposer
	a.openFG = win.ForegroundWindow()
	cfg := a.cfg
	info := a.modeInfoLocked()
	refresh := time.Since(a.lastRefresh) > 10*time.Second
	if refresh {
		a.lastRefresh = time.Now()
	}
	a.mu.Unlock()

	log.Printf("chat box opened over %q", win.WindowTitle(win.ForegroundWindow()))
	a.engine.Reset()
	a.inp.Capture(true)
	application.InvokeSync(func() {
		a.ov.show(overlayWidth, composerHeight, cfg.BottomOffset, cfg.Scale)
	})
	a.emit(EvMode, info)
	if refresh {
		a.steam.RefreshProfiles()
	}
	a.engine.AckAll()
}

// closeComposer stops catching keys and hides the chat box.
func (a *App) closeComposer(reason string) {
	a.mu.Lock()
	if a.mode != ModeComposer {
		a.mu.Unlock()
		return
	}
	caught, passed, hooked := a.inp.Stats()
	log.Printf("chat box closed (%s): hook active=%v, keys caught=%d, keys passed to game=%d", reason, hooked, caught, passed)
	a.mode = ModeHidden
	a.openFG = 0
	info := a.modeInfoLocked()
	a.scheduleHideLocked(a.animDelay())
	a.mu.Unlock()
	a.inp.Capture(false)
	a.emit(EvMode, info)
	a.engine.Reset()
}

func (a *App) scheduleHideLocked(d time.Duration) {
	if a.hideTimer != nil {
		a.hideTimer.Stop()
	}
	a.hideTimer = time.AfterFunc(d, func() {
		a.mu.Lock()
		hidden := a.mode == ModeHidden
		a.mu.Unlock()
		if hidden {
			application.InvokeSync(a.ov.hide)
		}
	})
}

// toast shows a new message at the bottom middle while the box is closed.
func (a *App) toast(ent chat.Entry) {
	a.mu.Lock()
	if !a.ready || !a.cfg.ShowToasts || a.mode == ModeComposer {
		a.mu.Unlock()
		return
	}
	a.stopTimersLocked()
	wasHidden := a.mode == ModeHidden
	if wasHidden {
		a.toasts = nil // a fresh set after the last one faded out
	}
	a.mode = ModeToast
	cfg := a.cfg
	a.toasts = append(a.toasts, ent)
	if n := len(a.toasts) - cfg.ToastMax; n > 0 {
		a.toasts = append([]chat.Entry(nil), a.toasts[n:]...)
	}
	list := Toasts{Entries: append([]chat.Entry(nil), a.toasts...)}
	info := a.modeInfoLocked()
	a.toastTimer = time.AfterFunc(time.Duration(cfg.ToastSeconds)*time.Second, a.endToast)
	a.mu.Unlock()

	if wasHidden {
		application.InvokeSync(func() {
			a.ov.show(overlayWidth, toastHeight, cfg.BottomOffset, cfg.Scale)
		})
	}
	a.emit(EvToasts, list)
	a.emit(EvMode, info)
}

func (a *App) endToast() {
	a.mu.Lock()
	if a.mode != ModeToast {
		a.mu.Unlock()
		return
	}
	a.mode = ModeHidden
	info := a.modeInfoLocked()
	a.scheduleHideLocked(a.animDelay())
	a.mu.Unlock()
	a.emit(EvMode, info)
}

func (a *App) showSettings() {
	if a.settings == nil {
		return
	}
	// Focusing before WebView2 has loaded crashes inside Wails, so an early
	// request waits for settingsLoaded.
	a.mu.Lock()
	if !a.settingsReady {
		a.settingsPending = true
		a.mu.Unlock()
		return
	}
	a.mu.Unlock()
	application.InvokeAsync(func() {
		a.settings.Show()
		a.settings.Focus()
	})
}

// settingsLoaded runs when the settings page has loaded.
func (a *App) settingsLoaded() {
	a.mu.Lock()
	first := !a.settingsReady
	pending := a.settingsPending
	a.settingsReady, a.settingsPending = true, false
	a.mu.Unlock()
	log.Printf("settings loaded (pending show: %v)", pending)
	if first && pending {
		a.showSettings()
	}
}

func (a *App) snapshot() Snapshot {
	local := a.localLogins()
	a.mu.Lock()
	defer a.mu.Unlock()
	return Snapshot{
		Local:    local,
		Toasts:   Toasts{Entries: append([]chat.Entry{}, a.toasts...)},
		Mode:     a.modeInfoLocked(),
		Log:      a.engine.Log(),
		Composer: a.engine.Composer(),
		Status:   a.status,
		Config:   a.configViewLocked(),
		Login:    a.login,
		Layout:   Layout{Width: overlayWidth, ComposerHeight: composerHeight, ToastHeight: toastHeight},
	}
}

func (a *App) configViewLocked() ConfigView {
	return ConfigView{Config: a.cfg, HotkeyLabel: hotkeyLabel(a.cfg.Hotkey)}
}

func hotkeyLabel(h config.Hotkey) string {
	s := ""
	if h.Ctrl {
		s += "Ctrl + "
	}
	if h.Alt {
		s += "Alt + "
	}
	if h.Shift {
		s += "Shift + "
	}
	return s + input.CodeLabel(h.Code)
}

func (a *App) setLogin(l LoginState) {
	a.mu.Lock()
	a.login = l
	a.mu.Unlock()
	a.emit(EvLogin, l)
}

func qrRows(url string) []string {
	q, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return nil
	}
	q.DisableBorder = true
	bits := q.Bitmap()
	rows := make([]string, len(bits))
	for i, r := range bits {
		b := make([]byte, len(r))
		for j, v := range r {
			if v {
				b[j] = '1'
			} else {
				b[j] = '0'
			}
		}
		rows[i] = string(b)
	}
	return rows
}

// ---- Service bound to the UI ----

// ChatService is the API the UI calls.
type ChatService struct{ a *App }

// Snapshot returns the full state.
func (s *ChatService) Snapshot() Snapshot { return s.a.snapshot() }

// OverlayReady is called once by the overlay page after it has loaded.
func (s *ChatService) OverlayReady() {
	a := s.a
	a.mu.Lock()
	if a.ready {
		a.mu.Unlock()
		return
	}
	a.mu.Unlock()
	application.InvokeSync(func() {
		// Show once off screen so WebView2 marks the page visible, then hide.
		a.ov.w.SetPosition(-32000, -32000)
		a.ov.w.Show()
		a.ov.attach()
		a.ov.hide()
	})
	a.mu.Lock()
	a.ready = true
	a.mu.Unlock()
	go a.demo()
}

// SaveConfig applies and saves settings. It returns the stored values.
func (s *ChatService) SaveConfig(c config.Config) (ConfigView, error) {
	a := s.a
	c.Sanitize()
	a.mu.Lock()
	old := a.cfg
	c.MachineSeed = old.MachineSeed
	c.LoginSource, c.SteamAccount = old.LoginSource, old.SteamAccount
	a.mu.Unlock()

	if c.Hotkey != old.Hotkey {
		hk, err := a.hotkey(c.Hotkey)
		if err != nil {
			return ConfigView{}, err
		}
		if err := a.inp.SetHotkey(hk); err != nil {
			return ConfigView{}, errors.New("that key is already in use by another app")
		}
	}
	if err := a.store.Save(c); err != nil {
		return ConfigView{}, err
	}
	a.mu.Lock()
	a.cfg = c
	view := a.configViewLocked()
	a.mu.Unlock()
	if c.Scale != old.Scale {
		application.InvokeSync(func() { a.ov.w.SetZoom(c.Scale) })
	}
	a.emit(EvConfig, view)
	return view, nil
}

// PreviewToast shows a sample toast so the user can check position and look.
func (s *ChatService) PreviewToast() {
	s.a.toast(chat.Entry{ID: -time.Now().UnixMilli(), Time: time.Now().UnixMilli(), Kind: chat.KindIn, PeerName: "myChat", Text: "This is how new messages look in game."})
}

// HideSettings hides the settings window.
func (s *ChatService) HideSettings() {
	application.InvokeAsync(func() { s.a.settings.Hide() })
}

// MinimiseSettings minimises the settings window.
func (s *ChatService) MinimiseSettings() {
	application.InvokeAsync(func() { s.a.settings.Minimise() })
}

// Quit exits the app.
func (s *ChatService) Quit() { s.a.app.Quit() }

// ---- adapters ----

type backend struct{ c *steam.Client }

func toChat(f steam.Friend) chat.Friend {
	game := ""
	if f.GameID == 730 {
		game = "Counter-Strike 2"
	}
	name, alt := f.Name, ""
	if f.Nickname != "" && f.Nickname != f.Name {
		name, alt = f.Nickname, f.Name
	}
	return chat.Friend{ID: f.SteamID, Name: name, AltName: alt, Online: f.PersonaState != 0, InGame: game, Avatar: f.AvatarURL()}
}

func (b backend) SelfID() uint64 { return b.c.SelfID() }
func (b backend) Friends() []chat.Friend {
	fs := b.c.Friends()
	out := make([]chat.Friend, len(fs))
	for i, f := range fs {
		out[i] = toChat(f)
	}
	return out
}
func (b backend) Lookup(id uint64) (chat.Friend, bool) {
	f, ok := b.c.Friend(id)
	return toChat(f), ok
}
func (b backend) Send(ctx context.Context, to uint64, text string) error {
	return b.c.SendMessage(ctx, to, text)
}
func (b backend) Ack(partner uint64, t time.Time) { b.c.Ack(partner, t) }

type notifier struct{ a *App }

func (n notifier) EntryAdded(e chat.Entry)              { n.a.emit(EvEntry, e) }
func (n notifier) EntryUpdated(e chat.Entry)            { n.a.emit(EvEntryUpdated, e) }
func (n notifier) LogCleared()                          { n.a.emit(EvLogCleared, nil) }
func (n notifier) ComposerChanged(s chat.ComposerState) { n.a.emit(EvComposer, s) }
