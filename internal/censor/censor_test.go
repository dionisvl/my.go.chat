package censor

import "testing"

func TestClean(t *testing.T) {
	c := New([]string{"bad", "word"})

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"custom profanity masked", "this is bad", "this is ***"},
		{"clean text unchanged", "hello world", "hello world"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.Clean(tt.input); got != tt.want {
				t.Errorf("Clean(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewWithEmptyDictionary(t *testing.T) {
	c := New(nil)
	if got := c.Clean("hello"); got != "hello" {
		t.Errorf("Clean(%q) = %q, want unchanged", "hello", got)
	}
}
