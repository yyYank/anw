package port

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var sample = []Listener{
	{Port: 3000, PID: 873, Name: "node", Command: "node server.js"},
	{Port: 8080, PID: 900, Name: "java", Command: "java -jar app.jar"},
}

func press(m tea.Model, keys ...tea.KeyMsg) tea.Model {
	for _, k := range keys {
		m, _ = m.Update(k)
	}
	return m
}

var (
	keyDown  = tea.KeyMsg{Type: tea.KeyDown}
	keyUp    = tea.KeyMsg{Type: tea.KeyUp}
	keyEnter = tea.KeyMsg{Type: tea.KeyEnter}
	keyEsc   = tea.KeyMsg{Type: tea.KeyEsc}
	keyJ     = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
)

// 下に移動して Enter を押すと、その行が選ばれること
func TestSelector_下に移動してEnterで選ぶ(t *testing.T) {
	m := press(NewSelector(sample), keyDown, keyEnter).(Selector)

	got, ok := m.Selected()
	if !ok || got.PID != 900 {
		t.Errorf("got %+v, ok=%v", got, ok)
	}
}

// j でも下に移動できること
func TestSelector_jで下に移動(t *testing.T) {
	m := press(NewSelector(sample), keyJ, keyEnter).(Selector)

	if got, _ := m.Selected(); got.PID != 900 {
		t.Errorf("got %+v", got)
	}
}

// 一覧の端より先には移動しないこと
func TestSelector_端より先には移動しない(t *testing.T) {
	m := press(NewSelector(sample), keyUp, keyEnter).(Selector)

	if got, _ := m.Selected(); got.PID != 873 {
		t.Errorf("got %+v", got)
	}
}

// Esc で何も選ばずに終わること
func TestSelector_Escでキャンセル(t *testing.T) {
	m := press(NewSelector(sample), keyEsc).(Selector)

	if _, ok := m.Selected(); ok {
		t.Error("選ばれてしまった")
	}
}

// 一覧にポート番号・PID・プロセス名・コマンドが表示されること
func TestSelector_一覧を表示する(t *testing.T) {
	view := NewSelector(sample).View()

	for _, s := range []string{"3000", "873", "node", "node server.js", "8080", "java -jar app.jar"} {
		if !strings.Contains(view, s) {
			t.Errorf("%q が表示されていない:\n%s", s, view)
		}
	}
}
