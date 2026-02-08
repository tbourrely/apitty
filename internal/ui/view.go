package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tbourrel/apitty/internal/model"
	"github.com/tbourrel/apitty/internal/text"
)

// RenderMain renders the main application view
func RenderMain(m model.Model) string {
	var sections []string

	// Calculate responsive widths - use full terminal width
	boxWidth := m.Width - 6
	if boxWidth < 40 {
		boxWidth = 40
	}

	// Use most of the terminal height for response
	responseHeight := m.Height - 14
	if responseHeight < 5 {
		responseHeight = 5
	}

	// Title
	sections = append(sections, TitleStyle.Render("🌐 APITTY - API Testing TUI"))

	// Request box: Method dropdown + URL input + Send button in single box
	var requestContent strings.Builder

	// Show selected method (no dropdown UI)
	if m.Focus == model.FocusMethod {
		requestContent.WriteString(SelectedMethodStyle.Render(model.Methods[m.MethodIdx]))
	} else {
		requestContent.WriteString(MethodStyle.Render(model.Methods[m.MethodIdx]))
	}
	requestContent.WriteString(" ")

	// URL input on same line
	urlDisplay := m.URLInput.View()
	requestContent.WriteString(urlDisplay)

	// Headers button with top margin
	requestContent.WriteString("\n\n")
	headerCount := len(m.RequestHeaders)
	headerBtn := fmt.Sprintf("Headers: %d", headerCount)
	if m.Focus == model.FocusHeaders {
		requestContent.WriteString(FocusedButtonStyle.Render(headerBtn))
	} else {
		requestContent.WriteString(ButtonStyle.Render(headerBtn))
	}

	// Apply box style based on focus
	requestBoxStyle := InputBoxStyle.Width(boxWidth)
	if m.Focus == model.FocusMethod || m.Focus == model.FocusURL || m.Focus == model.FocusHeaders {
		requestBoxStyle = FocusedInputBoxStyle.Width(boxWidth)
	}
	sections = append(sections, requestBoxStyle.Render(requestContent.String()))

	// Response display
	viewType := "Body"
	if m.CurrentView == model.ViewHeaders {
		viewType = "Headers"
	}
	responseLabel := LabelStyle.Render(fmt.Sprintf("Response - %s", viewType))
	if m.StatusCode != "" {
		responseLabel += " - " + lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575")).
			Render(m.StatusCode)
	}

	var responseView string
	if m.Response == "" {
		responseView = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true).
			Render("Response will appear here...")
	} else {
		responseView = m.Viewport.View()
	}

	responseDisplay := responseLabel + "\n" + responseView
	if m.Response != "" && m.CurrentView == model.ViewBody {
		responseDisplay += "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render(fmt.Sprintf("%.0f%%", m.Viewport.ScrollPercent()*100))
	}

	responseBoxStyle := ResponseBoxStyle.Width(boxWidth).Height(responseHeight)
	if m.Focus == model.FocusResponse {
		responseBoxStyle = responseBoxStyle.BorderForeground(lipgloss.Color("#FF00FF"))
	}
	sections = append(sections, responseBoxStyle.Render(responseDisplay))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// RenderFullscreen renders the fullscreen response view
func RenderFullscreen(m model.Model) string {
	viewType := "Body"
	if m.CurrentView == model.ViewHeaders {
		viewType = "Headers"
	}
	responseLabel := LabelStyle.Render(fmt.Sprintf("Response - %s (Fullscreen)", viewType))
	if m.StatusCode != "" {
		responseLabel += " - " + lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575")).
			Render(m.StatusCode)
	}

	responseView := m.Viewport.View()

	responseDisplay := responseLabel + "\n" + responseView
	if m.CurrentView == model.ViewBody {
		responseDisplay += "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render(fmt.Sprintf("%.0f%%", m.Viewport.ScrollPercent()*100))
	}

	fullscreenBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF00FF")).
		Padding(1).
		Width(m.Width - 4).
		Height(m.Height - 2)

	return "\n" + fullscreenBox.Render(responseDisplay)
}

// RenderHelp renders the help screen
func RenderHelp(m model.Model) string {
	helpBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.Width - 4).
		Height(m.Height - 2)

	scrollInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(fmt.Sprintf("%.0f%%", m.HelpViewport.ScrollPercent()*100))

	return "\n" + helpBox.Render(m.HelpViewport.View()+"\n"+scrollInfo)
}

// RenderHeadersForm renders the headers form modal
func RenderHeadersForm(m model.Model) string {
	var content strings.Builder

	content.WriteString(TitleStyle.Render("Request Headers"))
	content.WriteString("\n\n")

	// Show mode
	var modeStr string
	if m.HeaderFormMode == model.HeaderModeList {
		modeStr = LabelStyle.Render("[ LIST MODE ]")
	} else {
		modeStr = LabelStyle.Render("[ EDIT MODE ]")
	}
	content.WriteString(modeStr)
	content.WriteString("\n\n")

	// Show existing headers in list mode
	if m.HeaderFormMode == model.HeaderModeList {
		if len(m.RequestHeaders) > 0 {
			content.WriteString(LabelStyle.Render("Headers:"))
			content.WriteString("\n")
			for i, h := range m.RequestHeaders {
				prefix := "  "
				lineStyle := lipgloss.NewStyle()
				if i == m.HeaderSelectedIdx {
					prefix = "➤ "
					lineStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#FF00FF")).
						Bold(true)
				}
				headerLine := fmt.Sprintf("%s%d. %s: %s", prefix, i+1, h.Key, h.Value)
				content.WriteString(lineStyle.Render(headerLine))
				content.WriteString("\n")
			}
		} else {
			content.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Italic(true).
				Render("No headers yet. Press 'a' to add one."))
		}
		content.WriteString("\n\n")

		// List mode instructions
		instructions := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render("j/k: navigate • a: add new • e/enter: edit • d/x: delete • esc/q: close")
		content.WriteString(instructions)
	} else {
		// Edit mode - show input form
		content.WriteString(LabelStyle.Render("Edit Header:"))
		content.WriteString("\n\n")

		keyLabel := "Key:   "
		if m.HeaderFocusField == 0 {
			keyLabel = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF00FF")).
				Bold(true).
				Render("Key:   ")
		}
		content.WriteString(keyLabel + m.HeaderKeyInput.View() + "\n")

		valLabel := "Value: "
		if m.HeaderFocusField == 1 {
			valLabel = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF00FF")).
				Bold(true).
				Render("Value: ")
		}
		content.WriteString(valLabel + m.HeaderValInput.View() + "\n\n")

		// Edit mode instructions
		instructions := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render("tab: switch field • enter: save • esc: cancel • ctrl+c: close")
		content.WriteString(instructions)
	}

	formBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(2, 4).
		Width(m.Width - 4).
		Height(m.Height - 2)

	return "\n" + formBox.Render(content.String())
}

// RenderCurlImport renders the cURL import modal
func RenderCurlImport(m model.Model) string {
	var content strings.Builder

	content.WriteString(TitleStyle.Render("Import from cURL"))
	content.WriteString("\n\n")

	content.WriteString(LabelStyle.Render("Paste your curl command:"))
	content.WriteString("\n\n")
	content.WriteString(m.CurlInput.View())
	content.WriteString("\n\n")

	// Instructions
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("enter: import • esc: cancel")
	content.WriteString(instructions)

	// Example
	content.WriteString("\n\n")
	example := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true).
		Render("Example: curl -X POST https://api.example.com/users -H \"Content-Type: application/json\"")
	content.WriteString(example)

	modalBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(2, 4).
		Width(m.Width - 20).
		Height(14)

	// Center the modal
	return "\n\n\n" + lipgloss.Place(
		m.Width, m.Height,
		lipgloss.Center, lipgloss.Center,
		modalBox.Render(content.String()),
	)
}

// GetHelpContent returns the help text content
func GetHelpContent() string {
	return `
╭─────────────────────────────────────────────────────────────╮
│                   APITTY - KEYBINDINGS                      │
╰─────────────────────────────────────────────────────────────╯

GLOBAL KEYBINDINGS
  ?         Toggle this help manual
  q         Quit the application
  ctrl+c    Quit the application
  ctrl+s    Send HTTP request (from anywhere)
  h         Open headers form (add/edit request headers)
  i         Import from cURL command

NAVIGATION
  tab       Cycle forward through fields (Method → URL → Response)
  shift+tab Cycle backward through fields

METHOD SELECTOR (when focused)
  j / ↓     Next HTTP method (GET → POST → PUT → PATCH → DELETE)
  k / ↑     Previous HTTP method

URL INPUT (when focused)
  enter     Send HTTP request
  Type normally - all keys work including h/j/k/l

RESPONSE BOX (when focused)
  t         Toggle between Body and Headers view
  f         Toggle fullscreen mode
  j / ↓     Scroll down one line
  k / ↑     Scroll up one line
  d         Scroll down half page
  u         Scroll up half page
  g         Jump to top
  G         Jump to bottom
  w         Toggle text wrapping

FULLSCREEN MODE (when active)
  f         Exit fullscreen
  t         Toggle between Body and Headers
  All scroll keys (j/k/d/u/g/G) work as normal

MOUSE SUPPORT
  Scroll    Scroll the response box (when focused)
  Click     Switch focus between elements

HELP PAGE NAVIGATION
  j/k       Scroll up/down
  d/u       Scroll half page
  g/G       Jump to top/bottom
  ?/q/esc   Close help

Press ? to close this help
`
}

// UpdateViewportContent updates the viewport content with proper wrapping
func UpdateViewportContent(m *model.Model) {
	content := m.Response
	if m.CurrentView == model.ViewHeaders && m.ResponseHeaders != "" {
		content = m.ResponseHeaders
	}
	m.Viewport.SetContent(text.WrapText(content, m.Viewport.Width))
}
