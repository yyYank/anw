package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeToml(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// toml の history_dir が環境変数より優先されること
func TestLoad_tomlが環境変数より優先(t *testing.T) {
	t.Setenv("ANW_HISTORY_DIR", "/from/env")
	path := writeToml(t, `history_dir = "/from/toml"`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HistoryDir != "/from/toml" {
		t.Errorf("got %q", cfg.HistoryDir)
	}
}

// toml がなければ環境変数を使うこと
func TestLoad_tomlがなければ環境変数(t *testing.T) {
	t.Setenv("ANW_HISTORY_DIR", "/from/env")

	cfg, err := Load(filepath.Join(t.TempDir(), "none.toml"))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HistoryDir != "/from/env" {
		t.Errorf("got %q", cfg.HistoryDir)
	}
}

// toml に history_dir がなければ環境変数を使うこと
func TestLoad_tomlにキーがなければ環境変数(t *testing.T) {
	t.Setenv("ANW_HISTORY_DIR", "/from/env")
	path := writeToml(t, "")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HistoryDir != "/from/env" {
		t.Errorf("got %q", cfg.HistoryDir)
	}
}

// toml も環境変数もなければ ~/.anw を使うこと
func TestLoad_どちらもなければホームのanw(t *testing.T) {
	t.Setenv("ANW_HISTORY_DIR", "")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(filepath.Join(t.TempDir(), "none.toml"))
	if err != nil {
		t.Fatal(err)
	}

	if want := filepath.Join(home, ".anw"); cfg.HistoryDir != want {
		t.Errorf("got %q, want %q", cfg.HistoryDir, want)
	}
}

// toml の書式が壊れていればエラーになること
func TestLoad_壊れたtomlはエラー(t *testing.T) {
	path := writeToml(t, `history_dir = `)

	if _, err := Load(path); err == nil {
		t.Error("エラーにならなかった")
	}
}
