package cmd

import (
	"strings"
	"testing"
)

// 動詞ごとの対象の一覧から、その動詞のフラグとヘルプが作られていること
func TestVerbs_フラグとヘルプは対象の一覧から作られる(t *testing.T) {
	for _, v := range verbs {
		t.Run(v.cmd.Name(), func(t *testing.T) {
			for _, tg := range v.targets {
				f := v.cmd.Flags().Lookup(tg.name)
				if f == nil {
					t.Errorf("--%s のフラグがない", tg.name)
					continue
				}
				if f.Usage != tg.desc {
					t.Errorf("--%s の説明が %q、want %q", tg.name, f.Usage, tg.desc)
				}
				if !strings.Contains(v.cmd.Use, "--"+tg.name) {
					t.Errorf("ヘルプに --%s がない: %q", tg.name, v.cmd.Use)
				}
			}
		})
	}
}

// 対象を取る動詞がすべて一覧に入っていること
func TestVerbs_対象を取る動詞がそろっている(t *testing.T) {
	got := map[string]bool{}
	for _, v := range verbs {
		got[v.cmd.Name()] = true
	}
	for _, name := range []string{"view", "edit", "kill", "delete"} {
		if !got[name] {
			t.Errorf("%s が一覧にない", name)
		}
	}
}
