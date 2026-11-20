package aisession

import (
	"os"
	"path/filepath"
	"testing"
)

// テスト用のホームにセッションのファイルを作る
func writeSession(t *testing.T, home, rel, content string) {
	t.Helper()
	path := filepath.Join(home, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func userLine(id, cwd, ts, text string) string {
	return `{"type":"user","message":{"role":"user","content":"` + text + `"},"timestamp":"` + ts + `","cwd":"` + cwd + `","sessionId":"` + id + `"}` + "\n"
}

// 2種類のエージェントのセッションを集め、最終更新の新しい順に並べること
func TestCollect_集めて新しい順に並べる(t *testing.T) {
	home := t.TempDir()
	writeSession(t, home, ".claude/projects/-work-app/s1.jsonl", userLine("s1", "/work/app", "2026-11-01T10:00:00Z", "古い"))
	writeSession(t, home, ".claude/projects/-work-web/s2.jsonl", userLine("s2", "/work/web", "2026-11-03T10:00:00Z", "新しい"))
	writeSession(t, home, ".codex/sessions/2026/11/02/rollout-c1.jsonl", codexJSONL)

	got, err := Collect(home)
	if err != nil {
		t.Fatal(err)
	}

	var ids []string
	for _, s := range got {
		ids = append(ids, s.ID)
	}
	if want := "s2,c1,s1"; join(ids) != want {
		t.Errorf("got %s, want %s", join(ids), want)
	}
}

// 発言のないセッションと、サブディレクトリにあるサブエージェントのファイルは含めないこと
func TestCollect_発言のないものとサブエージェントは含めない(t *testing.T) {
	home := t.TempDir()
	writeSession(t, home, ".claude/projects/-work-app/s1.jsonl", userLine("s1", "/work/app", "2026-11-01T10:00:00Z", "本体"))
	writeSession(t, home, ".claude/projects/-work-app/empty.jsonl", `{"type":"permission-mode","sessionId":"e"}`+"\n")
	writeSession(t, home, ".claude/projects/-work-app/s1/subagents/a.jsonl", userLine("sub", "/work/app", "2026-11-05T10:00:00Z", "サブ"))

	got, err := Collect(home)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 || got[0].ID != "s1" {
		t.Errorf("got %+v", got)
	}
}

// どちらのエージェントのディレクトリもなければ、空で返すこと
func TestCollect_ディレクトリがなくてもよい(t *testing.T) {
	got, err := Collect(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %+v", got)
	}
}

func join(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += ","
		}
		out += s
	}
	return out
}
