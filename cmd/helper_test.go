package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/yyYank/anw/internal/branch"
	"github.com/yyYank/anw/internal/choose"
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

	branches []branch.Branch
	// branchPick は ブランチの一覧の画面で選ぶ行。nil なら選ばずにやめる
	branchPick []int
	// branchShown はブランチの一覧の画面が表示されたか
	branchShown bool
	// deleted は削除したブランチ。force で消したものには "!" を付ける
	deleted []string
	// deleteFail は削除に失敗させるブランチ
	deleteFail map[string]bool

	// targetPick は対象を選ぶ画面で選ぶ対象の名前。空なら選ばずにやめる
	targetPick string
	// targetItems は対象を選ぶ画面に出た一覧。表示されなければ nil
	targetItems []choose.Item
	// targetTitle は対象を選ぶ画面の見出し
	targetTitle string

	// out はコマンドの標準出力
	out bytes.Buffer
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

	// 本物のリポジトリのブランチを調べたり消したりしないよう、偽物に差し替える
	origBranches, origSelectBranches, origDelete := listBranches, selectBranches, deleteBranch
	listBranches = func() ([]branch.Branch, error) { return c.branches, nil }
	selectBranches = func(bs []branch.Branch) ([]branch.Branch, error) {
		c.branchShown = true
		var picked []branch.Branch
		for _, i := range c.branchPick {
			picked = append(picked, bs[i])
		}
		return picked, nil
	}
	deleteBranch = func(name string, force bool) error {
		if c.deleteFail[name] {
			return fmt.Errorf("%s は削除できません", name)
		}
		if force {
			name += "!"
		}
		c.deleted = append(c.deleted, name)
		return nil
	}
	t.Cleanup(func() { listBranches, selectBranches, deleteBranch = origBranches, origSelectBranches, origDelete })

	// 対象を選ぶ画面を出さないよう、偽物に差し替える
	origChoose := chooseTarget
	chooseTarget = func(title string, items []choose.Item) (string, bool, error) {
		c.targetTitle, c.targetItems = title, items
		return c.targetPick, c.targetPick != "", nil
	}
	t.Cleanup(func() { chooseTarget = origChoose })
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
	c.out.Reset()
	rootCmd.SetOut(&c.out)
	rootCmd.SetErr(io.Discard)
	return rootCmd.Execute()
}
