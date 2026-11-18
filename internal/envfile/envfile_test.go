package envfile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// テスト用ディレクトリに偽の env ファイルを作る
func writeFile(t *testing.T, root, rel string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("DUMMY=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// .env と .env.* を配下から探し、相対パスで返すこと
func TestFind_envファイルを探す(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".env")
	writeFile(t, root, ".env.local")
	writeFile(t, root, "app/.env.production")
	writeFile(t, root, "README.md")
	writeFile(t, root, "app/env.txt")
	writeFile(t, root, ".envrc")

	got, err := Find(root)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{".env", ".env.local", filepath.Join("app", ".env.production")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// .git・node_modules・vendor 配下は探さないこと
func TestFind_除外ディレクトリは探さない(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".git/.env")
	writeFile(t, root, "node_modules/pkg/.env")
	writeFile(t, root, "vendor/.env.local")
	writeFile(t, root, "app/.env")

	got, err := Find(root)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{filepath.Join("app", ".env")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
