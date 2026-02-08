package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tbourrel/apitty/internal/json"
	"github.com/tbourrel/apitty/internal/model"
)

// SendRequestCmd performs the HTTP request in a goroutine and returns a tea.Cmd
func SendRequestCmd(method, url string, headers []model.HeaderPair, body string) tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 15 * time.Second}
		var reqBody io.Reader
		if method == "POST" || method == "PUT" || method == "PATCH" {
			reqBody = strings.NewReader(body)
		}
		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return model.ResponseMsg{Resp: "", Headers: "", Status: "", Err: err}
		}
		// Apply request headers
		for _, h := range headers {
			if h.Key != "" {
				req.Header.Set(h.Key, h.Value)
			}
		}
		resp, err := client.Do(req)
		if err != nil {
			return model.ResponseMsg{Resp: "", Headers: "", Status: "", Err: err}
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		// Build headers string
		var headersBuilder strings.Builder
		for k, v := range resp.Header {
			headersBuilder.WriteString(fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", ")))
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return model.ResponseMsg{Resp: "", Headers: "", Status: "", Err: err}
		}
		pretty := json.TryPrettyJSON(respBody)
		return model.ResponseMsg{Resp: pretty, Headers: headersBuilder.String(), Status: resp.Status, Err: nil}
	}
}
