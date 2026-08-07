## Why

The council-review-action runs untrusted PR diff content through an
LLM on Vertex AI. Prior to this change, the only protection against
tool misuse was prompt-level instruction ("do not use shell"). This
is a soft boundary -- the model can ignore instructions via prompt
injection.

OpenCode provides runtime enforcement mechanisms:
1. Permission config via `opencode.json` or `OPENCODE_CONFIG_CONTENT`
   (hard tool deny/allow rules enforced by the runtime)
2. `--pure` CLI flag (plugin isolation)
3. Agent frontmatter `permission:` config (per-agent restrictions)

The existing agent frontmatter uses deprecated `tools:` syntax
which still works but should be migrated to `permission:`.

OpenCode has no `--permissions` CLI flag. A prior commit (`a74535c`)
incorrectly added `--permissions read,glob,grep` to the `opencode
run` invocation -- this flag does not exist, making the invocation
silently fail or ignore the flag. This change replaces that broken
mechanism with the correct approach.

Tracked as complytime/nunya#406 (defense-in-depth proposal).

## What Changes

- Replace non-existent `--permissions read,glob,grep` flag with
  `OPENCODE_CONFIG_CONTENT` permission config in `run-review.sh`
  that denies bash, edit, webfetch, websearch, and skill
- Add `--pure` flag to skip external MCP plugins
- Explicitly deny `external_directory` and `doom_loop` (the only
  permissions that default to `"ask"`) to avoid TTY hangs in CI
- Migrate all 9 Divisor agents from deprecated `tools:` frontmatter
  to `permission:` syntax
- Document the sandboxing decision in `decisions.md`
- Update `security-risks.md` to reflect implemented mitigations

## Capabilities

### New Capabilities

- `runtime-permission-config`: Hard tool boundary via
  `OPENCODE_CONFIG_CONTENT` env var that denies edit (covers
  write/edit/patch), bash, webfetch, websearch, and skill
  regardless of prompt instructions or agent config
- `plugin-isolation`: `--pure` flag prevents external MCP plugins
  from loading, closing the plugin bypass vector

### Modified Capabilities

- `divisor-agent-frontmatter`: All 9 Divisor agents migrated from
  deprecated `tools:` to `permission:` syntax with appropriate
  deny rules per agent role
- `security-risk-register`: T2 and A3 entries updated from
  "tracked" to "implemented"

### Removed Capabilities

(none)

## Impact

- `council-review-action/scripts/run-review.sh` -- permission
  config injection, `--pure`, explicit ask-default denials,
  remove broken
  `--permissions` flag
- `.opencode/agents/divisor-*.md` (9 files) -- frontmatter
  migration from `tools:` to `permission:`
- `council-review-action/docs/decisions.md` -- new decision entry
- `council-review-action/docs/security-risks.md` -- updated
  mitigations
- No Go code changes, no workflow changes, no new dependencies

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The sandbox does not change how agents communicate. The Divisor
personas still produce self-describing JSON review output via
`--format json`. The permission config restricts tool access
but does not alter the artifact format or inter-agent protocol.

### II. Composability First

**Assessment**: PASS

The composite action remains independently consumable. The
permission config and `--pure` flag are self-contained in
`run-review.sh` -- consumers do not need to configure anything
extra. Repos without `uf init` still get the single-agent
fallback with the same sandbox.

### III. Observable Quality

**Assessment**: PASS

The review JSON output format is unchanged. The sandbox adds
observability: if a tool call is denied, OpenCode logs the
denial in stderr which is captured in `review_err.txt`.

### IV. Testability

**Assessment**: PASS

The permission config is a JSON structure injected via env var,
testable by inspecting `OPENCODE_CONFIG_CONTENT`. The agent
frontmatter changes are testable by parsing YAML. The `--pure`
flag is a CLI argument visible in the invocation. The
permission denials are in `OPENCODE_CONFIG_CONTENT` JSON.
