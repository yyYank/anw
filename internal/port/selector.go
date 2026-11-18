package port

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Selector は Listener の一覧から1つを選ぶ TUI。
type Selector struct {
	items    []Listener
	cursor   int
	selected bool
}

// NewSelector は items を選ぶ Selector を返す。
func NewSelector(items []Listener) Selector {
	return Selector{items: items}
}

// Init は tea.Model の実装。
func (m Selector) Init() tea.Cmd { return nil }

// Update はキー入力でカーソルを動かし、Enter で選択、Esc / q / Ctrl+C で終了する。
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
	b.WriteString("kill するプロセスを選んでください (↑↓/jk: 移動, Enter: kill, q: やめる)\n\n")
	fmt.Fprintf(&b, "  %-6s %-7s %-16s %s\n", "PORT", "PID", "NAME", "COMMAND")
	for i, l := range m.items {
		mark := " "
		if i == m.cursor {
			mark = ">"
		}
		fmt.Fprintf(&b, "%s %-6d %-7d %-16s %s\n", mark, l.Port, l.PID, l.Name, l.Command)
	}
	return b.String()
}

// Selected は選ばれた Listener を返す。選ばれていなければ ok は false。
func (m Selector) Selected() (Listener, bool) {
	if !m.selected {
		return Listener{}, false
	}
	return m.items[m.cursor], true
}
