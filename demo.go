package main

import (
	"log"
	"os"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"mychat/internal/chat"
	"mychat/internal/win"
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
	case "toasts":
		a.mu.Lock()
		a.cfg.ToastSeconds = 3600
		a.mu.Unlock()
		msgs := []struct{ id, name, text string }{
			{"1", "Mestre Bene", "bora mais uma?"},
			{"1", "Mestre Bene", "to no lobby"},
			{"2", "mira", "eu vou tb"},
			{"1", "Mestre Bene", "chama o resto"},
		}
		for i, m := range msgs {
			t := time.Now().UnixMilli()
			a.toast(chat.Entry{ID: -t - int64(i), Time: t, Kind: chat.KindIn, PeerID: m.id, PeerName: m.name, Text: m.text})
			time.Sleep(700 * time.Millisecond)
		}
	case "typing":
		// Real key capture: open, inject H and I, report what arrived.
		a.openComposer()
		time.Sleep(400 * time.Millisecond)
		for _, vk := range []byte{'H', 'I'} {
			win.KeyEvent(vk, false)
			win.KeyEvent(vk, true)
			time.Sleep(60 * time.Millisecond)
		}
		time.Sleep(500 * time.Millisecond)
		log.Printf("demo typing: composer text=%q", a.engine.Composer().Text)
		time.Sleep(4 * time.Second)
		a.closeComposer("demo")
	case "composer":
		a.engine.System("Demo: chat box without key capture.")
		a.engine.Incoming(1, sample.Text, nil, time.Now(), false)
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
