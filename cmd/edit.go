package cmd

import "github.com/spf13/cobra"

func init() {
	newVerb("edit", "対象を選んで編集する", "", []target{
		{name: "env", desc: "env ファイルを $EDITOR で編集",
			run: func(cmd *cobra.Command, args []string) error { return editEnv() }},
	})
}
