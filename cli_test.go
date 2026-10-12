package main

import "testing"

func TestPreview(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"hello", 10, "hello"},                     // short: unchanged
		{"a\nb\t\tc", 10, "a b c"},                 // whitespace runs → one space
		{"  padded  ", 10, "padded"},               // trimmed
		{"hello world", 5, "hell…"},                // cut: max-1 chars + …
		{"héllo", 5, "héllo"},                      // 5 runes (6 bytes) fits
		{"日本語テキスト", 4, "日本語…"},                     // cut by runes, not bytes
		{"line one\nline two", 12, "line one li…"}, // flatten, then cut
	}
	for _, c := range cases {
		if got := preview(c.in, c.max); got != c.want {
			t.Errorf("preview(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
		}
	}
}
