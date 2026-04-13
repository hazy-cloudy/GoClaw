// Package pet 提供桌宠通道，通过WebSocket连接桌面客户端，处理API请求
package pet

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"net/http"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/channels"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/pet"
)

// Request 定义了客户端请求的类型
type Request = pet.Request

// Response 定义了服务端响应的类型
type Response = pet.Response

// Push 定义了服务端推送的类型
type Push = pet.Push

// AudioPush 音频推送数据结构
type AudioPush struct {
	AudioType   string `json:"audio_type"`   // 音频类型: audio, voice
	ContentType string `json:"content_type"` // MIME类型: audio/mpeg, audio/wav等
	Data        string `json:"data"`         // base64编码的音频数据
}

const (
	pingInterval = 30 * time.Second // 心跳间隔
	readTimeout  = 60 * time.Second // 读取超时时间
)

// PetChannel 桌面宠通道
// 通过WebSocket连接桌面客户端，处理API请求
type PetChannel struct {
	*channels.BaseChannel                     // 基础通道
	config                config.PetConfig    // 通道配置
	msgBus                *bus.MessageBus     // 消息总线
	upgrader              websocket.Upgrader  // WebSocket升级器
	connections           map[string]*petConn // 连接映射表：connID -> connection
	connsMu               sync.RWMutex        // 连接映射表的读写锁
	ctx                   context.Context     // 上下文
	cancel                context.CancelFunc  // 取消函数
	service               *pet.PetService     // 桌宠服务
}

// petConn 表示一个WebSocket连接
type petConn struct {
	id        string          // 连接ID
	conn      *websocket.Conn // WebSocket连接
	sessionID string          // 会话ID
	writeMu   sync.Mutex      // 写操作锁
	closed    bool            // 连接是否关闭
}

// NewPetChannel 创建Pet Channel
// cfg: pet channel 配置
// msgBus: 消息总线
// workspacePath: 工作区路径，用于存储角色配置等文件
func NewPetChannel(cfg config.PetConfig, msgBus *bus.MessageBus, workspacePath string) (*PetChannel, error) {
	ctx, cancel := context.WithCancel(context.Background())

	base := channels.NewBaseChannel("pet", cfg, msgBus, cfg.AllowFrom)

	// Origin 检查函数，用于 WebSocket 升级时的安全验证
	checkOrigin := func(r *http.Request) bool {
		if len(cfg.AllowOrigins) == 0 {
			return true
		}
		origin := r.Header.Get("Origin")
		for _, allowed := range cfg.AllowOrigins {
			if allowed == "*" || allowed == origin {
				return true
			}
		}
		return false
	}

	pc := &PetChannel{
		BaseChannel: base,
		config:      cfg,
		msgBus:      msgBus,
		upgrader: websocket.Upgrader{
			CheckOrigin:     checkOrigin,
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		connections: make(map[string]*petConn),
		ctx:         ctx,
		cancel:      cancel,
	}

	// 创建 PetService，传递工作区路径用于存储角色配置
	pc.service = pet.NewPetService(msgBus, pet.PetServiceConfig{
		WorkspacePath: workspacePath,
	})
	pc.service.SetPushHandler(pc.handleServicePush)
	pc.service.Start()

	logger.Infof("pet: created PetChannel with config: %v", cfg)

	return pc, nil
}

// handleServicePush 处理 PetService 的推送
func (c *PetChannel) handleServicePush(push any) {
	c.connsMu.RLock()
	defer c.connsMu.RUnlock()

	for _, pc := range c.connections {
		if err := pc.writeJSON(push); err != nil {
			logger.Warnf("pet: failed to push to conn_id=%s: %v", pc.id, err)
		}
	}
}

// Name 通道名称
func (c *PetChannel) Name() string {
	return "pet"
}

// Service 返回 PetService 实例，供 LLMTagHook 等组件访问情绪引擎和动作管理器
func (c *PetChannel) Service() *pet.PetService {
	return c.service
}

// Start 启动通道
func (c *PetChannel) Start(ctx context.Context) error {
	go func() {
		addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
		httpServer := &http.Server{
			Addr:         addr,
			Handler:      c,
			ReadTimeout:  readTimeout,
			WriteTimeout: 10 * time.Second,
		}
		logger.Infof("pet: starting WebSocket server at %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("pet: WebSocket server error: %v", err)
		}
	}()

	go c.runHeartbeat()

	logger.InfoCF("pet", "Pet channel started", nil)
	return nil
}

// Stop 停止通道
func (c *PetChannel) Stop(ctx context.Context) error {
	c.cancel()
	c.connsMu.Lock()
	for _, pc := range c.connections {
		pc.conn.Close()
	}
	c.connections = make(map[string]*petConn)
	c.connsMu.Unlock()
	logger.InfoCF("pet", "Pet channel stopped", nil)
	return nil
}

// Send 发送消息（实现Channel接口）
func (c *PetChannel) Send(ctx context.Context, msg bus.OutboundMessage) ([]string, error) {
	if !c.IsRunning() {
		return nil, channels.ErrNotRunning
	}
	logger.Infof("pet: Send, msg=%v", msg)

	// 如果是 cron 提醒（chatID 以 "cron-" 开头），广播到所有客户端
	if strings.HasPrefix(msg.ChatID, "cron-") {
		c.broadcastReminder(msg.Content)
	} else {
		c.sendToClient(msg.ChatID, msg.Content)
	}
	return nil, nil
}

// SendMedia 发送媒体消息（实现MediaSender接口）
func (c *PetChannel) SendMedia(ctx context.Context, msg bus.OutboundMediaMessage) ([]string, error) {
	if !c.IsRunning() {
		return nil, channels.ErrNotRunning
	}

	logger.Infof("pet: SendMedia, msg=%v", msg)

	for _, part := range msg.Parts {
		if part.Type == "audio" || part.Type == "voice" {
			// 解析媒体引用获取文件路径
			localPath, err := c.GetMediaStore().Resolve(part.Ref)
			if err != nil {
				logger.Warnf("pet: failed to resolve media ref %s: %v", part.Ref, err)
				continue
			}

			// 读取音频文件并发送到客户端
			if err := c.sendAudioToClient(localPath, part.Type); err != nil {
				logger.Warnf("pet: failed to send audio: %v", err)
				continue
			}
		}
	}

	return nil, nil
}

// sendAudioToClient 发送音频文件到客户端
func (c *PetChannel) sendAudioToClient(filePath string, audioType string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read audio file: %w", err)
	}

	// 将音频数据转换为 base64
	audioBase64 := base64.StdEncoding.EncodeToString(data)

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(filePath))
	contentType := "audio/mpeg"
	if ext == ".wav" {
		contentType = "audio/wav"
	} else if ext == ".ogg" {
		contentType = "audio/ogg"
	} else if ext == ".opus" {
		contentType = "audio/opus"
	}

	c.connsMu.RLock()
	defer c.connsMu.RUnlock()

	logger.Infof("pet: sending audio to %d connections, size=%d", len(c.connections), len(data))

	for connID, pc := range c.connections {
		logger.Infof("pet: sending audio to conn_id=%s", connID)
		pc.writeJSON(pet.Push{
			Type:     "push",
			PushType: pet.PushTypeAudio,
			Data: mustMarshal(AudioPush{
				AudioType:   audioType,
				ContentType: contentType,
				Data:        audioBase64,
			}),
			Timestamp: time.Now().Unix(),
		})
	}

	return nil
}

// IsRunning 是否运行
func (c *PetChannel) IsRunning() bool {
	return c.ctx.Err() == nil
}

// IsAllowed 是否允许
func (c *PetChannel) IsAllowed(senderID string) bool {
	return true
}

// IsAllowedSender 是否允许发送者
func (c *PetChannel) IsAllowedSender(sender bus.SenderInfo) bool {
	return true
}

// ReasoningChannelID 推理通道ID
func (c *PetChannel) ReasoningChannelID() string {
	return ""
}

// runHeartbeat 定期发送心跳
func (c *PetChannel) runHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.broadcastPush(pet.Push{
				Type:      "push",
				PushType:  pet.PushTypeHeartbeat,
				Data:      mustMarshal(map[string]interface{}{"timestamp": time.Now().Unix()}),
				Timestamp: time.Now().Unix(),
			})
		}
	}
}

// sendToClient 发送消息到指定session的客户端
func (c *PetChannel) sendToClient(sessionID, content string) {
	c.connsMu.RLock()
	defer c.connsMu.RUnlock()

	logger.Infof("pet: sendToClient, sessionID=%s, content=%s, connections_count=%d", sessionID, content, len(c.connections))

	matched := false
	for connID, pc := range c.connections {
		logger.Infof("pet: checking connection conn_id=%s, sessionID=%s", connID, pc.sessionID)
		// 匹配条件：sessionID相同、broadcast、或连接是默认会话（用于桌宠客户端）
		if pc.sessionID == sessionID || sessionID == "broadcast" || pc.sessionID == "default" {
			logger.Infof("pet: matched! sending to conn_id=%s", connID)
			data, _ := json.Marshal(content)
			pc.writeJSON(Response{
				Status: pet.StatusOK,
				Action: pet.ActionChat,
				Data:   data,
			})
			matched = true
		}
	}

	// 如果没有匹配任何连接，但有连接存在，则广播到所有连接（用于cron提醒等场景）
	if !matched && len(c.connections) > 0 && sessionID != "default" {
		logger.Infof("pet: no direct match, broadcasting to all connections")
		for connID, pc := range c.connections {
			logger.Infof("pet: broadcasting to conn_id=%s", connID)
			data, _ := json.Marshal(content)
			pc.writeJSON(Response{
				Status: pet.StatusOK,
				Action: pet.ActionChat,
				Data:   data,
			})
		}
	}
}

// broadcastReminder 广播提醒消息到所有连接的客户端
func (c *PetChannel) broadcastReminder(content string) {
	c.connsMu.RLock()
	defer c.connsMu.RUnlock()

	logger.Infof("pet: broadcastReminder, content=%s, connections_count=%d", content, len(c.connections))

	for connID, pc := range c.connections {
		logger.Infof("pet: sending reminder to conn_id=%s", connID)
		data, _ := json.Marshal(content)
		pc.writeJSON(Response{
			Status: pet.StatusOK,
			Action: pet.ActionChat,
			Data:   data,
		})
	}
}

// broadcastPush 广播推送消息
func (c *PetChannel) broadcastPush(push pet.Push) {
	c.connsMu.RLock()
	defer c.connsMu.RUnlock()

	for _, pc := range c.connections {
		if err := pc.writeJSON(push); err != nil {
			logger.Warnf("pet: failed to broadcast push to conn_id=%s: %v", pc.id, err)
		}
	}
}

// ServeHTTP 处理HTTP请求（WebSocket升级）
func (c *PetChannel) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/ws" || r.URL.Path == "/" {
		c.handleWebSocket(w, r)
		return
	}
	http.NotFound(w, r)
}

// handleWebSocket 处理WebSocket连接
func (c *PetChannel) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !c.IsRunning() {
		http.Error(w, "channel not running", http.StatusServiceUnavailable)
		return
	}

	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Warnf("pet: failed to upgrade: %v", err)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = "default"
	}

	pc := &petConn{
		id:        generateConnID(),
		conn:      conn,
		sessionID: sessionID,
	}

	c.connsMu.Lock()
	c.connections[pc.id] = pc
	c.connsMu.Unlock()

	c.service.RegisterSession(pc.id, sessionID)

	logger.Infof("pet: client connected, conn_id=%s, session=%s", pc.id, sessionID)

	go c.readLoop(pc)
}

// readLoop 读取客户端消息
func (c *PetChannel) readLoop(pc *petConn) {
	defer func() {
		c.service.UnregisterSession(pc.id)
		pc.conn.Close()
		c.connsMu.Lock()
		delete(c.connections, pc.id)
		c.connsMu.Unlock()
		logger.Infof("pet: client disconnected, conn_id=%s", pc.id)
	}()

	pc.conn.SetReadDeadline(time.Now().Add(readTimeout))
	pc.conn.SetPongHandler(func(appData string) error {
		pc.conn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				pc.conn.WriteMessage(websocket.PingMessage, nil)
			}
		}
	}()

	for {
		_, rawMsg, err := pc.conn.ReadMessage()
		if err != nil {
			return
		}

		var req Request
		if err := json.Unmarshal(rawMsg, &req); err != nil {
			pc.writeJSON(Response{Status: pet.StatusError, Error: "invalid request"})
			continue
		}

		c.handleRequest(pc, req)
	}
}

// handleRequest 处理客户端请求
func (c *PetChannel) handleRequest(pc *petConn, req Request) {
	c.service.HandleRequest(pc.id, req)
}

// writeJSON 发送JSON到客户端
func (pc *petConn) writeJSON(v any) error {
	if pc.closed {
		return fmt.Errorf("connection closed")
	}
	pc.writeMu.Lock()
	defer pc.writeMu.Unlock()
	err := pc.conn.WriteJSON(v)
	if err != nil {
		logger.Warnf("pet: writeJSON failed conn_id=%s, err=%v", pc.id, err)
	}
	return err
}

// generateConnID 生成连接ID
func generateConnID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// mustMarshal 将任意数据转换为适合嵌入推送信封的原始JSON
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{"error": "marshal error"}`)
	}
	return data
}

// petStreamer 实现流式输出，通过WebSocket发送增量内容
type petStreamer struct {
	channel   *PetChannel
	sessionID string
	lastLen   int
	buffer    string
	chatID    int64

	// 状态机相关
	textBuffer  strings.Builder // 收集 [text:] 中的文本
	tagBuffer   strings.Builder // 收集其他标签内容
	inTextTag   bool            // 是否在 [text:...] 标签内
	inOtherTag  bool            // 是否在其他标签内（如 [emotion:...]）
	pendingText string          // 已收集待发送的文本（遇到标签前的）
}

// BeginStream 实现StreamingCapable接口
func (c *PetChannel) BeginStream(ctx context.Context, sessionID string) (channels.Streamer, error) {
	return &petStreamer{
		channel:   c,
		sessionID: sessionID,
		chatID:    0,
	}, nil
}

// Update 发送增量内容到客户端，实时解析 [text:xxx] 标签并发送文本块
func (s *petStreamer) Update(ctx context.Context, content string) error {
	if s == nil || s.channel == nil {
		return nil
	}

	s.lastLen = len(content)
	s.buffer += content

	sendPending := func() {
		if len(s.buffer) > 0 {
			s.chatID++
			s.channel.sendStreamChunk(s.sessionID, s.chatID, "text", s.buffer, false)
			s.buffer = ""
		}
	}

	if len(s.buffer) > 0 {
		if s.inTextTag {
			// 正在收集 [text:...] 标签内的内容
			if idx := strings.Index(s.buffer, "]"); idx >= 0 {
				// 找到标签结束
				textContent := s.buffer[:idx]
				s.buffer = s.buffer[idx+1:]
				s.inTextTag = false
				i := strings.Index(s.buffer, "]")
				s.buffer = s.buffer[:i]
			}
			sendPending()
		} else if strings.Contains(s.buffer, "[text:") {
			s.inTextTag = true
			i := strings.Index(s.buffer, "[text:")
			s.buffer = s.buffer[i+6:]
			sendPending()
		}
	}

	return nil
}

// Finalize 发送最终完成标记，包含完整内容
func (s *petStreamer) Finalize(ctx context.Context, content string) error {
	if s == nil || s.channel == nil {
		return nil
	}

	// 清理任何残留的 [text:] 标签
	cleanContent := cleanStreamTags(content)

	// 发送最终内容块
	if cleanContent != "" {
		s.chatID++
		s.channel.sendStreamChunk(s.sessionID, s.chatID, "final", cleanContent, true)
	} else {
		// 即使内容为空，也发送 final 标记
		s.chatID++
		s.channel.sendStreamChunk(s.sessionID, s.chatID, "final", "", true)
	}

	return nil
}

// Cancel 取消流式输出
func (s *petStreamer) Cancel(ctx context.Context) {
}

// findSegment 查找第一个完整的段落（以换行或句号结尾）
// 返回: 段落内容, 剩余内容, 是否找到
func (s *petStreamer) findSegment(buffer string) (string, string, bool) {
	runes := []rune(buffer)
	for i := len(runes) - 1; i >= 0; i-- {
		c := runes[i]
		if c == '\n' || c == '.' {
			segment := string(runes[:i+1])
			remaining := string(runes[i+1:])
			return segment, remaining, true
		}
	}
	return "", buffer, false
}

// cleanSegment 清理段落，去除纯空白、无用符号和标签
func (s *petStreamer) cleanSegment(segment string) string {
	// 先移除所有标签
	segment = cleanStreamTags(segment)

	runes := []rune(segment)
	cleaned := make([]rune, 0, len(runes))

	for i, r := range runes {
		if r == '\n' {
			hasTextBefore := false
			for j := 0; j < i; j++ {
				if !isWhitespace(runes[j]) {
					hasTextBefore = true
					break
				}
			}
			if !hasTextBefore {
				continue
			}
		}
		cleaned = append(cleaned, r)
	}

	result := string(cleaned)
	result = removeTrailingDots(result)
	return strings.TrimRight(result, " \t")
}

// cleanStreamTags 清理流式输出中的所有标签
func cleanStreamTags(content string) string {
	// 匹配所有 [xxx:yyy] 或 [xxx] 格式的标签
	tagPattern := regexp.MustCompile(`\[[^\]]+\]`)
	content = tagPattern.ReplaceAllString(content, "")

	// 清理多余的空行
	emptyLinePattern := regexp.MustCompile(`\n{3,}`)
	content = emptyLinePattern.ReplaceAllString(content, "\n\n")

	return strings.TrimSpace(content)
}

// isWhitespace 检查是否为空白字符
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// removeTrailingDots 去除末尾连续的句号（后面无文字时）
func removeTrailingDots(s string) string {
	runes := []rune(s)
	for len(runes) > 1 && runes[len(runes)-1] == '.' {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}

// detectContentType 根据内容前缀判断类型
func detectContentType(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```go") || strings.HasPrefix(text, "```golang") {
		return "go"
	}
	if strings.HasPrefix(text, "```json") {
		return "json"
	}
	if strings.HasPrefix(text, "```python") || strings.HasPrefix(text, "```py") {
		return "python"
	}
	if strings.HasPrefix(text, "```javascript") || strings.HasPrefix(text, "```js") {
		return "javascript"
	}
	if strings.HasPrefix(text, "```typescript") || strings.HasPrefix(text, "```ts") {
		return "typescript"
	}
	if strings.HasPrefix(text, "```html") {
		return "html"
	}
	if strings.HasPrefix(text, "```css") {
		return "css"
	}
	if strings.HasPrefix(text, "```sql") {
		return "sql"
	}
	if strings.HasPrefix(text, "```bash") || strings.HasPrefix(text, "```sh") {
		return "bash"
	}
	if strings.HasPrefix(text, "```xml") {
		return "xml"
	}
	if strings.HasPrefix(text, "```yaml") || strings.HasPrefix(text, "```yml") {
		return "yaml"
	}
	if strings.HasPrefix(text, "```markdown") || strings.HasPrefix(text, "```md") {
		return "markdown"
	}
	if strings.HasPrefix(text, "```") {
		return "code"
	}
	return "text"
}

// sendStreamChunk 发送流式数据块
func (c *PetChannel) sendStreamChunk(sessionID string, chatID int64, contentType, text string, isFinal bool) {
	c.connsMu.RLock()
	defer c.connsMu.RUnlock()

	var emotion, act string
	if c.service != nil {
		emotion, _ = c.service.EmotionEngine().GetDominantEmotion()
	}

	streamData := StreamData{
		ChatID:      chatID,
		ContentType: contentType,
		Text:        text,
		Emotion:     emotion,
		Action:      act,
	}

	for connID, pc := range c.connections {
		// 匹配条件：sessionID相同，或者是默认会话（用于桌宠客户端）
		if pc.sessionID == sessionID || sessionID == "broadcast" || pc.sessionID == "default" {
			pc.writeJSON(PetStreamResponse{
				Type:      "push",
				PushType:  "ai_chat",
				Data:      streamData,
				IsFinal:   isFinal,
				Timestamp: time.Now().Unix(),
			})
			_ = connID
		}
	}
}

// sendResponse 发送普通响应
func (c *PetChannel) sendResponse(sessionID string, action string, data map[string]interface{}) {
	c.connsMu.RLock()
	defer c.connsMu.RUnlock()

	for connID, pc := range c.connections {
		if pc.sessionID == sessionID || sessionID == "broadcast" {
			pc.writeJSON(Response{
				Status: pet.StatusOK,
				Action: action,
				Data:   mustMarshal(data),
			})
			_ = connID
		}
	}
}

// StreamData 流式数据内容
type StreamData struct {
	ChatID      int64  `json:"chat_id"`
	ContentType string `json:"type"`
	Text        string `json:"text"`
	Emotion     string `json:"emotion,omitempty"`
	Action      string `json:"action,omitempty"`
}

// PetStreamResponse 流式响应结构
type PetStreamResponse struct {
	Type      string     `json:"type"`
	PushType  string     `json:"push_type"`
	Data      StreamData `json:"data"`
	IsFinal   bool       `json:"is_final"`
	Timestamp int64      `json:"timestamp"`
}
