package config

import (
	"path/filepath"
	"reflect"
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
