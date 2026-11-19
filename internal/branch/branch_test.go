package branch

import (
	"reflect"
	"testing"
)

// for-each-ref の出力と --merged の出力から、削除できるブランチを集められること
func TestCollect_マージ済みとリモート削除済みを集める(t *testing.T) {
	refs := "*\x00feature/current\x00\n" +
		" \x00main\x00\n" +
		" \x00feature/merged\x00\n" +
		" \x00feature/gone\x00[gone]\n" +
		" \x00feature/both\x00[gone]\n" +
		" \x00feature/wip\x00[ahead 1]\n"
	merged := "main\nfeature/merged\nfeature/both\nfeature/current\n"

	got := Collect(refs, merged, "main")

	want := []Branch{
		{Name: "feature/both", Merged: true, Gone: true},
		{Name: "feature/gone", Gone: true},
		{Name: "feature/merged", Merged: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// 削除できる理由を、表示用の文字列にできること
func TestReason_理由を表示する(t *testing.T) {
	cases := map[Branch]string{
		{Merged: true}:             "マージ済み",
		{Gone: true}:               "リモート削除済み",
		{Merged: true, Gone: true}: "マージ済み, リモート削除済み",
	}
	for b, want := range cases {
		if got := b.Reason(); got != want {
			t.Errorf("%+v: got %q, want %q", b, got, want)
		}
	}
}
