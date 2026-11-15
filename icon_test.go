package main

import (
	"os"
	"testing"
)

// TestWriteAppIcon regenerates build/appicon.png:
// MYCHAT_ICON=1 go test -run WriteAppIcon .
func TestWriteAppIcon(t *testing.T) {
	if os.Getenv("MYCHAT_ICON") == "" {
		t.Skip("set MYCHAT_ICON=1")
	}
	if err := os.WriteFile("build/appicon.png", drawIcon(1024, true), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.TempDir()+"/mychat-tray.png", drawIcon(64, false), 0o644); err != nil {
		t.Fatal(err)
	}
}
