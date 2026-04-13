package sandbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// CheBackend implements Backend for Eclipse Che / Dev
// Spaces workspace provisioning.
type CheBackend struct {
	// cheURL is the Che server URL.
	cheURL string

	// useChectl is true when chectl is available.
	useChectl bool
}

// Name returns the backend identifier.
func (b *CheBackend) Name() string { return BackendChe }

// Create provisions a CDE workspace via Eclipse Che /
// Dev Spaces from the project's devfile.
//
// Steps:
//  1. Verify CDE access (chectl or REST API)
//  2. Locate devfile.yaml in project directory
//  3. Create workspace via chectl or REST API
//  4. Wait for workspace to reach RUNNING state
//
// Returns error if devfile is missing or CDE is
// unreachable.
func (b *CheBackend) Create(opts Options) error {
	opts.defaults()

	wsName := cheWorkspaceName(opts)

	// Check for devfile.yaml.
	devfilePath := filepath.Join(opts.ProjectDir, "devfile.yaml")
	if _, err := os.Stat(devfilePath); err != nil {
		return fmt.Errorf(
			"devfile.yaml not found in %s, add a devfile or use --backend podman",
			opts.ProjectDir)
	}

	if b.useChectl {
		return b.createWithChectl(opts, wsName, devfilePath)
	}
	return b.createWithAPI(opts, wsName, devfilePath)
}

// createWithChectl creates a workspace via the chectl CLI.
func (b *CheBackend) createWithChectl(opts Options, wsName, devfilePath string) error {
	out, err := opts.ExecCmd("chectl", "workspace:create",
		"--devfile="+devfilePath,
		"--name="+wsName)
	if err != nil {
		outStr := strings.TrimSpace(string(out))
		if strings.Contains(outStr, "already exists") {
			proj := projectName(opts.ProjectDir)
			return fmt.Errorf(
				"CDE workspace already exists for %s, use `uf sandbox start` or `uf sandbox destroy`",
				proj)
		}
		return fmt.Errorf("failed to create CDE workspace: %s", outStr)
	}

	// Start the workspace.
	if out, err := opts.ExecCmd("chectl", "workspace:start",
		"--name="+wsName); err != nil {
		return fmt.Errorf("failed to start CDE workspace: %s",
			strings.TrimSpace(string(out)))
	}

	fmt.Fprintf(opts.Stderr, "CDE workspace created: %s\n", wsName)
	return nil
}

// createWithAPI creates a workspace via the Che REST API.
func (b *CheBackend) createWithAPI(opts Options, wsName, devfilePath string) error {
	devfileData, err := opts.ReadFile(devfilePath)
	if err != nil {
		return fmt.Errorf("read devfile: %w", err)
	}

	url := strings.TrimRight(b.cheURL, "/") + "/api/workspace/devfile"
	req, err := http.NewRequest("POST", url, bytes.NewReader(devfileData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-yaml")

	// Add auth token if available.
	cfg, _ := LoadConfig(opts)
	token := cfg.Che.Token
	if envToken := opts.Getenv("UF_CHE_TOKEN"); envToken != "" {
		token = envToken
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := opts.HTTPDo(req)
	if err != nil {
		return fmt.Errorf(
			"cannot reach Che at %s, check UF_CHE_URL and authentication",
			b.cheURL)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("che API error (%d): %s",
			resp.StatusCode, strings.TrimSpace(string(body)))
	}

	fmt.Fprintf(opts.Stderr, "CDE workspace created: %s\n", wsName)
	return nil
}

// Start starts a stopped CDE workspace.
func (b *CheBackend) Start(opts Options) error {
	opts.defaults()

	wsName := cheWorkspaceName(opts)

	if b.useChectl {
		if out, err := opts.ExecCmd("chectl", "workspace:start",
			"--name="+wsName); err != nil {
			return fmt.Errorf("failed to start CDE workspace: %s",
				strings.TrimSpace(string(out)))
		}
	} else {
		url := strings.TrimRight(b.cheURL, "/") +
			"/api/workspace/" + wsName + "/runtime"
		body := []byte(`{"status":"RUNNING"}`)
		req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		addCheAuth(req, opts)

		resp, err := opts.HTTPDo(req)
		if err != nil {
			return fmt.Errorf("cannot reach Che at %s", b.cheURL)
		}
		defer func() { _ = resp.Body.Close() }()
	}

	if !opts.Detach {
		return b.Attach(opts)
	}

	fmt.Fprintf(opts.Stdout, "CDE workspace started: %s\n", wsName)
	return nil
}

// Stop stops a running CDE workspace. Workspace state
// is preserved by the CDE platform.
func (b *CheBackend) Stop(opts Options) error {
	opts.defaults()

	wsName := cheWorkspaceName(opts)

	if b.useChectl {
		if out, err := opts.ExecCmd("chectl", "workspace:stop",
			"--name="+wsName); err != nil {
			return fmt.Errorf("failed to stop CDE workspace: %s",
				strings.TrimSpace(string(out)))
		}
	} else {
		url := strings.TrimRight(b.cheURL, "/") +
			"/api/workspace/" + wsName + "/runtime"
		body := []byte(`{"status":"STOPPED"}`)
		req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		addCheAuth(req, opts)

		resp, err := opts.HTTPDo(req)
		if err != nil {
			return fmt.Errorf("cannot reach Che at %s", b.cheURL)
		}
		defer func() { _ = resp.Body.Close() }()
	}

	fmt.Fprintf(opts.Stdout, "CDE workspace stopped.\n")
	return nil
}

// Destroy permanently deletes the CDE workspace.
func (b *CheBackend) Destroy(opts Options) error {
	opts.defaults()

	wsName := cheWorkspaceName(opts)

	if b.useChectl {
		if out, err := opts.ExecCmd("chectl", "workspace:delete",
			"--name="+wsName, "--yes"); err != nil {
			return fmt.Errorf("failed to delete CDE workspace: %s",
				strings.TrimSpace(string(out)))
		}
	} else {
		url := strings.TrimRight(b.cheURL, "/") +
			"/api/workspace/" + wsName
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		addCheAuth(req, opts)

		resp, err := opts.HTTPDo(req)
		if err != nil {
			return fmt.Errorf("cannot reach Che at %s", b.cheURL)
		}
		defer func() { _ = resp.Body.Close() }()
	}

	fmt.Fprintf(opts.Stdout, "CDE workspace destroyed.\n")
	return nil
}

// Status returns the current state of the CDE workspace.
func (b *CheBackend) Status(opts Options) (WorkspaceStatus, error) {
	opts.defaults()

	wsName := cheWorkspaceName(opts)
	ws := WorkspaceStatus{
		Backend: BackendChe,
		Name:    wsName,
		Mode:    ModePersistent,
	}

	if b.useChectl {
		out, err := opts.ExecCmd("chectl", "workspace:list",
			"--output=json")
		if err != nil {
			return ws, nil
		}
		return parseCheWorkspaceList(out, wsName, ws)
	}

	// REST API path.
	url := strings.TrimRight(b.cheURL, "/") +
		"/api/workspace/" + wsName
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ws, nil
	}
	addCheAuth(req, opts)

	resp, err := opts.HTTPDo(req)
	if err != nil {
		return ws, nil
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	return parseCheWorkspaceJSON(body, ws)
}

// Attach connects the TUI to the CDE workspace's
// OpenCode server via the Che endpoint URL.
func (b *CheBackend) Attach(opts Options) error {
	opts.defaults()

	if _, err := opts.LookPath("opencode"); err != nil {
		return fmt.Errorf(
			"opencode not found. Install: brew install anomalyco/tap/opencode")
	}

	// Get the workspace status to find the server URL.
	ws, err := b.Status(opts)
	if err != nil {
		return fmt.Errorf("get workspace status: %w", err)
	}

	serverURL := ws.ServerURL
	if serverURL == "" {
		serverURL = fmt.Sprintf("http://localhost:%d", DefaultServerPort)
	}

	return opts.ExecInteractive("opencode", "attach", serverURL)
}

// cheWorkspaceName returns the CDE workspace name for
// the current project: "uf-<project-name>".
func cheWorkspaceName(opts Options) string {
	if opts.WorkspaceName != "" {
		return opts.WorkspaceName
	}
	return "uf-" + projectName(opts.ProjectDir)
}

// addCheAuth adds the authorization header to a Che
// REST API request if a token is available.
func addCheAuth(req *http.Request, opts Options) {
	cfg, _ := LoadConfig(opts)
	token := cfg.Che.Token
	if envToken := opts.Getenv("UF_CHE_TOKEN"); envToken != "" {
		token = envToken
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

// cheWorkspaceInfo is the subset of Che workspace JSON
// that we parse.
type cheWorkspaceInfo struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Config struct {
		Name string `json:"name"`
	} `json:"config"`
	Runtime struct {
		Machines map[string]struct {
			Servers map[string]struct {
				URL string `json:"url"`
			} `json:"servers"`
		} `json:"machines"`
	} `json:"runtime"`
}

// parseCheWorkspaceList parses chectl workspace:list JSON
// output and finds the matching workspace.
func parseCheWorkspaceList(data []byte, wsName string, ws WorkspaceStatus) (WorkspaceStatus, error) {
	var workspaces []cheWorkspaceInfo
	if err := json.Unmarshal(data, &workspaces); err != nil {
		return ws, nil
	}

	for _, w := range workspaces {
		if w.Config.Name == wsName {
			ws.Exists = true
			ws.ID = w.ID
			ws.Running = strings.EqualFold(w.Status, "RUNNING")
			ws.Persistent = true
			ws.ServerURL = extractCheServerURL(w)
			ws.DemoEndpoints = extractCheEndpoints(w)
			return ws, nil
		}
	}

	return ws, nil
}

// parseCheWorkspaceJSON parses a single Che workspace
// JSON response.
func parseCheWorkspaceJSON(data []byte, ws WorkspaceStatus) (WorkspaceStatus, error) {
	var w cheWorkspaceInfo
	if err := json.Unmarshal(data, &w); err != nil {
		return ws, nil
	}

	ws.Exists = true
	ws.ID = w.ID
	ws.Running = strings.EqualFold(w.Status, "RUNNING")
	ws.Persistent = true
	ws.ServerURL = extractCheServerURL(w)
	ws.DemoEndpoints = extractCheEndpoints(w)
	return ws, nil
}

// extractCheServerURL finds the OpenCode server URL from
// Che workspace runtime info.
func extractCheServerURL(w cheWorkspaceInfo) string {
	for _, machine := range w.Runtime.Machines {
		for name, server := range machine.Servers {
			if strings.Contains(name, "opencode") || strings.Contains(name, "4096") {
				return server.URL
			}
		}
	}
	return ""
}

// extractCheEndpoints extracts demo endpoints from Che
// workspace runtime info.
func extractCheEndpoints(w cheWorkspaceInfo) []DemoEndpoint {
	var endpoints []DemoEndpoint
	for _, machine := range w.Runtime.Machines {
		for name, server := range machine.Servers {
			// Skip the OpenCode server itself.
			if strings.Contains(name, "opencode") || strings.Contains(name, "4096") {
				continue
			}
			if server.URL != "" {
				protocol := "https"
				if strings.HasPrefix(server.URL, "http://") {
					protocol = "http"
				}
				endpoints = append(endpoints, DemoEndpoint{
					Name:     name,
					URL:      server.URL,
					Protocol: protocol,
				})
			}
		}
	}
	return endpoints
}
