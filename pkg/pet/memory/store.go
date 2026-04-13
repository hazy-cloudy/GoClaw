package memory

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sipeed/picoclaw/pkg/logger"
)

// Store 记忆存储
// 使用SQLite数据库持久化存储角色的记忆信息
type Store struct {
	db *sql.DB
}

// Memory 记忆条目
// 代表角色的一条记忆
type Memory struct {
	ID          int64     // 记忆唯一ID
	CharacterID string    // 所属角色ID
	Content     string    // 记忆内容
	MemoryType  string    // 记忆类型
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
}

// 记忆类型常量
const (
	MemoryTypeConversation = "conversation" // 对话记忆
	MemoryTypePreference   = "preference"   // 偏好记忆
	MemoryTypeFact         = "fact"         // 事实记忆
)

// NewStore 创建记忆存储实例
// workspacePath: 工作区路径，数据库文件将创建在此目录下
func NewStore(workspacePath string) (*Store, error) {
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	dbPath := filepath.Join(workspacePath, "pet_memory.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	logger.Infof("memory: store initialized at %s", dbPath)
	return s, nil
}

// initSchema 初始化数据库表结构
func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS memories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		character_id TEXT NOT NULL,
		content TEXT NOT NULL,
		memory_type TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_character_memories ON memories(character_id);
	CREATE INDEX IF NOT EXISTS idx_character_type ON memories(character_id, memory_type);
	`
	_, err := s.db.Exec(schema)
	return err
}

// Add 添加新记忆
// characterID: 角色ID
// content: 记忆内容
// memoryType: 记忆类型（conversation/preference/fact）
func (s *Store) Add(characterID, content, memoryType string) (*Memory, error) {
	now := time.Now().Unix()

	result, err := s.db.Exec(
		"INSERT INTO memories (character_id, content, memory_type, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		characterID, content, memoryType, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert memory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &Memory{
		ID:         id,
		Content:    content,
		MemoryType: memoryType,
		CreatedAt:  time.Unix(now, 0),
		UpdatedAt:  time.Unix(now, 0),
	}, nil
}

// List 获取角色的所有记忆
// 按更新时间倒序排列
func (s *Store) List(characterID string) ([]*Memory, error) {
	rows, err := s.db.Query(
		"SELECT id, character_id, content, memory_type, created_at, updated_at FROM memories WHERE character_id = ? ORDER BY updated_at DESC",
		characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		var m Memory
		var createdAt, updatedAt int64
		if err := rows.Scan(&m.ID, &m.CharacterID, &m.Content, &m.MemoryType, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}
		m.CreatedAt = time.Unix(createdAt, 0)
		m.UpdatedAt = time.Unix(updatedAt, 0)
		memories = append(memories, &m)
	}

	return memories, nil
}

// ListByType 获取角色指定类型的所有记忆
// 按更新时间倒序排列
func (s *Store) ListByType(characterID, memoryType string) ([]*Memory, error) {
	rows, err := s.db.Query(
		"SELECT id, character_id, content, memory_type, created_at, updated_at FROM memories WHERE character_id = ? AND memory_type = ? ORDER BY updated_at DESC",
		characterID, memoryType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		var m Memory
		var createdAt, updatedAt int64
		if err := rows.Scan(&m.ID, &m.CharacterID, &m.Content, &m.MemoryType, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}
		m.CreatedAt = time.Unix(createdAt, 0)
		m.UpdatedAt = time.Unix(updatedAt, 0)
		memories = append(memories, &m)
	}

	return memories, nil
}

// Delete 删除指定记忆
func (s *Store) Delete(id int64) error {
	_, err := s.db.Exec("DELETE FROM memories WHERE id = ?", id)
	return err
}

// DeleteCharacter 删除角色的所有记忆
func (s *Store) DeleteCharacter(characterID string) error {
	_, err := s.db.Exec("DELETE FROM memories WHERE character_id = ?", characterID)
	return err
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	return s.db.Close()
}
