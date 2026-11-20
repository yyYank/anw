package choose

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle = lipgloss.NewStyle().Reverse(true)
	descStyle   = lipgloss.NewStyle().Faint(true)
	helpStyle   = lipgloss.NewStyle().Faint(true)
)

// Item は選択肢の名前と一行の説明。
type Item struct {
	Name string
	Desc string
}

// Model は Item の一覧から1つを選ぶ TUI。
type Model struct {
	title  string
	items  []Item
	cursor int
	chosen bool
}

// New は title を見出しにして items から選ぶ Model を返す。
func New(title string, items []Item) Model {
	return Model{title: title, items: items}
}

// Init は tea.Model の実装。
func (m Model) Init() tea.Cmd { return nil }

// Update は ↑↓ / jk で移動、Enter で決定、q / Esc / Ctrl+C で終了する。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, max(len(m.items)-1, 0))
	case "enter":
		m.chosen = len(m.items) > 0
		return m, tea.Quit
	case "esc", "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// View は見出しと一覧を描画する。
func (m Model) View() string {
	width := 0
	for _, it := range m.items {
		width = max(width, lipgloss.Width(it.Name))
	}

	var b strings.Builder
	if m.title != "" {
		b.WriteString(m.title + "\n\n")
	}
	for i, it := range m.items {
		name := fmt.Sprintf("%-*s", width, it.Name)
		if i == m.cursor {
			b.WriteString(cursorStyle.Render("> "+name) + "  " + descStyle.Render(it.Desc) + "\n")
		} else {
			b.WriteString("  " + name + "  " + descStyle.Render(it.Desc) + "\n")
		}
	}
	b.WriteString("\n" + helpStyle.Render("↑↓/jk: 移動  Enter: 決定  q: 終了"))
	return b.String()
}

// Chosen は選ばれた Item の名前を返す。選ばれていなければ ok は false。
func (m Model) Chosen() (string, bool) {
	if !m.chosen {
		return "", false
	}
	return m.items[m.cursor].Name, true
}
