package cmd

import "github.com/spf13/cobra"

var (
	deleteBranchFlag bool
	deleteForce      bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete --branch [--force]",
	Short: "対象を選んで削除する",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteBranches(cmd.OutOrStdout(), cmd.ErrOrStderr(), deleteForce)
	},
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteBranchFlag, "branch", false, "マージ済み・リモート削除済みのローカルブランチ")
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "マージされていなくても強制削除する")
	deleteCmd.MarkFlagsOneRequired("branch")
	rootCmd.AddCommand(deleteCmd)
}
