package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
)

type TTSManager struct {
	enabled   bool
	volume    int
	speaking  bool
	voiceName string
	emotion   string
	mu        sync.Mutex
}

type TTSConfig struct {
	Enabled bool
	Volume  int
}

var emotionParams = map[string]struct{ Rate, Pitch string }{
	"happy":     {"+15%", "+10Hz"},
	"excited":   {"+25%", "+15Hz"},
	"sad":       {"-15%", "-10Hz"},
	"angry":     {"+20%", "-15Hz"},
	"shy":       {"-10%", "+5Hz"},
	"scared":    {"-5%", "-20Hz"},
	"surprised": {"+10%", "+20Hz"},
	"sleepy":    {"-30%", "-15Hz"},
	"bored":     {"-20%", "-5Hz"},
	"worried":   {"-10%", "-10Hz"},
	"default":   {"+0%", "+0Hz"},
}

func NewTTSManager() *TTSManager {
	return &TTSManager{
		enabled:   true,
		volume:    80,
		speaking:  false,
		voiceName: "zh-CN-XiaoxiaoNeural",
		emotion:   "default",
	}
}

func (t *TTSManager) Speak(text string) error {
	if !t.enabled || text == "" {
		return nil
	}
	cleaned := cleanText(text)
	if cleaned == "" {
		return nil
	}
	go t.speakAsync(cleaned, t.emotion)
	return nil
}

func (t *TTSManager) SpeakWithEmotion(text string, emotion string) error {
	if !t.enabled || text == "" {
		return nil
	}
	cleaned := cleanText(text)
	if cleaned == "" {
		return nil
	}
	go t.speakAsync(cleaned, emotion)
	return nil
}

func (t *TTSManager) speakAsync(text string, emotion string) {
	t.mu.Lock()
	t.speaking = true
	t.mu.Unlock()
	defer func() {
		t.mu.Lock()
		t.speaking = false
		t.mu.Unlock()
	}()

	params := emotionParams["default"]
	if p, ok := emotionParams[emotion]; ok {
		params = p
	}

	t.speakWithPython(text, params.Rate, params.Pitch)
}

func (t *TTSManager) speakWithPython(text, rate, pitch string) error {
	script := fmt.Sprintf(`
import asyncio
import edge_tts
import tempfile
import os
import ctypes
import time

async def main():
    try:
        text = '''%s'''
        rate = '%s'
        pitch = '%s'
        
        tmp = tempfile.NamedTemporaryFile(suffix='.mp3', delete=False)
        tmp_path = tmp.name
        tmp.close()
        
        await edge_tts.Communicate(text, "zh-CN-XiaoxiaoNeural", rate=rate, pitch=pitch).save(tmp_path)
        
        mci = ctypes.windll.winmm.mciSendStringW
        buf = ctypes.create_unicode_buffer(256)
        path_escaped = tmp_path.replace('\\', '\\\\')
        
        try:
            mci('close pet_audio', buf, 256, None)
        except:
            pass
        
        mci(f'open "{path_escaped}" type mpegvideo alias pet_audio', buf, 256, None)
        mci('play pet_audio', buf, 256, None)
        
        while True:
            time.sleep(0.2)
            mci('status pet_audio mode', buf, 256, None)
            if buf.value != 'playing':
                break
    except:
        pass
    finally:
        try:
            mci('close pet_audio', buf, 256, None)
        except:
            pass
        try:
            os.unlink(tmp_path)
        except:
            pass

asyncio.run(main())
`, text, rate, pitch)

	tmpFile := os.TempDir() + "\\go-claw-tts.py"
	if err := os.WriteFile(tmpFile, []byte(script), 0644); err != nil {
		return fmt.Errorf("write script failed: %w", err)
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command("python", tmpFile)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	cmd.Run()
	return nil
}

func cleanText(text string) string {
	text = strings.TrimSpace(text)
	emojiPattern := regexp.MustCompile(`[\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F1E0}-\x{1F1FF}\x{2600}-\x{26FF}]+`)
	text = emojiPattern.ReplaceAllString(text, "")
	invalidChars := regexp.MustCompile(`[【】\[\]{}【】《》<>"'\\|@#$%^&*+=:=/\\|]`)
	text = invalidChars.ReplaceAllString(text, "")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	if len(text) < 2 {
		return ""
	}
	if len(text) > 200 {
		text = text[:200]
	}
	return text
}

func (t *TTSManager) SetEmotion(emotion string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := emotionParams[emotion]; ok {
		t.emotion = emotion
	}
}

func (t *TTSManager) SetEnabled(enabled bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.enabled = enabled
}

func (t *TTSManager) SetVolume(volume int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	t.volume = volume
}

func (t *TTSManager) IsSpeaking() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.speaking
}

func (t *TTSManager) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.speaking = false
}

// PlayAudio plays base64 encoded audio data
func (t *TTSManager) PlayAudio(audioBase64 string, audioType string) error {
	if !t.enabled || audioBase64 == "" {
		return nil
	}

	data, err := base64.StdEncoding.DecodeString(audioBase64)
	if err != nil {
		return fmt.Errorf("failed to decode audio: %w", err)
	}

	tmpFile := os.TempDir() + "\\go-claw-audio.mp3"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write audio file: %w", err)
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command("python", "-c", fmt.Sprintf(`
import ctypes
import time

mci = ctypes.windll.winmm.mciSendStringW
buf = ctypes.create_unicode_buffer(256)

try:
    mci('close pet_audio', buf, 256, None)
except:
    pass

mci('open "%s" type mpegvideo alias pet_audio', buf, 256, None)
mci('play pet_audio', buf, 256, None)

while True:
    time.sleep(0.2)
    mci('status pet_audio mode', buf, 256, None)
    if buf.value != 'playing':
        break

mci('close pet_audio', buf, 256, None)
`, strings.ReplaceAll(tmpFile, "\\", "\\\\")))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	cmd.Run()
	return nil
}

func (t *TTSManager) GetConfig() string {
	config := TTSConfig{
		Enabled: t.enabled,
		Volume:  t.volume,
	}
	data, _ := json.Marshal(config)
	return string(data)
}

func (t *TTSManager) LoadConfig(configJSON string) {
	var config TTSConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err == nil {
		t.mu.Lock()
		t.enabled = config.Enabled
		t.volume = config.Volume
		t.mu.Unlock()
	}
}
