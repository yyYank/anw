package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	viewEnvFlag    bool
	viewPortFlag   bool
	viewBranchFlag bool
)

var viewCmd = &cobra.Command{
	Use:   "view --env | --port [ポート番号] | --branch",
	Short: "見る。何も変更しない",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if viewPortFlag {
			return viewPort(cmd.OutOrStdout(), args)
		}
		if len(args) > 0 {
			return errors.New("--port 以外には引数を指定できません")
		}
		if viewBranchFlag {
			return viewBranches(cmd.OutOrStdout())
		}
		return viewEnv()
	},
}

func init() {
	viewCmd.Flags().BoolVar(&viewEnvFlag, "env", false, "env ファイルの一覧と内容のプレビュー")
	viewCmd.Flags().BoolVar(&viewPortFlag, "port", false, "LISTEN 中のポートの一覧")
	viewCmd.Flags().BoolVar(&viewBranchFlag, "branch", false, "削除できるブランチの一覧")
	viewCmd.MarkFlagsOneRequired("env", "port", "branch")
	viewCmd.MarkFlagsMutuallyExclusive("env", "port", "branch")
	rootCmd.AddCommand(viewCmd)
}
