package cmd

import "github.com/spf13/cobra"

func init() {
	newVerb("view", "見る。何も変更しない", "何を見ますか？", "", []target{
		{name: "env", desc: "env ファイルの一覧と内容のプレビュー",
			run: func(cmd *cobra.Command, args []string) error { return viewEnv() }},
		{name: "port", desc: "LISTEN 中のポートの一覧", argHint: "[ポート番号]",
			run: func(cmd *cobra.Command, args []string) error { return viewPort(cmd.OutOrStdout(), args) }},
		{name: "branch", desc: "削除できるブランチの一覧",
			run: func(cmd *cobra.Command, args []string) error { return viewBranches(cmd.OutOrStdout()) }},
		{name: "ai", desc: "コーディングエージェントのセッション", argHint: "[.]",
			run: func(cmd *cobra.Command, args []string) error { return viewAI(cmd.OutOrStdout(), args) }},
		{name: "dotfiles", desc: "ホームの設定ファイルの一覧と内容のプレビュー",
			run: func(cmd *cobra.Command, args []string) error { return viewDotfiles() }},
	})
}
