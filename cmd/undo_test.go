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

// undo の説明が、ファイル操作だけでなくブランチの削除にも触れていること
func TestUndo_説明は対応する操作に触れる(t *testing.T) {
	if undoCmd.Short != "直前の操作（tidy・delete --branch）を元に戻す" {
		t.Errorf("Short = %q", undoCmd.Short)
	}
}
