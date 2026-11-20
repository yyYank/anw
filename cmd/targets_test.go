package cmd

import (
	"strings"
	"testing"

	"github.com/yyYank/anw/internal/port"
)

// 動詞ごとの対象の一覧から、その動詞のフラグとヘルプが作られていること
func TestVerbs_フラグとヘルプは対象の一覧から作られる(t *testing.T) {
	for _, v := range verbs {
		t.Run(v.cmd.Name(), func(t *testing.T) {
			for _, tg := range v.targets {
				f := v.cmd.Flags().Lookup(tg.name)
				if f == nil {
					t.Errorf("--%s のフラグがない", tg.name)
					continue
				}
				if f.Usage != tg.desc {
					t.Errorf("--%s の説明が %q、want %q", tg.name, f.Usage, tg.desc)
				}
				if !strings.Contains(v.cmd.Use, "--"+tg.name) {
					t.Errorf("ヘルプに --%s がない: %q", tg.name, v.cmd.Use)
				}
			}
		})
	}
}

// 対象を取る動詞がすべて一覧に入っていること
func TestVerbs_対象を取る動詞がそろっている(t *testing.T) {
	got := map[string]bool{}
	for _, v := range verbs {
		got[v.cmd.Name()] = true
	}
	for _, name := range []string{"view", "edit", "kill", "delete"} {
		if !got[name] {
			t.Errorf("%s が一覧にない", name)
		}
	}
}

// 対象を省略すると、その動詞で使える対象だけを見出し付きで一覧すること
func TestVerbs_対象を省略すると一覧から選ばせる(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run("view"); err != nil {
		t.Fatal(err)
	}

	if cli.targetTitle != "view: 何を見ますか？" {
		t.Errorf("見出しが %q", cli.targetTitle)
	}
	var names []string
	for _, it := range cli.targetItems {
		if it.Desc == "" {
			t.Errorf("%s に説明がない", it.Name)
		}
		names = append(names, it.Name)
	}
	if got, want := strings.Join(names, ","), "env,port,branch,dotfiles"; got != want {
		t.Errorf("一覧が %s、want %s", got, want)
	}
}

// 一覧で選んだ対象は、フラグを付けたときと同じ処理に進むこと
func TestVerbs_選んだ対象の処理に進む(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env"}
	cli.targetPick = "env"

	if err := cli.run("view"); err != nil {
		t.Fatal(err)
	}

	if cli.envPicker == nil || *cli.envPicker {
		t.Error("view --env と同じ閲覧だけの一覧が表示されていない")
	}
}

// 一覧で選ばずにやめたら、何もせずに終わること
func TestVerbs_選ばずにやめたら何もしない(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env"}

	if err := cli.run("view"); err != nil {
		t.Fatal(err)
	}

	if cli.envPicker != nil {
		t.Error("env の一覧が表示された")
	}
}

// フラグを付けたら、対象を選ぶ画面を挟まないこと
func TestVerbs_フラグがあれば選ばせない(t *testing.T) {
	cli := setupCLI(t)
	cli.envFiles = []string{".env"}

	if err := cli.run("view", "--env"); err != nil {
		t.Fatal(err)
	}

	if cli.targetItems != nil {
		t.Error("対象を選ぶ画面が表示された")
	}
}

// 対象が1つしかない動詞は、選ぶ画面を挟まずにその対象へ進むこと
func TestVerbs_対象が1つなら選ばせない(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{{Port: 3000, PID: 873, Name: "node"}}
	cli.pick = -1

	if err := cli.run("kill"); err != nil {
		t.Fatal(err)
	}

	if cli.targetItems != nil {
		t.Error("対象を選ぶ画面が表示された")
	}
	if !cli.shown {
		t.Error("ポートを選ぶ画面に進んでいない")
	}
	if len(cli.killed) != 0 {
		t.Errorf("kill された: %v", cli.killed)
	}
}

// 対象を省略して値だけ渡したら、エラーにして何もしないこと
func TestVerbs_対象なしの値はエラー(t *testing.T) {
	cli := setupCLI(t)
	cli.listeners = []port.Listener{{Port: 3000, PID: 873, Name: "node"}}
	cli.targetPick = "port"

	if err := cli.run("view", "3000"); err == nil {
		t.Error("エラーにならなかった")
	}

	if cli.targetItems != nil {
		t.Error("対象を選ぶ画面が表示された")
	}
}

// その動詞で使えない対象を指定したら、使える対象を返すこと
func TestVerbs_合わない対象は使える対象を返す(t *testing.T) {
	cli := setupCLI(t)

	err := cli.run("kill", "--env")

	if err == nil || err.Error() != "kill で使えるのは --port です" {
		t.Errorf("err = %v", err)
	}
	if len(cli.killed) != 0 {
		t.Errorf("kill された: %v", cli.killed)
	}
}

// 対象が複数ある動詞では、使える対象を並べて返すこと
func TestVerbs_合わない対象は使える対象を並べて返す(t *testing.T) {
	cli := setupCLI(t)

	err := cli.run("edit", "--port")

	if err == nil || err.Error() != "edit で使えるのは --env, --dotfiles です" {
		t.Errorf("err = %v", err)
	}
	err = cli.run("view", "--foo")
	if err == nil || err.Error() != "view で使えるのは --env, --port, --branch, --dotfiles です" {
		t.Errorf("err = %v", err)
	}
}

// anw だけを打つと、動詞と対象の対応表を出すこと
func TestRoot_対応表を出す(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run(); err != nil {
		t.Fatal(err)
	}

	out := cli.out.String()
	for _, want := range [][]string{
		{"view", "--env", "--port", "--branch", "--dotfiles"},
		{"edit", "--env", "--dotfiles"},
		{"kill", "--port"},
		{"delete", "--branch"},
		{"tidy"},
		{"undo"},
	} {
		found := false
		for _, line := range strings.Split(out, "\n") {
			if fields := strings.Fields(line); len(fields) > 0 && fields[0] == want[0] {
				found = strings.Join(fields, " ") == strings.Join(want, " ")
				break
			}
		}
		if !found {
			t.Errorf("%v の行がない:\n%s", want, out)
		}
	}
}

// 対応表は仕様の表と同じ順に並び、見出しは他の表と同じ英字であること
func TestRoot_対応表の並びと見出し(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run(); err != nil {
		t.Fatal(err)
	}

	var firsts []string
	for _, line := range strings.Split(cli.out.String(), "\n") {
		if fields := strings.Fields(line); len(fields) > 0 && !strings.HasPrefix(line, "詳しくは") {
			firsts = append(firsts, fields[0])
		}
	}
	if got, want := strings.Join(firsts, ","), "VERB,view,edit,kill,delete,tidy,undo"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
