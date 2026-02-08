package model

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Methods contains all supported HTTP methods
var Methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

// ResponseView represents the current view mode for responses
type ResponseView int

const (
	// ViewBody shows the response body
	ViewBody ResponseView = iota
	// ViewHeaders shows the response headers
	ViewHeaders
)

// FocusArea represents which UI element is currently focused
type FocusArea int

const (
	// FocusMethod indicates the method selector is focused
	FocusMethod FocusArea = iota
	// FocusURL indicates the URL input is focused
	FocusURL
	// FocusResponse indicates the response viewer is focused
	FocusResponse
	// FocusHelp indicates the help screen is focused
	FocusHelp
	// FocusHeaders indicates the headers form is focused
	FocusHeaders
)

// HeaderFormMode represents the current state of the headers form
type HeaderFormMode int

const (
	// HeaderModeList shows the list of headers
	HeaderModeList HeaderFormMode = iota
	// HeaderModeEdit allows editing a header
	HeaderModeEdit
)

// HeaderPair represents a single HTTP header key-value pair
type HeaderPair struct {
	Key   string
	Value string
}

// Model represents the application state
type Model struct {
	Focus             FocusArea
	MethodIdx         int
	URLInput          textinput.Model
	Body              string
	Response          string
	ResponseHeaders   string
	StatusCode        string
	Loading           bool
	MethodOpen        bool
	Width             int
	Height            int
	Viewport          viewport.Model
	ViewportReady     bool
	Fullscreen        bool
	CurrentView       ResponseView
	ShowHelp          bool
	ShowHeadersForm   bool
	RequestHeaders    []HeaderPair
	HeaderKeyInput    textinput.Model
	HeaderValInput    textinput.Model
	HeaderSelectedIdx int
	HeaderFormMode    HeaderFormMode
	HeaderFocusField  int
	HeaderIsEditing   bool
	ShowCurlImport    bool
	CurlInput         textinput.Model
	HelpViewport      viewport.Model
}

// ResponseMsg represents the message returned from an HTTP request
type ResponseMsg struct {
	Resp    string
	Headers string
	Status  string
	Err     error
}

// InitialModel creates and returns a new model with default values
func InitialModel() Model {
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

	curlInput := textinput.New()
	curlInput.Placeholder = "Paste curl command here..."
	curlInput.CharLimit = 2000
	curlInput.Width = 80

	vp := viewport.New(0, 0)
	helpVp := viewport.New(0, 0)
	vp.KeyMap = viewport.KeyMap{} // Disable default keybindings

	return Model{
		Focus:             FocusMethod,
		MethodIdx:         0,
		URLInput:          ti,
		Width:             100,
		Height:            40,
		MethodOpen:        false,
		Viewport:          vp,
		CurrentView:       ViewBody,
		RequestHeaders:    []HeaderPair{},
		HeaderKeyInput:    headerKey,
		HeaderValInput:    headerVal,
		HeaderSelectedIdx: 0,
		HeaderFormMode:    HeaderModeList,
		HeaderFocusField:  0,
		CurlInput:         curlInput,
		HelpViewport:      helpVp,
	}
}

// Init implements tea.Model interface
func (m Model) Init() tea.Cmd {
	return nil
}
