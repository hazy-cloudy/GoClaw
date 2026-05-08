package activity

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Store struct {
	baseDir string
	mu      sync.Mutex
}

// NewStore 创建活动记录存储。
//
// 这里选择按月写 JSONL，而不是一开始就引入更复杂的数据库表，
// 是为了让第一阶段更容易落地和排查：
// - 文件可直接打开看
// - 出问题时方便人工检查
// - 后续真要迁移到 DB，也不影响上层接口
func NewStore(workspacePath string) (*Store, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("workspace path is empty")
	}
	baseDir := filepath.Join(workspacePath, "pet_activity")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create activity dir: %w", err)
	}
	return &Store{baseDir: baseDir}, nil
}

func (s *Store) monthFile(t time.Time) string {
	return filepath.Join(s.baseDir, t.Format("2006-01")+".jsonl")
}

// Append 追加一条活动记录。
//
// 这里用互斥锁的原因很简单：桌宠后端可能同时处理多个请求，
// 如果多个 goroutine 同时往一个 jsonl 文件里追加，容易把文件写坏。
func (s *Store) Append(event *Event) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.monthFile(event.CreatedAt), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open activity log: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(event); err != nil {
		return fmt.Errorf("encode event: %w", err)
	}
	return nil
}

// ListRange 查询一个时间区间内的活动记录。
//
// 主动性第一阶段的两个主要使用场景：
// - 最近 24 小时：判断当前用户活跃度、有没有未完成事项
// - 最近 7 天：生成 weekly_report
func (s *Store) ListRange(start, end time.Time) ([]*Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if end.Before(start) {
		return nil, fmt.Errorf("invalid range")
	}

	// 一个查询区间可能跨月，所以这里先把涉及到的月份文件都找出来。
	type monthKey struct {
		year  int
		month time.Month
	}
	months := make([]monthKey, 0, 4)
	cursor := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	last := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())
	for !cursor.After(last) {
		months = append(months, monthKey{year: cursor.Year(), month: cursor.Month()})
		cursor = cursor.AddDate(0, 1, 0)
	}

	var events []*Event
	for _, mk := range months {
		path := filepath.Join(s.baseDir, fmt.Sprintf("%04d-%02d.jsonl", mk.year, mk.month))
		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("open activity file: %w", err)
		}

		// 逐行读取 JSONL；坏行直接跳过，避免一条脏数据把整次查询打崩。
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var ev Event
			if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
				continue
			}
			if ev.CreatedAt.Before(start) || ev.CreatedAt.After(end) {
				continue
			}
			copyEv := ev
			events = append(events, &copyEv)
		}
		_ = f.Close()
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("scan activity file: %w", err)
		}
	}
	return events, nil
}
