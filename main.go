package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type httpMethod int

const (
	get httpMethod = iota
	post
	put
	patch
	deleteMethod
)

var methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	methodStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	selectedMethodStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Bold(true).
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFFFF"))

	methodBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2)

	inputBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2)

	focusedInputBoxStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FF00FF")).
				Padding(1, 2)

	buttonStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4"))

	focusedButtonStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#FF00FF"))

	responseBoxStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#04B575")).
				Padding(1, 2)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

type responseView int

const (
	viewBody responseView = iota
	viewHeaders
)

type focusArea int

const (
	focusMethod focusArea = iota
	focusURL
	focusResponse
	focusHelp
	focusHeaders
)

type headerFormMode int

const (
	headerModeList headerFormMode = iota
	headerModeEdit
)

type model struct {
	focus            focusArea
	methodIdx        int
	urlInput         textinput.Model
	headers          string
	body             string
	response         string
	responseHeaders  string
	statusCode       string
	submitting       bool
	loading          bool
	methodOpen       bool // dropdown open state
	width            int
	height           int
	viewport         viewport.Model
	viewportReady    bool
	fullscreen       bool // fullscreen response view
	currentView      responseView // body or headers
	showHelp         bool // show help manual
	showHeadersForm  bool // show headers form
	requestHeaders   []HeaderPair // request headers
	headerKeyInput   textinput.Model
	headerValInput   textinput.Model
	headerSelectedIdx int  // which header is selected in the list
	headerFormMode   headerFormMode
	headerFocusField int // 0 = key, 1 = value
	headerIsEditing  bool // true if editing existing, false if adding new
}

type HeaderPair struct {
	Key   string
	Value string
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "https://api.example.com/endpoint"
	ti.CharLimit = 500
	ti.Width = 80

	headerKey := textinput.New()
	headerKey.Placeholder = "Header-Name"
	headerKey.CharLimit = 100
	headerKey.Width = 30

	headerVal := textinput.New()
	headerVal.Placeholder = "Header-Value"
	headerVal.CharLimit = 200
	headerVal.Width = 50

	vp := viewport.New(0, 0)
	vp.KeyMap = viewport.KeyMap{} // Disable default keybindings

	return model{
		focus:             focusMethod,
		methodIdx:         0,
		urlInput:          ti,
		width:             100,
		height:            40,
		methodOpen:        false,
		viewport:          vp,
		currentView:       viewBody,
		requestHeaders:    []HeaderPair{},
		headerKeyInput:    headerKey,
		headerValInput:    headerVal,
		headerSelectedIdx: 0,
		headerFormMode:    headerModeList,
		headerFocusField:  0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

type responseMsg struct {
	resp    string
	headers string
	status  string
	err     error
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update viewport size
		if !m.viewportReady {
			m.viewport = viewport.New(m.width-10, m.height-14)
			m.viewportReady = true
		} else {
			m.viewport.Width = m.width - 10
			m.viewport.Height = m.height - 14
		}
		
		// Update text input width
		m.urlInput.Width = m.width - 26
		return m, nil

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			if m.focus == focusResponse && m.response != "" {
				m.viewport.LineUp(3)
			}
			return m, nil
		case tea.MouseWheelDown:
			if m.focus == focusResponse && m.response != "" {
				m.viewport.LineDown(3)
			}
			return m, nil
		case tea.MouseLeft:
			// Click to focus - simplified click detection
			// Since we don't track exact box positions, we use vertical position as rough guide
			// Top area (method/url) vs bottom area (response)
			if msg.Y < 6 {
				// Clicked in top area - cycle between method and URL
				if m.focus == focusURL {
					m.urlInput.Blur()
					m.focus = focusMethod
				} else {
					m.focus = focusURL
					m.urlInput.Focus()
					cmds = append(cmds, textinput.Blink)
				}
			} else {
				// Clicked in response area
				if m.focus == focusURL {
					m.urlInput.Blur()
				}
				m.focus = focusResponse
			}
			return m, tea.Batch(cmds...)
		}

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		// If headers form is open, handle it separately
		if m.showHeadersForm {
			return m.updateHeadersForm(msg)
		}

		// If URL is focused, let text input handle most keys
		if m.focus == focusURL {
			switch msg.String() {
			case "tab", "shift+tab", "ctrl+c", "ctrl+s", "enter", "?":
				// Let these fall through to navigation/actions
			default:
				// Let text input handle the key (including h/j/k/l for typing)
				m.urlInput, cmd = m.urlInput.Update(msg)
				return m, cmd
			}
		}

		// If response is focused, handle scrolling with vim motions
		if m.focus == focusResponse && m.response != "" {
			switch msg.String() {
			case "f":
				// Toggle fullscreen
				m.fullscreen = !m.fullscreen
				return m, nil
			case "t":
				// Toggle between body and headers
				if m.currentView == viewBody {
					m.currentView = viewHeaders
				} else {
					m.currentView = viewBody
				}
				// Update viewport content
				content := m.response
				if m.currentView == viewHeaders {
					content = m.responseHeaders
				}
				m.viewport.SetContent(wrapText(content, m.viewport.Width))
				m.viewport.GotoTop()
				return m, nil
			case "j", "down":
				m.viewport.LineDown(1)
				return m, nil
			case "k", "up":
				m.viewport.LineUp(1)
				return m, nil
			case "d":
				m.viewport.HalfViewDown()
				return m, nil
			case "u":
				m.viewport.HalfViewUp()
				return m, nil
			case "g":
				m.viewport.GotoTop()
				return m, nil
			case "G":
				m.viewport.GotoBottom()
				return m, nil
			}
		}

		// If method is focused, use j/k to cycle through methods
		if m.focus == focusMethod {
			switch msg.String() {
			case "j", "down":
				m.methodIdx++
				if m.methodIdx >= len(methods) {
					m.methodIdx = 0
				}
				return m, nil
			case "k", "up":
				m.methodIdx--
				if m.methodIdx < 0 {
					m.methodIdx = len(methods) - 1
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "?":
			// Toggle help manual
			m.showHelp = !m.showHelp
			return m, nil

		case "h":
			// Open headers form (only when not in URL input)
			if m.focus != focusURL && !m.loading {
				m.showHeadersForm = true
				m.headerFormMode = headerModeList
				m.headerSelectedIdx = 0
				m.headerFocusField = 0
				m.headerKeyInput.SetValue("")
				m.headerValInput.SetValue("")
				m.headerKeyInput.Blur()
				m.headerValInput.Blur()
				return m, nil
			}
			return m, nil

		case "ctrl+s":
			// Send request with ctrl+s
			if m.urlInput.Value() != "" && !m.loading {
				m.response = ""
				m.statusCode = "Sending..."
				m.loading = true
				return m, sendRequestCmd(methods[m.methodIdx], m.urlInput.Value(), m.requestHeaders, m.body)
			}
			return m, nil

		case "tab":
			// Tab cycles forward through all fields
			m.methodOpen = false
			if m.focus == focusMethod {
				m.focus = focusURL
				m.urlInput.Focus()
				cmds = append(cmds, textinput.Blink)
			} else if m.focus == focusURL {
				m.urlInput.Blur()
				m.focus = focusResponse
			} else {
				m.focus = focusMethod
			}
			return m, tea.Batch(cmds...)

		case "shift+tab":
			// Shift+tab cycles backward through all fields
			m.methodOpen = false
			if m.focus == focusMethod {
				m.focus = focusResponse
			} else if m.focus == focusURL {
				m.urlInput.Blur()
				m.focus = focusMethod
			} else {
				m.focus = focusURL
				m.urlInput.Focus()
				cmds = append(cmds, textinput.Blink)
			}
			return m, tea.Batch(cmds...)

		case "enter":
			if m.focus == focusURL {
				// Send request when enter is pressed in URL input
				if m.urlInput.Value() != "" && !m.loading {
					m.response = ""
					m.statusCode = "Sending..."
					m.loading = true
					return m, sendRequestCmd(methods[m.methodIdx], m.urlInput.Value(), m.requestHeaders, m.body)
				}
			}
			return m, nil
		}

	case responseMsg:
		m.loading = false
		if msg.err != nil {
			m.response = fmt.Sprintf("Error: %v", msg.err)
			m.responseHeaders = ""
			m.statusCode = "Error"
		} else {
			m.response = msg.resp
			m.responseHeaders = msg.headers
			m.statusCode = msg.status
		}
		// Update viewport content with wrapping
		content := m.response
		if m.currentView == viewHeaders && m.responseHeaders != "" {
			content = m.responseHeaders
		}
		m.viewport.SetContent(wrapText(content, m.viewport.Width))
		m.viewport.GotoTop()
	}

	// Update text input
	m.urlInput, cmd = m.urlInput.Update(msg)
	cmds = append(cmds, cmd)

	// Update viewport only when response is focused
	if m.focus == focusResponse {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}
	
	return m, tea.Batch(cmds...)
}

func updateInput(current string, msg tea.KeyMsg) string {
	switch msg.Type {
	case tea.KeyRunes:
		return current + msg.String()
	case tea.KeyBackspace, tea.KeyDelete:
		if len(current) > 0 {
			return current[:len(current)-1]
		}
	}
	return current
}

func (m model) View() string {
	// If showing help manual
	if m.showHelp {
		return m.renderHelp()
	}

	// If showing headers form
	if m.showHeadersForm {
		return m.renderHeadersForm()
	}

	// If in fullscreen mode, show only response
	if m.fullscreen && m.response != "" {
		viewType := "Body"
		if m.currentView == viewHeaders {
			viewType = "Headers"
		}
		responseLabel := labelStyle.Render(fmt.Sprintf("Response - %s (Fullscreen)", viewType))
		if m.statusCode != "" {
			responseLabel += " - " + lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#04B575")).
				Render(m.statusCode)
		}

		// Use full screen for viewport with room for borders
		m.viewport.Width = m.width - 8
		m.viewport.Height = m.height - 8
		responseView := m.viewport.View()

		responseDisplay := responseLabel + "\n" + responseView
		if m.currentView == viewBody {
			responseDisplay += "\n" + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Render(fmt.Sprintf("%.0f%%", m.viewport.ScrollPercent()*100))
		}

		fullscreenBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF00FF")).
			Padding(1).
			Width(m.width - 4).
			Height(m.height - 2)

		return "\n" + fullscreenBox.Render(responseDisplay)
	}

	var sections []string

	// Calculate responsive widths - use full terminal width
	boxWidth := m.width - 6
	if boxWidth < 40 {
		boxWidth = 40
	}

	// Use most of the terminal height for response
	responseHeight := m.height - 14
	if responseHeight < 5 {
		responseHeight = 5
	}

	// Title
	sections = append(sections, titleStyle.Render("🌐 APITTY - API Testing TUI"))

	// Request box: Method dropdown + URL input + Send button in single box
	var requestContent strings.Builder

	// Show selected method (no dropdown UI)
	if m.focus == focusMethod {
		requestContent.WriteString(selectedMethodStyle.Render(methods[m.methodIdx]))
	} else {
		requestContent.WriteString(methodStyle.Render(methods[m.methodIdx]))
	}
	requestContent.WriteString(" ")

	// URL input on same line
	urlDisplay := m.urlInput.View()
	requestContent.WriteString(urlDisplay)

	// Headers button with top margin
	requestContent.WriteString("\n\n")
	headerCount := len(m.requestHeaders)
	headerBtn := fmt.Sprintf("Headers: %d", headerCount)
	if m.focus == focusHeaders {
		requestContent.WriteString(focusedButtonStyle.Render(headerBtn))
	} else {
		requestContent.WriteString(buttonStyle.Render(headerBtn))
	}

	// Apply box style based on focus
	requestBoxStyle := inputBoxStyle.Width(boxWidth)
	if m.focus == focusMethod || m.focus == focusURL || m.focus == focusHeaders {
		requestBoxStyle = focusedInputBoxStyle.Width(boxWidth)
	}
	sections = append(sections, requestBoxStyle.Render(requestContent.String()))

	// Response display
	viewType := "Body"
	if m.currentView == viewHeaders {
		viewType = "Headers"
	}
	responseLabel := labelStyle.Render(fmt.Sprintf("Response - %s", viewType))
	if m.statusCode != "" {
		responseLabel += " - " + lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575")).
			Render(m.statusCode)
	}

	var responseView string
	if m.response == "" {
		responseView = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true).
			Render("Response will appear here...")
	} else {
		// Update viewport dimensions
		m.viewport.Width = boxWidth - 2
		m.viewport.Height = responseHeight - 2
		responseView = m.viewport.View()
	}

	responseDisplay := responseLabel + "\n" + responseView
	if m.response != "" && m.currentView == viewBody {
		responseDisplay += "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render(fmt.Sprintf("%.0f%%", m.viewport.ScrollPercent()*100))
	}

	responseBoxStyle := responseBoxStyle.Width(boxWidth).Height(responseHeight)
	if m.focus == focusResponse {
		responseBoxStyle = responseBoxStyle.BorderForeground(lipgloss.Color("#FF00FF"))
	}
	sections = append(sections, responseBoxStyle.Render(responseDisplay))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m model) renderHelp() string {
	helpContent := `
╭─────────────────────────────────────────────────────────────╮
│                   APITTY - KEYBINDINGS                      │
╰─────────────────────────────────────────────────────────────╯

GLOBAL KEYBINDINGS
  ?         Toggle this help manual
  q         Quit the application
  ctrl+c    Quit the application
  ctrl+s    Send HTTP request (from anywhere)

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

FULLSCREEN MODE (when active)
  f         Exit fullscreen
  t         Toggle between Body and Headers
  All scroll keys (j/k/d/u/g/G) work as normal

MOUSE SUPPORT
  Scroll    Scroll the response box (when focused)
  Click     Switch focus between elements

Press ? to close this help
`

	helpBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(m.width - 4).
		Height(m.height - 2)

	return "\n" + helpBox.Render(helpContent)
}

// sendRequestCmd performs the HTTP request in a goroutine and returns a tea.Cmd
func sendRequestCmd(method, url string, headers []HeaderPair, body string) tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 15 * time.Second}
		var reqBody io.Reader
		if method == "POST" || method == "PUT" || method == "PATCH" {
			reqBody = strings.NewReader(body)
		}
		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return responseMsg{resp: "", headers: "", status: "", err: err}
		}
		// Apply request headers
		for _, h := range headers {
			if h.Key != "" {
				req.Header.Set(h.Key, h.Value)
			}
		}
		resp, err := client.Do(req)
		if err != nil {
			return responseMsg{resp: "", headers: "", status: "", err: err}
		}
		defer resp.Body.Close()
		
		// Build headers string
		var headersBuilder strings.Builder
		for k, v := range resp.Header {
			headersBuilder.WriteString(fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", ")))
		}
		
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return responseMsg{resp: "", headers: "", status: "", err: err}
		}
		pretty := tryPrettyJSON(respBody)
		return responseMsg{resp: pretty, headers: headersBuilder.String(), status: resp.Status, err: nil}
	}
}

// tryPrettyJSON tries to pretty-print JSON, falls back to string if not JSON
func tryPrettyJSON(data []byte) string {
	trim := bytes.TrimSpace(data)
	if len(trim) == 0 {
		return ""
	}
	if trim[0] == '{' || trim[0] == '[' {
		var out bytes.Buffer
		err := jsonIndent(&out, trim)
		if err == nil {
			return colorizeJSON(out.String())
		}
	}
	return string(data)
}

// colorizeJSON adds ANSI color codes to JSON for syntax highlighting
func colorizeJSON(jsonStr string) string {
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

// wrapText wraps text to fit within the given width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	
	var result strings.Builder
	lines := strings.Split(text, "\n")
	
	for _, line := range lines {
		// Handle ANSI color codes - strip them for length calculation
		visibleLen := visibleLength(line)
		
		if visibleLen <= width {
			result.WriteString(line)
			result.WriteByte('\n')
			continue
		}
		
		// Wrap long lines
		currentPos := 0
		for currentPos < len(line) {
			// Find how many characters fit in width
			chunkEnd := findChunkEnd(line, currentPos, width)
			
			if chunkEnd <= currentPos {
				break
			}
			
			// Extract chunk
			chunk := line[currentPos:chunkEnd]
			
			// Try to break at a good position (space, comma, etc.)
			if chunkEnd < len(line) {
				// Look back for a good break point
				for i := len(chunk) - 1; i >= max(0, len(chunk)-15); i-- {
					if chunk[i] == ' ' || chunk[i] == ',' || chunk[i] == ':' {
						chunk = chunk[:i+1]
						chunkEnd = currentPos + len(chunk)
						break
					}
				}
			}
			
			result.WriteString(chunk)
			result.WriteByte('\n')
			
			// Skip leading spaces on next line
			currentPos = chunkEnd
			for currentPos < len(line) && line[currentPos] == ' ' {
				currentPos++
			}
		}
	}
	
	return result.String()
}

// visibleLength returns the length of string without ANSI codes
func visibleLength(s string) int {
	length := 0
	inAnsi := false
	
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inAnsi = true
			continue
		}
		
		if inAnsi {
			if s[i] == 'm' {
				inAnsi = false
			}
			continue
		}
		
		length++
	}
	
	return length
}

// findChunkEnd finds where to end a chunk based on visible width
func findChunkEnd(s string, start, width int) int {
	visibleCount := 0
	pos := start
	inAnsi := false
	
	for pos < len(s) {
		if s[pos] == '\x1b' {
			inAnsi = true
			pos++
			continue
		}
		
		if inAnsi {
			if s[pos] == 'm' {
				inAnsi = false
			}
			pos++
			continue
		}
		
		if visibleCount >= width {
			return pos
		}
		
		visibleCount++
		pos++
	}
	
	return pos
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// horizontalScroll applies horizontal offset to text while preserving ANSI colors
// jsonIndent indents JSON for pretty printing
func jsonIndent(out *bytes.Buffer, data []byte) error {
	type jsonRaw = map[string]interface{}
	type jsonArr = []interface{}
	// Try object
	if data[0] == '{' {
		var obj jsonRaw
		if err := jsonUnmarshal(data, &obj); err == nil {
			return jsonMarshalIndent(obj, out)
		}
	}
	// Try array
	if data[0] == '[' {
		var arr jsonArr
		if err := jsonUnmarshal(data, &arr); err == nil {
			return jsonMarshalIndent(arr, out)
		}
	}
	return fmt.Errorf("not JSON")
}

// Use encoding/json but avoid import collision with bubbletea
func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
func jsonMarshalIndent(v interface{}, out *bytes.Buffer) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	out.Write(b)
	return nil
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}


func (m model) updateHeadersForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
var cmd tea.Cmd

// Handle mode-specific keys
if m.headerFormMode == headerModeList {
// List mode - navigate and manage headers
switch msg.String() {
case "esc", "ctrl+c", "q":
m.showHeadersForm = false
return m, nil

case "a", "n":
// Add new header
m.headerFormMode = headerModeEdit
m.headerFocusField = 0
m.headerIsEditing = false
m.headerKeyInput.SetValue("")
m.headerValInput.SetValue("")
m.headerKeyInput.Focus()
m.headerValInput.Blur()
return m, textinput.Blink

case "j", "down":
if len(m.requestHeaders) > 0 {
m.headerSelectedIdx++
if m.headerSelectedIdx >= len(m.requestHeaders) {
m.headerSelectedIdx = 0
}
}
return m, nil

case "k", "up":
if len(m.requestHeaders) > 0 {
m.headerSelectedIdx--
if m.headerSelectedIdx < 0 {
m.headerSelectedIdx = len(m.requestHeaders) - 1
}
}
return m, nil

case "d", "x", "backspace", "delete":
// Delete selected header
if len(m.requestHeaders) > 0 && m.headerSelectedIdx >= 0 && m.headerSelectedIdx < len(m.requestHeaders) {
m.requestHeaders = append(m.requestHeaders[:m.headerSelectedIdx], m.requestHeaders[m.headerSelectedIdx+1:]...)
if m.headerSelectedIdx >= len(m.requestHeaders) && len(m.requestHeaders) > 0 {
m.headerSelectedIdx = len(m.requestHeaders) - 1
}
if len(m.requestHeaders) == 0 {
m.headerSelectedIdx = 0
}
}
return m, nil

case "e", "enter":
// Edit selected header
if len(m.requestHeaders) > 0 && m.headerSelectedIdx >= 0 && m.headerSelectedIdx < len(m.requestHeaders) {
h := m.requestHeaders[m.headerSelectedIdx]
m.headerFormMode = headerModeEdit
m.headerFocusField = 0
m.headerIsEditing = true
m.headerKeyInput.SetValue(h.Key)
m.headerValInput.SetValue(h.Value)
m.headerKeyInput.Focus()
m.headerValInput.Blur()
return m, textinput.Blink
}
return m, nil
}
return m, nil

} else {
// Edit mode - editing a header
switch msg.String() {
case "esc":
// Cancel edit and go back to list
m.headerFormMode = headerModeList
m.headerKeyInput.Blur()
m.headerValInput.Blur()
m.headerKeyInput.SetValue("")
m.headerValInput.SetValue("")
return m, nil

case "ctrl+c":
// Close form entirely
m.showHeadersForm = false
m.headerKeyInput.Blur()
m.headerValInput.Blur()
return m, nil

case "tab":
// Switch between key and value
if m.headerFocusField == 0 {
m.headerFocusField = 1
m.headerKeyInput.Blur()
m.headerValInput.Focus()
} else {
m.headerFocusField = 0
m.headerValInput.Blur()
m.headerKeyInput.Focus()
}
return m, textinput.Blink

case "enter":
// Save header
key := strings.TrimSpace(m.headerKeyInput.Value())
val := strings.TrimSpace(m.headerValInput.Value())

if key != "" {
if m.headerIsEditing {
// Update existing
m.requestHeaders[m.headerSelectedIdx] = HeaderPair{Key: key, Value: val}
} else {
// Add new
m.requestHeaders = append(m.requestHeaders, HeaderPair{Key: key, Value: val})
}
}

// Go back to list mode
m.headerFormMode = headerModeList
m.headerIsEditing = false
m.headerKeyInput.SetValue("")
m.headerValInput.SetValue("")
m.headerKeyInput.Blur()
m.headerValInput.Blur()
return m, nil

default:
// Pass through to text inputs
if m.headerFocusField == 0 {
m.headerKeyInput, cmd = m.headerKeyInput.Update(msg)
} else {
m.headerValInput, cmd = m.headerValInput.Update(msg)
}
return m, cmd
}
}
}

func (m model) renderHeadersForm() string {
var content strings.Builder

content.WriteString(titleStyle.Render("Request Headers"))
content.WriteString("\n\n")

// Show mode
var modeStr string
if m.headerFormMode == headerModeList {
modeStr = labelStyle.Render("[ LIST MODE ]")
} else {
modeStr = labelStyle.Render("[ EDIT MODE ]")
}
content.WriteString(modeStr)
content.WriteString("\n\n")

// Show existing headers in list mode
if m.headerFormMode == headerModeList {
if len(m.requestHeaders) > 0 {
content.WriteString(labelStyle.Render("Headers:"))
content.WriteString("\n")
for i, h := range m.requestHeaders {
prefix := "  "
lineStyle := lipgloss.NewStyle()
if i == m.headerSelectedIdx {
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
content.WriteString(labelStyle.Render("Edit Header:"))
content.WriteString("\n\n")

keyLabel := "Key:   "
if m.headerFocusField == 0 {
keyLabel = lipgloss.NewStyle().
Foreground(lipgloss.Color("#FF00FF")).
Bold(true).
Render("Key:   ")
}
content.WriteString(keyLabel + m.headerKeyInput.View() + "\n")

valLabel := "Value: "
if m.headerFocusField == 1 {
valLabel = lipgloss.NewStyle().
Foreground(lipgloss.Color("#FF00FF")).
Bold(true).
Render("Value: ")
}
content.WriteString(valLabel + m.headerValInput.View() + "\n\n")

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
Width(m.width - 4).
Height(m.height - 2)

return "\n" + formBox.Render(content.String())
}
