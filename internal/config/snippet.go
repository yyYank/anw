package config

import (
	"errors"
	"fmt"
)

// Step は snippet の1ステップ。Dir に移動してから anw を Args で実行する。
type Step struct {
	Dir  string   `toml:"dir"`
	Args []string `toml:"args"`
}

// ValidateSnippet は name の snippet の書き方を確かめ、ステップを返す。
// dir が存在するかや args が anw のコマンドとして正しいかは、呼び出し元で確かめる。
func (c Config) ValidateSnippet(name string) ([]Step, error) {
	steps, ok := c.Snippets[name]
	if !ok {
		return nil, fmt.Errorf("snippet %q がありません", name)
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("snippet %q にステップがありません", name)
	}
	var errs []error
	for i, s := range steps {
		switch {
		case s.Dir == "":
			errs = append(errs, fmt.Errorf("%d番目のステップに dir がありません", i+1))
		case len(s.Args) == 0:
			errs = append(errs, fmt.Errorf("%d番目のステップに args がありません", i+1))
		case s.Args[0] == "run":
			errs = append(errs, fmt.Errorf("%d番目のステップ: snippet から run は呼べません", i+1))
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("snippet %q: %w", name, errors.Join(errs...))
	}
	return steps, nil
}
