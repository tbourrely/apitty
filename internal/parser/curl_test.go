package parser

import (
	"testing"
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
