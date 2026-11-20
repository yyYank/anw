package cmd

import (
	"errors"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yyYank/anw/internal/dotfiles"
	"github.com/yyYank/anw/internal/envfile"
)

// テストで差し替えられるよう変数にしておく
var (
	findDotfiles   = dotfiles.Find
	homeDir        = os.UserHomeDir
	showFilePicker = runFilePicker
)

// findDotfilesInHome はホームの設定ファイルを探す。見つからなければエラーにする。
func findDotfilesInHome() (home string, files []string, err error) {
	if home, err = homeDir(); err != nil {
		return "", nil, err
	}
	if files, err = findDotfiles(home); err != nil {
		return "", nil, err
	}
	if len(files) == 0 {
		return "", nil, errors.New("設定ファイルが見つかりません")
	}
	return home, files, nil
}

// viewDotfiles はホームの設定ファイルの一覧と内容のプレビューを表示する。
func viewDotfiles() error {
	home, files, err := findDotfilesInHome()
	if err != nil {
		return err
	}
	_, _, err = showFilePicker("dotfiles", home, files, false)
	return err
}

// editDotfiles は選んだ設定ファイルを $EDITOR で開く。1つだけなら選ばずに開く。
func editDotfiles() error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return errors.New("$EDITOR が設定されていません")
	}
	home, files, err := findDotfilesInHome()
	if err != nil {
		return err
	}
	file := files[0]
	if len(files) > 1 {
		var ok bool
		if file, ok, err = showFilePicker("dotfiles", home, files, true); err != nil || !ok {
			return err
		}
	}
	return openEditor(editor, filepath.Join(home, file))
}

// runFilePicker は base からの相対パス files の一覧と内容のプレビューを、title を見出しにして表示する。
func runFilePicker(title, base string, files []string, selectable bool) (string, bool, error) {
	load := func(path string) string {
		data, err := os.ReadFile(filepath.Join(base, path))
		if err != nil {
			return err.Error()
		}
		return string(data)
	}
	m, err := tea.NewProgram(envfile.NewPicker(files, load, selectable).WithTitle(title)).Run()
	if err != nil {
		return "", false, err
	}
	file, ok := m.(envfile.Picker).Selected()
	return file, ok, nil
}
