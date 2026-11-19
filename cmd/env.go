package cmd

import (
	"errors"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yyYank/anw/internal/envfile"
)

// テストで差し替えられるよう変数にしておく
var (
	findEnvFiles  = envfile.Find
	showEnvPicker = runEnvPicker
	openEditor    = runEditor
)

// findEnvFilesInWd はカレントディレクトリ配下の env ファイルを探す。見つからなければエラーにする。
func findEnvFilesInWd() ([]string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	files, err := findEnvFiles(dir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("env ファイルが見つかりません")
	}
	return files, nil
}

// viewEnv は env ファイルの一覧と内容のプレビューを表示する。
func viewEnv() error {
	files, err := findEnvFilesInWd()
	if err != nil {
		return err
	}
	_, _, err = showEnvPicker(files, false)
	return err
}

// editEnv は選んだ env ファイルを $EDITOR で開く。1つだけなら選ばずに開く。
func editEnv() error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return errors.New("$EDITOR が設定されていません")
	}
	files, err := findEnvFilesInWd()
	if err != nil {
		return err
	}
	if len(files) == 1 {
		return openEditor(editor, files[0])
	}
	file, ok, err := showEnvPicker(files, true)
	if err != nil || !ok {
		return err
	}
	return openEditor(editor, file)
}

// runEnvPicker は env ファイルの一覧の画面を表示し、選ばれたファイルを返す。
func runEnvPicker(files []string, selectable bool) (string, bool, error) {
	load := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			return err.Error()
		}
		return string(data)
	}
	m, err := tea.NewProgram(envfile.NewPicker(files, load, selectable)).Run()
	if err != nil {
		return "", false, err
	}
	file, ok := m.(envfile.Picker).Selected()
	return file, ok, nil
}

// runEditor は editor で path を開く。editor には "code -w" のような引数付きの指定も使える。
func runEditor(editor, path string) error {
	c := exec.Command("sh", "-c", editor+` "$1"`, "sh", path)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}
