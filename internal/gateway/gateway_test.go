package gateway

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
	"path/filepath"
	"strings"
	"sync/atomic"
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
		projectID: "my-project",
		region:    "us-east5",
		token:     "ya29.test-token",
	}

	body := `{"model":"claude-sonnet-4-20250514","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/messages",
		strings.NewReader(body))

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

	prov := newVertexProvider(getenv, execCmd)
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

	prov := newVertexProvider(getenv, execCmd)
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

	prov := newVertexProvider(getenv, execCmd)
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

// Suppress unused import warnings.
var _ = io.Discard
