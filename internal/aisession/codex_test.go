package aisession

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// Codex の jsonl を真似た偽のセッション
const codexJSONL = `{"timestamp":"2026-11-02T09:00:00.000Z","type":"session_meta","payload":{"id":"c1","timestamp":"2026-11-02T09:00:00.000Z","cwd":"/work/api","git":{"branch":"feature/x","commit_hash":"abc"}}}
{"timestamp":"2026-11-02T09:00:01.000Z","type":"response_item","payload":{"type":"message","role":"developer","content":[{"type":"input_text","text":"システムの指示"}]}}
{"timestamp":"2026-11-02T09:00:02.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"# AGENTS.md instructions for /work/api\n..."}]}}
{"timestamp":"2026-11-02T09:00:03.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"<environment_context>\n<cwd>/work/api</cwd>\n</environment_context>"}]}}
{"timestamp":"2026-11-02T09:00:04.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"API のテストを足して"},{"type":"input_text","text":"境界値も"}]}}
{"timestamp":"2026-11-02T09:00:05.000Z","type":"response_item","payload":{"type":"reasoning","summary":[],"encrypted_content":"xxx"}}
{"timestamp":"2026-11-02T09:00:06.000Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","arguments":"{}","call_id":"f1"}}
{"timestamp":"2026-11-02T09:00:07.000Z","type":"response_item","payload":{"type":"function_call_output","call_id":"f1","output":[{"type":"input_text","text":"ok"}]}}
{"timestamp":"2026-11-02T09:00:08.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"apply_patch","input":"...","call_id":"f2"}}
{"timestamp":"2026-11-02T09:10:00.000Z","type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"追加しました"}]}}
{"timestamp":"2026-11-02T09:11:00.000Z","type":"event_msg","payload":{"type":"token_count"}}
`

// セッションの ID・作業ディレクトリ・ブランチ・最終更新・最初の発言を読み取ること
func TestParseCodex_セッションの情報を読み取る(t *testing.T) {
	s, err := ParseCodex(strings.NewReader(codexJSONL), "/home/.codex/sessions/c1.jsonl")
	if err != nil {
		t.Fatal(err)
	}

	if s.Agent != "codex" || s.ID != "c1" || s.Path != "/home/.codex/sessions/c1.jsonl" {
		t.Errorf("agent=%q id=%q path=%q", s.Agent, s.ID, s.Path)
	}
	if s.Cwd != "/work/api" || s.Branch != "feature/x" {
		t.Errorf("cwd=%q branch=%q", s.Cwd, s.Branch)
	}
	if want := time.Date(2026, 11, 2, 9, 11, 0, 0, time.UTC); !s.Updated.Equal(want) {
		t.Errorf("updated=%v, want %v", s.Updated, want)
	}
	if s.FirstPrompt != "API のテストを足して" {
		t.Errorf("first prompt=%q", s.FirstPrompt)
	}
}

// 発言だけを抜き出し、ツールの呼び出しは一行に畳み、差し込まれた指示・結果・思考は省くこと
func TestParseCodex_発言を抜き出す(t *testing.T) {
	s, err := ParseCodex(strings.NewReader(codexJSONL), "c1.jsonl")
	if err != nil {
		t.Fatal(err)
	}

	want := []Message{
		{Role: RoleUser, Text: "API のテストを足して\n境界値も"},
		{Role: RoleTool, Text: "exec_command"},
		{Role: RoleTool, Text: "apply_patch"},
		{Role: RoleAssistant, Text: "追加しました"},
	}
	if !reflect.DeepEqual(s.Messages, want) {
		t.Errorf("got %+v\nwant %+v", s.Messages, want)
	}
}
