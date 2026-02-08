package text

import (
	"strings"
	"testing"
)

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected int // expected number of lines
	}{
		{
			name:     "Short text no wrap",
			text:     "Hello",
			width:    20,
			expected: 1,
		},
		{
			name:     "Long text with wrap",
			text:     "This is a very long line that should be wrapped",
			width:    10,
			expected: 5, // Approximate, should be multiple lines
		},
		{
			name:     "Text with newlines",
			text:     "Line1\nLine2\nLine3",
			width:    50,
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.text, tt.width)
			lines := strings.Split(result, "\n")
			if len(lines) < tt.expected {
				t.Errorf("expected at least %d lines, got %d", tt.expected, len(lines))
			}
		})
	}
}

func TestVisibleLength(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Plain text",
			input:    "Hello World",
			expected: 11,
		},
		{
			name:     "Text with ANSI codes",
			input:    "\x1b[31mRed Text\x1b[0m",
			expected: 8, // Only "Red Text" counts
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VisibleLength(tt.input)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}
