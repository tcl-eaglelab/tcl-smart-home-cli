package output

import (
	"testing"
)

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no ansi", "hello", "hello"},
		{"color code", "\033[31mred\033[0m", "red"},
		{"bold", "\033[1mbold\033[0m", "bold"},
		{"multiple codes", "\033[31m\033[1mredBold\033[0m", "redBold"},
		{"empty string", "", ""},
		{"nested codes", "\033[36m\033[1mCLI\033[0m", "CLI"},
		{"plain with brackets", "a[1]b", "a[1]b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAnsi(tt.input)
			if got != tt.want {
				t.Errorf("stripAnsi(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"ascii", "hello", 5},
		{"empty", "", 0},
		{"chinese", "你好", 4},
		{"mixed", "hi你好", 6},
		{"japanese", "テスト", 6},
		{"korean", "한글", 4},
		{"emoji basic", "abc", 3},
		{"single CJK char", "中", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runeWidth(tt.input)
			if got != tt.want {
				t.Errorf("runeWidth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsWide(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"ascii A", 'A', false},
		{"ascii space", ' ', false},
		{"CJK unified", '中', true},
		{"hiragana", 'あ', true},
		{"katakana", 'ア', true},
		{"hangul", '한', true},
		{"fullwidth exclamation", '！', true},
		{"latin digit", '0', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWide(tt.r)
			if got != tt.want {
				t.Errorf("isWide(%q) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestStatusBadge(t *testing.T) {
	tests := []struct {
		name   string
		online bool
	}{
		{"online", true},
		{"offline", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StatusBadge(tt.online)
			stripped := stripAnsi(got)
			if tt.online {
				if stripped != "● online" {
					t.Errorf("StatusBadge(true) stripped = %q, want %q", stripped, "● online")
				}
			} else {
				if stripped != "○ offline" {
					t.Errorf("StatusBadge(false) stripped = %q, want %q", stripped, "○ offline")
				}
			}
		})
	}
}
