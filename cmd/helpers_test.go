package cmd

import (
	"testing"
)

func TestNonEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"non-empty string", "hello", "hello"},
		{"empty string", "", "-"},
		{"whitespace", " ", " "},
		{"chinese", "你好", "你好"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nonEmpty(tt.input)
			if got != tt.want {
				t.Errorf("nonEmpty(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
