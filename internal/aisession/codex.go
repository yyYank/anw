package aisession

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

// codexLine は Codex の jsonl の1行のうち、使う項目だけ。
type codexLine struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Payload   struct {
		Type    string `json:"type"`
		ID      string `json:"id"`
		Cwd     string `json:"cwd"`
		Role    string `json:"role"`
		Name    string `json:"name"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Git struct {
			Branch string `json:"branch"`
		} `json:"git"`
	} `json:"payload"`
}

// codexInjected は Codex が user の発言として差し込む指示や環境情報の書き出し。
var codexInjected = []string{"<", "# AGENTS.md instructions"}

// ParseCodex は Codex の jsonl を読み、Session にする。読めない行は飛ばす。
func ParseCodex(r io.Reader, path string) (Session, error) {
	s := Session{Agent: "codex", Path: path}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 64*1024*1024)
	for sc.Scan() {
		var l codexLine
		if json.Unmarshal(sc.Bytes(), &l) != nil {
			continue
		}
		s.touch(l.Timestamp)
		p := l.Payload
		switch {
		case l.Type == "session_meta":
			s.ID, s.Cwd, s.Branch = p.ID, p.Cwd, p.Git.Branch
		case l.Type != "response_item":
		case p.Type == "function_call" || p.Type == "custom_tool_call":
			s.add(RoleTool, p.Name)
		case p.Type == "message" && (p.Role == RoleUser || p.Role == RoleAssistant):
			var texts []string
			for _, c := range p.Content {
				texts = append(texts, c.Text)
			}
			text := strings.Join(texts, "\n")
			if p.Role == RoleUser && isCodexInjected(text) {
				continue
			}
			s.add(p.Role, text)
		}
	}
	return s, sc.Err()
}

func isCodexInjected(text string) bool {
	text = strings.TrimSpace(text)
	for _, prefix := range codexInjected {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}
