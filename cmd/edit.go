package cmd

import "github.com/spf13/cobra"

var editEnvFlag bool

var editCmd = &cobra.Command{
	Use:   "edit --env",
	Short: "対象を選んで編集する",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return editEnv()
	},
}

func init() {
	editCmd.Flags().BoolVar(&editEnvFlag, "env", false, "env ファイルを $EDITOR で編集")
	editCmd.MarkFlagsOneRequired("env")
	rootCmd.AddCommand(editCmd)
}
