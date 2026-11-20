package cmd

import "github.com/spf13/cobra"

func init() {
	newVerb("kill", "プロセスを止める", "何を止めますか？", "", []target{
		{name: "port", desc: "ポートを使っているプロセスを kill", argHint: "[ポート番号]",
			run: func(cmd *cobra.Command, args []string) error { return killPort(cmd.OutOrStdout(), args) }},
	})
}
