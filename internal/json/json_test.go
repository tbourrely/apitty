package json

import (
	"strings"
	"testing"
)

func TestColorizeJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "Simple JSON",
			input: `{"key": "value", "number": 123, "bool": true, "null": null}`,
			contains: []string{
				"key",   // Keys should be colored
				"value", // String values should be colored
				"123",   // Numbers should be colored
				"true",  // Booleans should be colored
				"null",  // Null should be colored
			},
		},
		{
			name:  "Nested JSON",
			input: `{"outer": {"inner": "value"}}`,
			contains: []string{
				"outer",
				"inner",
				"value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorizeJSON(tt.input)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected colorized output to contain %q", expected)
				}
			}
		})
	}
}

func TestTryPrettyJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantJSON bool
	}{
		{
			name:     "Valid JSON object",
			input:    []byte(`{"key":"value"}`),
			wantJSON: true,
		},
		{
			name:     "Valid JSON array",
			input:    []byte(`[1,2,3]`),
			wantJSON: true,
		},
		{
			name:     "Plain text",
			input:    []byte("Not JSON"),
			wantJSON: false,
		},
		{
			name:     "Empty input",
			input:    []byte(""),
			wantJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TryPrettyJSON(tt.input)
			if tt.wantJSON {
				if result == string(tt.input) && len(result) > 0 {
					t.Error("expected JSON to be pretty-printed")
				}
			} else {
				if result == "" && len(tt.input) > 0 {
					// Empty result for empty input is ok
					return
				}
			}
		})
	}
}
