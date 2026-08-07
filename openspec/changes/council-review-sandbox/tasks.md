<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Runtime Sandbox (run-review.sh)

- [x] 1.1 Build `PERMISSION_CONFIG` JSON with denials for `edit`,
  `bash`, `webfetch`, `websearch`, `skill`. Do NOT deny `task`
  (required for multi-agent subagent invocation, see D3).
- [x] 1.2 Merge permission config into `OPENCODE_CONFIG_CONTENT`
  alongside existing Vertex AI provider config (for
  `google-vertex-anthropic`). For non-Vertex providers, export
  permission config alone.
- [x] 1.3 Remove non-existent `--permissions read,glob,grep` flag
  from the `opencode run` invocation
- [x] 1.4 Add `--pure` flag to skip external MCP plugins
- [x] 1.5 Deny `external_directory` and `doom_loop` in permission
  config (the only ask-default permissions) to avoid TTY hangs
- [x] 1.6 Update inline comments explaining the defense-in-depth
  layers and referencing nunya#406

## 2. Agent Frontmatter Migration (tools: to permission:)

Review agents -- deny edit, bash, webfetch:

- [x] 2.1 [P] Migrate `.opencode/agents/divisor-adversary.md`
  from `tools:` to `permission: { edit: deny, bash: deny,
  webfetch: deny }`
- [x] 2.2 [P] Migrate `.opencode/agents/divisor-architect.md`
  from `tools:` to `permission: { edit: deny, bash: deny,
  webfetch: deny }`
- [x] 2.3 [P] Migrate `.opencode/agents/divisor-guard.md`
  from `tools:` to `permission: { edit: deny, bash: deny,
  webfetch: deny }`
- [x] 2.4 [P] Migrate `.opencode/agents/divisor-sre.md`
  from `tools:` to `permission: { edit: deny, bash: deny,
  webfetch: deny }`
- [x] 2.5 [P] Migrate `.opencode/agents/divisor-testing.md`
  from `tools:` to `permission: { edit: deny, bash: deny,
  webfetch: deny }`

Content agents -- deny only what is inappropriate:

- [x] 2.6 [P] Migrate `.opencode/agents/divisor-curator.md`
  from `tools:` to `permission: { edit: deny, webfetch: deny }`
  (bash allowed for interactive use)
- [x] 2.7 [P] Migrate `.opencode/agents/divisor-envoy.md`
  from `tools:` to `permission: { bash: deny, webfetch: deny }`
  (edit allowed for content creation)
- [x] 2.8 [P] Migrate `.opencode/agents/divisor-herald.md`
  from `tools:` to `permission: { bash: deny, webfetch: deny }`
  (edit allowed for content creation)
- [x] 2.9 [P] Migrate `.opencode/agents/divisor-scribe.md`
  from `tools:` to `permission: { bash: deny, webfetch: deny }`
  (edit allowed for documentation)

## 3. Documentation

- [x] 3.1 [P] Add "Runtime sandbox (defense-in-depth)" decision to
  `council-review-action/docs/decisions.md` -- document
  `OPENCODE_CONFIG_CONTENT`, `--pure`, why ask-default
  permissions are denied explicitly, why task is NOT
  denied, rationale for rejecting `--agent plan`, agent
  frontmatter as defense-in-depth layer
- [x] 3.2 [P] Update T2 (network exfiltration) in
  `council-review-action/docs/security-risks.md` -- replace
  `--permissions` references with `OPENCODE_CONFIG_CONTENT`
  permission config
- [x] 3.3 Update A3 (secrets in environment) in
  `council-review-action/docs/security-risks.md` -- defense-in-
  depth table updated, replace `--permissions` with
  `OPENCODE_CONFIG_CONTENT`, update `tools:` references to
  `permission:`

## 4. Scaffold Asset Sync

- [x] 4.0 Copy migrated `divisor-*.md` agents to
  `internal/scaffold/assets/opencode/agents/` to fix
  `TestEmbeddedAssets_MatchSource` drift detection

## 5. Validation

- [x] 5.1 Verify `opencode run --help` confirms `--pure` and
  flag exists (pre-flight check)
- [x] 5.2 Verify `OPENCODE_CONFIG_CONTENT` with permission config
  is accepted by OpenCode (local smoke test)
- [x] 5.3 Verify no `tools:` frontmatter remains in any
  `divisor-*.md` file
- [x] 5.4 Verify no reference to `--permissions` flag remains in
  code or docs (except as historical context in decisions.md)
- [ ] 5.5 E2E test PR to confirm review quality is maintained
  (deferred to integration testing via org-infra)
