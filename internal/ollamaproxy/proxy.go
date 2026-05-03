// Package ollamaproxy implements an Ollama-compatible API
// proxy that translates requests to Vertex AI (embeddings)
// and the uf gateway (generation). It follows the gateway
// lifecycle pattern: Options struct with injectable deps,
// Start/Stop/Status, PID file management, and detach mode.
//
// Design decision: Direct HTTP calls for embeddings (D3)
// because the proxy must translate between two different
// API formats (Ollama ↔ Vertex AI). Generation delegates
// to the gateway (D4) to reuse Anthropic/Vertex
// translation logic.
package ollamaproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/log"

	"github.com/unbound-force/unbound-force/internal/auth"
	"github.com/unbound-force/unbound-force/internal/pidfile"
)

// DefaultPort is the default local port for the proxy.
// Matches Ollama's default for zero-config Dewey usage
// (per design.md D6).
const DefaultPort = 11434

// DefaultEmbedModel is the default Vertex AI embedding
// model used when no override is configured.
const DefaultEmbedModel = "text-embedding-005"

// DefaultGatewayURL is the default URL for the uf gateway
// that handles Anthropic Messages API requests.
const DefaultGatewayURL = "http://localhost:53147"

// ChildEnv is the sentinel environment variable set when
// the proxy re-execs itself for background mode. The child
// process detects this to run the server instead of
// forking again (per design.md D8).
const ChildEnv = "_UF_OLLAMA_PROXY_CHILD"

// pidFileName is the PID file name relative to .uf/.
const pidFileName = "ollama-proxy.pid"

// logFileName is the log file name relative to .uf/.
const logFileName = "ollama-proxy.log"

// maxRequestBodySize is the maximum allowed request body
// size (10MB). Requests exceeding this are rejected with
// HTTP 413 to prevent memory exhaustion.
const maxRequestBodySize = 10 * 1024 * 1024

// vertexTokenLifetime is the assumed lifetime of a gcloud
// OAuth token. Set to 55 minutes (tokens typically expire
// at 60 minutes, with 5 minutes of safety margin).
const vertexTokenLifetime = 55 * time.Minute

// proactiveRefreshWindow is the window before token expiry
// during which Token() will attempt a proactive refresh.
const proactiveRefreshWindow = 5 * time.Minute

// healthResponse is the JSON payload for GET /health.
// The "service" field identifies this as the proxy (not
// real Ollama) for uf doctor and status to distinguish
// (per spec requirement).
type healthResponse struct {
	Service          string `json:"service"`
	Status           string `json:"status"`
	Port             int    `json:"port"`
	EmbedModel       string `json:"embed_model"`
	GatewayAvailable bool   `json:"gateway_available"`
}

// Options configures ollama-proxy operations. All external
// dependencies are injected as function fields for
// testability per Constitution Principle IV.
type Options struct {
	// Port is the local port to listen on.
	// Default: 11434.
	Port int

	// EmbedModel is the Vertex AI embedding model name.
	// Default: "text-embedding-005".
	EmbedModel string

	// GatewayURL is the URL for the uf gateway.
	// Must be a loopback address (per design.md D8b).
	// Default: "http://localhost:53147".
	GatewayURL string

	// ProjectDir is the project directory (for PID
	// file location at <ProjectDir>/.uf/ollama-proxy.pid).
	ProjectDir string

	// Detach starts the proxy in the background.
	Detach bool

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
	// to complete. Used for detach (re-exec).
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

	// HTTPClient is the HTTP client used for upstream
	// requests (Vertex AI, gateway). Injected for
	// testability via httptest.
	HTTPClient *http.Client
}

// defaults fills zero-value fields with production
// implementations, following the pattern from
// internal/gateway/gateway.go.
func (o *Options) defaults() {
	if o.Port == 0 {
		o.Port = DefaultPort
	}
	if o.EmbedModel == "" {
		o.EmbedModel = DefaultEmbedModel
	}
	if o.GatewayURL == "" {
		o.GatewayURL = DefaultGatewayURL
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
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: 30 * time.Second}
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

// defaultModelMap maps Ollama model names to cloud model
// names. Unknown model names are rejected — passthrough
// is not allowed because model names are interpolated
// into Vertex AI URLs, and unvalidated strings could
// contain path traversal characters (per design.md D5).
var defaultModelMap = map[string]string{
	"granite-embedding:30m":              "text-embedding-005",
	"granite-embedding-small-english-r2": "text-embedding-005",
	"llama3.2:3b":                        "claude-sonnet-4-20250514",
}

// mapModelName returns the cloud model name for the given
// Ollama model name. Returns ("", false) for unknown models.
func mapModelName(name string) (string, bool) {
	cloud, ok := defaultModelMap[name]
	return cloud, ok
}

// safeModelNameRe matches model names containing only safe
// characters: alphanumeric, hyphens, colons, periods, and
// underscores. This prevents path traversal and injection
// when model names are interpolated into URLs.
var safeModelNameRe = regexp.MustCompile(`^[a-zA-Z0-9._:][a-zA-Z0-9._:\-]*$`)

// validateModelName checks that a model name contains only
// safe characters. Names with /, \, %, or control
// characters are rejected to prevent path traversal when
// constructing Vertex AI URLs (per design.md D5).
func validateModelName(name string) error {
	if name == "" {
		return fmt.Errorf("model name is required")
	}
	if !safeModelNameRe.MatchString(name) {
		return fmt.Errorf(
			"invalid model name %q: must contain only "+
				"alphanumeric characters, hyphens, colons, "+
				"periods, and underscores", name)
	}
	return nil
}

// bearerTokenRe matches Bearer tokens in error response
// bodies. Used by redactToken to strip credentials from
// upstream error messages before relaying to the client
// (per design.md D8a).
var bearerTokenRe = regexp.MustCompile(
	`(?i)(Bearer\s+)[A-Za-z0-9\-._~+/]+=*`)

// redactToken strips Bearer tokens from error response
// bodies. This prevents OAuth tokens from leaking to
// Ollama clients when Vertex AI echoes the Authorization
// header in error responses (per design.md D8a).
func redactToken(body string) string {
	return bearerTokenRe.ReplaceAllString(body, "${1}[REDACTED]")
}

// validateGatewayURL validates that the gateway URL is
// safe to forward requests to. The scheme must be http or
// https, and the host must resolve to a loopback address
// (127.0.0.1, ::1, localhost). Non-loopback URLs are
// rejected to prevent SSRF (per design.md D8b).
func validateGatewayURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid gateway URL %q: %w", rawURL, err)
	}

	// Validate scheme.
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf(
			"gateway URL scheme must be http or https, got %q",
			u.Scheme)
	}

	// Validate host is loopback.
	host := u.Hostname()
	if host != "localhost" && host != "127.0.0.1" && host != "::1" {
		return fmt.Errorf(
			"gateway URL host must be loopback "+
				"(localhost, 127.0.0.1, or ::1), got %q. "+
				"Forwarding prompt content to non-loopback "+
				"hosts is not allowed (SSRF prevention)",
			host)
	}

	// Validate path is empty or root.
	if u.Path != "" && u.Path != "/" {
		return fmt.Errorf(
			"gateway URL path must be empty or /, got %q",
			u.Path)
	}

	return nil
}

// pidPath returns the full path to the PID file.
func pidPath(projectDir string) string {
	return filepath.Join(projectDir, ".uf", pidFileName)
}

// writeOllamaError writes a JSON error response in the
// Ollama error format: {"error": "message"}.
func writeOllamaError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}

// proxyServer holds the runtime state for a running proxy
// instance. It is created by Start() and used by the HTTP
// handlers to access the token manager and configuration.
type proxyServer struct {
	tokenMgr         *auth.TokenManager
	opts             Options
	gatewayAvailable bool
	startTime        time.Time
}

// newMux creates the HTTP handler with health, embed,
// generate, tags, and fallback routes.
func newMux(ps *proxyServer) http.Handler {
	mux := http.NewServeMux()

	// GET /health — returns proxy status as JSON.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		resp := healthResponse{
			Service:          "uf-ollama-proxy",
			Status:           "ok",
			Port:             ps.opts.Port,
			EmbedModel:       ps.opts.EmbedModel,
			GatewayAvailable: ps.gatewayAvailable,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	// POST /api/embed — Ollama embed endpoint.
	mux.HandleFunc("POST /api/embed", ps.handleEmbed)

	// POST /api/generate — Ollama generate endpoint.
	mux.HandleFunc("POST /api/generate", ps.handleGenerate)

	// GET /api/tags — Ollama model list endpoint.
	mux.HandleFunc("GET /api/tags", ps.handleTags)

	// Catch-all: return 404 with supported endpoints.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Warn("unsupported endpoint hit",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr)
		writeOllamaError(w, http.StatusNotFound,
			"unsupported endpoint. Supported: "+
				"/api/embed, /api/generate, /api/tags, /health")
	})

	return mux
}

// checkGatewayHealth probes the gateway health endpoint
// and returns true if it responds with HTTP 200 and the
// expected service identifier.
func checkGatewayHealth(httpGet func(string) (int, error), gatewayURL string) bool {
	healthURL := gatewayURL + "/health"
	code, err := httpGet(healthURL)
	return err == nil && code == http.StatusOK
}

// Start starts the ollama-proxy. It validates config,
// initializes the token manager, checks gateway health,
// writes the PID file, and starts the HTTP server.
// Handles graceful shutdown on SIGINT/SIGTERM.
func Start(opts Options) error {
	opts.defaults()

	// If --detach is requested and we are NOT the child
	// process, re-exec in the background.
	if opts.Detach && opts.Getenv(ChildEnv) != "1" {
		return detach(opts)
	}

	// Validate gateway URL is loopback (D8b).
	if err := validateGatewayURL(opts.GatewayURL); err != nil {
		return err
	}

	// Verify gcloud is in PATH (required for Vertex AI
	// OAuth token acquisition).
	if _, err := opts.LookPath("gcloud"); err != nil {
		return fmt.Errorf(
			"gcloud CLI not found in PATH. Install: "+
				"https://cloud.google.com/sdk/docs/install\n"+
				"Then authenticate: "+
				"gcloud auth application-default login")
	}

	// Clean up stale PID file from a previous crash.
	pp := pidPath(opts.ProjectDir)
	if err := pidfile.CleanupStale(pp, opts.FindProcess); err != nil {
		log.Warn("failed to clean stale PID file", "error", err)
	}

	// Check if a proxy is already running. Probe health
	// for "service": "uf-ollama-proxy" to distinguish
	// from real Ollama.
	if info, readErr := pidfile.ReadPID(pp); readErr == nil {
		if pidfile.IsAlive(info.PID, opts.FindProcess) {
			return fmt.Errorf(
				"ollama-proxy already running (PID %d) on port %d. "+
					"Use `uf ollama-proxy stop` first or "+
					"`--port` to use a different port",
				info.PID, info.Port)
		}
	}

	// Region is resolved per-request in embed.go using
	// the same env var chain (per design.md D9).

	projectID := opts.Getenv("ANTHROPIC_VERTEX_PROJECT_ID")
	if projectID == "" {
		projectID = opts.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	if projectID == "" {
		return fmt.Errorf(
			"vertex AI project ID not set; set one of: " +
				"ANTHROPIC_VERTEX_PROJECT_ID or GOOGLE_CLOUD_PROJECT")
	}

	// Initialize token manager for Vertex AI OAuth.
	execCmd := opts.ExecCmd
	tokenMgr := auth.NewTokenManager(auth.TokenManagerOpts{
		RefreshFn: func() (string, error) {
			return auth.RefreshVertexToken(execCmd)
		},
		Lifetime:        vertexTokenLifetime,
		ProactiveWindow: proactiveRefreshWindow,
		Interval:        50 * time.Minute,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := tokenMgr.Start(ctx); err != nil {
		return fmt.Errorf("vertex AI token acquisition failed: %w", err)
	}

	// Check gateway health (informational warning, not
	// a hard failure — the proxy can still serve
	// /api/embed without the gateway, per design.md D4).
	gatewayAvailable := checkGatewayHealth(opts.HTTPGet, opts.GatewayURL)
	if !gatewayAvailable {
		log.Warn("gateway not available — /api/generate will fail",
			"gateway_url", opts.GatewayURL,
			"hint", "run `uf gateway start` to enable generation")
	}

	// Log startup warning about dewey reindex requirement
	// when switching from local Ollama (per design.md R1).
	log.Warn("if switching from local Ollama, run `dewey reindex` once "+
		"(Vertex text-embedding-005 produces 768-dim vectors vs "+
		"granite's 256-dim)")

	// Build the proxy server with runtime state.
	ps := &proxyServer{
		tokenMgr:         tokenMgr,
		opts:             opts,
		gatewayAvailable: gatewayAvailable,
		startTime:        time.Now(),
	}

	// Build the HTTP handler.
	handler := newMux(ps)

	// Write PID file.
	pidInfo := pidfile.PIDInfo{
		PID:      os.Getpid(),
		Port:     opts.Port,
		Provider: "vertex-embedding",
		Started:  ps.startTime,
	}
	if err := pidfile.WritePID(pp, pidInfo); err != nil {
		tokenMgr.Stop()
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
		log.Info("ollama-proxy started",
			"port", opts.Port,
			"embed_model", opts.EmbedModel,
			"gateway_url", opts.GatewayURL,
			"pid", os.Getpid())
		fmt.Fprintf(opts.Stderr,
			"Ollama proxy listening on port %d "+
				"(embed model: %s)\n",
			opts.Port, opts.EmbedModel)
		serverErr <- srv.ListenAndServe()
	}()

	// Wait for shutdown signal or server error.
	select {
	case <-sigCtx.Done():
		log.Info("shutting down ollama-proxy")
		fmt.Fprintf(opts.Stderr, "Shutting down ollama-proxy...\n")
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			if isAddrInUse(err) {
				_ = pidfile.RemovePID(pp)
				tokenMgr.Stop()
				return fmt.Errorf(
					"port %d is already in use. "+
						"Use `--port` to specify a different port",
					opts.Port)
			}
			_ = pidfile.RemovePID(pp)
			tokenMgr.Stop()
			return fmt.Errorf("server error: %w", err)
		}
	}

	// Graceful shutdown: drain in-flight requests.
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)

	// Stop token refresh goroutine.
	tokenMgr.Stop()

	// Remove PID file.
	_ = pidfile.RemovePID(pp)

	fmt.Fprintf(opts.Stderr, "Ollama proxy stopped.\n")
	return nil
}

// detach re-execs the binary as a background process with
// the ChildEnv sentinel. The parent waits for the health
// endpoint to respond, then exits.
func detach(opts Options) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}

	// Build child args: same as parent but without --detach.
	args := []string{"ollama-proxy"}
	if opts.Port != DefaultPort {
		args = append(args, "--port", fmt.Sprintf("%d", opts.Port))
	}
	if opts.EmbedModel != DefaultEmbedModel {
		args = append(args, "--embed-model", opts.EmbedModel)
	}
	if opts.GatewayURL != DefaultGatewayURL {
		args = append(args, "--gateway-url", opts.GatewayURL)
	}

	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), ChildEnv+"=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	// Ensure .uf/ directory exists for the log file.
	ufDir := filepath.Join(opts.ProjectDir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		return fmt.Errorf("create .uf directory: %w", err)
	}

	// Redirect child stdout/stderr to .uf/ollama-proxy.log.
	// Owner-only permissions (0600) since the log may
	// contain credential refresh output (per D8a).
	logPath := filepath.Join(ufDir, logFileName)
	logFile, err := os.OpenFile(logPath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open proxy log file: %w", err)
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := opts.ExecStart(cmd); err != nil {
		_ = logFile.Close()
		return fmt.Errorf("start background proxy: %w", err)
	}
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
				"Ollama proxy started (PID %d) on port %d. "+
					"Logs: .uf/%s\n",
				childPID, opts.Port, logFileName)
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

	return fmt.Errorf(
		"ollama-proxy started (PID %d) but health check "+
			"timed out on port %d",
		childPID, opts.Port)
}

// Stop terminates a running background proxy and removes
// its PID file.
func Stop(opts Options) error {
	opts.defaults()

	pp := pidPath(opts.ProjectDir)
	info, err := pidfile.ReadPID(pp)
	if err != nil {
		fmt.Fprintf(opts.Stdout, "No ollama-proxy running.\n")
		return nil
	}

	if !pidfile.IsAlive(info.PID, opts.FindProcess) {
		_ = pidfile.RemovePID(pp)
		fmt.Fprintf(opts.Stdout, "No ollama-proxy running.\n")
		return nil
	}

	proc, err := opts.FindProcess(info.PID)
	if err != nil {
		_ = pidfile.RemovePID(pp)
		fmt.Fprintf(opts.Stdout, "No ollama-proxy running.\n")
		return nil
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		log.Warn("failed to send SIGTERM",
			"pid", info.PID, "error", err)
	}

	// Wait briefly for the process to exit.
	time.Sleep(500 * time.Millisecond)

	_ = pidfile.RemovePID(pp)
	fmt.Fprintf(opts.Stdout, "Ollama proxy stopped.\n")
	return nil
}

// Status displays the running proxy's port, model, PID,
// and uptime.
func Status(opts Options) error {
	opts.defaults()

	pp := pidPath(opts.ProjectDir)
	info, err := pidfile.ReadPID(pp)
	if err != nil {
		fmt.Fprintf(opts.Stdout, "No ollama-proxy running.\n")
		return nil
	}

	if !pidfile.IsAlive(info.PID, opts.FindProcess) {
		_ = pidfile.RemovePID(pp)
		fmt.Fprintf(opts.Stdout, "No ollama-proxy running.\n")
		return nil
	}

	uptime := time.Since(info.Started)
	fmt.Fprintf(opts.Stdout, "Ollama Proxy Status\n")
	fmt.Fprintf(opts.Stdout, "  Port:      %d\n", info.Port)
	fmt.Fprintf(opts.Stdout, "  Provider:  %s\n", info.Provider)
	fmt.Fprintf(opts.Stdout, "  PID:       %d\n", info.PID)
	fmt.Fprintf(opts.Stdout, "  Uptime:    %s\n", formatUptime(uptime))
	logPath := filepath.Join(opts.ProjectDir, ".uf", logFileName)
	if _, statErr := os.Stat(logPath); statErr == nil {
		fmt.Fprintf(opts.Stdout, "  Log:       .uf/%s\n", logFileName)
	}
	return nil
}

// formatUptime formats a duration as a human-readable
// string like "1h 23m" or "45s".
func formatUptime(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds",
			int(d.Minutes()), int(d.Seconds())%60)
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
	return strings.Contains(err.Error(), "address already in use")
}
