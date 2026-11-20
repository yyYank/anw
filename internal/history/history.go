package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yyYank/anw/internal/tidy"
)

// 操作の種類
const (
	KindTidy = "tidy"
)

// Record は直前の操作を元に戻すための記録。Kind で操作の種類を見分ける。
type Record struct {
	Kind  string      `json:"kind"`
	Moves []tidy.Move `json:"moves,omitempty"`
}

// Save は直前の操作として tidy の移動 moves を path に保存する。
func Save(path string, moves []tidy.Move) error {
	return saveRecord(path, Record{Kind: KindTidy, Moves: moves})
}

func saveRecord(path string, r Record) error {
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

// Undo は path に保存された直前の操作を元に戻し、戻した移動を返す。
// 移動先のディレクトリが空になった場合は削除する。
func Undo(path string) ([]tidy.Move, error) {
	r, err := Load(path)
	if err != nil {
		return nil, err
	}
	moves := r.Moves
	// 途中まで戻した状態にならないよう、先にすべての戻し先を確認する
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
		if err := os.Rename(m.To, m.From); err != nil {
			return undone, err
		}
		// 空でなければ失敗するので、エラーは無視してよい
		_ = os.Remove(filepath.Dir(m.To))
		undone = append(undone, m)
	}
	return undone, os.Remove(path)
}
