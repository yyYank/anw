package cmd

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/yyYank/anw/internal/choose"
)

// テストで差し替えられるよう変数にしておく
var chooseTarget = runChooser

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
	cmd      *cobra.Command
	targets  []target
	flags    map[string]*bool
	question string
}

// verbs は対象を取る動詞の一覧。フラグとヘルプはここから作る。
var verbs []*verb

// newVerb は targets からフラグ・ヘルプ・処理の振り分けを組み立てた動詞を作り、ルートに登録する。
// question は対象を省略したときに出す問いかけ。useSuffix はヘルプの対象の後ろに付ける表記。
func newVerb(name, short, question, useSuffix string, targets []target) *verb {
	v := &verb{targets: targets, flags: map[string]*bool{}, question: name + ": " + question}

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
			t, ok := v.flagged()
			if !ok {
				if len(args) > 0 {
					return errors.New("値を渡すときは対象のフラグも指定してください")
				}
				var err error
				if t, ok, err = v.ask(); err != nil {
					return err
				}
				if !ok {
					return errCanceled
				}
			}
			if len(args) > 0 && t.argHint == "" {
				return fmt.Errorf("%s 以外には引数を指定できません", v.argTargets())
			}
			return t.run(cmd, args)
		},
	}
	for _, t := range targets {
		v.flags[t.name] = v.cmd.Flags().Bool(t.name, false, t.desc)
	}
	if len(names) > 1 {
		v.cmd.MarkFlagsMutuallyExclusive(names...)
	}
	v.cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		if strings.HasPrefix(err.Error(), "unknown flag") {
			return fmt.Errorf("%s で使えるのは %s です", name, v.targetFlags(", "))
		}
		return err
	})

	verbs = append(verbs, v)
	rootCmd.AddCommand(v.cmd)
	return v
}

// flagged はフラグで指定された対象を返す。指定がなければ ok は false。
func (v *verb) flagged() (target, bool) {
	for _, t := range v.targets {
		if *v.flags[t.name] {
			return t, true
		}
	}
	return target{}, false
}

// ask は使える対象を一覧して選ばせる。対象が1つしかなければ選ばせずにそれを返す。
func (v *verb) ask() (target, bool, error) {
	if len(v.targets) == 1 {
		return v.targets[0], true, nil
	}
	items := make([]choose.Item, len(v.targets))
	for i, t := range v.targets {
		items[i] = choose.Item{Name: t.name, Desc: t.desc}
	}
	name, ok, err := chooseTarget(v.question, items)
	if err != nil || !ok {
		return target{}, false, err
	}
	for _, t := range v.targets {
		if t.name == name {
			return t, true, nil
		}
	}
	return target{}, false, fmt.Errorf("不明な対象です: %s", name)
}

// targetFlags は使える対象のフラグを "--env" のように sep でつないで並べる。
func (v *verb) targetFlags(sep string) string {
	var names []string
	for _, t := range v.targets {
		names = append(names, "--"+t.name)
	}
	return strings.Join(names, sep)
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

// runChooser は対象を選ぶ画面を表示し、選ばれた対象の名前を返す。
func runChooser(title string, items []choose.Item) (string, bool, error) {
	m, err := tea.NewProgram(choose.New(title, items)).Run()
	if err != nil {
		return "", false, err
	}
	name, ok := m.(choose.Model).Chosen()
	return name, ok, nil
}
