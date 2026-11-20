package aisession

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RepoOf は dir が属するリポジトリを、メインの作業ディレクトリのパスで返す。
// worktree は git の共通ディレクトリで判定するので、別ディレクトリでも同じリポジトリになる。
// git のリポジトリでなければ dir 自身を返す。dir が存在しなければ ok は false。
func RepoOf(dir string) (repo string, ok bool) {
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return dir, false
	}
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--path-format=absolute", "--git-common-dir").Output()
	if err != nil {
		return dir, true
	}
	common := strings.TrimSpace(string(out))
	if filepath.Base(common) == ".git" {
		return filepath.Dir(common), true
	}
	// bare リポジトリなどは共通ディレクトリそのものを使う
	return common, true
}
