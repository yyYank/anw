package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// anw tidy のあと anw undo で元の状態に戻ること
func TestUndo_tidyを元に戻す(t *testing.T) {
	cli := setupCLI(t)
	path := cli.createFile("memo.md", oct27)

	if err := cli.run("tidy"); err != nil {
		t.Fatal(err)
	}
	if err := cli.run("undo"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("元の場所に戻っていない: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cli.workDir, "2026-10-27")); !os.IsNotExist(err) {
		t.Errorf("日付ディレクトリが残っている: %v", err)
	}
}

// 元に戻せる操作がない場合はエラーになること
func TestUndo_履歴がなければエラー(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run("undo"); err == nil {
		t.Error("エラーにならなかった")
	}
}
