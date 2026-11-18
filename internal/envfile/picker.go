package envfile

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Picker は env ファイルの一覧と、カーソル位置のファイルの内容を表示する TUI。
type Picker struct {
	items      []string
	load       func(path string) string
	selectable bool
	cursor     int
	selected   bool
}

// NewPicker は items を表示する Picker を返す。
// load はプレビューに表示する内容を返す。selectable が true なら Enter で選べる。
func NewPicker(items []string, load func(path string) string, selectable bool) Picker {
	return Picker{items: items, load: load, selectable: selectable}
}

// Init は tea.Model の実装。
func (m Picker) Init() tea.Cmd { return nil }

// Update はキー入力でカーソルを動かし、Enter で選択、Esc / q / Ctrl+C で終了する。
func (m Picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case "enter":
		if m.selectable && len(m.items) > 0 {
			m.selected = true
			return m, tea.Quit
		}
	case "esc", "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// View は一覧とプレビューを描画する。
func (m Picker) View() string {
	var b strings.Builder
	if m.selectable {
		b.WriteString("編集するファイルを選んでください (↑↓/jk: 移動, Enter: 編集, q: やめる)\n\n")
	} else {
		b.WriteString("env ファイル (↑↓/jk: 移動, q: 終了)\n\n")
	}
	for i, item := range m.items {
		mark := " "
		if i == m.cursor {
			mark = ">"
		}
		fmt.Fprintf(&b, "%s %s\n", mark, item)
	}
	if len(m.items) > 0 {
		fmt.Fprintf(&b, "\n--- %s ---\n%s\n", m.items[m.cursor], m.load(m.items[m.cursor]))
	}
	return b.String()
}

// Selected は選ばれたファイルを返す。選ばれていなければ ok は false。
func (m Picker) Selected() (string, bool) {
	if !m.selected {
		return "", false
	}
	return m.items[m.cursor], true
}
