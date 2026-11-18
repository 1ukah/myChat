package main

import _ "embed"

//go:embed build/trayicon.ico
var trayICO []byte

func trayIcon() []byte { return trayICO }
