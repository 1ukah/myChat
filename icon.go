package main

import _ "embed"

// trayICO is the tray icon: every size Windows asks for (16 to 64 px),
// rendered from the icon's SVGs by build/icon/render.mjs, as is
// build/windows/icon.ico. Windows picks the size that matches the display
// scale instead of shrinking one big picture.
//
//go:embed build/trayicon.ico
var trayICO []byte

func trayIcon() []byte { return trayICO }
