package aisession

import (
	"strings"
	"time"
)

// 発言の話し手
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	// RoleTool はツールの呼び出し。Text にはツール名だけを入れる
	RoleTool = "tool"
)

// Session はコーディングエージェントのセッション。エージェントごとの形式から変換した共通の形。
type Session struct {
	Agent       string
	ID          string
	Path        string
	Cwd         string
	Branch      string
	Updated     time.Time
	FirstPrompt string
	Messages    []Message
}

// Message はセッションの中の1つの発言。
type Message struct {
	Role string
	Text string
}

// add は発言を足す。最初のユーザーの発言は、一行目を FirstPrompt にする。
func (s *Session) add(role, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if role == RoleUser && s.FirstPrompt == "" {
		s.FirstPrompt, _, _ = strings.Cut(text, "\n")
	}
	s.Messages = append(s.Messages, Message{Role: role, Text: text})
}

// touch は ts が今の最終更新より新しければ更新する。
func (s *Session) touch(ts string) {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err == nil && t.After(s.Updated) {
		s.Updated = t
	}
}
