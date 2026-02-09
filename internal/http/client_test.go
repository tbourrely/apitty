package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tbourrel/apitty/internal/model"
)

func TestSendRequestCmd_HeadersAreSent(t *testing.T) {
	// Create a test server that captures received headers
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	tests := []struct {
		name            string
		headers         []model.HeaderPair
		expectedHeaders map[string]string
	}{
		{
			name: "Single Authorization header",
			headers: []model.HeaderPair{
				{Key: "Authorization", Value: "Bearer test-token-123"},
			},
			expectedHeaders: map[string]string{
				"Authorization": "Bearer test-token-123",
			},
		},
		{
			name: "Multiple headers including Authorization",
			headers: []model.HeaderPair{
				{Key: "Authorization", Value: "Bearer secret-token"},
				{Key: "Content-Type", Value: "application/json"},
				{Key: "Accept", Value: "application/json"},
			},
			expectedHeaders: map[string]string{
				"Authorization": "Bearer secret-token",
				"Content-Type":  "application/json",
				"Accept":        "application/json",
			},
		},
		{
			name: "Authorization with different schemes",
			headers: []model.HeaderPair{
				{Key: "Authorization", Value: "Basic dXNlcjpwYXNz"},
			},
			expectedHeaders: map[string]string{
				"Authorization": "Basic dXNlcjpwYXNz",
			},
		},
		{
			name: "Custom headers",
			headers: []model.HeaderPair{
				{Key: "X-API-Key", Value: "my-api-key-123"},
				{Key: "X-Custom-Header", Value: "custom-value"},
			},
			expectedHeaders: map[string]string{
				"X-API-Key":       "my-api-key-123",
				"X-Custom-Header": "custom-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset received headers
			receivedHeaders = nil

			// Execute the command function directly
			cmd := SendRequestCmd("GET", server.URL, tt.headers, "")
			result := cmd()

			// Check for errors
			if responseMsg, ok := result.(model.ResponseMsg); ok {
				if responseMsg.Err != nil {
					t.Fatalf("unexpected error: %v", responseMsg.Err)
				}

				// Verify HTTP status
				if responseMsg.Status != "200 OK" {
					t.Errorf("expected status '200 OK', got '%s'", responseMsg.Status)
				}
			} else {
				t.Fatalf("unexpected result type: %T", result)
			}

			// Verify all expected headers were received
			for key, expectedValue := range tt.expectedHeaders {
				receivedValue := receivedHeaders.Get(key)
				if receivedValue != expectedValue {
					t.Errorf("header %s: expected '%s', got '%s'", key, expectedValue, receivedValue)
				}
			}
		})
	}
}

func TestSendRequestCmd_EmptyHeaderKeyNotSent(t *testing.T) {
	// Create a test server
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Headers with empty keys should not be sent
	headers := []model.HeaderPair{
		{Key: "", Value: "should-not-be-sent"},
		{Key: "Valid-Header", Value: "valid-value"},
	}

	cmd := SendRequestCmd("GET", server.URL, headers, "")
	result := cmd()

	if responseMsg, ok := result.(model.ResponseMsg); ok {
		if responseMsg.Err != nil {
			t.Fatalf("unexpected error: %v", responseMsg.Err)
		}
	}

	// Verify only the valid header was sent
	if receivedHeaders.Get("Valid-Header") != "valid-value" {
		t.Errorf("expected Valid-Header to be sent")
	}

	// Verify the header with empty key was not sent (it shouldn't appear with empty name)
	// Note: Go HTTP client adds User-Agent and Accept-Encoding by default
	foundEmptyValue := false
	for _, values := range receivedHeaders {
		for _, value := range values {
			if value == "should-not-be-sent" {
				foundEmptyValue = true
				break
			}
		}
	}
	if foundEmptyValue {
		t.Error("header with empty key was incorrectly sent")
	}
}

func TestSendRequestCmd_AuthorizationWith401Response(t *testing.T) {
	// Simulate a server that returns 401 even with Authorization header
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		// Intentionally return 401 to test that we still capture the header
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
	}))
	defer server.Close()

	headers := []model.HeaderPair{
		{Key: "Authorization", Value: "Bearer token-123"},
	}

	cmd := SendRequestCmd("GET", server.URL, headers, "")
	result := cmd()

	if responseMsg, ok := result.(model.ResponseMsg); ok {
		if responseMsg.Err != nil {
			t.Fatalf("unexpected error: %v", responseMsg.Err)
		}

		// Verify we got 401 status
		if responseMsg.Status != "401 Unauthorized" {
			t.Errorf("expected status '401 Unauthorized', got '%s'", responseMsg.Status)
		}
	}

	// Most importantly: verify the Authorization header WAS sent
	if receivedAuth != "Bearer token-123" {
		t.Errorf("Authorization header not received correctly. Expected 'Bearer token-123', got '%s'", receivedAuth)
	}

	if receivedAuth == "" {
		t.Error("Authorization header was NOT sent to the server!")
	}
}
