package utils

import (
	"testing"
)

func TestGetRandomColor(t *testing.T) {
	color := GetRandomColor()

	if len(color) != 7 {
		t.Errorf("Expected color length 7, got %d", len(color))
	}

	if color[0] != '#' {
		t.Errorf("Expected color to start with #, got %s", string(color[0]))
	}

	for i := 1; i < len(color); i++ {
		char := color[i]
		isValid := (char >= '0' && char <= '9') || (char >= 'A' && char <= 'F')
		if !isValid {
			t.Errorf("Invalid hex character at position %d: %c", i, char)
		}
	}
}

func TestGetRandomColorUniqueness(t *testing.T) {
	colors := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		color := GetRandomColor()
		colors[color] = true
	}

	uniqueColors := len(colors)
	if uniqueColors < iterations/2 {
		t.Errorf("Expected at least %d unique colors, got %d", iterations/2, uniqueColors)
	}
}
