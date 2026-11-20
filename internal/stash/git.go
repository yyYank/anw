package stash

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// Saved は消した stash を作り直すための記録。
type Saved struct {
	// Repo はリポジトリの作業ディレクトリの絶対パス
	Repo    string `json:"repo"`
	Commit  string `json:"commit"`
	Message string `json:"message"`
	// Index は消したときの stash の番号。作り直すときの並びに使う
	Index int `json:"index"`
}

// List はカレントディレクトリのリポジトリの stash を、新しい順にすべて返す。
func List() ([]Stash, error) {
	out, err := git("stash", "list", "--format=%gd%x00%H%x00%ct%x00%gs")
	if err != nil {
		return nil, err
	}
	refs, err := git("for-each-ref", "--format=%(refname:short)", "refs/heads")
	if err != nil {
		return nil, err
	}
	locals := map[string]bool{}
	for _, name := range strings.Split(refs, "\n") {
		locals[strings.TrimSpace(name)] = true
	}
	return Parse(out, locals), nil
}

// Drop は stashes を消し、作り直すための記録を返す。
// 消すと後ろの番号がずれるので、番号の大きいものから消す。
func Drop(stashes []Stash) ([]Saved, error) {
	repo, err := git("rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	targets := make([]Saved, len(stashes))
	for i, s := range stashes {
		idx, err := index(s.Ref)
		if err != nil {
			return nil, err
		}
		targets[i] = Saved{Repo: strings.TrimSpace(repo), Commit: s.Commit, Message: s.Message, Index: idx}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].Index > targets[j].Index })

	var dropped []Saved
	for _, t := range targets {
		if _, err := git("stash", "drop", "-q", fmt.Sprintf("stash@{%d}", t.Index)); err != nil {
			return dropped, err
		}
		dropped = append(dropped, t)
	}
	return dropped, nil
}

// CheckRestore は s を作り直せるかを確かめる。コミットがもうなければエラーにする。
func CheckRestore(s Saved) error {
	if _, err := git("-C", s.Repo, "cat-file", "-e", s.Commit+"^{commit}"); err != nil {
		return fmt.Errorf("stash のコミットがもうありません: %s (%s)", s.Message, s.Commit)
	}
	return nil
}

// Restore は saved の stash を同じコミットとメッセージで作り直す。
// git stash store は一番上に積むので、古いもの（番号の大きいもの）から積んで前後関係を保つ。
func Restore(saved []Saved) error {
	sorted := append([]Saved(nil), saved...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Index > sorted[j].Index })
	for _, s := range sorted {
		if _, err := git("-C", s.Repo, "stash", "store", "-m", s.Message, s.Commit); err != nil {
			return err
		}
	}
	return nil
}

// index は stash@{n} の n を返す。
func index(ref string) (int, error) {
	n, ok := strings.CutPrefix(ref, "stash@{")
	if !ok || !strings.HasSuffix(n, "}") {
		return 0, fmt.Errorf("stash の参照ではありません: %q", ref)
	}
	return strconv.Atoi(strings.TrimSuffix(n, "}"))
}

// git は git コマンドを実行して標準出力を返す。失敗したら標準エラーの内容をエラーにする。
func git(args ...string) (string, error) {
	var stderr bytes.Buffer
	c := exec.Command("git", args...)
	c.Stderr = &stderr
	out, err := c.Output()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		return "", err
	}
	return string(out), nil
}
