## Context

Pinkman (Spec 032) classifies licenses as OSI-approved
or not, but does not distinguish permissive from copyleft
licenses. The Unbound Force ecosystem uses Apache-2.0
(Spec 002, section 2.2). Adding a copyleft dependency
(GPL-3.0, AGPL-3.0) would impose derivative work
obligations incompatible with Apache-2.0. Pinkman
currently gives `adopt` to any project with an
OSI-approved license and healthy signals, creating a
real risk of recommending incompatible dependencies.

The proposal (constitution alignment: all PASS) adds a
compatibility tier layered on the existing OSI check.
No hero agents, schema registry, or Go code are
modified.

## Goals / Non-Goals

### Goals
- Prevent Pinkman from recommending `adopt` for
  dependencies whose licenses conflict with Apache-2.0.
- Classify every detected license into a compatibility
  tier (permissive, weak-copyleft, strong-copyleft)
  based on derivative work obligations.
- Produce a per-project compatibility verdict
  (compatible, caution, incompatible) that factors into
  the recommendation verdict.
- Display the compatibility tier and verdict in all
  output formats and Dewey learnings.

### Non-Goals
- Legal advice. Pinkman's classification is a screening
  tool, not a legal opinion. The `caution` verdict
  explicitly directs users to seek legal review.
- License compatibility analysis between dependencies
  (inter-dependency license conflicts). Only
  project-to-ecosystem compatibility is assessed.
- Configurable project license. The reference license
  is hardcoded as Apache-2.0 per Spec 002.
- SPDX `AND` expressions. Conjunctive licenses (both
  apply simultaneously) fall through to `unknown` /
  `caution` as a safe default requiring human review.
- SPDX `WITH` exceptions. License exceptions (e.g.,
  Classpath exception on GPL) fall through to `unknown`
  / `caution`. Mapping individual exceptions to tier
  adjustments is too complex for v1.
- Modifying any hero agent file, schema registry entry,
  or Go source code.

## Decisions

### D1: Three-tier classification

| Tier | Licenses | Derivative work obligation |
|------|----------|---------------------------|
| `permissive` | MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, Unlicense, 0BSD, Zlib, BSL-1.0 (Boost Software License — not to be confused with BUSL-1.1 Business Source License) | None or minimal — attribution only |
| `weak-copyleft` | LGPL-2.1-only, LGPL-2.1-or-later, LGPL-3.0-only, LGPL-3.0-or-later, MPL-2.0, EPL-2.0, EUPL-1.2, Artistic-2.0 | File-level or linking-exception — modifications to the library must be shared, but the consuming project is not a derivative work if linked as a library |
| `strong-copyleft` | GPL-2.0-only, GPL-2.0-or-later, GPL-3.0-only, GPL-3.0-or-later, AGPL-3.0-only, AGPL-3.0-or-later | Full — any derivative work must be distributed under the same license |

**Rationale**: Three tiers map cleanly to three
risk levels for an Apache-2.0 project. Permissive
licenses have no conflict. Strong-copyleft licenses
are definitively incompatible (Apache-2.0 cannot
satisfy GPL's copyleft requirements). Weak-copyleft
licenses are situational — they may be compatible
depending on usage (linking vs modification), requiring
human judgment.

Licenses not in the tier table (unrecognized SPDX
identifiers) receive the tier `unknown` and are
treated as `caution` for the compatibility verdict.

### D2: Compatibility verdict mapping

| Tier | Verdict | Rationale |
|------|---------|-----------|
| `permissive` | `compatible` | No derivative work obligation conflicts with Apache-2.0 |
| `weak-copyleft` | `caution` | May be compatible depending on usage pattern (linking vs modification). Requires human legal review |
| `strong-copyleft` | `incompatible` | Derivative work obligations cannot be satisfied under Apache-2.0 |
| `unknown` | `caution` | Unclassified license — requires human review |
| `not_approved` (non-OSI) | `incompatible` | Existing behavior preserved — non-OSI licenses are already excluded from `adopt` |
| `manual_review` | `caution` | Non-standard license text — requires human review |

**Rationale**: The verdicts mirror the existing
OSI classification pattern (binary with an escape
hatch) but add a middle ground (`caution`) that
acknowledges the weak-copyleft gray area. The verdicts
are conservative — `caution` means "don't auto-adopt,
get human input."

### D3: Recommendation verdict gate

The compatibility verdict acts as a hard gate on the
recommendation verdict:

| Compatibility | Allowed recommendations |
|---------------|----------------------|
| `compatible` | adopt, evaluate, defer, avoid |
| `caution` | evaluate, defer, avoid |
| `incompatible` | avoid only |

**Rationale**: A `compatible` license does not
guarantee `adopt` — other factors (maintenance health,
trend trajectory) still apply. But an `incompatible`
license overrides all positive signals — no amount of
stars or contributor activity makes a GPL-3.0
dependency safe for an Apache-2.0 project. The
`caution` tier caps at `evaluate` to flag the need
for human review without outright rejecting the
project.

### D4: Dual-license compatibility

For dual-licensed projects (SPDX `OR` expression),
evaluate each option independently and use the most
favorable compatible tier:

- `MIT OR GPL-3.0-only` → permissive (MIT is permissive)
- `LGPL-3.0-only OR GPL-3.0-only` → weak-copyleft
  (LGPL is more favorable)
- `GPL-3.0-only OR AGPL-3.0-only` → strong-copyleft
  (both are strong-copyleft)

**Rationale**: Dual-license models exist specifically
to offer flexibility. Evaluating only the most
restrictive option would misrepresent the project's
actual compatibility.

### D5: Fallback license list annotations

The hardcoded fallback license list gains tier
annotations. Each license is listed with its tier
so that compatibility verdicts can be produced even
when the OSI website is unreachable:

```
MIT (permissive), Apache-2.0 (permissive),
BSD-2-Clause (permissive), ...
GPL-3.0-only (strong-copyleft), ...
LGPL-3.0-only (weak-copyleft), ...
```

**Rationale**: The fallback list is the only data
source when the OSI website is unavailable.
Compatibility tiering must work in fallback mode.

### D6: Reference license source

The reference license (Apache-2.0) is hardcoded in
the agent file, not read from `LICENSE` at runtime.

**Rationale**: Pinkman has `bash: false` and cannot
execute shell commands to detect the project license.
It could use the `read` tool to parse `LICENSE`, but
Spec 002 already establishes Apache-2.0 as the
recommended license for all hero repositories. A
hardcoded reference is simpler, deterministic, and
avoids edge cases where `LICENSE` is missing or
contains non-standard text. If the project license
changes from Apache-2.0, update the reference in the
License Compatibility section and re-evaluate all
tier-to-verdict mappings.

## Risks / Trade-offs

### R1: Oversimplification of weak-copyleft (accepted)

Weak-copyleft licenses (LGPL, MPL-2.0) have nuanced
compatibility depending on usage patterns (static vs
dynamic linking, file-level copyleft boundary). The
blanket `caution` verdict does not distinguish these
subtleties. This is accepted because Pinkman is a
screening tool, not a legal advisor — `caution` means
"get human input," which is the correct response to
nuanced legal questions.

### R2: Static tier mapping (accepted)

The tier classification is a static mapping embedded
in the agent file. If new licenses are approved by OSI
or existing licenses change their terms, the mapping
must be manually updated. This is low risk because
license classifications change rarely (years, not
months) and the live OSI fetch + `unknown` tier
provide a safety net for unrecognized licenses.

### R3: No inter-dependency license analysis

This change only assesses project-to-ecosystem
compatibility (project license vs Apache-2.0). It does
not check whether a project's own dependencies have
license conflicts with each other. This is a deliberate
non-goal to keep scope manageable — inter-dependency
license analysis is a substantially more complex
feature.
