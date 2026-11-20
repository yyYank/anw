package choose

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var items = []Item{
	{Name: "env", Desc: "配下の env ファイル"},
	{Name: "port", Desc: "LISTEN 中のポート"},
	{Name: "branch", Desc: "削除できるブランチ"},
}

func press(m tea.Model, keys ...tea.KeyMsg) Model {
	for _, k := range keys {
		m, _ = m.Update(k)
	}
	return m.(Model)
}

var (
	keyDown  = tea.KeyMsg{Type: tea.KeyDown}
	keyUp    = tea.KeyMsg{Type: tea.KeyUp}
	keyEnter = tea.KeyMsg{Type: tea.KeyEnter}
	keyEsc   = tea.KeyMsg{Type: tea.KeyEsc}
	keyJ     = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	keyK     = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	keyQ     = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
)

// 見出しと、対象の名前・説明が一覧に表示されること
func TestModel_見出しと名前と説明を表示する(t *testing.T) {
	view := New("view: 何を見ますか？", items).View()

	for _, s := range []string{"view: 何を見ますか？", "env", "配下の env ファイル", "branch", "削除できるブランチ"} {
		if !strings.Contains(view, s) {
			t.Errorf("%q が表示されていない:\n%s", s, view)
		}
	}
}

// ↓ / j で下に、↑ / k で上に移動し、Enter で決定すること
func TestModel_移動してEnterで決定する(t *testing.T) {
	tests := []struct {
		name string
		keys []tea.KeyMsg
		want string
	}{
		{"そのまま", nil, "env"},
		{"下", []tea.KeyMsg{keyDown}, "port"},
		{"j", []tea.KeyMsg{keyJ, keyJ}, "branch"},
		{"上", []tea.KeyMsg{keyDown, keyDown, keyUp}, "port"},
		{"k", []tea.KeyMsg{keyJ, keyK}, "env"},
		{"下の端", []tea.KeyMsg{keyDown, keyDown, keyDown}, "branch"},
		{"上の端", []tea.KeyMsg{keyUp}, "env"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := press(New("", items), append(tt.keys, keyEnter)...)

			got, ok := m.Chosen()
			if !ok || got != tt.want {
				t.Errorf("got %q, ok=%v, want %q", got, ok, tt.want)
			}
		})
	}
}

// q / Esc で何も選ばずに終わること
func TestModel_qとEscで終了する(t *testing.T) {
	for _, k := range []tea.KeyMsg{keyQ, keyEsc} {
		t.Run(k.String(), func(t *testing.T) {
			if _, ok := press(New("", items), k).Chosen(); ok {
				t.Error("選ばれてしまった")
			}
		})
	}
}
