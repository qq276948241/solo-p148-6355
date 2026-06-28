package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadScoreStore_NoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "best_scores.json")
	store, err := LoadScoreStore(path)
	if err != nil {
		t.Fatalf("文件不存在时应返回空 ScoreStore，但报错: %v", err)
	}
	if store == nil {
		t.Fatal("文件不存在时不应返回 nil")
	}
	if len(store.Scores) != 0 {
		t.Fatalf("空 ScoreStore 的 Scores 应为空 map，实际: %v", store.Scores)
	}
}

func TestLoadScoreStore_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "best_scores.json")
	os.WriteFile(path, []byte("not json"), 0644)
	_, err := LoadScoreStore(path)
	if err == nil {
		t.Fatal("无效 JSON 应返回错误")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "best_scores.json")
	store := NewScoreStore()
	store.Update("第一关", 42)
	store.Update("第二关", 30)
	if err := store.Save(path); err != nil {
		t.Fatalf("保存失败: %v", err)
	}
	loaded, err := LoadScoreStore(path)
	if err != nil {
		t.Fatalf("加载失败: %v", err)
	}
	if len(loaded.Scores) != 2 {
		t.Fatalf("应加载2条记录，实际: %d", len(loaded.Scores))
	}
	if loaded.Scores["第一关"] != 42 {
		t.Errorf("第一关成绩应为42，实际: %d", loaded.Scores["第一关"])
	}
	if loaded.Scores["第二关"] != 30 {
		t.Errorf("第二关成绩应为30，实际: %d", loaded.Scores["第二关"])
	}
}

func TestBestDisplay_NoRecord(t *testing.T) {
	store := NewScoreStore()
	display := store.BestDisplay("不存在的关")
	if display != "--" {
		t.Errorf("无记录时应显示'--'，实际: %s", display)
	}
}

func TestBestDisplay_WithRecord(t *testing.T) {
	store := NewScoreStore()
	store.Update("第一关", 15)
	display := store.BestDisplay("第一关")
	if display != "15" {
		t.Errorf("有记录时应显示步数，期望'15'，实际: %s", display)
	}
}

func TestUpdate_FirstRecord(t *testing.T) {
	store := NewScoreStore()
	updated := store.Update("第一关", 100)
	if !updated {
		t.Error("首次记录应返回 true")
	}
	if store.Scores["第一关"] != 100 {
		t.Errorf("成绩应为100，实际: %d", store.Scores["第一关"])
	}
}

func TestUpdate_BetterScore(t *testing.T) {
	store := NewScoreStore()
	store.Update("第一关", 100)
	updated := store.Update("第一关", 50)
	if !updated {
		t.Error("更好成绩应返回 true")
	}
	if store.Scores["第一关"] != 50 {
		t.Errorf("成绩应更新为50，实际: %d", store.Scores["第一关"])
	}
}

func TestUpdate_WorseScore(t *testing.T) {
	store := NewScoreStore()
	store.Update("第一关", 50)
	updated := store.Update("第一关", 100)
	if updated {
		t.Error("更差成绩不应更新")
	}
	if store.Scores["第一关"] != 50 {
		t.Errorf("成绩应保持50，实际: %d", store.Scores["第一关"])
	}
}

func TestUpdate_SameScore(t *testing.T) {
	store := NewScoreStore()
	store.Update("第一关", 50)
	updated := store.Update("第一关", 50)
	if updated {
		t.Error("相同成绩不应更新")
	}
}

func TestBest_NoRecord(t *testing.T) {
	store := NewScoreStore()
	_, ok := store.Best("不存在的关")
	if ok {
		t.Error("无记录时 ok 应为 false")
	}
}

func TestBest_WithRecord(t *testing.T) {
	store := NewScoreStore()
	store.Update("第一关", 42)
	score, ok := store.Best("第一关")
	if !ok {
		t.Error("有记录时 ok 应为 true")
	}
	if score != 42 {
		t.Errorf("成绩应为42，实际: %d", score)
	}
}

func TestLoadScoreStore_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "best_scores.json")
	os.WriteFile(path, []byte("{}"), 0644)
	store, err := LoadScoreStore(path)
	if err != nil {
		t.Fatalf("空 JSON 对象不应报错: %v", err)
	}
	if store.Scores == nil {
		t.Error("Scores 不应为 nil")
	}
}

func TestSaveOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "best_scores.json")
	store := NewScoreStore()
	store.Update("第一关", 100)
	store.Save(path)
	store.Update("第一关", 50)
	store.Save(path)
	loaded, _ := LoadScoreStore(path)
	if loaded.Scores["第一关"] != 50 {
		t.Errorf("覆盖保存后成绩应为50，实际: %d", loaded.Scores["第一关"])
	}
}
