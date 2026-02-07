# Apitty 🐱

A fast and elegant terminal-based HTTP client with a beautiful TUI (Terminal User Interface). Built with Go and Bubble Tea.

## Features

✨ **Intuitive TUI** - Beautiful terminal interface with boxes and visual feedback  
🎯 **HTTP Methods** - Support for GET, POST, PUT, PATCH, and DELETE  
📝 **Request Headers** - Easy header management with a dedicated form  
📋 **cURL Import** - Import requests directly from cURL commands  
🎨 **Syntax Highlighting** - Colored JSON responses for better readability  
⌨️ **Vim Motions** - Navigate with h/j/k/l for a smooth experience  
🖱️ **Mouse Support** - Click and scroll through the interface  
📖 **Response Viewer** - Toggle between response body and headers  
🔍 **Fullscreen Mode** - Focus on responses with fullscreen view  
📜 **Scrollable Responses** - Smooth scrolling with text wrapping support

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/apitty.git
cd apitty

# Build the binary
go build -o apitty

# Run it
./apitty
```

## Quick Start

1. Launch the application: `./apitty`
2. Select an HTTP method with `j/k`
3. Navigate to the URL input with `Tab`
4. Type your URL
5. Press `Enter` to send the request
6. View the response in the box below

## Keybindings

### Global
- `?` - Toggle help manual
- `q` / `Ctrl+C` - Quit application
- `Ctrl+S` - Send HTTP request (from anywhere)
- `h` - Open headers form
- `i` - Import from cURL command

### Navigation
- `Tab` - Cycle forward through fields
- `Shift+Tab` - Cycle backward through fields

### Method Selector
- `j` / `↓` - Next HTTP method
- `k` / `↑` - Previous HTTP method

### URL Input
- `Enter` - Send HTTP request
- Type normally - all keys work

### Response Box
- `t` - Toggle between Body and Headers view
- `f` - Toggle fullscreen mode
- `j` / `↓` - Scroll down one line
- `k` / `↑` - Scroll up one line
- `d` - Scroll down half page
- `u` - Scroll up half page
- `g` - Jump to top
- `G` - Jump to bottom
- `w` - Toggle text wrapping

### Headers Form
- `j/k` - Navigate between headers
- `a` - Add new header
- `e` - Edit selected header
- `d` / `Backspace` - Delete selected header
- `Esc` - Close form

### cURL Import
- Type/paste cURL command
- `Enter` - Import and populate fields
- `Esc` - Cancel

## Usage Examples

### Simple GET Request
1. Select `GET` method
2. Enter URL: `https://api.github.com/users/octocat`
3. Press `Enter`

### POST with Headers
1. Select `POST` method
2. Enter URL: `https://httpbin.org/post`
3. Press `h` to open headers form
4. Add header: `Content-Type: application/json`
5. Press `Esc` to close form
6. Press `Enter` to send request

### Import from cURL
1. Press `i` to open import modal
2. Paste: `curl -X POST https://api.example.com -H "Authorization: Bearer token"`
3. Press `Enter`
4. All fields are populated automatically

## Features in Detail

### Response Viewer
- **Body View**: See the JSON response with syntax highlighting
- **Headers View**: Toggle with `t` to see response headers
- **Fullscreen Mode**: Press `f` for distraction-free viewing
- **Text Wrapping**: Toggle with `w` for long lines

### Syntax Highlighting
JSON responses are automatically colored:
- Keys in cyan
- String values in green
- Numbers in yellow
- Booleans in magenta
- Null in red

### Mouse Support
- Click to switch between fields
- Scroll wheel to navigate responses
- Click response box to focus

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

MIT License - See LICENSE file for details

## Author

Built with ❤️ using Go and Bubble Tea
