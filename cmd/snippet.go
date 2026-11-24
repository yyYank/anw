package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yyYank/anw/internal/choose"
	"github.com/yyYank/anw/internal/config"
	"github.com/yyYank/anw/internal/snippetedit"
)

// テストで差し替えられるよう変数にしておく
var showSnippetEditor = runSnippetEditor

// editSnippet は snippet を書く画面を開き、保存するたびに内容を確かめてから snippets/名前 に書く。
// 名前を省略したら snippets にある snippet から選ばせる。
func editSnippet(args []string) error {
	// 保存のたびにフラグを初期値に戻すので、先に値を取っておく
	cfgPath, err := configFile()
	if err != nil {
		return err
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	dir := config.SnippetDir(cfgPath)

	var name string
	if len(args) == 1 {
		name = args[0]
	} else {
		var ok bool
		if name, ok, err = chooseSnippetFile(dir); err != nil {
			return err
		}
		if !ok {
			return errCanceled
		}
	}
	if name == "" || strings.HasPrefix(name, ".") || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("snippet の名前に使えません: %q", name)
	}

	path := filepath.Join(dir, name)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if _, ok := cfg.Snippets[name]; ok {
			return fmt.Errorf("snippet %q は config.toml に書かれています。config.toml を編集してください", name)
		}
	} else if err != nil {
		return err
	}

	save := func(text string) error {
		steps, err := config.ParseSnippet(text)
		if err != nil {
			return err
		}
		if _, err := (config.Config{Snippets: map[string][]config.Step{name: steps}}).ValidateSnippet(name); err != nil {
			return err
		}
		if err := validateSteps(steps); err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		return os.WriteFile(path, []byte(text), 0o644)
	}
	saved, err := showSnippetEditor(name, string(data), save)
	if err != nil {
		return err
	}
	if !saved {
		return errCanceled
	}
	return nil
}

// chooseSnippetFile は dir にある snippet のファイルから1つ選ばせる。
func chooseSnippetFile(dir string) (string, bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", false, err
	}
	var items []choose.Item
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			items = append(items, choose.Item{Name: e.Name(), Desc: filepath.Join(dir, e.Name())})
		}
	}
	if len(items) == 0 {
		return "", false, errors.New("編集できる snippet がありません。新しく書くときは名前を指定してください (anw edit --snippet 名前)")
	}
	return chooseTarget("どの snippet を編集しますか？", items)
}

// runSnippetEditor は name の snippet を text から編集する画面を表示し、一度でも保存したかを返す。
func runSnippetEditor(name, text string, save func(string) error) (bool, error) {
	m, err := tea.NewProgram(snippetedit.New(name, text, save)).Run()
	if err != nil {
		return false, err
	}
	return m.(snippetedit.Model).Saved(), nil
}
