package cmd

import (
	"path/filepath"
	"reflect"
	"testing"
)

// view --dotfiles で、ホームを基準にした閲覧だけの一覧を表示し、何も開かないこと
func TestView_dotfilesは閲覧だけ(t *testing.T) {
	cli := setupCLI(t)
	cli.dotfiles = []string{".zshrc", ".gitconfig"}

	if err := cli.run("view", "--dotfiles"); err != nil {
		t.Fatal(err)
	}

	p := cli.filePicker
	if p == nil || p.selectable || p.title != "dotfiles" || p.base != cli.home {
		t.Errorf("閲覧だけの一覧が表示されていない: %+v", p)
	}
	if len(cli.edited) != 0 {
		t.Errorf("開かれた: %v", cli.edited)
	}
}

// edit --dotfiles で見つかったファイルが1つだけなら、選ばずにホームからのパスで開くこと
func TestEdit_dotfilesが1つなら選ばずに開く(t *testing.T) {
	cli := setupCLI(t)
	cli.dotfiles = []string{".zshrc"}

	if err := cli.run("edit", "--dotfiles"); err != nil {
		t.Fatal(err)
	}

	if cli.filePicker != nil {
		t.Error("一覧が表示された")
	}
	if want := []string{filepath.Join(cli.home, ".zshrc")}; !reflect.DeepEqual(cli.edited, want) {
		t.Errorf("edited %v, want %v", cli.edited, want)
	}
}

// edit --dotfiles で複数あれば、一覧から選んだファイルを開くこと
func TestEdit_dotfilesを一覧から選んで開く(t *testing.T) {
	cli := setupCLI(t)
	cli.dotfiles = []string{".gitconfig", ".config/git/config"}
	cli.filePick = 1

	if err := cli.run("edit", "--dotfiles"); err != nil {
		t.Fatal(err)
	}

	if cli.filePicker == nil || !cli.filePicker.selectable {
		t.Error("選択できる一覧が表示されていない")
	}
	if want := []string{filepath.Join(cli.home, ".config/git/config")}; !reflect.DeepEqual(cli.edited, want) {
		t.Errorf("edited %v, want %v", cli.edited, want)
	}
}

// edit --dotfiles で選ばずにやめたら、何も開かないこと
func TestEdit_dotfilesを選ばずにやめたら開かない(t *testing.T) {
	cli := setupCLI(t)
	cli.dotfiles = []string{".gitconfig", ".zshrc"}
	cli.filePick = -1

	if err := cli.run("edit", "--dotfiles"); err != nil {
		t.Fatal(err)
	}

	if len(cli.edited) != 0 {
		t.Errorf("開かれた: %v", cli.edited)
	}
}

// $EDITOR が未設定なら、エラーにして何も開かないこと
func TestEdit_dotfilesはEDITORがなければエラー(t *testing.T) {
	cli := setupCLI(t)
	t.Setenv("EDITOR", "")
	cli.dotfiles = []string{".zshrc"}

	if err := cli.run("edit", "--dotfiles"); err == nil {
		t.Error("エラーにならなかった")
	}
	if len(cli.edited) != 0 {
		t.Errorf("開かれた: %v", cli.edited)
	}
}

// 見つからなければ、一覧を出さずにエラーにすること
func TestView_dotfilesが見つからなければエラー(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run("view", "--dotfiles"); err == nil {
		t.Error("エラーにならなかった")
	}
	if cli.filePicker != nil {
		t.Error("一覧が表示された")
	}
}
