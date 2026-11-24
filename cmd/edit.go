package cmd

import "github.com/spf13/cobra"

func init() {
	newVerb("edit", "対象を選んで編集する", "何を編集しますか？", "", []target{
		{name: "env", desc: "env ファイルを $EDITOR で編集",
			run: func(cmd *cobra.Command, args []string) error { return editEnv() }},
		{name: "dotfiles", desc: "ホームの設定ファイルを $EDITOR で編集",
			run: func(cmd *cobra.Command, args []string) error { return editDotfiles() }},
		{name: "snippet", desc: "snippet を書いて snippets/名前 に保存", argHint: "[名前]",
			run: func(cmd *cobra.Command, args []string) error { return editSnippet(args) }},
	})
}
