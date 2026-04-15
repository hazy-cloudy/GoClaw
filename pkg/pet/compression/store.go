package compression

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sipeed/picoclaw/pkg/logger"
)

// ConversationStore 对话存储
// 持久化存储对话记录，支持软删除
// 数据库文件：pet_conversation.db
type ConversationStore struct {
	db        *sql.DB             // SQLite 数据库连接
	threshold int                 // 触发压缩的消息数阈值
	callback  OnThresholdCallback // 达到阈值时的回调
	mu        sync.RWMutex        // 读写锁
}

// NewConversationStore 创建对话存储实例
// workspacePath: 工作区路径，数据库文件将创建在此目录下
// threshold: 触发压缩的消息数阈值
// callback: 达到阈值时的回调函数
func NewConversationStore(workspacePath string, threshold int, callback OnThresholdCallback) (*ConversationStore, error) {
	if threshold <= 0 {
		threshold = DefaultThreshold
	}

	// 确保目录存在
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 打开数据库
	dbPath := filepath.Join(workspacePath, "pet_conversation.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &ConversationStore{
		db:        db,
		threshold: threshold,
		callback:  callback,
	}

	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	logger.Infof("compression: conversation store initialized at %s", dbPath)
	return s, nil
}

// initSchema 初始化数据库表结构
// conversations 表字段说明：
//   - id: 自增主键
//   - character_id: 角色ID
//   - role: 对话角色 (user/pet)
//   - content: 对话内容
//   - compressed: 是否已压缩 (0/1)
//   - created_at: 创建时间戳
//   - deleted_at: 软删除时间戳 (NULL表示未删除)
func (s *ConversationStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS conversations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		character_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		compressed INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		deleted_at INTEGER DEFAULT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_character_conversations ON conversations(character_id);
	CREATE INDEX IF NOT EXISTS idx_compressed ON conversations(character_id, compressed);
	`
	_, err := s.db.Exec(schema)
	return err
}

// Add 添加新对话
// characterID: 角色ID
// role: 对话角色 (user/pet)
// content: 对话内容
// 添加成功后异步检查是否达到压缩阈值
func (s *ConversationStore) Add(characterID, role, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	_, err := s.db.Exec(
		"INSERT INTO conversations (character_id, role, content, created_at) VALUES (?, ?, ?, ?)",
		characterID, role, content, now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert conversation: %w", err)
	}

	// 异步检查是否达到压缩阈值
	go s.checkThreshold(characterID)
	return nil
}

// checkThreshold 检查是否达到压缩阈值
// 如果未压缩对话数达到阈值，触发回调
func (s *ConversationStore) checkThreshold(characterID string) {
	s.mu.RLock()
	count, err := s.getUncompressedCountLocked(characterID)
	s.mu.RUnlock()
	if err != nil {
		logger.Warnf("compression: failed to get uncompressed count: %v", err)
		return
	}

	// 达到阈值时触发回调
	if s.callback != nil && count >= s.threshold {
		entries, err := s.GetUncompressed(characterID)
		if err != nil {
			logger.Warnf("compression: failed to get uncompressed entries: %v", err)
			return
		}
		s.callback(characterID, entries)
	}
}

// getUncompressedCountLocked 获取未压缩对话数量（内部锁定版本）
// caller must hold at least a read lock
func (s *ConversationStore) getUncompressedCountLocked(characterID string) (int, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM conversations WHERE character_id = ? AND compressed = 0 AND deleted_at IS NULL",
		characterID,
	).Scan(&count)
	return count, err
}

// GetUncompressed 获取未压缩的对话列表
// characterID: 角色ID
// 返回按时间正序排列的未压缩对话
func (s *ConversationStore) GetUncompressed(characterID string) ([]*ConversationEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		"SELECT id, role, content, created_at FROM conversations WHERE character_id = ? AND compressed = 0 AND deleted_at IS NULL ORDER BY created_at ASC",
		characterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*ConversationEntry
	for rows.Next() {
		var e ConversationEntry
		var createdAt int64
		if err := rows.Scan(&e.ID, &e.Role, &e.Content, &createdAt); err != nil {
			return nil, err
		}
		e.Timestamp = time.Unix(createdAt, 0)
		entries = append(entries, &e)
	}
	return entries, nil
}

// GetAll 获取所有对话（包括已压缩的）
// characterID: 角色ID
// limit: 返回条数限制
// 返回按时间倒序排列的对话
func (s *ConversationStore) GetAll(characterID string, limit int) ([]*ConversationEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		"SELECT id, role, content, created_at FROM conversations WHERE character_id = ? AND deleted_at IS NULL ORDER BY created_at DESC LIMIT ?",
		characterID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*ConversationEntry
	for rows.Next() {
		var e ConversationEntry
		var createdAt int64
		if err := rows.Scan(&e.ID, &e.Role, &e.Content, &createdAt); err != nil {
			return nil, err
		}
		e.Timestamp = time.Unix(createdAt, 0)
		entries = append(entries, &e)
	}
	return entries, nil
}

// MarkCompressed 标记对话为已压缩
// ids: 要标记的对话ID列表
func (s *ConversationStore) MarkCompressed(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 构建批量更新语句
	query := "UPDATE conversations SET compressed = 1 WHERE id IN ("
	args := make([]any, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	_, err := s.db.Exec(query, args...)
	return err
}

// SoftDeleteOld 软删除旧的已压缩对话
// days: 保留天数，超过此天数的已压缩对话将被软删除
func (s *ConversationStore) SoftDeleteOld(days int) error {
	if days <= 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -days).Unix()
	_, err := s.db.Exec(
		"UPDATE conversations SET deleted_at = ? WHERE compressed = 1 AND deleted_at IS NULL AND created_at < ?",
		time.Now().Unix(), cutoff,
	)
	return err
}

// SetCallback 设置达到阈值时的回调函数
func (s *ConversationStore) SetCallback(callback OnThresholdCallback) {
	s.callback = callback
}

// Count 获取未压缩对话数量
// characterID: 角色ID
func (s *ConversationStore) Count(characterID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getUncompressedCountLocked(characterID)
}

// Threshold 获取当前阈值
func (s *ConversationStore) Threshold() int {
	return s.threshold
}

// Close 关闭数据库连接
func (s *ConversationStore) Close() error {
	return s.db.Close()
}
