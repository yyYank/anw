package cmd

import "github.com/spf13/cobra"

var deleteForce bool

func init() {
	v := newVerb("delete", "対象を選んで削除する", "何を削除しますか？", "[--force]", []target{
		{name: "branch", desc: "マージ済み・リモート削除済みのローカルブランチ",
			run: func(cmd *cobra.Command, args []string) error {
				return deleteBranches(cmd.OutOrStdout(), cmd.ErrOrStderr(), deleteForce)
			}},
	})
	v.cmd.Flags().BoolVar(&deleteForce, "force", false, "マージされていなくても強制削除する")
}
