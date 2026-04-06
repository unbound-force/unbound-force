## 1. Create Agent Files

- [x] 1.1 Create `.opencode/agents/divisor-scribe.md`
  following the `divisor-architect.md` structural
  pattern: YAML frontmatter (description, mode: subagent,
  model, temperature: 0.1, tools: write+edit true,
  bash false), Role heading, Step 0: Prior Learnings
  (Dewey), content creation workflows for technical
  documentation, reference to content.md pack (TD
  section), Out of Scope section
- [x] 1.2 Create `.opencode/agents/divisor-herald.md`
  following the same pattern: temperature: 0.4,
  workflows for blog posts, release notes, feature
  announcements, changelog entries, reference to
  content.md pack (BA section)
- [x] 1.3 Create `.opencode/agents/divisor-envoy.md`
  following the same pattern: temperature: 0.5,
  workflows for press releases, social media, community
  updates, partnership communications, reference to
  content.md pack (PR section)

## 2. Create Convention Packs

- [x] 2.1 Create `.opencode/unbound/packs/content.md`
  with YAML frontmatter (pack_id: content, language:
  Any, version: 1.0.0), sections for Voice & Brand
  (VB-NNN), Technical Documentation (TD-NNN), Blog &
  Announcements (BA-NNN), Public Relations (PR-NNN),
  Fact-Checking & Accuracy (FA-NNN), Formatting (FT-NNN).
  Model after website repo's `markdown.md` pack structure.
- [x] 2.2 Create `.opencode/unbound/packs/content-custom.md`
  user-owned stub following `default-custom.md` pattern

## 3. Scaffold Asset Registration

- [x] 3.1 Copy `.opencode/agents/divisor-scribe.md` to
  `internal/scaffold/assets/opencode/agents/divisor-scribe.md`
- [x] 3.2 Copy `.opencode/agents/divisor-herald.md` to
  `internal/scaffold/assets/opencode/agents/divisor-herald.md`
- [x] 3.3 Copy `.opencode/agents/divisor-envoy.md` to
  `internal/scaffold/assets/opencode/agents/divisor-envoy.md`
- [x] 3.4 Copy `.opencode/unbound/packs/content.md` to
  `internal/scaffold/assets/opencode/unbound/packs/content.md`
- [x] 3.5 Copy `.opencode/unbound/packs/content-custom.md` to
  `internal/scaffold/assets/opencode/unbound/packs/content-custom.md`
- [x] 3.6 Update `expectedAssetPaths` in
  `internal/scaffold/scaffold_test.go` to include the
  5 new files (3 agents + 2 packs)

## 4. Pack Deployment Logic

- [x] 4.1 Update `shouldDeployPack()` in
  `internal/scaffold/scaffold.go` to always deploy
  `content` and `content-custom` packs (add to the
  always-deploy condition alongside `default`,
  `default-custom`, and `severity`)

## 5. Documentation

- [x] 5.1 Update `AGENTS.md` — add the 3 new agents to
  the agent file listing in Project Structure section
- [x] 5.2 Update `AGENTS.md` — add the 2 new packs to
  the convention packs listing
- [x] 5.3 Add Recent Changes entry to `AGENTS.md`

## 6. Verification

- [x] 6.1 Run `go test -race -count=1 ./internal/scaffold/...`
  to verify drift detection and asset count tests pass
- [x] 6.2 Run `go build ./...` to confirm clean
  compilation
- [x] 6.3 Verify constitution alignment: each agent is
  independently invocable (Composability), produces
  self-describing output (Autonomous Collaboration),
  and is testable via scaffold drift detection
  (Testability)
