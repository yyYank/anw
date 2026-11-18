package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yyYank/anw/internal/tidy"
)

// Save は直前の操作として moves を path に保存する。
func Save(path string, moves []tidy.Move) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(moves)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Load は path に保存された直前の操作を読み込む。
func Load(path string) ([]tidy.Move, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var moves []tidy.Move
	if err := json.Unmarshal(data, &moves); err != nil {
		return nil, err
	}
	return moves, nil
}

// Undo は path に保存された直前の操作を元に戻し、戻した移動を返す。
// 移動先のディレクトリが空になった場合は削除する。
func Undo(path string) ([]tidy.Move, error) {
	moves, err := Load(path)
	if err != nil {
		return nil, err
	}
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
