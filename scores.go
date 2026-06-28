package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type ScoreStore struct {
	Scores map[string]int `json:"scores"`
}

func NewScoreStore() *ScoreStore {
	return &ScoreStore{Scores: make(map[string]int)}
}

func LoadScoreStore(path string) (*ScoreStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewScoreStore(), nil
		}
		return nil, fmt.Errorf("读取成绩文件失败: %w", err)
	}
	var store ScoreStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("解析成绩文件失败: %w", err)
	}
	if store.Scores == nil {
		store.Scores = make(map[string]int)
	}
	return &store, nil
}

func (s *ScoreStore) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化成绩失败: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入成绩文件失败: %w", err)
	}
	return nil
}

func (s *ScoreStore) Best(levelName string) (int, bool) {
	score, ok := s.Scores[levelName]
	return score, ok
}

func (s *ScoreStore) BestDisplay(levelName string) string {
	score, ok := s.Scores[levelName]
	if !ok {
		return "--"
	}
	return fmt.Sprintf("%d", score)
}

func (s *ScoreStore) Update(levelName string, steps int) bool {
	best, exists := s.Scores[levelName]
	if !exists || steps < best {
		s.Scores[levelName] = steps
		return true
	}
	return false
}
