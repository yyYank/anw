package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

// testCLI はテスト用の toml を読ませ、テスト用ディレクトリでコマンドを実行する
type testCLI struct {
	t          *testing.T
	configPath string
	historyDir string
	workDir    string
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
	return &testCLI{t: t, configPath: configPath, historyDir: historyDir, workDir: workDir}
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
