package main

import (
	"context"
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 创建日志文件
	logFile, err := os.OpenFile("go-claw.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Println("Failed to open log file:", err)
	} else {
		log.SetOutput(logFile)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		defer logFile.Close()
	}

	log.Println("===== GoClaw Starting =====")
	log.Println("Log file: go-claw.log")

	app := NewApp()
	toolsInstance := NewTools()
	tts := NewTTSManager()

	err = wails.Run(&options.App{
		Title:            "GoClaw",
		Width:            320,
		Height:           500,
		MinWidth:         320,
		MinHeight:        500,
		MaxWidth:         320,
		MaxHeight:        500,
		Frameless:        true,
		AlwaysOnTop:      true,
		DisableResize:    true,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		CSSDragProperty:  "--wails-draggable",
		CSSDragValue:     "true",
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			DisableFramelessWindowDecorations: true,
			BackdropType:                      windows.None,
		},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			toolsInstance.SetContext(ctx)
		},
		OnDomReady:    app.domReady,
		OnBeforeClose: app.beforeClose,
		OnShutdown:    app.shutdown,
		Bind: []interface{}{
			app,
			toolsInstance,
			tts,
		},
	})

	if err != nil {
		log.Fatal("Error running application:", err)
	}
}
