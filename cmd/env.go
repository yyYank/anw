package cmd

import (
	"errors"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yyYank/anw/internal/envfile"
)

// テストで差し替えられるよう変数にしておく
var (
	findEnvFiles  = envfile.Find
	showEnvPicker = runEnvPicker
	openEditor    = runEditor
)

var (
	envView bool
	envEdit bool
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "配下の env ファイルを探して、閲覧・編集する",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if envView && envEdit {
			return errors.New("--view と --edit は同時に指定できません")
		}
		editor := os.Getenv("EDITOR")
		if envEdit && editor == "" {
			return errors.New("$EDITOR が設定されていません")
		}
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		files, err := findEnvFiles(dir)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return errors.New("env ファイルが見つかりません")
		}

		if !envEdit {
			_, _, err := showEnvPicker(files, false)
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
	},
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

func init() {
	envCmd.Flags().BoolVar(&envView, "view", false, "一覧と内容のプレビュー")
	envCmd.Flags().BoolVar(&envEdit, "edit", false, "選んだファイルを $EDITOR で編集")
	rootCmd.AddCommand(envCmd)
}
