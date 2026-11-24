package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// [[snippet.名前]] のステップを順番どおりに読み、dir の ~ を展開すること
func TestLoad_snippetを読む(t *testing.T) {
	t.Setenv("HOME", "/home/me")
	path := writeToml(t, `
[[snippet.morning]]
dir = "~/Desktop"
args = ["tidy"]

[[snippet.morning]]
dir = "/tmp/dl"
args = ["tidy", "--by", "month"]
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	want := []Step{
		{Dir: filepath.Join("/home/me", "Desktop"), Args: []string{"tidy"}},
		{Dir: "/tmp/dl", Args: []string{"tidy", "--by", "month"}},
	}
	if got := cfg.Snippets["morning"]; !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// snippet の書き方の誤りを、実行前に見つけられること
func TestValidateSnippet_書き方の誤りはエラー(t *testing.T) {
	cfg := Config{Snippets: map[string][]Step{
		"ok":      {{Dir: "/tmp", Args: []string{"tidy"}}},
		"empty":   {},
		"nodir":   {{Args: []string{"tidy"}}},
		"noargs":  {{Dir: "/tmp"}},
		"nest":    {{Dir: "/tmp", Args: []string{"run", "--snippet", "ok"}}},
		"partial": {{Dir: "/tmp", Args: []string{"tidy"}}, {Dir: "/tmp"}},
	}}

	if _, err := cfg.ValidateSnippet("ok"); err != nil {
		t.Errorf("ok: %v", err)
	}
	for _, name := range []string{"none", "empty", "nodir", "noargs", "nest", "partial"} {
		if _, err := cfg.ValidateSnippet(name); err == nil {
			t.Errorf("%s がエラーにならなかった", name)
		}
	}
}

// cd の行で dir を決め、anw の行を1ステップとして読むこと
func TestParseSnippet_cdとanwの行をステップにする(t *testing.T) {
	t.Setenv("HOME", "/home/me")

	steps, err := ParseSnippet("# 朝の片付け\ncd ~/Desktop\nanw tidy\n\ncd /tmp/dl\nanw tidy --by month\nanw view --env\n")
	if err != nil {
		t.Fatal(err)
	}

	want := []Step{
		{Dir: filepath.Join("/home/me", "Desktop"), Args: []string{"tidy"}},
		{Dir: "/tmp/dl", Args: []string{"tidy", "--by", "month"}},
		{Dir: "/tmp/dl", Args: []string{"view", "--env"}},
	}
	if !reflect.DeepEqual(steps, want) {
		t.Errorf("got %+v, want %+v", steps, want)
	}
}

// cd と anw 以外の行や、cd より前の anw は、行番号つきのエラーにすること
func TestParseSnippet_書き方の誤りは行番号つきのエラー(t *testing.T) {
	cases := map[string]string{
		"cd より前の anw":  "anw tidy",
		"シェルのコマンド":     "cd /tmp\nrm memo.md",
		"cd に dir がない": "cd\nanw tidy",
	}
	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ParseSnippet(text)
			if err == nil || !strings.Contains(err.Error(), "行目") {
				t.Errorf("err = %v", err)
			}
		})
	}
}

// 設定ファイルの隣の snippets/名前 を、名前の snippet として読むこと
func TestLoad_snippetsのファイルを読む(t *testing.T) {
	path := writeToml(t, "")
	writeSnippetFile(t, path, "morning", "cd /tmp\nanw tidy\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	steps, err := cfg.ValidateSnippet("morning")
	if err != nil {
		t.Fatal(err)
	}
	if want := []Step{{Dir: "/tmp", Args: []string{"tidy"}}}; !reflect.DeepEqual(steps, want) {
		t.Errorf("got %+v, want %+v", steps, want)
	}
	if got := cfg.SnippetNames(); !reflect.DeepEqual(got, []string{"morning"}) {
		t.Errorf("names = %v", got)
	}
}

// 読めない snippet のファイルや、toml と名前が重なる snippet は、その snippet を使うときだけエラーにすること
func TestLoad_読めないsnippetは使うときにエラー(t *testing.T) {
	path := writeToml(t, "[[snippet.dup]]\ndir = \"/tmp\"\nargs = [\"tidy\"]\n")
	writeSnippetFile(t, path, "dup", "cd /tmp\nanw tidy\n")
	writeSnippetFile(t, path, "broken", "rm -r /\n")
	writeSnippetFile(t, path, "ok", "cd /tmp\nanw tidy\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("ほかの設定まで読めなくなった: %v", err)
	}

	for _, name := range []string{"dup", "broken"} {
		if _, err := cfg.ValidateSnippet(name); err == nil {
			t.Errorf("%s がエラーにならなかった", name)
		}
	}
	if _, err := cfg.ValidateSnippet("ok"); err != nil {
		t.Errorf("ok: %v", err)
	}
	if got := cfg.SnippetNames(); !reflect.DeepEqual(got, []string{"broken", "dup", "ok"}) {
		t.Errorf("names = %v", got)
	}
}

// writeSnippetFile は設定ファイル configPath の隣の snippets/name に text を書く
func writeSnippetFile(t *testing.T, configPath, name, text string) {
	t.Helper()
	dir := SnippetDir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}
