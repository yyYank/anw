package config

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Step は snippet の1ステップ。Dir に移動してから anw を Args で実行する。
type Step struct {
	Dir  string   `toml:"dir"`
	Args []string `toml:"args"`
}

// ValidateSnippet は name の snippet の書き方を確かめ、ステップを返す。
// dir が存在するかや args が anw のコマンドとして正しいかは、呼び出し元で確かめる。
func (c Config) ValidateSnippet(name string) ([]Step, error) {
	if err := c.snippetErrs[name]; err != nil {
		return nil, fmt.Errorf("snippet %q: %w", name, err)
	}
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

// SnippetDir は設定ファイル configPath の隣の、snippet のファイルを置くディレクトリを返す。
func SnippetDir(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "snippets")
}

// SnippetNames は読めなかったものも含めて、snippet の名前を名前順に返す。
func (c Config) SnippetNames() []string {
	names := slices.Collect(maps.Keys(c.Snippets))
	for name := range c.snippetErrs {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// ParseSnippet は snippet のファイルの中身を読み、ステップを返す。
// cd の行で以降のステップの dir を決め、anw の行を1ステップにする。空行と # で始まる行は読み飛ばす。
func ParseSnippet(text string) ([]Step, error) {
	var steps []Step
	var errs []error
	dir := ""
	for i, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		switch fields[0] {
		case "cd":
			if len(fields) != 2 {
				errs = append(errs, fmt.Errorf("%d行目: cd には dir を1つ書いてください", i+1))
				continue
			}
			d, err := expandPath(fields[1])
			if err != nil {
				return nil, err
			}
			dir = d
		case "anw":
			if dir == "" {
				errs = append(errs, fmt.Errorf("%d行目: 先に cd で dir を書いてください", i+1))
				continue
			}
			steps = append(steps, Step{Dir: dir, Args: fields[1:]})
		default:
			errs = append(errs, fmt.Errorf("%d行目: cd か anw で始めてください", i+1))
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return steps, nil
}

// loadSnippetFiles は dir のファイルを、ファイル名の snippet として読む。
// 読めないファイルや toml と名前が重なるものは、ほかの設定を使えるよう snippetErrs に入れておく。
func (c *Config) loadSnippetFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || strings.HasPrefix(name, ".") {
			continue
		}
		path := filepath.Join(dir, name)
		if _, ok := c.Snippets[name]; ok {
			c.setSnippetErr(name, fmt.Errorf("config.toml と %s の両方にあります", path))
			delete(c.Snippets, name)
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			c.setSnippetErr(name, err)
			continue
		}
		steps, err := ParseSnippet(string(data))
		if err != nil {
			c.setSnippetErr(name, fmt.Errorf("%s: %w", path, err))
			continue
		}
		if c.Snippets == nil {
			c.Snippets = map[string][]Step{}
		}
		c.Snippets[name] = steps
	}
	return nil
}

func (c *Config) setSnippetErr(name string, err error) {
	if c.snippetErrs == nil {
		c.snippetErrs = map[string]error{}
	}
	c.snippetErrs[name] = err
}
