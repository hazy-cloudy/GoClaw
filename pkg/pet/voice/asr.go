package voice

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/sipeed/picoclaw/pkg/logger"
)

// 音频帧类型标识
const (
	FrameTypeAudio = 0x01 // 音频帧类型标识
)

// AudioReceiver 音频接收者接口
// 定义ASR事件的回调处理
type AudioReceiver interface {
	// OnAudio 收到音频数据时的回调
	OnAudio(sessionID string, data []byte) error
	// OnSilence 检测到静音时的回调，返回识别结果
	OnSilence(sessionID string) (string, error)
}

// ASRHandler 语音识别处理器
// 负责解析和分发音频帧
type ASRHandler struct {
	receiver AudioReceiver // 音频接收者
}

// NewASRHandler 创建ASR处理器实例
func NewASRHandler(receiver AudioReceiver) *ASRHandler {
	return &ASRHandler{receiver: receiver}
}

// HandleFrame 处理接收到的音频帧
// 根据frameType分发到对应的处理方法
func (h *ASRHandler) HandleFrame(sessionID string, frameType byte, data []byte) error {
	switch frameType {
	case FrameTypeAudio:
		return h.receiver.OnAudio(sessionID, data)
	default:
		logger.WarnCF("pet-voice", "unknown frame type", map[string]any{
			"session_id": sessionID,
			"frame_type": frameType,
		})
	}
	return nil
}

// ParseAudioFrame 解析音频帧
// 帧格式: [1字节类型][4字节长度][N字节数据]
// 返回: frameType, data, error
func ParseAudioFrame(raw []byte) (frameType byte, data []byte, err error) {
	if len(raw) < 5 {
		return 0, nil, io.ErrShortBuffer
	}

	frameType = raw[0]
	length := binary.BigEndian.Uint32(raw[1:5]) // 大端序读取长度

	if len(raw) < 5+int(length) {
		return 0, nil, io.ErrShortBuffer
	}

	data = raw[5 : 5+length]
	return frameType, data, nil
}

// BuildAudioFrame 构建音频帧
// 帧格式: [1字节类型][4字节长度][N字节数据]
func BuildAudioFrame(data []byte) []byte {
	frame := make([]byte, 5+len(data))
	frame[0] = FrameTypeAudio
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)
	return frame
}

// SessionAudioBuffer 会话音频缓冲区
// 用于累积接收的音频数据
type SessionAudioBuffer struct {
	sessionID string
	buffer    *bytes.Buffer
}

// NewSessionAudioBuffer 创建会话音频缓冲区
func NewSessionAudioBuffer(sessionID string) *SessionAudioBuffer {
	return &SessionAudioBuffer{
		sessionID: sessionID,
		buffer:    bytes.NewBuffer(nil),
	}
}

// Write 将音频数据写入缓冲区
func (s *SessionAudioBuffer) Write(data []byte) (int, error) {
	return s.buffer.Write(data)
}

// GetAudio 获取缓冲区中的所有音频数据
func (s *SessionAudioBuffer) GetAudio() []byte {
	return s.buffer.Bytes()
}

// Reset 清空缓冲区
func (s *SessionAudioBuffer) Reset() {
	s.buffer.Reset()
}

// Len 返回缓冲区中音频数据的长度
func (s *SessionAudioBuffer) Len() int {
	return s.buffer.Len()
}
