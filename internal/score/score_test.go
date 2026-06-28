package score

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "best_scores.json")
	sb, err := Load(path)
	if err != nil {
		t.Fatalf("文件不存在时应返回空 ScoreBoard，但报错: %v", err)
	}
	if sb == nil {
		t.Fatal("文件不存在时不应返回 nil")
	}
	if sb == nil || len(sb.scores) != 0 {
		t.Fatalf("空 ScoreBoard 的 scores 应为空 map，实际: %v", sb.scores)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "best_scores.json")
	os.WriteFile(path, []byte("not json"), 0644)
	_, err := Load(path)
	if err == nil {
		t.Fatal("无效 JSON 应返回错误")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "best_scores.json")
	sb := NewScoreBoard(path)
	sb.Update("第一关", 42)
	sb.Update("第二关", 30)
	if err := sb.Save(); err != nil {
		t.Fatalf("保存失败: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("加载失败: %v", err)
	}
	if len(loaded.scores) != 2 {
		t.Fatalf("应加载2条记录，实际: %d", len(loaded.scores))
	}
	if loaded.scores["第一关"] != 42 {
		t.Errorf("第一关成绩应为42，实际: %d", loaded.scores["第一关"])
	}
	if loaded.scores["第二关"] != 30 {
		t.Errorf("第二关成绩应为30，实际: %d", loaded.scores["第二关"])
	}
}

func TestSave_Atomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "best_scores.json")
	sb := NewScoreBoard(path)
	sb.Update("第一关", 100)
	if err := sb.Save(); err != nil {
		t.Fatalf("第一次保存失败: %v", err)
	}
	origData, _ := os.ReadFile(path)
	sb.Update("第一关", 50)
	if err := sb.Save(); err != nil {
		t.Fatalf("第二次保存失败: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	jsonCount := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			jsonCount++
		}
	}
	if jsonCount != 1 {
		t.Errorf("目录里应只剩1个 .json 文件，实际: %d", jsonCount)
	}
	loaded, _ := Load(path)
	if loaded.scores["第一关"] != 50 {
		t.Errorf("最终成绩应为50，实际: %d", loaded.scores["第一关"])
	}
	_ = origData
}

func TestBestDisplay_NoRecord(t *testing.T) {
	sb := NewScoreBoard("dummy.json")
	display := sb.BestDisplay("不存在的关")
	if display != "--" {
		t.Errorf("无记录时应显示'--'，实际: %s", display)
	}
}

func TestBestDisplay_WithRecord(t *testing.T) {
	sb := NewScoreBoard("dummy.json")
	sb.Update("第一关", 15)
	display := sb.BestDisplay("第一关")
	if display != "15" {
		t.Errorf("有记录时应显示步数，期望'15'，实际: %s", display)
	}
}

func TestUpdate_FirstRecord(t *testing.T) {
	sb := NewScoreBoard("dummy.json")
	updated := sb.Update("第一关", 100)
	if !updated {
		t.Error("首次记录应返回 true")
	}
	if sb.scores["第一关"] != 100 {
		t.Errorf("成绩应为100，实际: %d", sb.scores["第一关"])
	}
}

func TestUpdate_BetterScore(t *testing.T) {
	sb := NewScoreBoard("dummy.json")
	sb.Update("第一关", 100)
	updated := sb.Update("第一关", 50)
	if !updated {
		t.Error("更好成绩应返回 true")
	}
	if sb.scores["第一关"] != 50 {
		t.Errorf("成绩应更新为50，实际: %d", sb.scores["第一关"])
	}
}

func TestUpdate_WorseScore(t *testing.T) {
	sb := NewScoreBoard("dummy.json")
	sb.Update("第一关", 50)
	updated := sb.Update("第一关", 100)
	if updated {
		t.Error("更差成绩不应更新")
	}
	if sb.scores["第一关"] != 50 {
		t.Errorf("成绩应保持50，实际: %d", sb.scores["第一关"])
	}
}

func TestUpdate_SameScore(t *testing.T) {
	sb := NewScoreBoard("dummy.json")
	sb.Update("第一关", 50)
	updated := sb.Update("第一关", 50)
	if updated {
		t.Error("相同成绩不应更新")
	}
}

func TestGetBest_NoRecord(t *testing.T) {
	sb := NewScoreBoard("dummy.json")
	_, ok := sb.GetBest("不存在的关")
	if ok {
		t.Error("无记录时 ok 应为 false")
	}
}

func TestGetBest_WithRecord(t *testing.T) {
	sb := NewScoreBoard("dummy.json")
	sb.Update("第一关", 42)
	score, ok := sb.GetBest("第一关")
	if !ok {
		t.Error("有记录时 ok 应为 true")
	}
	if score != 42 {
		t.Errorf("成绩应为42，实际: %d", score)
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "best_scores.json")
	os.WriteFile(path, []byte("{}"), 0644)
	sb, err := Load(path)
	if err != nil {
		t.Fatalf("空 JSON 对象不应报错: %v", err)
	}
	if sb.scores == nil {
		t.Error("scores 不应为 nil")
	}
}

func TestSaveOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "best_scores.json")
	sb := NewScoreBoard(path)
	sb.Update("第一关", 100)
	sb.Save()
	sb.Update("第一关", 50)
	sb.Save()
	loaded, _ := Load(path)
	if loaded.scores["第一关"] != 50 {
		t.Errorf("覆盖保存后成绩应为50，实际: %d", loaded.scores["第一关"])
	}
}
