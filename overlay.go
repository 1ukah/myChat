package main

import (
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"mychat/internal/win"
)

// Overlay window sizes in CSS pixels. The UI is designed for exactly these
// sizes (see docs/UI_CONTRACT.md).
const (
	overlayWidth   = 640
	composerHeight = 360
	// The pop-up area is as tall as the chat box: the window is see-through
	// and click-through, so pop-ups stack up from the bottom inside it.
	toastHeight = composerHeight
)

// overlay controls the chat window at the Win32 level. It never activates:
// the game keeps keyboard focus, and clicks pass through to the game.
type overlay struct {
	w    *application.WebviewWindow
	hwnd uintptr

	mu sync.Mutex
	// home is where the 640×360 area is on screen (physical pixels). While
	// a picture is open the window covers the whole monitor, and the UI
	// keeps drawing the chat box or pop-ups at this same place.
	home win.RECT
	// big is set while the window covers the monitor.
	big bool
	// concealed is set while the window is invisible during a resize (see
	// concealLocked); revealTimer brings it back if the UI never says it
	// has redrawn.
	concealed   bool
	revealTimer *time.Timer
}

// revealWait is the longest the window stays invisible after a resize.
const revealWait = 300 * time.Millisecond

// concealLocked makes the window invisible for a resize between the 640×360
// size and the whole monitor. WebView2 redraws a moment after the window
// changes size, and until then shows its last frame at the window's top
// left corner: the chat box would jump to the corner of the monitor for a
// frame or two. The UI calls reveal once it has drawn at the new size.
func (o *overlay) concealLocked() {
	if !win.IsWindowVisible(o.hwnd) {
		return
	}
	o.concealed = true
	if o.revealTimer != nil {
		o.revealTimer.Stop()
	}
	o.revealTimer = time.AfterFunc(revealWait, func() { application.InvokeAsync(o.reveal) })
	win.LayeredInvisible(o.hwnd)
}

// reveal makes the window visible again after concealLocked.
func (o *overlay) reveal() {
	o.mu.Lock()
	was := o.concealed
	o.concealed = false
	if o.revealTimer != nil {
		o.revealTimer.Stop()
		o.revealTimer = nil
	}
	o.mu.Unlock()
	if was {
		win.LayeredOpaque(o.hwnd)
	}
}

func (o *overlay) attach() {
	o.hwnd = uintptr(o.w.NativeWindow())
	// Wails creates a captioned window; make it a bare popup so Windows draws
	// no title bar or frame.
	win.SetStyle(o.hwnd, win.OverlayStyle)
	win.SetExStyle(o.hwnd, win.OverlayExStyle)
	win.LayeredOpaque(o.hwnd)
	win.SquareCorners(o.hwnd)
	win.NoBorder(o.hwnd)
	win.SetWindowPos(o.hwnd, 0, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOACTIVATE|win.SWP_FRAMECHANGED)
}

// show places the window at the bottom middle of the monitor that holds the
// active window (the game) and shows it without taking focus.
func (o *overlay) show(cssW, cssH, bottomOffset int, scale float64) {
	mon, dpi := win.MonitorOf(win.ForegroundWindow())
	k := dpi * scale
	w := int32(float64(cssW) * k)
	h := int32(float64(cssH) * k)
	x := mon.Left + (mon.Right-mon.Left-w)/2
	y := mon.Bottom - int32(float64(bottomOffset)*dpi) - h
	if y < mon.Top {
		y = mon.Top
	}
	o.mu.Lock()
	o.home = win.RECT{Left: x, Top: y, Right: x + w, Bottom: y + h}
	if o.big {
		o.big = false
		o.concealLocked()
	}
	o.mu.Unlock()
	win.SetWindowPos(o.hwnd, win.HWND_TOPMOST, x, y, w, h, win.SWP_NOACTIVATE|win.SWP_SHOWWINDOW|win.SWP_NOOWNERZORDER)
}

// showViewer makes the window cover the monitor it is on, for a picture
// opened at full size. The window stays click-through.
func (o *overlay) showViewer() {
	o.mu.Lock()
	if !o.big {
		o.big = true
		o.concealLocked()
	}
	o.mu.Unlock()
	mon, _ := win.MonitorOf(o.hwnd)
	win.SetWindowPos(o.hwnd, win.HWND_TOPMOST, mon.Left, mon.Top, mon.Right-mon.Left, mon.Bottom-mon.Top,
		win.SWP_NOACTIVATE|win.SWP_SHOWWINDOW|win.SWP_NOOWNERZORDER)
}

// screenRect turns a box in CSS pixels of the 640×360 area into screen
// pixels.
func (o *overlay) screenRect(x, y, w, h float64) win.RECT {
	o.mu.Lock()
	home := o.home
	o.mu.Unlock()
	kx := float64(home.Right-home.Left) / overlayWidth
	ky := float64(home.Bottom-home.Top) / composerHeight
	return win.RECT{
		Left:   home.Left + int32(x*kx),
		Top:    home.Top + int32(y*ky),
		Right:  home.Left + int32((x+w)*kx),
		Bottom: home.Top + int32((y+h)*ky),
	}
}

func (o *overlay) hide() {
	win.ShowWindow(o.hwnd, win.SW_HIDE)
}
