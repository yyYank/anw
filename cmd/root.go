package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

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
