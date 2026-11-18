package envfile

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var files = []string{".env", "app/.env.local"}

// 実ファイルを読まず、ファイル名に応じた偽の内容を返す
func fakeLoad(path string) string {
	return "CONTENT_OF " + path
}

func press(m tea.Model, keys ...tea.KeyMsg) Picker {
	for _, k := range keys {
		m, _ = m.Update(k)
	}
	return m.(Picker)
}

var (
	keyDown  = tea.KeyMsg{Type: tea.KeyDown}
	keyEnter = tea.KeyMsg{Type: tea.KeyEnter}
	keyEsc   = tea.KeyMsg{Type: tea.KeyEsc}
)

// 一覧と、カーソル位置のファイルの内容が表示されること
func TestPicker_一覧とプレビューを表示する(t *testing.T) {
	view := NewPicker(files, fakeLoad, false).View()

	for _, s := range []string{".env", "app/.env.local", "CONTENT_OF .env"} {
		if !strings.Contains(view, s) {
			t.Errorf("%q が表示されていない:\n%s", s, view)
		}
	}
}

// カーソルを動かすとプレビューが切り替わること
func TestPicker_移動するとプレビューが切り替わる(t *testing.T) {
	view := press(NewPicker(files, fakeLoad, false), keyDown).View()

	if !strings.Contains(view, "CONTENT_OF app/.env.local") {
		t.Errorf("プレビューが切り替わっていない:\n%s", view)
	}
}

// 選択できる画面では Enter でカーソル位置のファイルが選ばれること
func TestPicker_Enterで選ぶ(t *testing.T) {
	m := press(NewPicker(files, fakeLoad, true), keyDown, keyEnter)

	got, ok := m.Selected()
	if !ok || got != "app/.env.local" {
		t.Errorf("got %q, ok=%v", got, ok)
	}
}

// 閲覧だけの画面では Enter を押しても選ばれないこと
func TestPicker_閲覧ではEnterで選ばない(t *testing.T) {
	m := press(NewPicker(files, fakeLoad, false), keyEnter)

	if _, ok := m.Selected(); ok {
		t.Error("選ばれてしまった")
	}
}

// Esc で何も選ばずに終わること
func TestPicker_Escでキャンセル(t *testing.T) {
	m := press(NewPicker(files, fakeLoad, true), keyEsc)

	if _, ok := m.Selected(); ok {
		t.Error("選ばれてしまった")
	}
}
