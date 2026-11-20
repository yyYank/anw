package branch

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yyYank/anw/internal/multiselect"
)

// Selector はブランチの一覧から、削除するものを複数選ぶ TUI。
type Selector struct {
	items []Branch
	m     multiselect.Model
}

// NewSelector は items から選ぶ Selector を返す。
func NewSelector(items []Branch) Selector {
	ms := make([]multiselect.Item, len(items))
	for i, b := range items {
		ms[i] = multiselect.Item{Name: b.Name, Reason: b.Reason()}
	}
	title := "削除するブランチを選んでください (↑↓/jk: 移動, Space: 選択, Enter: 削除, q: やめる)"
	return Selector{items: items, m: multiselect.New(title, ms)}
}

// Init は tea.Model の実装。
func (s Selector) Init() tea.Cmd { return nil }

// Update は multiselect.Model に任せる。
func (s Selector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := s.m.Update(msg)
	s.m = m.(multiselect.Model)
	return s, cmd
}

// View は一覧を描画する。
func (s Selector) View() string { return s.m.View() }

// Selected は選ばれたブランチを返す。印がなければカーソル位置のブランチを返す。
func (s Selector) Selected() []Branch {
	var result []Branch
	for _, i := range s.m.Selected() {
		result = append(result, s.items[i])
	}
	return result
}
