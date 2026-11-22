package tidy

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// テスト用のファイルを作り、更新日時を設定する
func createFile(t *testing.T, dir, name string, mtime time.Time) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(name), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 12, 0, 0, 0, time.Local)
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("%s が存在しない: %v", path, err)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("%s が存在してはいけない: %v", path, err)
	}
}

// ファイルが更新日時の yyyy-mm-dd ディレクトリに移動されること
func TestRun_更新日時のディレクトリに移動する(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "memo.md", date(2026, 10, 27))

	if _, err := Run(dir, ByDate); err != nil {
		t.Fatal(err)
	}

	assertExists(t, filepath.Join(dir, "2026-10-27", "memo.md"))
	assertNotExists(t, filepath.Join(dir, "memo.md"))
}

// ディレクトリは移動されないこと
func TestRun_ディレクトリは移動しない(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "photos")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	mtime := date(2026, 10, 27)
	if err := os.Chtimes(sub, mtime, mtime); err != nil {
		t.Fatal(err)
	}

	if _, err := Run(dir, ByDate); err != nil {
		t.Fatal(err)
	}

	assertExists(t, sub)
	assertNotExists(t, filepath.Join(dir, "2026-10-27", "photos"))
}

// 隠しファイルは移動されないこと
func TestRun_隠しファイルは対象外(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".DS_Store", date(2026, 10, 27))

	if _, err := Run(dir, ByDate); err != nil {
		t.Fatal(err)
	}

	assertExists(t, filepath.Join(dir, ".DS_Store"))
	assertNotExists(t, filepath.Join(dir, "2026-10-27"))
}

// 移動先に同名のファイルがある場合は連番を付けること
func TestRun_同名ファイルには連番を付ける(t *testing.T) {
	dir := t.TempDir()
	destDir := filepath.Join(dir, "2026-10-29")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	createFile(t, destDir, "report.pdf", date(2026, 10, 29))
	createFile(t, destDir, "report (1).pdf", date(2026, 10, 29))
	createFile(t, dir, "report.pdf", date(2026, 10, 29))

	moves, err := Run(dir, ByDate)
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(destDir, "report (2).pdf")
	assertExists(t, want)
	if len(moves) != 1 || moves[0].To != want {
		t.Errorf("移動先が想定と違う: %+v", moves)
	}
}
