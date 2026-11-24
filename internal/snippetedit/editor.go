// Package snippetedit は snippet を書く、vim のように操作できるテキストエディタの TUI。
package snippetedit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle = lipgloss.NewStyle().Reverse(true)
	helpStyle   = lipgloss.NewStyle().Faint(true)
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type mode int

const (
	insertMode mode = iota
	normalMode
	commandMode
)

// Model は snippet のテキストを編集する TUI。入力モードで始まり、Esc でノーマルモードに切り替える。
type Model struct {
	title string
	lines [][]rune
	row   int
	col   int
	mode  mode
	// command はノーマルモードで : のあとに入力中のコマンド
	command string
	// message は保存の結果やエラーなど、最下行に出す一言
	message string
	isErr   bool
	// savedText は最後に保存した内容。開いたときは開いた内容
	savedText string
	saved     bool
	save      func(string) error
}

// New は name の snippet の text を編集する Model を返す。保存するたびに save を呼ぶ。
// カーソルは末尾に置き、そのまま書き足せるようにする。
func New(name, text string, save func(string) error) Model {
	var lines [][]rune
	for _, l := range strings.Split(text, "\n") {
		lines = append(lines, []rune(l))
	}
	m := Model{title: name, lines: lines, savedText: text, save: save}
	m.row = len(lines) - 1
	m.col = len(lines[m.row])
	return m
}

// Saved は一度でも保存したかを返す。
func (m Model) Saved() bool { return m.saved }

// Init は tea.Model の実装。
func (m Model) Init() tea.Cmd { return nil }

// Update はキー入力を、モードごとの操作として扱う。Ctrl+S はどのモードでも保存、Ctrl+C は保存せずに終了する。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.Type {
	case tea.KeyCtrlS:
		m.write()
		return m, nil
	case tea.KeyCtrlC:
		return m, tea.Quit
	}
	switch m.mode {
	case insertMode:
		m.updateInsert(key)
	case normalMode:
		m.updateNormal(key)
	case commandMode:
		return m.updateCommand(key)
	}
	return m, nil
}

func (m *Model) updateInsert(key tea.KeyMsg) {
	line := m.lines[m.row]
	switch key.Type {
	case tea.KeyEsc:
		m.mode = normalMode
		m.left()
	case tea.KeyEnter:
		rest := append([]rune{}, line[m.col:]...)
		m.lines[m.row] = line[:m.col]
		m.lines = append(m.lines[:m.row+1], append([][]rune{rest}, m.lines[m.row+1:]...)...)
		m.row, m.col = m.row+1, 0
	case tea.KeyBackspace:
		switch {
		case m.col > 0:
			m.lines[m.row] = append(line[:m.col-1:m.col-1], line[m.col:]...)
			m.col--
		case m.row > 0:
			prev := m.lines[m.row-1]
			m.col = len(prev)
			m.lines[m.row-1] = append(prev[:len(prev):len(prev)], line...)
			m.lines = append(m.lines[:m.row], m.lines[m.row+1:]...)
			m.row--
		}
	case tea.KeyRunes, tea.KeySpace:
		r := key.Runes
		if key.Type == tea.KeySpace {
			r = []rune{' '}
		}
		m.lines[m.row] = append(line[:m.col:m.col], append(append([]rune{}, r...), line[m.col:]...)...)
		m.col += len(r)
	default:
		m.move(key.String())
	}
}

func (m *Model) updateNormal(key tea.KeyMsg) {
	m.message = ""
	switch k := key.String(); k {
	case "i":
		m.mode = insertMode
	case "a":
		m.mode = insertMode
		if len(m.lines[m.row]) > 0 {
			m.col++
		}
	case "o":
		m.lines = append(m.lines[:m.row+1], append([][]rune{{}}, m.lines[m.row+1:]...)...)
		m.row, m.col = m.row+1, 0
		m.mode = insertMode
	case "x":
		line := m.lines[m.row]
		if len(line) > 0 {
			m.lines[m.row] = append(line[:m.col:m.col], line[m.col+1:]...)
			m.clamp()
		}
	case "0":
		m.col = 0
	case "$":
		m.col = len(m.lines[m.row])
		m.clamp()
	case ":":
		m.mode, m.command = commandMode, ""
	case "h":
		m.move("left")
	case "j":
		m.move("down")
	case "k":
		m.move("up")
	case "l":
		m.move("right")
	default:
		m.move(k)
	}
}

func (m Model) updateCommand(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEsc:
		m.mode = normalMode
	case tea.KeyBackspace:
		if m.command == "" {
			m.mode = normalMode
			break
		}
		r := []rune(m.command)
		m.command = string(r[:len(r)-1])
	case tea.KeyEnter:
		m.mode = normalMode
		return m.exec(m.command)
	case tea.KeyRunes:
		m.command += string(key.Runes)
	}
	return m, nil
}

// exec は : のあとに入力したコマンドを実行する。
func (m Model) exec(command string) (tea.Model, tea.Cmd) {
	switch command {
	case "w":
		m.write()
	case "q":
		if m.text() != m.savedText {
			m.setMessage("保存していない変更があります (:w で保存、:q! で破棄して終了)", true)
			return m, nil
		}
		return m, tea.Quit
	case "q!":
		return m, tea.Quit
	case "wq", "x":
		if m.write() {
			return m, tea.Quit
		}
	default:
		m.setMessage("不明なコマンドです: "+command, true)
	}
	return m, nil
}

// write は今の内容を保存する。保存できなければエラーを表示して false を返す。
func (m *Model) write() bool {
	text := m.text()
	if err := m.save(text); err != nil {
		m.setMessage(err.Error(), true)
		return false
	}
	m.savedText, m.saved = text, true
	m.setMessage("保存しました", false)
	return true
}

func (m *Model) setMessage(msg string, isErr bool) {
	m.message, m.isErr = msg, isErr
}

func (m Model) text() string {
	lines := make([]string, len(m.lines))
	for i, l := range m.lines {
		lines[i] = string(l)
	}
	return strings.Join(lines, "\n")
}

// move は矢印キーの名前 dir の向きにカーソルを動かす。
func (m *Model) move(dir string) {
	switch dir {
	case "left":
		m.left()
	case "right":
		m.col++
	case "up":
		if m.row > 0 {
			m.row--
		}
	case "down":
		if m.row < len(m.lines)-1 {
			m.row++
		}
	}
	m.clamp()
}

func (m *Model) left() {
	if m.col > 0 {
		m.col--
	}
}

// clamp はカーソルを行の中に収める。入力モードでは行末の次、ほかのモードでは行末の文字までにする。
func (m *Model) clamp() {
	end := len(m.lines[m.row])
	if m.mode != insertMode && end > 0 {
		end--
	}
	m.col = min(m.col, end)
}

// View はテキストと、モード・コマンド・メッセージの行を描画する。
func (m Model) View() string {
	var b strings.Builder
	b.WriteString("snippet: " + m.title + "\n")
	b.WriteString(helpStyle.Render("cd <dir> と anw <コマンド> を1行ずつ書く。Ctrl+S / :w 保存、:wq 保存して終了、:q 終了、:q! 破棄して終了") + "\n\n")
	for i, l := range m.lines {
		if i != m.row {
			b.WriteString(string(l) + "\n")
			continue
		}
		under := " "
		if m.col < len(l) {
			under = string(l[m.col])
		}
		after := ""
		if m.col+1 < len(l) {
			after = string(l[m.col+1:])
		}
		b.WriteString(string(l[:m.col]) + cursorStyle.Render(under) + after + "\n")
	}
	b.WriteString("\n")
	switch {
	case m.mode == commandMode:
		b.WriteString(":" + m.command)
	case m.message != "" && m.isErr:
		b.WriteString(errStyle.Render(m.message))
	case m.message != "":
		b.WriteString(m.message)
	case m.mode == insertMode:
		b.WriteString(helpStyle.Render("-- 入力 -- (Esc でノーマルモード)"))
	default:
		b.WriteString(helpStyle.Render("-- ノーマル -- (i で入力、: でコマンド)"))
	}
	b.WriteString("\n")
	return b.String()
}
