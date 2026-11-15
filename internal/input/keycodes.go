//go:build windows

package input

import (
	"fmt"
	"strings"

	"mychat/internal/win"
)

// Scan codes (set 1) for KeyboardEvent.code values whose virtual key depends
// on the keyboard layout. MapVirtualKeyEx turns them into the right VK.
var codeScan = map[string]uint32{
	"Backquote": 0x29, "Minus": 0x0C, "Equal": 0x0D, "BracketLeft": 0x1A, "BracketRight": 0x1B,
	"Backslash": 0x2B, "Semicolon": 0x27, "Quote": 0x28, "Comma": 0x33, "Period": 0x34,
	"Slash": 0x35, "IntlBackslash": 0x56, "IntlRo": 0x73,
}

// Layout-independent virtual keys.
var codeVK = map[string]uint32{
	"Enter": 0x0D, "Tab": 0x09, "Space": 0x20, "Backspace": 0x08, "Escape": 0x1B,
	"Insert": 0x2D, "Delete": 0x2E, "Home": 0x24, "End": 0x23, "PageUp": 0x21, "PageDown": 0x22,
	"ArrowUp": 0x26, "ArrowDown": 0x28, "ArrowLeft": 0x25, "ArrowRight": 0x27,
	"Pause": 0x13, "ScrollLock": 0x91, "CapsLock": 0x14,
	"Numpad0": 0x60, "Numpad1": 0x61, "Numpad2": 0x62, "Numpad3": 0x63, "Numpad4": 0x64,
	"Numpad5": 0x65, "Numpad6": 0x66, "Numpad7": 0x67, "Numpad8": 0x68, "Numpad9": 0x69,
	"NumpadMultiply": 0x6A, "NumpadAdd": 0x6B, "NumpadSubtract": 0x6D, "NumpadDecimal": 0x6E,
	"NumpadDivide": 0x6F, "NumpadEnter": 0x0D,
}

// CodeToVK converts a KeyboardEvent.code to a Windows virtual key code.
func CodeToVK(code string) (uint32, error) {
	switch {
	case strings.HasPrefix(code, "Key") && len(code) == 4:
		return uint32(code[3]), nil
	case strings.HasPrefix(code, "Digit") && len(code) == 6:
		return uint32(code[5]), nil
	case strings.HasPrefix(code, "F") && len(code) <= 3:
		var n int
		if _, err := fmt.Sscanf(code[1:], "%d", &n); err == nil && n >= 1 && n <= 24 {
			return uint32(0x70 + n - 1), nil
		}
	}
	if vk, ok := codeVK[code]; ok {
		return vk, nil
	}
	if sc, ok := codeScan[code]; ok {
		if vk := win.MapVirtualKeyEx(sc, 3, win.KeyboardLayout(0)); vk != 0 {
			return vk, nil
		}
	}
	return 0, fmt.Errorf("unsupported key %q", code)
}

// CodeLabel returns a short display name for a KeyboardEvent.code.
func CodeLabel(code string) string {
	switch {
	case strings.HasPrefix(code, "Key") && len(code) == 4:
		return code[3:]
	case strings.HasPrefix(code, "Digit") && len(code) == 6:
		return code[5:]
	case strings.HasPrefix(code, "Numpad"):
		return "Num " + strings.TrimPrefix(code, "Numpad")
	case strings.HasPrefix(code, "Arrow"):
		return strings.TrimPrefix(code, "Arrow")
	}
	switch code {
	case "Backquote":
		return "`"
	case "Minus":
		return "-"
	case "Equal":
		return "="
	case "BracketLeft":
		return "["
	case "BracketRight":
		return "]"
	case "Backslash", "IntlBackslash":
		return "\\"
	case "Semicolon":
		return ";"
	case "Quote":
		return "'"
	case "Comma":
		return ","
	case "Period":
		return "."
	case "Slash":
		return "/"
	}
	return code
}
