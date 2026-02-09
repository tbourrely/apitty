package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpClient "github.com/tbourrel/apitty/internal/http"
	"github.com/tbourrel/apitty/internal/model"
	"github.com/tbourrel/apitty/internal/parser"
)

// TestEndToEnd_CurlImportToHTTPRequest tests the complete flow:
// 1. User pastes curl command
// 2. Parser extracts headers
// 3. Headers are sent in HTTP request
// 4. Server receives correct headers
func TestEndToEnd_CurlImportToHTTPRequest(t *testing.T) {
	tests := []struct {
		name             string
		curlCommand      string
		expectAuthHeader string
	}{
		{
			name:             "GitHub API style",
			curlCommand:      `curl https://api.github.com/user -H "Authorization: Bearer ghp_xxxxxxxxxxxxxxxxxxxx"`,
			expectAuthHeader: "Bearer ghp_xxxxxxxxxxxxxxxxxxxx",
		},
		{
			name:             "JWT token",
			curlCommand:      `curl https://api.example.com/data -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"`,
			expectAuthHeader: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
		},
		{
			name:             "Basic auth",
			curlCommand:      `curl https://api.example.com/data -H "Authorization: Basic dXNlcjpwYXNzd29yZA=="`,
			expectAuthHeader: "Basic dXNlcjpwYXNzd29yZA==",
		},
		{
			name:             "API Key",
			curlCommand:      `curl https://api.example.com/data -H "X-API-Key: secret-key-123"`,
			expectAuthHeader: "secret-key-123",
		},
		{
			name:             "Multiple headers with auth",
			curlCommand:      `curl -X POST https://api.example.com/data -H "Content-Type: application/json" -H "Authorization: Bearer token123" -H "Accept: application/json"`,
			expectAuthHeader: "Bearer token123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that captures received headers
			var receivedAuth string
			var receivedAPIKey string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedAuth = r.Header.Get("Authorization")
				receivedAPIKey = r.Header.Get("X-API-Key")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success": true}`))
			}))
			defer server.Close()

			// Step 1: Parse curl command (simulating user paste)
			method, _, headers := parser.ParseCurlCommand(tt.curlCommand)

			// Verify headers were parsed
			if len(headers) == 0 {
				t.Fatal("no headers were parsed from curl command")
			}

			// Step 2: Simulate sending request with parsed headers (like pressing Ctrl+S)
			cmd := httpClient.SendRequestCmd(method, server.URL, headers, "")
			result := cmd()

			// Step 3: Verify request succeeded
			if responseMsg, ok := result.(model.ResponseMsg); ok {
				if responseMsg.Err != nil {
					t.Fatalf("request failed: %v", responseMsg.Err)
				}

				if responseMsg.Status != "200 OK" {
					t.Errorf("expected status '200 OK', got '%s'", responseMsg.Status)
				}
			} else {
				t.Fatalf("unexpected result type: %T", result)
			}

			// Step 4: Verify server received the expected auth header
			receivedValue := receivedAuth
			if receivedAuth == "" && receivedAPIKey != "" {
				receivedValue = receivedAPIKey
			}

			if receivedValue != tt.expectAuthHeader {
				t.Errorf("server did not receive correct auth header")
				t.Errorf("  Expected: '%s'", tt.expectAuthHeader)
				t.Errorf("  Received: '%s'", receivedValue)
				t.Errorf("\nParsed headers:")
				for i, h := range headers {
					t.Errorf("  [%d] %s: %s", i, h.Key, h.Value)
				}
			}
		})
	}
}

// TestEndToEnd_CurlImport401Response simulates the exact scenario:
// User imports curl with Authorization, but still gets 401
func TestEndToEnd_CurlImport401Response(t *testing.T) {
	// Create a server that always returns 401 but logs what it received
	var receivedAuth string
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized", "message": "Invalid token"}`))
	}))
	defer server.Close()

	curlCommand := `curl https://api.example.com/protected -H "Authorization: Bearer my-secret-token"`

	// Parse curl
	method, _, headers := parser.ParseCurlCommand(curlCommand)

	// Send request
	cmd := httpClient.SendRequestCmd(method, server.URL, headers, "")
	result := cmd()

	// Check response
	responseMsg, ok := result.(model.ResponseMsg)
	if !ok {
		t.Fatalf("unexpected result type: %T", result)
	}

	// Should NOT have an error (401 is a valid HTTP response, not an error)
	if responseMsg.Err != nil {
		t.Fatalf("unexpected error: %v", responseMsg.Err)
	}

	// Should get 401 status
	if responseMsg.Status != "401 Unauthorized" {
		t.Errorf("expected status '401 Unauthorized', got '%s'", responseMsg.Status)
	}

	// THE KEY TEST: Even though we got 401, verify the Authorization header WAS sent
	if receivedAuth != "Bearer my-secret-token" {
		t.Errorf("Authorization header was not sent correctly!")
		t.Errorf("  Expected: 'Bearer my-secret-token'")
		t.Errorf("  Received: '%s'", receivedAuth)
		t.Errorf("\nThis means the 401 was due to invalid/expired token, NOT missing header")
	} else {
		t.Logf("✓ Authorization header WAS sent correctly")
		t.Logf("✓ The 401 response was due to the token being invalid/expired, not a missing header")
	}

	if requestCount != 1 {
		t.Errorf("expected 1 request, got %d", requestCount)
	}
}
