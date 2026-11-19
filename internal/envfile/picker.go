package envfile

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// ウィンドウサイズがまだ分からないときに使う幅
const defaultWidth = 80

var (
	paneStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	titleStyle  = lipgloss.NewStyle().Bold(true)
	cursorStyle = lipgloss.NewStyle().Reverse(true)
	helpStyle   = lipgloss.NewStyle().Faint(true)
)

// Picker は env ファイルの一覧と、カーソル位置のファイルの内容を表示する TUI。
type Picker struct {
	items      []string
	load       func(path string) string
	selectable bool
	cursor     int
	selected   bool
	width      int
	height     int
}

// NewPicker は items を表示する Picker を返す。
// load はプレビューに表示する内容を返す。selectable が true なら Enter で選べる。
func NewPicker(items []string, load func(path string) string, selectable bool) Picker {
	return Picker{items: items, load: load, selectable: selectable}
}

// Init は tea.Model の実装。
func (m Picker) Init() tea.Cmd { return nil }

// Update はキー入力でカーソルを動かし、Enter で選択、Esc / q / Ctrl+C で終了する。
// ウィンドウサイズの変更を受け取ると、描画をその大きさに合わせる。
func (m Picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = size.Width, size.Height
		return m, nil
	}
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

// View は左に一覧、右にカーソル位置のファイルのプレビューを並べて描画する。
func (m Picker) View() string {
	width := m.width
	if width <= 0 {
		width = defaultWidth
	}
	help := "↑↓/jk: 移動  q: 終了"
	if m.selectable {
		help = "↑↓/jk: 移動  Enter: 編集  q: やめる"
	}

	// 枠とパディングで左右それぞれ4桁使う
	const frame = 4
	listWidth := 0
	for _, item := range m.items {
		listWidth = max(listWidth, lipgloss.Width(item)+2)
	}
	listWidth = min(listWidth, (width-2*frame)/2)
	previewWidth := width - 2*frame - listWidth

	// 高さが分かっていれば、枠2行・タイトル1行・ヘルプ1行を除いた行数に収める
	rows := -1
	if m.height > 0 {
		rows = max(m.height-4, 1)
	}

	list := []string{titleStyle.Render("env ファイル")}
	start, end := window(len(m.items), m.cursor, rows)
	for i := start; i < end; i++ {
		line := "  " + m.items[i]
		if i == m.cursor {
			line = cursorStyle.Render("> " + m.items[i])
		}
		list = append(list, line)
	}

	preview := []string{}
	if len(m.items) > 0 {
		preview = append(preview, titleStyle.Render(m.items[m.cursor]))
		content := strings.Split(strings.TrimRight(m.load(m.items[m.cursor]), "\n"), "\n")
		_, end := window(len(content), 0, rows)
		preview = append(preview, content[:end]...)
	}

	panes := lipgloss.JoinHorizontal(lipgloss.Top,
		paneStyle.Render(fit(list, listWidth)),
		paneStyle.Render(fit(preview, previewWidth)),
	)
	return panes + "\n" + helpStyle.Render(ansi.Truncate(help, width, "…"))
}

// window は n 行のうち、cursor が見える最大 rows 行の範囲 [start, end) を返す。rows が負なら全体を返す。
func window(n, cursor, rows int) (start, end int) {
	if rows < 0 || n <= rows {
		return 0, n
	}
	start = max(cursor-rows+1, 0)
	return start, start + rows
}

// fit は各行を width 桁に切り詰め、幅をそろえて1つの文字列にする。
func fit(lines []string, width int) string {
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "…")
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(lines, "\n"))
}

// Selected は選ばれたファイルを返す。選ばれていなければ ok は false。
func (m Picker) Selected() (string, bool) {
	if !m.selected {
		return "", false
	}
	return m.items[m.cursor], true
}
