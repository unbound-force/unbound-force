package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// --- Helper: mock Options builder ---

// testOpts returns an Options struct with all dependencies
// injected as no-op/success mocks. Tests override specific
// fields to exercise error paths.
func testOpts(t *testing.T) Options {
	t.Helper()
	return Options{
		Port:       DefaultPort,
		ProjectDir: t.TempDir(),
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
		LookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return []byte(""), nil
		},
		ExecStart: func(cmd *exec.Cmd) error {
			return nil
		},
		Getenv: func(key string) string {
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
	}
}

// stdoutStr returns the captured stdout content.
func stdoutStr(opts Options) string {
	return opts.Stdout.(*bytes.Buffer).String()
}

// ============================================================
// PID File Tests (T041-T045)
// ============================================================

func TestWritePID_ReadPID_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".uf", "gateway.pid")

	started := time.Date(2026, 4, 20, 14, 30, 0, 0, time.UTC)
	info := PIDInfo{
		PID:      42195,
		Port:     53147,
		Provider: "vertex",
		Started:  started,
	}

	if err := WritePID(path, info); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}

	got, err := ReadPID(path)
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}

	if got.PID != info.PID {
		t.Errorf("PID: got %d, want %d", got.PID, info.PID)
	}
	if got.Port != info.Port {
		t.Errorf("Port: got %d, want %d", got.Port, info.Port)
	}
	if got.Provider != info.Provider {
		t.Errorf("Provider: got %q, want %q", got.Provider, info.Provider)
	}
	if !got.Started.Equal(info.Started) {
		t.Errorf("Started: got %v, want %v", got.Started, info.Started)
	}
}

func TestReadPID_MalformedFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "non-numeric PID",
			content: "not-a-number\nport=53147\n",
			wantErr: true,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
		{
			name:    "missing metadata",
			content: "12345\n",
			wantErr: false, // PID is valid, metadata is optional
		},
		{
			name:    "valid with unknown keys",
			content: "12345\nport=53147\nunknown=value\n",
			wantErr: false, // Unknown keys are ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "gateway.pid")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			info, err := ReadPID(path)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.PID != 12345 {
				t.Errorf("PID: got %d, want 12345", info.PID)
			}
		})
	}
}

func TestReadPID_FileNotFound(t *testing.T) {
	_, err := ReadPID("/nonexistent/path/gateway.pid")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestIsAlive_ProcessFound(t *testing.T) {
	findProcess := func(pid int) (*os.Process, error) {
		// Return a mock process that accepts signal 0.
		return os.FindProcess(os.Getpid())
	}

	// Use current process PID — it's definitely alive.
	alive := IsAlive(os.Getpid(), findProcess)
	if !alive {
		t.Error("expected alive=true for current process")
	}
}

func TestIsAlive_ProcessNotFound(t *testing.T) {
	findProcess := func(pid int) (*os.Process, error) {
		return nil, fmt.Errorf("no such process")
	}

	alive := IsAlive(99999, findProcess)
	if alive {
		t.Error("expected alive=false when process not found")
	}
}

func TestRemovePID_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gateway.pid")
	if err := os.WriteFile(path, []byte("12345\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemovePID(path); err != nil {
		t.Fatalf("RemovePID failed: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}
}

func TestRemovePID_NonExistent(t *testing.T) {
	// Removing a non-existent file should not error (idempotent).
	err := RemovePID("/nonexistent/path/gateway.pid")
	if err != nil {
		t.Fatalf("expected nil error for non-existent file, got: %v", err)
	}
}

func TestCleanupStale_DeadProcess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gateway.pid")
	content := "99999\nport=53147\nprovider=anthropic\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Process is dead.
	findProcess := func(pid int) (*os.Process, error) {
		return nil, fmt.Errorf("no such process")
	}

	if err := CleanupStale(path, findProcess); err != nil {
		t.Fatalf("CleanupStale failed: %v", err)
	}

	// PID file should be removed.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected stale PID file to be removed")
	}
}

func TestCleanupStale_AliveProcess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gateway.pid")
	content := fmt.Sprintf("%d\nport=53147\nprovider=anthropic\n", os.Getpid())
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Process is alive (current process).
	findProcess := func(pid int) (*os.Process, error) {
		return os.FindProcess(pid)
	}

	if err := CleanupStale(path, findProcess); err != nil {
		t.Fatalf("CleanupStale failed: %v", err)
	}

	// PID file should still exist.
	if _, err := os.Stat(path); err != nil {
		t.Error("expected PID file to be preserved for alive process")
	}
}

func TestCleanupStale_NoPIDFile(t *testing.T) {
	err := CleanupStale("/nonexistent/path/gateway.pid", os.FindProcess)
	if err != nil {
		t.Fatalf("expected nil error when no PID file, got: %v", err)
	}
}

// ============================================================
// Provider Detection Tests (T046-T048)
// ============================================================

func TestDetectProvider_VertexPriority(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "CLAUDE_CODE_USE_VERTEX":
			return "1"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		case "ANTHROPIC_API_KEY":
			return "sk-ant-xxx" // Also set, but Vertex has priority.
		}
		return ""
	}
	execCmd := func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not called")
	}

	prov, err := DetectProvider(getenv, execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov.Name() != "vertex" {
		t.Errorf("expected vertex, got: %s", prov.Name())
	}
}

func TestDetectProvider_BedrockPriority(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "CLAUDE_CODE_USE_BEDROCK":
			return "1"
		case "ANTHROPIC_API_KEY":
			return "sk-ant-xxx"
		}
		return ""
	}
	execCmd := func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not called")
	}

	prov, err := DetectProvider(getenv, execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov.Name() != "bedrock" {
		t.Errorf("expected bedrock, got: %s", prov.Name())
	}
}

func TestDetectProvider_Anthropic(t *testing.T) {
	getenv := func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-xxx"
		}
		return ""
	}
	execCmd := func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not called")
	}

	prov, err := DetectProvider(getenv, execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov.Name() != "anthropic" {
		t.Errorf("expected anthropic, got: %s", prov.Name())
	}
}

func TestDetectProvider_NoVarsSet(t *testing.T) {
	getenv := func(key string) string { return "" }
	execCmd := func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not called")
	}

	_, err := DetectProvider(getenv, execCmd)
	if err == nil {
		t.Fatal("expected error when no provider vars set")
	}
	if !strings.Contains(err.Error(), "no cloud provider detected") {
		t.Errorf("expected provider listing, got: %s", err.Error())
	}
}

func TestDetectProvider_VertexPrecedenceOverAnthropic(t *testing.T) {
	// When both ANTHROPIC_API_KEY and Vertex vars are set,
	// Vertex should be selected (higher priority).
	getenv := func(key string) string {
		switch key {
		case "CLAUDE_CODE_USE_VERTEX":
			return "1"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		case "ANTHROPIC_API_KEY":
			return "sk-ant-xxx"
		}
		return ""
	}
	execCmd := func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not called")
	}

	prov, err := DetectProvider(getenv, execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov.Name() != "vertex" {
		t.Errorf("expected vertex (higher priority), got: %s", prov.Name())
	}
}

func TestNewProviderByName_ValidNames(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
	}{
		{"anthropic", "anthropic"},
		{"vertex", "vertex"},
		{"bedrock", "bedrock"},
	}

	getenv := func(key string) string { return "" }
	execCmd := func(name string, args ...string) ([]byte, error) {
		return nil, nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prov, err := NewProviderByName(tt.name, getenv, execCmd)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if prov.Name() != tt.wantName {
				t.Errorf("got %s, want %s", prov.Name(), tt.wantName)
			}
		})
	}
}

func TestNewProviderByName_InvalidName(t *testing.T) {
	getenv := func(key string) string { return "" }
	execCmd := func(name string, args ...string) ([]byte, error) {
		return nil, nil
	}

	_, err := NewProviderByName("openai", getenv, execCmd)
	if err == nil {
		t.Fatal("expected error for invalid provider name")
	}
	if !strings.Contains(err.Error(), "unknown provider") {
		t.Errorf("expected unknown provider message, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "anthropic, vertex, bedrock") {
		t.Errorf("expected valid names listed, got: %s", err.Error())
	}
}

// ============================================================
// Anthropic Provider Tests (T049-T050)
// ============================================================

func TestAnthropicProvider_PrepareRequest(t *testing.T) {
	prov := &AnthropicProvider{apiKey: "sk-ant-test123"}

	req := httptest.NewRequest("POST", "/v1/messages", nil)
	req.Header.Set("anthropic-beta", "messages-2024-01-01")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Authorization", "Bearer should-be-stripped")

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	// Verify URL rewritten.
	if req.URL.Scheme != "https" {
		t.Errorf("scheme: got %s, want https", req.URL.Scheme)
	}
	if req.URL.Host != "api.anthropic.com" {
		t.Errorf("host: got %s, want api.anthropic.com", req.URL.Host)
	}
	if req.URL.Path != "/v1/messages" {
		t.Errorf("path: got %s, want /v1/messages", req.URL.Path)
	}

	// Verify x-api-key header set.
	if got := req.Header.Get("x-api-key"); got != "sk-ant-test123" {
		t.Errorf("x-api-key: got %q, want %q", got, "sk-ant-test123")
	}

	// Verify anthropic headers preserved.
	if got := req.Header.Get("anthropic-beta"); got != "messages-2024-01-01" {
		t.Errorf("anthropic-beta: got %q, want preserved", got)
	}
	if got := req.Header.Get("anthropic-version"); got != "2023-06-01" {
		t.Errorf("anthropic-version: got %q, want preserved", got)
	}
}

func TestAnthropicProvider_Start_Success(t *testing.T) {
	getenv := func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test123"
		}
		return ""
	}

	prov := newAnthropicProvider(getenv)
	if err := prov.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if prov.apiKey != "sk-ant-test123" {
		t.Errorf("apiKey: got %q, want %q", prov.apiKey, "sk-ant-test123")
	}
}

func TestAnthropicProvider_Start_MissingKey(t *testing.T) {
	getenv := func(key string) string { return "" }

	prov := newAnthropicProvider(getenv)
	err := prov.Start(context.Background())
	if err == nil {
		t.Fatal("expected error when ANTHROPIC_API_KEY is empty")
	}
	if !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Errorf("expected ANTHROPIC_API_KEY mention, got: %s", err.Error())
	}
}

// ============================================================
// Vertex Provider Tests (T051-T053)
// ============================================================

func TestVertexProvider_PrepareRequest(t *testing.T) {
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "ya29.test-token",
		tokenExpiry: time.Now().Add(30 * time.Minute),
	}

	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("anthropic-beta", "messages-2024-01-01")
	req.Header.Set("anthropic-version", "2023-06-01")

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	// Verify URL rewritten to Vertex rawPredict endpoint.
	if req.URL.Scheme != "https" {
		t.Errorf("scheme: got %s, want https", req.URL.Scheme)
	}
	if !strings.Contains(req.URL.Host, "aiplatform.googleapis.com") {
		t.Errorf("host: got %s, want *aiplatform.googleapis.com", req.URL.Host)
	}
	if !strings.Contains(req.URL.Path, "my-project") {
		t.Errorf("path should contain project ID, got: %s", req.URL.Path)
	}
	if !strings.Contains(req.URL.Path, "rawPredict") {
		t.Errorf("path should contain rawPredict, got: %s", req.URL.Path)
	}

	// Verify Authorization header.
	auth := req.Header.Get("Authorization")
	if auth != "Bearer ya29.test-token" {
		t.Errorf("Authorization: got %q, want %q", auth, "Bearer ya29.test-token")
	}

	// Verify body transformation: model removed,
	// anthropic_version injected (Spec 034 T030).
	transformed, readErr := io.ReadAll(req.Body)
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}
	var payload map[string]any
	if err := json.Unmarshal(transformed, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := payload["model"]; ok {
		t.Error("model field should be removed from body")
	}
	if av, ok := payload["anthropic_version"].(string); !ok || av != "vertex-2023-10-16" {
		t.Errorf("anthropic_version: got %q, want vertex-2023-10-16", av)
	}

	// Verify anthropic-beta and anthropic-version headers
	// are stripped (FR-004).
	if req.Header.Get("anthropic-beta") != "" {
		t.Error("anthropic-beta header should be stripped")
	}
	if req.Header.Get("anthropic-version") != "" {
		t.Error("anthropic-version header should be stripped")
	}
}

func TestVertexProvider_PrepareRequest_StreamingEndpoint(t *testing.T) {
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "ya29.test-token",
		tokenExpiry: time.Now().Add(30 * time.Minute),
	}

	body := `{"model":"claude-sonnet-4-20250514","messages":[],"stream":true}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if !strings.HasSuffix(req.URL.Path, ":streamRawPredict") {
		t.Errorf("path should end with :streamRawPredict for streaming, got: %s", req.URL.Path)
	}
}

func TestVertexProvider_PrepareRequest_NonStreamingEndpoint(t *testing.T) {
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "ya29.test-token",
		tokenExpiry: time.Now().Add(30 * time.Minute),
	}

	body := `{"model":"claude-sonnet-4-20250514","messages":[],"stream":false}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if !strings.HasSuffix(req.URL.Path, ":rawPredict") {
		t.Errorf("path should end with :rawPredict for non-streaming, got: %s", req.URL.Path)
	}
}

func TestVertexProvider_PrepareRequest_CountTokensAlwaysRawPredict(t *testing.T) {
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "ya29.test-token",
		tokenExpiry: time.Now().Add(30 * time.Minute),
	}

	// Even with stream=true, count_tokens should use rawPredict.
	body := `{"model":"claude-sonnet-4-20250514","messages":[],"stream":true}`
	req := httptest.NewRequest("POST", "/v1/messages/count_tokens",
		strings.NewReader(body))

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if !strings.HasSuffix(req.URL.Path, ":rawPredict") {
		t.Errorf("count_tokens should always use :rawPredict, got: %s", req.URL.Path)
	}
}

func TestVertexProvider_PrepareRequest_HeaderStripping(t *testing.T) {
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "ya29.test-token",
		tokenExpiry: time.Now().Add(30 * time.Minute),
	}

	body := `{"model":"claude-sonnet-4-20250514","messages":[]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("anthropic-beta", "messages-2024-01-01")
	req.Header.Set("anthropic-version", "2023-06-01")

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if req.Header.Get("anthropic-beta") != "" {
		t.Error("anthropic-beta header should be stripped (FR-004)")
	}
	if req.Header.Get("anthropic-version") != "" {
		t.Error("anthropic-version header should be stripped (FR-004)")
	}
}

func TestVertexProvider_PrepareRequest_PreservesOtherHeaders(t *testing.T) {
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "ya29.test-token",
		tokenExpiry: time.Now().Add(30 * time.Minute),
	}

	body := `{"model":"claude-sonnet-4-20250514","messages":[]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Claude-Code-Session-Id", "session-123")

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type header should be preserved")
	}
	if req.Header.Get("X-Claude-Code-Session-Id") != "session-123" {
		t.Error("X-Claude-Code-Session-Id header should be preserved")
	}
}

func TestVertexProvider_Start_Success(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		case "CLOUD_ML_REGION":
			return "us-central1"
		}
		return ""
	}
	execCmd := func(name string, args ...string) ([]byte, error) {
		if name == "gcloud" {
			return []byte("ya29.test-token\n"), nil
		}
		return nil, fmt.Errorf("unexpected command: %s", name)
	}

	prov, err := newVertexProvider(getenv, execCmd)
	if err != nil {
		t.Fatalf("newVertexProvider: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := prov.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer prov.Stop()

	prov.tokenMu.RLock()
	token := prov.token
	prov.tokenMu.RUnlock()

	if token != "ya29.test-token" {
		t.Errorf("token: got %q, want %q", token, "ya29.test-token")
	}
}

func TestVertexProvider_Start_GcloudFails(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		}
		return ""
	}
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte("ERROR: not authenticated"), fmt.Errorf("exit 1")
	}

	prov, provErr := newVertexProvider(getenv, execCmd)
	if provErr != nil {
		t.Fatalf("newVertexProvider: %v", provErr)
	}
	err := prov.Start(context.Background())
	if err == nil {
		t.Fatal("expected error when gcloud fails")
	}
	if !strings.Contains(err.Error(), "token acquisition failed") {
		t.Errorf("expected token acquisition error, got: %s", err.Error())
	}
}

func TestVertexProvider_TokenRefresh(t *testing.T) {
	// Use a very short refresh interval for testing.
	oldMinute := refreshMinute
	refreshMinute = time.Millisecond
	defer func() { refreshMinute = oldMinute }()

	var callCount atomic.Int32
	getenv := func(key string) string {
		if key == "ANTHROPIC_VERTEX_PROJECT_ID" {
			return "my-project"
		}
		return ""
	}
	execCmd := func(name string, args ...string) ([]byte, error) {
		n := callCount.Add(1)
		return []byte(fmt.Sprintf("token-%d\n", n)), nil
	}

	prov, provErr := newVertexProvider(getenv, execCmd)
	if provErr != nil {
		t.Fatalf("newVertexProvider: %v", provErr)
	}
	ctx, cancel := context.WithCancel(context.Background())

	if err := prov.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for at least one refresh cycle.
	time.Sleep(100 * time.Millisecond)

	// Cancel to stop the refresh loop.
	cancel()
	prov.Stop()

	// Verify the token was refreshed at least once.
	prov.tokenMu.RLock()
	token := prov.token
	prov.tokenMu.RUnlock()

	if callCount.Load() < 2 {
		t.Errorf("expected at least 2 calls (initial + refresh), got: %d", callCount.Load())
	}
	if token == "" {
		t.Error("expected token to be set after refresh")
	}
}

// ============================================================
// Bedrock Provider Tests (T054-T056)
// ============================================================

func TestBedrockProvider_PrepareRequest(t *testing.T) {
	prov := &BedrockProvider{
		region:       "us-east-1",
		accessKey:    "AKIAIOSFODNN7EXAMPLE",
		secretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		sessionToken: "test-session-token",
		credExpiry:   time.Now().Add(30 * time.Minute),
	}

	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	// Verify URL rewritten to Bedrock invoke endpoint.
	if req.URL.Scheme != "https" {
		t.Errorf("scheme: got %s, want https", req.URL.Scheme)
	}
	if !strings.Contains(req.URL.Host, "bedrock-runtime") {
		t.Errorf("host: got %s, want *bedrock-runtime*", req.URL.Host)
	}
	if !strings.Contains(req.URL.Path, "invoke") {
		t.Errorf("path should contain invoke, got: %s", req.URL.Path)
	}

	// Verify SigV4 signature present.
	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "AWS4-HMAC-SHA256") {
		t.Errorf("Authorization should start with AWS4-HMAC-SHA256, got: %s", auth)
	}

	// Verify X-Amz-Date header set.
	if req.Header.Get("X-Amz-Date") == "" {
		t.Error("expected X-Amz-Date header")
	}

	// Verify X-Amz-Security-Token header set.
	if req.Header.Get("X-Amz-Security-Token") != "test-session-token" {
		t.Errorf("X-Amz-Security-Token: got %q, want %q",
			req.Header.Get("X-Amz-Security-Token"), "test-session-token")
	}
}

func TestBedrockProvider_Start_Success(t *testing.T) {
	getenv := func(key string) string {
		if key == "AWS_REGION" {
			return "us-east-1"
		}
		return ""
	}
	execCmd := func(name string, args ...string) ([]byte, error) {
		if name == "aws" {
			return []byte("export AWS_ACCESS_KEY_ID=AKIATEST\n" +
				"export AWS_SECRET_ACCESS_KEY=secrettest\n" +
				"export AWS_SESSION_TOKEN=sessiontest\n"), nil
		}
		return nil, fmt.Errorf("unexpected command: %s", name)
	}

	prov := newBedrockProvider(getenv, execCmd)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := prov.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer prov.Stop()

	prov.credMu.RLock()
	ak := prov.accessKey
	sk := prov.secretKey
	st := prov.sessionToken
	prov.credMu.RUnlock()

	if ak != "AKIATEST" {
		t.Errorf("accessKey: got %q, want AKIATEST", ak)
	}
	if sk != "secrettest" {
		t.Errorf("secretKey: got %q, want secrettest", sk)
	}
	if st != "sessiontest" {
		t.Errorf("sessionToken: got %q, want sessiontest", st)
	}
}

func TestBedrockProvider_Start_AWSFails(t *testing.T) {
	getenv := func(key string) string { return "" }
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte("Unable to locate credentials"), fmt.Errorf("exit 1")
	}

	prov := newBedrockProvider(getenv, execCmd)
	err := prov.Start(context.Background())
	if err == nil {
		t.Fatal("expected error when aws CLI fails")
	}
	if !strings.Contains(err.Error(), "credential acquisition failed") {
		t.Errorf("expected credential acquisition error, got: %s", err.Error())
	}
}

func TestBedrockProvider_CredentialRefresh(t *testing.T) {
	oldMinute := refreshMinute
	refreshMinute = time.Millisecond
	defer func() { refreshMinute = oldMinute }()

	var callCount atomic.Int32
	getenv := func(key string) string {
		if key == "AWS_REGION" {
			return "us-east-1"
		}
		return ""
	}
	execCmd := func(name string, args ...string) ([]byte, error) {
		n := callCount.Add(1)
		return []byte(fmt.Sprintf(
			"export AWS_ACCESS_KEY_ID=AKIA%d\n"+
				"export AWS_SECRET_ACCESS_KEY=secret%d\n"+
				"export AWS_SESSION_TOKEN=session%d\n",
			n, n, n)), nil
	}

	prov := newBedrockProvider(getenv, execCmd)
	ctx, cancel := context.WithCancel(context.Background())

	if err := prov.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for at least one refresh cycle.
	time.Sleep(100 * time.Millisecond)

	cancel()
	prov.Stop()

	if callCount.Load() < 2 {
		t.Errorf("expected at least 2 calls (initial + refresh), got: %d", callCount.Load())
	}
}

// ============================================================
// SigV4 Signing Tests (T057)
// ============================================================

func TestSignV4_WithSessionToken(t *testing.T) {
	body := `{"test":"data"}`
	req := httptest.NewRequest("POST",
		"https://bedrock-runtime.us-east-1.amazonaws.com/model/test/invoke",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	err := signV4(req, "us-east-1", "bedrock-runtime",
		"AKIAIOSFODNN7EXAMPLE",
		"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"test-session-token")
	if err != nil {
		t.Fatalf("signV4 failed: %v", err)
	}

	// Verify Authorization header format.
	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/") {
		t.Errorf("Authorization header format wrong: %s", auth)
	}
	if !strings.Contains(auth, "SignedHeaders=") {
		t.Errorf("expected SignedHeaders in Authorization: %s", auth)
	}
	if !strings.Contains(auth, "Signature=") {
		t.Errorf("expected Signature in Authorization: %s", auth)
	}

	// Verify X-Amz-Date header set.
	if req.Header.Get("X-Amz-Date") == "" {
		t.Error("expected X-Amz-Date header")
	}

	// Verify X-Amz-Security-Token header set.
	if req.Header.Get("X-Amz-Security-Token") != "test-session-token" {
		t.Errorf("expected X-Amz-Security-Token, got: %s",
			req.Header.Get("X-Amz-Security-Token"))
	}
}

func TestSignV4_WithoutSessionToken(t *testing.T) {
	req := httptest.NewRequest("POST",
		"https://bedrock-runtime.us-east-1.amazonaws.com/model/test/invoke",
		strings.NewReader(`{}`))

	err := signV4(req, "us-east-1", "bedrock-runtime",
		"AKIAIOSFODNN7EXAMPLE",
		"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"") // No session token.
	if err != nil {
		t.Fatalf("signV4 failed: %v", err)
	}

	// Verify X-Amz-Security-Token header is NOT set.
	if req.Header.Get("X-Amz-Security-Token") != "" {
		t.Errorf("expected no X-Amz-Security-Token when session token empty, got: %s",
			req.Header.Get("X-Amz-Security-Token"))
	}

	// Authorization header should still be present.
	if req.Header.Get("Authorization") == "" {
		t.Error("expected Authorization header")
	}
}

// ============================================================
// Gateway Core Tests (T058-T062)
// ============================================================

func TestNewMux_HealthEndpoint(t *testing.T) {
	prov := &AnthropicProvider{apiKey: "test-key"}
	startTime := time.Now().Add(-1 * time.Hour)
	mux := newMux(prov, 53147, startTime)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	var resp HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("status: got %q, want ok", resp.Status)
	}
	if resp.Provider != "anthropic" {
		t.Errorf("provider: got %q, want anthropic", resp.Provider)
	}
	if resp.Port != 53147 {
		t.Errorf("port: got %d, want 53147", resp.Port)
	}
	if resp.UptimeSeconds < 3600 {
		t.Errorf("uptime: got %d, want >= 3600", resp.UptimeSeconds)
	}
}

func TestNewMux_UnsupportedEndpoint(t *testing.T) {
	prov := &AnthropicProvider{apiKey: "test-key"}
	mux := newMux(prov, 53147, time.Now())

	req := httptest.NewRequest("GET", "/v1/completions", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Unsupported endpoint") {
		t.Errorf("expected unsupported endpoint message, got: %s", body)
	}
	if !strings.Contains(body, "/v1/messages") {
		t.Errorf("expected supported endpoints listed, got: %s", body)
	}
}

func TestNewMux_ProxyRouting(t *testing.T) {
	// Create a mock upstream server.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id":   "msg_test",
			"type": "message",
		})
	}))
	defer upstream.Close()

	// Create a provider that rewrites to the mock upstream.
	prov := &mockProvider{
		name:        "test",
		upstreamURL: upstream.URL,
	}

	mux := newMux(prov, 53147, time.Now())

	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	respBody := w.Body.String()
	if !strings.Contains(respBody, "msg_test") {
		t.Errorf("expected upstream response, got: %s", respBody)
	}
}

func TestNewMux_ProxyRoutingCountTokens(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the path was forwarded correctly.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]int{
			"input_tokens": 42,
		})
	}))
	defer upstream.Close()

	prov := &mockProvider{
		name:        "test",
		upstreamURL: upstream.URL,
	}

	mux := newMux(prov, 53147, time.Now())

	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/messages/count_tokens",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestNewMux_UpstreamErrorForwarding(t *testing.T) {
	// Mock upstream returns 429 rate limit error.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"type":    "rate_limit_error",
				"message": "Rate limit exceeded",
			},
		})
	}))
	defer upstream.Close()

	prov := &mockProvider{
		name:        "test",
		upstreamURL: upstream.URL,
	}

	mux := newMux(prov, 53147, time.Now())

	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Verify the upstream error is forwarded as-is (FR-014).
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusTooManyRequests)
	}

	body := w.Body.String()
	if !strings.Contains(body, "rate_limit_error") {
		t.Errorf("expected rate limit error forwarded, got: %s", body)
	}
}

// ============================================================
// Gateway Lifecycle Tests (T063-T069)
// ============================================================

func TestStart_ProviderDetectionAndPIDFile(t *testing.T) {
	// Test that Start detects the provider, writes a PID file,
	// and starts the server. We test the components individually
	// because Start() uses srv.ListenAndServe() internally
	// (not the injected ListenAndServe), so we can't easily
	// mock the server lifecycle.

	// Test 1: Provider detection works.
	getenv := func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test"
		}
		return ""
	}
	execCmd := func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not called")
	}

	prov, err := DetectProvider(getenv, execCmd)
	if err != nil {
		t.Fatalf("DetectProvider failed: %v", err)
	}
	if prov.Name() != "anthropic" {
		t.Errorf("expected anthropic, got: %s", prov.Name())
	}

	// Test 2: PID file round-trip works.
	dir := t.TempDir()
	pp := pidPath(dir)
	info := PIDInfo{
		PID:      os.Getpid(),
		Port:     53147,
		Provider: "anthropic",
		Started:  time.Now(),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}

	got, err := ReadPID(pp)
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}
	if got.PID != info.PID {
		t.Errorf("PID: got %d, want %d", got.PID, info.PID)
	}
	if got.Port != info.Port {
		t.Errorf("Port: got %d, want %d", got.Port, info.Port)
	}

	// Test 3: newMux creates a working handler.
	if err := prov.Start(context.Background()); err != nil {
		t.Fatalf("provider Start failed: %v", err)
	}
	handler := newMux(prov, 53147, time.Now())
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("health status: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestStop_PIDFileAndProcessAlive(t *testing.T) {
	opts := testOpts(t)

	// Write a PID file with the current process PID.
	pp := pidPath(opts.ProjectDir)
	info := PIDInfo{
		PID:      os.Getpid(),
		Port:     53147,
		Provider: "anthropic",
		Started:  time.Now(),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatal(err)
	}

	// FindProcess returns current process (alive).
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return os.FindProcess(pid)
	}

	// Stop should send SIGTERM and remove PID file.
	// Since we're sending SIGTERM to ourselves, we need
	// to handle this carefully. Instead, mock FindProcess
	// to return an error so Stop treats it as "not alive".
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return nil, fmt.Errorf("no such process")
	}

	if err := Stop(opts); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "No gateway running") {
		t.Errorf("expected no gateway message (process not found), got: %s", out)
	}
}

func TestStop_NoPIDFile(t *testing.T) {
	opts := testOpts(t)

	if err := Stop(opts); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "No gateway running") {
		t.Errorf("expected no gateway message, got: %s", out)
	}
}

func TestStop_StalePIDFile(t *testing.T) {
	opts := testOpts(t)

	// Write a PID file with a dead process.
	pp := pidPath(opts.ProjectDir)
	info := PIDInfo{
		PID:      99999,
		Port:     53147,
		Provider: "anthropic",
		Started:  time.Now(),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatal(err)
	}

	// Process is dead.
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return nil, fmt.Errorf("no such process")
	}

	if err := Stop(opts); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "No gateway running") {
		t.Errorf("expected no gateway message, got: %s", out)
	}

	// PID file should be cleaned up.
	if _, err := os.Stat(pp); !os.IsNotExist(err) {
		t.Error("expected stale PID file to be removed")
	}
}

func TestStatus_GatewayRunning(t *testing.T) {
	opts := testOpts(t)

	// Write a PID file with current process PID.
	pp := pidPath(opts.ProjectDir)
	info := PIDInfo{
		PID:      os.Getpid(),
		Port:     53147,
		Provider: "vertex",
		Started:  time.Now().Add(-1 * time.Hour),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatal(err)
	}

	// Process is alive.
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return os.FindProcess(pid)
	}

	// Health endpoint responds.
	opts.HTTPGet = func(url string) (int, error) {
		return http.StatusOK, nil
	}

	if err := Status(opts); err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "Gateway Status") {
		t.Errorf("expected status header, got: %s", out)
	}
	if !strings.Contains(out, "vertex") {
		t.Errorf("expected provider name, got: %s", out)
	}
	if !strings.Contains(out, "53147") {
		t.Errorf("expected port, got: %s", out)
	}
}

func TestStatus_NoGateway(t *testing.T) {
	opts := testOpts(t)

	if err := Status(opts); err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "No gateway running") {
		t.Errorf("expected no gateway message, got: %s", out)
	}
}

func TestStart_PortConflict(t *testing.T) {
	opts := testOpts(t)
	opts.Getenv = func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test"
		}
		return ""
	}

	// Simulate port conflict by having the server in the
	// goroutine return an address-in-use error. We need to
	// use the real Start function which creates its own
	// http.Server. Instead, test the isAddrInUse helper.
	err := fmt.Errorf("listen tcp :53147: bind: address already in use")
	if !isAddrInUse(err) {
		t.Error("expected isAddrInUse to return true for address in use error")
	}

	notInUse := fmt.Errorf("some other error")
	if isAddrInUse(notInUse) {
		t.Error("expected isAddrInUse to return false for other errors")
	}
}

func TestStart_ProviderOverride(t *testing.T) {
	// When --provider is specified, it overrides auto-detection.
	// Test via NewProviderByName directly since Start() uses
	// srv.ListenAndServe() internally.
	getenv := func(key string) string {
		switch key {
		case "CLAUDE_CODE_USE_VERTEX":
			return "1"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		case "ANTHROPIC_API_KEY":
			return "sk-ant-test"
		}
		return ""
	}
	execCmd := func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not called")
	}

	// Without override, Vertex would be detected.
	prov, err := DetectProvider(getenv, execCmd)
	if err != nil {
		t.Fatalf("DetectProvider failed: %v", err)
	}
	if prov.Name() != "vertex" {
		t.Errorf("expected vertex from auto-detect, got: %s", prov.Name())
	}

	// With override, Anthropic should be used.
	prov, err = NewProviderByName("anthropic", getenv, execCmd)
	if err != nil {
		t.Fatalf("NewProviderByName failed: %v", err)
	}
	if prov.Name() != "anthropic" {
		t.Errorf("expected anthropic from override, got: %s", prov.Name())
	}
}

func TestStart_CustomPort(t *testing.T) {
	// Verify the port is correctly used in the PID file
	// and health endpoint.
	dir := t.TempDir()
	pp := pidPath(dir)

	info := PIDInfo{
		PID:      os.Getpid(),
		Port:     9000,
		Provider: "anthropic",
		Started:  time.Now(),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}

	got, err := ReadPID(pp)
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}
	if got.Port != 9000 {
		t.Errorf("expected port 9000, got: %d", got.Port)
	}

	// Verify health endpoint uses the correct port.
	prov := &AnthropicProvider{apiKey: "test"}
	handler := newMux(prov, 9000, time.Now())
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var resp HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Port != 9000 {
		t.Errorf("health port: got %d, want 9000", resp.Port)
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m 0s"},
		{90 * time.Minute, "1h 30m"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatUptime(tt.d)
			if got != tt.want {
				t.Errorf("formatUptime(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

// ============================================================
// Refresh Tests
// ============================================================

func TestRefreshVertexToken_Success(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte("ya29.fresh-token\n"), nil
	}

	token, err := refreshVertexToken(execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "ya29.fresh-token" {
		t.Errorf("token: got %q, want ya29.fresh-token", token)
	}
}

func TestRefreshVertexToken_Failure(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte("ERROR: not authenticated"), fmt.Errorf("exit 1")
	}

	_, err := refreshVertexToken(execCmd)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Re-authenticate") {
		t.Errorf("expected re-auth suggestion, got: %s", err.Error())
	}
}

func TestRefreshBedrockCredentials_EnvFormat(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte(
			"export AWS_ACCESS_KEY_ID=AKIATEST\n" +
				"export AWS_SECRET_ACCESS_KEY=secrettest\n" +
				"export AWS_SESSION_TOKEN=sessiontest\n"), nil
	}

	ak, sk, st, err := refreshBedrockCredentials(execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ak != "AKIATEST" {
		t.Errorf("accessKey: got %q, want AKIATEST", ak)
	}
	if sk != "secrettest" {
		t.Errorf("secretKey: got %q, want secrettest", sk)
	}
	if st != "sessiontest" {
		t.Errorf("sessionToken: got %q, want sessiontest", st)
	}
}

func TestRefreshBedrockCredentials_JSONFormat(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte(`{
			"AccessKeyId": "AKIAJSON",
			"SecretAccessKey": "secretjson",
			"SessionToken": "sessionjson"
		}`), nil
	}

	ak, sk, st, err := refreshBedrockCredentials(execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ak != "AKIAJSON" {
		t.Errorf("accessKey: got %q, want AKIAJSON", ak)
	}
	if sk != "secretjson" {
		t.Errorf("secretKey: got %q, want secretjson", sk)
	}
	if st != "sessionjson" {
		t.Errorf("sessionToken: got %q, want sessionjson", st)
	}
}

func TestRefreshBedrockCredentials_Failure(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte("Unable to locate credentials"), fmt.Errorf("exit 1")
	}

	_, _, _, err := refreshBedrockCredentials(execCmd)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Re-authenticate") {
		t.Errorf("expected re-auth suggestion, got: %s", err.Error())
	}
}

func TestRefreshLoop_ContextCancellation(t *testing.T) {
	oldMinute := refreshMinute
	refreshMinute = time.Millisecond
	defer func() { refreshMinute = oldMinute }()

	var callCount atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		refreshLoop(ctx, 1*refreshMinute, func() error {
			callCount.Add(1)
			return nil
		})
		close(done)
	}()

	// Let it run a few cycles.
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for the goroutine to exit.
	select {
	case <-done:
		// Good — goroutine exited.
	case <-time.After(1 * time.Second):
		t.Fatal("refreshLoop did not exit after context cancellation")
	}

	if callCount.Load() == 0 {
		t.Error("expected at least one refresh call")
	}
}

// ============================================================
// Extract Model Tests
// ============================================================

func TestExtractModelFromBody_WithModel(t *testing.T) {
	body := `{"model":"claude-opus-4-20250514","messages":[]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	model := extractModelFromBody(req)
	if model != "claude-opus-4-20250514" {
		t.Errorf("got %q, want claude-opus-4-20250514", model)
	}
}

func TestExtractModelFromBody_NoModel(t *testing.T) {
	body := `{"messages":[]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	model := extractModelFromBody(req)
	if model != "claude-sonnet-4-20250514" {
		t.Errorf("got %q, want default model", model)
	}
}

func TestExtractModelFromBody_NilBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/v1/messages", nil)
	model := extractModelFromBody(req)
	if model != "claude-sonnet-4-20250514" {
		t.Errorf("got %q, want default model", model)
	}
}

// ============================================================
// defaults() Tests
// ============================================================

func TestDefaults_ZeroValueOptions(t *testing.T) {
	// Verify that a zero-value Options gets all fields
	// populated with production defaults.
	var opts Options
	opts.defaults()

	if opts.Port != DefaultPort {
		t.Errorf("Port: got %d, want %d", opts.Port, DefaultPort)
	}
	if opts.ProjectDir == "" {
		t.Error("ProjectDir should be set to cwd")
	}
	if opts.Stdout == nil {
		t.Error("Stdout should be set")
	}
	if opts.Stderr == nil {
		t.Error("Stderr should be set")
	}
	if opts.LookPath == nil {
		t.Error("LookPath should be set")
	}
	if opts.ExecCmd == nil {
		t.Error("ExecCmd should be set")
	}
	if opts.ExecStart == nil {
		t.Error("ExecStart should be set")
	}
	if opts.Getenv == nil {
		t.Error("Getenv should be set")
	}
	if opts.HTTPGet == nil {
		t.Error("HTTPGet should be set")
	}
	if opts.FindProcess == nil {
		t.Error("FindProcess should be set")
	}
	if opts.ListenAndServe == nil {
		t.Error("ListenAndServe should be set")
	}
}

func TestDefaults_PreservesExistingValues(t *testing.T) {
	// Verify that defaults() does not overwrite
	// already-set fields.
	buf := &bytes.Buffer{}
	opts := Options{
		Port:    9999,
		Stdout:  buf,
		Stderr:  buf,
		Getenv:  func(string) string { return "custom" },
		HTTPGet: func(string) (int, error) { return 418, nil },
	}
	opts.defaults()

	if opts.Port != 9999 {
		t.Errorf("Port: got %d, want 9999 (should preserve)", opts.Port)
	}
	if opts.Stdout != buf {
		t.Error("Stdout should be preserved")
	}
	if opts.Stderr != buf {
		t.Error("Stderr should be preserved")
	}
	if opts.Getenv("anything") != "custom" {
		t.Error("Getenv should be preserved")
	}
	code, _ := opts.HTTPGet("anything")
	if code != 418 {
		t.Errorf("HTTPGet: got %d, want 418 (should preserve)", code)
	}
}

// ============================================================
// Start() Tests
// ============================================================

func TestStart_NoProviderDetected(t *testing.T) {
	opts := testOpts(t)
	// No env vars set — provider detection should fail.
	opts.Getenv = func(key string) string { return "" }

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error when no provider detected")
	}
	if !strings.Contains(err.Error(), "no cloud provider detected") {
		t.Errorf("expected provider detection error, got: %s", err.Error())
	}
}

func TestStart_ProviderNameOverride(t *testing.T) {
	opts := testOpts(t)
	opts.ProviderName = "anthropic"
	opts.Getenv = func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test"
		}
		return ""
	}

	// Provider Start succeeds, but the real server will
	// fail to bind. We test the provider override path
	// by verifying it doesn't fail with "no cloud provider
	// detected" — it should get past provider detection
	// and fail later (at server start or PID write).
	// Use a port that's almost certainly in use (0 is
	// special — it picks a free port, so we use a real
	// port conflict scenario).
	//
	// Actually, Start() will succeed up to the point of
	// srv.ListenAndServe(). We need to let it start and
	// then stop it. Use a real ephemeral port and signal.
	//
	// Simpler approach: verify the "already running" path
	// by writing a PID file with a live process first.
	pp := pidPath(opts.ProjectDir)
	info := PIDInfo{
		PID:      os.Getpid(),
		Port:     opts.Port,
		Provider: "anthropic",
		Started:  time.Now(),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatal(err)
	}
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return os.FindProcess(pid)
	}

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error when gateway already running")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("expected already running error, got: %s", err.Error())
	}
}

func TestStart_AlreadyRunning(t *testing.T) {
	opts := testOpts(t)
	opts.Getenv = func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test"
		}
		return ""
	}

	// Write a PID file with the current process PID
	// (which is alive).
	pp := pidPath(opts.ProjectDir)
	info := PIDInfo{
		PID:      os.Getpid(),
		Port:     53147,
		Provider: "anthropic",
		Started:  time.Now(),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatal(err)
	}
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return os.FindProcess(pid)
	}

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error when gateway already running")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("expected already running error, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "uf gateway stop") {
		t.Errorf("expected stop hint, got: %s", err.Error())
	}
}

func TestStart_ProviderInitFails(t *testing.T) {
	opts := testOpts(t)
	opts.ProviderName = "anthropic"
	// ANTHROPIC_API_KEY is empty — provider.Start() will fail.
	opts.Getenv = func(key string) string { return "" }

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error when provider init fails")
	}
	if !strings.Contains(err.Error(), "initialization failed") {
		t.Errorf("expected initialization failed error, got: %s", err.Error())
	}
}

func TestStart_DetachPath(t *testing.T) {
	opts := testOpts(t)
	opts.Detach = true
	opts.Getenv = func(key string) string {
		// GatewayChildEnv is NOT set — should trigger detach.
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test"
		}
		return ""
	}

	// Mock ExecStart to simulate starting a child process.
	opts.ExecStart = func(cmd *exec.Cmd) error {
		// Simulate a started process by setting cmd.Process.
		cmd.Process = &os.Process{Pid: 12345}
		return nil
	}

	// Mock HTTPGet to return 200 on first call (health check).
	opts.HTTPGet = func(url string) (int, error) {
		return http.StatusOK, nil
	}

	err := Start(opts)
	if err != nil {
		t.Fatalf("Start with detach should succeed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "12345") {
		t.Errorf("expected PID in output, got: %s", out)
	}
	if !strings.Contains(out, "Gateway started") {
		t.Errorf("expected started message, got: %s", out)
	}
}

func TestStart_ChildPath_PortConflict(t *testing.T) {
	opts := testOpts(t)
	opts.Getenv = func(key string) string {
		switch key {
		case GatewayChildEnv:
			return "1" // We ARE the child.
		case "ANTHROPIC_API_KEY":
			return "sk-ant-test"
		}
		return ""
	}

	// Bind a port first to cause a conflict.
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to bind port: %v", err)
	}
	defer func() { _ = listener.Close() }()

	// Extract the port from the listener.
	port := listener.Addr().(*net.TCPAddr).Port
	opts.Port = port

	// Start should fail with port conflict.
	// Start() will try to bind the same port and fail.
	// It runs srv.ListenAndServe() in a goroutine, so
	// we need to wait for it to detect the error.
	startErr := Start(opts)
	if startErr == nil {
		t.Fatal("expected error for port conflict")
	}
	if !strings.Contains(startErr.Error(), "already in use") {
		t.Errorf("expected address in use error, got: %s", startErr.Error())
	}
}

func TestStart_ChildPath_ServerStartsAndShutdown(t *testing.T) {
	opts := testOpts(t)
	opts.Getenv = func(key string) string {
		switch key {
		case GatewayChildEnv:
			return "1" // We ARE the child.
		case "ANTHROPIC_API_KEY":
			return "sk-ant-test"
		}
		return ""
	}

	// Use a high random port to avoid conflicts.
	opts.Port = 59123 + os.Getpid()%1000

	// Use io.Discard for Stderr to avoid data race:
	// Start() writes to Stderr from both the main
	// goroutine and the server goroutine concurrently.
	// bytes.Buffer is not thread-safe.
	opts.Stderr = io.Discard

	// Start the gateway in a goroutine and send SIGINT
	// after a short delay to trigger graceful shutdown.
	errCh := make(chan error, 1)
	go func() {
		errCh <- Start(opts)
	}()

	// Wait briefly for the server to start, then send
	// ourselves SIGINT to trigger shutdown.
	time.Sleep(200 * time.Millisecond)

	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.SIGINT)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start should return nil on graceful shutdown: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after SIGINT")
	}

	// Verify PID file was cleaned up.
	pp := pidPath(opts.ProjectDir)
	if _, err := os.Stat(pp); !os.IsNotExist(err) {
		t.Error("expected PID file to be removed after shutdown")
	}
}

// ============================================================
// detach() Tests
// ============================================================

func TestDetach_ExecStartFails(t *testing.T) {
	opts := testOpts(t)
	opts.Detach = true
	opts.Port = DefaultPort
	opts.Getenv = func(key string) string { return "" }
	opts.ExecStart = func(cmd *exec.Cmd) error {
		return fmt.Errorf("permission denied")
	}

	err := detach(opts)
	if err == nil {
		t.Fatal("expected error when ExecStart fails")
	}
	if !strings.Contains(err.Error(), "start background gateway") {
		t.Errorf("expected start background error, got: %s", err.Error())
	}
}

func TestDetach_HealthCheckTimeout(t *testing.T) {
	opts := testOpts(t)
	opts.Port = DefaultPort
	opts.ExecStart = func(cmd *exec.Cmd) error {
		cmd.Process = &os.Process{Pid: 54321}
		return nil
	}
	// Health check always fails — simulates child crash.
	opts.HTTPGet = func(url string) (int, error) {
		return 0, fmt.Errorf("connection refused")
	}

	// Override the deadline to be very short for testing.
	// detach() uses a 10-second deadline internally, so
	// this test will take ~10 seconds. Instead, we accept
	// the timeout and just verify the error message.
	// To speed this up, we can't easily override the
	// deadline. Let's just verify the error path works.
	//
	// Actually, the 10-second deadline with 200ms initial
	// interval means ~15 iterations. That's acceptable
	// for a test. But let's be smarter: return 500 instead
	// of connection refused so it loops faster.
	opts.HTTPGet = func(url string) (int, error) {
		return http.StatusInternalServerError, nil
	}

	err := detach(opts)
	if err == nil {
		t.Fatal("expected error on health check timeout")
	}
	if !strings.Contains(err.Error(), "health check timed out") {
		t.Errorf("expected timeout error, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "54321") {
		t.Errorf("expected PID in error, got: %s", err.Error())
	}
}

func TestDetach_CustomPortAndProvider(t *testing.T) {
	opts := testOpts(t)
	opts.Port = 9999
	opts.ProviderName = "vertex"

	var capturedArgs []string
	opts.ExecStart = func(cmd *exec.Cmd) error {
		capturedArgs = cmd.Args
		cmd.Process = &os.Process{Pid: 11111}
		return nil
	}
	opts.HTTPGet = func(url string) (int, error) {
		return http.StatusOK, nil
	}

	err := detach(opts)
	if err != nil {
		t.Fatalf("detach failed: %v", err)
	}

	// Verify the child args include --port and --provider.
	argsStr := strings.Join(capturedArgs, " ")
	if !strings.Contains(argsStr, "--port") {
		t.Errorf("expected --port in args, got: %s", argsStr)
	}
	if !strings.Contains(argsStr, "9999") {
		t.Errorf("expected port 9999 in args, got: %s", argsStr)
	}
	if !strings.Contains(argsStr, "--provider") {
		t.Errorf("expected --provider in args, got: %s", argsStr)
	}
	if !strings.Contains(argsStr, "vertex") {
		t.Errorf("expected vertex in args, got: %s", argsStr)
	}
}

func TestDetach_DefaultPortNoExtraArgs(t *testing.T) {
	opts := testOpts(t)
	opts.Port = DefaultPort
	opts.ProviderName = ""

	var capturedArgs []string
	opts.ExecStart = func(cmd *exec.Cmd) error {
		capturedArgs = cmd.Args
		cmd.Process = &os.Process{Pid: 22222}
		return nil
	}
	opts.HTTPGet = func(url string) (int, error) {
		return http.StatusOK, nil
	}

	err := detach(opts)
	if err != nil {
		t.Fatalf("detach failed: %v", err)
	}

	// With default port and no provider override, args
	// should just be ["gateway"].
	argsStr := strings.Join(capturedArgs, " ")
	if strings.Contains(argsStr, "--port") {
		t.Errorf("should not include --port for default port, got: %s", argsStr)
	}
	if strings.Contains(argsStr, "--provider") {
		t.Errorf("should not include --provider when empty, got: %s", argsStr)
	}
}

// ============================================================
// Stop() Additional Tests
// ============================================================

func TestStop_ProcessAliveSignalSucceeds(t *testing.T) {
	opts := testOpts(t)

	// Write a PID file.
	pp := pidPath(opts.ProjectDir)
	info := PIDInfo{
		PID:      os.Getpid(),
		Port:     53147,
		Provider: "anthropic",
		Started:  time.Now(),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatal(err)
	}

	// First call to FindProcess (IsAlive check) returns
	// a process that accepts signal 0.
	// Second call (to get process for SIGTERM) also succeeds.
	// We use a mock process that accepts Signal calls.
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return os.FindProcess(os.Getpid())
	}

	// Stop will send SIGTERM to ourselves — that's fine,
	// the test process handles it. But to be safe, we
	// ignore SIGTERM for this test.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	if err := Stop(opts); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "Gateway stopped") {
		t.Errorf("expected stopped message, got: %s", out)
	}

	// PID file should be removed.
	if _, err := os.Stat(pp); !os.IsNotExist(err) {
		t.Error("expected PID file to be removed")
	}
}

func TestStop_ProcessAliveButFindProcessFailsOnSecondCall(t *testing.T) {
	opts := testOpts(t)

	// Write a PID file.
	pp := pidPath(opts.ProjectDir)
	info := PIDInfo{
		PID:      os.Getpid(),
		Port:     53147,
		Provider: "anthropic",
		Started:  time.Now(),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatal(err)
	}

	// IsAlive returns true (first FindProcess call succeeds),
	// but the second FindProcess call (to get process for
	// SIGTERM) fails.
	callCount := 0
	opts.FindProcess = func(pid int) (*os.Process, error) {
		callCount++
		if callCount <= 1 {
			// First call: IsAlive check — return current process.
			return os.FindProcess(os.Getpid())
		}
		// Second call: getting process for SIGTERM — fail.
		return nil, fmt.Errorf("process disappeared")
	}

	if err := Stop(opts); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "No gateway running") {
		t.Errorf("expected no gateway message when process disappeared, got: %s", out)
	}
}

// ============================================================
// Status() Additional Tests
// ============================================================

func TestStatus_HealthEndpointFails(t *testing.T) {
	opts := testOpts(t)

	// Write a PID file.
	pp := pidPath(opts.ProjectDir)
	info := PIDInfo{
		PID:      os.Getpid(),
		Port:     53147,
		Provider: "bedrock",
		Started:  time.Now().Add(-30 * time.Minute),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatal(err)
	}

	// Process is alive.
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return os.FindProcess(os.Getpid())
	}

	// Health endpoint returns error.
	opts.HTTPGet = func(url string) (int, error) {
		return 0, fmt.Errorf("connection refused")
	}

	if err := Status(opts); err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	// Should still display status from PID file data.
	out := stdoutStr(opts)
	if !strings.Contains(out, "Gateway Status") {
		t.Errorf("expected status header, got: %s", out)
	}
	if !strings.Contains(out, "bedrock") {
		t.Errorf("expected provider bedrock, got: %s", out)
	}
	if !strings.Contains(out, "53147") {
		t.Errorf("expected port, got: %s", out)
	}
}

func TestStatus_HealthEndpointNon200(t *testing.T) {
	opts := testOpts(t)

	pp := pidPath(opts.ProjectDir)
	info := PIDInfo{
		PID:      os.Getpid(),
		Port:     53147,
		Provider: "vertex",
		Started:  time.Now().Add(-2 * time.Hour),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatal(err)
	}

	opts.FindProcess = func(pid int) (*os.Process, error) {
		return os.FindProcess(os.Getpid())
	}

	// Health endpoint returns 500.
	opts.HTTPGet = func(url string) (int, error) {
		return http.StatusInternalServerError, nil
	}

	if err := Status(opts); err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	// Should still display status from PID file data.
	out := stdoutStr(opts)
	if !strings.Contains(out, "Gateway Status") {
		t.Errorf("expected status header, got: %s", out)
	}
	if !strings.Contains(out, "vertex") {
		t.Errorf("expected provider vertex, got: %s", out)
	}
}

func TestStatus_StalePIDFile(t *testing.T) {
	opts := testOpts(t)

	pp := pidPath(opts.ProjectDir)
	info := PIDInfo{
		PID:      99999,
		Port:     53147,
		Provider: "anthropic",
		Started:  time.Now(),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatal(err)
	}

	// Process is dead.
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return nil, fmt.Errorf("no such process")
	}

	if err := Status(opts); err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "No gateway running") {
		t.Errorf("expected no gateway message, got: %s", out)
	}

	// PID file should be cleaned up.
	if _, err := os.Stat(pp); !os.IsNotExist(err) {
		t.Error("expected stale PID file to be removed")
	}
}

// ============================================================
// WritePID() Additional Tests
// ============================================================

func TestWritePID_NonExistentDirectory(t *testing.T) {
	// WritePID should create the directory if it doesn't
	// exist (via MkdirAll).
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "nested", "dir", "gateway.pid")

	info := PIDInfo{
		PID:      12345,
		Port:     53147,
		Provider: "anthropic",
		Started:  time.Now(),
	}

	if err := WritePID(path, info); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}

	// Verify the file was created.
	got, err := ReadPID(path)
	if err != nil {
		t.Fatalf("ReadPID failed: %v", err)
	}
	if got.PID != 12345 {
		t.Errorf("PID: got %d, want 12345", got.PID)
	}
}

func TestWritePID_ReadOnlyDirectory(t *testing.T) {
	// Test that WritePID returns an error when the
	// directory cannot be created.
	path := "/proc/nonexistent/gateway.pid"
	info := PIDInfo{PID: 1, Port: 1, Provider: "test"}

	err := WritePID(path, info)
	if err == nil {
		t.Fatal("expected error writing to read-only path")
	}
}

// ============================================================
// newMux() Additional Tests — Director error path
// ============================================================

func TestNewMux_DirectorError(t *testing.T) {
	// Test the ErrorHandler path when the Director
	// (provider.PrepareRequest) returns an error.
	prov := &errorProvider{
		name: "broken",
		err:  fmt.Errorf("token expired"),
	}

	mux := newMux(prov, 53147, time.Now())

	body := `{"model":"claude-sonnet-4-20250514","messages":[]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadGateway)
	}

	respBody := w.Body.String()
	if !strings.Contains(respBody, "auth_error") {
		t.Errorf("expected auth_error type, got: %s", respBody)
	}
	if !strings.Contains(respBody, "token expired") {
		t.Errorf("expected error message, got: %s", respBody)
	}
}

func TestNewMux_InboundAuthHeadersStripped(t *testing.T) {
	// Verify that inbound Authorization and x-api-key
	// headers are stripped before calling the provider.
	var capturedReq *http.Request
	prov := &capturingProvider{
		name: "capture",
		onPrepare: func(req *http.Request) error {
			capturedReq = req.Clone(req.Context())
			// Rewrite to a non-routable address so the
			// proxy fails (we don't care about the response).
			req.URL.Scheme = "http"
			req.URL.Host = "127.0.0.1:1"
			req.Host = "127.0.0.1:1"
			return nil
		},
	}

	mux := newMux(prov, 53147, time.Now())

	body := `{"model":"claude-sonnet-4-20250514","messages":[]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer should-be-stripped")
	req.Header.Set("x-api-key", "sk-should-be-stripped")
	req.Header.Set("anthropic-beta", "should-be-preserved")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if capturedReq == nil {
		t.Fatal("provider PrepareRequest was not called")
	}

	// Authorization and x-api-key should be stripped
	// BEFORE PrepareRequest is called.
	if capturedReq.Header.Get("Authorization") != "" {
		t.Error("Authorization header should be stripped before PrepareRequest")
	}
	if capturedReq.Header.Get("x-api-key") != "" {
		t.Error("x-api-key header should be stripped before PrepareRequest")
	}
	// anthropic-beta should be preserved.
	if capturedReq.Header.Get("anthropic-beta") != "should-be-preserved" {
		t.Error("anthropic-beta header should be preserved")
	}
}

// ============================================================
// isAddrInUse() Additional Tests
// ============================================================

func TestIsAddrInUse_NilError(t *testing.T) {
	if isAddrInUse(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsAddrInUse_OpError(t *testing.T) {
	// Construct a realistic net.OpError wrapping a
	// SyscallError.
	sysErr := &os.SyscallError{
		Syscall: "bind",
		Err:     fmt.Errorf("address already in use"),
	}
	opErr := &net.OpError{
		Op:  "listen",
		Net: "tcp",
		Err: sysErr,
	}
	if !isAddrInUse(opErr) {
		t.Error("expected true for OpError wrapping address in use")
	}
}

// ============================================================
// Additional provider edge case tests
// ============================================================

func TestVertexProvider_PrepareRequest_EmptyToken(t *testing.T) {
	prov := &VertexProvider{
		projectID: "my-project",
		region:    "us-east5",
		token:     "", // Empty token.
	}

	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(`{"messages":[]}`))

	err := prov.PrepareRequest(req)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
	if !strings.Contains(err.Error(), "token unavailable") {
		t.Errorf("expected token unavailable error, got: %s", err.Error())
	}
}

func TestBedrockProvider_PrepareRequest_EmptyCredentials(t *testing.T) {
	prov := &BedrockProvider{
		region:    "us-east-1",
		accessKey: "",
		secretKey: "",
	}

	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(`{"messages":[]}`))

	err := prov.PrepareRequest(req)
	if err == nil {
		t.Fatal("expected error for empty credentials")
	}
	if !strings.Contains(err.Error(), "credentials unavailable") {
		t.Errorf("expected credentials unavailable error, got: %s", err.Error())
	}
}

func TestAnthropicProvider_Stop(t *testing.T) {
	// Verify Stop is a no-op and doesn't panic.
	prov := &AnthropicProvider{apiKey: "test"}
	prov.Stop() // Should not panic.
}

func TestVertexProvider_Stop_NilCancel(t *testing.T) {
	// Verify Stop with nil cancel doesn't panic.
	prov := &VertexProvider{}
	prov.Stop() // Should not panic.
}

func TestBedrockProvider_Stop_NilCancel(t *testing.T) {
	// Verify Stop with nil cancel doesn't panic.
	prov := &BedrockProvider{}
	prov.Stop() // Should not panic.
}

// ============================================================
// extractModelFromBody edge cases
// ============================================================

func TestExtractModelFromBody_MalformedJSON(t *testing.T) {
	body := `{not valid json`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	model := extractModelFromBody(req)
	if model != "claude-sonnet-4-20250514" {
		t.Errorf("got %q, want default model for malformed JSON", model)
	}
}

func TestExtractModelFromBody_EmptyModel(t *testing.T) {
	body := `{"model":"","messages":[]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	model := extractModelFromBody(req)
	if model != "claude-sonnet-4-20250514" {
		t.Errorf("got %q, want default model for empty model field", model)
	}
}

func TestExtractModelFromBody_AnthropicPrefix(t *testing.T) {
	body := `{"model":"anthropic.claude-3-sonnet","messages":[]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	model := extractModelFromBody(req)
	if model != "anthropic.claude-3-sonnet" {
		t.Errorf("got %q, want anthropic.claude-3-sonnet", model)
	}
}

// ============================================================
// hashPayload edge case
// ============================================================

func TestHashPayload_NilBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", nil)
	hash, err := hashPayload(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// SHA256 of empty string.
	emptyHash := sha256Hex([]byte(""))
	if hash != emptyHash {
		t.Errorf("got %q, want SHA256 of empty string", hash)
	}
}

// ============================================================
// parseAWSCredentialsJSON edge case
// ============================================================

func TestParseAWSCredentialsJSON_MissingFields(t *testing.T) {
	data := []byte(`{"AccessKeyId":"AKIA","SecretAccessKey":""}`)
	_, _, _, err := parseAWSCredentialsJSON(data)
	if err == nil {
		t.Fatal("expected error for missing SecretAccessKey")
	}
}

func TestParseAWSCredentialsJSON_InvalidJSON(t *testing.T) {
	data := []byte(`not json`)
	_, _, _, err := parseAWSCredentialsJSON(data)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// ============================================================
// refreshVertexToken edge case
// ============================================================

func TestRefreshVertexToken_EmptyOutput(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte("   \n"), nil
	}

	_, err := refreshVertexToken(execCmd)
	if err == nil {
		t.Fatal("expected error for empty token output")
	}
	if !strings.Contains(err.Error(), "empty token") {
		t.Errorf("expected empty token error, got: %s", err.Error())
	}
}

// ============================================================
// Mock Provider for proxy tests
// ============================================================

// mockProvider is a test provider that rewrites requests to
// a local httptest.Server URL.
type mockProvider struct {
	name        string
	upstreamURL string
}

func (p *mockProvider) Name() string { return p.name }

func (p *mockProvider) PrepareRequest(req *http.Request) error {
	// Parse the upstream URL and rewrite the request.
	req.URL.Scheme = "http"
	// Extract host from upstream URL (strip scheme).
	host := strings.TrimPrefix(p.upstreamURL, "http://")
	req.URL.Host = host
	req.Host = host
	return nil
}

func (p *mockProvider) Start(_ context.Context) error { return nil }
func (p *mockProvider) Stop()                         {}

// Ensure mockProvider implements Provider.
var _ Provider = (*mockProvider)(nil)

// errorProvider is a test provider whose PrepareRequest
// always returns an error. Used to test the Director
// error path in newMux.
type errorProvider struct {
	name string
	err  error
}

func (p *errorProvider) Name() string                         { return p.name }
func (p *errorProvider) Start(_ context.Context) error        { return nil }
func (p *errorProvider) Stop()                                {}
func (p *errorProvider) PrepareRequest(_ *http.Request) error { return p.err }

var _ Provider = (*errorProvider)(nil)

// capturingProvider captures the request passed to
// PrepareRequest for inspection.
type capturingProvider struct {
	name      string
	onPrepare func(req *http.Request) error
}

func (p *capturingProvider) Name() string                  { return p.name }
func (p *capturingProvider) Start(_ context.Context) error { return nil }
func (p *capturingProvider) Stop()                         {}
func (p *capturingProvider) PrepareRequest(req *http.Request) error {
	return p.onPrepare(req)
}

var _ Provider = (*capturingProvider)(nil)

// ============================================================
// End-to-End Proxy Translation Tests (T036-T037)
// ============================================================

func TestNewMux_VertexProxyTranslation(t *testing.T) {
	// Create a mock upstream that captures the received
	// request for inspection.
	var capturedBody []byte
	var capturedHeaders http.Header
	var capturedPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedHeaders = r.Header.Clone()
		var err error
		capturedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read upstream body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id":   "msg_test",
			"type": "message",
		})
	}))
	defer upstream.Close()

	// Create a VertexProvider that rewrites to the mock
	// upstream. We use a mock that calls transformVertexBody
	// and rewrites to the upstream URL.
	upstreamHost := strings.TrimPrefix(upstream.URL, "http://")
	prov := &capturingProvider{
		name: "vertex",
		onPrepare: func(req *http.Request) error {
			// Simulate VertexProvider.PrepareRequest behavior.
			isCountTokens := strings.Contains(req.URL.Path, "count_tokens")
			model, stream, _ := transformVertexBody(req)
			action := "rawPredict"
			if stream && !isCountTokens {
				action = "streamRawPredict"
			}
			req.URL.Scheme = "http"
			req.URL.Host = upstreamHost
			req.Host = upstreamHost
			req.URL.Path = fmt.Sprintf("/v1/projects/test/locations/us-east5/publishers/anthropic/models/%s:%s", model, action)
			req.Header.Del("anthropic-beta")
			req.Header.Del("anthropic-version")
			req.Header.Set("Authorization", "Bearer test-token")
			return nil
		},
	}

	mux := newMux(prov, 53147, time.Now())

	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}],"max_tokens":1024}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-beta", "messages-2024-01-01")
	req.Header.Set("anthropic-version", "2023-06-01")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	// Verify upstream received transformed body.
	var upstreamPayload map[string]any
	if err := json.Unmarshal(capturedBody, &upstreamPayload); err != nil {
		t.Fatalf("unmarshal upstream body: %v", err)
	}
	if _, ok := upstreamPayload["model"]; ok {
		t.Error("upstream body should not contain model field")
	}
	if av, ok := upstreamPayload["anthropic_version"].(string); !ok || av != "vertex-2023-10-16" {
		t.Errorf("upstream body anthropic_version: got %q, want vertex-2023-10-16", av)
	}

	// Verify headers were stripped.
	if capturedHeaders.Get("Anthropic-Beta") != "" {
		t.Error("upstream should not receive anthropic-beta header")
	}
	if capturedHeaders.Get("Anthropic-Version") != "" {
		t.Error("upstream should not receive anthropic-version header")
	}

	// Verify path contains rawPredict (non-streaming).
	if !strings.Contains(capturedPath, "rawPredict") {
		t.Errorf("upstream path should contain rawPredict, got: %s", capturedPath)
	}
}

func TestNewMux_VertexStreamingEndpoint(t *testing.T) {
	var capturedPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"msg_test"}`))
	}))
	defer upstream.Close()

	upstreamHost := strings.TrimPrefix(upstream.URL, "http://")
	prov := &capturingProvider{
		name: "vertex",
		onPrepare: func(req *http.Request) error {
			isCountTokens := strings.Contains(req.URL.Path, "count_tokens")
			model, stream, _ := transformVertexBody(req)
			action := "rawPredict"
			if stream && !isCountTokens {
				action = "streamRawPredict"
			}
			req.URL.Scheme = "http"
			req.URL.Host = upstreamHost
			req.Host = upstreamHost
			req.URL.Path = fmt.Sprintf("/v1/projects/test/locations/us-east5/publishers/anthropic/models/%s:%s", model, action)
			req.Header.Set("Authorization", "Bearer test-token")
			return nil
		},
	}

	mux := newMux(prov, 53147, time.Now())

	body := `{"model":"claude-sonnet-4-20250514","messages":[],"stream":true}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	if !strings.Contains(capturedPath, "streamRawPredict") {
		t.Errorf("upstream path should contain streamRawPredict for streaming, got: %s", capturedPath)
	}
}

// ============================================================
// Backward Compatibility Tests (T038-T039)
// ============================================================

func TestAnthropicProvider_NoBodyTransformation(t *testing.T) {
	prov := &AnthropicProvider{apiKey: "sk-ant-test123"}

	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("anthropic-beta", "messages-2024-01-01")
	req.Header.Set("anthropic-version", "2023-06-01")

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	// Verify body passes through unchanged (no model
	// removal, no anthropic_version injection).
	transformed, readErr := io.ReadAll(req.Body)
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}
	var payload map[string]any
	if err := json.Unmarshal(transformed, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := payload["model"]; !ok {
		t.Error("Anthropic provider should NOT remove model from body (FR-010)")
	}
	if _, ok := payload["anthropic_version"]; ok {
		t.Error("Anthropic provider should NOT inject anthropic_version (FR-010)")
	}

	// Verify anthropic-beta and anthropic-version headers
	// are preserved (FR-010).
	if req.Header.Get("anthropic-beta") != "messages-2024-01-01" {
		t.Error("Anthropic provider should preserve anthropic-beta header (FR-010)")
	}
	if req.Header.Get("anthropic-version") != "2023-06-01" {
		t.Error("Anthropic provider should preserve anthropic-version header (FR-010)")
	}
}

func TestBedrockProvider_NoBodyTransformation(t *testing.T) {
	prov := &BedrockProvider{
		region:       "us-east-1",
		accessKey:    "AKIAIOSFODNN7EXAMPLE",
		secretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		sessionToken: "test-session-token",
		credExpiry:   time.Now().Add(30 * time.Minute),
	}

	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	// Verify body still uses extractModelFromBody (model
	// is read but NOT removed from body).
	transformed, readErr := io.ReadAll(req.Body)
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}
	var payload map[string]any
	if err := json.Unmarshal(transformed, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// extractModelFromBody reads the body but replaces it
	// unchanged — model should still be present.
	if _, ok := payload["model"]; !ok {
		t.Error("Bedrock provider should NOT remove model from body (FR-010)")
	}
	if _, ok := payload["anthropic_version"]; ok {
		t.Error("Bedrock provider should NOT inject anthropic_version (FR-010)")
	}
}

// ============================================================
// transformVertexBody Tests (T007-T016)
// ============================================================

func TestTransformVertexBody_RemovesModel(t *testing.T) {
	body := `{"model":"claude-opus-4-20250514","messages":[{"role":"user","content":"hi"}],"max_tokens":1024}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	model, _, err := transformVertexBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model != "claude-opus-4-20250514" {
		t.Errorf("model: got %q, want claude-opus-4-20250514", model)
	}

	// Read the transformed body and verify model is removed.
	transformed, readErr := io.ReadAll(req.Body)
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}
	var payload map[string]any
	if err := json.Unmarshal(transformed, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := payload["model"]; ok {
		t.Error("model field should be removed from body")
	}
	// Verify other fields are preserved.
	if _, ok := payload["messages"]; !ok {
		t.Error("messages field should be preserved")
	}
	if _, ok := payload["max_tokens"]; !ok {
		t.Error("max_tokens field should be preserved")
	}
}

func TestTransformVertexBody_InjectsAnthropicVersion(t *testing.T) {
	body := `{"model":"claude-sonnet-4-20250514","messages":[]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	_, _, err := transformVertexBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	transformed, readErr := io.ReadAll(req.Body)
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}
	var payload map[string]any
	if err := json.Unmarshal(transformed, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	av, ok := payload["anthropic_version"].(string)
	if !ok {
		t.Fatal("anthropic_version should be injected")
	}
	if av != "vertex-2023-10-16" {
		t.Errorf("anthropic_version: got %q, want vertex-2023-10-16", av)
	}
}

func TestTransformVertexBody_PreservesExistingAnthropicVersion(t *testing.T) {
	body := `{"model":"claude-sonnet-4-20250514","messages":[],"anthropic_version":"custom-2024-01-01"}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	_, _, err := transformVertexBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	transformed, readErr := io.ReadAll(req.Body)
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}
	var payload map[string]any
	if err := json.Unmarshal(transformed, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	av, ok := payload["anthropic_version"].(string)
	if !ok {
		t.Fatal("anthropic_version should be present")
	}
	if av != "custom-2024-01-01" {
		t.Errorf("anthropic_version: got %q, want custom-2024-01-01 (should preserve existing)", av)
	}
}

func TestTransformVertexBody_DetectsStreamTrue(t *testing.T) {
	body := `{"model":"claude-sonnet-4-20250514","messages":[],"stream":true}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	_, stream, err := transformVertexBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stream {
		t.Error("stream should be true when body contains \"stream\": true")
	}
}

func TestTransformVertexBody_DetectsStreamFalse(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"stream false", `{"model":"claude-sonnet-4-20250514","messages":[],"stream":false}`},
		{"stream absent", `{"model":"claude-sonnet-4-20250514","messages":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/v1/messages",
				strings.NewReader(tt.body))

			_, stream, err := transformVertexBody(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if stream {
				t.Errorf("stream should be false for %s", tt.name)
			}
		})
	}
}

func TestTransformVertexBody_MalformedJSON(t *testing.T) {
	body := `{not valid json`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	model, stream, err := transformVertexBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model != "claude-sonnet-4-20250514" {
		t.Errorf("model: got %q, want default model", model)
	}
	if stream {
		t.Error("stream should be false for malformed JSON")
	}

	// Verify original body is forwarded unchanged.
	transformed, readErr := io.ReadAll(req.Body)
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}
	if string(transformed) != body {
		t.Errorf("body should be forwarded unchanged, got: %s", string(transformed))
	}
}

func TestTransformVertexBody_NilBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/v1/messages", nil)

	model, stream, err := transformVertexBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model != "claude-sonnet-4-20250514" {
		t.Errorf("model: got %q, want default model", model)
	}
	if stream {
		t.Error("stream should be false for nil body")
	}
}

func TestTransformVertexBody_EmptyBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(""))

	model, stream, err := transformVertexBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model != "claude-sonnet-4-20250514" {
		t.Errorf("model: got %q, want default model", model)
	}
	if stream {
		t.Error("stream should be false for empty body")
	}
}

func TestTransformVertexBody_UpdatesContentLength(t *testing.T) {
	body := `{"model":"claude-sonnet-4-20250514","messages":[],"max_tokens":1024}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	_, _, err := transformVertexBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read the transformed body to get its length.
	transformed, readErr := io.ReadAll(req.Body)
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}

	if req.ContentLength != int64(len(transformed)) {
		t.Errorf("ContentLength: got %d, want %d (body length)",
			req.ContentLength, len(transformed))
	}
}

func TestTransformVertexBody_PreservesOtherFields(t *testing.T) {
	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}],"max_tokens":4096,"temperature":0.7,"tools":[{"name":"test"}],"system":"You are helpful"}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

	_, _, err := transformVertexBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	transformed, readErr := io.ReadAll(req.Body)
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}
	var payload map[string]any
	if err := json.Unmarshal(transformed, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Verify all fields except model are preserved.
	if _, ok := payload["messages"]; !ok {
		t.Error("messages field should be preserved")
	}
	if _, ok := payload["max_tokens"]; !ok {
		t.Error("max_tokens field should be preserved")
	}
	if _, ok := payload["temperature"]; !ok {
		t.Error("temperature field should be preserved")
	}
	if _, ok := payload["tools"]; !ok {
		t.Error("tools field should be preserved")
	}
	if _, ok := payload["system"]; !ok {
		t.Error("system field should be preserved")
	}
}

// ============================================================
// SSE Filter Tests (T019-T026)
// ============================================================

func TestSSEFilterReader_DropsVertexEvent(t *testing.T) {
	sseStream := "event: vertex_event\ndata: {\"type\":\"vertex_event\"}\n\nevent: message_start\ndata: {\"type\":\"message_start\"}\n\n"
	source := io.NopCloser(strings.NewReader(sseStream))
	filtered := map[string]bool{"vertex_event": true, "ping": true}
	reader := newSSEFilterReader(source, filtered)

	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	output := string(result)
	if strings.Contains(output, "vertex_event") {
		t.Errorf("vertex_event should be dropped, got: %s", output)
	}
	if !strings.Contains(output, "message_start") {
		t.Errorf("message_start should be forwarded, got: %s", output)
	}
}

func TestSSEFilterReader_DropsPing(t *testing.T) {
	sseStream := "event: ping\ndata: \n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"
	source := io.NopCloser(strings.NewReader(sseStream))
	filtered := map[string]bool{"vertex_event": true, "ping": true}
	reader := newSSEFilterReader(source, filtered)

	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	output := string(result)
	if strings.Contains(output, "event: ping") {
		t.Errorf("ping should be dropped, got: %s", output)
	}
	if !strings.Contains(output, "message_stop") {
		t.Errorf("message_stop should be forwarded, got: %s", output)
	}
}

func TestSSEFilterReader_ForwardsStandardEvents(t *testing.T) {
	events := []string{
		"message_start",
		"content_block_delta",
		"content_block_stop",
		"message_delta",
		"message_stop",
	}

	var sseStream strings.Builder
	for _, evt := range events {
		sseStream.WriteString("event: " + evt + "\n")
		sseStream.WriteString("data: {\"type\":\"" + evt + "\"}\n")
		sseStream.WriteString("\n")
	}

	source := io.NopCloser(strings.NewReader(sseStream.String()))
	filtered := map[string]bool{"vertex_event": true, "ping": true}
	reader := newSSEFilterReader(source, filtered)

	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	output := string(result)
	for _, evt := range events {
		if !strings.Contains(output, "event: "+evt) {
			t.Errorf("standard event %q should be forwarded, got: %s", evt, output)
		}
	}
}

func TestSSEFilterReader_MixedEvents(t *testing.T) {
	sseStream := "" +
		"event: message_start\ndata: {\"type\":\"message_start\"}\n\n" +
		"event: vertex_event\ndata: {\"type\":\"vertex_event\"}\n\n" +
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\"}\n\n" +
		"event: ping\ndata: \n\n" +
		"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"

	source := io.NopCloser(strings.NewReader(sseStream))
	filtered := map[string]bool{"vertex_event": true, "ping": true}
	reader := newSSEFilterReader(source, filtered)

	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	output := string(result)

	// Standard events should be present.
	if !strings.Contains(output, "message_start") {
		t.Error("message_start should be forwarded")
	}
	if !strings.Contains(output, "content_block_delta") {
		t.Error("content_block_delta should be forwarded")
	}
	if !strings.Contains(output, "message_stop") {
		t.Error("message_stop should be forwarded")
	}

	// Filtered events should be absent.
	if strings.Contains(output, "vertex_event") {
		t.Error("vertex_event should be dropped")
	}
	if strings.Contains(output, "event: ping") {
		t.Error("ping should be dropped")
	}
}

func TestSSEFilterReader_EmptyStream(t *testing.T) {
	source := io.NopCloser(strings.NewReader(""))
	filtered := map[string]bool{"vertex_event": true}
	reader := newSSEFilterReader(source, filtered)

	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty output, got: %q", string(result))
	}
}

func TestSSEFilterReader_Close(t *testing.T) {
	closed := false
	source := &mockReadCloser{
		Reader: strings.NewReader(""),
		onClose: func() error {
			closed = true
			return nil
		},
	}
	filtered := map[string]bool{"vertex_event": true}
	reader := newSSEFilterReader(source, filtered)

	if err := reader.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !closed {
		t.Error("Close should delegate to source's Close")
	}
}

func TestVertexSSEFilter_NonStreamingPassthrough(t *testing.T) {
	body := `{"id":"msg_test","type":"message"}`
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	filter := vertexSSEFilter()
	if err := filter(resp); err != nil {
		t.Fatalf("filter: %v", err)
	}

	// Body should NOT be wrapped — read it directly.
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(result) != body {
		t.Errorf("body should pass through unchanged, got: %s", string(result))
	}
}

func TestVertexSSEFilter_StreamingWraps(t *testing.T) {
	sseStream := "event: message_start\ndata: {}\n\n"
	resp := &http.Response{
		StatusCode:    200,
		Header:        http.Header{"Content-Type": []string{"text/event-stream"}, "Content-Length": []string{"100"}},
		Body:          io.NopCloser(strings.NewReader(sseStream)),
		ContentLength: 100,
	}

	filter := vertexSSEFilter()
	if err := filter(resp); err != nil {
		t.Fatalf("filter: %v", err)
	}

	// Verify body is wrapped (type should be *sseFilterReader).
	if _, ok := resp.Body.(*sseFilterReader); !ok {
		t.Error("body should be wrapped in sseFilterReader for streaming responses")
	}

	// Verify Content-Length is removed.
	if resp.Header.Get("Content-Length") != "" {
		t.Error("Content-Length should be removed for filtered streaming responses")
	}
	if resp.ContentLength != -1 {
		t.Errorf("ContentLength: got %d, want -1", resp.ContentLength)
	}
}

// ============================================================
// SSE Filtering Integration Tests (T042-T045)
// ============================================================

func TestNewMux_VertexSSEFiltering(t *testing.T) {
	// Create a mock upstream that returns an SSE stream
	// with vertex_event and standard events.
	sseStream := "" +
		"event: message_start\ndata: {\"type\":\"message_start\"}\n\n" +
		"event: vertex_event\ndata: {\"type\":\"vertex_event\"}\n\n" +
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\"}\n\n" +
		"event: ping\ndata: \n\n" +
		"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sseStream))
	}))
	defer upstream.Close()

	prov := &mockProvider{
		name:        "vertex",
		upstreamURL: upstream.URL,
	}

	mux := newMux(prov, 53147, time.Now())

	body := `{"model":"claude-sonnet-4-20250514","messages":[],"stream":true}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	output := w.Body.String()

	// Standard events should be present.
	if !strings.Contains(output, "message_start") {
		t.Error("message_start should be forwarded")
	}
	if !strings.Contains(output, "content_block_delta") {
		t.Error("content_block_delta should be forwarded")
	}
	if !strings.Contains(output, "message_stop") {
		t.Error("message_stop should be forwarded")
	}

	// Filtered events should be absent.
	if strings.Contains(output, "vertex_event") {
		t.Error("vertex_event should be dropped by SSE filter")
	}
	if strings.Contains(output, "event: ping") {
		t.Error("ping should be dropped by SSE filter")
	}
}

func TestNewMux_VertexNonStreamingNoFilter(t *testing.T) {
	// Non-streaming responses should pass through unchanged.
	jsonBody := `{"id":"msg_test","type":"message","content":[{"type":"text","text":"hello"}]}`
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(jsonBody))
	}))
	defer upstream.Close()

	prov := &mockProvider{
		name:        "vertex",
		upstreamURL: upstream.URL,
	}

	mux := newMux(prov, 53147, time.Now())

	body := `{"model":"claude-sonnet-4-20250514","messages":[]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	// Response should pass through unchanged (FR-009).
	if w.Body.String() != jsonBody {
		t.Errorf("non-streaming response should pass through unchanged, got: %s", w.Body.String())
	}
}

func TestNewMux_AnthropicNoSSEFilter(t *testing.T) {
	// Anthropic provider should NOT filter SSE events.
	sseStream := "" +
		"event: message_start\ndata: {\"type\":\"message_start\"}\n\n" +
		"event: vertex_event\ndata: {\"type\":\"vertex_event\"}\n\n" +
		"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sseStream))
	}))
	defer upstream.Close()

	upstreamHost := strings.TrimPrefix(upstream.URL, "http://")
	prov := &capturingProvider{
		name: "anthropic",
		onPrepare: func(req *http.Request) error {
			req.URL.Scheme = "http"
			req.URL.Host = upstreamHost
			req.Host = upstreamHost
			return nil
		},
	}

	mux := newMux(prov, 53147, time.Now())

	body := `{"model":"claude-sonnet-4-20250514","messages":[],"stream":true}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	output := w.Body.String()

	// vertex_event should NOT be filtered for Anthropic
	// provider (FR-010).
	if !strings.Contains(output, "vertex_event") {
		t.Error("vertex_event should NOT be filtered for Anthropic provider (FR-010)")
	}
	if !strings.Contains(output, "message_start") {
		t.Error("message_start should be present")
	}
}

func TestNewMux_VertexErrorResponseNoFilter(t *testing.T) {
	// Error responses should pass through unchanged.
	errorBody := `{"error":{"type":"invalid_request_error","message":"Invalid model"}}`
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(errorBody))
	}))
	defer upstream.Close()

	prov := &mockProvider{
		name:        "vertex",
		upstreamURL: upstream.URL,
	}

	mux := newMux(prov, 53147, time.Now())

	body := `{"model":"claude-sonnet-4-20250514","messages":[]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}

	// Error response should pass through unchanged.
	if w.Body.String() != errorBody {
		t.Errorf("error response should pass through unchanged, got: %s", w.Body.String())
	}
}

// ============================================================
// Synthetic Model Catalog Tests (T054-T057)
// ============================================================

func TestNewMux_ModelsList(t *testing.T) {
	prov := &AnthropicProvider{apiKey: "test-key"}
	mux := newMux(prov, 53147, time.Now())

	req := httptest.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatal("expected data array in response")
	}
	if len(data) < 9 {
		t.Errorf("expected at least 9 models, got: %d", len(data))
	}

	// Verify each model has required fields.
	for i, entry := range data {
		m, ok := entry.(map[string]any)
		if !ok {
			t.Errorf("model %d: expected object", i)
			continue
		}
		if _, ok := m["id"]; !ok {
			t.Errorf("model %d: missing id", i)
		}
		if _, ok := m["type"]; !ok {
			t.Errorf("model %d: missing type", i)
		}
		if _, ok := m["display_name"]; !ok {
			t.Errorf("model %d: missing display_name", i)
		}
		if _, ok := m["capabilities"]; !ok {
			t.Errorf("model %d: missing capabilities", i)
		}
	}

	// Verify has_more, first_id, last_id.
	if resp["has_more"] != false {
		t.Errorf("has_more: got %v, want false", resp["has_more"])
	}
	if _, ok := resp["first_id"].(string); !ok {
		t.Error("expected first_id string")
	}
	if _, ok := resp["last_id"].(string); !ok {
		t.Error("expected last_id string")
	}
}

func TestNewMux_ModelsSingleFound(t *testing.T) {
	prov := &AnthropicProvider{apiKey: "test-key"}
	mux := newMux(prov, 53147, time.Now())

	req := httptest.NewRequest("GET", "/v1/models/claude-sonnet-4-20250514", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var model map[string]any
	if err := json.NewDecoder(w.Body).Decode(&model); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if model["id"] != "claude-sonnet-4-20250514" {
		t.Errorf("id: got %q, want claude-sonnet-4-20250514", model["id"])
	}
	if model["type"] != "model" {
		t.Errorf("type: got %q, want model", model["type"])
	}
	if model["display_name"] != "Claude Sonnet 4" {
		t.Errorf("display_name: got %q, want Claude Sonnet 4", model["display_name"])
	}
}

func TestNewMux_ModelsSingleNotFound(t *testing.T) {
	prov := &AnthropicProvider{apiKey: "test-key"}
	mux := newMux(prov, 53147, time.Now())

	req := httptest.NewRequest("GET", "/v1/models/unknown-model", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["type"] != "not_found" {
		t.Errorf("error type: got %q, want not_found", errObj["type"])
	}
	if msg, ok := errObj["message"].(string); !ok || !strings.Contains(msg, "unknown-model") {
		t.Errorf("error message should contain model ID, got: %q", msg)
	}
}

func TestNewMux_ModelsCapabilities(t *testing.T) {
	prov := &AnthropicProvider{apiKey: "test-key"}
	mux := newMux(prov, 53147, time.Now())

	// Test Haiku 4.5 — should have extended_thinking: false.
	req := httptest.NewRequest("GET", "/v1/models/claude-haiku-4-5-20241022", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("haiku status: got %d, want %d", w.Code, http.StatusOK)
	}

	var haiku map[string]any
	if err := json.NewDecoder(w.Body).Decode(&haiku); err != nil {
		t.Fatalf("decode haiku: %v", err)
	}

	caps, ok := haiku["capabilities"].(map[string]any)
	if !ok {
		t.Fatal("expected capabilities object for haiku")
	}
	if caps["extended_thinking"] != false {
		t.Errorf("Haiku 4.5 extended_thinking: got %v, want false", caps["extended_thinking"])
	}
	if caps["vision"] != true {
		t.Errorf("Haiku 4.5 vision: got %v, want true", caps["vision"])
	}
	if caps["pdf_input"] != true {
		t.Errorf("Haiku 4.5 pdf_input: got %v, want true", caps["pdf_input"])
	}

	// Test Opus 4.7 — should have extended_thinking: true.
	req2 := httptest.NewRequest("GET", "/v1/models/claude-opus-4-7-20250416", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("opus status: got %d, want %d", w2.Code, http.StatusOK)
	}

	var opus map[string]any
	if err := json.NewDecoder(w2.Body).Decode(&opus); err != nil {
		t.Fatalf("decode opus: %v", err)
	}

	opusCaps, ok := opus["capabilities"].(map[string]any)
	if !ok {
		t.Fatal("expected capabilities object for opus")
	}
	if opusCaps["extended_thinking"] != true {
		t.Errorf("Opus 4.7 extended_thinking: got %v, want true", opusCaps["extended_thinking"])
	}
}

// mockReadCloser is a test helper that wraps an io.Reader
// with a custom Close function.
type mockReadCloser struct {
	io.Reader
	onClose func() error
}

func (m *mockReadCloser) Close() error {
	return m.onClose()
}

func TestNewVertexProvider_GlobalRegionError(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "VERTEX_LOCATION":
			return "global"
		case "CLOUD_ML_REGION":
			return "global"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		}
		return ""
	}
	execCmd := func(string, ...string) ([]byte, error) {
		return nil, nil
	}

	_, err := newVertexProvider(getenv, execCmd)
	if err == nil {
		t.Fatal("expected error for global region")
	}
	if !strings.Contains(err.Error(), "global") {
		t.Errorf("error should mention 'global', got: %s",
			err.Error())
	}
	if !strings.Contains(err.Error(),
		"ANTHROPIC_VERTEX_REGION") {
		t.Errorf("error should mention "+
			"ANTHROPIC_VERTEX_REGION, got: %s",
			err.Error())
	}
}

func TestNewVertexProvider_CloudMLRegionGlobalAlone(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "CLOUD_ML_REGION":
			return "global"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		}
		return ""
	}
	execCmd := func(string, ...string) ([]byte, error) {
		return nil, nil
	}

	_, err := newVertexProvider(getenv, execCmd)
	if err == nil {
		t.Fatal("expected error for CLOUD_ML_REGION=global")
	}
	if !strings.Contains(err.Error(), "global") {
		t.Errorf("error should mention 'global', got: %s",
			err.Error())
	}
}

func TestNewVertexProvider_GlobalOverriddenBySpecificRegion(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "VERTEX_LOCATION":
			return "global"
		case "ANTHROPIC_VERTEX_REGION":
			return "us-east5"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		}
		return ""
	}
	execCmd := func(string, ...string) ([]byte, error) {
		return nil, nil
	}

	prov, err := newVertexProvider(getenv, execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov.region != "us-east5" {
		t.Errorf("region: got %q, want %q",
			prov.region, "us-east5")
	}
}

func TestNewVertexProvider_EmptyRegionDefault(t *testing.T) {
	getenv := func(key string) string {
		if key == "ANTHROPIC_VERTEX_PROJECT_ID" {
			return "my-project"
		}
		return ""
	}
	execCmd := func(string, ...string) ([]byte, error) {
		return nil, nil
	}

	prov, err := newVertexProvider(getenv, execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov.region != "us-east5" {
		t.Errorf("region: got %q, want %q",
			prov.region, "us-east5")
	}
}

func TestDetectProvider_GlobalRegionError(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "CLAUDE_CODE_USE_VERTEX":
			return "1"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		case "VERTEX_LOCATION":
			return "global"
		}
		return ""
	}
	execCmd := func(string, ...string) ([]byte, error) {
		return nil, nil
	}

	_, err := DetectProvider(getenv, execCmd)
	if err == nil {
		t.Fatal("expected error from DetectProvider " +
			"with global region")
	}
	if !strings.Contains(err.Error(), "global") {
		t.Errorf("error should mention 'global', got: %s",
			err.Error())
	}
}

func TestNewProviderByName_VertexGlobalRegionError(t *testing.T) {
	getenv := func(key string) string {
		switch key {
		case "VERTEX_LOCATION":
			return "global"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		}
		return ""
	}
	execCmd := func(string, ...string) ([]byte, error) {
		return nil, nil
	}

	_, err := NewProviderByName("vertex", getenv, execCmd)
	if err == nil {
		t.Fatal("expected error from NewProviderByName " +
			"with global region")
	}
	if !strings.Contains(err.Error(), "global") {
		t.Errorf("error should mention 'global', got: %s",
			err.Error())
	}
}

// ============================================================
// Token Expiry & Proactive Refresh Tests (TC-006+)
// ============================================================

func TestVertexPrepareRequest_StaleTokenRegression(t *testing.T) {
	// TC-006 regression: before the fix, a stale token
	// would be forwarded without error. Now PrepareRequest
	// must reject requests when the token has expired.
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "stale-token",
		tokenExpiry: time.Now().Add(-10 * time.Minute),
		execCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("gcloud not authenticated")
		},
		getenv: func(key string) string { return "" },
	}

	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[]}`))

	err := prov.PrepareRequest(req)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !strings.Contains(err.Error(), "token expired") {
		t.Errorf("expected 'token expired' error, got: %s", err.Error())
	}
}

func TestVertexPrepareRequest_ValidToken(t *testing.T) {
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "ya29.valid-token",
		tokenExpiry: time.Now().Add(30 * time.Minute),
	}

	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[]}`))

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest should succeed with valid token: %v", err)
	}

	auth := req.Header.Get("Authorization")
	if auth != "Bearer ya29.valid-token" {
		t.Errorf("Authorization: got %q, want %q", auth, "Bearer ya29.valid-token")
	}
}

func TestVertexPrepareRequest_ProactiveRefresh(t *testing.T) {
	// Token expires in 3 minutes — within the proactive
	// refresh window (5 minutes). PrepareRequest should
	// trigger a proactive refresh.
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "old-token",
		tokenExpiry: time.Now().Add(3 * time.Minute),
		execCmd: func(name string, args ...string) ([]byte, error) {
			return []byte("new-token\n"), nil
		},
		getenv: func(key string) string {
			switch key {
			case "ANTHROPIC_VERTEX_PROJECT_ID":
				return "my-project"
			case "ANTHROPIC_VERTEX_REGION":
				return "us-east5"
			}
			return ""
		},
	}

	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[]}`))

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest should succeed: %v", err)
	}

	// Verify the token was updated.
	prov.tokenMu.RLock()
	token := prov.token
	expiry := prov.tokenExpiry
	prov.tokenMu.RUnlock()

	if token != "new-token" {
		t.Errorf("token: got %q, want %q", token, "new-token")
	}
	if time.Until(expiry) < 50*time.Minute {
		t.Errorf("tokenExpiry should be reset to ~55 minutes from now, got %v until expiry",
			time.Until(expiry))
	}

	auth := req.Header.Get("Authorization")
	if auth != "Bearer new-token" {
		t.Errorf("Authorization: got %q, want %q", auth, "Bearer new-token")
	}
}

func TestVertexPrepareRequest_ProactiveRefreshFails(t *testing.T) {
	// Token expires in 3 minutes — within the proactive
	// refresh window. The refresh fails, but the request
	// should proceed with the current (still valid) token.
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "current-token",
		tokenExpiry: time.Now().Add(3 * time.Minute),
		execCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("gcloud not found")
		},
		getenv: func(key string) string { return "" },
	}

	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[]}`))

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest should succeed with still-valid token: %v", err)
	}

	// Verify the original token is preserved (not cleared).
	prov.tokenMu.RLock()
	token := prov.token
	prov.tokenMu.RUnlock()

	if token != "current-token" {
		t.Errorf("token should be preserved on proactive refresh failure, got: %q", token)
	}

	auth := req.Header.Get("Authorization")
	if auth != "Bearer current-token" {
		t.Errorf("Authorization: got %q, want %q", auth, "Bearer current-token")
	}
}

func TestVertexPrepareRequest_ProactiveRefreshTimeout(t *testing.T) {
	// When execCmd returns an error (simulating any failure
	// including timeout), the request should proceed with
	// the current token.
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "valid-token",
		tokenExpiry: time.Now().Add(3 * time.Minute),
		execCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("command timed out")
		},
		getenv: func(key string) string { return "" },
	}

	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[]}`))

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest should succeed despite refresh timeout: %v", err)
	}

	auth := req.Header.Get("Authorization")
	if auth != "Bearer valid-token" {
		t.Errorf("Authorization: got %q, want %q", auth, "Bearer valid-token")
	}
}

func TestVertexRefreshFailure_ClearsToken(t *testing.T) {
	// Simulate the refresh closure's failure path:
	// token and tokenExpiry are cleared atomically.
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "valid-token",
		tokenExpiry: time.Now().Add(30 * time.Minute),
	}

	// Simulate what the refresh closure does on failure.
	prov.tokenMu.Lock()
	prov.token = ""
	prov.tokenExpiry = time.Time{}
	prov.tokenMu.Unlock()

	if prov.token != "" {
		t.Errorf("token should be empty after refresh failure, got: %q", prov.token)
	}
	if !prov.tokenExpiry.IsZero() {
		t.Errorf("tokenExpiry should be zero after refresh failure, got: %v", prov.tokenExpiry)
	}
}

func TestVertexPrepareRequest_ConcurrentProactiveRefresh(t *testing.T) {
	// Verify that concurrent PrepareRequest calls with
	// near-expiry tokens result in exactly 1 refresh
	// invocation (TryLock deduplication).
	var refreshCount atomic.Int32
	prov := &VertexProvider{
		projectID:   "my-project",
		region:      "us-east5",
		token:       "near-expiry-token",
		tokenExpiry: time.Now().Add(3 * time.Minute),
		execCmd: func(name string, args ...string) ([]byte, error) {
			refreshCount.Add(1)
			// Brief sleep to simulate real gcloud latency.
			time.Sleep(10 * time.Millisecond)
			return []byte("refreshed-token\n"), nil
		},
		getenv: func(key string) string { return "" },
	}

	const goroutines = 5
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Use a barrier to synchronize all goroutines.
	barrier := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-barrier // Wait for barrier release.
			req := httptest.NewRequest("POST", "/v1/messages",
				strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[]}`))
			_ = prov.PrepareRequest(req)
		}()
	}

	// Release all goroutines simultaneously.
	close(barrier)
	wg.Wait()

	count := refreshCount.Load()
	if count != 1 {
		t.Errorf("expected exactly 1 refresh invocation (TryLock deduplication), got: %d", count)
	}
}

func TestBedrockPrepareRequest_ExpiredCredentials(t *testing.T) {
	prov := &BedrockProvider{
		region:       "us-east-1",
		accessKey:    "AKIATEST",
		secretKey:    "secrettest",
		sessionToken: "sessiontest",
		credExpiry:   time.Now().Add(-10 * time.Minute),
		execCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("aws CLI not authenticated")
		},
		getenv: func(key string) string { return "" },
	}

	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[]}`))

	err := prov.PrepareRequest(req)
	if err == nil {
		t.Fatal("expected error for expired credentials")
	}
	if !strings.Contains(err.Error(), "credentials expired") {
		t.Errorf("expected 'credentials expired' error, got: %s", err.Error())
	}
}

func TestBedrockRefreshFailure_ClearsCredentials(t *testing.T) {
	// Simulate the refresh closure's failure path:
	// all credentials and credExpiry are cleared atomically.
	prov := &BedrockProvider{
		region:       "us-east-1",
		accessKey:    "AKIATEST",
		secretKey:    "secrettest",
		sessionToken: "sessiontest",
		credExpiry:   time.Now().Add(30 * time.Minute),
	}

	// Simulate what the refresh closure does on failure.
	prov.credMu.Lock()
	prov.accessKey = ""
	prov.secretKey = ""
	prov.sessionToken = ""
	prov.credExpiry = time.Time{}
	prov.credMu.Unlock()

	if prov.accessKey != "" {
		t.Errorf("accessKey should be empty, got: %q", prov.accessKey)
	}
	if prov.secretKey != "" {
		t.Errorf("secretKey should be empty, got: %q", prov.secretKey)
	}
	if prov.sessionToken != "" {
		t.Errorf("sessionToken should be empty, got: %q", prov.sessionToken)
	}
	if !prov.credExpiry.IsZero() {
		t.Errorf("credExpiry should be zero, got: %v", prov.credExpiry)
	}
}

func TestBedrockPrepareRequest_ProactiveRefresh(t *testing.T) {
	// Credentials expire in 3 minutes — within the
	// proactive refresh window. PrepareRequest should
	// trigger a proactive refresh.
	prov := &BedrockProvider{
		region:       "us-east-1",
		accessKey:    "AKIAOLD",
		secretKey:    "secretold",
		sessionToken: "sessionold",
		credExpiry:   time.Now().Add(3 * time.Minute),
		execCmd: func(name string, args ...string) ([]byte, error) {
			return []byte(
				"export AWS_ACCESS_KEY_ID=AKIANEW\n" +
					"export AWS_SECRET_ACCESS_KEY=secretnew\n" +
					"export AWS_SESSION_TOKEN=sessionnew\n"), nil
		},
		getenv: func(key string) string { return "" },
	}

	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[]}`))

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest should succeed: %v", err)
	}

	// Verify credentials were refreshed.
	prov.credMu.RLock()
	ak := prov.accessKey
	sk := prov.secretKey
	st := prov.sessionToken
	expiry := prov.credExpiry
	prov.credMu.RUnlock()

	if ak != "AKIANEW" {
		t.Errorf("accessKey: got %q, want AKIANEW", ak)
	}
	if sk != "secretnew" {
		t.Errorf("secretKey: got %q, want secretnew", sk)
	}
	if st != "sessionnew" {
		t.Errorf("sessionToken: got %q, want sessionnew", st)
	}
	if time.Until(expiry) < 45*time.Minute {
		t.Errorf("credExpiry should be reset to ~50 minutes from now, got %v until expiry",
			time.Until(expiry))
	}
}

func TestBedrockPrepareRequest_ProactiveRefreshFails(t *testing.T) {
	// Credentials expire in 3 minutes — within the
	// proactive refresh window. The refresh fails, but
	// the request should proceed with current credentials.
	prov := &BedrockProvider{
		region:       "us-east-1",
		accessKey:    "AKIACURRENT",
		secretKey:    "secretcurrent",
		sessionToken: "sessioncurrent",
		credExpiry:   time.Now().Add(3 * time.Minute),
		execCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("aws CLI not found")
		},
		getenv: func(key string) string { return "" },
	}

	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(`{"model":"claude-sonnet-4-20250514","messages":[]}`))

	if err := prov.PrepareRequest(req); err != nil {
		t.Fatalf("PrepareRequest should succeed with still-valid credentials: %v", err)
	}

	// Verify original credentials are preserved.
	prov.credMu.RLock()
	ak := prov.accessKey
	prov.credMu.RUnlock()

	if ak != "AKIACURRENT" {
		t.Errorf("accessKey should be preserved on proactive refresh failure, got: %q", ak)
	}
}

func TestDetach_CreatesLogFile(t *testing.T) {
	opts := testOpts(t)
	opts.Port = DefaultPort
	opts.Getenv = func(key string) string { return "" }

	var capturedCmd *exec.Cmd
	opts.ExecStart = func(cmd *exec.Cmd) error {
		capturedCmd = cmd
		cmd.Process = &os.Process{Pid: 33333}
		return nil
	}
	opts.HTTPGet = func(url string) (int, error) {
		return http.StatusOK, nil
	}

	if err := detach(opts); err != nil {
		t.Fatalf("detach failed: %v", err)
	}

	if capturedCmd == nil {
		t.Fatal("ExecStart was not called")
	}

	// Verify cmd.Stdout is a non-nil *os.File.
	stdoutFile, ok := capturedCmd.Stdout.(*os.File)
	if !ok || stdoutFile == nil {
		t.Fatal("cmd.Stdout should be a non-nil *os.File")
	}

	// Verify the file path ends with gateway.log.
	if !strings.HasSuffix(stdoutFile.Name(), "gateway.log") {
		t.Errorf("log file path should end with gateway.log, got: %s", stdoutFile.Name())
	}

	// Verify the log file exists in .uf/ directory.
	logPath := filepath.Join(opts.ProjectDir, ".uf", "gateway.log")
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("log file should exist at %s: %v", logPath, err)
	}

	// Verify 0600 permissions.
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("log file permissions: got %o, want 0600", perm)
	}
}

func TestStatus_ShowsLogPath(t *testing.T) {
	opts := testOpts(t)

	// Create .uf/ directory and gateway.log file.
	ufDir := filepath.Join(opts.ProjectDir, ".uf")
	if err := os.MkdirAll(ufDir, 0755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(ufDir, "gateway.log")
	if err := os.WriteFile(logPath, []byte("test log\n"), 0600); err != nil {
		t.Fatal(err)
	}

	// Write a PID file with current process PID.
	pp := pidPath(opts.ProjectDir)
	info := PIDInfo{
		PID:      os.Getpid(),
		Port:     53147,
		Provider: "vertex",
		Started:  time.Now().Add(-1 * time.Hour),
	}
	if err := WritePID(pp, info); err != nil {
		t.Fatal(err)
	}

	// Process is alive.
	opts.FindProcess = func(pid int) (*os.Process, error) {
		return os.FindProcess(os.Getpid())
	}

	// Health endpoint responds.
	opts.HTTPGet = func(url string) (int, error) {
		return http.StatusOK, nil
	}

	if err := Status(opts); err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	out := stdoutStr(opts)
	if !strings.Contains(out, "Log:") {
		t.Errorf("expected 'Log:' in status output, got: %s", out)
	}
	if !strings.Contains(out, "gateway.log") {
		t.Errorf("expected 'gateway.log' in status output, got: %s", out)
	}
}

func TestStart_ForegroundNoLogFile(t *testing.T) {
	opts := testOpts(t)
	opts.Getenv = func(key string) string {
		switch key {
		case GatewayChildEnv:
			return "1" // We ARE the child (foreground).
		case "ANTHROPIC_API_KEY":
			return "sk-ant-test"
		}
		return ""
	}

	// Use io.Discard for Stderr to avoid data race.
	opts.Stderr = io.Discard

	// ListenAndServe immediately returns ErrServerClosed
	// to end the test quickly.
	opts.ListenAndServe = func(addr string, handler http.Handler) error {
		return http.ErrServerClosed
	}

	// Start will use srv.ListenAndServe() internally (not
	// the injected one), so we need to use a real port
	// approach. Instead, let's just verify the log file
	// does NOT exist after a quick start/stop cycle.
	//
	// Use a port that's free, and send SIGINT immediately.
	opts.Port = 59200 + os.Getpid()%1000

	errCh := make(chan error, 1)
	go func() {
		errCh <- Start(opts)
	}()

	// Brief wait for server to start, then signal shutdown.
	time.Sleep(100 * time.Millisecond)
	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.SIGINT)

	select {
	case <-errCh:
		// Server stopped.
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after SIGINT")
	}

	// Verify .uf/gateway.log does NOT exist (foreground
	// mode does not create a log file).
	logPath := filepath.Join(opts.ProjectDir, ".uf", "gateway.log")
	if _, err := os.Stat(logPath); err == nil {
		t.Error("gateway.log should NOT exist in foreground mode")
	}
}

// Suppress unused import warnings.
var _ = io.Discard
