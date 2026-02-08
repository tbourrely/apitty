package parser

import (
	"strings"

	"github.com/tbourrel/apitty/internal/model"
)

// ParseCurlCommand parses a curl command and returns method, URL, and headers
func ParseCurlCommand(curlCmd string) (method string, url string, headers []model.HeaderPair) {
	// Remove leading "curl " and trim
	curlCmd = strings.TrimSpace(curlCmd)
	if strings.HasPrefix(curlCmd, "curl ") {
		curlCmd = strings.TrimSpace(curlCmd[5:])
	}

	// Parse using a simple state machine
	inQuote := false
	inSingleQuote := false
	var currentArg strings.Builder
	var args []string

	for i := 0; i < len(curlCmd); i++ {
		ch := curlCmd[i]

		switch {
		case ch == '"' && !inSingleQuote:
			inQuote = !inQuote
		case ch == '\'' && !inQuote:
			inSingleQuote = !inSingleQuote
		case ch == ' ' && !inQuote && !inSingleQuote:
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		case ch == '\\' && i+1 < len(curlCmd):
			// Handle escape sequences
			i++
			currentArg.WriteByte(curlCmd[i])
		default:
			currentArg.WriteByte(ch)
		}
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	// Parse arguments
	method = "GET"
	url = ""
	headerMap := make(map[string]string)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-X" || arg == "--request":
			if i+1 < len(args) {
				method = strings.ToUpper(args[i+1])
				i++
			}
		case arg == "-H" || arg == "--header":
			if i+1 < len(args) {
				headerStr := args[i+1]
				parts := strings.SplitN(headerStr, ":", 2)
				if len(parts) == 2 {
					headerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
				i++
			}
		case strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://"):
			url = arg
		case strings.HasPrefix(arg, "-"):
			// Skip other flags we don't handle
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++ // Skip the value too
			}
		default:
			// If we haven't found a URL yet and this doesn't start with -, it might be the URL
			if url == "" && !strings.HasPrefix(arg, "-") {
				url = arg
			}
		}
	}

	// Convert header map to slice
	headers = make([]model.HeaderPair, 0, len(headerMap))
	for key, val := range headerMap {
		headers = append(headers, model.HeaderPair{Key: key, Value: val})
	}

	return method, url, headers
}
