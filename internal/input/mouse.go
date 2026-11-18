//go:build windows

package input

import (
	"log"
	"unsafe"

	"mychat/internal/win"

	"golang.org/x/sys/windows"
)

type Spot struct {
	Rect win.RECT
	ID   string
}

type Mouse struct {
	ID    string
	Hover bool
}

type mouseState struct {
	spots  []Spot
	viewer bool
}

func (m *Manager) SetMouse(spots []Spot, viewer bool) {
	m.mouse.Store(&mouseState{spots: spots, viewer: viewer})
	m.RefreshMouse()
}

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

var (
	mouseHandle uintptr
	mouseCB     = windows.NewCallback(mouseProc)
	ms          struct {
		m      *Manager
		upL    bool
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
