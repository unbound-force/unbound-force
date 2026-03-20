# Research: The Divisor Architecture

**Spec**: 005-the-divisor-architecture
**Date**: 2026-03-19

## R1: Convention Pack Format

**Decision**: Markdown with YAML frontmatter.

**Rationale**: The primary consumers of convention packs
are AI agent personas (LLMs), which parse natural language
prose more reliably than structured data. Markdown allows
rich descriptions with examples and rationale — critical
for a reviewer that needs to understand *why* a rule
exists. The agents already consume Markdown files
(AGENTS.md, constitution, spec.md). YAML frontmatter
provides machine-parseable metadata (pack_id, language,
version) for tooling.

**Alternatives considered**:
- Pure YAML: More machine-parseable but harder for AI
  agents to interpret rule nuance. Would require a
  runtime parser in the agent tool chain.
- JSON: Too verbose for human-authored convention rules.
  Not aligned with existing ecosystem (all agent/command
  files are Markdown).

## R2: Convention Pack Sections

**Decision**: Six required H2 sections matching FR-007,
plus YAML frontmatter for metadata.

**Format**:
```text
---
pack_id: {language}
language: {Language Name}
version: 1.0.0
---
# Convention Pack: {Language Name}
## Coding Style
## Architectural Patterns
## Security Checks
## Testing Conventions
## Documentation Requirements
## Custom Rules
```

**Rationale**: Each section maps 1:1 to the FR-007
requirement. H2 headers make sections scannable by both
humans and AI agents. H3 subsections within a section
provide topical grouping (e.g., "Formatting", "Naming"
under "Coding Style").

## R3: Rule Identifiers

**Decision**: Each rule gets a stable prefixed numeric
identifier: CS-NNN (Coding Style), AP-NNN (Architectural
Patterns), SC-NNN (Security Checks), TC-NNN (Testing
Conventions), DR-NNN (Documentation Requirements),
CR-NNN (Custom Rules).

**Rationale**: Enables traceability in review findings
(a finding can cite "CS-006" instead of quoting the full
rule text). Supports Mx F trend analysis across reviews
(Spec 007). Supports the Swarm learning loop (Spec 008).

**Alternatives considered**:
- Unnumbered bullet points: Simpler but prevents
  machine-traceable references. Trend analysis would
  require fuzzy text matching.

## R4: Severity Indicators

**Decision**: RFC 2119 keywords `[MUST]`, `[SHOULD]`,
`[MAY]` inline with each rule.

**Rationale**: Consistent with the vocabulary used
throughout the project (constitution, specs, AGENTS.md).
Maps to finding severity: `[MUST]` violation -> MAJOR or
CRITICAL finding, `[SHOULD]` -> MINOR, `[MAY]` -> INFO.
Exact severity is left to the persona's judgment because
context matters.

## R5: Convention Pack Ownership (Split Model)

**Decision**: Split each convention pack into two files:
- `{lang}.md` — **tool-owned**, auto-updated by
  `unbound init` when content changes
- `{lang}-custom.md` — **user-owned**, never overwritten,
  contains project-specific `custom_rules[]` extensions

**Rationale**: Solves the conflict between automatic pack
updates (new language idioms) and user customization
(project-specific rules). The scaffold engine operates at
file granularity (full file write, `bytes.Equal` for diff
detection). Section-level merge within a single file would
require a content parser, violating the existing clean
pattern. Two files keep the architecture simple.

**Alternatives considered**:
- Single file, user-owned: Simpler but fails Automated
  Governance — convention pack improvements would not
  reach users automatically.
- Single file, tool-owned: Overwrites user customizations.
- Single file with protected section marker: Requires
  section-level parser, fragile and complex.

## R6: Agent Template Design — Pack Loading

**Decision**: All `divisor-*.md` agents include a
standardized `### Convention Pack` subsection under
`## Source Documents` that instructs the agent to read
all `*.md` files from `.opencode/unbound/packs/`.

**Rationale**: Reading "all `*.md` files" (rather than a
single named file) allows projects to have both a
language pack and a custom pack loaded simultaneously.
The `[PACK]` tag on checklist headings signals
pack-dependent sections. Agents skip these sections with
a note when no pack is loaded — graceful degradation per
spec US2 acceptance scenario 3.

**Alternatives considered**:
- Hardcoded single file path: Simpler but prevents
  the split ownership model (R5) and multi-pack
  scenarios.
- Agent-specific pack sections: Each persona reads
  only its relevant sections. Rejected because all
  personas benefit from reading the full pack for
  context, even if they prioritize certain sections.

## R7: Persona-to-Pack-Section Mapping

**Decision**: Each persona prioritizes specific convention
pack sections but may reference any section for context.

| Persona | Priority Pack Sections |
|---------|----------------------|
| Guard | `custom_rules` |
| Architect | `coding_style`, `architectural_patterns`, `documentation_requirements` |
| Adversary | `security_checks`, `custom_rules` |
| SRE | `architectural_patterns`, `custom_rules` |
| Testing | `testing_conventions` |

**Rationale**: Prevents five personas all checking the
same thing. Each persona has a clear "lane" within the
pack. The mapping is advisory, not exclusive — personas
MAY reference any section when relevant to their finding.

## R8: Agent Content Classification

**Decision**: Agent content is classified into three
buckets during the `reviewer-*` to `divisor-*` migration:

1. **STAYS in agent**: Persona-specific identity, focus
   areas, universal checks, decision criteria, output
   format.
2. **MOVES to convention pack**: Language-specific rules
   (Go conventions, framework patterns, test framework
   requirements).
3. **BECOMES GENERIC + pack reference**: Checklist items
   that currently hardcode project-specific details but
   should become generic with a `[PACK]` tag for
   convention-pack-driven specifics.

**Rationale**: This classification ensures agents are
project-agnostic (work for any project) while convention
packs provide the language/project specificity. The
existing Go-specific content extracted from all five
reviewer agents forms the initial `go.md` convention pack.

## R9: Scaffold Engine — Subset Deployment

**Decision**: Add `DivisorOnly bool` to the `Options`
struct. Use a predicate function `isDivisorAsset(relPath)`
to filter files during `fs.WalkDir` when `DivisorOnly`
is true.

The predicate matches:
- `opencode/agents/divisor-*.md` (prefix match)
- `opencode/command/review-council.md` (exact match)
- `opencode/unbound/packs/*` (prefix match)

**Rationale**: A predicate function is more extensible
than an explicit list (adding a new `divisor-foo.md`
persona requires zero filter changes) and more precise
than a broad prefix match (which would miss agents in the
shared `opencode/agents/` directory). The `divisor-`
naming convention established in the spec makes prefix
matching reliable.

**Alternatives considered**:
- Explicit file list: Requires updating the list every
  time a persona is added, breaking dynamic discovery.
- `opencode/divisor/` prefix only: Would miss agents
  and commands in shared directories.

## R10: Scaffold Engine — Language Detection

**Decision**: Add `Lang string` to the `Options` struct.
Implement `detectLang(targetDir)` that checks for marker
files in priority order:

1. `go.mod` → "go"
2. `tsconfig.json` → "typescript"
3. `package.json` → "typescript"
4. `pyproject.toml` → "python"
5. `Cargo.toml` → "rust"

`--lang` flag takes priority over auto-detection. If
neither provides a language, fall back to "default".

**Rationale**: Marker files are standard, reliable
indicators of project language. Priority order resolves
ambiguity (a Go project with a `package.json` for tooling
should get the Go pack). The function is pure and easily
testable with `t.TempDir()`.

## R11: Convention Pack Filtering

**Decision**: Add `shouldDeployPack(relPath, lang)` filter.
Only deploy packs matching the resolved language plus
`default` (always deployed as fallback). Filter applies
in both full and `--divisor` modes.

The function matches: `{lang}.md`, `{lang}-custom.md`,
`default.md`, `default-custom.md`. All other packs are
skipped.

**Rationale**: A Go project should not receive
`typescript.md`. This aligns with the Zero-Waste Mandate
from the constitution.

## R12: `review-council.md` Discovery Pattern Change

**Decision**: Change the discovery prefix from
`reviewer-*.md` to `divisor-*.md`. Update the known
roles reference table to use `divisor-*` names.

**Rationale**: Minimal change. The command was already
designed with dynamic discovery. Only the prefix string
and role table entries change. The dynamic discovery
architecture is already correct.

## R13: Embedded Asset Count

**Decision**: Add 12 new embedded files, bringing the
total from 33 to 45:
- 5 `divisor-*.md` agents (user-owned)
- 1 `review-council.md` command (tool-owned, was non-embedded)
- 3 canonical convention packs (tool-owned)
- 3 custom convention pack stubs (user-owned)

Remove `review-council.md` from `knownNonEmbeddedFiles`.
Add `.opencode/divisor` to `canonicalDirs` in drift
detection tests.

## R14: Go Code Estimate

**Decision**: ~200-250 lines new Go code, ~50 lines
modified, ~300 lines new test code.

New functions: `isDivisorAsset()`, `isConventionPack()`,
`shouldDeployPack()`, `detectLang()`.

Modified functions: `Run()`, `isToolOwned()`,
`printSummary()`, `newInitCmd()`.

New tests: `TestIsDivisorAsset`, `TestDetectLang`,
`TestShouldDeployPack`, `TestRun_DivisorSubset`,
`TestRun_DivisorSubset_WithLangFlag`,
`TestRun_DivisorSubset_DefaultFallback`.

Risk: Low — all changes are additive. No existing behavior
changes when `DivisorOnly=false` and `Lang=""`.
<!-- scaffolded by unbound vdev -->
