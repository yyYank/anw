package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// snippetFile はテスト用の toml の隣の snippets/name のパスを返す
func (c *testCLI) snippetFile(name string) string {
	return filepath.Join(filepath.Dir(c.configPath), "snippets", name)
}

// 新しい名前なら空で開き、保存した内容を snippets/名前 に書いて、run で実行できること
func TestEditSnippet_新しいsnippetを書いて保存する(t *testing.T) {
	cli, a, _ := setupSnippet(t)
	cli.snippetInputs = []string{"cd " + a + "\nanw tidy"}
	createFileIn(t, a, "memo.md")

	if err := cli.run("edit", "--snippet", "morning"); err != nil {
		t.Fatal(err)
	}

	if cli.snippetEditor == nil || cli.snippetEditor.name != "morning" || cli.snippetEditor.text != "" {
		t.Fatalf("空の snippet として開いていない: %+v", cli.snippetEditor)
	}
	data, err := os.ReadFile(cli.snippetFile("morning"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "cd "+a+"\nanw tidy\n" {
		t.Errorf("保存された内容が %q", data)
	}
	if err := cli.run("run", "--snippet", "morning"); err != nil {
		t.Fatal(err)
	}
	assertMoved(t, a, "memo.md")
}

// もう一度開くと、保存されている内容から編集できること
func TestEditSnippet_保存された内容から開く(t *testing.T) {
	cli, a, b := setupSnippet(t)
	cli.snippetInputs = []string{"cd " + a + "\nanw tidy\n"}
	if err := cli.run("edit", "--snippet", "morning"); err != nil {
		t.Fatal(err)
	}

	cli.snippetInputs = []string{"cd " + a + "\nanw tidy\ncd " + b + "\nanw tidy\n"}
	if err := cli.run("edit", "--snippet", "morning"); err != nil {
		t.Fatal(err)
	}

	if want := "cd " + a + "\nanw tidy\n"; cli.snippetEditor.text != want {
		t.Errorf("開いた内容が %q、want %q", cli.snippetEditor.text, want)
	}
	if data, _ := os.ReadFile(cli.snippetFile("morning")); !strings.Contains(string(data), b) {
		t.Errorf("更新されていない: %q", data)
	}
}

// 実行できない内容は保存せず、エラーを画面に返すこと
func TestEditSnippet_正しくない内容は保存しない(t *testing.T) {
	cases := map[string]string{
		"dir がない":    "cd {a}/none\nanw tidy",
		"コマンドがない":    "cd {a}\nanw nope",
		"シェルのコマンド":   "cd {a}\nrm memo.md",
		"run を呼んでいる": "cd {a}\nanw run --snippet s",
		"ステップがない":    "# まだ何もない",
	}
	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			cli, a, _ := setupSnippet(t)
			cli.snippetInputs = []string{strings.ReplaceAll(text, "{a}", a)}

			// 一度も保存せずに閉じたので errCanceled を返す
			if err := cli.run("edit", "--snippet", "morning"); !errors.Is(err, errCanceled) {
				t.Errorf("err = %v, want errCanceled", err)
			}

			if len(cli.snippetSaveErrs) != 1 || cli.snippetSaveErrs[0] == nil {
				t.Errorf("保存がエラーにならなかった: %v", cli.snippetSaveErrs)
			}
			if _, err := os.Stat(cli.snippetFile("morning")); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("ファイルが書かれた: %v", err)
			}
		})
	}
}

// 名前を省略したら、snippets にある snippet から選ばせること
func TestEditSnippet_名前がなければ一覧から選ばせる(t *testing.T) {
	cli, a, _ := setupSnippet(t)
	for _, name := range []string{"night", "morning"} {
		cli.snippetInputs = []string{"cd " + a + "\nanw tidy"}
		if err := cli.run("edit", "--snippet", name); err != nil {
			t.Fatal(err)
		}
	}
	cli.targetPick = "night"
	cli.snippetInputs = nil

	// 保存せずに閉じたので errCanceled を返す
	if err := cli.run("edit", "--snippet"); !errors.Is(err, errCanceled) {
		t.Errorf("err = %v, want errCanceled", err)
	}

	var names []string
	for _, it := range cli.targetItems {
		names = append(names, it.Name)
	}
	if got := strings.Join(names, ","); got != "morning,night" {
		t.Errorf("一覧が %s", got)
	}
	if cli.snippetEditor.name != "night" {
		t.Errorf("開いた snippet が %s", cli.snippetEditor.name)
	}
}

// 名前を省略して snippets が1つもなければ、名前を指定するようエラーにすること
func TestEditSnippet_snippetsがなければ名前を求める(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run("edit", "--snippet"); err == nil || !strings.Contains(err.Error(), "名前") {
		t.Errorf("err = %v", err)
	}
	if cli.snippetEditor != nil {
		t.Error("エディタが開いた")
	}
}

// config.toml に書いた snippet や、ファイル名にできない名前は、開かずにエラーにすること
func TestEditSnippet_編集できない名前はエラー(t *testing.T) {
	cli, _, _ := setupSnippet(t, `dir = "{a}"`+"\n"+`args = ["tidy"]`)

	for _, name := range []string{"s", "../x", ".hidden"} {
		if err := cli.run("edit", "--snippet", name); err == nil {
			t.Errorf("%s がエラーにならなかった", name)
		}
	}
	if cli.snippetEditor != nil {
		t.Error("エディタが開いた")
	}
}
