//go:build windows

package win

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32  = windows.NewLazySystemDLL("user32.dll")
	dwmapi  = windows.NewLazySystemDLL("dwmapi.dll")
	shcore  = windows.NewLazySystemDLL("shcore.dll")
	kernel3 = windows.NewLazySystemDLL("kernel32.dll")

	pSetWindowsHookExW        = user32.NewProc("SetWindowsHookExW")
	pUnhookWindowsHookEx      = user32.NewProc("UnhookWindowsHookEx")
	pCallNextHookEx           = user32.NewProc("CallNextHookEx")
	pGetMessageW              = user32.NewProc("GetMessageW")
	pPostThreadMessageW       = user32.NewProc("PostThreadMessageW")
	pRegisterHotKey           = user32.NewProc("RegisterHotKey")
	pUnregisterHotKey         = user32.NewProc("UnregisterHotKey")
	pGetAsyncKeyState         = user32.NewProc("GetAsyncKeyState")
	pGetKeyState              = user32.NewProc("GetKeyState")
	pToUnicodeEx              = user32.NewProc("ToUnicodeEx")
	pGetKeyboardLayout        = user32.NewProc("GetKeyboardLayout")
	pMapVirtualKeyExW         = user32.NewProc("MapVirtualKeyExW")
	pGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	pGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	pSetWindowPos             = user32.NewProc("SetWindowPos")
	pShowWindow               = user32.NewProc("ShowWindow")
	pGetWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	pSetWindowLongPtrW        = user32.NewProc("SetWindowLongPtrW")
	pMonitorFromWindow        = user32.NewProc("MonitorFromWindow")
	pGetMonitorInfoW          = user32.NewProc("GetMonitorInfoW")
	pIsWindowVisible          = user32.NewProc("IsWindowVisible")
	pSetLayeredWindowAttrs    = user32.NewProc("SetLayeredWindowAttributes")
	pDwmSetWindowAttribute    = dwmapi.NewProc("DwmSetWindowAttribute")
	pGetDpiForMonitor         = shcore.NewProc("GetDpiForMonitor")
	pGetCurrentThreadId       = kernel3.NewProc("GetCurrentThreadId")
	pGetModuleHandleW         = kernel3.NewProc("GetModuleHandleW")
	pKeybdEvent               = user32.NewProc("keybd_event")
	pGetWindowTextW           = user32.NewProc("GetWindowTextW")
	pGetCursorInfo            = user32.NewProc("GetCursorInfo")
)

const (
	WH_KEYBOARD_LL = 13
	WH_MOUSE_LL    = 14
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	WM_SYSKEYUP    = 0x0105
	WM_HOTKEY      = 0x0312
	WM_MOUSEMOVE   = 0x0200
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_RBUTTONDOWN = 0x0204
	WM_RBUTTONUP   = 0x0205
	WM_APP         = 0x8000
	WM_QUIT        = 0x0012

	MOD_ALT      = 0x1
	MOD_CONTROL  = 0x2
	MOD_SHIFT    = 0x4
	MOD_NOREPEAT = 0x4000

	LLKHF_EXTENDED = 0x01
	LLKHF_INJECTED = 0x10
	LLKHF_UP       = 0x80

	GWL_STYLE   = -16
	GWL_EXSTYLE = -20

	WS_POPUP        = 0x80000000
	WS_CLIPSIBLINGS = 0x04000000
	WS_CLIPCHILDREN = 0x02000000

	OverlayStyle = WS_POPUP | WS_CLIPSIBLINGS | WS_CLIPCHILDREN

	WS_EX_NOREDIRECTIONBITMAP = 0x00200000

	WS_EX_TOPMOST       = 0x00000008
	WS_EX_TRANSPARENT   = 0x00000020
	WS_EX_TOOLWINDOW    = 0x00000080
	WS_EX_LAYERED       = 0x00080000
	WS_EX_NOACTIVATE    = 0x08000000
	WS_EX_APPWINDOW     = 0x00040000
	WS_EX_CONTROLPARENT = 0x00010000

	OverlayExStyle = WS_EX_CONTROLPARENT | WS_EX_NOREDIRECTIONBITMAP | WS_EX_LAYERED | WS_EX_TRANSPARENT | WS_EX_TOPMOST | WS_EX_TOOLWINDOW | WS_EX_NOACTIVATE

	SW_HIDE           = 0
	SW_SHOWNOACTIVATE = 4

	SWP_NOSIZE        = 0x0001
	SWP_NOMOVE        = 0x0002
	SWP_NOACTIVATE    = 0x0010
	SWP_FRAMECHANGED  = 0x0020
	SWP_SHOWWINDOW    = 0x0040
	SWP_NOOWNERZORDER = 0x0200

	MONITOR_DEFAULTTOPRIMARY = 1
)

var HWND_TOPMOST = ^uintptr(0)

type POINT struct{ X, Y int32 }

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type MSLLHOOKSTRUCT struct {
	Pt          POINT
	MouseData   uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type RECT struct{ Left, Top, Right, Bottom int32 }

func (r RECT) Contains(p POINT) bool {
	return p.X >= r.Left && p.X < r.Right && p.Y >= r.Top && p.Y < r.Bottom
}

type cursorInfo struct {
	CbSize  uint32
	Flags   uint32
	HCursor uintptr
	Pt      POINT
}

func CursorVisible() bool {
	ci := cursorInfo{CbSize: uint32(unsafe.Sizeof(cursorInfo{}))}
	if r, _, _ := pGetCursorInfo.Call(uintptr(unsafe.Pointer(&ci))); r == 0 {
		return false
	}
	return ci.Flags&1 != 0 && ci.HCursor != 0
}

type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}

func CurrentThreadID() uint32 {
	r, _, _ := pGetCurrentThreadId.Call()
	return uint32(r)
}

func SetWindowsHookEx(idHook int, fn uintptr, mod uintptr, threadID uint32) (uintptr, error) {
	r, _, err := pSetWindowsHookExW.Call(uintptr(idHook), fn, mod, uintptr(threadID))
	if r == 0 {
		return 0, err
	}
	return r, nil
}

func WindowTitle(hwnd uintptr) string {
	var buf [256]uint16
	n, _, _ := pGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return windows.UTF16ToString(buf[:n])
}

func ModuleHandle() uintptr {
	r, _, _ := pGetModuleHandleW.Call(0)
	return r
}

func KeyEvent(vk byte, up bool) {
	var flags uintptr
	if up {
		flags = 2
	}
	pKeybdEvent.Call(uintptr(vk), 0, flags, 0)
}

func UnhookWindowsHookEx(h uintptr) bool {
	r, _, _ := pUnhookWindowsHookEx.Call(h)
	return r != 0
}

func CallNextHookEx(nCode int, wParam, lParam uintptr) uintptr {
	r, _, _ := pCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return r
}

func GetMessage(msg *MSG) int32 {
	r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0)
	return int32(r)
}

func PostThreadMessage(threadID uint32, msg uint32, wParam, lParam uintptr) bool {
	r, _, _ := pPostThreadMessageW.Call(uintptr(threadID), uintptr(msg), wParam, lParam)
	return r != 0
}

func RegisterHotKey(id int, mods, vk uint32) error {
	r, _, err := pRegisterHotKey.Call(0, uintptr(id), uintptr(mods), uintptr(vk))
	if r == 0 {
		return err
	}
	return nil
}

func UnregisterHotKey(id int) {
	pUnregisterHotKey.Call(0, uintptr(id))
}

func AsyncKeyDown(vk int) bool {
	r, _, _ := pGetAsyncKeyState.Call(uintptr(vk))
	return r&0x8000 != 0
}

func KeyToggled(vk int) bool {
	r, _, _ := pGetKeyState.Call(uintptr(vk))
	return r&1 != 0
}

func ToUnicodeEx(vk, scan uint32, state *[256]byte, buf []uint16, hkl uintptr) int32 {
	r, _, _ := pToUnicodeEx.Call(uintptr(vk), uintptr(scan), uintptr(unsafe.Pointer(state)),
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0, hkl)
	return int32(r)
}

func KeyboardLayout(threadID uint32) uintptr {
	r, _, _ := pGetKeyboardLayout.Call(uintptr(threadID))
	return r
}

func MapVirtualKeyEx(code, mapType uint32, hkl uintptr) uint32 {
	r, _, _ := pMapVirtualKeyExW.Call(uintptr(code), uintptr(mapType), hkl)
	return uint32(r)
}

func ForegroundWindow() uintptr {
	r, _, _ := pGetForegroundWindow.Call()
	return r
}

func WindowThreadID(hwnd uintptr) uint32 {
	r, _, _ := pGetWindowThreadProcessId.Call(hwnd, 0)
	return uint32(r)
}

func ForegroundLayout() uintptr {
	return KeyboardLayout(WindowThreadID(ForegroundWindow()))
}

func SetWindowPos(hwnd, after uintptr, x, y, w, h int32, flags uint32) bool {
	r, _, _ := pSetWindowPos.Call(hwnd, after, uintptr(x), uintptr(y), uintptr(w), uintptr(h), uintptr(flags))
	return r != 0
}

func ShowWindow(hwnd uintptr, cmd int) {
	pShowWindow.Call(hwnd, uintptr(cmd))
}

func LayeredOpaque(hwnd uintptr) {
	pSetLayeredWindowAttrs.Call(hwnd, 0, 255, 2)
}

func LayeredInvisible(hwnd uintptr) {
	pSetLayeredWindowAttrs.Call(hwnd, 0, 0, 2)
}

func IsWindowVisible(hwnd uintptr) bool {
	r, _, _ := pIsWindowVisible.Call(hwnd)
	return r != 0
}

func GetExStyle(hwnd uintptr) uintptr {
	var idx int32 = GWL_EXSTYLE
	r, _, _ := pGetWindowLongPtrW.Call(hwnd, uintptr(idx))
	return r
}

func SetStyle(hwnd, style uintptr) {
	var idx int32 = GWL_STYLE
	pSetWindowLongPtrW.Call(hwnd, uintptr(idx), style)
}

func SetExStyle(hwnd, style uintptr) {
	var idx int32 = GWL_EXSTYLE
	pSetWindowLongPtrW.Call(hwnd, uintptr(idx), style)
}

func MonitorOf(hwnd uintptr) (RECT, float64) {
	hm, _, _ := pMonitorFromWindow.Call(hwnd, MONITOR_DEFAULTTOPRIMARY)
	mi := MONITORINFO{CbSize: uint32(unsafe.Sizeof(MONITORINFO{}))}
	pGetMonitorInfoW.Call(hm, uintptr(unsafe.Pointer(&mi)))
	var dx, dy uint32
	scale := 1.0
	if pGetDpiForMonitor.Find() == nil {
		if r, _, _ := pGetDpiForMonitor.Call(hm, 0, uintptr(unsafe.Pointer(&dx)), uintptr(unsafe.Pointer(&dy))); r == 0 && dx > 0 {
			scale = float64(dx) / 96
		}
	}
	return mi.RcMonitor, scale
}

const (
	dwmwaWindowCornerPreference = 33
	dwmwaSystemBackdropType     = 38
	dwmwaUseImmersiveDarkMode   = 20
	dwmwaBorderColor            = 34
)

func dwmSet(hwnd uintptr, attr uint32, val uint32) {
	pDwmSetWindowAttribute.Call(hwnd, uintptr(attr), uintptr(unsafe.Pointer(&val)), 4)
}

func SquareCorners(hwnd uintptr) { dwmSet(hwnd, dwmwaWindowCornerPreference, 1) }

func NoBorder(hwnd uintptr) { dwmSet(hwnd, dwmwaBorderColor, 0xFFFFFFFE) }

func DarkFrame(hwnd uintptr) { dwmSet(hwnd, dwmwaUseImmersiveDarkMode, 1) }

func SystemBackdrop(hwnd uintptr, kind uint32) { dwmSet(hwnd, dwmwaSystemBackdropType, kind) }
