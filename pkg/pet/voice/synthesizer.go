package voice

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Synthesizer 语音合成器
// 整合TTS提供者和消息发送者，支持从带标签的文本中解析出多个语音片段及其参数
type Synthesizer struct {
	provider StreamingTTS // TTS提供者
	sender   *Sender      // 消息发送者
}

// NewSynthesizer 创建新的合成器实例
func NewSynthesizer(provider StreamingTTS, sender *Sender) *Synthesizer {
	return &Synthesizer{
		provider: provider,
		sender:   sender,
	}
}

// ParsedText 解析后的文本片段及其语音参数
type ParsedText struct {
	Text   string      // 文本内容
	Params VoiceParams // 语音参数
}

// 预编译的正则表达式，避免重复编译
var (
	textTagRegex  = regexp.MustCompile(`\[text:([^\]]+)\]`)
	voiceTagRegex = regexp.MustCompile(`\[voice:([^\]]+)\]`)
)

// ParseAndSynthesize 解析文本标签并执行语音合成
// 支持格式: [text:要说的内容] 或 [text:内容]-[voice:speed:1.2,emotion:happy]
// 解析后调用TTS提供者合成语音，并通过sender发送音频块
func (s *Synthesizer) ParseAndSynthesize(sessionID string, chatID int64, content string, emotion string) error {
	fmt.Printf("pet-voice: ParseAndSynthesize called, content=%s\n", content)
	texts := s.parseTextTags(content)
	fmt.Printf("pet-voice: parseTextTags returned %d texts\n", len(texts))
	if len(texts) == 0 {
		return nil
	}

	for _, parsed := range texts {
		if parsed.Text == "" {
			continue
		}

		params := parsed.Params

		// 设置默认参数
		if params.Speed == 0 {
			params.Speed = 1.0
		}
		if params.Vol == 0 {
			params.Vol = 1.0
		}

		fmt.Printf("pet-voice: synthesizing, text=%s\n", parsed.Text)

		ch, err := s.provider.Synthesize(parsed.Text, params)
		if err != nil {
			fmt.Printf("pet-voice: synthesize failed, error=%v\n", err)
			s.sender.SendError(sessionID, chatID, err.Error())
			continue
		}

		fmt.Printf("pet-voice: waiting for audio chunks...\n")
		// 从通道读取音频块并发送
		for chunk := range ch {
			fmt.Printf("pet-voice: received chunk, size=%d, isLast=%v\n", len(chunk.Data), chunk.IsLast)
			if err := s.sender.SendAudioChunk(sessionID, chatID, chunk); err != nil {
				fmt.Printf("pet-voice: send audio chunk failed, error=%v\n", err)
				break
			}
			if chunk.IsLast {
				fmt.Printf("pet-voice: received final chunk, done\n")
			}
		}
		fmt.Printf("pet-voice: chunk channel closed, done\n")
	}

	return nil
}

// parseTextTags 解析文本中的标签
// 支持格式: [text:xxx](-[voice:yyy])?
// voice部分是可选的
// 如果没有匹配到任何标签，则返回整个内容作为纯文本
func (s *Synthesizer) parseTextTags(content string) []ParsedText {
	var results []ParsedText

	// 组合正则: [text:内容](-[voice:参数])?
	combinedRegex := regexp.MustCompile(`\[text:([^\]]+)\](?:-\[voice:([^\]]+)\])?`)
	matches := combinedRegex.FindAllStringSubmatch(content, -1)

	// 如果没有匹配到任何 [text:] 标签，将整个内容作为纯文本返回
	if len(matches) == 0 {
		cleaned := CleanTextTags(content)
		if cleaned != "" {
			results = append(results, ParsedText{
				Text:   cleaned,
				Params: DefaultVoiceParams(),
			})
		}
		return results
	}

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		text := strings.TrimSpace(match[1])
		var params VoiceParams

		// 若存在voice标签，则解析参数
		if len(match) >= 3 && match[2] != "" {
			params = s.parseVoiceParams(match[2])
		}

		results = append(results, ParsedText{
			Text:   text,
			Params: params,
		})
	}

	return results
}

// parseVoiceParams 解析语音参数字符串
// 格式: key:value,key2:value2
// 支持的键: speed, vol/volume, pitch, emotion/emo
func (s *Synthesizer) parseVoiceParams(voiceStr string) VoiceParams {
	params := DefaultVoiceParams()

	pairs := strings.Split(voiceStr, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "speed":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				params.Speed = v
			}
		case "vol", "volume":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				params.Vol = v
			}
		case "pitch":
			if v, err := strconv.Atoi(value); err == nil {
				params.Pitch = v
			}
		}
	}

	return params
}

// CleanTextTags 移除文本中所有标签，保留纯文本
func CleanTextTags(content string) string {
	tagPattern := regexp.MustCompile(`\[[^\]]+\]`)
	return tagPattern.ReplaceAllString(content, "")
}

// ExtractPureText 提取所有[text:xxx]标签中的文本内容
// 若没有text标签，则返回CleanTextTags的结果
func ExtractPureText(content string) string {
	var sb strings.Builder
	textRegex := regexp.MustCompile(`\[text:([^\]]+)\]`)
	matches := textRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			if sb.Len() > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(match[1])
		}
	}

	if sb.Len() == 0 {
		return CleanTextTags(content)
	}

	return sb.String()
}
