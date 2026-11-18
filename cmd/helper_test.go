package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/yyYank/anw/internal/port"
)

// testCLI はテスト用の toml を読ませ、テスト用ディレクトリでコマンドを実行する
type testCLI struct {
	t          *testing.T
	configPath string
	historyDir string
	workDir    string
	listeners  []port.Listener
	killed     []int
	// pick は一覧の画面で選ぶ行。-1 なら選ばずにやめる
	pick  int
	shown bool

	envFiles []string
	// envPick は env の一覧の画面で選ぶ行。-1 なら選ばずにやめる
	envPick int
	// envPicker は env の一覧の画面が表示されたときの selectable。表示されなければ nil
	envPicker *bool
	edited    []string
}

// setupCLI はテスト用の toml を書き出し、テスト用ディレクトリをカレントディレクトリにする
func setupCLI(t *testing.T) *testCLI {
	t.Helper()
	dir := t.TempDir()
	historyDir := filepath.Join(dir, "history")
	configPath := filepath.Join(dir, "config.toml")
	content := fmt.Sprintf("history_dir = %q\n", historyDir)
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	workDir := filepath.Join(dir, "work")
	if err := os.Mkdir(workDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(workDir)
	c := &testCLI{t: t, configPath: configPath, historyDir: historyDir, workDir: workDir}

	// 本物のプロセスを調べたり止めたりしないよう、偽物に差し替える
	origList, origKill, origSelect := listListeners, killProcess, selectListener
	listListeners = func() ([]port.Listener, error) { return c.listeners, nil }
	killProcess = func(pid int) error {
		c.killed = append(c.killed, pid)
		return nil
	}
	selectListener = func(ls []port.Listener) (port.Listener, bool, error) {
		c.shown = true
		if c.pick < 0 {
			return port.Listener{}, false, nil
		}
		return ls[c.pick], true, nil
	}
	t.Cleanup(func() { listListeners, killProcess, selectListener = origList, origKill, origSelect })

	// 本物の env ファイルを読んだりエディタを起動したりしないよう、偽物に差し替える
	t.Setenv("EDITOR", "fake-editor")
	origFind, origPicker, origEditor := findEnvFiles, showEnvPicker, openEditor
	findEnvFiles = func(string) ([]string, error) { return c.envFiles, nil }
	showEnvPicker = func(files []string, selectable bool) (string, bool, error) {
		c.envPicker = &selectable
		if !selectable || c.envPick < 0 {
			return "", false, nil
		}
		return files[c.envPick], true, nil
	}
	openEditor = func(editor, path string) error {
		c.edited = append(c.edited, path)
		return nil
	}
	t.Cleanup(func() { findEnvFiles, showEnvPicker, openEditor = origFind, origPicker, origEditor })
	return c
}

// createFile はテスト用ディレクトリにテスト対象のファイルを作り、更新日時を設定する
func (c *testCLI) createFile(name string, mtime time.Time) string {
	c.t.Helper()
	path := filepath.Join(c.workDir, name)
	if err := os.WriteFile(path, []byte(name), 0o644); err != nil {
		c.t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		c.t.Fatal(err)
	}
	return path
}

// run はフラグを初期値に戻し、テスト用の toml を指定してルートコマンドを実行する
func (c *testCLI) run(args ...string) error {
	c.t.Helper()
	reset := func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	}
	rootCmd.PersistentFlags().VisitAll(reset)
	for _, sub := range rootCmd.Commands() {
		sub.Flags().VisitAll(reset)
	}
	rootCmd.SetArgs(append([]string{"--config", c.configPath}, args...))
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	return rootCmd.Execute()
}
