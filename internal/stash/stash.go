package stash

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Stash は1つの stash と、その元のブランチを表す。
type Stash struct {
	// Ref は stash@{n} の形の参照
	Ref        string
	Commit     string
	Time       time.Time
	Message    string
	Branch     string
	BranchGone bool
}

// Reason は元のブランチと何日前かを表示用の文字列にする。
func (s Stash) Reason(now time.Time) string {
	branch := s.Branch
	if s.BranchGone {
		branch += " (ブランチなし)"
	}
	days := int(now.Sub(s.Time).Hours() / 24)
	return fmt.Sprintf("%s, %d日前", branch, days)
}

// Parse は git stash list --format=%gd%x00%H%x00%ct%x00%gs の出力 out を読む。
// locals はローカルのブランチ名の集合で、元のブランチがここになければ BranchGone にする。
func Parse(out string, locals map[string]bool) []Stash {
	var result []Stash
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Split(line, "\x00")
		if len(fields) != 4 {
			continue
		}
		sec, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			continue
		}
		s := Stash{Ref: fields[0], Commit: fields[1], Time: time.Unix(sec, 0), Message: fields[3]}
		s.Branch = branchOf(s.Message)
		s.BranchGone = !locals[s.Branch]
		result = append(result, s)
	}
	return result
}

// branchOf は "WIP on <branch>: ..." や "On <branch>: ..." の形のメッセージから元のブランチを取り出す。
func branchOf(message string) string {
	rest, ok := strings.CutPrefix(message, "WIP on ")
	if !ok {
		rest, _ = strings.CutPrefix(message, "On ")
	}
	branch, _, _ := strings.Cut(rest, ":")
	return branch
}
