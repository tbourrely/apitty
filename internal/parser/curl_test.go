package parser

import (
	"testing"

	"github.com/tbourrel/apitty/internal/model"
)

func TestParseCurlCommand(t *testing.T) {
	tests := []struct {
		name            string
		curl            string
		expectedMethod  string
		expectedURL     string
		expectedHeaders int
	}{
		{
			name:            "Simple GET",
			curl:            "curl https://api.example.com",
			expectedMethod:  "GET",
			expectedURL:     "https://api.example.com",
			expectedHeaders: 0,
		},
		{
			name:            "POST with headers",
			curl:            `curl -X POST https://api.example.com -H "Content-Type: application/json" -H "Authorization: Bearer token"`,
			expectedMethod:  "POST",
			expectedURL:     "https://api.example.com",
			expectedHeaders: 2,
		},
		{
			name:            "GET with single header",
			curl:            `curl https://api.example.com -H "Accept: application/json"`,
			expectedMethod:  "GET",
			expectedURL:     "https://api.example.com",
			expectedHeaders: 1,
		},
		{
			name:            "PUT request",
			curl:            "curl -X PUT https://api.example.com/resource",
			expectedMethod:  "PUT",
			expectedURL:     "https://api.example.com/resource",
			expectedHeaders: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, url, headers := ParseCurlCommand(tt.curl)

			if method != tt.expectedMethod {
				t.Errorf("expected method %s, got %s", tt.expectedMethod, method)
			}

			if url != tt.expectedURL {
				t.Errorf("expected URL %s, got %s", tt.expectedURL, url)
			}

			if len(headers) != tt.expectedHeaders {
				t.Errorf("expected %d headers, got %d", tt.expectedHeaders, len(headers))
			}
		})
	}
}

func TestParseCurlCommandHeaders(t *testing.T) {
	method, url, headers := ParseCurlCommand(`curl https://example.com -H "Content-Type: application/json"`)

	if method != "GET" {
		t.Errorf("expected GET, got %s", method)
	}

	if url != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", url)
	}

	if len(headers) != 1 {
		t.Fatalf("expected 1 header, got %d", len(headers))
	}

	if headers[0].Key != "Content-Type" {
		t.Errorf("expected header key 'Content-Type', got %s", headers[0].Key)
	}

	if headers[0].Value != "application/json" {
		t.Errorf("expected header value 'application/json', got %s", headers[0].Value)
	}
}

func TestParseCurlCommand_AuthorizationHeaders(t *testing.T) {
	tests := []struct {
		name            string
		curl            string
		expectedKey     string
		expectedValue   string
		expectedHeaders int
	}{
		{
			name:            "Bearer token with double quotes",
			curl:            `curl https://api.example.com -H "Authorization: Bearer abc123"`,
			expectedKey:     "Authorization",
			expectedValue:   "Bearer abc123",
			expectedHeaders: 1,
		},
		{
			name:            "Bearer token with single quotes",
			curl:            `curl https://api.example.com -H 'Authorization: Bearer abc123'`,
			expectedKey:     "Authorization",
			expectedValue:   "Bearer abc123",
			expectedHeaders: 1,
		},
		{
			name:            "Bearer token with long value",
			curl:            `curl https://api.example.com -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"`,
			expectedKey:     "Authorization",
			expectedValue:   "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			expectedHeaders: 1,
		},
		{
			name:            "Basic auth",
			curl:            `curl https://api.example.com -H "Authorization: Basic dXNlcjpwYXNzd29yZA=="`,
			expectedKey:     "Authorization",
			expectedValue:   "Basic dXNlcjpwYXNzd29yZA==",
			expectedHeaders: 1,
		},
		{
			name:            "Token scheme",
			curl:            `curl https://api.example.com -H "Authorization: Token abcdef123456"`,
			expectedKey:     "Authorization",
			expectedValue:   "Token abcdef123456",
			expectedHeaders: 1,
		},
		{
			name:            "Authorization with extra spaces",
			curl:            `curl https://api.example.com -H "Authorization:  Bearer  abc123  "`,
			expectedKey:     "Authorization",
			expectedValue:   "Bearer  abc123", // TrimSpace only trims leading/trailing
			expectedHeaders: 1,
		},
		{
			name:            "X-API-Key header",
			curl:            `curl https://api.example.com -H "X-API-Key: secret-key-123"`,
			expectedKey:     "X-API-Key",
			expectedValue:   "secret-key-123",
			expectedHeaders: 1,
		},
		{
			name:            "Multiple headers including auth",
			curl:            `curl https://api.example.com -H "Content-Type: application/json" -H "Authorization: Bearer token123"`,
			expectedKey:     "Authorization",
			expectedValue:   "Bearer token123",
			expectedHeaders: 2,
		},
		{
			name:            "Authorization before URL",
			curl:            `curl -H "Authorization: Bearer token123" https://api.example.com`,
			expectedKey:     "Authorization",
			expectedValue:   "Bearer token123",
			expectedHeaders: 1,
		},
		{
			name:            "Authorization with --header instead of -H",
			curl:            `curl https://api.example.com --header "Authorization: Bearer token123"`,
			expectedKey:     "Authorization",
			expectedValue:   "Bearer token123",
			expectedHeaders: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, url, headers := ParseCurlCommand(tt.curl)

			if method != "GET" {
				t.Errorf("expected method GET, got %s", method)
			}

			if url != "https://api.example.com" {
				t.Errorf("expected URL https://api.example.com, got %s", url)
			}

			if len(headers) != tt.expectedHeaders {
				t.Fatalf("expected %d headers, got %d", tt.expectedHeaders, len(headers))
			}

			// Find the expected header
			found := false
			for _, h := range headers {
				if h.Key == tt.expectedKey {
					found = true
					if h.Value != tt.expectedValue {
						t.Errorf("header %s: expected value '%s', got '%s'", tt.expectedKey, tt.expectedValue, h.Value)
					}
					break
				}
			}

			if !found {
				t.Errorf("expected header '%s' not found in parsed headers", tt.expectedKey)
				for i, h := range headers {
					t.Logf("  [%d] %s: %s", i, h.Key, h.Value)
				}
			}
		})
	}
}

func TestParseCurlCommand_ProblematicFormats(t *testing.T) {
	tests := []struct {
		name        string
		curl        string
		expectError bool
		checkHeader func(t *testing.T, headers []model.HeaderPair)
	}{
		{
			name: "Header with colon in value",
			curl: `curl https://example.com -H "Authorization: Bearer token:with:colons"`,
			checkHeader: func(t *testing.T, headers []model.HeaderPair) {
				if len(headers) != 1 {
					t.Fatalf("expected 1 header, got %d", len(headers))
				}
				if headers[0].Key != "Authorization" {
					t.Errorf("expected key 'Authorization', got '%s'", headers[0].Key)
				}
				if headers[0].Value != "Bearer token:with:colons" {
					t.Errorf("expected value 'Bearer token:with:colons', got '%s'", headers[0].Value)
				}
			},
		},
		{
			name: "Header without colon (malformed)",
			curl: `curl https://example.com -H "Authorization Bearer token"`,
			checkHeader: func(t *testing.T, headers []model.HeaderPair) {
				// Should be ignored (no colon separator)
				if len(headers) != 0 {
					t.Errorf("malformed header should be ignored, but got %d headers", len(headers))
				}
			},
		},
		{
			name: "Empty header value",
			curl: `curl https://example.com -H "Authorization: "`,
			checkHeader: func(t *testing.T, headers []model.HeaderPair) {
				if len(headers) != 1 {
					t.Fatalf("expected 1 header, got %d", len(headers))
				}
				if headers[0].Key != "Authorization" {
					t.Errorf("expected key 'Authorization', got '%s'", headers[0].Key)
				}
				if headers[0].Value != "" {
					t.Errorf("expected empty value, got '%s'", headers[0].Value)
				}
			},
		},
		{
			name: "Escaped quotes in header value",
			curl: `curl https://example.com -H "Authorization: Bearer \"token\""`,
			checkHeader: func(t *testing.T, headers []model.HeaderPair) {
				if len(headers) != 1 {
					t.Fatalf("expected 1 header, got %d", len(headers))
				}
				// Escaped quotes should be preserved
				t.Logf("Parsed value: '%s'", headers[0].Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, headers := ParseCurlCommand(tt.curl)
			if tt.checkHeader != nil {
				tt.checkHeader(t, headers)
			}
		})
	}
}
