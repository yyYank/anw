package cmd

import (
	"reflect"
	"testing"
)

// edit --env で見つかったファイルが1つだけなら、選ばずにエディタで開くこと
func TestEdit_1つだけなら選ばずに開く(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env"}

	if err := cli.run("edit", "--env"); err != nil {
		t.Fatal(err)
	}

	if cli.envPicker != nil {
		t.Error("一覧が表示された")
	}
	if want := []string{".env"}; !reflect.DeepEqual(cli.edited, want) {
		t.Errorf("edited %v, want %v", cli.edited, want)
	}
}

// edit --env で複数あれば、一覧から選んだファイルを開くこと
func TestEdit_一覧から選んで開く(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env", ".env.local"}
	cli.envPick = 1

	if err := cli.run("edit", "--env"); err != nil {
		t.Fatal(err)
	}

	if cli.envPicker == nil || !*cli.envPicker {
		t.Error("選択できる一覧が表示されていない")
	}
	if want := []string{".env.local"}; !reflect.DeepEqual(cli.edited, want) {
		t.Errorf("edited %v, want %v", cli.edited, want)
	}
}

// edit --env で一覧から選ばずにやめたら、何も開かないこと
func TestEdit_選ばずにやめたら開かない(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env", ".env.local"}
	cli.envPick = -1

	if err := cli.run("edit", "--env"); err != nil {
		t.Fatal(err)
	}

	if len(cli.edited) != 0 {
		t.Errorf("開かれた: %v", cli.edited)
	}
}

// $EDITOR が未設定なら、エラーにして何も開かないこと
func TestEdit_EDITORがなければエラー(t *testing.T) {
	cli := setupCLI(t)
	t.Setenv("EDITOR", "")
	cli.envFiles = []string{".env"}

	if err := cli.run("edit", "--env"); err == nil {
		t.Error("エラーにならなかった")
	}

	if len(cli.edited) != 0 {
		t.Errorf("開かれた: %v", cli.edited)
	}
}

// edit で対象のフラグがなければ、対象が env だけなので選ばせずに env の処理に進むこと
func TestEdit_対象がなければenvに進む(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env"}

	if err := cli.run("edit"); err != nil {
		t.Fatal(err)
	}

	if cli.targetItems != nil {
		t.Error("対象を選ぶ画面が表示された")
	}
	if want := []string{".env"}; !reflect.DeepEqual(cli.edited, want) {
		t.Errorf("edited %v, want %v", cli.edited, want)
	}
}
