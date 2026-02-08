package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tbourrel/apitty/internal/http"
	"github.com/tbourrel/apitty/internal/model"
	"github.com/tbourrel/apitty/internal/parser"
	"github.com/tbourrel/apitty/internal/text"
)

// Update handles all messages and updates the model
func Update(m model.Model, msg tea.Msg) (model.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Update viewport size
		if !m.ViewportReady {
			m.Viewport.Width = m.Width - 10
			m.Viewport.Height = m.Height - 14
			m.ViewportReady = true
		} else {
			m.Viewport.Width = m.Width - 10
			m.Viewport.Height = m.Height - 14
		}

		// Update text input width
		m.URLInput.Width = m.Width - 26
		return m, nil

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonWheelUp {
			if m.Focus == model.FocusResponse && m.Response != "" {
				m.Viewport.ScrollUp(3)
			}
			return m, nil
		}
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonWheelDown {
			if m.Focus == model.FocusResponse && m.Response != "" {
				m.Viewport.ScrollDown(3)
			}
			return m, nil
		}
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Click to focus
			if msg.Y < 6 {
				// Clicked in top area
				if m.Focus == model.FocusURL {
					m.URLInput.Blur()
					m.Focus = model.FocusMethod
				} else {
					m.Focus = model.FocusURL
					m.URLInput.Focus()
					cmds = append(cmds, textinput.Blink)
				}
			} else {
				// Clicked in response area
				if m.Focus == model.FocusURL {
					m.URLInput.Blur()
				}
				m.Focus = model.FocusResponse
			}
			return m, tea.Batch(cmds...)
		}

	case tea.KeyMsg:
		if m.Loading {
			return m, nil
		}

		// If help is showing, handle help-specific keys
		if m.ShowHelp {
			return handleHelpKeys(m, msg)
		}

		// If headers form is open, handle it separately
		if m.ShowHeadersForm {
			return updateHeadersForm(m, msg)
		}

		// If curl import is open, handle it separately
		if m.ShowCurlImport {
			return updateCurlImport(m, msg)
		}

		// If URL is focused, let text input handle most keys
		if m.Focus == model.FocusURL {
			switch msg.String() {
			case "tab", "shift+tab", "ctrl+c", "ctrl+s", "enter", "?":
				// Let these fall through to navigation/actions
			default:
				// Let text input handle the key
				m.URLInput, cmd = m.URLInput.Update(msg)
				return m, cmd
			}
		}

		// If response is focused, handle scrolling
		if m.Focus == model.FocusResponse && m.Response != "" {
			if handleResponseKeys(&m, msg) {
				return m, nil
			}
		}

		// If method is focused, use j/k to cycle
		if m.Focus == model.FocusMethod {
			if handleMethodKeys(&m, msg) {
				return m, nil
			}
		}

		// Handle global keys
		return handleGlobalKeys(m, msg, cmds)

	case model.ResponseMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Response = fmt.Sprintf("Error: %v", msg.Err)
			m.ResponseHeaders = ""
			m.StatusCode = "Error"
		} else {
			m.Response = msg.Resp
			m.ResponseHeaders = msg.Headers
			m.StatusCode = msg.Status
		}
		// Update viewport content with wrapping
		content := m.Response
		if m.CurrentView == model.ViewHeaders && m.ResponseHeaders != "" {
			content = m.ResponseHeaders
		}
		m.Viewport.SetContent(text.WrapText(content, m.Viewport.Width))
		m.Viewport.GotoTop()
	}

	// Update text input
	m.URLInput, cmd = m.URLInput.Update(msg)
	cmds = append(cmds, cmd)

	// Update viewport only when response is focused
	if m.Focus == model.FocusResponse {
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func handleHelpKeys(m model.Model, msg tea.KeyMsg) (model.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "q", "esc":
		m.ShowHelp = false
		return m, nil
	case "j", "down":
		m.HelpViewport.ScrollDown(1)
		return m, nil
	case "k", "up":
		m.HelpViewport.ScrollUp(1)
		return m, nil
	case "d":
		m.HelpViewport.HalfPageDown()
		return m, nil
	case "u":
		m.HelpViewport.HalfPageUp()
		return m, nil
	case "g":
		m.HelpViewport.GotoTop()
		return m, nil
	case "G":
		m.HelpViewport.GotoBottom()
		return m, nil
	}
	return m, nil
}

func handleResponseKeys(m *model.Model, msg tea.KeyMsg) bool {
	switch msg.String() {
	case "f":
		m.Fullscreen = !m.Fullscreen
		return true
	case "t":
		if m.CurrentView == model.ViewBody {
			m.CurrentView = model.ViewHeaders
		} else {
			m.CurrentView = model.ViewBody
		}
		content := m.Response
		if m.CurrentView == model.ViewHeaders {
			content = m.ResponseHeaders
		}
		m.Viewport.SetContent(text.WrapText(content, m.Viewport.Width))
		m.Viewport.GotoTop()
		return true
	case "j", "down":
		m.Viewport.ScrollDown(1)
		return true
	case "k", "up":
		m.Viewport.ScrollUp(1)
		return true
	case "d":
		m.Viewport.HalfPageDown()
		return true
	case "u":
		m.Viewport.HalfPageUp()
		return true
	case "g":
		m.Viewport.GotoTop()
		return true
	case "G":
		m.Viewport.GotoBottom()
		return true
	}
	return false
}

func handleMethodKeys(m *model.Model, msg tea.KeyMsg) bool {
	switch msg.String() {
	case "j", "down":
		m.MethodIdx++
		if m.MethodIdx >= len(model.Methods) {
			m.MethodIdx = 0
		}
		return true
	case "k", "up":
		m.MethodIdx--
		if m.MethodIdx < 0 {
			m.MethodIdx = len(model.Methods) - 1
		}
		return true
	}
	return false
}

func handleGlobalKeys(m model.Model, msg tea.KeyMsg, cmds []tea.Cmd) (model.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "?":
		if !m.ShowHelp {
			m.ShowHelp = true
			m.HelpViewport.Width = m.Width - 8
			m.HelpViewport.Height = m.Height - 8
			m.HelpViewport.SetContent(GetHelpContent())
			m.HelpViewport.GotoTop()
		} else {
			m.ShowHelp = false
		}
		return m, nil

	case "i":
		if m.Focus != model.FocusURL && !m.Loading {
			m.ShowCurlImport = true
			m.CurlInput.SetValue("")
			m.CurlInput.Focus()
			return m, textinput.Blink
		}
		return m, nil

	case "h":
		if m.Focus != model.FocusURL && !m.Loading {
			m.ShowHeadersForm = true
			m.HeaderFormMode = model.HeaderModeList
			m.HeaderSelectedIdx = 0
			m.HeaderFocusField = 0
			m.HeaderKeyInput.SetValue("")
			m.HeaderValInput.SetValue("")
			m.HeaderKeyInput.Blur()
			m.HeaderValInput.Blur()
			return m, nil
		}
		return m, nil

	case "ctrl+s":
		if m.URLInput.Value() != "" && !m.Loading {
			m.Response = ""
			m.StatusCode = "Sending..."
			m.Loading = true
			return m, http.SendRequestCmd(model.Methods[m.MethodIdx], m.URLInput.Value(), m.RequestHeaders, m.Body)
		}
		return m, nil

	case "tab":
		m.MethodOpen = false
		switch m.Focus {
		case model.FocusMethod:
			m.Focus = model.FocusURL
			m.URLInput.Focus()
			cmds = append(cmds, textinput.Blink)
		case model.FocusURL:
			m.URLInput.Blur()
			m.Focus = model.FocusResponse
		default:
			m.Focus = model.FocusMethod
		}
		return m, tea.Batch(cmds...)

	case "shift+tab":
		m.MethodOpen = false
		switch m.Focus {
		case model.FocusMethod:
			m.Focus = model.FocusResponse
		case model.FocusURL:
			m.URLInput.Blur()
			m.Focus = model.FocusMethod
		default:
			m.Focus = model.FocusURL
			m.URLInput.Focus()
			cmds = append(cmds, textinput.Blink)
		}
		return m, tea.Batch(cmds...)

	case "enter":
		if m.Focus == model.FocusURL {
			if m.URLInput.Value() != "" && !m.Loading {
				m.Response = ""
				m.StatusCode = "Sending..."
				m.Loading = true
				return m, http.SendRequestCmd(model.Methods[m.MethodIdx], m.URLInput.Value(), m.RequestHeaders, m.Body)
			}
		}
		return m, nil
	}

	return m, nil
}

func updateHeadersForm(m model.Model, msg tea.KeyMsg) (model.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.HeaderFormMode == model.HeaderModeList {
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			m.ShowHeadersForm = false
			return m, nil

		case "a", "n":
			m.HeaderFormMode = model.HeaderModeEdit
			m.HeaderFocusField = 0
			m.HeaderIsEditing = false
			m.HeaderKeyInput.SetValue("")
			m.HeaderValInput.SetValue("")
			m.HeaderKeyInput.Focus()
			m.HeaderValInput.Blur()
			return m, textinput.Blink

		case "j", "down":
			if len(m.RequestHeaders) > 0 {
				m.HeaderSelectedIdx++
				if m.HeaderSelectedIdx >= len(m.RequestHeaders) {
					m.HeaderSelectedIdx = 0
				}
			}
			return m, nil

		case "k", "up":
			if len(m.RequestHeaders) > 0 {
				m.HeaderSelectedIdx--
				if m.HeaderSelectedIdx < 0 {
					m.HeaderSelectedIdx = len(m.RequestHeaders) - 1
				}
			}
			return m, nil

		case "d", "x", "backspace", "delete":
			if len(m.RequestHeaders) > 0 && m.HeaderSelectedIdx >= 0 && m.HeaderSelectedIdx < len(m.RequestHeaders) {
				m.RequestHeaders = append(m.RequestHeaders[:m.HeaderSelectedIdx], m.RequestHeaders[m.HeaderSelectedIdx+1:]...)
				if m.HeaderSelectedIdx >= len(m.RequestHeaders) && len(m.RequestHeaders) > 0 {
					m.HeaderSelectedIdx = len(m.RequestHeaders) - 1
				}
				if len(m.RequestHeaders) == 0 {
					m.HeaderSelectedIdx = 0
				}
			}
			return m, nil

		case "e", "enter":
			if len(m.RequestHeaders) > 0 && m.HeaderSelectedIdx >= 0 && m.HeaderSelectedIdx < len(m.RequestHeaders) {
				h := m.RequestHeaders[m.HeaderSelectedIdx]
				m.HeaderFormMode = model.HeaderModeEdit
				m.HeaderFocusField = 0
				m.HeaderIsEditing = true
				m.HeaderKeyInput.SetValue(h.Key)
				m.HeaderValInput.SetValue(h.Value)
				m.HeaderKeyInput.Focus()
				m.HeaderValInput.Blur()
				return m, textinput.Blink
			}
			return m, nil
		}
		return m, nil

	} else {
		switch msg.String() {
		case "esc":
			m.HeaderFormMode = model.HeaderModeList
			m.HeaderKeyInput.Blur()
			m.HeaderValInput.Blur()
			m.HeaderKeyInput.SetValue("")
			m.HeaderValInput.SetValue("")
			return m, nil

		case "ctrl+c":
			m.ShowHeadersForm = false
			m.HeaderKeyInput.Blur()
			m.HeaderValInput.Blur()
			return m, nil

		case "tab":
			if m.HeaderFocusField == 0 {
				m.HeaderFocusField = 1
				m.HeaderKeyInput.Blur()
				m.HeaderValInput.Focus()
			} else {
				m.HeaderFocusField = 0
				m.HeaderValInput.Blur()
				m.HeaderKeyInput.Focus()
			}
			return m, textinput.Blink

		case "enter":
			key := strings.TrimSpace(m.HeaderKeyInput.Value())
			val := strings.TrimSpace(m.HeaderValInput.Value())

			if key != "" {
				if m.HeaderIsEditing {
					m.RequestHeaders[m.HeaderSelectedIdx] = model.HeaderPair{Key: key, Value: val}
				} else {
					m.RequestHeaders = append(m.RequestHeaders, model.HeaderPair{Key: key, Value: val})
				}
			}

			m.HeaderFormMode = model.HeaderModeList
			m.HeaderIsEditing = false
			m.HeaderKeyInput.SetValue("")
			m.HeaderValInput.SetValue("")
			m.HeaderKeyInput.Blur()
			m.HeaderValInput.Blur()
			return m, nil

		default:
			if m.HeaderFocusField == 0 {
				m.HeaderKeyInput, cmd = m.HeaderKeyInput.Update(msg)
			} else {
				m.HeaderValInput, cmd = m.HeaderValInput.Update(msg)
			}
			return m, cmd
		}
	}
}

func updateCurlImport(m model.Model, msg tea.KeyMsg) (model.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc", "ctrl+c":
		m.ShowCurlImport = false
		m.CurlInput.Blur()
		return m, nil

	case "enter":
		curlCmd := m.CurlInput.Value()
		if curlCmd != "" {
			method, url, headers := parser.ParseCurlCommand(curlCmd)

			// Apply parsed values
			if url != "" {
				m.URLInput.SetValue(url)
			}

			// Set method
			for idx, meth := range model.Methods {
				if meth == method {
					m.MethodIdx = idx
					break
				}
			}

			// Add headers
			m.RequestHeaders = headers
		}
		m.ShowCurlImport = false
		m.CurlInput.Blur()
		m.CurlInput.SetValue("")
		return m, nil

	default:
		m.CurlInput, cmd = m.CurlInput.Update(msg)
		return m, cmd
	}
}

// View renders the appropriate view based on current state
func View(m model.Model) string {
	if m.ShowHelp {
		return RenderHelp(m)
	}

	if m.ShowCurlImport {
		return RenderCurlImport(m)
	}

	if m.ShowHeadersForm {
		return RenderHeadersForm(m)
	}

	if m.Fullscreen && m.Response != "" {
		// Update viewport dimensions for fullscreen
		m.Viewport.Width = m.Width - 8
		m.Viewport.Height = m.Height - 8
		return RenderFullscreen(m)
	}

	// Update viewport dimensions for normal view
	boxWidth := m.Width - 6
	if boxWidth < 40 {
		boxWidth = 40
	}
	responseHeight := m.Height - 14
	if responseHeight < 5 {
		responseHeight = 5
	}
	m.Viewport.Width = boxWidth - 2
	m.Viewport.Height = responseHeight - 2

	return RenderMain(m)
}
