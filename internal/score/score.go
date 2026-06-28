package score

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ScoreBoard struct {
	path   string
	scores map[string]int
}

type scoreFile struct {
	Scores map[string]int `json:"scores"`
}

func NewScoreBoard(path string) *ScoreBoard {
	return &ScoreBoard{
		path:   path,
		scores: make(map[string]int),
	}
}

func Load(path string) (*ScoreBoard, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewScoreBoard(path), nil
		}
		return nil, fmt.Errorf("读取成绩文件失败: %w", err)
	}
	var file scoreFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("解析成绩文件失败: %w", err)
	}
	if file.Scores == nil {
		file.Scores = make(map[string]int)
	}
	return &ScoreBoard{
		path:   path,
		scores: file.Scores,
	}, nil
}

func (sb *ScoreBoard) Save() error {
	data, err := json.MarshalIndent(scoreFile{Scores: sb.scores}, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化成绩失败: %w", err)
	}

	dir := filepath.Dir(sb.path)
	tmp, err := os.CreateTemp(dir, "best_scores_*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("写入临时文件失败: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("同步临时文件失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}
	if err := os.Rename(tmpPath, sb.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("替换成绩文件失败: %w", err)
	}
	return nil
}

func (sb *ScoreBoard) GetBest(levelName string) (int, bool) {
	score, ok := sb.scores[levelName]
	return score, ok
}

func (sb *ScoreBoard) BestDisplay(levelName string) string {
	score, ok := sb.scores[levelName]
	if !ok {
		return "--"
	}
	return fmt.Sprintf("%d", score)
}

func (sb *ScoreBoard) Update(levelName string, steps int) bool {
	best, exists := sb.scores[levelName]
	if !exists || steps < best {
		sb.scores[levelName] = steps
		return true
	}
	return false
}
