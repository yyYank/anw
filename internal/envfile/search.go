package envfile

import (
	"strings"
	"unicode"
)

// fuzzyMatch は query の文字が、この順で s に含まれるかを調べる。大文字小文字は区別しない。
// 一致すれば、一致した s の文字の位置 (rune 単位) を返す。
func fuzzyMatch(s, query string) ([]int, bool) {
	q := []rune(strings.ToLower(query))
	if len(q) == 0 {
		return nil, true
	}
	var pos []int
	for i, r := range []rune(s) {
		if unicode.ToLower(r) == q[len(pos)] {
			pos = append(pos, i)
			if len(pos) == len(q) {
				return pos, true
			}
		}
	}
	return nil, false
}

// containsFold は s が query を含むかを調べる。大文字小文字は区別しない。
func containsFold(s, query string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(query))
}
