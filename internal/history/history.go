package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yyYank/anw/internal/branch"
	"github.com/yyYank/anw/internal/stash"
	"github.com/yyYank/anw/internal/tidy"
)

// 操作の種類
const (
	KindTidy         = "tidy"
	KindDeleteBranch = "delete-branch"
	KindDropStash    = "drop-stash"
	// KindSnippet は snippet で行った複数の操作。Steps に行った順に入る
	KindSnippet = "snippet"
)

// Record は直前の操作を元に戻すための記録。Kind で操作の種類を見分ける。
type Record struct {
	Kind     string         `json:"kind"`
	Moves    []tidy.Move    `json:"moves,omitempty"`
	Branches []branch.Saved `json:"branches,omitempty"`
	Stashes  []stash.Saved  `json:"stashes,omitempty"`
	Steps    []Record       `json:"steps,omitempty"`
}

// Save は直前の操作として tidy の移動 moves を path に保存する。
func Save(path string, moves []tidy.Move) error {
	return SaveRecord(path, Record{Kind: KindTidy, Moves: moves})
}

// SaveBranches は直前の操作として、削除したブランチの記録 saved を path に保存する。
func SaveBranches(path string, saved []branch.Saved) error {
	return SaveRecord(path, Record{Kind: KindDeleteBranch, Branches: saved})
}

// SaveStashes は直前の操作として、消した stash の記録 saved を path に保存する。
func SaveStashes(path string, saved []stash.Saved) error {
	return SaveRecord(path, Record{Kind: KindDropStash, Stashes: saved})
}

// SaveRecord は直前の操作として r を path に保存する。
func SaveRecord(path string, r Record) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Load は path に保存された直前の操作を読み込む。
// 移動の配列だけの古い形式は、tidy の記録として読む。
func Load(path string) (Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Record{}, err
	}
	var moves []tidy.Move
	if json.Unmarshal(data, &moves) == nil {
		return Record{Kind: KindTidy, Moves: moves}, nil
	}
	var r Record
	if err := json.Unmarshal(data, &r); err != nil {
		return Record{}, err
	}
	return r, nil
}

// Undo は path に保存された直前の操作を元に戻し、戻した記録を返す。戻せたら履歴を消す。
// 途中まで戻した状態にならないよう、先にすべて戻せるかを確かめ、1件でも戻せなければ何もしない。
func Undo(path string) (Record, error) {
	r, err := Load(path)
	if err != nil {
		return Record{}, err
	}
	if r.Kind == KindTidy {
		// 途中で失敗しても、戻せた分を呼び出し元に伝える
		moves, err := undoMoves(r.Moves)
		if err != nil {
			return Record{Kind: KindTidy, Moves: moves}, err
		}
	} else if err := undoRecord(r); err != nil {
		return Record{}, err
	}
	return r, os.Remove(path)
}

// undoRecord は r の操作を元に戻す。
func undoRecord(r Record) error {
	switch r.Kind {
	case KindTidy:
		if _, err := undoMoves(r.Moves); err != nil {
			return err
		}
	case KindDeleteBranch:
		if err := undoBranches(r.Branches); err != nil {
			return err
		}
	case KindDropStash:
		for _, s := range r.Stashes {
			if err := stash.CheckRestore(s); err != nil {
				return err
			}
		}
		if err := stash.Restore(r.Stashes); err != nil {
			return err
		}
	case KindSnippet:
		for i := len(r.Steps) - 1; i >= 0; i-- {
			if err := undoRecord(r.Steps[i]); err != nil {
				return fmt.Errorf("snippet の%d番目の操作を戻せませんでした: %w", i+1, err)
			}
		}
	default:
		return fmt.Errorf("元に戻せない種類の操作です: %q", r.Kind)
	}
	return nil
}

// undoMoves は tidy の移動を逆向きに戻し、戻した移動を返す。移動先のディレクトリが空になった場合は削除する。
func undoMoves(moves []tidy.Move) ([]tidy.Move, error) {
	for _, m := range moves {
		if _, err := os.Lstat(m.From); err == nil {
			return nil, fmt.Errorf("元の場所に同名のファイルがあります: %s", m.From)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	var undone []tidy.Move
	for i := len(moves) - 1; i >= 0; i-- {
		m := moves[i]
		// 移動元のディレクトリが中身ごと移された後でも戻せるよう、作り直す
		if err := os.MkdirAll(filepath.Dir(m.From), 0o755); err != nil {
			return undone, err
		}
		if err := os.Rename(m.To, m.From); err != nil {
			return undone, err
		}
		removeEmptyDirs(filepath.Dir(m.To), m.From)
		undone = append(undone, m)
	}
	return undone, nil
}

// undoBranches は削除したブランチを作り直す。
func undoBranches(saved []branch.Saved) error {
	for _, s := range saved {
		if err := branch.CheckRestore(s); err != nil {
			return err
		}
	}
	for _, s := range saved {
		if err := branch.Restore(s); err != nil {
			return err
		}
	}
	return nil
}

// removeEmptyDirs は dir から上の階層へ、空のディレクトリを消していく。
// from を含むディレクトリまで来たら止める。
func removeEmptyDirs(dir, from string) {
	for !strings.HasPrefix(from, dir+string(filepath.Separator)) {
		// 空でなければ失敗するので、そこで止めてよい
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}
