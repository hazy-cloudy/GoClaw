package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type tools struct {
	ctx context.Context
}

type toolDef struct {
	Name string                 `json:"Name"`
	Args map[string]interface{} `json:"Args"`
}

func NewTools() *tools {
	return &tools{}
}

func (t *tools) SetContext(ctx context.Context) {
	t.ctx = ctx
}

func (t *tools) Execute(toolJSON string) string {
	if toolJSON == "" || toolJSON == "null" {
		return "无工具"
	}

	var tool toolDef
	if err := json.Unmarshal([]byte(toolJSON), &tool); err != nil {
		return "解析失败"
	}

	if tool.Name == "open_app" {
		app, _ := tool.Args["app"].(string)
		if app == "browser" {
			return t.openBrowserSearch("")
		}
	}

	if tool.Name == "web_search" {
		query := ""
		if q, ok := tool.Args["query"].(string); ok {
			query = q
		}
		return t.openBrowserSearch(query)
	}

	if tool.Name == "reminder" {
		text := ""
		delaySeconds := 60
		if val, ok := tool.Args["text"].(string); ok {
			text = val
		}
		if val, ok := tool.Args["delay_seconds"].(float64); ok {
			delaySeconds = int(val)
		}
		return t.setReminder(text, delaySeconds)
	}

	return "未知工具"
}

func (t *tools) openBrowserSearch(query string) string {
	var urlStr string
	if query != "" {
		encoded := url.QueryEscape(query)
		urlStr = "https://www.bing.com/search?q=" + encoded
	} else {
		urlStr = "https://www.bing.com"
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C", "start", urlStr)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x08000000,
		}
		err := cmd.Start()
		if err != nil {
			return fmt.Sprintf("打开失败: %v", err)
		}
		if query != "" {
			return fmt.Sprintf("已搜索: %s", query)
		}
		return "已打开浏览器"
	}

	return "不支持的系统"
}

func (t *tools) setReminder(text string, delaySeconds int) string {
	if text == "" {
		text = "提醒"
	}

	go func() {
		time.Sleep(time.Duration(delaySeconds) * time.Second)
		if t.ctx != nil {
			wailsruntime.EventsEmit(t.ctx, "reminder-trigger", map[string]interface{}{
				"text": text,
			})
		}
	}()

	return fmt.Sprintf("已设置 %d秒 后提醒: %s", delaySeconds, text)
}
