package branch

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// Branch は削除できるローカルブランチと、その理由を表す。
type Branch struct {
	Name   string
	Merged bool
	Gone   bool
}

// Reason は削除できる理由を表示用の文字列にする。
func (b Branch) Reason() string {
	var rs []string
	if b.Merged {
		rs = append(rs, "マージ済み")
	}
	if b.Gone {
		rs = append(rs, "リモート削除済み")
	}
	return strings.Join(rs, ", ")
}

// refsFormat は for-each-ref で、現在のブランチの印・名前・upstream との差を NUL 区切りで出す書式。
const refsFormat = "%(HEAD)%00%(refname:short)%00%(upstream:track)"

// Collect は for-each-ref の出力 refs と、git branch --merged の出力 merged から、
// 削除できるブランチを名前順に返す。現在のブランチと defaultBranch は含めない。
func Collect(refs, merged, defaultBranch string) []Branch {
	isMerged := map[string]bool{}
	for _, name := range strings.Split(merged, "\n") {
		isMerged[strings.TrimSpace(name)] = true
	}

	var result []Branch
	for _, line := range strings.Split(refs, "\n") {
		fields := strings.Split(line, "\x00")
		if len(fields) != 3 {
			continue
		}
		head, name, track := fields[0], fields[1], fields[2]
		if head == "*" || name == defaultBranch {
			continue
		}
		b := Branch{Name: name, Merged: isMerged[name], Gone: track == "[gone]"}
		if b.Merged || b.Gone {
			result = append(result, b)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

// List はカレントディレクトリのリポジトリで、削除できるローカルブランチを返す。
// デフォルトブランチは origin/HEAD から決める。fetch はしない。
func List() ([]Branch, error) {
	remoteHead, err := git("symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err != nil {
		return nil, fmt.Errorf("デフォルトブランチが分かりません (git remote set-head origin --auto で設定できます): %w", err)
	}
	remoteHead = strings.TrimSpace(remoteHead)
	refs, err := git("for-each-ref", "--format="+refsFormat, "refs/heads")
	if err != nil {
		return nil, err
	}
	merged, err := git("branch", "--merged", remoteHead, "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	return Collect(refs, merged, strings.TrimPrefix(remoteHead, "origin/")), nil
}

// Delete はローカルブランチを削除する。マージされていないブランチは消さない。
// force なら、マージされていなくても強制削除する。
func Delete(name string, force bool) error {
	args := []string{"branch", "--delete", name}
	if force {
		args = []string{"branch", "--delete", "--force", name}
	}
	_, err := git(args...)
	return err
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
