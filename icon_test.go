package main

import (
	"bytes"
	"image/png"
	"testing"
)

func TestTrayIconIsPNG(t *testing.T) {
	cfg, err := png.DecodeConfig(bytes.NewReader(trayIcon()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != 64 || cfg.Height != 64 {
		t.Fatalf("tray icon is %dx%d, want 64x64", cfg.Width, cfg.Height)
	}
}
