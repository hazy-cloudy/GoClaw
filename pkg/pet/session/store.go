package session

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sipeed/picoclaw/pkg/logger"
)

// Store 会话存储
// 使用SQLite数据库持久化存储会话和消息
type Store struct {
	db *sql.DB
}

// Session 会话
type Session struct {
	ID           int64     // 会话ID
	CharacterID  string    // 角色ID
	Title        string    // 会话标题
	Active       bool      // 是否为当前活跃会话
	CreatedAt    time.Time // 创建时间
	UpdatedAt    time.Time // 更新时间
}

// Message 消息
type Message struct {
	ID        int64     // 消息ID
	SessionID int64     // 会话ID
	Role      string    // 角色 (user/assistant)
	Content   string    // 内容
	Timestamp time.Time // 时间戳
}

// NewStore 创建会话存储实例
func NewStore(workspacePath string) (*Store, error) {
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	dbPath := filepath.Join(workspacePath, "pet_sessions.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 配置 SQLite
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		logger.Warnf("session: failed to enable WAL: %v", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		logger.Warnf("session: failed to set busy_timeout: %v", err)
	}

	s := &Store{db: db}

	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	logger.Infof("session: store initialized at %s", dbPath)
	return s, nil
}

// initSchema 初始化数据库表结构
func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		character_id TEXT NOT NULL,
		title TEXT NOT NULL DEFAULT '新对话',
		active INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_character_sessions ON sessions(character_id);
	CREATE INDEX IF NOT EXISTS idx_active_session ON sessions(character_id, active);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_session_messages ON messages(session_id);
	CREATE INDEX IF NOT EXISTS idx_message_timestamp ON messages(session_id, timestamp);
	`
	_, err := s.db.Exec(schema)
	return err
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// CreateSession 创建新会话
func (s *Store) CreateSession(characterID string) (*Session, error) {
	// 先将当前活跃会话设为非活跃
	_, err := s.db.Exec(
		"UPDATE sessions SET active = 0 WHERE character_id = ? AND active = 1",
		characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate sessions: %w", err)
	}

	now := time.Now().Unix()
	result, err := s.db.Exec(
		"INSERT INTO sessions (character_id, title, active, created_at, updated_at) VALUES (?, ?, 1, ?, ?)",
		characterID, "新对话", now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &Session{
		ID:          id,
		CharacterID: characterID,
		Title:       "新对话",
		Active:      true,
		CreatedAt:   time.Unix(now, 0),
		UpdatedAt:   time.Unix(now, 0),
	}, nil
}

// GetActiveSession 获取角色的当前活跃会话
func (s *Store) GetActiveSession(characterID string) (*Session, error) {
	var session Session
	var createdAt, updatedAt int64
	var active int

	err := s.db.QueryRow(
		"SELECT id, character_id, title, active, created_at, updated_at FROM sessions WHERE character_id = ? AND active = 1",
		characterID,
	).Scan(&session.ID, &session.CharacterID, &session.Title, &active, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	session.Active = active == 1
	session.CreatedAt = time.Unix(createdAt, 0)
	session.UpdatedAt = time.Unix(updatedAt, 0)
	return &session, nil
}

// GetOrCreateActiveSession 获取或创建活跃会话
func (s *Store) GetOrCreateActiveSession(characterID string) (*Session, error) {
	session, err := s.GetActiveSession(characterID)
	if err != nil {
		return nil, err
	}
	if session != nil {
		return session, nil
	}
	return s.CreateSession(characterID)
}

// ListSessions 列出角色的所有会话
func (s *Store) ListSessions(characterID string) ([]*Session, error) {
	rows, err := s.db.Query(
		"SELECT id, character_id, title, active, created_at, updated_at FROM sessions WHERE character_id = ? ORDER BY updated_at DESC",
		characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		var createdAt, updatedAt int64
		var active int

		if err := rows.Scan(&session.ID, &session.CharacterID, &session.Title, &active, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		session.Active = active == 1
		session.CreatedAt = time.Unix(createdAt, 0)
		session.UpdatedAt = time.Unix(updatedAt, 0)
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// SetActiveSession 设置活跃会话
func (s *Store) SetActiveSession(sessionID int64, characterID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 先将所有会话设为非活跃
	_, err = tx.Exec("UPDATE sessions SET active = 0 WHERE character_id = ?", characterID)
	if err != nil {
		return fmt.Errorf("failed to deactivate sessions: %w", err)
	}

	// 将目标会话设为活跃
	result, err := tx.Exec("UPDATE sessions SET active = 1 WHERE id = ? AND character_id = ?", sessionID, characterID)
	if err != nil {
		return fmt.Errorf("failed to activate session: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("session not found")
	}

	return tx.Commit()
}

// UpdateSessionTitle 更新会话标题
func (s *Store) UpdateSessionTitle(sessionID int64, title string) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		"UPDATE sessions SET title = ?, updated_at = ? WHERE id = ?",
		title, now, sessionID,
	)
	return err
}

// DeleteSession 删除会话及其所有消息
func (s *Store) DeleteSession(sessionID int64) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

// AddMessage 添加消息
func (s *Store) AddMessage(sessionID int64, role, content string) (*Message, error) {
	now := time.Now().Unix()
	result, err := s.db.Exec(
		"INSERT INTO messages (session_id, role, content, timestamp) VALUES (?, ?, ?, ?)",
		sessionID, role, content, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// 更新会话的更新时间
	_, err = s.db.Exec("UPDATE sessions SET updated_at = ? WHERE id = ?", now, sessionID)
	if err != nil {
		logger.Warnf("session: failed to update session timestamp: %v", err)
	}

	return &Message{
		ID:        id,
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Timestamp: time.Unix(now, 0),
	}, nil
}

// GetMessages 获取会话的所有消息
func (s *Store) GetMessages(sessionID int64) ([]*Message, error) {
	rows, err := s.db.Query(
		"SELECT id, session_id, role, content, timestamp FROM messages WHERE session_id = ? ORDER BY timestamp ASC",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		var timestamp int64

		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.Timestamp = time.Unix(timestamp, 0)
		messages = append(messages, &msg)
	}

	return messages, nil
}

// GetRecentMessages 获取会话的最近N条消息
func (s *Store) GetRecentMessages(sessionID int64, limit int) ([]*Message, error) {
	rows, err := s.db.Query(
		"SELECT id, session_id, role, content, timestamp FROM messages WHERE session_id = ? ORDER BY timestamp DESC LIMIT ?",
		sessionID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		var timestamp int64

		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.Timestamp = time.Unix(timestamp, 0)
		messages = append(messages, &msg)
	}

	// 反转数组以按时间正序返回
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// DeleteMessage 删除消息
func (s *Store) DeleteMessage(messageID int64) error {
	_, err := s.db.Exec("DELETE FROM messages WHERE id = ?", messageID)
	return err
}

// GetSessionByID 根据ID获取会话
func (s *Store) GetSessionByID(sessionID int64) (*Session, error) {
	var session Session
	var createdAt, updatedAt int64
	var active int

	err := s.db.QueryRow(
		"SELECT id, character_id, title, active, created_at, updated_at FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&session.ID, &session.CharacterID, &session.Title, &active, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	session.Active = active == 1
	session.CreatedAt = time.Unix(createdAt, 0)
	session.UpdatedAt = time.Unix(updatedAt, 0)
	return &session, nil
}
