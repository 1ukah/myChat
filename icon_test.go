package main

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"testing"
)

func TestTrayIconHasEverySmallSize(t *testing.T) {
	ico := trayIcon()
	if len(ico) < 6 || binary.LittleEndian.Uint16(ico[0:]) != 0 || binary.LittleEndian.Uint16(ico[2:]) != 1 {
		t.Fatal("tray icon is not an ICO file")
	}
	count := int(binary.LittleEndian.Uint16(ico[4:]))
	sizes := map[int]bool{}
	for i := 0; i < count; i++ {
		entry := ico[6+16*i:]
		w, h := int(entry[0]), int(entry[1])
		if w == 0 {
			w = 256
		}
		if h == 0 {
			h = 256
		}
		size := binary.LittleEndian.Uint32(entry[8:])
		offset := binary.LittleEndian.Uint32(entry[12:])
		cfg, err := png.DecodeConfig(bytes.NewReader(ico[offset : offset+size]))
		if err != nil {
			t.Fatalf("entry %d: %v", i, err)
		}
		if cfg.Width != w || cfg.Height != h {
			t.Fatalf("entry %d says %dx%d but holds %dx%d", i, w, h, cfg.Width, cfg.Height)
		}
		sizes[w] = true
	}
	for _, want := range []int{16, 20, 24, 32} {
		if !sizes[want] {
			t.Errorf("tray icon has no %d px image", want)
		}
	}
}
