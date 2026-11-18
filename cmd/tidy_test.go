package cmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func runRoot(t *testing.T, args ...string) error {
	t.Helper()
	rootCmd.SetArgs(args)
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	return rootCmd.Execute()
}

// anw tidy --by date でカレントディレクトリのファイルが振り分けられること
func TestTidy_カレントディレクトリを振り分ける(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	path := filepath.Join(dir, "memo.md")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	mtime := time.Date(2026, 10, 27, 12, 0, 0, 0, time.Local)
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}

	if err := runRoot(t, "tidy", "--by", "date"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "2026-10-27", "memo.md")); err != nil {
		t.Errorf("振り分けられていない: %v", err)
	}
}

// --by に date 以外を指定するとエラーになること
func TestTidy_不明なbyはエラー(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := runRoot(t, "tidy", "--by", "size"); err == nil {
		t.Error("エラーにならなかった")
	}
}
