package aisession

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// source はエージェントごとのセッションの置き場所と読み取り方。
type source struct {
	// dir はホームからの相対パス
	dir string
	// depth は dir からセッションのファイルまでの階層。これ以外の階層のファイルは読まない
	depth int
	parse func(r io.Reader, path string) (Session, error)
}

var sources = []source{
	// projects/<プロジェクト>/<セッション>.jsonl
	{dir: filepath.Join(".claude", "projects"), depth: 2, parse: ParseClaude},
	// sessions/<年>/<月>/<日>/<セッション>.jsonl
	{dir: filepath.Join(".codex", "sessions"), depth: 4, parse: ParseCodex},
}

// Collect は home 配下のエージェントのセッションを集め、最終更新の新しい順に返す。
// 発言のないセッションと、読めないファイルは含めない。
func Collect(home string) ([]Session, error) {
	var result []Session
	for _, src := range sources {
		root := filepath.Join(home, src.dir)
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if path == root && os.IsNotExist(err) {
					return fs.SkipDir
				}
				// 読めないディレクトリは飛ばす
				if d != nil && d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			depth := len(strings.Split(rel, string(filepath.Separator)))
			if d.IsDir() {
				if path != root && depth >= src.depth {
					return fs.SkipDir
				}
				return nil
			}
			if depth != src.depth || filepath.Ext(path) != ".jsonl" {
				return nil
			}
			if s, ok := parseFile(path, src.parse); ok {
				result = append(result, s)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Updated.After(result[j].Updated) })
	return result, nil
}

func parseFile(path string, parse func(io.Reader, string) (Session, error)) (Session, bool) {
	f, err := os.Open(path)
	if err != nil {
		return Session{}, false
	}
	defer f.Close()
	s, err := parse(f, path)
	if err != nil || len(s.Messages) == 0 {
		return Session{}, false
	}
	return s, true
}
