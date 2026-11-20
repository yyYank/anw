package dotfiles

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// maxDepth は ~/.config からの階層の上限
	maxDepth = 3
	// maxSize はこれを超えるファイルを設定ファイルとみなさない大きさ
	maxSize = 1 << 20
	// sniffSize はテキストかどうかを判定するために読む大きさ
	sniffSize = 8 << 10
)

// skipDirs は ~/.config 配下で中を見ないディレクトリ名。
var skipDirs = map[string]bool{"cache": true, "Cache": true, "logs": true}

// Find は home 直下のドットファイルと ~/.config 配下の設定ファイルを探し、home からの相対パスで返す。
// テキストでないもの、大きすぎるものは含めない。
func Find(home string) ([]string, error) {
	var result []string

	entries, err := os.ReadDir(home)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			continue
		}
		// シンボリックリンクはリンク先で判定する
		if isConfigFile(filepath.Join(home, e.Name())) {
			result = append(result, e.Name())
		}
	}

	configFiles, err := walk(filepath.Join(home, ".config"), ".config", 1)
	if err != nil {
		return nil, err
	}
	result = append(result, configFiles...)

	sort.Strings(result)
	return result, nil
}

// walk は dir の中の設定ファイルを rel からの相対パスで返す。depth は dir の中身の階層。
// dotfiles 管理でよく使われるので、シンボリックリンクのディレクトリもたどる。
// 階層の上限があるので、リンクが循環していても止まる。
func walk(dir, rel string, depth int) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// ~/.config がない、または読めないディレクトリは飛ばす。~/.config 自体が読めないときだけエラーにする
		if depth > 1 || errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var result []string
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		info, err := os.Stat(path)
		if err != nil {
			// リンク切れなどは飛ばす
			continue
		}
		if info.IsDir() {
			if skipDirs[e.Name()] || depth >= maxDepth {
				continue
			}
			sub, err := walk(path, filepath.Join(rel, e.Name()), depth+1)
			if err != nil {
				return nil, err
			}
			result = append(result, sub...)
			continue
		}
		if isConfigFile(path) {
			result = append(result, filepath.Join(rel, e.Name()))
		}
	}
	return result, nil
}

// isConfigFile は path が設定ファイルとして扱える、大きすぎないテキストファイルかを返す。
func isConfigFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxSize {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, sniffSize)
	n, err := io.ReadFull(f, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return false
	}
	return !bytes.Contains(buf[:n], []byte{0})
}
