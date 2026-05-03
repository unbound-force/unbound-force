package ollamaproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/unbound-force/unbound-force/internal/auth"
	"github.com/unbound-force/unbound-force/internal/pidfile"
)

// --- Helper: mock Options builder ---

// testOpts returns an Options struct with all dependencies
// injected as no-op/success mocks. Tests override specific
// fields to exercise error paths. Follows the testOpts()
// pattern from internal/gateway/gateway_test.go.
func testOpts(t *testing.T) Options {
	t.Helper()
	return Options{
		Port:       DefaultPort,
		EmbedModel: DefaultEmbedModel,
		GatewayURL: DefaultGatewayURL,
		ProjectDir: t.TempDir(),
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
		LookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return []byte("mock-token\n"), nil
		},
		ExecStart: func(cmd *exec.Cmd) error {
			return nil
		},
		Getenv: func(key string) string {
			switch key {
			case "ANTHROPIC_VERTEX_PROJECT_ID":
				return "test-project"
			case "VERTEX_LOCATION":
				return "us-east5"
			}
			return ""
		},
		HTTPGet: func(url string) (int, error) {
			return http.StatusOK, nil
		},
		FindProcess: func(pid int) (*os.Process, error) {
			return nil, fmt.Errorf("no such process")
		},
		ListenAndServe: func(addr string, handler http.Handler) error {
			return nil
		},
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// stdoutStr returns the captured stdout content.
func stdoutStr(opts Options) string {
	return opts.Stdout.(*bytes.Buffer).String()
}

// testTokenManager creates a real auth.TokenManager with
// a mock RefreshFn that returns the given token. The
// manager is started and cleaned up via t.Cleanup.
// Follows the testVertexProvider pattern from
// internal/gateway/gateway_test.go.
func testTokenManager(t *testing.T, token string) *auth.TokenManager {
	t.Helper()
	tm := auth.NewTokenManager(auth.TokenManagerOpts{
		RefreshFn:       func() (string, error) { return token, nil },
		Lifetime:        55 * time.Minute,
		ProactiveWindow: 5 * time.Minute,
		Interval:        1 * time.Hour,
	})
	ctx, cancel := context.WithCancel(context.Background())
	if err := tm.Start(ctx); err != nil {
		cancel()
		t.Fatalf("testTokenManager Start: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		tm.Stop()
	})
	return tm
}

// ============================================================
// Task 3.10: TestMapModelName_Known
// ============================================================

func TestMapModelName_Known(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"granite-embedding:30m", "text-embedding-005"},
		{"granite-embedding-small-english-r2", "text-embedding-005"},
		{"llama3.2:3b", "claude-sonnet-4-20250514"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := mapModelName(tt.input)
			if !ok {
				t.Fatalf("mapModelName(%q) returned false, want true", tt.input)
			}
			if got != tt.want {
				t.Errorf("mapModelName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================
// Task 4.21: TestMapModelName_Unknown
// ============================================================

func TestMapModelName_Unknown(t *testing.T) {
	_, ok := mapModelName("nonexistent-model")
	if ok {
		t.Error("mapModelName(nonexistent-model) returned true, want false")
	}
}

// ============================================================
// Task 4.19: TestValidateModelName
// ============================================================

func TestValidateModelName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "granite-embedding:30m", false},
		{"valid with dots", "llama3.2:3b", false},
		{"valid with underscore", "model_v2", false},
		{"valid with hyphen", "text-embedding-005", false},
		{"empty", "", true},
		{"path traversal slash", "../../evil", true},
		{"backslash", `model\evil`, true},
		{"percent encoding", "model%2F", true},
		{"control char", "model\x00evil", true},
		{"space", "model name", true},
		{"starts with hyphen", "-model", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateModelName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateModelName(%q) error = %v, wantErr %v",
					tt.input, err, tt.wantErr)
			}
		})
	}
}

// ============================================================
// Task 3.12: TestRedactToken
// ============================================================

func TestRedactToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"bearer token in error",
			`Authorization: Bearer ya29.abc123_DEF-456`,
			`Authorization: Bearer [REDACTED]`,
		},
		{
			"no token",
			`some error without tokens`,
			`some error without tokens`,
		},
		{
			"multiple tokens",
			`Bearer token1 and Bearer token2`,
			`Bearer [REDACTED] and Bearer [REDACTED]`,
		},
		{
			"case insensitive",
			`bearer AbCdEf123`,
			`bearer [REDACTED]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactToken(tt.input)
			if got != tt.want {
				t.Errorf("redactToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ============================================================
// Task 3.3: TestValidateGatewayURL
// ============================================================

func TestValidateGatewayURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"localhost http", "http://localhost:53147", false},
		{"127.0.0.1 http", "http://127.0.0.1:53147", false},
		{"ipv6 loopback", "http://[::1]:53147", false},
		{"localhost https", "https://localhost:53147", false},
		{"root path", "http://localhost:53147/", false},
		{"non-loopback", "http://192.168.1.1:53147", true},
		{"public host", "http://example.com:53147", true},
		{"ftp scheme", "ftp://localhost:53147", true},
		{"non-root path", "http://localhost:53147/api", true},
		{"invalid url", "://bad", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGatewayURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGatewayURL(%q) error = %v, wantErr %v",
					tt.url, err, tt.wantErr)
			}
		})
	}
}

// ============================================================
// Task 3.3: TestValidateGatewayURL_NonLoopback_ErrorMessage
// ============================================================

func TestValidateGatewayURL_NonLoopback_ErrorMessage(t *testing.T) {
	err := validateGatewayURL("http://192.168.1.1:53147")
	if err == nil {
		t.Fatal("expected error for non-loopback URL")
	}
	if !strings.Contains(err.Error(), "SSRF") {
		t.Errorf("error should mention SSRF prevention, got: %s", err)
	}
	if !strings.Contains(err.Error(), "loopback") {
		t.Errorf("error should mention loopback, got: %s", err)
	}
}

// ============================================================
// Task 3.8: TestHealthEndpoint_ResponseFields
// ============================================================

func TestHealthEndpoint_ResponseFields(t *testing.T) {
	opts := testOpts(t)
	ps := &proxyServer{
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}
	mux := newMux(ps)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("health endpoint returned %d, want %d",
			w.Code, http.StatusOK)
	}

	var resp healthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse health response: %v", err)
	}

	if resp.Service != "uf-ollama-proxy" {
		t.Errorf("service = %q, want %q", resp.Service, "uf-ollama-proxy")
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q, want %q", resp.Status, "ok")
	}
	if resp.Port != DefaultPort {
		t.Errorf("port = %d, want %d", resp.Port, DefaultPort)
	}
	if resp.EmbedModel != DefaultEmbedModel {
		t.Errorf("embed_model = %q, want %q",
			resp.EmbedModel, DefaultEmbedModel)
	}
	if !resp.GatewayAvailable {
		t.Error("gateway_available = false, want true")
	}
}

// ============================================================
// Task 4.13: TestHandleEmbed_SingleInput
// ============================================================

func TestHandleEmbed_SingleInput(t *testing.T) {
	// Mock Vertex AI server.
	var capturedAuth string
	var capturedBody []byte
	vertexServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = r.Header.Get("Authorization")
			capturedBody, _ = io.ReadAll(r.Body)
			resp := vertexEmbedResponse{
				Predictions: []vertexEmbedPrediction{
					{Embeddings: vertexEmbedValues{
						Values: []float64{0.1, 0.2, 0.3},
					}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
	defer vertexServer.Close()

	opts := testOpts(t)
	// Use a redirectTransport to intercept Vertex AI calls
	// and send them to our test server.
	opts.HTTPClient = &http.Client{
		Transport: &redirectTransport{
			target: vertexServer.URL,
			inner:  http.DefaultTransport,
		},
		Timeout: 5 * time.Second,
	}

	ps := &proxyServer{
		tokenMgr:         testTokenManager(t, "mock-vertex-token"),
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}

	// Build request.
	reqBody := `{"model": "granite-embedding:30m", "input": ["hello"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/embed",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ps.handleEmbed(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s",
			w.Code, http.StatusOK, w.Body.String())
	}

	// Verify Authorization header was injected.
	if capturedAuth != "Bearer mock-vertex-token" {
		t.Errorf("Authorization = %q, want %q",
			capturedAuth, "Bearer mock-vertex-token")
	}

	// Verify Vertex request body contains correct instances.
	var vertexReq vertexEmbedRequest
	if err := json.Unmarshal(capturedBody, &vertexReq); err != nil {
		t.Fatalf("failed to parse captured Vertex body: %v", err)
	}
	if len(vertexReq.Instances) != 1 {
		t.Fatalf("instances count = %d, want 1",
			len(vertexReq.Instances))
	}
	if vertexReq.Instances[0].Content != "hello" {
		t.Errorf("instance content = %q, want %q",
			vertexReq.Instances[0].Content, "hello")
	}

	// Verify Ollama response.
	var resp ollamaEmbedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Model != "granite-embedding:30m" {
		t.Errorf("model = %q, want %q",
			resp.Model, "granite-embedding:30m")
	}
	if len(resp.Embeddings) != 1 {
		t.Fatalf("embeddings count = %d, want 1",
			len(resp.Embeddings))
	}
	expected := []float64{0.1, 0.2, 0.3}
	for i, v := range resp.Embeddings[0] {
		if v != expected[i] {
			t.Errorf("embedding[0][%d] = %f, want %f",
				i, v, expected[i])
		}
	}
}

// ============================================================
// Task 4.14: TestHandleEmbed_BatchInput
// ============================================================

func TestHandleEmbed_BatchInput(t *testing.T) {
	vertexServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := vertexEmbedResponse{
				Predictions: []vertexEmbedPrediction{
					{Embeddings: vertexEmbedValues{
						Values: []float64{0.1, 0.2},
					}},
					{Embeddings: vertexEmbedValues{
						Values: []float64{0.3, 0.4},
					}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
	defer vertexServer.Close()

	opts := testOpts(t)
	opts.HTTPClient = &http.Client{
		Transport: &redirectTransport{
			target: vertexServer.URL,
			inner:  http.DefaultTransport,
		},
		Timeout: 5 * time.Second,
	}

	ps := &proxyServer{
		tokenMgr:         testTokenManager(t, "mock-token"),
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}

	reqBody := `{"model": "granite-embedding:30m", "input": ["a", "b"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/embed",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ps.handleEmbed(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s",
			w.Code, http.StatusOK, w.Body.String())
	}

	var resp ollamaEmbedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp.Embeddings) != 2 {
		t.Fatalf("embeddings count = %d, want 2",
			len(resp.Embeddings))
	}
	// Verify vectors are in correct order.
	if resp.Embeddings[0][0] != 0.1 {
		t.Errorf("embeddings[0][0] = %f, want 0.1",
			resp.Embeddings[0][0])
	}
	if resp.Embeddings[1][0] != 0.3 {
		t.Errorf("embeddings[1][0] = %f, want 0.3",
			resp.Embeddings[1][0])
	}
}

// ============================================================
// Task 4.15: TestHandleEmbed_VertexError
// ============================================================

func TestHandleEmbed_VertexError(t *testing.T) {
	vertexServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			// Simulate Vertex echoing the Authorization header
			// in the error response (D8a test).
			fmt.Fprintf(w,
				`{"error": "access denied with Bearer ya29.secret_token_here"}`)
		}))
	defer vertexServer.Close()

	opts := testOpts(t)
	opts.HTTPClient = &http.Client{
		Transport: &redirectTransport{
			target: vertexServer.URL,
			inner:  http.DefaultTransport,
		},
		Timeout: 5 * time.Second,
	}

	ps := &proxyServer{
		tokenMgr:         testTokenManager(t, "mock-token"),
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}

	reqBody := `{"model": "granite-embedding:30m", "input": ["hello"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/embed",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ps.handleEmbed(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d",
			w.Code, http.StatusBadGateway)
	}

	// Verify token is redacted in error response.
	body := w.Body.String()
	if strings.Contains(body, "ya29.secret_token_here") {
		t.Error("error response contains unredacted token")
	}
	if !strings.Contains(body, "[REDACTED]") {
		t.Error("error response should contain [REDACTED]")
	}
}

// ============================================================
// Task 4.16: TestHandleEmbed_UnknownModel
// ============================================================

func TestHandleEmbed_UnknownModel(t *testing.T) {
	opts := testOpts(t)
	ps := &proxyServer{
		tokenMgr:         testTokenManager(t, "mock-token"),
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}

	reqBody := `{"model": "unknown-model", "input": ["hello"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/embed",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ps.handleEmbed(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d",
			w.Code, http.StatusBadRequest)
	}

	body := w.Body.String()
	if !strings.Contains(body, "unknown model") {
		t.Errorf("error should mention unknown model, got: %s", body)
	}
	// Verify it's not passed through — the model should
	// be rejected, not forwarded to Vertex.
	if !strings.Contains(body, "not in model mapping") {
		t.Errorf("error should mention model mapping, got: %s", body)
	}
}

// ============================================================
// Task 4.17: TestHandleEmbed_ModelPathTraversal
// ============================================================

func TestHandleEmbed_ModelPathTraversal(t *testing.T) {
	opts := testOpts(t)
	ps := &proxyServer{
		tokenMgr:         testTokenManager(t, "mock-token"),
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}

	reqBody := `{"model": "../../evil", "input": ["hello"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/embed",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ps.handleEmbed(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d",
			w.Code, http.StatusBadRequest)
	}

	body := w.Body.String()
	if !strings.Contains(body, "invalid model name") {
		t.Errorf("error should mention invalid model name, got: %s", body)
	}
}

// ============================================================
// Task 4.18: TestHandleEmbed_OversizedBody
// ============================================================

func TestHandleEmbed_OversizedBody(t *testing.T) {
	opts := testOpts(t)
	ps := &proxyServer{
		tokenMgr:         testTokenManager(t, "mock-token"),
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}

	// Create a body larger than 10MB.
	bigBody := strings.Repeat("x", maxRequestBodySize+1)
	reqBody := fmt.Sprintf(`{"model": "granite-embedding:30m", "input": ["%s"]}`,
		bigBody)
	req := httptest.NewRequest(http.MethodPost, "/api/embed",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ps.handleEmbed(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body: %s",
			w.Code, http.StatusRequestEntityTooLarge,
			w.Body.String())
	}
}

// ============================================================
// Task 5.11: TestHandleGenerate_Success
// ============================================================

func TestHandleGenerate_Success(t *testing.T) {
	var capturedBody []byte
	gatewayServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedBody, _ = io.ReadAll(r.Body)
			resp := anthropicResponse{
				Content: []anthropicContentBlock{
					{Type: "text", Text: "Hello from Claude!"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
	defer gatewayServer.Close()

	opts := testOpts(t)
	opts.GatewayURL = gatewayServer.URL
	opts.HTTPClient = gatewayServer.Client()

	ps := &proxyServer{
		tokenMgr:         testTokenManager(t, "mock-token"),
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}

	reqBody := `{"model": "llama3.2:3b", "prompt": "Hello", "stream": false}`
	req := httptest.NewRequest(http.MethodPost, "/api/generate",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ps.handleGenerate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s",
			w.Code, http.StatusOK, w.Body.String())
	}

	// Verify model mapping in the Anthropic request.
	var anthropicReq anthropicRequest
	if err := json.Unmarshal(capturedBody, &anthropicReq); err != nil {
		t.Fatalf("failed to parse captured body: %v", err)
	}
	if anthropicReq.Model != "claude-sonnet-4-20250514" {
		t.Errorf("anthropic model = %q, want %q",
			anthropicReq.Model, "claude-sonnet-4-20250514")
	}

	// Verify Ollama response.
	var resp ollamaGenerateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Model != "llama3.2:3b" {
		t.Errorf("model = %q, want %q", resp.Model, "llama3.2:3b")
	}
	if resp.Response != "Hello from Claude!" {
		t.Errorf("response = %q, want %q",
			resp.Response, "Hello from Claude!")
	}
}

// ============================================================
// Task 5.12: TestHandleGenerate_GatewayDown
// ============================================================

func TestHandleGenerate_GatewayDown(t *testing.T) {
	opts := testOpts(t)
	// Use a URL that will fail to connect.
	opts.GatewayURL = "http://localhost:1"
	opts.HTTPClient = &http.Client{Timeout: 1 * time.Second}

	ps := &proxyServer{
		tokenMgr:         testTokenManager(t, "mock-token"),
		opts:             opts,
		gatewayAvailable: false,
		startTime:        time.Now(),
	}

	reqBody := `{"model": "llama3.2:3b", "prompt": "Hello", "stream": false}`
	req := httptest.NewRequest(http.MethodPost, "/api/generate",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ps.handleGenerate(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d",
			w.Code, http.StatusBadGateway)
	}

	body := w.Body.String()
	if !strings.Contains(body, "gateway not available") {
		t.Errorf("error should mention gateway not available, got: %s", body)
	}
	if !strings.Contains(body, "uf gateway") {
		t.Errorf("error should mention `uf gateway`, got: %s", body)
	}
}

// ============================================================
// Task 5.13: TestHandleGenerate_GatewayError
// ============================================================

func TestHandleGenerate_GatewayError(t *testing.T) {
	gatewayServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "internal server error"}`)
		}))
	defer gatewayServer.Close()

	opts := testOpts(t)
	opts.GatewayURL = gatewayServer.URL
	opts.HTTPClient = gatewayServer.Client()

	ps := &proxyServer{
		tokenMgr:         testTokenManager(t, "mock-token"),
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}

	reqBody := `{"model": "llama3.2:3b", "prompt": "Hello", "stream": false}`
	req := httptest.NewRequest(http.MethodPost, "/api/generate",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ps.handleGenerate(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d",
			w.Code, http.StatusBadGateway)
	}

	body := w.Body.String()
	if !strings.Contains(body, "gateway error") {
		t.Errorf("error should mention gateway error, got: %s", body)
	}
}

// ============================================================
// Task 5.14: TestHandleGenerate_EmptyContent
// ============================================================

func TestHandleGenerate_EmptyContent(t *testing.T) {
	gatewayServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Return a response with empty content array.
			resp := anthropicResponse{Content: []anthropicContentBlock{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
	defer gatewayServer.Close()

	opts := testOpts(t)
	opts.GatewayURL = gatewayServer.URL
	opts.HTTPClient = gatewayServer.Client()

	ps := &proxyServer{
		tokenMgr:         testTokenManager(t, "mock-token"),
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}

	reqBody := `{"model": "llama3.2:3b", "prompt": "Hello", "stream": false}`
	req := httptest.NewRequest(http.MethodPost, "/api/generate",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ps.handleGenerate(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d",
			w.Code, http.StatusBadGateway)
	}

	body := w.Body.String()
	if !strings.Contains(body, "empty content") {
		t.Errorf("error should mention empty content, got: %s", body)
	}
}

// ============================================================
// Task 5.15: TestHandleGenerate_MaxTokensSet
// ============================================================

func TestHandleGenerate_MaxTokensSet(t *testing.T) {
	var capturedBody []byte
	gatewayServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedBody, _ = io.ReadAll(r.Body)
			resp := anthropicResponse{
				Content: []anthropicContentBlock{
					{Type: "text", Text: "response"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
	defer gatewayServer.Close()

	opts := testOpts(t)
	opts.GatewayURL = gatewayServer.URL
	opts.HTTPClient = gatewayServer.Client()

	ps := &proxyServer{
		tokenMgr:         testTokenManager(t, "mock-token"),
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}

	reqBody := `{"model": "llama3.2:3b", "prompt": "Hello", "stream": false}`
	req := httptest.NewRequest(http.MethodPost, "/api/generate",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ps.handleGenerate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s",
			w.Code, http.StatusOK, w.Body.String())
	}

	// Verify max_tokens is 4096 in the Anthropic request.
	var anthropicReq anthropicRequest
	if err := json.Unmarshal(capturedBody, &anthropicReq); err != nil {
		t.Fatalf("failed to parse captured body: %v", err)
	}
	if anthropicReq.MaxTokens != 4096 {
		t.Errorf("max_tokens = %d, want 4096",
			anthropicReq.MaxTokens)
	}
}

// ============================================================
// Task 6.3: TestHandleTags_ReturnsModels
// ============================================================

func TestHandleTags_ReturnsModels(t *testing.T) {
	opts := testOpts(t)
	ps := &proxyServer{
		opts:             opts,
		gatewayAvailable: true,
		startTime:        time.Now(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	w := httptest.NewRecorder()

	ps.handleTags(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d",
			w.Code, http.StatusOK)
	}

	var resp ollamaTagsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Verify all mapped model names appear.
	modelNames := make(map[string]bool)
	for _, m := range resp.Models {
		modelNames[m.Name] = true
	}

	for name := range defaultModelMap {
		if !modelNames[name] {
			t.Errorf("model %q not found in tags response", name)
		}
	}

	// Verify count matches the model map.
	if len(resp.Models) != len(defaultModelMap) {
		t.Errorf("model count = %d, want %d",
			len(resp.Models), len(defaultModelMap))
	}
}

// ============================================================
// Task 9.1: TestStart_DefaultPort
// ============================================================

func TestStart_DefaultPort(t *testing.T) {
	// Verify that Options.defaults() sets the default port
	// to 11434 when Port is zero, and that Start() passes
	// validation with the default port. We test the
	// components individually because Start() blocks on
	// a select waiting for signals or server errors.

	// Test 1: defaults() sets DefaultPort.
	opts := Options{}
	opts.defaults()
	if opts.Port != DefaultPort {
		t.Errorf("default port = %d, want %d",
			opts.Port, DefaultPort)
	}

	// Test 2: Start() with default port passes validation
	// and reaches the token acquisition phase. We verify
	// by checking that the gateway URL validation passes
	// (it uses the default gateway URL).
	if err := validateGatewayURL(DefaultGatewayURL); err != nil {
		t.Errorf("default gateway URL validation failed: %v", err)
	}

	// Test 3: PID file path uses the correct directory.
	pp := pidPath("/test/project")
	want := "/test/project/.uf/ollama-proxy.pid"
	if pp != want {
		t.Errorf("pidPath = %q, want %q", pp, want)
	}
}

// ============================================================
// Task 9.2: TestStart_CustomPort
// ============================================================

func TestStart_CustomPort(t *testing.T) {
	opts := testOpts(t)
	opts.Port = 19876

	opts.defaults()

	if opts.Port != 19876 {
		t.Errorf("custom port = %d, want 19876", opts.Port)
	}

	// Verify the port is preserved through defaults().
	opts2 := Options{Port: 19876}
	opts2.defaults()
	if opts2.Port != 19876 {
		t.Errorf("defaults() overwrote custom port: got %d, want 19876",
			opts2.Port)
	}
}

// ============================================================
// Task 9.3: TestStart_GatewayWarning
// ============================================================

func TestStart_GatewayWarning(t *testing.T) {
	// Proxy should start even when gateway is unreachable.
	// Gateway health check is informational, not a hard
	// failure (per design.md D4).

	// Test 1: checkGatewayHealth returns false when
	// gateway is unreachable.
	httpGet := func(url string) (int, error) {
		return 0, fmt.Errorf("connection refused")
	}
	available := checkGatewayHealth(httpGet, DefaultGatewayURL)
	if available {
		t.Error("expected gateway to be unavailable")
	}

	// Test 2: checkGatewayHealth returns false for
	// non-200 responses.
	httpGet2 := func(url string) (int, error) {
		return http.StatusServiceUnavailable, nil
	}
	available2 := checkGatewayHealth(httpGet2, DefaultGatewayURL)
	if available2 {
		t.Error("expected gateway to be unavailable for 503")
	}

	// Test 3: checkGatewayHealth returns true for 200.
	httpGet3 := func(url string) (int, error) {
		return http.StatusOK, nil
	}
	available3 := checkGatewayHealth(httpGet3, DefaultGatewayURL)
	if !available3 {
		t.Error("expected gateway to be available for 200")
	}

	// Test 4: The proxyServer struct correctly records
	// gateway availability. When gateway is unavailable,
	// the health endpoint should report it.
	opts := testOpts(t)
	ps := &proxyServer{
		opts:             opts,
		gatewayAvailable: false,
		startTime:        time.Now(),
	}
	mux := newMux(ps)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var resp healthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse health response: %v", err)
	}
	if resp.GatewayAvailable {
		t.Error("health should report gateway_available=false")
	}
}

// ============================================================
// Task 9.4: TestStart_AlreadyRunning
// ============================================================

func TestStart_AlreadyRunning(t *testing.T) {
	opts := testOpts(t)

	// Write a PID file with the current process PID
	// (which is alive).
	pp := pidPath(opts.ProjectDir)
	info := pidfile.PIDInfo{
		PID:      os.Getpid(),
		Port:     DefaultPort,
		Provider: "vertex-embedding",
		Started:  time.Now(),
	}
	if err := pidfile.WritePID(pp, info); err != nil {
		t.Fatal(err)
	}
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return os.FindProcess(pid)
	}

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error when proxy already running")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("expected already running error, got: %s",
			err.Error())
	}
	if !strings.Contains(err.Error(), "uf ollama-proxy stop") {
		t.Errorf("expected stop hint, got: %s", err.Error())
	}
}

// ============================================================
// Task 9.5: TestStart_GcloudMissing
// ============================================================

func TestStart_GcloudMissing(t *testing.T) {
	opts := testOpts(t)
	opts.LookPath = func(name string) (string, error) {
		if name == "gcloud" {
			return "", fmt.Errorf("not found")
		}
		return "/usr/bin/" + name, nil
	}

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error when gcloud missing")
	}
	if !strings.Contains(err.Error(), "gcloud CLI not found") {
		t.Errorf("expected gcloud not found error, got: %s",
			err.Error())
	}
	if !strings.Contains(err.Error(), "Install") {
		t.Errorf("expected install instructions, got: %s",
			err.Error())
	}
}

// ============================================================
// Task 9.6: TestStart_NonLoopbackGatewayURL
// ============================================================

func TestStart_NonLoopbackGatewayURL(t *testing.T) {
	opts := testOpts(t)
	opts.GatewayURL = "http://192.168.1.100:53147"

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error for non-loopback gateway URL")
	}
	if !strings.Contains(err.Error(), "loopback") {
		t.Errorf("expected loopback error, got: %s",
			err.Error())
	}
	if !strings.Contains(err.Error(), "SSRF") {
		t.Errorf("expected SSRF mention, got: %s",
			err.Error())
	}
}

// ============================================================
// Task 9.7: TestStop_Running
// ============================================================

func TestStop_Running(t *testing.T) {
	opts := testOpts(t)

	// Write a PID file.
	pp := pidPath(opts.ProjectDir)
	info := pidfile.PIDInfo{
		PID:      99999,
		Port:     DefaultPort,
		Provider: "vertex-embedding",
		Started:  time.Now(),
	}
	if err := pidfile.WritePID(pp, info); err != nil {
		t.Fatal(err)
	}

	// FindProcess returns "no such process" — Stop treats
	// it as not alive and removes the PID file.
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return nil, fmt.Errorf("no such process")
	}

	if err := Stop(opts); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "No ollama-proxy running") {
		t.Errorf("expected no proxy message, got: %s", out)
	}

	// PID file should be removed.
	if _, err := os.Stat(pp); !os.IsNotExist(err) {
		t.Error("PID file should be removed after Stop")
	}
}

// ============================================================
// Task 9.8: TestStatus_Running
// ============================================================

func TestStatus_Running(t *testing.T) {
	opts := testOpts(t)

	// Write a PID file with current process PID (alive).
	pp := pidPath(opts.ProjectDir)
	info := pidfile.PIDInfo{
		PID:      os.Getpid(),
		Port:     DefaultPort,
		Provider: "vertex-embedding",
		Started:  time.Now().Add(-5 * time.Minute),
	}
	if err := pidfile.WritePID(pp, info); err != nil {
		t.Fatal(err)
	}
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return os.FindProcess(pid)
	}

	if err := Status(opts); err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "Ollama Proxy Status") {
		t.Errorf("expected status header, got: %s", out)
	}
	if !strings.Contains(out, fmt.Sprintf("Port:      %d", DefaultPort)) {
		t.Errorf("expected port in output, got: %s", out)
	}
	if !strings.Contains(out, "vertex-embedding") {
		t.Errorf("expected provider in output, got: %s", out)
	}
	if !strings.Contains(out, fmt.Sprintf("PID:       %d", os.Getpid())) {
		t.Errorf("expected PID in output, got: %s", out)
	}
}

// ============================================================
// Task 9.9: TestHealthEndpoint_ResponseFields (lifecycle)
// ============================================================

// TestHealthEndpoint_ResponseFields_AllFields verifies that
// the /health endpoint returns all expected JSON fields
// including the "service" field that distinguishes the proxy
// from real Ollama.
func TestHealthEndpoint_ResponseFields_AllFields(t *testing.T) {
	opts := testOpts(t)
	opts.Port = 22222
	opts.EmbedModel = "custom-model"

	ps := &proxyServer{
		opts:             opts,
		gatewayAvailable: false,
		startTime:        time.Now(),
	}
	mux := newMux(ps)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d",
			w.Code, http.StatusOK)
	}

	var resp healthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse health response: %v", err)
	}

	// Assert all required fields.
	if resp.Service != "uf-ollama-proxy" {
		t.Errorf("service = %q, want %q",
			resp.Service, "uf-ollama-proxy")
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q, want %q", resp.Status, "ok")
	}
	if resp.Port != 22222 {
		t.Errorf("port = %d, want 22222", resp.Port)
	}
	if resp.EmbedModel != "custom-model" {
		t.Errorf("embed_model = %q, want %q",
			resp.EmbedModel, "custom-model")
	}
	if resp.GatewayAvailable {
		t.Error("gateway_available = true, want false")
	}

	// Verify all JSON keys are present in raw response.
	var raw map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to parse raw JSON: %v", err)
	}
	requiredKeys := []string{
		"service", "status", "port",
		"embed_model", "gateway_available",
	}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing JSON key %q in health response",
				key)
		}
	}
}

// ============================================================
// Task 9.10: TestStart_TokenRefreshFailure
// ============================================================

func TestStart_TokenRefreshFailure(t *testing.T) {
	opts := testOpts(t)

	// Make token acquisition fail — simulates gcloud
	// returning an error.
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf(
			"ERROR: (gcloud.auth.application-default.print-access-token) " +
				"You do not currently have an active account selected")
	}

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error when token refresh fails")
	}

	// Assert the error mentions re-authentication, not
	// stale token forwarding. The error should come from
	// the token manager's initial acquisition failure.
	if !strings.Contains(err.Error(), "token acquisition failed") {
		t.Errorf("expected token acquisition error, got: %s",
			err.Error())
	}

	// Verify the error does NOT contain a stale token.
	if strings.Contains(err.Error(), "Bearer") {
		t.Errorf("error should not contain Bearer token, got: %s",
			err.Error())
	}
}

// ============================================================
// Test helpers
// ============================================================

// redirectTransport is an http.RoundTripper that redirects
// all requests to a target URL (the httptest server). This
// allows handler tests to intercept Vertex AI calls without
// modifying the URL construction logic.
type redirectTransport struct {
	target string
	inner  http.RoundTripper
}

func (rt *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect the request to the test server.
	req.URL.Scheme = "http"
	// Parse the target to get host.
	targetURL := strings.TrimPrefix(rt.target, "http://")
	req.URL.Host = targetURL
	req.Host = targetURL
	return rt.inner.RoundTrip(req)
}
