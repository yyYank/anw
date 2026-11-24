package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/yyYank/anw/internal/aisession"
)

var sampleSessions = []aisession.Session{
	{
		Agent: "claude", ID: "s1", Cwd: "/work/app", Branch: "fix/login",
		Updated:     time.Date(2026, 11, 3, 10, 0, 0, 0, time.Local),
		FirstPrompt: "ログインのバグを直して",
		Messages: []aisession.Message{
			{Role: aisession.RoleUser, Text: "ログインのバグを直して"},
			{Role: aisession.RoleTool, Text: "Read"},
			{Role: aisession.RoleAssistant, Text: "直しました"},
		},
	},
	{
		Agent: "codex", ID: "c1", Cwd: "/work/app-wt", Branch: "feature/x",
		Updated:     time.Date(2026, 11, 2, 9, 0, 0, 0, time.Local),
		FirstPrompt: "API のテストを足して",
		Messages:    []aisession.Message{{Role: aisession.RoleUser, Text: "API のテストを足して"}},
	},
	{
		Agent: "claude", ID: "s2", Cwd: "/work/web", Branch: "main",
		Updated:     time.Date(2026, 11, 1, 8, 0, 0, 0, time.Local),
		FirstPrompt: "README を書いて",
		Messages:    []aisession.Message{{Role: aisession.RoleUser, Text: "README を書いて"}},
	},
}

// view --ai で、すべてのセッションを日時・種類・作業ディレクトリ・ブランチ・最初の発言付きで一覧すること
func TestView_aiはセッションを一覧する(t *testing.T) {
	cli := setupCLI(t)
	cli.sessions = sampleSessions
	cli.sessionPick = -1

	if err := cli.run("view", "--ai"); err != nil {
		t.Fatal(err)
	}

	p := cli.sessionPicker
	if p == nil || len(p.labels) != 3 {
		t.Fatalf("一覧が表示されていない: %+v", p)
	}
	for _, s := range []string{"11-03 10:00", "claude", "work/app", "fix/login", "ログインのバグを直して"} {
		if !strings.Contains(p.labels[0], s) {
			t.Errorf("%q が一覧にない: %q", s, p.labels[0])
		}
	}
	if p.keys[0] != "work/app" || p.fullKeys[0] != "/work/app" {
		t.Errorf("絞り込みの文字列が表示と合っていない: %q, %q", p.keys[0], p.fullKeys[0])
	}
	for _, s := range []string{"ログインのバグを直して", "Read", "直しました"} {
		if !strings.Contains(p.contents[0], s) {
			t.Errorf("%q がプレビューにない: %q", s, p.contents[0])
		}
	}
}

// 選んで Enter を押すと、そのセッションを再開するコマンドを出力すること
func TestView_aiは再開コマンドを出す(t *testing.T) {
	tests := []struct {
		pick int
		want string
	}{
		{0, "cd '/work/app' && claude --resume s1\n"},
		{1, "cd '/work/app-wt' && codex resume c1\n"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			cli := setupCLI(t)
			cli.sessions = sampleSessions
			cli.sessionPick = tt.pick

			if err := cli.run("view", "--ai"); err != nil {
				t.Fatal(err)
			}

			if got := cli.out.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// 選ばずにやめたら、何も出力しないこと
func TestView_aiで選ばずにやめたら何も出さない(t *testing.T) {
	cli := setupCLI(t)
	cli.sessions = sampleSessions
	cli.sessionPick = -1

	if err := cli.run("view", "--ai"); err != nil {
		t.Fatal(err)
	}

	if got := cli.out.String(); got != "" {
		t.Errorf("出力された: %q", got)
	}
}

// view --ai . で、カレントと同じリポジトリ（別の worktree を含む）のセッションだけにすること
func TestView_aiのドットは同じリポジトリだけ(t *testing.T) {
	cli := setupCLI(t)
	cli.sessions = sampleSessions
	cli.sessionPick = -1
	cli.repos = map[string]string{cli.workDir: "/work/app", "/work/app-wt": "/work/app"}

	if err := cli.run("view", "--ai", "."); err != nil {
		t.Fatal(err)
	}

	p := cli.sessionPicker
	if p == nil || strings.Join(p.fullKeys, ",") != "/work/app,/work/app-wt" {
		t.Errorf("同じリポジトリのセッションだけになっていない: %+v", p)
	}
}

// 作業ディレクトリがもう存在しないセッションには印を付けること
func TestView_aiは消えた作業ディレクトリに印を付ける(t *testing.T) {
	cli := setupCLI(t)
	cli.sessions = sampleSessions
	cli.sessionPick = -1
	cli.missing = map[string]bool{"/work/app-wt": true}

	if err := cli.run("view", "--ai"); err != nil {
		t.Fatal(err)
	}

	p := cli.sessionPicker
	if strings.Contains(p.labels[0], "削除済み") || !strings.Contains(p.labels[1], "削除済み") {
		t.Errorf("印が正しくない: %q", p.labels)
	}
}

// 一覧は日時・種類・リポジトリの「親/名前」・ブランチの順に出し、切り替え先ではフルパスを出すこと
func TestView_aiは日時と種類のあとにリポジトリを出す(t *testing.T) {
	cli := setupCLI(t)
	updated := time.Date(2026, 11, 24, 15, 6, 0, 0, time.Local)
	cli.sessions = []aisession.Session{
		{Agent: "claude", ID: "s1", Cwd: "/Users/me/ghq/github.com/yyYank/anw-wt/cmd", Branch: "main", Updated: updated},
		{Agent: "codex", ID: "s2", Cwd: "/Users/me/tmp/scratch", Branch: "dev", Updated: updated},
		{Agent: "claude", ID: "s3", Cwd: "/Users/me/x/" + strings.Repeat("long", 10), Branch: "main", Updated: updated},
	}
	cli.sessionPick = -1
	// worktree のサブディレクトリでも、本体のリポジトリにまとまる
	cli.repos = map[string]string{"/Users/me/ghq/github.com/yyYank/anw-wt/cmd": "/Users/me/ghq/github.com/yyYank/anw"}

	if err := cli.run("view", "--ai"); err != nil {
		t.Fatal(err)
	}

	p := cli.sessionPicker
	for i, dir := range []string{"yyYank/anw", "tmp/scratch"} {
		s := cli.sessions[i]
		// 列は一覧の画面でそろえるので、タブで区切って渡す
		if want := "11-24 15:06\t" + s.Agent + "\t" + dir + "\t" + s.Branch + "\t"; !strings.HasPrefix(p.labels[i], want) {
			t.Errorf("labels[%d] = %q, want 先頭が %q", i, p.labels[i], want)
		}
		if p.keys[i] != dir || p.fullKeys[i] != s.Cwd {
			t.Errorf("絞り込みの文字列が %q, %q", p.keys[i], p.fullKeys[i])
		}
		if !strings.Contains(p.fullLabels[i], "\t"+s.Cwd+"\t") {
			t.Errorf("fullLabels[%d] = %q にフルパスがない", i, p.fullLabels[i])
		}
	}
	// 長い名前は切り詰める
	if strings.Contains(p.labels[2], strings.Repeat("long", 10)) {
		t.Errorf("長い名前が切り詰められていない: %q", p.labels[2])
	}
}

// セッションがなければ、一覧を出さずにエラーにすること
func TestView_aiのセッションがなければエラー(t *testing.T) {
	cli := setupCLI(t)

	if err := cli.run("view", "--ai"); err == nil {
		t.Error("エラーにならなかった")
	}
	if cli.sessionPicker != nil {
		t.Error("一覧が表示された")
	}
}

// --ai に . 以外の値を渡すとエラーにすること
func TestView_aiにドット以外はエラー(t *testing.T) {
	cli := setupCLI(t)
	cli.sessions = sampleSessions

	if err := cli.run("view", "--ai", "foo"); err == nil {
		t.Error("エラーにならなかった")
	}
	if cli.sessionPicker != nil {
		t.Error("一覧が表示された")
	}
}
