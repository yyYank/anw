package cmd

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yyYank/anw/internal/config"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "anw",
	Short: "anyway、なんとかする。日常の細かい面倒をとりあえず片づけるコマンド集",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printVerbTable(cmd.OutOrStdout())
	},
}

// printVerbTable は動詞と、その動詞で使える対象の対応表を出力する。
func printVerbTable(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "動詞\t対象")
	for _, v := range verbs {
		fmt.Fprintf(tw, "%s\t%s\n", v.cmd.Name(), v.targetFlags(" "))
	}
	// 対象を取らない動詞
	for _, name := range []string{"tidy", "undo"} {
		fmt.Fprintf(tw, "%s\t\n", name)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "\n詳しくは anw <動詞> --help")
	return err
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
