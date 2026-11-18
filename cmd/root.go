package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/yyYank/anw/internal/config"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "anw",
	Short: "anyway、なんとかする。日常の細かい面倒をとりあえず片づけるコマンド集",
}

// Execute はルートコマンドを実行する。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// loadConfig は --config の toml を読み込む。未指定なら既定のパスを使う。
func loadConfig() (config.Config, error) {
	path := configPath
	if path == "" {
		p, err := config.DefaultPath()
		if err != nil {
			return config.Config{}, err
		}
		path = p
	}
	return config.Load(path)
}

// historyPath は操作の履歴を保存するファイルのパスを返す。
func historyPath() (string, error) {
	cfg, err := loadConfig()
	if err != nil {
		return "", err
	}
	return cfg.HistoryFile(), nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "設定ファイルのパス (既定: ~/.config/anw/config.toml)")
}
