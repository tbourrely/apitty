package model

import (
	"testing"
)

func TestInitialModel(t *testing.T) {
	m := InitialModel()

	if m.Focus != FocusMethod {
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

	if m.CurrentView != ViewBody {
		t.Errorf("expected initial response view to be ViewBody")
	}

	if len(m.RequestHeaders) != 0 {
		t.Errorf("expected no headers initially, got %d", len(m.RequestHeaders))
	}
}

func TestMethodsArray(t *testing.T) {
	expectedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	if len(Methods) != len(expectedMethods) {
		t.Errorf("expected %d methods, got %d", len(expectedMethods), len(Methods))
	}

	for i, method := range expectedMethods {
		if Methods[i] != method {
			t.Errorf("expected method at index %d to be %s, got %s", i, method, Methods[i])
		}
	}
}

func TestTextInputNotNil(t *testing.T) {
	m := InitialModel()

	if m.URLInput.Placeholder == "" {
		t.Error("URL input should have placeholder")
	}

	m.ShowHeadersForm = true
	m.HeaderFormMode = HeaderModeEdit

	// These should not panic
	_ = m.HeaderKeyInput.Value()
	_ = m.HeaderValInput.Value()
}

func TestViewportNotNil(t *testing.T) {
	m := InitialModel()
	m.Width = 100
	m.Height = 40

	// Viewport should be initialized
	if m.Viewport.Width == 0 && m.Viewport.Height == 0 {
		// Width and height can be 0 initially, but viewport should exist
		_ = m.Viewport.View() // Should not panic
	}
}
