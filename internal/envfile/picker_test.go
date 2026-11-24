package envfile

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var files = []string{".env", "app/.env.local"}

// 実ファイルを読まず、ファイル名に応じた偽の内容を返す
func fakeLoad(path string) string {
	return "CONTENT_OF " + path
}

func press(m tea.Model, keys ...tea.KeyMsg) Picker {
	for _, k := range keys {
		m, _ = m.Update(k)
	}
	return m.(Picker)
}

var (
	keyDown  = tea.KeyMsg{Type: tea.KeyDown}
	keyEnter = tea.KeyMsg{Type: tea.KeyEnter}
	keyEsc   = tea.KeyMsg{Type: tea.KeyEsc}
)

// 一覧と、カーソル位置のファイルの内容が表示されること
func TestPicker_一覧とプレビューを表示する(t *testing.T) {
	view := NewPicker(files, fakeLoad, false).View()

	for _, s := range []string{".env", "app/.env.local", "CONTENT_OF .env"} {
		if !strings.Contains(view, s) {
			t.Errorf("%q が表示されていない:\n%s", s, view)
		}
	}
}

// カーソルを動かすとプレビューが切り替わること
func TestPicker_移動するとプレビューが切り替わる(t *testing.T) {
	view := press(NewPicker(files, fakeLoad, false), keyDown).View()

	if !strings.Contains(view, "CONTENT_OF app/.env.local") {
		t.Errorf("プレビューが切り替わっていない:\n%s", view)
	}
}

// 選択できる画面では Enter でカーソル位置のファイルが選ばれること
func TestPicker_Enterで選ぶ(t *testing.T) {
	m := press(NewPicker(files, fakeLoad, true), keyDown, keyEnter)

	got, ok := m.Selected()
	if !ok || got != "app/.env.local" {
		t.Errorf("got %q, ok=%v", got, ok)
	}
}

// 閲覧だけの画面では Enter を押しても選ばれないこと
func TestPicker_閲覧ではEnterで選ばない(t *testing.T) {
	m := press(NewPicker(files, fakeLoad, false), keyEnter)

	if _, ok := m.Selected(); ok {
		t.Error("選ばれてしまった")
	}
}

// Esc で何も選ばずに終わること
func TestPicker_Escでキャンセル(t *testing.T) {
	m := press(NewPicker(files, fakeLoad, true), keyEsc)

	if _, ok := m.Selected(); ok {
		t.Error("選ばれてしまった")
	}
}

// ウィンドウ幅が分かると、左に一覧・右にプレビューが横並びで表示されること
func TestPicker_一覧とプレビューを左右に並べる(t *testing.T) {
	m, _ := NewPicker(files, fakeLoad, true).Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	view := m.View()

	for _, line := range strings.Split(view, "\n") {
		// 一覧の ".env" とプレビューの "CONTENT_OF .env" が同じ行にあれば ".env" は2回現れる
		if strings.Contains(line, "CONTENT_OF .env") && strings.Count(line, ".env") >= 2 {
			return
		}
	}
	t.Errorf("一覧とプレビューが同じ行に並んでいない:\n%s", view)
}

// 描画がウィンドウ幅に収まること
func TestPicker_ウィンドウ幅に収まる(t *testing.T) {
	m, _ := NewPicker(files, fakeLoad, true).Update(tea.WindowSizeMsg{Width: 40, Height: 20})

	for _, line := range strings.Split(m.View(), "\n") {
		if w := lipgloss.Width(line); w > 40 {
			t.Errorf("幅 %d がウィンドウ幅 40 を超えている: %q", w, line)
		}
	}
}

var searchFiles = []string{".env", "apps/front/.env.local", "apps/api/.env"}

var searchContents = map[string]string{
	".env":                  "DEBUG=true\n",
	"apps/front/.env.local": "NEXT_PUBLIC_URL=http://localhost:3000\n",
	"apps/api/.env":         "API_PORT=8080\nDB_HOST=localhost\n",
}

func loadSearch(path string) string { return searchContents[path] }

// typed は文字の入力を表すキー
func typed(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

var keyCtrlG = tea.KeyMsg{Type: tea.KeyCtrlG}

// / でパスを部分列で絞り込み、Enter で通常モードに戻ってから Enter で選べること
func TestPicker_スラッシュでパスを絞り込む(t *testing.T) {
	for _, q := range []string{"front", "apps/front", "afl"} {
		t.Run(q, func(t *testing.T) {
			m := press(NewPicker(searchFiles, loadSearch, true), typed("/"), typed(q), keyEnter)
			if m.Mode() != "" {
				t.Fatalf("1回目の Enter で通常モードに戻っていない: %q", m.Mode())
			}
			if _, ok := m.Selected(); ok {
				t.Fatal("1回目の Enter で選ばれてしまった")
			}

			m = press(m, keyEnter)

			if got, ok := m.Selected(); !ok || got != "apps/front/.env.local" {
				t.Errorf("got %q, ok=%v", got, ok)
			}
		})
	}
}

// Ctrl+G でファイルの中身から絞り込めること。大文字小文字は区別しない
func TestPicker_CtrlGで中身から絞り込む(t *testing.T) {
	for _, q := range []string{"8080", "API", "port"} {
		t.Run(q, func(t *testing.T) {
			m := press(NewPicker(searchFiles, loadSearch, true), keyCtrlG, typed(q), keyEnter, keyEnter)

			if got, ok := m.Selected(); !ok || got != "apps/api/.env" {
				t.Errorf("got %q, ok=%v", got, ok)
			}
		})
	}
}

// 入力モードで Esc を押すと、絞り込みを解除して通常モードに戻ること
func TestPicker_Escで絞り込みを解除する(t *testing.T) {
	m := press(NewPicker(searchFiles, loadSearch, true), typed("/"), typed("front"), keyEsc)

	if m.Mode() != "" {
		t.Errorf("通常モードに戻っていない: %q", m.Mode())
	}
	view := m.View()
	for _, f := range searchFiles {
		if !strings.Contains(view, f) {
			t.Errorf("%q が表示されていない:\n%s", f, view)
		}
	}
	// 通常モードの Esc は今までどおり終了する
	if _, ok := press(m, keyEsc).Selected(); ok {
		t.Error("選ばれてしまった")
	}
}

// 入力モードでは j / k / q は移動や終了ではなく文字として入ること
func TestPicker_入力モードではjkqを文字として扱う(t *testing.T) {
	m := press(NewPicker(searchFiles, loadSearch, true), keyCtrlG, typed("j"), typed("k"), typed("q"))

	if m.Mode() == "" || m.Query() != "jkq" {
		t.Errorf("mode=%q query=%q", m.Mode(), m.Query())
	}
}

// 一致するファイルがなければ、Enter を押しても何も選ばれないこと
func TestPicker_一致がなければ選ばない(t *testing.T) {
	m := press(NewPicker(searchFiles, loadSearch, true), typed("/"), typed("zzz"), keyEnter, keyEnter)

	if _, ok := m.Selected(); ok {
		t.Error("選ばれてしまった")
	}
}

// 入力中は今のモードと入力した文字が表示されること
func TestPicker_入力中はモードを表示する(t *testing.T) {
	view := press(NewPicker(searchFiles, loadSearch, true), keyCtrlG, typed("8080")).View()

	if !strings.Contains(view, "8080") || !strings.Contains(view, "grep") {
		t.Errorf("モードと入力が表示されていない:\n%s", view)
	}
}

// WithTitle で一覧の見出しを変えられること。指定しなければ "env ファイル" のまま
func TestPicker_見出しを変えられる(t *testing.T) {
	if view := NewPicker(files, fakeLoad, false).View(); !strings.Contains(view, "env ファイル") {
		t.Errorf("既定の見出しがない:\n%s", view)
	}

	view := NewPicker(files, fakeLoad, false).WithTitle("dotfiles").View()

	if !strings.Contains(view, "dotfiles") || strings.Contains(view, "env ファイル") {
		t.Errorf("見出しが変わっていない:\n%s", view)
	}
}

// 中身を直接渡して作れ、WithLabels で表示を変えても、/ の絞り込みは元の文字列で行うこと
func TestPicker_表示と絞り込みの文字列を分ける(t *testing.T) {
	keys := []string{"/work/app", "/work/web"}
	labels := []string{"11-01 claude app 「ログイン」", "11-03 codex web 「API」"}
	contents := []string{"会話1", "会話2"}
	m := NewPickerWithContents(keys, contents, true).WithLabels(labels)

	view := m.View()
	if !strings.Contains(view, labels[0]) || !strings.Contains(view, "会話1") {
		t.Errorf("表示の文字列か中身が出ていない:\n%s", view)
	}

	// 表示の文字列にしかない「codex」では一致せず、元の文字列の「web」で一致する
	if _, ok := press(m, typed("/"), typed("codex"), keyEnter, keyEnter).Selected(); ok {
		t.Error("表示の文字列で一致してしまった")
	}
	got := press(m, typed("/"), typed("web"), keyEnter, keyEnter)
	if i, ok := got.SelectedIndex(); !ok || i != 1 {
		t.Errorf("SelectedIndex = %d, ok=%v", i, ok)
	}
}

// WithAltLabels を渡すと、p で表示の文字列を切り替えられること。絞り込みは元の文字列のまま
func TestPicker_pで表示を切り替える(t *testing.T) {
	keys := []string{"/home/me/ghq/github.com/yyYank/anw", "/home/me/ghq/github.com/yyYank/web"}
	short := []string{"yyYank/anw  ログイン", "yyYank/web  API"}
	full := []string{"/home/me/ghq/github.com/yyYank/anw  ログイン", "/home/me/ghq/github.com/yyYank/web  API"}
	m := NewPickerWithContents(keys, []string{"会話1", "会話2"}, true).WithLabels(short).WithAltLabels(full)

	if view := m.View(); !strings.Contains(view, short[0]) || strings.Contains(view, "/home/me") {
		t.Errorf("最初は短い表示になっていない:\n%s", view)
	}
	m = press(m, typed("p"))
	if view := m.View(); !strings.Contains(view, "/home/me/ghq") {
		t.Errorf("p で切り替わっていない:\n%s", view)
	}
	m = press(m, typed("p"))
	if view := m.View(); strings.Contains(view, "/home/me") {
		t.Errorf("もう一度 p で戻っていない:\n%s", view)
	}

	// 入力モードの p は文字として扱う
	got := press(m, typed("/"), typed("web"), keyEnter, keyEnter)
	if i, ok := got.SelectedIndex(); !ok || i != 1 {
		t.Errorf("SelectedIndex = %d, ok=%v", i, ok)
	}
	if strings.Contains(press(m, typed("/"), typed("p")).View(), "/home/me") {
		t.Error("入力モードの p で切り替わった")
	}
}

// 切り替える表示がなければ、p では何も変わらず、キーの説明にも出さないこと
func TestPicker_切り替える表示がなければpは何もしない(t *testing.T) {
	m := NewPicker(files, fakeLoad, false)

	if before, after := m.View(), press(m, typed("p")).View(); before != after {
		t.Errorf("表示が変わった:\n%s\n---\n%s", before, after)
	}
	if strings.Contains(m.View(), "p:") {
		t.Errorf("p の説明が出ている:\n%s", m.View())
	}
}
