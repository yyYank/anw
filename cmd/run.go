package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/yyYank/anw/internal/config"
	"github.com/yyYank/anw/internal/history"
)

// runSnippetSteps は名前で指定した snippet のステップを順に実行する。
func runSnippetSteps(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("snippet の名前を指定してください (一覧は anw view --snippet)")
	}
	// ステップを実行するたびにフラグを初期値に戻すので、先に値を取っておく
	name, cfgPath := args[0], configPath
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	steps, err := cfg.ValidateSnippet(name)
	if err != nil {
		return err
	}
	if err := validateSteps(steps); err != nil {
		return fmt.Errorf("snippet %q: %w", name, err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	// 各ステップの操作をまとめて1回の操作として保存し、anw undo でまとめて戻せるようにする
	var records []history.Record
	snippetRecords = &records
	runErr := runSteps(cmd, steps, cfgPath)
	snippetRecords = nil
	if len(records) > 0 {
		if err := saveHistory(history.Record{Kind: history.KindSnippet, Steps: records}); err != nil {
			return errors.Join(runErr, err)
		}
	}
	return runErr
}

// runSteps は steps を順に実行する。失敗するか一覧の画面でやめたら、そこで止める。
func runSteps(cmd *cobra.Command, steps []config.Step, cfgPath string) error {
	for i, s := range steps {
		fmt.Fprintf(cmd.OutOrStdout(), "[%d/%d] %s: anw %s\n", i+1, len(steps), s.Dir, strings.Join(s.Args, " "))
		if err := os.Chdir(s.Dir); err != nil {
			return fmt.Errorf("%d番目のステップで失敗しました: %w", i+1, err)
		}
		err := runStep(s.Args, cfgPath)
		if errors.Is(err, errCanceled) {
			return err
		}
		if err != nil {
			return fmt.Errorf("%d番目のステップで失敗しました: %w", i+1, err)
		}
	}
	return nil
}

// runStep は anw を args で実行する。--config は snippet を実行したときと同じものを使う。
func runStep(args []string, cfgPath string) error {
	resetFlags()
	if cfgPath != "" {
		args = append([]string{"--config", cfgPath}, args...)
	}
	rootCmd.SetArgs(args)
	_, err := rootCmd.ExecuteC()
	return err
}

// validateSteps は、各ステップの dir が存在し、args が anw のコマンドとして正しいかを確かめる。
func validateSteps(steps []config.Step) error {
	defer resetFlags()
	var errs []error
	for i, s := range steps {
		if info, err := os.Stat(s.Dir); err != nil || !info.IsDir() {
			errs = append(errs, fmt.Errorf("%d番目のステップ: dir がありません: %s", i+1, s.Dir))
			continue
		}
		if err := validateArgs(s.Args); err != nil {
			errs = append(errs, fmt.Errorf("%d番目のステップ: anw %s: %w", i+1, strings.Join(s.Args, " "), err))
		}
	}
	return errors.Join(errs...)
}

// validateArgs は args を実行せずに、コマンド・フラグ・引数が正しいかを確かめる。
func validateArgs(args []string) error {
	sub, rest, err := rootCmd.Find(args)
	if err != nil {
		return err
	}
	// runCmd と直接比べると初期化が循環するので、名前で比べる
	if sub == rootCmd || sub.Name() == "run" {
		return fmt.Errorf("snippet から実行できるコマンドではありません")
	}
	resetFlags()
	if err := sub.ParseFlags(rest); err != nil {
		return err
	}
	if err := sub.ValidateArgs(sub.Flags().Args()); err != nil {
		return err
	}
	if err := sub.ValidateRequiredFlags(); err != nil {
		return err
	}
	return sub.ValidateFlagGroups()
}

// resetFlags はすべてのコマンドのフラグを初期値に戻す。
func resetFlags() {
	reset := func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	}
	rootCmd.PersistentFlags().VisitAll(reset)
	for _, sub := range rootCmd.Commands() {
		sub.Flags().VisitAll(reset)
	}
}

func init() {
	newVerb("run", "snippet に書いた anw のコマンドを順に実行する", "何を実行しますか？", "", []target{
		{name: "snippet", desc: "snippet に書いた anw のコマンドを順に実行", argHint: "名前", run: runSnippetSteps},
	})
}
