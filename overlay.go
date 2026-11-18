package main

import (
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"mychat/internal/win"
)

const (
	overlayWidth   = 640
	composerHeight = 360
	toastHeight    = composerHeight
)

type overlay struct {
	w    *application.WebviewWindow
	hwnd uintptr

	mu          sync.Mutex
	home        win.RECT
	big         bool
	concealed   bool
	revealTimer *time.Timer
}

const revealWait = 300 * time.Millisecond

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
	win.SetStyle(o.hwnd, win.OverlayStyle)
	win.SetExStyle(o.hwnd, win.OverlayExStyle)
	win.LayeredOpaque(o.hwnd)
	win.SquareCorners(o.hwnd)
	win.NoBorder(o.hwnd)
	win.SetWindowPos(o.hwnd, 0, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOACTIVATE|win.SWP_FRAMECHANGED)
}

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
	if o.big && win.IsWindowVisible(o.hwnd) {
		bigMon, _ := win.MonitorOf(o.hwnd)
		if bigMon == mon {
			o.home = win.RECT{Left: x, Top: y, Right: x + w, Bottom: y + h}
			o.mu.Unlock()
			return
		}
	}
	o.home = win.RECT{Left: x, Top: y, Right: x + w, Bottom: y + h}
	if o.big {
		o.big = false
		o.concealLocked()
	}
	o.mu.Unlock()
	win.SetWindowPos(o.hwnd, win.HWND_TOPMOST, x, y, w, h, win.SWP_NOACTIVATE|win.SWP_SHOWWINDOW|win.SWP_NOOWNERZORDER)
}

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
