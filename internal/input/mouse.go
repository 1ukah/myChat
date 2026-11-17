//go:build windows

package input

import (
	"log"
	"unsafe"

	"mychat/internal/win"

	"golang.org/x/sys/windows"
)

// The overlay is click-through: the game gets every click. Pictures and
// names in it can still be clicked, through a low-level mouse hook that
// takes a click for itself only when it lands on one of them (or anywhere
// while a picture is open in the viewer) and the mouse pointer is visible.
// While the game aims with a hidden pointer, clicks are never taken. The
// hook also reports which spot the pointer is over, so the overlay can
// show it as a link. It is only installed while it has something to do.

// Spot is something on screen that reacts to the mouse.
type Spot struct {
	Rect win.RECT // screen pixels
	ID   string
}

// Mouse is a click the overlay took for itself, or a change of the spot
// under the pointer.
type Mouse struct {
	// ID is the spot clicked or now under the pointer. A click with ""
	// is a click while the viewer is open; a hover with "" means none.
	ID    string
	Hover bool
}

type mouseState struct {
	spots  []Spot
	viewer bool
}

// SetMouse sets the spots and whether the viewer is open.
func (m *Manager) SetMouse(spots []Spot, viewer bool) {
	m.mouse.Store(&mouseState{spots: spots, viewer: viewer})
	m.RefreshMouse()
}

// RefreshMouse installs or removes the hook as needed. Spots only react
// while the pointer is visible, so it is checked again from time to time;
// the open viewer needs the hook either way (a click closes it).
func (m *Manager) RefreshMouse() {
	s := m.mouse.Load()
	want := s != nil && (s.viewer || (len(s.spots) > 0 && win.CursorVisible()))
	m.mu.Lock()
	if m.mouseOn == want || m.threadID == 0 {
		m.mu.Unlock()
		return
	}
	m.mouseOn = want
	m.mu.Unlock()
	msg := uint32(msgMouseOff)
	if want {
		msg = msgMouseOn
	}
	win.PostThreadMessage(m.threadID, msg, 0, 0)
}

// Mouse hook state. It is only touched on the input thread.
var (
	mouseHandle uintptr
	mouseCB     = windows.NewCallback(mouseProc)
	ms          struct {
		m      *Manager
		upL    bool // the press was taken: take its release too
		upR    bool
		hover  string
		events chan Mouse
	}
)

func startMouse(m *Manager) {
	if mouseHandle != 0 {
		return
	}
	ms.m, ms.events = m, m.mouseEvents
	h, err := win.SetWindowsHookEx(win.WH_MOUSE_LL, mouseCB, win.ModuleHandle(), 0)
	if err != nil {
		log.Printf("mouse hook: %v", err)
		return
	}
	mouseHandle = h
}

func stopMouse() {
	if mouseHandle != 0 {
		win.UnhookWindowsHookEx(mouseHandle)
		mouseHandle = 0
	}
	ms.upL, ms.upR = false, false
	setHover("")
}

func sendMouse(e Mouse) {
	select {
	case ms.events <- e:
	default:
	}
}

func setHover(id string) {
	if ms.hover == id {
		return
	}
	ms.hover = id
	sendMouse(Mouse{ID: id, Hover: true})
}

func spotAt(s *mouseState, pt win.POINT) string {
	if s == nil || s.viewer {
		return ""
	}
	for _, sp := range s.spots {
		if sp.Rect.Contains(pt) {
			return sp.ID
		}
	}
	return ""
}

// mouseProc must return fast: it sees every mouse move while installed.
func mouseProc(nCode, wParam, lParam uintptr) uintptr {
	if int32(nCode) < 0 || mouseHandle == 0 {
		return win.CallNextHookEx(int(int32(nCode)), wParam, lParam)
	}
	switch wParam {
	case win.WM_MOUSEMOVE:
		pt := (*win.MSLLHOOKSTRUCT)(unsafe.Pointer(lParam)).Pt
		setHover(spotAt(ms.m.mouse.Load(), pt))
	case win.WM_LBUTTONDOWN, win.WM_RBUTTONDOWN:
		s := ms.m.mouse.Load()
		if s == nil {
			break
		}
		left := wParam == win.WM_LBUTTONDOWN
		if s.viewer {
			// Any click closes the viewer. With the pointer hidden the
			// player is aiming: the click still goes to the game.
			sendMouse(Mouse{})
			if !win.CursorVisible() {
				break
			}
			if left {
				ms.upL = true
			} else {
				ms.upR = true
			}
			return 1
		}
		if !left {
			break
		}
		pt := (*win.MSLLHOOKSTRUCT)(unsafe.Pointer(lParam)).Pt
		if id := spotAt(s, pt); id != "" && win.CursorVisible() {
			sendMouse(Mouse{ID: id})
			ms.upL = true
			return 1
		}
	case win.WM_LBUTTONUP:
		if ms.upL {
			ms.upL = false
			return 1
		}
	case win.WM_RBUTTONUP:
		if ms.upR {
			ms.upR = false
			return 1
		}
	}
	return win.CallNextHookEx(int(int32(nCode)), wParam, lParam)
}
