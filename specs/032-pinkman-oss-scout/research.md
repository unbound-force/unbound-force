# Research: Pinkman OSS Scout

## R1: Implementation Approach -- Agent-Only vs Go CLI

**Decision**: Implement Pinkman as an OpenCode agent file
(`.opencode/agents/pinkman.md`) with a `/scout` slash
command (`.opencode/command/scout.md`). No Go CLI backend
package.

**Rationale**: Pinkman's core operations are AI reasoning
tasks -- interpreting web content, classifying licenses,
assessing trend signals, generating natural-language
reports. These map naturally to the agent runtime's
capabilities. The `webfetch` tool provides access to
public data sources, and `read`/`write` handle file I/O
for manifest parsing and report persistence. This keeps
the implementation to 2 Markdown files (plus scaffold
copies and test updates), which is dramatically simpler
than a Go CLI backend (which would require a new
`internal/pinkman/` package, Cobra commands, etc.).

Prior art: The onboarding agent (Spec 031) follows this
exact pattern -- agent file + command, no Go backend.
The Divisor Scribe, Herald, and Envoy agents (Spec 026)
are also agent-only implementations.

**Alternatives considered**:
- Go CLI backend (like Muti-Mind or Mx F): Adds 500+
  lines of Go code, a new `cmd/pinkman/` entry point,
  GoReleaser multi-binary config, and Homebrew formula.
  Unjustified complexity for operations that are better
  expressed as agent instructions.
- Standalone binary in a separate repo: Would require
  full hero status (Spec 002 compliance), its own
  constitution, and CI pipeline. Overkill for a utility
  agent.
- MCP server: Pinkman does not serve other tools; it is
  invoked by users. An MCP server would be the wrong
  interaction model.

## R2: OSI License List Retrieval

**Decision**: Retrieve the OSI-approved license list from
`https://opensource.org/licenses/` via `webfetch` at
invocation time. Maintain a hardcoded fallback set of
well-known OSI-approved licenses for use when the OSI
site is unreachable.

**Rationale**: The spec (FR-003) mandates the OSI-approved
list as the "sole authority." Live retrieval ensures the
agent always uses the current list. The OSI publishes the
list at a stable URL. The fallback set prevents total
failure when the site is down -- the agent reports "using
fallback license list, live OSI verification unavailable"
in results.

The fallback set includes licenses that have been on the
OSI list for 10+ years and are overwhelmingly common:
MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC,
MPL-2.0, LGPL-2.1, LGPL-3.0, GPL-2.0, GPL-3.0,
AGPL-3.0, Unlicense, Artistic-2.0, EPL-2.0, EUPL-1.2,
0BSD, Zlib, BSL-1.0. All of these are OSI-approved.

**Alternatives considered**:
- Static list only (embedded in agent file): Would
  become stale as OSI approves new licenses. Violates
  the "sole authority" requirement.
- SPDX license list instead of OSI: SPDX includes
  non-OSI-approved licenses. The spec explicitly requires
  OSI approval, not just SPDX presence.
- License detection library: Would require Go code.
  Agent-based classification using webfetch is sufficient
  for project-level license detection (root LICENSE
  file).

## R3: Data Sources for Project Discovery

**Decision**: Use GitHub's public web interface as the
primary discovery source. Use package registries
(pkg.go.dev, npmjs.com, crates.io) as secondary sources
for dependency resolution and version information.

**Rationale**: GitHub hosts the vast majority of open
source projects and provides publicly accessible
metadata: stars, forks, contributor count, release dates,
license files, dependency manifests (go.mod,
package.json, Cargo.toml). The `webfetch` tool can
retrieve this data from public URLs without API keys.

For dependency resolution, package registries provide
authoritative version and dependency information that
is more reliable than parsing raw manifest files from
repository HTML.

Rate limiting: GitHub's public web interface has no
formal rate limit, but aggressive scraping may trigger
anti-bot measures. The agent handles this gracefully per
FR-012 (report partial results).

**Alternatives considered**:
- GitHub REST API: Requires a personal access token for
  reasonable rate limits (60 req/hr unauthenticated vs.
  5000 req/hr authenticated). The spec says "no
  proprietary or paid data sources required." PATs are
  free but add setup friction.
- GitHub GraphQL API: Same PAT requirement. Richer
  queries but heavier setup.
- Multiple search engines (Google, Bing): Unreliable for
  structured project metadata. Better suited for broad
  discovery but not for extracting stars, forks, license
  files.
- Libraries.io: Excellent dependency data but requires
  API key. Could be added as an optional enhancement.

## R4: Dependency Manifest Parsing

**Decision**: Parse dependency manifests by reading the
raw file content and extracting dependency names and
versions using pattern matching. Support `go.mod`
(primary, since Unbound Force is a Go project),
`package.json`, `Cargo.toml`, `requirements.txt`, and
`pyproject.toml`.

**Rationale**: Manifest files follow well-documented
formats with consistent structure. The agent can extract
dependency names and version constraints using its
language understanding capabilities without a dedicated
parser. For `go.mod`, the `require` block lists
dependencies one per line in `module/path vX.Y.Z` format.

For overlap detection, the agent compares dependency
names across all scouted projects in the result set.
Version conflicts are detected when the same dependency
appears at different versions.

**Alternatives considered**:
- Shell out to `go list -m all`: Requires Go toolchain
  and the target project's source code. Pinkman operates
  on remote projects, not local ones (except for the
  audit mode).
- Dedicated Go parser package: Requires Go code in a
  new internal package. Overkill for agent-based parsing
  of well-structured text files.

## R5: Report Format and Storage

**Decision**: Store scouting reports as Markdown files
with YAML frontmatter at `.uf/pinkman/reports/`. File
naming: `YYYY-MM-DDTHH-MM-SS-<sanitized-query>.md`.
Optionally store in Dewey via `dewey_store_learning`.

**Rationale**: Markdown with YAML frontmatter is the
established format for persistent data in the Unbound
Force ecosystem (Spec 004 backlog items, Spec 007
impediments, Spec 031 profiles). YAML frontmatter
provides machine-parseable metadata (producer, version,
timestamp, query, result count). The Markdown body
contains the human-readable report.

The `.uf/pinkman/` directory follows Spec 025's
convention for per-project runtime data. Reports are
git-ignored (runtime data, not source artifacts).

Dewey integration uses `dewey_store_learning` with tag
`pinkman` for semantic search across past scouting
results. This enables the agent to answer "have we
evaluated this project before?" without re-scouting.

**Alternatives considered**:
- JSON files: Less human-readable for reports that
  include narrative assessment text.
- Standard artifact envelope (JSON): The envelope format
  (Spec 009) is designed for inter-hero artifacts.
  Scouting reports are primarily for human consumption,
  not hero consumption. If inter-hero consumption is
  needed later, the report can be wrapped in an envelope.
- Database (SQLite): Overkill for report storage.
  Introduces complexity without significant benefit over
  file-based storage.

## R6: Scaffold Integration

**Decision**: Add two new embedded assets to the scaffold
engine:
1. `opencode/agents/pinkman.md` — user-owned
2. `opencode/command/scout.md` — tool-owned

Update `expectedAssetPaths` in tests (35 → 37).
No new `initSubTools()` delegation needed -- files are
deployed by the standard asset walk.

**Rationale**: This follows the exact pattern used by
every prior agent and command addition (Specs 005, 006,
007, 018, 019, 026, 031). The scaffold engine's
`fs.WalkDir` automatically picks up new files under
`internal/scaffold/assets/`. Tool ownership for the
command file ensures the invocation interface stays
canonical. User ownership for the agent file allows
scouting behavior customization.

**Alternatives considered**:
- External tool delegation (like `specify init`): Pinkman
  is part of the core ecosystem, not a third-party tool.
  Embedding is correct.
- No scaffold integration: Users would have to manually
  create the agent file. Violates zero-friction principle.

## R7: Trend Signal Quantification

**Decision**: Use three primary trend indicators derived
from publicly available GitHub data:
1. **Star growth rate**: Stars gained in the last 90 days
   as a percentage of total stars
2. **Release velocity**: Number of releases in the last
   6 months
3. **Contributor activity**: Number of unique contributors
   with commits in the last 90 days

Secondary indicators (reported when available):
- Fork count trajectory
- Issue response time (median time to first response)
- Dependency adoption count (how many other projects
  depend on it)

**Rationale**: These three primary indicators are
available from GitHub's public web interface and provide
a balanced view of community traction (stars), project
maturity (releases), and active development
(contributors). The 90-day window captures recent
momentum without being noisy from daily fluctuations.

**Alternatives considered**:
- Social media mentions: Requires sentiment analysis,
  which the spec explicitly excludes.
- Download counts: Not consistently available across
  package registries.
- Google Trends data: Too broad -- measures general
  interest, not developer adoption.
