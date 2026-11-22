package tidy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Move は1件のファイル移動を表す。
type Move struct {
	From string
	To   string
}

// By は振り分けの基準を表す。
type By string

// ByDate は更新日時の yyyy-mm-dd ディレクトリに振り分ける。
const ByDate By = "date"

// ParseBy は --by の値を By にする。対応していない値ならエラーにする。
func ParseBy(s string) (By, error) {
	if By(s) != ByDate {
		return "", fmt.Errorf("--by に指定できるのは date のみです: %q", s)
	}
	return By(s), nil
}

// Run は dir 直下のファイルを、by の基準で振り分ける。
func Run(dir string, by By) ([]Move, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var moves []Move
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return moves, err
		}
		destDir := filepath.Join(dir, info.ModTime().Format("2006-01-02"))
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return moves, err
		}
		from := filepath.Join(dir, e.Name())
		to, err := uniquePath(destDir, e.Name())
		if err != nil {
			return moves, err
		}
		if err := os.Rename(from, to); err != nil {
			return moves, err
		}
		moves = append(moves, Move{From: from, To: to})
	}
	return moves, nil
}

// uniquePath は dir 内で name と衝突しないパスを返す。
// 衝突する場合は "report (1).pdf" のように連番を付ける。
func uniquePath(dir, name string) (string, error) {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	candidate := filepath.Join(dir, name)
	for i := 1; ; i++ {
		_, err := os.Lstat(candidate)
		if os.IsNotExist(err) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
		candidate = filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
	}
}
