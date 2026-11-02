---
tag: permission-gating
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:09:54Z
identity: permission-gating-20261102T050954-jay-flowers
tier: draft
---

Glob-based bash permission rules that gate short CLI flags must use the glued pattern form (`-f*`), not the spaced form (`-f *`). A spaced pattern requires the literal ` -f ` substring and silently fails to match glued usage like `-ftitle=x`; a glued pattern's trailing `*` absorbs the space and covers both forms in a single rule. In unbound-force's gh api gating (issue #569), spaced-only `-f`/`-F` rules let glued-form field flags fall through to the `"gh api*": "allow"` rule — and gh auto-switches to POST when request fields are present, so those are unguarded mutations. The refine-bash-permissions design covered glued forms for `-X` method flags (8 spaced + 8 glued rules) but omitted them for data-sending flags; spec FR-002 carries glued `-XPOST` scenarios while FR-003 has no glued short-flag scenario. Triage corroboration tip: when the target file is machine-local/gitignored in the reviewing workspace (opencode.json is gitignored in fork workspaces), verify quoted config against the canonical spec/design artifacts and CHANGELOG rather than the local file.
