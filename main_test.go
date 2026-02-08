package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tbourrel/apitty/internal/json"
	"github.com/tbourrel/apitty/internal/model"
	"github.com/tbourrel/apitty/internal/parser"
	"github.com/tbourrel/apitty/internal/text"
	"github.com/tbourrel/apitty/internal/ui"
)

func TestInitialModel(t *testing.T) {
	m := model.InitialModel()

	if m.Focus != model.FocusMethod {
		t.Errorf("expected initial focus to be FocusMethod, got %v", m.Focus)
	}

	if m.MethodIdx != 0 {
		t.Errorf("expected initial method index to be 0 (GET), got %d", m.MethodIdx)
	}

	if m.URLInput.Value() != "" {
		t.Errorf("expected initial URL to be empty, got %s", m.URLInput.Value())
	}

	if m.Response != "" {
		t.Errorf("expected initial response to be empty")
	}

	if m.CurrentView != model.ViewBody {
		t.Errorf("expected initial response view to be ViewBody")
	}

	if len(m.RequestHeaders) != 0 {
		t.Errorf("expected no headers initially, got %d", len(m.RequestHeaders))
	}
}

func TestMethodCycling(t *testing.T) {
	m := model.InitialModel()
	m.Focus = model.FocusMethod

	tests := []struct {
		key      tea.KeyMsg
		expected int
	}{
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 1}, // GET -> POST
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 2}, // POST -> PUT
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 3}, // PUT -> PATCH
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 4}, // PATCH -> DELETE
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 0}, // DELETE -> GET (wrap)
		{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, 4}, // GET -> DELETE (reverse)
	}

	for i, tt := range tests {
		newModel, _ := ui.Update(m, tt.key)
		m = newModel
		if m.MethodIdx != tt.expected {
			t.Errorf("test %d: expected method index %d, got %d", i, tt.expected, m.MethodIdx)
		}
	}
}

func TestResponseViewToggle(t *testing.T) {
	m := model.InitialModel()
	m.Focus = model.FocusResponse
	m.Response = "test response"
	m.ResponseHeaders = "Content-Type: application/json"
	m.ViewportReady = true

	// Toggle from Body to Headers with "t"
	newModel, _ := ui.Update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = newModel
	if m.CurrentView != model.ViewHeaders {
		t.Errorf("expected response view to be ViewHeaders, got %v", m.CurrentView)
	}

	// Toggle back to Body
	newModel, _ = ui.Update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = newModel
	if m.CurrentView != model.ViewBody {
		t.Errorf("expected response view to be ViewBody, got %v", m.CurrentView)
	}
}

func TestFullscreenToggle(t *testing.T) {
	m := model.InitialModel()
	m.Focus = model.FocusResponse
	m.Response = "test response"

	if m.Fullscreen {
		t.Error("expected fullscreen to be disabled initially")
	}

	// Toggle fullscreen on
	newModel, _ := ui.Update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = newModel
	if !m.Fullscreen {
		t.Error("expected fullscreen to be enabled")
	}

	// Toggle fullscreen off
	newModel, _ = ui.Update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = newModel
	if m.Fullscreen {
		t.Error("expected fullscreen to be disabled")
	}
}

func TestFullscreenOnlyWhenResponseHasContent(t *testing.T) {
	m := model.InitialModel()
	m.Focus = model.FocusResponse
	m.Response = "" // No content

	// Try to toggle fullscreen
	newModel, _ := ui.Update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = newModel
	if m.Fullscreen {
		t.Error("expected fullscreen to remain disabled when no response content")
	}
}

func TestColorizeJSON(t *testing.T) {
	input := `{"key": "value", "number": 123}`
	result := json.ColorizeJSON(input)

	if result == "" {
		t.Error("expected colorized output")
	}
}

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, url, headers := parser.ParseCurlCommand(tt.curl)

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

func TestWrapText(t *testing.T) {
	result := text.WrapText("Hello World", 5)
	if result == "" {
		t.Error("expected wrapped text")
	}
}

func TestURLInputEditable(t *testing.T) {
	m := model.InitialModel()
	m.Focus = model.FocusURL
	m.URLInput.Focus()

	// Type a URL
	testURL := "https://api.example.com"
	for _, r := range testURL {
		m.URLInput, _ = m.URLInput.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if m.URLInput.Value() != testURL {
		t.Errorf("expected URL to be %s, got %s", testURL, m.URLInput.Value())
	}
}

func TestHelpToggle(t *testing.T) {
	m := model.InitialModel()

	if m.ShowHelp {
		t.Error("expected help to be hidden initially")
	}

	// Toggle help on
	newModel, _ := ui.Update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel
	if !m.ShowHelp {
		t.Error("expected help to be shown")
	}

	// Toggle help off
	newModel, _ = ui.Update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel
	if m.ShowHelp {
		t.Error("expected help to be hidden")
	}
}

func TestHeaderFormToggle(t *testing.T) {
	m := model.InitialModel()

	if m.ShowHeadersForm {
		t.Error("expected header form to be hidden initially")
	}

	// Toggle header form on
	newModel, _ := ui.Update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = newModel
	if !m.ShowHeadersForm {
		t.Error("expected header form to be shown")
	}

	// Close with Esc
	newModel, _ = ui.Update(m, tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel
	if m.ShowHeadersForm {
		t.Error("expected header form to be hidden after Esc")
	}
}

func TestCurlImportToggle(t *testing.T) {
	m := model.InitialModel()

	if m.ShowCurlImport {
		t.Error("expected curl import to be hidden initially")
	}

	// Toggle curl import on
	newModel, _ := ui.Update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = newModel
	if !m.ShowCurlImport {
		t.Error("expected curl import to be shown")
	}

	// Close with Esc
	newModel, _ = ui.Update(m, tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel
	if m.ShowCurlImport {
		t.Error("expected curl import to be hidden after Esc")
	}
}

func TestMethodsArray(t *testing.T) {
	expectedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	if len(model.Methods) != len(expectedMethods) {
		t.Errorf("expected %d methods, got %d", len(expectedMethods), len(model.Methods))
	}

	for i, method := range expectedMethods {
		if model.Methods[i] != method {
			t.Errorf("expected method at index %d to be %s, got %s", i, method, model.Methods[i])
		}
	}
}

func TestView(t *testing.T) {
	m := model.InitialModel()
	m.Width = 100
	m.Height = 40

	// Just test that View doesn't panic
	view := ui.View(m)
	if view == "" {
		t.Error("expected non-empty view")
	}
}
