package cmd

import "github.com/spf13/cobra"

var killPortFlag bool

var killCmd = &cobra.Command{
	Use:   "kill --port [ポート番号]",
	Short: "プロセスを止める",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return killPort(cmd.OutOrStdout(), args)
	},
}

func init() {
	killCmd.Flags().BoolVar(&killPortFlag, "port", false, "ポートを使っているプロセスを kill")
	killCmd.MarkFlagsOneRequired("port")
	rootCmd.AddCommand(killCmd)
}
