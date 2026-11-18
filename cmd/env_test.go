package cmd

import (
	"reflect"
	"testing"
)

// --edit で見つかったファイルが1つだけなら、選ばずにエディタで開くこと
func TestEnv_1つだけなら選ばずに開く(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env"}

	if err := cli.run("env", "--edit"); err != nil {
		t.Fatal(err)
	}

	if cli.envPicker != nil {
		t.Error("一覧が表示された")
	}
	if want := []string{".env"}; !reflect.DeepEqual(cli.edited, want) {
		t.Errorf("edited %v, want %v", cli.edited, want)
	}
}

// --edit で複数あれば、一覧から選んだファイルを開くこと
func TestEnv_一覧から選んで開く(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env", ".env.local"}
	cli.envPick = 1

	if err := cli.run("env", "--edit"); err != nil {
		t.Fatal(err)
	}

	if cli.envPicker == nil || !*cli.envPicker {
		t.Error("選択できる一覧が表示されていない")
	}
	if want := []string{".env.local"}; !reflect.DeepEqual(cli.edited, want) {
		t.Errorf("edited %v, want %v", cli.edited, want)
	}
}

// --edit で一覧から選ばずにやめたら、何も開かないこと
func TestEnv_選ばずにやめたら開かない(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env", ".env.local"}
	cli.envPick = -1

	if err := cli.run("env", "--edit"); err != nil {
		t.Fatal(err)
	}

	if len(cli.edited) != 0 {
		t.Errorf("開かれた: %v", cli.edited)
	}
}

// $EDITOR が未設定なら、エラーにして何も開かないこと
func TestEnv_EDITORがなければエラー(t *testing.T) {
	cli := setupCLI(t)
	t.Setenv("EDITOR", "")
	cli.envFiles = []string{".env"}

	if err := cli.run("env", "--edit"); err == nil {
		t.Error("エラーにならなかった")
	}

	if len(cli.edited) != 0 {
		t.Errorf("開かれた: %v", cli.edited)
	}
}

// --view とフラグなしでは、閲覧だけの一覧を表示し、何も開かないこと
func TestEnv_viewは閲覧だけ(t *testing.T) {
	for _, args := range [][]string{{"env", "--view"}, {"env"}} {
		t.Run(args[len(args)-1], func(t *testing.T) {
			cli := setupCLI(t)
			cli.envFiles = []string{".env"}

			if err := cli.run(args...); err != nil {
				t.Fatal(err)
			}

			if cli.envPicker == nil || *cli.envPicker {
				t.Error("閲覧だけの一覧が表示されていない")
			}
			if len(cli.edited) != 0 {
				t.Errorf("開かれた: %v", cli.edited)
			}
		})
	}
}

// env ファイルがなければ、一覧を出さずにエラーにすること
func TestEnv_見つからなければエラー(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run("env", "--view"); err == nil {
		t.Error("エラーにならなかった")
	}

	if cli.envPicker != nil {
		t.Error("一覧が表示された")
	}
}

// --view と --edit を同時に指定するとエラーになること
func TestEnv_viewとeditの同時指定はエラー(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env"}

	if err := cli.run("env", "--view", "--edit"); err == nil {
		t.Error("エラーにならなかった")
	}

	if cli.envPicker != nil || len(cli.edited) != 0 {
		t.Error("一覧の表示かファイルの編集が行われた")
	}
}
