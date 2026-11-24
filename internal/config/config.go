package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config は anw の設定。
type Config struct {
	HistoryDir string `toml:"history_dir"`
	// Snippets は名前ごとの snippet のステップ。toml では [[snippet.名前]] で書く
	Snippets map[string][]Step `toml:"snippet"`
	// snippetErrs は読めなかった snippet ごとのエラー。使うときにエラーにする
	snippetErrs map[string]error
}

// DefaultPath は設定ファイルの既定のパスを返す。
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "anw", "config.toml"), nil
}

// Load は path の toml を読み込む。
// 値がない項目は、環境変数、既定値の順に補う。
func Load(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}
	if cfg.HistoryDir == "" {
		cfg.HistoryDir = os.Getenv("ANW_HISTORY_DIR")
	}
	if cfg.HistoryDir != "" {
		dir, err := expandPath(cfg.HistoryDir)
		if err != nil {
			return Config{}, err
		}
		cfg.HistoryDir = dir
	}
	if cfg.HistoryDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, err
		}
		cfg.HistoryDir = filepath.Join(home, ".anw")
	}
	for _, steps := range cfg.Snippets {
		for i := range steps {
			if steps[i].Dir == "" {
				continue
			}
			dir, err := expandPath(steps[i].Dir)
			if err != nil {
				return Config{}, err
			}
			steps[i].Dir = dir
		}
	}
	if err := cfg.loadSnippetFiles(SnippetDir(path)); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// expandPath は先頭の ~ をホームディレクトリに置き換え、環境変数を展開する。
func expandPath(p string) (string, error) {
	p = os.ExpandEnv(p)
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, strings.TrimPrefix(p, "~")), nil
}

// HistoryFile は操作の履歴を保存するファイルのパスを返す。
func (c Config) HistoryFile() string {
	return filepath.Join(c.HistoryDir, "history.json")
}
