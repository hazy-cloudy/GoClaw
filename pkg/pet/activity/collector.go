package activity

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

// NewID 生成一条活动记录的简易唯一 ID。
// 第一阶段只要能区分事件即可，不需要引入更重的 UUID 依赖。
func NewID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf[:])
}

// ClassifyText 是一个“足够简单”的文本分类器。
//
// 这里的分类不是最终智能分类，而是为了先把数据粗分到：
// code / doc / debug / pptx / config / other
// 后续周报和催办先基于这个粗分类就能工作起来。
func ClassifyText(text string) Category {
	lower := strings.ToLower(strings.TrimSpace(text))
	switch {
	case strings.Contains(lower, "ppt") || strings.Contains(lower, "slide") || strings.Contains(lower, "presentation"):
		return CategoryPPTX
	case strings.Contains(lower, "bug") || strings.Contains(lower, "error") || strings.Contains(lower, "报错") || strings.Contains(lower, "排查"):
		return CategoryDebug
	case strings.Contains(lower, "readme") || strings.Contains(lower, "文档") || strings.Contains(lower, "md"):
		return CategoryDoc
	case strings.Contains(lower, "config") || strings.Contains(lower, "配置"):
		return CategoryConfig
	case strings.Contains(lower, "code") || strings.Contains(lower, "go ") || strings.Contains(lower, "ts") || strings.Contains(lower, "修复"):
		return CategoryCode
	default:
		return CategoryOther
	}
}

// BuildUserMessageEvent 把用户输入转成一条活动记录。
//
// 这一步相当于把“用户说过的话”转成后续可聚合的数据。
func BuildUserMessageEvent(characterID, sessionID, text string) *Event {
	trimmed := strings.TrimSpace(text)
	title := trimmed
	if len([]rune(title)) > 32 {
		title = string([]rune(title)[:32])
	}
	if title == "" {
		title = "empty_message"
	}
	return &Event{
		ID:          NewID(),
		CharacterID: characterID,
		SessionID:   sessionID,
		Type:        EventUserMessage,
		Category:    ClassifyText(trimmed),
		Status:      StatusDone,
		Title:       title,
		Summary:     trimmed,
		CreatedAt:   time.Now(),
	}
}
