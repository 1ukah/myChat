package main

import _ "embed"

// trayPNG is the tray icon, rendered from frontend/src/assets/mychat-icon.svg
// (as is build/appicon.png, which becomes the Windows .ico).
//
//go:embed build/trayicon.png
var trayPNG []byte

func trayIcon() []byte { return trayPNG }
