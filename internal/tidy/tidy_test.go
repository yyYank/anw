package tidy

import (
	"os"
	"path/filepath"
	"reflect"
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

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

// --by month でファイルが yyyy-mm/yyyy-mm-dd ディレクトリに移動されること
func TestRun_monthはファイルを月と日の入れ子に移動する(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "memo.md", date(2026, 10, 27))

	if _, err := Run(dir, ByMonth); err != nil {
		t.Fatal(err)
	}

	assertExists(t, filepath.Join(dir, "2026-10", "2026-10-27", "memo.md"))
	assertNotExists(t, filepath.Join(dir, "memo.md"))
}

// --by month で yyyy-mm-dd ディレクトリが丸ごと yyyy-mm ディレクトリの下に移動されること
func TestRun_monthは日付ディレクトリを月の下に移動する(t *testing.T) {
	dir := t.TempDir()
	mkdir(t, filepath.Join(dir, "2026-10-27"))
	createFile(t, filepath.Join(dir, "2026-10-27"), "memo.md", date(2026, 10, 27))

	moves, err := Run(dir, ByMonth)
	if err != nil {
		t.Fatal(err)
	}

	assertExists(t, filepath.Join(dir, "2026-10", "2026-10-27", "memo.md"))
	assertNotExists(t, filepath.Join(dir, "2026-10-27"))
	want := []Move{{From: filepath.Join(dir, "2026-10-27"), To: filepath.Join(dir, "2026-10", "2026-10-27")}}
	if !reflect.DeepEqual(moves, want) {
		t.Errorf("got %+v, want %+v", moves, want)
	}
}

// 移動先の yyyy-mm/yyyy-mm-dd が既にあれば、中身を1件ずつ統合し、空になった元のディレクトリを消すこと
func TestRun_monthは既存の日付ディレクトリに統合する(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "2026-10", "2026-10-27")
	mkdir(t, dest)
	createFile(t, dest, "a.md", date(2026, 10, 27))
	src := filepath.Join(dir, "2026-10-27")
	mkdir(t, src)
	createFile(t, src, "a.md", date(2026, 10, 27))
	createFile(t, src, "b.md", date(2026, 10, 27))

	if _, err := Run(dir, ByMonth); err != nil {
		t.Fatal(err)
	}

	assertExists(t, filepath.Join(dest, "a.md"))
	assertExists(t, filepath.Join(dest, "a (1).md"))
	assertExists(t, filepath.Join(dest, "b.md"))
	assertNotExists(t, src)
}

// --by month でも、日付の名前でないディレクトリは移動されないこと
func TestRun_monthは日付以外のディレクトリを移動しない(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"photos", "2026-10", "2026-13-01"} {
		mkdir(t, filepath.Join(dir, name))
	}

	if _, err := Run(dir, ByMonth); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"photos", "2026-10", "2026-13-01"} {
		assertExists(t, filepath.Join(dir, name))
	}
	assertNotExists(t, filepath.Join(dir, "2026-10", "2026-10"))
}

// --by には date と month を指定でき、それ以外はエラーになること
func TestParseBy_dateとmonthだけ受け付ける(t *testing.T) {
	for s, want := range map[string]By{"date": ByDate, "month": ByMonth} {
		if got, err := ParseBy(s); err != nil || got != want {
			t.Errorf("%q: got %q, err %v", s, got, err)
		}
	}
	if _, err := ParseBy("size"); err == nil {
		t.Error("size がエラーにならなかった")
	}
}
