package tidy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Move は1件のファイル移動を表す。
type Move struct {
	From string
	To   string
}

// By は振り分けの基準を表す。
type By string

const (
	// ByDate は更新日時の yyyy-mm-dd ディレクトリに振り分ける。
	ByDate By = "date"
	// ByMonth は更新日時の yyyy-mm/yyyy-mm-dd ディレクトリに振り分ける。
	// 直下の yyyy-mm-dd ディレクトリも yyyy-mm ディレクトリの下にまとめる。
	ByMonth By = "month"
)

const dateLayout = "2006-01-02"

// ParseBy は --by の値を By にする。対応していない値ならエラーにする。
func ParseBy(s string) (By, error) {
	switch by := By(s); by {
	case ByDate, ByMonth:
		return by, nil
	}
	return "", fmt.Errorf("--by に指定できるのは date か month です: %q", s)
}

// Run は dir 直下のファイルを、by の基準で振り分ける。
func Run(dir string, by By) ([]Move, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var moves []Move
	if by == ByMonth {
		for _, e := range entries {
			if !isDateDir(e) {
				continue
			}
			moved, err := moveDateDir(dir, e.Name())
			moves = append(moves, moved...)
			if err != nil {
				return moves, err
			}
		}
	}
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return moves, err
		}
		day := info.ModTime().Format(dateLayout)
		destDir := filepath.Join(dir, day)
		if by == ByMonth {
			destDir = filepath.Join(dir, day[:len("2006-01")], day)
		}
		m, err := moveInto(filepath.Join(dir, e.Name()), destDir)
		if err != nil {
			return moves, err
		}
		moves = append(moves, m)
	}
	return moves, nil
}

// isDateDir は e が yyyy-mm-dd という名前のディレクトリかを返す。
func isDateDir(e os.DirEntry) bool {
	if !e.IsDir() {
		return false
	}
	_, err := time.Parse(dateLayout, e.Name())
	return err == nil
}

// moveDateDir は dir 直下の yyyy-mm-dd ディレクトリ name を yyyy-mm ディレクトリの下に移動する。
// 移動先に同じ名前のディレクトリが既にあれば、中身を1件ずつ移して統合し、空になった元のディレクトリを消す。
func moveDateDir(dir, name string) ([]Move, error) {
	src := filepath.Join(dir, name)
	dest := filepath.Join(dir, name[:len("2006-01")], name)
	if _, err := os.Lstat(dest); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return nil, err
		}
		if err := os.Rename(src, dest); err != nil {
			return nil, err
		}
		return []Move{{From: src, To: dest}}, nil
	} else if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return nil, err
	}
	var moves []Move
	for _, e := range entries {
		m, err := moveInto(filepath.Join(src, e.Name()), dest)
		if err != nil {
			return moves, err
		}
		moves = append(moves, m)
	}
	return moves, os.Remove(src)
}

// moveInto は from を destDir に移動する。destDir が無ければ作り、同名があれば連番を付ける。
func moveInto(from, destDir string) (Move, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return Move{}, err
	}
	to, err := uniquePath(destDir, filepath.Base(from))
	if err != nil {
		return Move{}, err
	}
	if err := os.Rename(from, to); err != nil {
		return Move{}, err
	}
	return Move{From: from, To: to}, nil
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
