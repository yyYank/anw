package snippetedit

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// saver は保存された内容を記録する。err があれば保存に失敗させる
type saver struct {
	saved []string
	err   error
}

func (s *saver) save(text string) error {
	if s.err != nil {
		return s.err
	}
	s.saved = append(s.saved, text)
	return nil
}

// press は keys を順に入力し、最後に返ったコマンドを返す。"\n" は Enter として入力する
func press(m *Model, keys ...tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	for _, k := range keys {
		var next tea.Model
		next, cmd = m.Update(k)
		*m = next.(Model)
	}
	return cmd
}

// typeText は s を1文字ずつ入力する
func typeText(m *Model, s string) tea.Cmd {
	var keys []tea.KeyMsg
	for _, r := range s {
		if r == '\n' {
			keys = append(keys, tea.KeyMsg{Type: tea.KeyEnter})
			continue
		}
		keys = append(keys, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return press(m, keys...)
}

var (
	esc   = tea.KeyMsg{Type: tea.KeyEsc}
	enter = tea.KeyMsg{Type: tea.KeyEnter}
	ctrlS = tea.KeyMsg{Type: tea.KeyCtrlS}
)

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

// 入力した内容を Ctrl+S で保存し、閉じずに編集を続けられること
func TestEditor_CtrlSで保存する(t *testing.T) {
	s := &saver{}
	m := New("morning", "", s.save)

	typeText(&m, "cd ~/Desktop\nanw tidy")
	if cmd := press(&m, ctrlS); isQuit(cmd) {
		t.Error("保存しただけで閉じた")
	}

	if len(s.saved) != 1 || s.saved[0] != "cd ~/Desktop\nanw tidy" {
		t.Errorf("saved = %q", s.saved)
	}
	if !m.Saved() {
		t.Error("保存したことになっていない")
	}
}

// Esc のあとの :w で保存し、:wq で保存して閉じること
func TestEditor_wで保存しwqで閉じる(t *testing.T) {
	s := &saver{}
	m := New("morning", "", s.save)

	typeText(&m, "cd /tmp")
	press(&m, esc)
	if cmd := typeText(&m, ":w\n"); isQuit(cmd) {
		t.Error(":w で閉じた")
	}
	typeText(&m, "o")
	typeText(&m, "anw tidy")
	press(&m, esc)
	if cmd := typeText(&m, ":wq\n"); !isQuit(cmd) {
		t.Error(":wq で閉じなかった")
	}

	if got := strings.Join(s.saved, "|"); got != "cd /tmp|cd /tmp\nanw tidy" {
		t.Errorf("saved = %q", got)
	}
}

// 保存に失敗したら、閉じずにエラーを表示すること
func TestEditor_保存できなければ閉じずにエラーを出す(t *testing.T) {
	s := &saver{err: errors.New("1行目: cd か anw で始めてください")}
	m := New("morning", "", s.save)

	typeText(&m, "rm memo.md")
	press(&m, esc)
	if cmd := typeText(&m, ":wq\n"); isQuit(cmd) {
		t.Error("保存できないのに閉じた")
	}

	if !strings.Contains(m.View(), "cd か anw で始めてください") {
		t.Errorf("エラーが表示されていない:\n%s", m.View())
	}
	if m.Saved() {
		t.Error("保存したことになっている")
	}
}

// 保存していない変更があれば :q では閉じず、:q! なら閉じること
func TestEditor_保存していなければqで閉じない(t *testing.T) {
	s := &saver{}
	m := New("morning", "", s.save)

	typeText(&m, "cd /tmp")
	press(&m, esc)
	if cmd := typeText(&m, ":q\n"); isQuit(cmd) {
		t.Error("保存していないのに :q で閉じた")
	}
	if cmd := typeText(&m, ":q!\n"); !isQuit(cmd) {
		t.Error(":q! で閉じなかった")
	}

	if len(s.saved) != 0 {
		t.Errorf("保存された: %q", s.saved)
	}
}

// 開いたときは保存された内容の末尾から書き足せること
func TestEditor_保存された内容から書き足す(t *testing.T) {
	s := &saver{}
	m := New("morning", "cd /tmp\nanw tidy\n", s.save)

	typeText(&m, "anw view --env")
	press(&m, ctrlS)

	if len(s.saved) != 1 || s.saved[0] != "cd /tmp\nanw tidy\nanw view --env" {
		t.Errorf("saved = %q", s.saved)
	}
	// 何も変えていなければ :q で閉じられる
	m = New("morning", "cd /tmp\n", s.save)
	press(&m, esc)
	if cmd := typeText(&m, ":q\n"); !isQuit(cmd) {
		t.Error("変更がないのに :q で閉じなかった")
	}
}

// ノーマルモードで hjkl で動き、x で1文字消し、i で入力に戻ること
func TestEditor_ノーマルモードで動いて消す(t *testing.T) {
	s := &saver{}
	m := New("morning", "", s.save)

	typeText(&m, "cd /tmpp\nanw tidyy")
	press(&m, esc)
	typeText(&m, "xk$x")
	typeText(&m, "0i")
	typeText(&m, "# ")
	press(&m, tea.KeyMsg{Type: tea.KeyBackspace}, tea.KeyMsg{Type: tea.KeyBackspace})
	press(&m, ctrlS)

	if len(s.saved) != 1 || s.saved[0] != "cd /tmp\nanw tidy" {
		t.Errorf("saved = %q", s.saved)
	}
}
