package envfile

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// ウィンドウサイズがまだ分からないときに使う幅
const defaultWidth = 80

// scrollEnd はプレビューを末尾まで動かすときの位置。描画のときに末尾の位置に直す
const scrollEnd = math.MaxInt

// 絞り込みの入力モード。通常モードは空文字
const (
	modeNormal = ""
	modePath   = "path"
	modeGrep   = "grep"
	// modeFind はプレビューの中を検索する語の入力モード
	modeFind = "find"
)

var (
	paneStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	activeStyle = paneStyle.BorderForeground(lipgloss.Color("6"))
	titleStyle  = lipgloss.NewStyle().Bold(true)
	cursorStyle = lipgloss.NewStyle().Reverse(true)
	matchStyle  = lipgloss.NewStyle().Bold(true).Underline(true)
	helpStyle   = lipgloss.NewStyle().Faint(true)
)

// Picker は env ファイルの一覧と、カーソル位置のファイルの内容を表示する TUI。
// / でパスを、Ctrl+G で中身を検索して一覧を絞り込める。
type Picker struct {
	title    string
	items    []string
	contents []string
	// labels は一覧とプレビューの見出しに表示する文字列。nil なら items を表示する。
	// タブで区切ると、表示中の行の中で列をそろえる
	labels []string
	// altItems と altLabels は p で切り替える、絞り込みと表示の文字列。nil なら切り替えない
	altItems   []string
	altLabels  []string
	showAlt    bool
	selectable bool
	// visible は絞り込んだ後に表示する items の添字。cursor はこの中の位置
	visible  []int
	cursor   int
	selected bool
	mode     string
	// filter は絞り込みに使っている検索の種類。query と組で、通常モードに戻っても残る
	filter string
	query  string
	width  int
	height int
	// focusPreview はプレビューの pane に移っているか。移っていれば j/k でプレビューを動かす
	focusPreview bool
	// scroll はプレビューの先頭に表示する行
	scroll int
	// find はプレビューの中を検索する語。n / N で次と前の一致へ飛ぶ
	find string
	// notice は一番下の行に一度だけ出す知らせ
	notice string
}

// NewPicker は items を表示する Picker を返す。
// load はプレビューと中身の検索に使う内容を返す。selectable が true なら Enter で選べる。
func NewPicker(items []string, load func(path string) string, selectable bool) Picker {
	contents := make([]string, len(items))
	for i, item := range items {
		contents[i] = load(item)
	}
	return NewPickerWithContents(items, contents, selectable)
}

// NewPickerWithContents は items を表示し、contents[i] を items[i] のプレビューと中身の検索に使う Picker を返す。
func NewPickerWithContents(items, contents []string, selectable bool) Picker {
	m := Picker{title: "env ファイル", items: items, contents: contents, selectable: selectable}
	m.refilter()
	return m
}

// WithLabels は表示する文字列を labels にした Picker を返す。/ の絞り込みは元の items で行う。
func (m Picker) WithLabels(labels []string) Picker {
	m.labels = labels
	return m
}

// WithAltLabels は p で切り替えたときの、/ の絞り込みの文字列を items に、表示の文字列を labels にした Picker を返す。
func (m Picker) WithAltLabels(items, labels []string) Picker {
	m.altItems, m.altLabels = items, labels
	return m
}

// item は i 番目の / の絞り込みに使う文字列を返す。
func (m Picker) item(i int) string {
	if m.showAlt {
		return m.altItems[i]
	}
	return m.items[i]
}

// label は i 番目に表示する文字列を返す。
func (m Picker) label(i int) string {
	if m.showAlt {
		return m.altLabels[i]
	}
	if m.labels != nil {
		return m.labels[i]
	}
	return m.items[i]
}

// WithTitle は一覧の見出しを title にした Picker を返す。
func (m Picker) WithTitle(title string) Picker {
	m.title = title
	return m
}

// Init は tea.Model の実装。
func (m Picker) Init() tea.Cmd { return nil }

// Mode は今の入力モードを返す。通常モードなら空文字。
func (m Picker) Mode() string { return m.mode }

// Query は絞り込みに使っている文字列を返す。
func (m Picker) Query() string { return m.query }

// Update はキー入力を処理する。ウィンドウサイズの変更を受け取ると、描画をその大きさに合わせる。
func (m Picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = size.Width, size.Height
		return m, nil
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.mode == modeFind {
		return m.updateFind(key)
	}
	if m.mode != modeNormal {
		return m.updateInput(key)
	}
	m.notice = ""
	if m.focusPreview {
		switch key.String() {
		case "/":
			m.mode, m.find = modeFind, ""
			return m, nil
		case "n":
			m.jumpToMatch(1)
			return m, nil
		case "N":
			m.jumpToMatch(-1)
			return m, nil
		}
	}
	switch key.String() {
	case "up", "k":
		m.moveFocused(-1)
	case "down", "j":
		m.moveFocused(1)
	case "right", "l":
		m.focusPreview = len(m.visible) > 0
	case "left", "h":
		m.focusPreview = false
	case "g":
		if m.focusPreview {
			m.scroll = 0
		}
	case "G":
		if m.focusPreview {
			m.scroll = scrollEnd
		}
	case "ctrl+d":
		if m.focusPreview {
			m.scrollBy(m.halfPage())
		}
	case "ctrl+u":
		if m.focusPreview {
			m.scrollBy(-m.halfPage())
		}
	case "/":
		m.startInput(modePath)
	case "ctrl+g":
		m.startInput(modeGrep)
	case "p":
		if m.altLabels != nil {
			m.showAlt = !m.showAlt
			m.refilter()
		}
	case "enter":
		if m.selectable && len(m.visible) > 0 {
			m.selected = true
			return m, tea.Quit
		}
	case "esc", "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// updateInput は入力モードのキー入力を処理する。
// Enter は絞り込みを残して、Esc は絞り込みを解除して通常モードに戻る。
func (m Picker) updateInput(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyUp, tea.KeyCtrlP:
		m.moveCursor(-1)
	case tea.KeyDown, tea.KeyCtrlN:
		m.moveCursor(1)
	case tea.KeyEnter:
		m.mode = modeNormal
	case tea.KeyEsc:
		m.mode, m.query = modeNormal, ""
		m.refilter()
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyBackspace:
		if r := []rune(m.query); len(r) > 0 {
			m.query = string(r[:len(r)-1])
			m.refilter()
		}
	case tea.KeyRunes, tea.KeySpace:
		m.query += string(key.Runes)
		m.refilter()
	}
	return m, nil
}

// updateFind はプレビューの中を検索する語の入力を処理する。
// Enter は今の位置から下の最初の一致へ飛び、Esc は検索をやめる。
func (m Picker) updateFind(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEnter:
		m.mode = modeNormal
		m.jumpToMatch(0)
	case tea.KeyEsc:
		m.mode, m.find = modeNormal, ""
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyBackspace:
		if r := []rune(m.find); len(r) > 0 {
			m.find = string(r[:len(r)-1])
		}
	case tea.KeyRunes, tea.KeySpace:
		m.find += string(key.Runes)
	}
	return m, nil
}

// jumpToMatch はプレビューを、find を含む行へ動かす。
// dir が 0 なら今の行から下、1 なら次の行から下、-1 なら前の行から上を探し、端まで行ったら反対の端から探す。
func (m *Picker) jumpToMatch(dir int) {
	if m.find == "" || len(m.visible) == 0 {
		return
	}
	_, previewWidth, _ := m.layout(m.columnWidths())
	lines := m.previewLines(m.visible[m.cursor], previewWidth)
	cur := m.clampedScroll()
	step := 1
	if dir < 0 {
		step = -1
	}
	for k := 0; k < len(lines); k++ {
		i := ((cur+dir+step*k)%len(lines) + len(lines)) % len(lines)
		if containsFold(ansi.Strip(lines[i]), m.find) {
			m.scroll = i
			return
		}
	}
	m.notice = "見つかりません: " + m.find
}

// highlightFind は line の中の find に一致する部分を強調する。
func (m Picker) highlightFind(line string) string {
	plain := ansi.Strip(line)
	lower := strings.ToLower(plain)
	// 小文字にすると長さが変わる文字があれば、行ごと強調する
	if m.find == "" || len(lower) != len(plain) {
		if m.find != "" && containsFold(plain, m.find) {
			return matchStyle.Render(plain)
		}
		return line
	}
	word := strings.ToLower(m.find)
	var b strings.Builder
	rest := 0
	for {
		at := strings.Index(lower[rest:], word)
		if at < 0 {
			break
		}
		b.WriteString(plain[rest : rest+at])
		b.WriteString(matchStyle.Render(plain[rest+at : rest+at+len(word)]))
		rest += at + len(word)
	}
	if rest == 0 {
		return line
	}
	return b.String() + plain[rest:]
}

// startInput は mode の入力モードに入り、新しく入力を始める。
func (m *Picker) startInput(mode string) {
	m.mode, m.filter, m.query = mode, mode, ""
	m.refilter()
}

// moveCursor はカーソルを delta 行動かす。一覧の外には出ない。プレビューは先頭に戻す。
func (m *Picker) moveCursor(delta int) {
	m.cursor = min(max(m.cursor+delta, 0), max(len(m.visible)-1, 0))
	m.resetScroll()
}

// moveFocused は、プレビューの pane にいればプレビューを、いなければカーソルを delta 行動かす。
func (m *Picker) moveFocused(delta int) {
	if m.focusPreview {
		m.scrollBy(delta)
		return
	}
	m.moveCursor(delta)
}

// resetScroll はプレビューを先頭に戻す。
func (m *Picker) resetScroll() {
	m.scroll = 0
}

// scrollBy はプレビューを delta 行動かす。先頭と末尾より先には動かない。
func (m *Picker) scrollBy(delta int) {
	m.scroll = max(m.clampedScroll()+delta, 0)
	m.scroll = m.clampedScroll()
}

// clampedScroll は scroll を、先頭から末尾までの間に収めた位置を返す。
func (m Picker) clampedScroll() int {
	if len(m.visible) == 0 {
		return 0
	}
	_, previewWidth, rows := m.layout(m.columnWidths())
	if rows < 0 {
		return 0
	}
	lines := m.previewLines(m.visible[m.cursor], previewWidth)
	return min(m.scroll, max(len(lines)-rows, 0))
}

// halfPage はプレビューの半ページの行数を返す。
func (m Picker) halfPage() int {
	_, _, rows := m.layout(m.columnWidths())
	return max(rows/2, 1)
}

// previewLines は i 番目の中身を、width 桁で折り返した行にして返す。中身で絞り込んでいれば一致した行を強調する。
func (m Picker) previewLines(i, width int) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimRight(m.contents[i], "\n"), "\n") {
		if m.filter == modeGrep && m.query != "" && containsFold(line, m.query) {
			line = matchStyle.Render(line)
		}
		lines = append(lines, strings.Split(ansi.Wrap(line, width, ""), "\n")...)
	}
	return lines
}

// refilter は query で一覧を絞り込み直し、カーソルを先頭に戻す。
func (m *Picker) refilter() {
	// Picker は値で受け渡されるので、元の slice を書き換えないよう作り直す
	var visible []int
	for i := range m.items {
		if m.matches(i) {
			visible = append(visible, i)
		}
	}
	m.visible, m.cursor = visible, 0
	m.resetScroll()
	m.focusPreview = m.focusPreview && len(visible) > 0
}

// matches は items[i] が今の絞り込みに一致するかを返す。
func (m Picker) matches(i int) bool {
	if m.query == "" {
		return true
	}
	if m.filter == modeGrep {
		return containsFold(m.contents[i], m.query)
	}
	_, ok := fuzzyMatch(m.item(i), m.query)
	return ok
}

// View は左に一覧、右にカーソル位置のファイルのプレビューを並べて描画する。
func (m Picker) View() string {
	width := m.screenWidth()
	widths := m.columnWidths()
	listWidth, previewWidth, rows := m.layout(widths)

	list := []string{titleStyle.Render(m.title)}
	start, end := window(len(m.visible), m.cursor, rows)
	for i := start; i < end; i++ {
		if i == m.cursor {
			list = append(list, cursorStyle.Render("> ")+m.highlightLabel(m.visible[i], widths, cursorStyle))
		} else {
			list = append(list, "  "+m.highlightLabel(m.visible[i], widths, lipgloss.NewStyle()))
		}
	}
	if len(m.visible) == 0 {
		list = append(list, helpStyle.Render("  (一致なし)"))
	}

	preview := []string{}
	if len(m.visible) > 0 {
		i := m.visible[m.cursor]
		preview = append(preview, titleStyle.Render(align(m.label(i), widths)))
		lines := m.previewLines(i, previewWidth)
		start := m.clampedScroll()
		end := len(lines)
		if rows >= 0 {
			end = min(start+rows, end)
		}
		for _, line := range lines[start:end] {
			preview = append(preview, m.highlightFind(line))
		}
	}

	listStyle, previewStyle := activeStyle, paneStyle
	if m.focusPreview {
		listStyle, previewStyle = paneStyle, activeStyle
	}
	panes := lipgloss.JoinHorizontal(lipgloss.Top,
		listStyle.Render(fit(list, listWidth)),
		previewStyle.Render(fit(preview, previewWidth)),
	)
	return panes + "\n" + ansi.Truncate(m.statusLine(), width, "…")
}

// screenWidth は画面の桁数を返す。まだ分からなければ defaultWidth を返す。
func (m Picker) screenWidth() int {
	if m.width <= 0 {
		return defaultWidth
	}
	return m.width
}

// layout は一覧とプレビューの桁数と、一覧に表示できる行数を返す。行数が分からなければ -1 を返す。
func (m Picker) layout(widths []int) (listWidth, previewWidth, rows int) {
	// 枠とパディングで左右それぞれ4桁使う
	const frame = 4
	width := m.screenWidth()
	// 一致がなくても「(一致なし)」が切れない幅は取る
	listWidth = lipgloss.Width("  (一致なし)")
	for _, i := range m.visible {
		listWidth = max(listWidth, lipgloss.Width(align(m.label(i), widths))+2)
	}
	listWidth = min(listWidth, (width-2*frame)/2)
	previewWidth = width - 2*frame - listWidth

	// 高さが分かっていれば、枠2行・タイトル1行・ヘルプ1行を除いた行数に収める
	rows = -1
	if m.height > 0 {
		rows = max(m.height-4, 1)
	}
	return listWidth, previewWidth, rows
}

// highlightLabel は i 番目を base で描画する。表示の文字列が絞り込みの文字列を含むときは、一致した文字を強調する。
func (m Picker) highlightLabel(i int, widths []int, base lipgloss.Style) string {
	label, item := align(m.label(i), widths), m.item(i)
	at := strings.Index(label, item)
	if at < 0 {
		return base.Render(label)
	}
	return base.Render(label[:at]) + m.highlightPath(item, base) + base.Render(label[at+len(item):])
}

// highlightPath は path を base で描画し、パスの絞り込みで一致した文字をさらに強調する。
func (m Picker) highlightPath(path string, base lipgloss.Style) string {
	var pos []int
	if m.filter == modePath && m.query != "" {
		pos, _ = fuzzyMatch(path, m.query)
	}
	if len(pos) == 0 {
		return base.Render(path)
	}
	hit := map[int]bool{}
	for _, p := range pos {
		hit[p] = true
	}
	// 強調の後で base が切れないよう、1文字ずつ描画する
	match := matchStyle.Inherit(base)
	var b strings.Builder
	for i, r := range []rune(path) {
		if hit[i] {
			b.WriteString(match.Render(string(r)))
		} else {
			b.WriteString(base.Render(string(r)))
		}
	}
	return b.String()
}

// columnWidths は表示中の行の、タブで区切った列ごとの最大の桁数を返す。
func (m Picker) columnWidths() []int {
	var widths []int
	for _, i := range m.visible {
		for c, cell := range strings.Split(m.label(i), "\t") {
			if c == len(widths) {
				widths = append(widths, 0)
			}
			widths[c] = max(widths[c], lipgloss.Width(cell))
		}
	}
	return widths
}

// align は label のタブで区切った列を widths の桁数まで空白で埋め、2桁の空白でつなぐ。最後の列は埋めない。
func align(label string, widths []int) string {
	cells := strings.Split(label, "\t")
	for c := range cells[:len(cells)-1] {
		cells[c] += strings.Repeat(" ", max(widths[c]-lipgloss.Width(cells[c]), 0))
	}
	return strings.Join(cells, "  ")
}

// statusLine は画面の一番下に出す、入力中の文字かキーの説明を返す。
func (m Picker) statusLine() string {
	switch m.mode {
	case modePath:
		return "/" + m.query + "▏" + helpStyle.Render("  Enter: 確定  Esc: 解除")
	case modeGrep:
		return "grep: " + m.query + "▏" + helpStyle.Render("  Enter: 確定  Esc: 解除")
	case modeFind:
		return "検索: " + m.find + "▏" + helpStyle.Render("  Enter: 移動  Esc: やめる")
	}
	if m.notice != "" {
		return m.notice
	}
	help := "↑↓/jk: 移動  →/l: プレビューへ  /: パス検索  Ctrl+G: 中身検索  q: 終了"
	if m.selectable {
		help = "↑↓/jk: 移動  →/l: プレビューへ  /: パス検索  Ctrl+G: 中身検索  Enter: 編集  q: やめる"
	}
	if m.focusPreview {
		help = "↑↓/jk: スクロール  g/G: 先頭/末尾  Ctrl+D/U: 半ページ  /: 検索  n/N: 次/前  ←/h: 一覧へ  q: やめる"
		if m.find != "" {
			help = "検索: " + m.find + "  " + help
		}
	}
	if m.altLabels != nil {
		help = strings.Replace(help, "  q:", "  p: 表示切替  q:", 1)
	}
	if m.query != "" {
		prefix := "/"
		if m.filter == modeGrep {
			prefix = "grep: "
		}
		return prefix + m.query + "  " + helpStyle.Render(help)
	}
	return helpStyle.Render(help)
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
	return m.items[m.visible[m.cursor]], true
}

// SelectedIndex は選ばれた要素の添字を返す。選ばれていなければ ok は false。
func (m Picker) SelectedIndex() (int, bool) {
	if !m.selected {
		return 0, false
	}
	return m.visible[m.cursor], true
}
