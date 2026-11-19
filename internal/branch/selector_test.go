package branch

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var items = []Branch{
	{Name: "feature/a", Merged: true},
	{Name: "feature/b", Gone: true},
	{Name: "feature/c", Merged: true},
}

func press(m tea.Model, keys ...tea.KeyMsg) Selector {
	for _, k := range keys {
		m, _ = m.Update(k)
	}
	return m.(Selector)
}

var (
	keyDown  = tea.KeyMsg{Type: tea.KeyDown}
	keySpace = tea.KeyMsg{Type: tea.KeySpace}
	keyEnter = tea.KeyMsg{Type: tea.KeyEnter}
	keyEsc   = tea.KeyMsg{Type: tea.KeyEsc}
)

// Space で印を付けたブランチを、Enter でまとめて選べること
func TestSelector_Spaceで複数選ぶ(t *testing.T) {
	m := press(NewSelector(items), keySpace, keyDown, keyDown, keySpace, keyEnter)

	want := []Branch{items[0], items[2]}
	if got := m.Selected(); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// Space をもう一度押すと印が外れること
func TestSelector_Spaceで印を外す(t *testing.T) {
	m := press(NewSelector(items), keySpace, keySpace, keyDown, keySpace, keyEnter)

	want := []Branch{items[1]}
	if got := m.Selected(); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// 印を付けずに Enter を押すと、カーソル位置のブランチが選ばれること
func TestSelector_印がなければカーソル位置を選ぶ(t *testing.T) {
	m := press(NewSelector(items), keyDown, keyEnter)

	want := []Branch{items[1]}
	if got := m.Selected(); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// Esc では何も選ばずに終わること
func TestSelector_Escでキャンセル(t *testing.T) {
	m := press(NewSelector(items), keySpace, keyEsc)

	if got := m.Selected(); len(got) != 0 {
		t.Errorf("選ばれてしまった: %+v", got)
	}
}

// 一覧にブランチ名と理由、印の有無が表示されること
func TestSelector_一覧を表示する(t *testing.T) {
	view := press(NewSelector(items), keySpace).View()

	for _, s := range []string{"[x] feature/a", "[ ] feature/b", "リモート削除済み"} {
		if !strings.Contains(view, s) {
			t.Errorf("%q が表示されていない:\n%s", s, view)
		}
	}
}
