package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/yyYank/anw/internal/aisession"
	"github.com/yyYank/anw/internal/envfile"
)

// テストで差し替えられるよう変数にしておく
var (
	collectSessions   = aisession.Collect
	repoOf            = aisession.RepoOf
	showSessionPicker = runSessionPicker
)

// firstPromptWidth は一覧に出す最初の発言の文字数の上限
const firstPromptWidth = 40

// dirWidth は一覧に出すリポジトリの「親/名前」の桁数の上限
const dirWidth = 30

// viewAI はエージェントのセッションを一覧し、選んだセッションを再開するコマンドを出力する。
// args が "." なら、カレントと同じリポジトリのセッションだけにする。
func viewAI(w io.Writer, args []string) error {
	if len(args) > 0 && args[0] != "." {
		return fmt.Errorf("--ai に指定できるのは . だけです: %q", args[0])
	}
	home, err := homeDir()
	if err != nil {
		return err
	}
	sessions, err := collectSessions(home)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		if sessions, err = sessionsInCurrentRepo(sessions); err != nil {
			return err
		}
	}
	if len(sessions) == 0 {
		return errors.New("セッションが見つかりません")
	}

	dirs := make([]string, len(sessions))
	fullDirs := make([]string, len(sessions))
	contents := make([]string, len(sessions))
	for i, s := range sessions {
		dirs[i], fullDirs[i] = sessionDirs(s)
		contents[i] = transcript(s)
	}
	labels := sessionLabels(sessions, dirs, dirWidth)
	fullLabels := sessionLabels(sessions, fullDirs, -1)
	i, ok, err := showSessionPicker(dirs, labels, fullDirs, fullLabels, contents)
	if err != nil || !ok {
		return err
	}
	_, err = fmt.Fprintln(w, resumeCommand(sessions[i]))
	return err
}

// sessionsInCurrentRepo は sessions のうち、カレントと同じリポジトリに属するものを返す。
func sessionsInCurrentRepo(sessions []aisession.Session) ([]aisession.Session, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	current, _ := repoOf(dir)
	repos := map[string]string{}
	var result []aisession.Session
	for _, s := range sessions {
		repo, ok := repos[s.Cwd]
		if !ok {
			repo, _ = repoOf(s.Cwd)
			repos[s.Cwd] = repo
		}
		if repo == current {
			result = append(result, s)
		}
	}
	return result, nil
}

// sessionDirs は一覧に出す作業ディレクトリを、リポジトリの「親/名前」とフルパスの2通りで返す。
// git の外や、作業ディレクトリがもう存在しなければ、作業ディレクトリの「親/名前」にする。存在しなければ印を付ける。
func sessionDirs(s aisession.Session) (short, full string) {
	repo, ok := repoOf(s.Cwd)
	short, full = parentAndName(repo), s.Cwd
	if !ok {
		short, full = short+" (削除済み)", full+" (削除済み)"
	}
	return short, full
}

// parentAndName は path の末尾2階層を返す。
func parentAndName(path string) string {
	return filepath.Join(filepath.Base(filepath.Dir(path)), filepath.Base(path))
}

// sessionLabels は一覧に出す行を、日時・種類・作業ディレクトリ dirs・ブランチ・最初の発言の順にタブで区切って返す。
// 列は一覧の画面が表示中の行に合わせてそろえる。limit が正なら、dirs をその桁数までに切り詰める。
func sessionLabels(sessions []aisession.Session, dirs []string, limit int) []string {
	labels := make([]string, len(sessions))
	for i, s := range sessions {
		dir := dirs[i]
		if limit > 0 {
			dir = ansi.Truncate(dir, limit, "…")
		}
		prompt := []rune(s.FirstPrompt)
		if len(prompt) > firstPromptWidth {
			prompt = append(prompt[:firstPromptWidth], '…')
		}
		labels[i] = strings.Join([]string{s.Updated.Local().Format("01-02 15:04"), s.Agent, dir, s.Branch, string(prompt)}, "\t")
	}
	return labels
}

// transcript はプレビューと中身の検索に使う、発言を並べた文字列。
func transcript(s aisession.Session) string {
	var b strings.Builder
	for _, m := range s.Messages {
		if m.Role == aisession.RoleTool {
			fmt.Fprintf(&b, "[tool] %s\n", m.Text)
			continue
		}
		fmt.Fprintf(&b, "%s: %s\n\n", m.Role, m.Text)
	}
	return b.String()
}

// resumeCommand はセッションを再開するコマンド。
func resumeCommand(s aisession.Session) string {
	cd := "cd " + shellQuote(s.Cwd) + " && "
	if s.Agent == "codex" {
		return cd + "codex resume " + s.ID
	}
	return cd + "claude --resume " + s.ID
}

// shellQuote は s をシェルの単一引用符で囲む。
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// runSessionPicker はセッションの一覧を表示し、選ばれたセッションの添字を返す。
// / は keys で絞り込み、p で fullKeys での絞り込みと fullLabels の表示に切り替えられる。
func runSessionPicker(keys, labels, fullKeys, fullLabels, contents []string) (int, bool, error) {
	p := envfile.NewPickerWithContents(keys, contents, true).WithLabels(labels).WithAltLabels(fullKeys, fullLabels).WithTitle("セッション")
	m, err := tea.NewProgram(p).Run()
	if err != nil {
		return 0, false, err
	}
	i, ok := m.(envfile.Picker).SelectedIndex()
	return i, ok, nil
}
