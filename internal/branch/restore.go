package branch

import (
	"fmt"
	"strings"
)

// Saved は削除したブランチを作り直すための記録。
type Saved struct {
	// Repo はリポジトリの作業ディレクトリの絶対パス
	Repo     string `json:"repo"`
	Name     string `json:"name"`
	Commit   string `json:"commit"`
	Upstream string `json:"upstream,omitempty"`
}

// Snapshot はカレントディレクトリのリポジトリで、ブランチ name を作り直すための記録を返す。
func Snapshot(name string) (Saved, error) {
	repo, err := git("rev-parse", "--show-toplevel")
	if err != nil {
		return Saved{}, err
	}
	commit, err := git("rev-parse", "--verify", "refs/heads/"+name)
	if err != nil {
		return Saved{}, err
	}
	// 上流ブランチがなければ空のままにする
	upstream, _ := git("rev-parse", "--abbrev-ref", name+"@{upstream}")
	return Saved{
		Repo:     strings.TrimSpace(repo),
		Name:     name,
		Commit:   strings.TrimSpace(commit),
		Upstream: strings.TrimSpace(upstream),
	}, nil
}

// CheckRestore は s を作り直せるかを確かめる。同じ名前のブランチがある、コミットがもうない場合はエラーにする。
func CheckRestore(s Saved) error {
	if _, err := git("-C", s.Repo, "rev-parse", "--verify", "--quiet", "refs/heads/"+s.Name); err == nil {
		return fmt.Errorf("同じ名前のブランチがすでにあります: %s", s.Name)
	}
	if _, err := git("-C", s.Repo, "cat-file", "-e", s.Commit+"^{commit}"); err != nil {
		return fmt.Errorf("コミットがもうありません: %s (%s)", s.Name, s.Commit)
	}
	return nil
}

// Restore は s のブランチを同じコミットで作り直し、上流ブランチがまだあれば設定し直す。
func Restore(s Saved) error {
	if _, err := git("-C", s.Repo, "branch", s.Name, s.Commit); err != nil {
		return err
	}
	if s.Upstream == "" {
		return nil
	}
	if _, err := git("-C", s.Repo, "rev-parse", "--verify", "--quiet", "refs/remotes/"+s.Upstream); err != nil {
		return nil
	}
	_, err := git("-C", s.Repo, "branch", "--set-upstream-to="+s.Upstream, s.Name)
	return err
}
