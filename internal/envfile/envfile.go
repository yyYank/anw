package envfile

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// skipDirs は探索しないディレクトリ名。
var skipDirs = map[string]bool{".git": true, "node_modules": true, "vendor": true}

// Find は root 配下の .env と .env.* を探し、root からの相対パスで返す。
func Find(root string) ([]string, error) {
	var result []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !isEnvFile(d.Name()) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		result = append(result, rel)
		return nil
	})
	return result, err
}

func isEnvFile(name string) bool {
	return name == ".env" || strings.HasPrefix(name, ".env.")
}
