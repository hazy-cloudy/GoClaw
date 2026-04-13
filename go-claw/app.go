package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"
	"unsafe"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("GoClaw started!")

	// 监听前端日志事件
	wailsruntime.EventsOn(a.ctx, "frontend-log", func(eventData ...interface{}) {
		if len(eventData) > 0 {
			if msg, ok := eventData[0].(string); ok {
				log.Println("[Frontend]", msg)
			}
		}
	})
}

func (a *App) domReady(ctx context.Context) {
	log.Println("DOM ready!")
}

func (a *App) beforeClose(ctx context.Context) bool {
	log.Println("Closing application...")
	return false
}

func (a *App) shutdown(ctx context.Context) {
	log.Println("Shutdown!")
}

func (a *App) Quit() {
	log.Println("Force quitting...")
	os.Exit(0)
}

func (a *App) GetAppInfo() map[string]string {
	return map[string]string{
		"name":    "GoClaw",
		"version": "0.1.0",
		"author":  "1024XEngineer",
	}
}

func (a *App) GetEmotion() string {
	log.Println("GetEmotion called")

	resultChan := make(chan string, 1)

	// 监听前端的情绪响应 - 使用正确的函数签名
	handler := func(eventData ...interface{}) {
		log.Printf("GetEmotion: received event data: %v", eventData)
		if len(eventData) > 0 {
			if data, ok := eventData[0].(map[string]interface{}); ok {
				if emotion, ok := data["emotion"].(string); ok {
					resultChan <- emotion
					return
				}
			}
		}
		resultChan <- "neutral"
	}

	// 注册临时监听器
	wailsruntime.EventsOn(a.ctx, "emotion-get-result", handler)
	defer wailsruntime.EventsOff(a.ctx, "emotion-get-result")

	// 请求前端获取情绪
	wailsruntime.EventsEmit(a.ctx, "emotion-get")
	log.Println("GetEmotion: emitted emotion-get event")

	// 等待结果，带超时
	select {
	case result := <-resultChan:
		log.Printf("GetEmotion: got result emotion=%s", result)
		return result
	case <-time.After(10 * time.Second):
		log.Println("GetEmotion: timeout")
		return "neutral"
	}
}

func (a *App) Show() {
	log.Println("Show window")
}

func (a *App) Hide() {
	log.Println("Hide window")
}

type Config struct {
	Character     string  `json:"character"`
	Opacity       float64 `json:"opacity"`
	SpeechEnabled bool    `json:"speech_enabled"`
	TouchEnabled  bool    `json:"touch_enabled"`
	AutoStart     bool    `json:"auto_start"`
	VoiceVolume   int     `json:"voice_volume"`
	PositionX     int     `json:"position_x"`
	PositionY     int     `json:"position_y"`
}

type InitSettings struct {
	Initialized   bool    `json:"initialized"`
	Character     string  `json:"character"`
	PetImage      string  `json:"pet_image"`
	WindowWidth   int     `json:"window_width"`
	WindowHeight  int     `json:"window_height"`
	PetScale      float64 `json:"pet_scale"`
	Opacity       float64 `json:"opacity"`
	SpeechEnabled bool    `json:"speech_enabled"`
	TouchEnabled  bool    `json:"touch_enabled"`
	VoiceVolume   int     `json:"voice_volume"`
}

func (a *App) LoadConfig() string {
	data, err := os.ReadFile("config.json")
	if err != nil {
		defaultConfig := Config{
			Character:     "lively",
			Opacity:       1.0,
			SpeechEnabled: true,
			TouchEnabled:  true,
			AutoStart:     false,
			VoiceVolume:   80,
			PositionX:     -1,
			PositionY:     -1,
		}
		result, _ := json.Marshal(defaultConfig)
		return string(result)
	}
	return string(data)
}

func (a *App) SaveConfig(jsonStr string) bool {
	err := os.WriteFile("config.json", []byte(jsonStr), 0644)
	if err != nil {
		log.Printf("Save config error: %v", err)
		return false
	}
	log.Println("Config saved")
	return true
}

func (a *App) LoadInitSettings() string {
	data, err := os.ReadFile("init_settings.json")
	if err != nil {
		defaultSettings := InitSettings{
			Initialized:   false,
			Character:     "lively",
			PetImage:      "default",
			WindowWidth:   280,
			WindowHeight:  380,
			PetScale:      1.0,
			Opacity:       1.0,
			SpeechEnabled: true,
			TouchEnabled:  true,
			VoiceVolume:   80,
		}
		result, _ := json.Marshal(defaultSettings)
		return string(result)
	}
	return string(data)
}

func (a *App) SaveInitSettings(jsonStr string) bool {
	err := os.WriteFile("init_settings.json", []byte(jsonStr), 0644)
	if err != nil {
		log.Printf("Save init settings error: %v", err)
		return false
	}
	log.Println("Init settings saved")
	return true
}

func (a *App) ResetInitSettings() bool {
	err := os.Remove("init_settings.json")
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Reset init settings error: %v", err)
		return false
	}
	log.Println("Init settings reset")
	return true
}

var (
	user32          = windows.NewLazyDLL("user32.dll")
	gwlExtStyle     = int32(-20)         // GWL_EXSTYLE
	wsExTransparent = uint32(0x00000020) // WS_EX_TRANSPARENT
	wsExLayered     = uint32(0x00080000) // WS_EX_LAYERED
)

// StartDrag starts window dragging from frontend
func (a *App) StartDrag() {
	hwnd := getForegroundWindow()
	if hwnd == 0 {
		hwnd = findWindowByTitle("GoClaw")
	}
	if hwnd == 0 {
		log.Println("StartDrag: could not find hwnd")
		return
	}

	log.Printf("StartDrag: hwnd = %d", hwnd)

	// Use WM_NCLBUTTONDOWN with HTCAPTION to start drag
	var procSendMessage = user32.NewProc("SendMessageW")
	var WM_NCLBUTTONDOWN = uint32(0x00A1)
	var HTCAPTION = uint32(2)

	// Get current mouse position
	var procGetCursorPos = user32.NewProc("GetCursorPos")
	type POINT struct {
		X int32
		Y int32
	}
	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	log.Printf("StartDrag: mouse at (%d, %d)", pt.X, pt.Y)

	// Send message to start drag
	procSendMessage.Call(uintptr(hwnd), uintptr(WM_NCLBUTTONDOWN), uintptr(HTCAPTION), 0)
	log.Println("StartDrag: sent message")
}

// SetClickThrough enables or disables click-through for the window
func (a *App) SetClickThrough(enabled bool) {
	log.Println("SetClickThrough called, enabled:", enabled)

	hwnd := getForegroundWindow()
	if hwnd == 0 {
		hwnd = findWindowByTitle("GoClaw")
	}

	if hwnd == 0 {
		log.Println("SetClickThrough: could not find hwnd")
		return
	}

	log.Printf("SetClickThrough: using hwnd = %d", hwnd)

	var procSetWindowRgn = user32.NewProc("SetWindowRgn")
	if enabled {
		ret, _, err := procSetWindowRgn.Call(uintptr(hwnd), 0, 1)
		log.Printf("SetClickThrough: SetWindowRgn ret = %d, err = %v", ret, err)
	} else {
		ret, _, err := procSetWindowRgn.Call(uintptr(hwnd), 0, 0)
		log.Printf("SetClickThrough: SetWindowRgn ret = %d, err = %v", ret, err)
	}

	var procSetWindowLong = user32.NewProc("SetWindowLongW")
	var procGetWindowLong = user32.NewProc("GetWindowLongW")

	extStyle, _, _ := procGetWindowLong.Call(uintptr(hwnd), uintptr(gwlExtStyle))
	log.Printf("SetClickThrough: current extStyle = %d", extStyle)

	if enabled {
		newStyle := extStyle | uintptr(wsExTransparent)
		procSetWindowLong.Call(uintptr(hwnd), uintptr(gwlExtStyle), newStyle)
		log.Printf("SetClickThrough: new style = %d", newStyle)
	} else {
		newStyle := extStyle &^ uintptr(wsExTransparent)
		procSetWindowLong.Call(uintptr(hwnd), uintptr(gwlExtStyle), newStyle)
		log.Printf("SetClickThrough: new style = %d", newStyle)
	}
}

func findWindowByTitle(title string) uintptr {
	var procFindWindow = user32.NewProc("FindWindowW")
	titlePtr, _ := windows.UTF16PtrFromString(title)
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	return hwnd
}

func getForegroundWindow() uintptr {
	var procGetForeground = user32.NewProc("GetForegroundWindow")
	hwnd, _, _ := procGetForeground.Call()
	return hwnd
}
