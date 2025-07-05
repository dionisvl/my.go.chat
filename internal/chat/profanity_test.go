package chat

import (
	"testing"
)

func TestLoadProfanities(t *testing.T) {
	profanities := "bad,word,test"
	LoadProfanities(profanities)

	if profanityDetector == nil {
		t.Error("profanityDetector should not be nil after LoadProfanities")
	}
}

func TestCensor(t *testing.T) {
	LoadProfanities("bad,word")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "censor russian profanity",
			input:    "хуй test",
			expected: "*** test",
		},
		{
			name:     "censor custom profanity",
			input:    "this is bad",
			expected: "this is ***",
		},
		{
			name:     "clean text unchanged",
			input:    "hello world",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Censor(tt.input)
			if result != tt.expected {
				t.Errorf("Censor(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
