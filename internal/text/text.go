package text

import (
	"strings"
)

// WrapText wraps text to fit within the given width
func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		// Handle ANSI color codes - strip them for length calculation
		visibleLen := VisibleLength(line)

		if visibleLen <= width {
			result.WriteString(line)
			result.WriteByte('\n')
			continue
		}

		// Wrap long lines
		currentPos := 0
		for currentPos < len(line) {
			// Find how many characters fit in width
			chunkEnd := findChunkEnd(line, currentPos, width)

			if chunkEnd <= currentPos {
				break
			}

			// Extract chunk
			chunk := line[currentPos:chunkEnd]

			// Try to break at a good position (space, comma, etc.)
			if chunkEnd < len(line) {
				// Look back for a good break point
				for i := len(chunk) - 1; i >= max(0, len(chunk)-15); i-- {
					if chunk[i] == ' ' || chunk[i] == ',' || chunk[i] == ':' {
						chunk = chunk[:i+1]
						chunkEnd = currentPos + len(chunk)
						break
					}
				}
			}

			result.WriteString(chunk)
			result.WriteByte('\n')

			// Skip leading spaces on next line
			currentPos = chunkEnd
			for currentPos < len(line) && line[currentPos] == ' ' {
				currentPos++
			}
		}
	}

	return result.String()
}

// VisibleLength returns the length of string without ANSI codes
func VisibleLength(s string) int {
	length := 0
	inAnsi := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inAnsi = true
			continue
		}

		if inAnsi {
			if s[i] == 'm' {
				inAnsi = false
			}
			continue
		}

		length++
	}

	return length
}

// findChunkEnd finds where to end a chunk based on visible width
func findChunkEnd(s string, start, width int) int {
	visibleCount := 0
	pos := start
	inAnsi := false

	for pos < len(s) {
		if s[pos] == '\x1b' {
			inAnsi = true
			pos++
			continue
		}

		if inAnsi {
			if s[pos] == 'm' {
				inAnsi = false
			}
			pos++
			continue
		}

		if visibleCount >= width {
			return pos
		}

		visibleCount++
		pos++
	}

	return pos
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
