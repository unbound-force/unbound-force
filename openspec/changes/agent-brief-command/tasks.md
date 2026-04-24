## 1. /agent-brief Command File

- [x] 1.1 Create `.opencode/command/agent-brief.md` with frontmatter (`description` field describing create/audit/improve lifecycle)
- [x] 1.2 Write mode detection logic (Step 1): check AGENTS.md existence, parse optional arguments (`create`, `audit`), route to appropriate mode
- [x] 1.3 Write project analysis step (Step 2): scan for language markers (`go.mod`, `package.json`, `Cargo.toml`, `pyproject.toml`), build system (`Makefile`, `justfile`), CI config (`.github/workflows/`), linter config (`.golangci.yml`, `ruff.toml`, `.eslintrc`), README.md, LICENSE, `.git/config` remote URL
- [x] 1.4 Write governance detection (Step 2 continued): scan for `.specify/memory/constitution.md`, `specs/` with `NNN-*/` subdirs, `openspec/config.yaml`, `.opencode/uf/packs/`, `opencode.json` MCP servers
- [x] 1.5 Write create mode flow (Step 3): generate Tier 1 sections from analysis (Project Overview from README, Build & Test Commands from Makefile/CI, Project Structure from directory listing, Code Conventions from linter config + language defaults, Technology Stack from dependency files)
- [x] 1.6 Write Tier 1C generation in create mode: Constitution section (read `.specify/memory/constitution.md`, summarize principles, state cross-framework applicability), Spec Framework section (describe detected frameworks, state constitution governs both when both present)
- [x] 1.7 Write Tier 2 stub generation in create mode: Architecture, Testing Conventions, Git & Workflow, Behavioral Constraints -- each with TODO guidance comments explaining what to fill and why it matters
- [x] 1.8 Write audit mode flow (Step 4): read AGENTS.md, check each section against detection patterns (per design D7), count lines, check for code blocks in build section, detect staleness (verify directories in Project Structure tree exist)
- [x] 1.9 Write context-sensitive audit checks: constitution reference when `.specify/memory/constitution.md` exists, spec framework when `specs/` or `openspec/` exist, cross-framework governance bridge when both frameworks detected
- [x] 1.10 Write audit scoring and report generation: tier completion metrics, quality metrics table, overall effectiveness label (Excellent/Strong/Adequate/Weak/Missing per design D3), improvement suggestions with specific content for missing sections
- [x] 1.11 Write improvement application flow: present suggestions, ask user confirmation, apply accepted changes (insert missing sections, update stale content)
- [x] 1.12 Write bridge file handling (Step 5): verify CLAUDE.md contains `@AGENTS.md`, verify .cursorrules references AGENTS.md, create or update missing bridge files following `uf init` patterns
- [x] 1.13 Write guardrails section: NEVER modify files outside AGENTS.md and bridge files, NEVER implement code, present all generated content to user before writing, respect existing project-specific sections

## 2. Scaffold Asset Integration

- [x] 2.1 Copy command file to `internal/scaffold/assets/opencode/command/agent-brief.md` (scaffold asset copy)
- [x] 2.2 Add `"opencode/command/agent-brief.md"` to `expectedAssetPaths` in `internal/scaffold/scaffold_test.go`
- [x] 2.3 Update the OpenCode commands count comment in `expectedAssetPaths` (increment by 1)
- [x] 2.4 Verify `isToolOwned()` already covers `opencode/command/` paths (no change expected -- existing pattern handles this)
- [x] 2.5 Verify `mapAssetPath()` already maps `opencode/` to `.opencode/` (no change expected)

## 3. Doctor Check Group

- [x] 3.1 Create `checkAgentContext(opts *Options) CheckGroup` function in `internal/doctor/checks.go` with group name "Agent Context"
- [x] 3.2 Implement AGENTS.md existence check (check #1): `os.Stat()`, Fail severity, InstallHint `"Run: /agent-brief in OpenCode"`
- [x] 3.3 Implement section header detection helper: `detectAGENTSmdSections(content []byte) map[string]bool` with regex patterns per design D7 for all Tier 1 and Tier 2 sections
- [x] 3.4 Implement Tier 1 section presence checks (checks #2-6): iterate over 5 required sections (Project Overview, Build Commands, Project Structure, Code Conventions, Technology Stack), Fail severity for each missing section
- [x] 3.5 Implement build section code block check (check #7): scan content between the Build section header and the next `##` header for triple-backtick fenced code blocks, Warn severity if no code blocks found
- [x] 3.6 Implement line count check (check #8): count newlines in content, Warn if >300 with InstallHint suggesting `/agent-brief` for condensing suggestions
- [x] 3.7 Implement constitution reference check (check #9): only when `.specify/memory/constitution.md` exists (via `os.Stat()`), scan for `(?i)constitution` in AGENTS.md content, Warn severity if absent
- [x] 3.8 Implement spec framework description check (check #10): only when `specs/` has `NNN-*/` subdirs or `openspec/config.yaml` exists, scan for `(?i)(speckit|openspec|spec.*framework|specification.*framework)`, Warn severity if absent
- [x] 3.9 Implement CLAUDE.md bridge check (check #11): read CLAUDE.md via `opts.ReadFile()`, check for `@AGENTS.md` or `AGENTS.md` reference, Warn severity if missing, InstallHint `"Run: /agent-brief in OpenCode"`
- [x] 3.10 Implement .cursorrules bridge check (check #12): read .cursorrules via `opts.ReadFile()`, check for `AGENTS.md` reference, Warn severity if missing
- [x] 3.11 Remove AGENTS.md existence check from `checkScaffoldedFiles()` (lines 483-498 of checks.go) to avoid duplication with the new check group
- [x] 3.12 Register `checkAgentContext(&opts)` in the `groups` slice in `Run()` function of `internal/doctor/doctor.go` (insert before `checkScaffoldedFiles` for logical grouping)

## 4. Doctor Tests

- [x] 4.1 Add `TestCheckAgentContext_NoFile` -- verify Fail result with correct InstallHint when AGENTS.md absent
- [x] 4.2 Add `TestCheckAgentContext_AllTier1Present` -- verify 5 Pass results when all Tier 1 section headers present
- [x] 4.3 Add `TestCheckAgentContext_MissingTier1Section` -- verify Fail result for each missing Tier 1 section individually
- [x] 4.4 Add `TestCheckAgentContext_BuildCodeBlocks` -- verify Warn when build section has no code blocks, Pass when code blocks present
- [x] 4.5 Add `TestCheckAgentContext_LineCount` -- verify Pass for <=300 lines, Warn for >300 lines
- [x] 4.6 Add `TestCheckAgentContext_ConstitutionReference` -- verify check runs only when `.specify/memory/constitution.md` exists, Warn when AGENTS.md doesn't reference it, Pass when it does
- [x] 4.7 Add `TestCheckAgentContext_ConstitutionSkipped` -- verify check is omitted from results when no `.specify/` directory exists
- [x] 4.8 Add `TestCheckAgentContext_SpecFrameworkReference` -- verify check runs when `specs/` or `openspec/` detected, Warn when AGENTS.md doesn't describe framework, Pass when it does
- [x] 4.9 Add `TestCheckAgentContext_SpecFrameworkSkipped` -- verify check omitted when neither `specs/` nor `openspec/` exist
- [x] 4.10 Add `TestCheckAgentContext_BridgeCLAUDEmd` -- verify Warn when CLAUDE.md missing or lacks `@AGENTS.md`, Pass when properly configured
- [x] 4.11 Add `TestCheckAgentContext_BridgeCursorrules` -- verify Warn when .cursorrules missing or lacks AGENTS.md reference, Pass when properly configured
- [x] 4.12 Add `TestCheckAgentContext_FullPass` -- integration test with complete AGENTS.md, constitution, spec dirs, and bridge files -- verify all 12 checks Pass

## 5. Scaffold Tests

- [x] 5.1 Update `expectedAssetPaths` count assertion in `TestDriftDetection_AllCanonicalAssetsEmbedded` (increment expected file count by 1)
- [x] 5.2 Add `TestRunInit_DeploysAgentBrief` -- verify `uf init` creates `.opencode/command/agent-brief.md` in target directory
- [x] 5.3 Verify scaffold drift detection passes (embedded asset matches live copy at `.opencode/command/agent-brief.md`)

## 6. Documentation Updates

- [x] 6.1 Update AGENTS.md "Project Structure" tree: add `agent-brief.md` under `.opencode/command/` entries
- [x] 6.2 Update AGENTS.md "Recent Changes" section: add entry summarizing the `/agent-brief` command and doctor changes
- [x] 6.3 Fix AGENTS.md doctor install hint documentation if any prose references the old "Run: uf init" hint

## 7. Verification

- [x] 7.1 Run `make check` (build, test, vet, lint) and verify all checks pass
- [x] 7.2 Verify scaffold drift detection: embedded asset at `internal/scaffold/assets/opencode/command/agent-brief.md` matches live copy at `.opencode/command/agent-brief.md`
- [x] 7.3 Verify constitution alignment: review all new code against the 4 principles (Autonomous Collaboration, Composability First, Observable Quality, Testability) -- confirm PASS assessment from proposal holds
