---
description: >
  Create, validate, and improve AGENTS.md -- the project briefing
  for AI coding agents. Auto-detects mode: creates from scratch
  when no AGENTS.md exists, audits and suggests improvements when
  one is present. Also ensures cross-tool bridge files (CLAUDE.md,
  .cursorrules) are properly configured.
---

# Command: /agent-brief

## Description

Manage the AGENTS.md lifecycle: create, audit, and improve the
project briefing that AI coding agents read at session start.

AGENTS.md is the API contract between humans and AI agents. It
tells agents how to build, test, and lint the project, what
conventions to follow, and what constraints to respect. Without
a good AGENTS.md, every agent session starts from cold context.

**Modes**:
- No AGENTS.md → **Create mode** (analyze project, generate file)
- AGENTS.md exists → **Audit mode** (score, report, suggest)
- `/agent-brief create` → Force create mode (overwrite)
- `/agent-brief audit` → Force audit mode (read-only)

## Instructions

### Step 1: Mode Detection

1. Check if AGENTS.md exists at the repository root.

2. Parse the user's argument (if any):
   - `create` → force create mode
   - `audit` → force audit mode
   - No argument → auto-detect based on file existence

3. Route to the appropriate mode:
   - **Create mode**: No AGENTS.md exists, or user passed `create`
   - **Audit mode**: AGENTS.md exists, or user passed `audit`

4. Announce the selected mode:
   - Create: `"No AGENTS.md found. Analyzing project to generate one..."`
   - Audit: `"Found AGENTS.md (N lines). Running quality audit..."`
   - Force create: `"Force-creating AGENTS.md. Existing file will be replaced after your review."`

### Step 2: Project Analysis

Regardless of mode, analyze the project to understand its
characteristics. Read the following files (skip any that do not
exist):

**Language & Dependencies**:
1. `go.mod` → Go project (extract module name, Go version, key deps)
2. `package.json` → Node/TypeScript project (extract name, scripts, key deps)
3. `Cargo.toml` → Rust project (extract name, edition, key deps)
4. `pyproject.toml` → Python project (extract name, version, key deps)
5. `tsconfig.json` → TypeScript confirmation

**Build System**:
6. `Makefile` or `justfile` → extract build/test/lint targets
7. `.github/workflows/` → read CI workflow files to find exact
   build, test, vet, and lint commands (these are the source of
   truth for build commands)

**Linter Configuration**:
8. `.golangci.yml` → Go linter rules
9. `ruff.toml` or `pyproject.toml [tool.ruff]` → Python linter
10. `.eslintrc*` or `eslint.config.*` → JavaScript/TypeScript linter

**Project Context**:
11. `README.md` → project description (first paragraph or heading)
12. `LICENSE` → license type
13. `.git/config` → remote URL for project/org name
14. Top-level directory listing → project structure

**Governance & Spec Framework** (Tier 1C triggers):
15. `.specify/memory/constitution.md` → constitution exists (read
    it to summarize principles)
16. `specs/` → check for `NNN-*/` subdirectories (Speckit)
17. `openspec/config.yaml` → OpenSpec configured
18. `.opencode/uf/packs/` → convention packs deployed
19. `opencode.json` → MCP servers configured

Record what was detected. This informs both create and audit
modes.

### Step 3: Create Mode

If in create mode, generate AGENTS.md using the project analysis.

#### 3a: Generate Tier 1 Sections (LLM-Filled)

For each Tier 1 section, read the relevant project files and
generate concrete, specific content. Do NOT use placeholder
text -- fill from actual project data.

**Project Overview** (2-5 lines):
- What the project is (from README first paragraph)
- Project type (CLI, library, web app, API, monorepo)
- Key domain context
- License

**Build & Test Commands**:
- Extract exact commands from Makefile targets or CI workflows
- Include the flags that matter (e.g., `-race -count=1` for Go)
- Use fenced code blocks (```) for all commands
- Group as: Build, Test, Lint (and any other common targets)

**Project Structure**:
- Generate a directory tree showing major directories
- Focus on top-level and one level deep
- Annotate each directory with its purpose
- Use the `text` code fence format

**Code Conventions**:
- Derive from linter config if present
- Include language-specific defaults:
  - Go: gofmt, goimports, error wrapping, import grouping
  - TypeScript: prettier, ESLint rules, naming conventions
  - Python: ruff/black, type hints, docstring style
  - Rust: clippy, formatting, error handling
- Include naming conventions, comment style, error handling

**Technology Stack**:
- Language and version (from go.mod, package.json, etc.)
- Key frameworks and libraries (from dependency files)
- Runtime requirements
- Version-pin important dependencies

#### 3b: Generate Tier 1C Sections (Context-Sensitive)

Only generate these sections when their triggers are detected:

**Constitution / Governance** (when `.specify/memory/constitution.md`
exists):
- Read the constitution file
- Summarize each principle in one line
- State that the constitution is the highest-authority document
- If both Speckit and OpenSpec are detected, explicitly state:
  "All work -- regardless of whether it uses the Speckit or
  OpenSpec workflow -- MUST align with these principles."

**Specification Framework** (when `specs/` or `openspec/` exist):
- Describe which framework(s) are in use
- Show the directory locations
- If both exist, include a comparison table:
  | Tier | Tool | Location | When to Use |
  |------|------|----------|-------------|
  | Strategic | Speckit | `specs/NNN-*/` | 3+ stories, cross-repo |
  | Tactical | OpenSpec | `openspec/changes/*/` | <3 stories, bug fix |
- State the governance bridge: "Both tiers share the org
  constitution as their governance bridge."

#### 3c: Generate Tier 2 Stubs

For each Tier 2 section, create a header with a TODO guidance
comment explaining what to fill and why it matters:

**Architecture / Key Patterns**:
```markdown
## Architecture
<!-- TODO: Describe the dominant design patterns in this project.
     Examples: "Options/Result structs", "Clean Architecture layers",
     "Repository pattern for data access", "Dependency injection via
     function fields". Agents use this to avoid introducing
     inconsistent patterns. -->
```

**Testing Conventions**:
```markdown
## Testing Conventions
<!-- TODO: Document testing framework, naming patterns, isolation
     requirements. Example: "Standard library testing package only.
     Test names: TestFoo_Description. Use t.TempDir() for filesystem
     tests. No assertion libraries." -->
```

**Git & Workflow**:
```markdown
## Git & Workflow
<!-- TODO: Commit format, branching strategy, PR requirements.
     Example: "Conventional commits (type: description), feature
     branches from main, 1 approving review required." -->
```

**Behavioral Constraints**:
```markdown
## Behavioral Constraints
<!-- TODO: Things agents must NEVER do. Negative instructions are
     often more impactful than positive ones.
     Examples:
     - "Never modify coverage thresholds to make tests pass."
     - "Never commit .env files or credentials."
     - "Never use os.Exit() in library code." -->
```

#### 3d: Present and Write

1. Show the complete generated AGENTS.md to the user.
2. Ask: "Does this look good? I can write it now, or you can
   suggest changes first."
3. On confirmation, write the file to `AGENTS.md` at repo root.
4. Proceed to Step 5 (Bridge Files).

### Step 4: Audit Mode

If in audit mode, read the existing AGENTS.md and evaluate it
against the context-sensitive section taxonomy.

#### 4a: Section Detection

Scan the file for section headers matching these patterns. A
section is "found" if any of its patterns match a `##` header
line (case-insensitive):

| Section | Tier | Detection Patterns |
|---------|------|--------------------|
| Project Overview | 1 | `overview`, `about` |
| Build & Test Commands | 1 | `build`, `development` |
| Project Structure | 1 | `structure`, `layout`, `directory` |
| Code Conventions | 1 | `convention`, `coding standard`, `style guide`, `coding convention` |
| Technology Stack | 1 | `technolog`, `tech stack`, `stack`, `active technologies` |
| Constitution | 1C | `constitution`, `governance` |
| Spec Framework | 1C | `specification`, `spec framework`, `speckit`, `openspec` |
| Architecture | 2 | `architect`, `pattern`, `design` |
| Testing Conventions | 2 | `test` |
| Git & Workflow | 2 | `git`, `workflow`, `branch` |
| Behavioral Constraints | 2 | `constraint`, `do not`, `rule`, `behavioral` |

Record which sections are found and which are missing.

#### 4b: Quality Metrics

1. **Line count**: Count total lines. Flag if >300.
2. **Build code blocks**: Check if the Build section contains
   at least one fenced code block (triple backtick). Flag if
   the section exists but has no code blocks.
3. **Staleness check**: If a Project Structure section exists and
   contains a directory tree (lines starting with `├`, `└`, `│`,
   or indented paths ending with `/`), verify that the listed
   directories actually exist on the filesystem. Flag any that
   do not exist.
4. **Constitution reference** (only when
   `.specify/memory/constitution.md` exists): Check if AGENTS.md
   mentions the constitution. Flag if absent.
5. **Spec framework reference** (only when `specs/` has numbered
   subdirs or `openspec/config.yaml` exists): Check if AGENTS.md
   describes the spec framework. Flag if absent.
6. **Cross-framework bridge** (only when BOTH `specs/` and
   `openspec/` exist): Check if AGENTS.md explicitly states the
   constitution governs both frameworks. Look for co-occurrence
   of "constitution" with "both"/"all"/"regardless"/"openspec"
   within the spec framework or constitution section.

#### 4c: Scoring

Calculate the overall effectiveness label:

| Label | Criteria |
|-------|----------|
| Excellent | 5/5 Tier 1 + 4/4 Tier 2 + all applicable Tier 1C |
| Strong | 5/5 Tier 1 + 2-3/4 Tier 2 |
| Adequate | 4-5/5 Tier 1 |
| Weak | 2-3/5 Tier 1 |
| Missing | 0-1/5 Tier 1 |

If Tier 1C sections are applicable (triggers detected) but
missing, downgrade the label by one level.

#### 4d: Generate Report

Produce a structured report:

```
## /agent-brief: Audit Report

### Section Coverage

| Section | Tier | Status | Notes |
|---------|:----:|:------:|-------|
| Project Overview | 1 | ✅ | [notes] |
| Build & Test | 1 | ✅ | [notes] |
| ... | | | |

### Quality Metrics

| Metric | Value | Status |
|--------|-------|--------|
| Total lines | N | ✅/⚠ |
| Tier 1 sections | N/5 | ✅/❌ |
| Tier 2 sections | N/4 | ✅/⚠ |
| Build code blocks | N | ✅/⚠ |
| Bridge: CLAUDE.md | present/missing | ✅/⚠ |
| Bridge: .cursorrules | present/missing | ✅/⚠ |

### Improvement Suggestions

[numbered list of specific, actionable suggestions with
generated content for missing sections]

### Overall Score: [Label] ([N/N] essential, [N/N] recommended)
```

#### 4e: Offer Improvements

If improvements were suggested:
1. Ask: "Would you like me to apply these improvements?"
2. On confirmation, apply the changes:
   - Insert missing sections at appropriate locations
   - Update stale directory references
   - Do NOT modify existing project-specific sections
3. After applying, re-run the audit to show updated score.

### Step 5: Bridge File Verification

After creating or improving AGENTS.md, verify cross-tool bridge
files exist. Bridge file creation is owned by `uf init`
(`ensureCLAUDEmd()` and `ensureCursorrules()`). This command
only checks their status and suggests running `uf init` if
they are missing or misconfigured.

**CLAUDE.md**:
1. Check if CLAUDE.md exists at repo root.
2. If it exists, check if it contains `@AGENTS.md`.
3. If missing or lacking the import:
   - Report: `"⚠ CLAUDE.md: missing or does not import AGENTS.md"`
   - Suggest: `"Run: uf init to create bridge files"`
4. If already configured:
   - Report: `"⊘ CLAUDE.md: already imports AGENTS.md"`

**.cursorrules**:
1. Check if .cursorrules exists at repo root.
2. If it exists, check if it references AGENTS.md.
3. If missing or lacking the reference:
   - Report: `"⚠ .cursorrules: missing or does not reference AGENTS.md"`
   - Suggest: `"Run: uf init to create bridge files"`
4. If already configured:
   - Report: `"⊘ .cursorrules: already references AGENTS.md"`

**Note**: `uf init` is the canonical owner of bridge file
creation. It generates CLAUDE.md with `@AGENTS.md` plus
convention pack `@` imports, and .cursorrules with AGENTS.md
reading instructions. Do NOT create bridge files with a
different marker -- defer to `uf init`.

### Step 6: Summary Report

Display a final summary:

**Create mode**:
```
## /agent-brief: Complete

### Created
  ✅ AGENTS.md: generated (N lines)
  [bridge file statuses]

### Next Steps
  Fill in the Tier 2 TODO sections (Architecture, Testing,
  Git & Workflow, Behavioral Constraints).
  Then run `uf init` to deploy convention packs and agents.
```

**Audit mode**:
```
## /agent-brief: Audit Complete

### Score: [Label]
  [section coverage summary]
  [quality metrics summary]
  [improvements applied or suggested]
```

## Guardrails

- **NEVER modify files outside AGENTS.md, CLAUDE.md, and
  .cursorrules** -- this command manages agent context files
  only.
- **NEVER implement code, modify source files, update tests,
  or change configuration** -- this command produces
  documentation artifacts.
- **ALWAYS present generated content to the user before
  writing** -- never auto-write without confirmation.
- **ALWAYS respect existing project-specific sections** -- when
  improving, insert missing sections but do not rewrite or
  remove sections the user has customized.
- **NEVER remove content** -- only add missing sections or
  update stale references. If a section should be condensed,
  suggest it but do not apply without confirmation.
- **Use actual project data** -- in create mode, fill Tier 1
  sections from real files (README, Makefile, go.mod, CI
  config). Do not use placeholder text or generic examples.
