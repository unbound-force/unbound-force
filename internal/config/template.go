// SPDX-License-Identifier: Apache-2.0

package config

// Template returns the full commented-out YAML template for the
// current config version. All 8 sections are present with inline
// comments documenting valid values and defaults.
func Template() string {
	return `# .uf/config.yaml
# Unbound Force project configuration.
# All values shown are defaults — only uncomment what you want to change.
# CLI flags and environment variables override these settings.
# Env var overrides: UF_PACKAGE_MANAGER, OLLAMA_MODEL, OLLAMA_HOST,
#   UF_SANDBOX_IMAGE, UF_SANDBOX_BACKEND, UF_SANDBOX_RUNTIME,
#   UF_CHE_URL, UF_CHE_TOKEN, UF_GATEWAY_PORT, UF_GATEWAY_PROVIDER,
#   UF_OLLAMA_PROXY_PORT, UF_OLLAMA_PROXY_EMBED_MODEL, UF_OLLAMA_PROXY_GATEWAY_URL

# ─── Setup Preferences ───────────────────────────────────────
# Controls how ` + "`uf setup`" + ` installs tools.
# setup:
#   package_manager: auto        # auto | homebrew | dnf | apt | manual
#   skip: []                     # tool names to skip: [ollama, dewey, ...]
#   tools:                       # per-tool install overrides
#     opencode:
#       method: auto             # auto | homebrew | curl | skip
#     gaze:
#       method: auto             # auto | homebrew | rpm | skip
#     node:
#       method: auto             # auto | nvm | fnm | mise | homebrew | skip
#       version: "22"            # target version when installing
#     gh:
#       method: auto             # auto | homebrew | skip
#     ollama:
#       method: auto             # auto | homebrew | skip
#     dewey:
#       method: auto             # auto | homebrew | skip
#     replicator:
#       method: auto             # auto | homebrew | skip

# ─── Scaffold Preferences ────────────────────────────────────
# Controls what ` + "`uf init`" + ` deploys.
# scaffold:
#   language: auto               # auto | go | typescript | python | rust

# ─── Embedding ────────────────────────────────────────────────
# Controls the embedding model used by Dewey.
# embedding:
#   model: granite-embedding:30m
#   dimensions: 256
#   provider: ollama             # ollama (only supported today)
#   host: http://localhost:11434

# ─── Sandbox ──────────────────────────────────────────────────
# Controls ` + "`uf sandbox`" + ` behavior.
# sandbox:
#   runtime: auto                # auto | podman | docker
#   backend: auto                # auto | podman | che
#   image: quay.io/unbound-force/opencode-dev:latest
#   resources:
#     memory: 8g
#     cpus: "4"
#   mode: isolated               # isolated | direct
#   che:
#     url: ""
#     token: ""
#   demo_ports: []

# ─── Gateway ──────────────────────────────────────────────────
# Controls ` + "`uf gateway`" + ` behavior.
# gateway:
#   port: 53147
#   provider: auto               # auto | anthropic | vertex | bedrock

# ─── Ollama Proxy ─────────────────────────────────────────────
# Controls ` + "`uf ollama-proxy`" + ` behavior.
# ollama_proxy:
#   port: 11434                  # local port (matches Ollama default)
#   embed_model: text-embedding-005  # Vertex AI embedding model
#   gateway_url: http://localhost:53147  # uf gateway URL for generation

# ─── Doctor ───────────────────────────────────────────────────
# Controls ` + "`uf doctor`" + ` check behavior.
# doctor:
#   skip: []                     # check names to skip
#   tools:                       # override tool severity
#     gaze: recommended          # required | recommended | optional

# ─── Workflow ─────────────────────────────────────────────────
# Controls hero lifecycle workflow.
# workflow:
#   execution_modes:
#     define: human               # human | swarm
#     implement: swarm
#     validate: swarm
#     review: swarm
#     accept: human
#     reflect: swarm
#   spec_review: false
`
}

// knownSections lists the top-level section names in the current
// config template. Used by InitFile to detect added/removed
// sections.
var knownSections = []string{
	"setup",
	"scaffold",
	"embedding",
	"sandbox",
	"gateway",
	"ollama_proxy",
	"doctor",
	"workflow",
}
