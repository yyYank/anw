package branch

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Selector はブランチの一覧から、削除するものを複数選ぶ TUI。
type Selector struct {
	items    []Branch
	marked   map[int]bool
	cursor   int
	selected bool
}

// NewSelector は items から選ぶ Selector を返す。
func NewSelector(items []Branch) Selector {
	return Selector{items: items, marked: map[int]bool{}}
}

// Init は tea.Model の実装。
func (m Selector) Init() tea.Cmd { return nil }

// Update はキー入力でカーソルを動かし、Space で印の付け外し、Enter で決定、Esc / q / Ctrl+C で終了する。
func (m Selector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case " ":
		// Selector は値で受け渡されるので、元の map を書き換えないよう作り直す
		marked := make(map[int]bool, len(m.marked)+1)
		for i := range m.marked {
			marked[i] = true
		}
		if marked[m.cursor] {
			delete(marked, m.cursor)
		} else {
			marked[m.cursor] = true
		}
		m.marked = marked
	case "enter":
		m.selected = len(m.items) > 0
		return m, tea.Quit
	case "esc", "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// View は一覧を描画する。
func (m Selector) View() string {
	var b strings.Builder
	b.WriteString("削除するブランチを選んでください (↑↓/jk: 移動, Space: 選択, Enter: 削除, q: やめる)\n\n")
	for i, br := range m.items {
		cur := " "
		if i == m.cursor {
			cur = ">"
		}
		mark := " "
		if m.marked[i] {
			mark = "x"
		}
		fmt.Fprintf(&b, "%s [%s] %s  (%s)\n", cur, mark, br.Name, br.Reason())
	}
	return b.String()
}

// Selected は選ばれたブランチを返す。印がなければカーソル位置のブランチを返す。
func (m Selector) Selected() []Branch {
	if !m.selected {
		return nil
	}
	if len(m.marked) == 0 {
		return []Branch{m.items[m.cursor]}
	}
	var result []Branch
	for i, br := range m.items {
		if m.marked[i] {
			result = append(result, br)
		}
	}
	return result
}
