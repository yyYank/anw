package multiselect

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Item は選択肢の名前と、表示する理由。
type Item struct {
	Name   string
	Reason string
}

// Model は一覧から複数を選ぶ TUI。
type Model struct {
	title    string
	items    []Item
	marked   map[int]bool
	cursor   int
	selected bool
}

// New は title を見出しにして items から選ぶ Model を返す。
func New(title string, items []Item) Model {
	return Model{title: title, items: items, marked: map[int]bool{}}
}

// Init は tea.Model の実装。
func (m Model) Init() tea.Cmd { return nil }

// Update はキー入力でカーソルを動かし、Space で印の付け外し、Enter で決定、Esc / q / Ctrl+C で終了する。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Model は値で受け渡されるので、元の map を書き換えないよう作り直す
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
func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.title + "\n\n")
	for i, it := range m.items {
		cur := " "
		if i == m.cursor {
			cur = ">"
		}
		mark := " "
		if m.marked[i] {
			mark = "x"
		}
		fmt.Fprintf(&b, "%s [%s] %s  (%s)\n", cur, mark, it.Name, it.Reason)
	}
	return b.String()
}

// Selected は選ばれた要素の添字を返す。印がなければカーソル位置を返す。
func (m Model) Selected() []int {
	if !m.selected {
		return nil
	}
	if len(m.marked) == 0 {
		return []int{m.cursor}
	}
	var result []int
	for i := range m.items {
		if m.marked[i] {
			result = append(result, i)
		}
	}
	return result
}
