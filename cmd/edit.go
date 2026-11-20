package cmd

import "github.com/spf13/cobra"

func init() {
	newVerb("edit", "対象を選んで編集する", "何を編集しますか？", "", []target{
		{name: "env", desc: "env ファイルを $EDITOR で編集",
			run: func(cmd *cobra.Command, args []string) error { return editEnv() }},
	})
}
