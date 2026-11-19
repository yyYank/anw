package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	viewEnvFlag  bool
	viewPortFlag bool
)

var viewCmd = &cobra.Command{
	Use:   "view --env | --port [ポート番号]",
	Short: "見る。何も変更しない",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if viewPortFlag {
			return viewPort(cmd.OutOrStdout(), args)
		}
		if len(args) > 0 {
			return errors.New("--env には引数を指定できません")
		}
		return viewEnv()
	},
}

func init() {
	viewCmd.Flags().BoolVar(&viewEnvFlag, "env", false, "env ファイルの一覧と内容のプレビュー")
	viewCmd.Flags().BoolVar(&viewPortFlag, "port", false, "LISTEN 中のポートの一覧")
	viewCmd.MarkFlagsOneRequired("env", "port")
	viewCmd.MarkFlagsMutuallyExclusive("env", "port")
	rootCmd.AddCommand(viewCmd)
}
