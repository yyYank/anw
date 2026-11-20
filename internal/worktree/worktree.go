package worktree

import (
	"sort"
	"strings"

	"github.com/yyYank/anw/internal/branch"
)

// Worktree は消してよい worktree と、その理由を表す。
type Worktree struct {
	Path    string
	Branch  string
	Missing bool
	Merged  bool
	Gone    bool
}

// Reason は消してよい理由を表示用の文字列にする。
func (w Worktree) Reason() string {
	var rs []string
	if w.Missing {
		rs = append(rs, "ディレクトリなし")
	}
	if w.Merged {
		rs = append(rs, "マージ済み")
	}
	if w.Gone {
		rs = append(rs, "リモート削除済み")
	}
	return strings.Join(rs, ", ")
}

// Collect は git worktree list --porcelain の出力 porcelain から、消してよい worktree をパス順に返す。
// ディレクトリが消えたものと、ブランチが deletable に含まれるものを選ぶ。
// メインの worktree（先頭）と current は含めない。
func Collect(porcelain, current string, deletable []branch.Branch) []Worktree {
	byName := map[string]branch.Branch{}
	for _, b := range deletable {
		byName[b.Name] = b
	}

	var result []Worktree
	for i, block := range strings.Split(strings.TrimSpace(porcelain), "\n\n") {
		var w Worktree
		for _, line := range strings.Split(block, "\n") {
			key, value, _ := strings.Cut(line, " ")
			switch key {
			case "worktree":
				w.Path = value
			case "branch":
				w.Branch = strings.TrimPrefix(value, "refs/heads/")
			case "prunable":
				w.Missing = true
			}
		}
		if i == 0 || w.Path == "" || w.Path == current {
			continue
		}
		if b, ok := byName[w.Branch]; ok {
			w.Merged, w.Gone = b.Merged, b.Gone
		}
		if w.Missing || w.Merged || w.Gone {
			result = append(result, w)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}
