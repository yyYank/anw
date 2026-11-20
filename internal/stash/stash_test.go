package stash

import (
	"reflect"
	"testing"
	"time"
)

// git stash list --format=%gd%x00%H%x00%ct%x00%gs を真似た出力
const listOutput = "stash@{0}\x00aaaa\x001762000000\x00WIP on feature/x: 1234 作業中\n" +
	"stash@{1}\x00bbbb\x001761000000\x00On main: 実験\n"

// stash の番号・コミット・日時・メッセージ・元のブランチを読み、元のブランチがもうないものに印を付けること
func TestParse_stashを読む(t *testing.T) {
	got := Parse(listOutput, map[string]bool{"main": true})

	want := []Stash{
		{Ref: "stash@{0}", Commit: "aaaa", Time: time.Unix(1762000000, 0), Message: "WIP on feature/x: 1234 作業中", Branch: "feature/x", BranchGone: true},
		{Ref: "stash@{1}", Commit: "bbbb", Time: time.Unix(1761000000, 0), Message: "On main: 実験", Branch: "main"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v\nwant %+v", got, want)
	}
}

// 理由には元のブランチと何日前かを出し、ブランチがもうなければそう書くこと
func TestReason_ブランチと日数(t *testing.T) {
	now := time.Unix(1762000000, 0).Add(12 * 24 * time.Hour)
	s := Parse(listOutput, map[string]bool{"main": true})

	if got := s[0].Reason(now); got != "feature/x (ブランチなし), 12日前" {
		t.Errorf("got %q", got)
	}
	if got := s[1].Reason(now); got != "main, 23日前" {
		t.Errorf("got %q", got)
	}
}
