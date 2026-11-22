package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var oct27 = time.Date(2026, 10, 27, 12, 0, 0, 0, time.Local)

// anw tidy --by date でカレントディレクトリのファイルが振り分けられること
func TestTidy_カレントディレクトリを振り分ける(t *testing.T) {
	cli := setupCLI(t)
	cli.createFile("memo.md", oct27)

	if err := cli.run("tidy", "--by", "date"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(cli.workDir, "2026-10-27", "memo.md")); err != nil {
		t.Errorf("振り分けられていない: %v", err)
	}
}

// --by に date 以外を指定するとエラーになり、ファイルは移動されないこと
func TestTidy_不明なbyはエラー(t *testing.T) {
	cli := setupCLI(t)
	path := cli.createFile("memo.md", oct27)

	if err := cli.run("tidy", "--by", "size"); err == nil {
		t.Error("エラーにならなかった")
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("ファイルが移動された: %v", err)
	}
}

// 履歴が toml の history_dir に保存されること
func TestTidy_履歴はtomlのhistory_dirに保存される(t *testing.T) {
	cli := setupCLI(t)
	cli.createFile("memo.md", oct27)

	if err := cli.run("tidy"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(cli.historyDir, "history.json")); err != nil {
		t.Errorf("履歴が保存されていない: %v", err)
	}
}

// anw tidy --by month の結果を anw undo で元に戻せること
func TestTidy_monthをundoで戻す(t *testing.T) {
	cli := setupCLI(t)
	path := cli.createFile("memo.md", oct27)
	if err := os.Mkdir(filepath.Join(cli.workDir, "2026-10-28"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := cli.run("tidy", "--by", "month"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cli.workDir, "2026-10", "2026-10-27", "memo.md")); err != nil {
		t.Fatalf("振り分けられていない: %v", err)
	}
	if err := cli.run("undo"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("元に戻っていない: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cli.workDir, "2026-10-28")); err != nil {
		t.Errorf("日付ディレクトリが元に戻っていない: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cli.workDir, "2026-10")); !os.IsNotExist(err) {
		t.Errorf("空の月ディレクトリが残っている: %v", err)
	}
}
