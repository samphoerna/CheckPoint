package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/src
//go:embed all:frontend/src
var assets embed.FS

// AppVersion is injected at build time via -ldflags "-X main.AppVersion=vX.Y.Z"
var AppVersion = "v0.1.2"

func main() {
	// Create an instance of the app structure
	app := NewApp(AppVersion)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "CheckPoint",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "Mac Diagnostic Tool",
				Message: "A system diagnostic tool for macOS",
			},
		},
		LogLevel: logger.ERROR,
	})

	if err != nil {
		log.Fatal("Error:", err)
	}
}
