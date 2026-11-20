package aisession

import (
	"bufio"
	"encoding/json"
	"io"
)

// claudeLine は Claude Code の jsonl の1行のうち、使う項目だけ。
type claudeLine struct {
	Type        string `json:"type"`
	IsSidechain bool   `json:"isSidechain"`
	IsMeta      bool   `json:"isMeta"`
	SessionID   string `json:"sessionId"`
	Cwd         string `json:"cwd"`
	GitBranch   string `json:"gitBranch"`
	Timestamp   string `json:"timestamp"`
	Message     struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// claudeContent は message.content が配列のときの要素。
type claudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Name string `json:"name"`
}

// ParseClaude は Claude Code の jsonl を読み、Session にする。読めない行は飛ばす。
func ParseClaude(r io.Reader, path string) (Session, error) {
	s := Session{Agent: "claude", Path: path}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 64*1024*1024)
	for sc.Scan() {
		var l claudeLine
		if json.Unmarshal(sc.Bytes(), &l) != nil {
			continue
		}
		if l.Type != "user" && l.Type != "assistant" {
			continue
		}
		if l.IsSidechain || l.IsMeta {
			continue
		}
		if s.ID == "" {
			s.ID = l.SessionID
		}
		if l.Cwd != "" {
			s.Cwd = l.Cwd
		}
		if l.GitBranch != "" {
			s.Branch = l.GitBranch
		}
		s.touch(l.Timestamp)

		var text string
		if json.Unmarshal(l.Message.Content, &text) == nil {
			s.add(l.Type, text)
			continue
		}
		var contents []claudeContent
		if json.Unmarshal(l.Message.Content, &contents) != nil {
			continue
		}
		for _, c := range contents {
			switch c.Type {
			case "text":
				s.add(l.Type, c.Text)
			case "tool_use":
				s.add(RoleTool, c.Name)
			}
		}
	}
	return s, sc.Err()
}
