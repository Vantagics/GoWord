package document

import "testing"

func TestCJKCountingForms(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{1, "一"},
		{10, "十"},
		{11, "十一"},
		{20, "二十"},
		{21, "二十一"},
		{100, "一百"},
		{101, "一百零一"},
		{110, "一百一十"},
		{111, "一百一十一"},
		{1000, "一千"},
		{1001, "一千零一"},
		{1010, "一千零一十"},
		{1100, "一千一百"},
		{2000, "二千"},
	}
	for _, tc := range cases {
		if got := cjkCounting(tc.n); got != tc.want {
			t.Errorf("%d = %s, want %s", tc.n, got, tc.want)
		}
	}
}
