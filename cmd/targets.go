package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// target は動詞で扱う対象。フラグの名前・説明と、その対象を選んだときの処理を持つ。
type target struct {
	name string
	desc string
	// argHint は値を取る対象のヘルプでの表記。値を取らなければ空
	argHint string
	run     func(cmd *cobra.Command, args []string) error
}

// verb は対象を取る動詞と、その動詞で使える対象の一覧。
type verb struct {
	cmd     *cobra.Command
	targets []target
	flags   map[string]*bool
}

// verbs は対象を取る動詞の一覧。フラグとヘルプはここから作る。
var verbs []*verb

// newVerb は targets からフラグ・ヘルプ・処理の振り分けを組み立てた動詞を作り、ルートに登録する。
// useSuffix はヘルプの対象の後ろに付ける表記。
func newVerb(name, short, useSuffix string, targets []target) *verb {
	v := &verb{targets: targets, flags: map[string]*bool{}}

	var uses, names []string
	for _, t := range targets {
		use := "--" + t.name
		if t.argHint != "" {
			use += " " + t.argHint
		}
		uses = append(uses, use)
		names = append(names, t.name)
	}
	use := name + " " + strings.Join(uses, " | ")
	if useSuffix != "" {
		use += " " + useSuffix
	}

	maxArgs := 0
	for _, t := range targets {
		if t.argHint != "" {
			maxArgs = 1
		}
	}

	v.cmd = &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.MaximumNArgs(maxArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			t := v.chosen()
			if len(args) > 0 && t.argHint == "" {
				return fmt.Errorf("%s 以外には引数を指定できません", v.argTargets())
			}
			return t.run(cmd, args)
		},
	}
	for _, t := range targets {
		v.flags[t.name] = v.cmd.Flags().Bool(t.name, false, t.desc)
	}
	v.cmd.MarkFlagsOneRequired(names...)
	if len(names) > 1 {
		v.cmd.MarkFlagsMutuallyExclusive(names...)
	}

	verbs = append(verbs, v)
	rootCmd.AddCommand(v.cmd)
	return v
}

// chosen はフラグで指定された対象を返す。
func (v *verb) chosen() target {
	for _, t := range v.targets {
		if *v.flags[t.name] {
			return t
		}
	}
	return v.targets[0]
}

// argTargets は値を取る対象のフラグを "--port" のように並べる。
func (v *verb) argTargets() string {
	var names []string
	for _, t := range v.targets {
		if t.argHint != "" {
			names = append(names, "--"+t.name)
		}
	}
	return strings.Join(names, ", ")
}
