package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"mychat/internal/win"
)

// Overlay window sizes in CSS pixels. The UI is designed for exactly these
// sizes (see docs/UI_CONTRACT.md).
const (
	overlayWidth   = 640
	composerHeight = 360
	toastHeight    = 92
)

// overlay controls the chat window at the Win32 level. It never activates:
// the game keeps keyboard focus, and clicks pass through to the game.
type overlay struct {
	w    *application.WebviewWindow
	hwnd uintptr
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
	win.SetWindowPos(o.hwnd, win.HWND_TOPMOST, x, y, w, h, win.SWP_NOACTIVATE|win.SWP_SHOWWINDOW|win.SWP_NOOWNERZORDER)
}

func (o *overlay) hide() {
	win.ShowWindow(o.hwnd, win.SW_HIDE)
}

func (o *overlay) setBlur(on bool) {
	win.AccentBlur(o.hwnd, on, 0x01000000)
}
