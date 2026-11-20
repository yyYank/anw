package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

	keys := make([]string, len(sessions))
	labels := make([]string, len(sessions))
	contents := make([]string, len(sessions))
	for i, s := range sessions {
		keys[i] = s.Cwd
		labels[i] = sessionLabel(s)
		contents[i] = transcript(s)
	}
	i, ok, err := showSessionPicker(keys, labels, contents)
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

// sessionLabel は一覧に出す1行。作業ディレクトリがもう存在しなければ印を付ける。
func sessionLabel(s aisession.Session) string {
	cwd := s.Cwd
	if _, ok := repoOf(s.Cwd); !ok {
		cwd += " (削除済み)"
	}
	prompt := []rune(s.FirstPrompt)
	if len(prompt) > firstPromptWidth {
		prompt = append(prompt[:firstPromptWidth], '…')
	}
	return fmt.Sprintf("%s  %-6s  %s  %s  %s", s.Updated.Local().Format("2006-01-02 15:04"), s.Agent, cwd, s.Branch, string(prompt))
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
func runSessionPicker(keys, labels, contents []string) (int, bool, error) {
	p := envfile.NewPickerWithContents(keys, contents, true).WithLabels(labels).WithTitle("セッション")
	m, err := tea.NewProgram(p).Run()
	if err != nil {
		return 0, false, err
	}
	i, ok := m.(envfile.Picker).SelectedIndex()
	return i, ok, nil
}
