//go:build windows

// Package input owns the global hotkey and, only while the chat box is open,
// a low-level keyboard hook. The hook catches typed keys so the game never
// receives them, while keys that were already held when the box opened (for
// example W) keep going to the game. The game keeps focus the whole time.
package input

import (
	"errors"
	"runtime"
	"sync"
	"unicode/utf16"
	"unsafe"

	"mychat/internal/win"

	"golang.org/x/sys/windows"
)

// Key is a non-text editing key.
type Key string

const (
	KeyEnter         Key = "Enter"
	KeyEscape        Key = "Escape"
	KeyBackspace     Key = "Backspace"
	KeyWordBackspace Key = "WordBackspace"
	KeyDelete        Key = "Delete"
	KeyLeft          Key = "Left"
	KeyRight         Key = "Right"
	KeyWordLeft      Key = "WordLeft"
	KeyWordRight     Key = "WordRight"
	KeyHome          Key = "Home"
	KeyEnd           Key = "End"
	KeyUp            Key = "Up"
	KeyDown          Key = "Down"
	KeyPageUp        Key = "PageUp"
	KeyPageDown      Key = "PageDown"
	KeyTab           Key = "Tab"
	KeyPaste         Key = "Paste"
	// KeyFocusLost means the user switched away (Alt+Tab or Win key).
	KeyFocusLost Key = "FocusLost"
)

// Event is one typed character or editing key.
type Event struct {
	Text  string
	Key   Key
	Shift bool
}

// Hotkey is the global key that opens the chat box.
type Hotkey struct {
	VK    uint32
	Ctrl  bool
	Shift bool
	Alt   bool
}

const (
	msgCaptureOn  = win.WM_APP + 1
	msgCaptureOff = win.WM_APP + 2
	msgSetHotkey  = win.WM_APP + 3
	hotkeyID      = 0x4d43
)

// Manager runs the input thread.
type Manager struct {
	Hotkeys <-chan struct{}
	Events  <-chan Event

	hotkeys  chan struct{}
	events   chan Event
	threadID uint32

	mu       sync.Mutex
	pending  Hotkey
	reply    chan error
	capturing bool
}

var errNoThread = errors.New("input: thread not running")

// Start launches the input thread and registers the hotkey.
func Start(hk Hotkey) (*Manager, error) {
	m := &Manager{
		hotkeys: make(chan struct{}, 4),
		events:  make(chan Event, 256),
		reply:   make(chan error, 1),
	}
	m.Hotkeys, m.Events = m.hotkeys, m.events
	ready := make(chan error, 1)
	go m.loop(hk, ready)
	if err := <-ready; err != nil {
		return m, err
	}
	return m, nil
}

// SetHotkey changes the global hotkey.
func (m *Manager) SetHotkey(hk Hotkey) error {
	if m.threadID == 0 {
		return errNoThread
	}
	m.mu.Lock()
	m.pending = hk
	m.mu.Unlock()
	win.PostThreadMessage(m.threadID, msgSetHotkey, 0, 0)
	return <-m.reply
}

// Capture turns the keyboard hook on or off.
func (m *Manager) Capture(on bool) {
	m.mu.Lock()
	if m.capturing == on {
		m.mu.Unlock()
		return
	}
	m.capturing = on
	m.mu.Unlock()
	msg := uint32(msgCaptureOff)
	if on {
		msg = msgCaptureOn
	}
	win.PostThreadMessage(m.threadID, msg, 0, 0)
}

// Close stops the input thread.
func (m *Manager) Close() {
	if m.threadID != 0 {
		win.PostThreadMessage(m.threadID, win.WM_QUIT, 0, 0)
	}
}

func registerHotkey(hk Hotkey) error {
	if hk.VK == 0 {
		return errors.New("input: no hotkey")
	}
	mods := uint32(win.MOD_NOREPEAT)
	if hk.Ctrl {
		mods |= win.MOD_CONTROL
	}
	if hk.Shift {
		mods |= win.MOD_SHIFT
	}
	if hk.Alt {
		mods |= win.MOD_ALT
	}
	return win.RegisterHotKey(hotkeyID, mods, hk.VK)
}

func (m *Manager) loop(hk Hotkey, ready chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	m.threadID = win.CurrentThreadID()

	// Make sure the thread has a message queue before anyone posts to it.
	var msg win.MSG
	err := registerHotkey(hk)
	ready <- err
	current := hk

	defer stopHook()
	defer win.UnregisterHotKey(hotkeyID)
	for win.GetMessage(&msg) > 0 {
		switch msg.Message {
		case win.WM_HOTKEY:
			select {
			case m.hotkeys <- struct{}{}:
			default:
			}
		case msgCaptureOn:
			startHook(m.events)
		case msgCaptureOff:
			stopHook()
		case msgSetHotkey:
			m.mu.Lock()
			next := m.pending
			m.mu.Unlock()
			win.UnregisterHotKey(hotkeyID)
			if err := registerHotkey(next); err != nil {
				_ = registerHotkey(current)
				m.reply <- err
				continue
			}
			current = next
			m.reply <- nil
		}
	}
}

// Hook state. It is only touched on the input thread.
var (
	hookHandle uintptr
	hookCB     = windows.NewCallback(hookProc)
	st         struct {
		down   [256]bool
		pass   [256]bool
		caps   bool
		num    bool
		hkl    uintptr
		events chan Event
	}
)

const (
	vkBack     = 0x08
	vkTab      = 0x09
	vkReturn   = 0x0D
	vkShift    = 0x10
	vkControl  = 0x11
	vkMenu     = 0x12
	vkCapital  = 0x14
	vkEscape   = 0x1B
	vkPrior    = 0x21
	vkNext     = 0x22
	vkEnd      = 0x23
	vkHome     = 0x24
	vkLeft     = 0x25
	vkUp       = 0x26
	vkRight    = 0x27
	vkDown     = 0x28
	vkDelete   = 0x2E
	vkLWin     = 0x5B
	vkRWin     = 0x5C
	vkNumLock  = 0x90
	vkLShift   = 0xA0
	vkRShift   = 0xA1
	vkLControl = 0xA2
	vkRControl = 0xA3
	vkLMenu    = 0xA4
	vkRMenu    = 0xA5
)

func startHook(events chan Event) {
	if hookHandle != 0 {
		return
	}
	st.down = [256]bool{}
	st.pass = [256]bool{}
	// Keys held right now (movement, the hotkey itself) keep going to the
	// game, including their release.
	for vk := 8; vk < 255; vk++ {
		if win.AsyncKeyDown(vk) {
			st.pass[vk] = true
			st.down[vk] = true
		}
	}
	st.caps = win.KeyToggled(vkCapital)
	st.num = win.KeyToggled(vkNumLock)
	st.hkl = win.ForegroundLayout()
	st.events = events
	hookHandle = win.SetWindowsHookEx(win.WH_KEYBOARD_LL, hookCB, 0, 0)
}

func stopHook() {
	if hookHandle != 0 {
		win.UnhookWindowsHookEx(hookHandle)
		hookHandle = 0
	}
	st.pass = [256]bool{}
	st.down = [256]bool{}
}

func emit(e Event) {
	select {
	case st.events <- e:
	default:
	}
}

// hookProc must return fast: Windows removes slow low-level hooks.
func hookProc(nCode, wParam, lParam uintptr) uintptr {
	if int32(nCode) < 0 || hookHandle == 0 {
		return win.CallNextHookEx(int(int32(nCode)), wParam, lParam)
	}
	kb := (*win.KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
	vk := kb.VkCode & 0xff
	up := kb.Flags&win.LLKHF_UP != 0

	if st.pass[vk] {
		st.down[vk] = !up
		if up {
			st.pass[vk] = false
		}
		return win.CallNextHookEx(int(int32(nCode)), wParam, lParam)
	}
	st.down[vk] = !up

	switch vk {
	case vkLWin, vkRWin:
		if !up {
			emit(Event{Key: KeyFocusLost})
		}
		return win.CallNextHookEx(int(int32(nCode)), wParam, lParam)
	case vkLMenu:
		// Left Alt goes through so Alt+Tab keeps working.
		return win.CallNextHookEx(int(int32(nCode)), wParam, lParam)
	case vkTab:
		if st.down[vkLMenu] {
			if !up {
				emit(Event{Key: KeyFocusLost})
			}
			return win.CallNextHookEx(int(int32(nCode)), wParam, lParam)
		}
	}
	if !up {
		translate(vk, kb.ScanCode, kb.Flags)
	}
	return 1
}

func translate(vk, scan, flags uint32) {
	shift := st.down[vkLShift] || st.down[vkRShift] || st.down[vkShift]
	altGr := st.down[vkRMenu]
	ctrl := (st.down[vkLControl] || st.down[vkRControl] || st.down[vkControl]) && !altGr

	key := Key("")
	switch vk {
	case vkReturn:
		key = KeyEnter
	case vkEscape:
		key = KeyEscape
	case vkBack:
		key = KeyBackspace
		if ctrl {
			key = KeyWordBackspace
		}
	case vkDelete:
		key = KeyDelete
	case vkLeft:
		key = KeyLeft
		if ctrl {
			key = KeyWordLeft
		}
	case vkRight:
		key = KeyRight
		if ctrl {
			key = KeyWordRight
		}
	case vkHome:
		key = KeyHome
	case vkEnd:
		key = KeyEnd
	case vkUp:
		key = KeyUp
	case vkDown:
		key = KeyDown
	case vkPrior:
		key = KeyPageUp
	case vkNext:
		key = KeyPageDown
	case vkTab:
		key = KeyTab
	case vkCapital:
		st.caps = !st.caps
		return
	case vkNumLock:
		st.num = !st.num
		return
	case vkShift, vkLShift, vkRShift, vkControl, vkLControl, vkRControl, vkMenu, vkRMenu:
		return
	}
	if key != "" {
		emit(Event{Key: key, Shift: shift})
		return
	}
	if ctrl {
		if vk == 'V' {
			emit(Event{Key: KeyPaste})
		}
		return
	}

	var ks [256]byte
	if shift {
		ks[vkShift], ks[vkLShift] = 0x80, 0x80
	}
	if altGr {
		ks[vkControl], ks[vkLControl], ks[vkMenu], ks[vkRMenu] = 0x80, 0x80, 0x80, 0x80
	}
	if st.caps {
		ks[vkCapital] = 1
	}
	if st.num {
		ks[vkNumLock] = 1
	}
	var buf [8]uint16
	n := win.ToUnicodeEx(vk, scan, &ks, buf[:], st.hkl)
	if n <= 0 {
		return // dead key (stored for the next key) or nothing
	}
	runes := utf16.Decode(buf[:n])
	out := runes[:0]
	for _, r := range runes {
		if r >= 0x20 && r != 0x7f {
			out = append(out, r)
		}
	}
	if len(out) > 0 {
		emit(Event{Text: string(out)})
	}
}
