package aisession

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// Claude Code の jsonl を真似た偽のセッション
const claudeJSONL = `{"type":"permission-mode","permissionMode":"default","sessionId":"s1"}
{"type":"user","isSidechain":false,"message":{"role":"user","content":"ログインのバグを直して\n詳しくは issue を見て"},"timestamp":"2026-11-01T10:00:00.000Z","cwd":"/work/app","sessionId":"s1","gitBranch":"fix/login"}
{"type":"assistant","isSidechain":false,"message":{"role":"assistant","content":[{"type":"thinking","thinking":"考え中"},{"type":"text","text":"見てみます"}]},"timestamp":"2026-11-01T10:00:05.000Z","cwd":"/work/app","sessionId":"s1","gitBranch":"fix/login"}
{"type":"assistant","isSidechain":false,"message":{"role":"assistant","content":[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"/work/app/login.go"}}]},"timestamp":"2026-11-01T10:00:06.000Z","cwd":"/work/app","sessionId":"s1","gitBranch":"fix/login"}
{"type":"user","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":"package main"}]},"timestamp":"2026-11-01T10:00:07.000Z","cwd":"/work/app","sessionId":"s1","gitBranch":"fix/login"}
{"type":"user","isSidechain":false,"isMeta":true,"message":{"role":"user","content":[{"type":"text","text":"メタ情報"}]},"timestamp":"2026-11-01T10:00:08.000Z","cwd":"/work/app","sessionId":"s1","gitBranch":"fix/login"}
{"type":"assistant","isSidechain":true,"message":{"role":"assistant","content":[{"type":"text","text":"サブエージェント"}]},"timestamp":"2026-11-01T10:00:09.000Z","cwd":"/work/app","sessionId":"s1","gitBranch":"fix/login"}
{"type":"user","isSidechain":false,"message":{"role":"user","content":[{"type":"text","text":"ありがとう"}]},"timestamp":"2026-11-01T10:05:00.000Z","cwd":"/work/app","sessionId":"s1","gitBranch":"fix/login"}
not json
`

// セッションの ID・作業ディレクトリ・ブランチ・最終更新・最初の発言を読み取ること
func TestParseClaude_セッションの情報を読み取る(t *testing.T) {
	s, err := ParseClaude(strings.NewReader(claudeJSONL), "/home/.claude/projects/x/s1.jsonl")
	if err != nil {
		t.Fatal(err)
	}

	if s.Agent != "claude" || s.ID != "s1" || s.Path != "/home/.claude/projects/x/s1.jsonl" {
		t.Errorf("agent=%q id=%q path=%q", s.Agent, s.ID, s.Path)
	}
	if s.Cwd != "/work/app" || s.Branch != "fix/login" {
		t.Errorf("cwd=%q branch=%q", s.Cwd, s.Branch)
	}
	if want := time.Date(2026, 11, 1, 10, 5, 0, 0, time.UTC); !s.Updated.Equal(want) {
		t.Errorf("updated=%v, want %v", s.Updated, want)
	}
	if s.FirstPrompt != "ログインのバグを直して" {
		t.Errorf("first prompt=%q", s.FirstPrompt)
	}
}

// 発言だけを抜き出し、ツールの呼び出しは一行に畳み、結果・思考・メタ情報・サブエージェントは省くこと
func TestParseClaude_発言を抜き出す(t *testing.T) {
	s, err := ParseClaude(strings.NewReader(claudeJSONL), "s1.jsonl")
	if err != nil {
		t.Fatal(err)
	}

	want := []Message{
		{Role: RoleUser, Text: "ログインのバグを直して\n詳しくは issue を見て"},
		{Role: RoleAssistant, Text: "見てみます"},
		{Role: RoleTool, Text: "Read"},
		{Role: RoleUser, Text: "ありがとう"},
	}
	if !reflect.DeepEqual(s.Messages, want) {
		t.Errorf("got %+v\nwant %+v", s.Messages, want)
	}
}
