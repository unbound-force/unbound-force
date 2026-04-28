package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
)

// DefaultPort is the default local port for the gateway.
// Chosen to avoid conflicts with common development ports
// (3000, 4096, 8080, etc.).
const DefaultPort = 53147

// GatewayChildEnv is the sentinel environment variable set
// when the gateway re-execs itself for background mode.
// The child process detects this to know it should run the
// server instead of forking again.
//
// Design decision: Re-exec with sentinel env var instead
// of syscall.Fork, because Go's runtime is multi-threaded
// and forking is unsafe (per research.md R5).
const GatewayChildEnv = "_UF_GATEWAY_CHILD"

// pidFileName is the PID file name relative to .uf/.
const pidFileName = "gateway.pid"

// HealthResponse is the JSON payload for GET /health.
// Per Constitution Principle III (Observable Quality),
// the health endpoint returns machine-parseable output.
type HealthResponse struct {
	Status        string `json:"status"`
	Provider      string `json:"provider"`
	Port          int    `json:"port"`
	PID           int    `json:"pid"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

// Options configures gateway operations. All external
// dependencies are injected as function fields for
// testability per Constitution Principle IV.
type Options struct {
	// Port is the local port to listen on.
	// Default: 53147.
	Port int

	// ProviderName overrides auto-detection.
	// Valid values: "anthropic", "vertex", "bedrock".
	// Default: "" (auto-detect from env vars).
	ProviderName string

	// Detach starts the gateway in the background.
	Detach bool

	// ProjectDir is the project directory (for PID
	// file location at <ProjectDir>/.uf/gateway.pid).
	ProjectDir string

	// Stdout is the writer for user-facing output.
	Stdout io.Writer

	// Stderr is the writer for progress/status messages.
	Stderr io.Writer

	// --- Injectable dependencies ---

	// LookPath finds a binary in PATH.
	LookPath func(string) (string, error)

	// ExecCmd runs a command and returns combined output.
	ExecCmd func(name string, args ...string) ([]byte, error)

	// ExecStart starts a command without waiting for it
	// to complete. Used for detach (re-exec). Returns
	// the started process and any error.
	ExecStart func(cmd *exec.Cmd) error

	// Getenv reads an environment variable.
	Getenv func(string) string

	// HTTPGet performs an HTTP GET and returns the status
	// code. Used for health check polling.
	HTTPGet func(url string) (int, error)

	// FindProcess looks up a process by PID.
	FindProcess func(int) (*os.Process, error)

	// ListenAndServe starts the HTTP server. Injected
	// for testability — tests can provide a no-op or
	// channel-based implementation.
	ListenAndServe func(addr string, handler http.Handler) error
}

// defaults fills zero-value fields with production
// implementations, following the pattern from
// internal/sandbox/sandbox.go.
func (o *Options) defaults() {
	if o.Port == 0 {
		o.Port = DefaultPort
	}
	if o.ProjectDir == "" {
		o.ProjectDir, _ = os.Getwd()
	}
	if o.Stdout == nil {
		o.Stdout = os.Stdout
	}
	if o.Stderr == nil {
		o.Stderr = os.Stderr
	}
	if o.LookPath == nil {
		o.LookPath = exec.LookPath
	}
	if o.ExecCmd == nil {
		o.ExecCmd = defaultExecCmd
	}
	if o.ExecStart == nil {
		o.ExecStart = defaultExecStart
	}
	if o.Getenv == nil {
		o.Getenv = os.Getenv
	}
	if o.HTTPGet == nil {
		o.HTTPGet = defaultHTTPGet
	}
	if o.FindProcess == nil {
		o.FindProcess = os.FindProcess
	}
	if o.ListenAndServe == nil {
		o.ListenAndServe = http.ListenAndServe
	}
}

// defaultExecCmd is the production implementation of ExecCmd.
func defaultExecCmd(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// defaultExecStart is the production implementation of ExecStart.
func defaultExecStart(cmd *exec.Cmd) error {
	return cmd.Start()
}

// defaultHTTPGet performs an HTTP GET and returns the status code.
func defaultHTTPGet(url string) (int, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url) //nolint:gosec // URL is constructed internally
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode, nil
}

// pidPath returns the full path to the PID file.
func pidPath(projectDir string) string {
	return filepath.Join(projectDir, ".uf", pidFileName)
}

// newMux creates the HTTP handler with health, proxy, and
// fallback routes. The proxy uses httputil.ReverseProxy
// with the provider's PrepareRequest as the Director
// function (per research.md R1).
func newMux(provider Provider, port int, startTime time.Time) http.Handler {
	mux := http.NewServeMux()

	// GET /health — returns gateway status as JSON (FR-006).
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		resp := HealthResponse{
			Status:        "ok",
			Provider:      provider.Name(),
			Port:          port,
			PID:           os.Getpid(),
			UptimeSeconds: int64(time.Since(startTime).Seconds()),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Build the reverse proxy with the provider's Director.
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Strip inbound auth headers — the provider
			// injects its own credentials (FR-013).
			req.Header.Del("Authorization")
			req.Header.Del("x-api-key")

			// Call provider to set upstream URL and inject
			// credentials. For Vertex, PrepareRequest also
			// transforms the body and strips anthropic-*
			// headers (Spec 034 FR-001 through FR-006).
			if err := provider.PrepareRequest(req); err != nil {
				log.Error("provider prepare request failed", "error", err)
				// Mark the request as failed so ErrorHandler
				// can return a proper error response.
				req.Header.Set("X-Gateway-Error", err.Error())
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			// Check if the Director set an error.
			if gwErr := r.Header.Get("X-Gateway-Error"); gwErr != "" {
				writeJSONError(w, http.StatusBadGateway, "auth_error", gwErr)
				return
			}
			writeJSONError(w, http.StatusBadGateway, "upstream_error",
				fmt.Sprintf("upstream connection failed: %v", err))
		},
	}

	// Apply SSE response filtering for Vertex provider.
	// Drops vertex_event and ping SSE events that OpenCode
	// cannot parse (FR-007, FR-008). Only applied when
	// provider is Vertex — Anthropic responses pass through
	// unchanged (FR-010).
	if provider.Name() == "vertex" {
		proxy.ModifyResponse = vertexSSEFilter()
	}

	// Build reusable handlers for both /v1/ and bare paths.
	// Some SDKs set ANTHROPIC_BASE_URL to include /v1
	// (e.g., "http://host:53147/v1"), producing bare
	// paths like POST /messages instead of /v1/messages.
	// Registering both ensures compatibility.

	listModelsHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data":     knownModels,
			"has_more": false,
			"first_id": knownModels[0].ID,
			"last_id":  knownModels[len(knownModels)-1].ID,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}

	lookupModel := func(w http.ResponseWriter, r *http.Request, prefix string) {
		modelID := strings.TrimPrefix(r.URL.Path, prefix)
		if modelID == "" {
			listModelsHandler(w, r)
			return
		}
		for _, m := range knownModels {
			if m.ID == modelID {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(m)
				return
			}
		}
		writeJSONError(w, http.StatusNotFound, "not_found",
			fmt.Sprintf("Model '%s' not found", modelID))
	}

	// --- /v1/ prefixed routes (standard SDK paths) ---

	// POST /v1/messages — proxy to upstream (FR-001).
	mux.HandleFunc("POST /v1/messages", proxy.ServeHTTP)

	// POST /v1/messages/count_tokens — proxy to upstream.
	mux.HandleFunc("POST /v1/messages/count_tokens", proxy.ServeHTTP)

	// GET /v1/models — synthetic model catalog (FR-013).
	mux.HandleFunc("GET /v1/models", listModelsHandler)

	// GET /v1/models/{model_id} — single model lookup.
	mux.HandleFunc("GET /v1/models/", func(w http.ResponseWriter, r *http.Request) {
		lookupModel(w, r, "/v1/models/")
	})

	// --- Bare routes (no /v1 prefix) ---
	// Handles the case where ANTHROPIC_BASE_URL includes
	// the /v1 path component, causing the SDK to send
	// requests like POST /messages instead of /v1/messages.

	mux.HandleFunc("POST /messages", proxy.ServeHTTP)
	mux.HandleFunc("POST /messages/count_tokens", proxy.ServeHTTP)
	mux.HandleFunc("GET /models", listModelsHandler)
	mux.HandleFunc("GET /models/", func(w http.ResponseWriter, r *http.Request) {
		lookupModel(w, r, "/models/")
	})

	// Catch-all: log and return 405 with supported endpoints.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Warn("unsupported endpoint hit",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"remote", r.RemoteAddr,
			"content-type", r.Header.Get("Content-Type"),
			"anthropic-version", r.Header.Get("anthropic-version"))
		writeJSONError(w, http.StatusMethodNotAllowed, "not_found",
			"Unsupported endpoint. Supported: /v1/messages, /v1/messages/count_tokens, /v1/models, /health")
	})

	return mux
}

// modelCapabilities describes the feature support for a
// Claude model on Vertex AI.
type modelCapabilities struct {
	Vision           bool `json:"vision"`
	ExtendedThinking bool `json:"extended_thinking"`
	PDFInput         bool `json:"pdf_input"`
}

// syntheticModel represents a Claude model available on
// Vertex AI. Since Vertex has no model listing API, the
// gateway maintains a hardcoded catalog matching the
// models Google documents as available (FR-013).
//
// Design decision: Hardcoded slice instead of dynamic
// discovery because Vertex has no REST API for listing
// Claude models. Model releases are infrequent (every
// few months), so manual updates are acceptable (per
// research.md R3).
type syntheticModel struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	DisplayName    string            `json:"display_name"`
	CreatedAt      int64             `json:"created_at"`
	MaxInputTokens int               `json:"max_input_tokens"`
	MaxTokens      int               `json:"max_tokens"`
	Capabilities   modelCapabilities `json:"capabilities"`
}

// knownModels is the catalog of Claude models available on
// Vertex AI as of April 2026 (per research.md R3). All
// models support vision and PDF input. All except Haiku 4.5
// support extended thinking.
var knownModels = []syntheticModel{
	{
		ID: "claude-opus-4-7-20250416", Type: "model",
		DisplayName: "Claude Opus 4.7", CreatedAt: 1744761600,
		MaxInputTokens: 200000, MaxTokens: 32000,
		Capabilities: modelCapabilities{Vision: true, ExtendedThinking: true, PDFInput: true},
	},
	{
		ID: "claude-sonnet-4-6-20250217", Type: "model",
		DisplayName: "Claude Sonnet 4.6", CreatedAt: 1739750400,
		MaxInputTokens: 200000, MaxTokens: 64000,
		Capabilities: modelCapabilities{Vision: true, ExtendedThinking: true, PDFInput: true},
	},
	{
		ID: "claude-opus-4-6-20250205", Type: "model",
		DisplayName: "Claude Opus 4.6", CreatedAt: 1738713600,
		MaxInputTokens: 200000, MaxTokens: 32000,
		Capabilities: modelCapabilities{Vision: true, ExtendedThinking: true, PDFInput: true},
	},
	{
		ID: "claude-opus-4-5-20241124", Type: "model",
		DisplayName: "Claude Opus 4.5", CreatedAt: 1732406400,
		MaxInputTokens: 200000, MaxTokens: 32000,
		Capabilities: modelCapabilities{Vision: true, ExtendedThinking: true, PDFInput: true},
	},
	{
		ID: "claude-sonnet-4-5-20241022", Type: "model",
		DisplayName: "Claude Sonnet 4.5", CreatedAt: 1729555200,
		MaxInputTokens: 200000, MaxTokens: 8000,
		Capabilities: modelCapabilities{Vision: true, ExtendedThinking: true, PDFInput: true},
	},
	{
		ID: "claude-opus-4-1-20250414", Type: "model",
		DisplayName: "Claude Opus 4.1", CreatedAt: 1744588800,
		MaxInputTokens: 200000, MaxTokens: 32000,
		Capabilities: modelCapabilities{Vision: true, ExtendedThinking: true, PDFInput: true},
	},
	{
		ID: "claude-haiku-4-5-20241022", Type: "model",
		DisplayName: "Claude Haiku 4.5", CreatedAt: 1729555200,
		MaxInputTokens: 200000, MaxTokens: 8000,
		Capabilities: modelCapabilities{Vision: true, ExtendedThinking: false, PDFInput: true},
	},
	{
		ID: "claude-opus-4-20250514", Type: "model",
		DisplayName: "Claude Opus 4", CreatedAt: 1747180800,
		MaxInputTokens: 200000, MaxTokens: 32000,
		Capabilities: modelCapabilities{Vision: true, ExtendedThinking: true, PDFInput: true},
	},
	{
		ID: "claude-sonnet-4-20250514", Type: "model",
		DisplayName: "Claude Sonnet 4", CreatedAt: 1747180800,
		MaxInputTokens: 200000, MaxTokens: 8000,
		Capabilities: modelCapabilities{Vision: true, ExtendedThinking: true, PDFInput: true},
	},
}

// writeJSONError writes a JSON error response matching the
// Anthropic error format (per contracts/gateway-api.md).
func writeJSONError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"type":    errType,
			"message": message,
		},
	})
}

// Start starts the gateway. It detects (or uses the
// overridden) provider, initializes credentials, writes
// the PID file, and starts the HTTP server. Handles
// graceful shutdown on SIGINT/SIGTERM.
func Start(opts Options) error {
	opts.defaults()

	// If --detach is requested and we are NOT the child
	// process, re-exec in the background.
	if opts.Detach && opts.Getenv(GatewayChildEnv) != "1" {
		return detach(opts)
	}

	// Resolve the provider (auto-detect or explicit).
	var prov Provider
	var err error
	if opts.ProviderName != "" {
		prov, err = NewProviderByName(opts.ProviderName, opts.Getenv, opts.ExecCmd)
	} else {
		prov, err = DetectProvider(opts.Getenv, opts.ExecCmd)
	}
	if err != nil {
		return err
	}

	// Clean up stale PID file from a previous crash.
	pp := pidPath(opts.ProjectDir)
	if err := CleanupStale(pp, opts.FindProcess); err != nil {
		log.Warn("failed to clean stale PID file", "error", err)
	}

	// Check if a gateway is already running.
	if info, readErr := ReadPID(pp); readErr == nil {
		if IsAlive(info.PID, opts.FindProcess) {
			return fmt.Errorf(
				"gateway already running (PID %d) on port %d. "+
					"Use `uf gateway stop` first or `--port` to use a different port",
				info.PID, info.Port)
		}
	}

	// Initialize the provider (acquire initial credentials).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := prov.Start(ctx); err != nil {
		return fmt.Errorf("provider %s initialization failed: %w", prov.Name(), err)
	}

	// Build the HTTP handler.
	startTime := time.Now()
	handler := newMux(prov, opts.Port, startTime)

	// Write PID file.
	pidInfo := PIDInfo{
		PID:      os.Getpid(),
		Port:     opts.Port,
		Provider: prov.Name(),
		Started:  startTime,
	}
	if err := WritePID(pp, pidInfo); err != nil {
		return fmt.Errorf("write PID file: %w", err)
	}

	// Set up graceful shutdown on SIGINT/SIGTERM.
	sigCtx, sigStop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer sigStop()

	// Create the server for graceful shutdown support.
	addr := fmt.Sprintf(":%d", opts.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Start the server in a goroutine.
	serverErr := make(chan error, 1)
	go func() {
		log.Info("gateway started",
			"provider", prov.Name(),
			"port", opts.Port,
			"pid", os.Getpid())
		fmt.Fprintf(opts.Stderr, "Gateway listening on port %d (provider: %s)\n",
			opts.Port, prov.Name())
		serverErr <- srv.ListenAndServe()
	}()

	// Wait for shutdown signal or server error.
	select {
	case <-sigCtx.Done():
		log.Info("shutting down gateway")
		fmt.Fprintf(opts.Stderr, "Shutting down gateway...\n")
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			// Check for port conflict.
			if isAddrInUse(err) {
				_ = RemovePID(pp)
				return fmt.Errorf(
					"port %d is already in use. Use `--port` to specify a different port",
					opts.Port)
			}
			_ = RemovePID(pp)
			return fmt.Errorf("server error: %w", err)
		}
	}

	// Graceful shutdown: drain in-flight requests.
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)

	// Stop the provider (cancel refresh goroutines).
	prov.Stop()

	// Remove PID file.
	_ = RemovePID(pp)

	fmt.Fprintf(opts.Stderr, "Gateway stopped.\n")
	return nil
}

// detach re-execs the binary as a background process with
// the GatewayChildEnv sentinel. The parent waits for the
// health endpoint to respond, then exits.
//
// Design decision: Re-exec instead of syscall.Fork because
// Go's runtime is multi-threaded (per research.md R5).
func detach(opts Options) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}

	// Build child args: same as parent but without --detach.
	args := []string{"gateway"}
	if opts.Port != DefaultPort {
		args = append(args, "--port", fmt.Sprintf("%d", opts.Port))
	}
	if opts.ProviderName != "" {
		args = append(args, "--provider", opts.ProviderName)
	}

	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), GatewayChildEnv+"=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	// Ensure .uf/ directory exists for the log file.
	// It should already exist from PID file creation,
	// but guard defensively.
	ufDir := filepath.Join(opts.ProjectDir, ".uf")
	if err := os.MkdirAll(ufDir, 0755); err != nil {
		return fmt.Errorf("create .uf directory: %w", err)
	}

	// Redirect child stdout/stderr to .uf/gateway.log
	// so auth diagnostics are captured for debugging.
	// Owner-only permissions (0600) since the log may
	// contain credential refresh output.
	logPath := filepath.Join(ufDir, "gateway.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open gateway log file: %w", err)
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := opts.ExecStart(cmd); err != nil {
		_ = logFile.Close()
		return fmt.Errorf("start background gateway: %w", err)
	}
	// Close the parent's handle — the child inherited
	// the file descriptor independently.
	_ = logFile.Close()

	childPID := cmd.Process.Pid

	// Wait for the health endpoint with exponential backoff.
	healthURL := fmt.Sprintf("http://localhost:%d/health", opts.Port)
	deadline := time.Now().Add(10 * time.Second)
	interval := 200 * time.Millisecond
	maxInterval := 2 * time.Second

	for time.Now().Before(deadline) {
		code, err := opts.HTTPGet(healthURL)
		if err == nil && code == http.StatusOK {
			fmt.Fprintf(opts.Stdout,
				"Gateway started (PID %d) on port %d. Logs: .uf/gateway.log\n",
				childPID, opts.Port)
			return nil
		}
		time.Sleep(interval)
		if interval < maxInterval {
			interval *= 2
			if interval > maxInterval {
				interval = maxInterval
			}
		}
	}

	// Health check timed out — the child may have failed.
	return fmt.Errorf(
		"gateway started (PID %d) but health check timed out on port %d",
		childPID, opts.Port)
}

// Stop terminates a running background gateway and removes
// its PID file. Prints "No gateway running." if no gateway
// is found (FR-008).
func Stop(opts Options) error {
	opts.defaults()

	pp := pidPath(opts.ProjectDir)
	info, err := ReadPID(pp)
	if err != nil {
		fmt.Fprintf(opts.Stdout, "No gateway running.\n")
		return nil
	}

	if !IsAlive(info.PID, opts.FindProcess) {
		// Stale PID file — clean up.
		_ = RemovePID(pp)
		fmt.Fprintf(opts.Stdout, "No gateway running.\n")
		return nil
	}

	// Send SIGTERM to the gateway process.
	proc, err := opts.FindProcess(info.PID)
	if err != nil {
		_ = RemovePID(pp)
		fmt.Fprintf(opts.Stdout, "No gateway running.\n")
		return nil
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		log.Warn("failed to send SIGTERM", "pid", info.PID, "error", err)
	}

	// Wait briefly for the process to exit.
	time.Sleep(500 * time.Millisecond)

	// Remove PID file.
	_ = RemovePID(pp)

	fmt.Fprintf(opts.Stdout, "Gateway stopped.\n")
	return nil
}

// Status displays the running gateway's provider, port,
// PID, and uptime. Prints "No gateway running." if no
// gateway is found (FR-008).
func Status(opts Options) error {
	opts.defaults()

	pp := pidPath(opts.ProjectDir)
	info, err := ReadPID(pp)
	if err != nil {
		fmt.Fprintf(opts.Stdout, "No gateway running.\n")
		return nil
	}

	if !IsAlive(info.PID, opts.FindProcess) {
		// Stale PID file — clean up.
		_ = RemovePID(pp)
		fmt.Fprintf(opts.Stdout, "No gateway running.\n")
		return nil
	}

	// Query the health endpoint for live data (informational
	// only — status is displayed regardless of health).
	healthURL := fmt.Sprintf("http://localhost:%d/health", info.Port)
	opts.HTTPGet(healthURL) //nolint:errcheck // best-effort

	printGatewayStatus(opts.Stdout, info, opts.ProjectDir)
	return nil
}

// printGatewayStatus writes the formatted gateway status
// to w. Extracted from Status() to avoid duplicating the
// display logic for the health-success and health-failure
// paths.
func printGatewayStatus(w io.Writer, info *PIDInfo, projectDir string) {
	uptime := time.Since(info.Started)
	fmt.Fprintf(w, "Gateway Status\n")
	fmt.Fprintf(w, "  Provider:  %s\n", info.Provider)
	fmt.Fprintf(w, "  Port:      %d\n", info.Port)
	fmt.Fprintf(w, "  PID:       %d\n", info.PID)
	fmt.Fprintf(w, "  Uptime:    %s\n", formatUptime(uptime))
	logPath := filepath.Join(projectDir, ".uf", "gateway.log")
	if _, statErr := os.Stat(logPath); statErr == nil {
		fmt.Fprintf(w, "  Log:       .uf/gateway.log\n")
	}
}

// formatUptime formats a duration as a human-readable
// string like "1h 23m" or "45s".
func formatUptime(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// isAddrInUse checks if an error indicates the address is
// already in use (EADDRINUSE).
func isAddrInUse(err error) bool {
	if err == nil {
		return false
	}
	// Check for net.OpError wrapping syscall.EADDRINUSE.
	if opErr, ok := err.(*net.OpError); ok {
		if sysErr, ok := opErr.Err.(*os.SyscallError); ok {
			return strings.Contains(sysErr.Err.Error(), "address already in use")
		}
	}
	return strings.Contains(err.Error(), "address already in use")
}
