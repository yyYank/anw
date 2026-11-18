package port

import (
	"errors"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// Listener は LISTEN 中のポートと、それを使っているプロセスを表す。
type Listener struct {
	Port    int
	PID     int
	Name    string
	Command string
}

// Parse は lsof -F pcn の出力を Listener の一覧にする。ポート番号順に並べる。
func Parse(out string) []Listener {
	type key struct{ pid, port int }
	var (
		result []Listener
		seen   = map[key]bool{}
		pid    int
		name   string
	)
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		value := line[1:]
		switch line[0] {
		case 'p':
			pid, _ = strconv.Atoi(value)
		case 'c':
			name = value
		case 'n':
			i := strings.LastIndex(value, ":")
			if i < 0 {
				continue
			}
			port, err := strconv.Atoi(value[i+1:])
			if err != nil || seen[key{pid, port}] {
				continue
			}
			seen[key{pid, port}] = true
			result = append(result, Listener{Port: port, PID: pid, Name: name})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Port < result[j].Port })
	return result
}

// Find は ls から port を使っているものを返す。
func Find(ls []Listener, port int) []Listener {
	var result []Listener
	for _, l := range ls {
		if l.Port == port {
			result = append(result, l)
		}
	}
	return result
}

// List は LISTEN 中の TCP ポートを lsof で調べ、コマンドを ps で補って返す。
func List() ([]Listener, error) {
	out, err := exec.Command("lsof", "-nP", "-iTCP", "-sTCP:LISTEN", "-F", "pcn").Output()
	if err != nil {
		// 該当がないとき lsof は出力なしで終了コード 1 を返す
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || len(out) > 0 {
			return nil, err
		}
	}
	ls := Parse(string(out))
	for i := range ls {
		cmd, err := exec.Command("ps", "-o", "command=", "-p", strconv.Itoa(ls[i].PID)).Output()
		if err == nil {
			ls[i].Command = strings.TrimSpace(string(cmd))
		}
	}
	return ls, nil
}
