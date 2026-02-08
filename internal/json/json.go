package json

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TryPrettyJSON tries to pretty-print JSON, falls back to string if not JSON
func TryPrettyJSON(data []byte) string {
	trim := bytes.TrimSpace(data)
	if len(trim) == 0 {
		return ""
	}
	if trim[0] == '{' || trim[0] == '[' {
		var out bytes.Buffer
		err := indent(&out, trim)
		if err == nil {
			return ColorizeJSON(out.String())
		}
	}
	return string(data)
}

// indent indents JSON for pretty printing
func indent(out *bytes.Buffer, data []byte) error {
	type jsonRaw = map[string]interface{}
	type jsonArr = []interface{}
	// Try object
	if data[0] == '{' {
		var obj jsonRaw
		if err := json.Unmarshal(data, &obj); err == nil {
			return marshalIndent(obj, out)
		}
	}
	// Try array
	if data[0] == '[' {
		var arr jsonArr
		if err := json.Unmarshal(data, &arr); err == nil {
			return marshalIndent(arr, out)
		}
	}
	return nil
}

func marshalIndent(v interface{}, out *bytes.Buffer) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	out.Write(b)
	return nil
}

// ColorizeJSON adds ANSI color codes to JSON for syntax highlighting
func ColorizeJSON(jsonStr string) string {
	var result strings.Builder
	var currentString strings.Builder
	inString := false
	inEscape := false
	isKey := false

	keyColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	stringColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	numberColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	boolNullColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	bracketColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF"))

	i := 0
	for i < len(jsonStr) {
		ch := jsonStr[i]

		if inString {
			if inEscape {
				currentString.WriteByte(ch)
				inEscape = false
			} else if ch == '\\' {
				currentString.WriteByte(ch)
				inEscape = true
			} else if ch == '"' {
				currentString.WriteByte('"')
				// Render the complete string with appropriate color
				if isKey {
					result.WriteString(keyColor.Render(currentString.String()))
				} else {
					result.WriteString(stringColor.Render(currentString.String()))
				}
				currentString.Reset()
				inString = false
			} else {
				currentString.WriteByte(ch)
			}
			i++
			continue
		}

		switch ch {
		case '"':
			// Detect if this is a key or a value
			isKey = false
			j := i + 1
			for j < len(jsonStr) && jsonStr[j] != '"' {
				if jsonStr[j] == '\\' {
					j++
				}
				j++
			}
			if j < len(jsonStr) {
				j++ // skip closing quote
				// skip whitespace
				for j < len(jsonStr) && (jsonStr[j] == ' ' || jsonStr[j] == '\t' || jsonStr[j] == '\n' || jsonStr[j] == '\r') {
					j++
				}
				if j < len(jsonStr) && jsonStr[j] == ':' {
					isKey = true
				}
			}

			currentString.WriteByte('"')
			inString = true

		case '{', '}', '[', ']':
			result.WriteString(bracketColor.Render(string(ch)))

		case 't', 'f', 'n':
			// Check for true, false, null
			if i+4 <= len(jsonStr) && jsonStr[i:i+4] == "true" {
				result.WriteString(boolNullColor.Render("true"))
				i += 3
			} else if i+5 <= len(jsonStr) && jsonStr[i:i+5] == "false" {
				result.WriteString(boolNullColor.Render("false"))
				i += 4
			} else if i+4 <= len(jsonStr) && jsonStr[i:i+4] == "null" {
				result.WriteString(boolNullColor.Render("null"))
				i += 3
			} else {
				result.WriteByte(ch)
			}

		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// Number
			start := i
			if ch == '-' {
				i++
			}
			for i < len(jsonStr) && (jsonStr[i] >= '0' && jsonStr[i] <= '9' || jsonStr[i] == '.' || jsonStr[i] == 'e' || jsonStr[i] == 'E' || jsonStr[i] == '+' || jsonStr[i] == '-') {
				i++
			}
			result.WriteString(numberColor.Render(jsonStr[start:i]))
			i--

		default:
			result.WriteByte(ch)
		}
		i++
	}

	return result.String()
}
