// Package pet 提供桌宠通道，通过WebSocket连接桌面客户端，处理API请求
package pet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

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
	c.sendToClient(msg.ChatID, msg.Content)
	return nil, nil
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

	for connID, pc := range c.connections {
		logger.Infof("pet: checking connection conn_id=%s, sessionID=%s", connID, pc.sessionID)
		if pc.sessionID == sessionID || sessionID == "broadcast" {
			logger.Infof("pet: matched! sending to conn_id=%s", connID)
			data, _ := json.Marshal(content)
			pc.writeJSON(Response{
				Status: pet.StatusOK,
				Action: pet.ActionChat,
				Data:   data,
			})
		}
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
