package main

import (
	"log"
	"os"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"mychat/internal/chat"
)

// demo shows the overlay with sample content for visual checks, without a
// Steam login and without catching keys. Enabled with MYCHAT_DEMO=toast or
// MYCHAT_DEMO=composer; it stays on screen until the app quits.
func (a *App) demo() {
	kind := os.Getenv("MYCHAT_DEMO")
	if kind == "" {
		return
	}
	log.Printf("demo: %s", kind)
	now := time.Now().UnixMilli()
	sample := chat.Entry{ID: -now, Time: now, Kind: chat.KindIn, PeerName: "Mestre Bene", Text: "bora mais uma? to no lobby"}
	switch kind {
	case "toast":
		a.mu.Lock()
		a.cfg.ToastSeconds = 3600
		a.mu.Unlock()
		a.toast(sample)
	case "composer":
		a.engine.System("Demo: chat box without key capture.")
		a.engine.Incoming(1, sample.Text, time.Now(), false)
		a.mu.Lock()
		a.stopTimersLocked()
		a.mode = ModeComposer
		cfg := a.cfg
		info := a.modeInfoLocked()
		a.mu.Unlock()
		application.InvokeSync(func() {
			a.ov.show(overlayWidth, composerHeight, cfg.BottomOffset, cfg.Scale)
		})
		a.emit(EvMode, info)
	}
}
