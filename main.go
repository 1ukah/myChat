package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"mychat/internal/config"
	"mychat/internal/win"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	store, err := config.Open()
	if err != nil {
		log.Fatal(err)
	}
	if f, err := os.OpenFile(filepath.Join(store.Dir(), "myChat.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600); err == nil {
		log.SetOutput(f)
		defer f.Close()
	}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %v\n%s", r, debug.Stack())
			panic(r)
		}
	}()
	log.Print("starting")
	a := newApp(store)

	app := application.New(application.Options{
		Name:        "myChat",
		Description: "Steam chat overlay for Counter-Strike 2",
		Services: []application.Service{
			application.NewService(&ChatService{a: a}),
		},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: a.media.Middleware,
		},
		Windows: application.WindowsOptions{
			DisableQuitOnLastWindowClosed: true,
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "com.mychat.cs2overlay",
			OnSecondInstanceLaunch: func(application.SecondInstanceData) {
				a.showSettings()
			},
		},
		OnShutdown: func() {
			a.steam.Stop()
			if a.inp != nil {
				a.inp.Close()
			}
		},
	})
	a.app = app

	ow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:                       "overlay",
		Title:                      "myChat overlay",
		URL:                        "/?view=overlay",
		Width:                      overlayWidth,
		Height:                     composerHeight,
		Frameless:                  true,
		AlwaysOnTop:                true,
		Hidden:                     true,
		DisableResize:              true,
		IgnoreMouseEvents:          true,
		BackgroundType:             application.BackgroundTypeTransparent,
		BackgroundColour:           application.NewRGBA(0, 0, 0, 0),
		Zoom:                       a.cfg.Scale,
		DefaultContextMenuDisabled: true,
		Windows: application.WindowsWindow{
			HiddenOnTaskbar:                   true,
			DisableFramelessWindowDecorations: true,
			DisableIcon:                       true,
			ExStyle:                           win.OverlayExStyle,
		},
	})
	a.ov.w = ow

	sw := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:                       "settings",
		Title:                      "myChat",
		URL:                        "/?view=settings",
		Width:                      980,
		Height:                     680,
		MinWidth:                   820,
		MinHeight:                  560,
		Frameless:                  true,
		Hidden:                     true,
		BackgroundColour:           application.NewRGB(20, 22, 25),
		DefaultContextMenuDisabled: true,
		Windows: application.WindowsWindow{
			Theme: application.Dark,
		},
	})
	sw.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		sw.Hide()
		e.Cancel()
	})
	sw.OnWindowEvent(events.Common.WindowRuntimeReady, func(*application.WindowEvent) { a.settingsLoaded() })
	a.settings = sw

	tray := app.SystemTray.New()
	tray.SetIcon(trayIcon())
	tray.SetTooltip("myChat")
	menu := application.NewMenu()
	menu.Add("Settings").OnClick(func(*application.Context) { a.showSettings() })
	menu.Add("Open chat").OnClick(func(*application.Context) { a.openComposer() })
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(*application.Context) { app.Quit() })
	tray.SetMenu(menu)
	tray.OnClick(a.showSettings)

	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		go a.start()
	})

	if err := app.Run(); err != nil {
		log.Printf("run: %v", err)
	}
	log.Print("exit")
}
