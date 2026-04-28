## Design Decisions

### D1. Lightweight single-agent, not council delegation

**Decision**: `/review-pr` stays as a single-agent command.
It does NOT delegate to Divisor persona agents even when
they are present.

**Rationale**: `/review-council` already provides the
multi-agent council experience for local tree review.
Adding Divisor delegation to `/review-pr` would duplicate
that capability and make the command significantly more
token-expensive. The two commands serve different workflow
moments:

- `/review-council`: pre-PR, deep multi-agent analysis
- `/review-pr`: post-PR, lean single-agent review with
  CI awareness

Keeping `/review-pr` lightweight makes it practical for
reviewing other people's PRs where you want a quick
assessment, not a full council session.

### D2. Convention packs as additive context

**Decision**: When `.opencode/uf/packs/` exists, the
command loads applicable packs and uses their numbered
rules (CS-001, AP-001, SC-001, etc.) alongside the
constitution for more specific findings. When packs are
absent, the command relies on the constitution alone.

**Rationale**: Convention packs provide precise, numbered
rules that make findings actionable ("violates CS-004: DRY
principle" vs "code is duplicated"). But the command must
work in repos that don't use UF's full scaffold, so packs
are additive, not required.

**Detection logic**:
1. Check if `.opencode/uf/packs/` directory exists
2. Always load `default.md` (language-agnostic rules)
3. Detect language: `go.mod` → `go.md`,
   `tsconfig.json` or `package.json` → `typescript.md`
4. Load corresponding `-custom.md` if present
5. Also load `severity.md` if present
6. Exclude `content.md` and `content-custom.md` — these
   contain writing standards (voice, brand, blog format)
   for the Scribe/Herald/Envoy agents, not code quality
   rules relevant to PR review

### D3. Generic constitution reference

**Decision**: The command reads `.specify/memory/constitution.md`
and evaluates against whatever principles it finds, rather
than hardcoding specific principle names or numbers.

**Rationale**: org-infra's constitution has 7 principles
(I-VII), UF's has 4 (I-IV), and other repos may have
different sets. Hardcoding principle references makes the
command brittle and repo-specific. The standardized version
instructs the AI to "extract all principles and their
MUST/SHOULD rules" generically.

**Fallback**: If no constitution file exists, the command
notes this and reviews against general software engineering
best practices (which the AI model already knows).

### D4. PR number auto-detection

**Decision**: The PR number argument is optional. When
omitted, the command detects the current branch's open PR
via `gh pr view --json number`.

**Rationale**: The most common use case is reviewing your
own PR on the current branch. Requiring the PR number is
unnecessary friction in that case. Explicit numbers are
still supported for reviewing other people's PRs.

**Error handling**: If no open PR exists for the current
branch, the command errors with:
```
No open PR found for branch '<branch>'.
Provide a PR number: /review-pr 42
```

### D5. Severity definitions: pack-first, inline fallback

**Decision**: The command checks for
`.opencode/uf/packs/severity.md`. If found, severity
definitions come from the pack. If not found, inline
definitions are used as fallback.

**Rationale**: The severity pack provides per-persona
examples and an auto-fix policy that ensure consistency
across the UF ecosystem. But repos without UF packs still
need severity definitions, so the inline fallback is
preserved.

### D6. Tool-owned file ownership

**Decision**: The command is deployed as a tool-owned file
under `.opencode/command/review-pr.md`.

**Rationale**: All files under `opencode/command/` are
tool-owned per the scaffold engine's `isToolOwned()`
function. This means `uf init` re-runs will automatically
update the command when improvements are made upstream.
This is the desired behavior — the whole point of
standardizing the command is that improvements flow to all
repos.

### D7. Preserved operational features

The following features from org-infra's original are
preserved without modification:

| Feature | Why preserved |
|---------|--------------|
| CI causality analysis (2a) | Distinguishes PR-caused from pre-existing failures — unique value not found in /review-council |
| Local tool detection (3) | Already repo-agnostic — detects Makefile, golangci-lint, ruff, yamllint, pre-commit |
| Scoped diff fetching (4) | Token-efficient — staged loading, skip binaries/lock files |
| Spec awareness (5) | Already checks both `specs/` and `openspec/` |
| Fix-branch offer (8) | Solid guardrails: local only, never pushes, non-trivial fixes are deferred to human |
| In-line PR comments (9) | Human confirmation gate: shows all comments, waits for explicit yes before posting |

## Architecture

No new Go packages or functions are introduced. The change
adds a single Markdown file to the embedded scaffold assets
and updates existing test assertions.

```
internal/scaffold/assets/opencode/command/
  ├── review-pr.md    ← NEW (tool-owned, ~340 lines)
  └── ... (7 existing commands unchanged)
```

The scaffold engine's existing machinery handles everything:
- `embed.FS` embeds the file at compile time
- `mapAssetPath()` maps `opencode/command/review-pr.md`
  to `.opencode/command/review-pr.md`
- `isToolOwned()` returns true for the `opencode/command/`
  prefix
- `insertMarkerAfterFrontmatter()` adds the version marker
- Drift detection test validates source ↔ asset parity
