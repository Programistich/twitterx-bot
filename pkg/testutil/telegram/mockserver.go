package telegram

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

// Call represents a single Telegram Bot API request received by the MockServer.
//
// The request body is captured both as raw bytes and, when possible, decoded into JSON
// (for application/json requests) or form values (for form submissions).
type Call struct {
	Method string
	Token  string

	Header  http.Header
	RawBody []byte

	// JSON is set when the request payload is JSON (or can be decoded as JSON).
	JSON map[string]any

	// Form is set when the request payload is form-encoded or multipart.
	Form url.Values
}

// JSONString returns a string value from the decoded JSON payload.
// Supports basic "dot notation" for nested objects (e.g. "reply_parameters.message_id").
func (c Call) JSONString(path string) (string, bool) {
	v, ok := c.jsonGet(path)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if ok {
		return s, true
	}
	return "", false
}

// JSONInt64 returns an int64 value from the decoded JSON payload.
// Supports basic "dot notation" for nested objects (e.g. "reply_parameters.message_id").
func (c Call) JSONInt64(path string) (int64, bool) {
	v, ok := c.jsonGet(path)
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int64:
		return x, true
	case int:
		return int64(x), true
	case json.Number:
		i, err := x.Int64()
		return i, err == nil
	case string:
		i, err := strconv.ParseInt(x, 10, 64)
		return i, err == nil
	default:
		return 0, false
	}
}

func (c Call) jsonGet(path string) (any, bool) {
	if c.JSON == nil {
		return nil, false
	}
	parts := strings.Split(path, ".")
	var cur any = c.JSON
	for i, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[p]
		if !ok {
			return nil, false
		}
		if i < len(parts)-1 {
			if s, ok := cur.(string); ok {
				var obj map[string]any
				if err := json.Unmarshal([]byte(s), &obj); err == nil {
					cur = obj
				}
			}
		}
	}
	return cur, true
}

type telegramError struct {
	code        int
	description string
}

type httpError struct {
	statusCode int
	body       string
}

// MockServer is an HTTP-level mock for Telegram Bot API endpoints.
//
// It is designed for tests which want to exercise gotgbot serialization and error paths,
// rather than mocking Go interfaces.
type MockServer struct {
	srv *httptest.Server

	mu sync.Mutex

	calls map[string][]Call

	// responses stores full JSON response bodies (already wrapped as Telegram "ok/result" payloads).
	responses map[string][]byte
	tgErrors  map[string]telegramError
	httpErrs  map[string]httpError

	nextMessageID int64
}

// NewMockServer starts a new Telegram Bot API mock server.
func NewMockServer() *MockServer {
	ms := &MockServer{
		calls:         make(map[string][]Call),
		responses:     make(map[string][]byte),
		tgErrors:      make(map[string]telegramError),
		httpErrs:      make(map[string]httpError),
		nextMessageID: 1,
	}

	ms.srv = httptest.NewServer(http.HandlerFunc(ms.handle))
	return ms
}

// URL returns the base URL suitable for gotgbot.RequestOpts.APIURL.
func (m *MockServer) URL() string {
	if m == nil || m.srv == nil {
		return ""
	}
	return strings.TrimRight(m.srv.URL, "/")
}

// Close stops the underlying test server.
func (m *MockServer) Close() {
	if m == nil || m.srv == nil {
		return
	}
	m.srv.Close()
}

// SetResponse overrides the response for a specific Telegram API method.
//
// The provided value becomes the "result" field of a Telegram-style success response:
// {"ok": true, "result": ...}
func (m *MockServer) SetResponse(method string, result any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	method = strings.TrimSpace(method)
	if method == "" {
		return nil
	}

	payload := map[string]any{
		"ok":     true,
		"result": result,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	m.responses[method] = b
	return nil
}

// SetTelegramError configures a Telegram-style error response (HTTP 200 + ok=false).
func (m *MockServer) SetTelegramError(method string, code int, description string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tgErrors[method] = telegramError{code: code, description: description}
}

// SetHTTPError configures a non-200 HTTP error response.
func (m *MockServer) SetHTTPError(method string, statusCode int, body string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.httpErrs[method] = httpError{statusCode: statusCode, body: body}
}

// GetCalls returns all recorded calls for the given Telegram API method.
func (m *MockServer) GetCalls(method string) []Call {
	m.mu.Lock()
	defer m.mu.Unlock()

	got := m.calls[method]
	out := make([]Call, len(got))
	copy(out, got)
	return out
}

func (m *MockServer) handle(w http.ResponseWriter, r *http.Request) {
	token, method, ok := parseTelegramPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	body, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()

	call := Call{
		Method:  method,
		Token:   token,
		Header:  r.Header.Clone(),
		RawBody: body,
	}

	// Best-effort payload decoding for assertions.
	ct := r.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(ct, "application/json"):
		var obj map[string]any
		if err := json.Unmarshal(body, &obj); err == nil {
			call.JSON = obj
		}
	case strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
		// Parse from raw body (r.ParseForm would read from r.Body which is now closed).
		if vals, err := url.ParseQuery(string(body)); err == nil {
			call.Form = vals
		}
	case strings.HasPrefix(ct, "multipart/form-data"):
		// Rebuild request body so ParseMultipartForm can work.
		r2 := new(http.Request)
		*r2 = *r
		r2.Body = io.NopCloser(strings.NewReader(string(body)))
		_ = r2.ParseMultipartForm(32 << 20)
		if r2.MultipartForm != nil {
			call.Form = url.Values{}
			for k, v := range r2.MultipartForm.Value {
				call.Form[k] = append([]string(nil), v...)
			}
		}
	default:
		// Some gotgbot methods still send JSON without explicit content-type in edge cases.
		// Try decoding anyway.
		var obj map[string]any
		if err := json.Unmarshal(body, &obj); err == nil {
			call.JSON = obj
		}
	}

	m.mu.Lock()
	m.calls[method] = append(m.calls[method], call)

	// Priority: HTTP error -> Telegram error -> explicit response override -> default response.
	if herr, exists := m.httpErrs[method]; exists {
		m.mu.Unlock()
		w.WriteHeader(herr.statusCode)
		_, _ = w.Write([]byte(herr.body))
		return
	}
	if tgErr, exists := m.tgErrors[method]; exists {
		m.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":          false,
			"error_code":  tgErr.code,
			"description": tgErr.description,
		})
		return
	}
	if resp, exists := m.responses[method]; exists {
		m.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(resp)
		return
	}

	// Default response path.
	chatID := extractChatID(call)
	msgID := m.nextMessageID
	m.nextMessageID++
	m.mu.Unlock()

	switch method {
	case "getMe":
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true,
			"result": map[string]any{
				"id":         123456,
				"is_bot":     true,
				"first_name": "MockBot",
				"username":   "mock_bot",
			},
		})
		return
	case "sendMessage", "sendPhoto", "sendVideo":
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":     true,
			"result": defaultMessage(chatID, msgID),
		})
		return
	case "sendMediaGroup":
		n := extractMediaGroupSize(call)
		if n <= 0 {
			n = 1
		}
		msgs := make([]any, 0, n)
		for i := 0; i < n; i++ {
			msgs = append(msgs, defaultMessage(chatID, msgID+int64(i)))
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":     true,
			"result": msgs,
		})
		return
	default:
		// Many methods return bool true on success (deleteMessage, answerCallbackQuery, sendChatAction, etc.).
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":     true,
			"result": true,
		})
		return
	}
}

func parseTelegramPath(path string) (token string, method string, ok bool) {
	// Expected: /bot<TOKEN>/<method>
	p := strings.TrimPrefix(path, "/")
	parts := strings.Split(p, "/")
	if len(parts) < 2 {
		return "", "", false
	}
	if !strings.HasPrefix(parts[0], "bot") {
		return "", "", false
	}
	token = strings.TrimPrefix(parts[0], "bot")
	method = parts[1]
	if token == "" || method == "" {
		return "", "", false
	}
	return token, method, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func extractChatID(call Call) int64 {
	if call.JSON != nil {
		if v, ok := call.JSONInt64("chat_id"); ok {
			return v
		}
	}
	if call.Form != nil {
		if s := strings.TrimSpace(call.Form.Get("chat_id")); s != "" {
			if v, err := strconv.ParseInt(s, 10, 64); err == nil {
				return v
			}
		}
	}
	return 0
}

func extractMediaGroupSize(call Call) int {
	if call.JSON == nil {
		return 0
	}
	v, ok := call.JSON["media"]
	if !ok {
		return 0
	}
	arr, ok := v.([]any)
	if !ok {
		return 0
	}
	return len(arr)
}

func defaultMessage(chatID int64, messageID int64) map[string]any {
	// Minimal shape required for gotgbot to decode Message responses.
	// Keep it stable/deterministic for tests (no timestamps needed).
	return map[string]any{
		"message_id": messageID,
		"date":       0,
		"chat": map[string]any{
			"id":   chatID,
			"type": "private",
		},
		"text": "",
	}
}
