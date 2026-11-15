//go:build windows

// Package win holds the small set of Win32 calls the app needs. Nothing here
// opens, reads or writes another process: only window handles and input.
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
	pSetWindowCompositionAttr = user32.NewProc("SetWindowCompositionAttribute")
	pSetLayeredWindowAttrs    = user32.NewProc("SetLayeredWindowAttributes")
	pDwmSetWindowAttribute    = dwmapi.NewProc("DwmSetWindowAttribute")
	pGetDpiForMonitor         = shcore.NewProc("GetDpiForMonitor")
	pGetCurrentThreadId       = kernel3.NewProc("GetCurrentThreadId")
)

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	WM_SYSKEYUP    = 0x0105
	WM_HOTKEY      = 0x0312
	WM_APP         = 0x8000
	WM_QUIT        = 0x0012

	MOD_ALT      = 0x1
	MOD_CONTROL  = 0x2
	MOD_SHIFT    = 0x4
	MOD_NOREPEAT = 0x4000

	LLKHF_EXTENDED = 0x01
	LLKHF_INJECTED = 0x10
	LLKHF_UP       = 0x80

	GWL_EXSTYLE = -20

	WS_EX_TOPMOST       = 0x00000008
	WS_EX_TRANSPARENT   = 0x00000020
	WS_EX_TOOLWINDOW    = 0x00000080
	WS_EX_LAYERED       = 0x00080000
	WS_EX_NOACTIVATE    = 0x08000000
	WS_EX_APPWINDOW     = 0x00040000
	WS_EX_CONTROLPARENT = 0x00010000

	// OverlayExStyle never activates, stays on top, has no taskbar button
	// and lets mouse clicks through.
	OverlayExStyle = WS_EX_CONTROLPARENT | WS_EX_LAYERED | WS_EX_TRANSPARENT | WS_EX_TOPMOST | WS_EX_TOOLWINDOW | WS_EX_NOACTIVATE

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

var HWND_TOPMOST = ^uintptr(0) // (HWND)-1

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

type RECT struct{ Left, Top, Right, Bottom int32 }

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

func SetWindowsHookEx(idHook int, fn uintptr, mod uintptr, threadID uint32) uintptr {
	r, _, _ := pSetWindowsHookExW.Call(uintptr(idHook), fn, mod, uintptr(threadID))
	return r
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

// ToUnicodeEx translates a key with the given keyboard state and layout.
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

// ForegroundWindow returns the active window handle. This is only a window
// handle, not a process handle.
func ForegroundWindow() uintptr {
	r, _, _ := pGetForegroundWindow.Call()
	return r
}

// WindowThreadID returns the GUI thread ID that owns hwnd.
func WindowThreadID(hwnd uintptr) uint32 {
	r, _, _ := pGetWindowThreadProcessId.Call(hwnd, 0)
	return uint32(r)
}

// ForegroundLayout returns the keyboard layout of the active window.
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

// LayeredOpaque makes a WS_EX_LAYERED window draw normally (alpha 255).
func LayeredOpaque(hwnd uintptr) {
	pSetLayeredWindowAttrs.Call(hwnd, 0, 255, 2) // LWA_ALPHA
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

func SetExStyle(hwnd, style uintptr) {
	var idx int32 = GWL_EXSTYLE
	pSetWindowLongPtrW.Call(hwnd, uintptr(idx), style)
}

// MonitorOf returns the monitor rectangle (physical pixels) and DPI scale of
// the monitor that holds hwnd.
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

// DWM attributes.
const (
	dwmwaWindowCornerPreference = 33
	dwmwaSystemBackdropType     = 38
	dwmwaUseImmersiveDarkMode   = 20
)

func dwmSet(hwnd uintptr, attr uint32, val uint32) {
	pDwmSetWindowAttribute.Call(hwnd, uintptr(attr), uintptr(unsafe.Pointer(&val)), 4)
}

// SquareCorners turns off Windows 11 rounded corners.
func SquareCorners(hwnd uintptr) { dwmSet(hwnd, dwmwaWindowCornerPreference, 1) }

// DarkFrame asks DWM for dark window chrome.
func DarkFrame(hwnd uintptr) { dwmSet(hwnd, dwmwaUseImmersiveDarkMode, 1) }

// SystemBackdrop sets the Windows 11 system backdrop (1 none, 3 acrylic).
func SystemBackdrop(hwnd uintptr, kind uint32) { dwmSet(hwnd, dwmwaSystemBackdropType, kind) }

type accentPolicy struct {
	AccentState   uint32
	AccentFlags   uint32
	GradientColor uint32
	AnimationID   uint32
}

type winCompAttrData struct {
	Attrib uint32
	PvData unsafe.Pointer
	CbData uintptr
}

// AccentBlur enables or disables DWM blur-behind for a window through the
// accent API. Unlike the Windows 11 system backdrop it also works while the
// window is inactive, which this overlay always is. tint is 0xAABBGGRR.
func AccentBlur(hwnd uintptr, on bool, tint uint32) bool {
	if pSetWindowCompositionAttr.Find() != nil {
		return false
	}
	ap := accentPolicy{}
	if on {
		ap.AccentState = 3 // ACCENT_ENABLE_BLURBEHIND
		ap.AccentFlags = 2
		ap.GradientColor = tint
	}
	data := winCompAttrData{Attrib: 19, PvData: unsafe.Pointer(&ap), CbData: unsafe.Sizeof(ap)} // WCA_ACCENT_POLICY
	r, _, _ := pSetWindowCompositionAttr.Call(hwnd, uintptr(unsafe.Pointer(&data)))
	return r != 0
}
