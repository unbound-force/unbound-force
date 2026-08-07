## Context

The council-review-action invokes `opencode run` to perform AI code
review using Divisor personas. The review processes untrusted PR diff
content. Prior to this change, the only protection against tool
misuse was prompt-level instruction ("do not use shell"). This is a
soft boundary -- the model can ignore instructions via prompt
injection.

A prior attempt (commit `a74535c`) added `--permissions
read,glob,grep` to the invocation. This flag does not exist in
OpenCode's CLI (`opencode run --help` confirms no `--permissions`
flag). The invocation silently failed or ignored the flag, leaving
the sandbox unenforced.

OpenCode's actual permission mechanism is config-based:
1. `opencode.json` or `OPENCODE_CONFIG_CONTENT` env var with
   `"permission"` key (hard deny/allow rules)
2. `--pure` CLI flag (plugin isolation, confirmed present)
3. Explicit deny of all `"ask"`-default permissions
   (`external_directory`, `doom_loop`) so no TTY prompt is needed
4. Agent frontmatter `permission:` (per-agent restrictions,
   replaces deprecated `tools:`)

## Goals / Non-Goals

### Goals

- Enforce tool restrictions at the OpenCode runtime level using
  the correct mechanism (`OPENCODE_CONFIG_CONTENT` permission
  config)
- Prevent external MCP plugins from bypassing restrictions
- Migrate Divisor agent frontmatter from deprecated `tools:` to
  `permission:` syntax
- Preserve multi-agent review (orchestrator MUST be able to
  invoke Divisor subagents via Task tool)
- Document sandboxing decisions and update security risk register

### Non-Goals

- Network-level egress blocking (tracked in org-infra#429, uses
  harden-runner -- independent of this change)
- Token consumption controls (tracked in org-infra#430)
- Comment-triggered invocation (tracked in org-infra#429)
- Migrating non-Divisor agents (constitution-check, gaze-reporter,
  gaze-test-generator, muti-mind-po) -- separate housekeeping

## Decisions

### D1: `OPENCODE_CONFIG_CONTENT` over non-existent `--permissions`

**Decision**: Inject runtime permission config via the
`OPENCODE_CONFIG_CONTENT` env var rather than a CLI flag.

**Why**: OpenCode has no `--permissions` CLI flag. Permissions are
configured via `opencode.json` or env var. `OPENCODE_CONFIG_CONTENT`
is documented as "runtime overrides" at precedence level 6 (above
project config). It merges with (not replaces) the project
`opencode.json`, so existing config is preserved.

`run-review.sh` already uses `OPENCODE_CONFIG_CONTENT` to inject
Vertex AI provider config for the `google-vertex-anthropic`
provider. The permission config merges into the same JSON object.

### D2: Deny list, not allowlist

**Decision**: Deny specific dangerous tools (`edit`, `bash`,
`webfetch`, `websearch`, `skill`) rather than attempting to
allowlist only safe tools.

**Why**: OpenCode's permission defaults allow `read`, `glob`,
`grep`. Denying the dangerous tools achieves the same result as
an allowlist while being more explicit about what is blocked.
OpenCode evaluates permission rules with last-match-wins, so
explicit denials override defaults.

### D3: Do NOT deny `task` -- required for multi-agent review

**Decision**: The permission config MUST NOT include
`"task": "deny"`.

**Why**: The council review orchestrator invokes Divisor subagents
via OpenCode's Task tool. The OpenCode docs state: "When set to
deny, the subagent is removed from the Task tool description
entirely, so the model won't attempt to invoke it." Denying `task`
would silently degrade multi-agent review to single-agent mode,
defeating the purpose of the council review.

The subagents themselves are sandboxed by their own `permission:`
frontmatter (edit: deny, bash: deny, webfetch: deny).

### D4: Explicit deny of ask-default permissions (no TTY)

**Decision**: Deny `external_directory` and `doom_loop` in the
permission config rather than using `--dangerously-skip-permissions`
or `--auto`.

**Why**: In CI there is no TTY for approval prompts. Only two
permissions default to `"ask"`: `external_directory` and
`doom_loop`. Rather than blanket-approving all non-denied
permissions via a skip flag, we deny these two explicitly. This
is more precise — only tools we've evaluated are allowed to run.

**Why not `--dangerously-skip-permissions` / `--auto`**: These
flags auto-approve everything not explicitly denied. If OpenCode
adds new tool types that default to `"ask"`, they would be
silently approved. Explicit denials require conscious evaluation
of each new tool type.

### D5: Skip `--agent plan` (incompatible with Divisor discovery)

**Decision**: Do not use `--agent plan` or
`OPENCODE_EXPERIMENTAL_PLAN_MODE`.

**Why**: `--agent plan` selects the built-in Plan agent, replacing
custom Divisor agent discovery. This breaks multi-persona review.
`OPENCODE_EXPERIMENTAL_PLAN_MODE` is "an instruction to the model,
not a hard sandbox" (OpenCode docs) -- the permission config
already provides the hard boundary, making plan mode redundant.

### D6: Migrate `tools:` to `permission:` in agent frontmatter

**Decision**: Replace deprecated `tools:` frontmatter with
`permission:` syntax in all 9 Divisor agents.

**Why**: OpenCode docs state "the legacy `tools` boolean config is
deprecated and has been merged into `permission`." While `tools:`
still works, migrating now avoids future breakage and aligns with
the permission config mechanism used in `run-review.sh`.

The migration preserves semantics:
- `tools: { bash: false }` becomes `permission: { bash: deny }`
- `tools: { write: false }` is covered by `permission: { edit: deny }`
  (OpenCode's `edit` permission governs `edit`, `write`, and `patch`
  tools — see https://opencode.ai/docs/permissions/)
- `tools: { read: true }` is dropped (read defaults to allow)
- Review agents: deny edit, bash, webfetch
- Content agents (curator, envoy, herald, scribe): deny only
  what they shouldn't use for interactive work

### D7: Config merge interaction with project `opencode.json`

**Decision**: Accept that `OPENCODE_CONFIG_CONTENT` merges with
the repo's `opencode.json` (which defines MCP servers for dewey
and replicator). The `--pure` flag prevents MCP servers from
loading regardless.

**Why**: OpenCode docs confirm "Configuration files are merged
together, not replaced" and `OPENCODE_CONFIG_CONTENT` is
precedence level 6 (overrides project config for conflicting
keys). The MCP servers in `opencode.json` would be present in the
merged config but `--pure` prevents them from loading. This is
the correct layered behavior.

## Risks / Trade-offs

**[R1] `OPENCODE_CONFIG_CONTENT` merge semantics**: If OpenCode
changes how `OPENCODE_CONFIG_CONTENT` merges with project config,
permission denials could be overridden. Mitigated by version
pinning (`opencode-ai@1.15.13`).

**[R2] Review depth reduction**: Denying `bash` means the model
cannot run read-only git commands (e.g., `git log`). The pre-
fetched context and diff file contain sufficient information for
code review. If review quality degrades, `bash` could be
selectively re-enabled with granular rules (e.g.,
`"bash": { "*": "deny", "git log*": "allow" }`).

**[R3] `--pure` blocks legitimate plugins**: If a future review
enhancement relies on an MCP plugin, `--pure` would block it.
Evaluate plugin needs when they arise.

**[R4] New ask-default permissions**: If OpenCode adds new tool types
that default to `"ask"`, they would cause TTY hangs in CI. Any tool not
explicitly denied runs without approval. The deny list must be
kept current as OpenCode adds new tool types. Mitigated by
version pinning and the pre-collected context approach (Layer 1)
that eliminates the need for most tools.
