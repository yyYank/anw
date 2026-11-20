package dotfiles

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// テスト用のホームにファイルを作る
func writeFile(t *testing.T, home, rel string, content []byte) {
	t.Helper()
	path := filepath.Join(home, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func text(rel string) []byte { return []byte("# " + rel + "\n") }

func find(t *testing.T, home string) []string {
	t.Helper()
	got, err := Find(home)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// ホーム直下のドットファイルを拾い、ドットでないファイルとドットディレクトリの中は拾わないこと
func TestFind_ホーム直下のドットファイル(t *testing.T) {
	home := t.TempDir()
	for _, rel := range []string{".zshrc", ".gitconfig", "memo.txt", ".ssh/config", ".claude/settings.json"} {
		writeFile(t, home, rel, text(rel))
	}

	got := find(t, home)

	want := []string{".gitconfig", ".zshrc"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// シンボリックリンクのドットファイルも拾うこと
func TestFind_シンボリックリンクも拾う(t *testing.T) {
	home := t.TempDir()
	writeFile(t, home, "dotfiles/zshrc", text("zshrc"))
	if err := os.Symlink(filepath.Join(home, "dotfiles", "zshrc"), filepath.Join(home, ".zshrc")); err != nil {
		t.Fatal(err)
	}

	got := find(t, home)

	if want := []string{".zshrc"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// ~/.config 配下は3階層まで拾い、それより深いものは拾わないこと
func TestFind_configは3階層まで(t *testing.T) {
	home := t.TempDir()
	for _, rel := range []string{
		".config/starship.toml",
		".config/git/config",
		".config/nvim/lua/plugins.lua",
		".config/nvim/lua/deep/too.lua",
	} {
		writeFile(t, home, rel, text(rel))
	}

	got := find(t, home)

	want := []string{
		filepath.Join(".config", "git", "config"),
		filepath.Join(".config", "nvim", "lua", "plugins.lua"),
		filepath.Join(".config", "starship.toml"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// テキストでないファイル、1MB を超えるファイル、cache・Cache・logs の中は拾わないこと
func TestFind_明らかに変なファイルは拾わない(t *testing.T) {
	home := t.TempDir()
	writeFile(t, home, ".config/app/settings.json", text("settings"))
	writeFile(t, home, ".config/app/state.db", []byte("SQLite\x00\x01\x02"))
	writeFile(t, home, ".config/app/huge.log", bytes.Repeat([]byte("a"), 1<<20+1))
	writeFile(t, home, ".config/app/cache/x.json", text("cache"))
	writeFile(t, home, ".config/Cache/y.json", text("Cache"))
	writeFile(t, home, ".config/app/logs/z.txt", text("logs"))
	writeFile(t, home, ".bin_history", []byte("\x00\x00"))

	got := find(t, home)

	want := []string{filepath.Join(".config", "app", "settings.json")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// ~/.config がなくてもエラーにならないこと
func TestFind_configがなくてもよい(t *testing.T) {
	home := t.TempDir()
	writeFile(t, home, ".zshrc", text(".zshrc"))

	if got, want := find(t, home), []string{".zshrc"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// ~/.config 自体や、その中のディレクトリがシンボリックリンクでも中を拾うこと
func TestFind_リンクされたディレクトリの中も拾う(t *testing.T) {
	home := t.TempDir()
	writeFile(t, home, "dotfiles/config/git/config", text("git"))
	writeFile(t, home, "dotfiles/nvim/init.lua", text("nvim"))
	if err := os.Symlink(filepath.Join(home, "dotfiles", "config"), filepath.Join(home, ".config")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(home, "dotfiles", "nvim"), filepath.Join(home, "dotfiles", "config", "nvim")); err != nil {
		t.Fatal(err)
	}

	got := find(t, home)

	want := []string{
		filepath.Join(".config", "git", "config"),
		filepath.Join(".config", "nvim", "init.lua"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// 読めないディレクトリがあっても、それを飛ばして残りを拾うこと
func TestFind_読めないディレクトリは飛ばす(t *testing.T) {
	home := t.TempDir()
	writeFile(t, home, ".config/git/config", text("git"))
	writeFile(t, home, ".config/secret/x.conf", text("x"))
	locked := filepath.Join(home, ".config", "secret")
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

	got := find(t, home)

	if want := []string{filepath.Join(".config", "git", "config")}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
