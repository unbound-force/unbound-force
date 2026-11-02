---
tag: bash-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:09:49Z
identity: bash-permissions-20261102T050949-jay-flowers
tier: draft
---

Glob-based bash permission patterns that gate CLI mutation flags must not assume a space between a shorthand flag and its value. cobra/pflag CLIs — including gh — accept glued shorthand values, so `-ftitle=x` is equivalent to `-f title=x`; a pattern requiring the spaced literal ` -f ` silently misses the glued form, and under last-match-wins ordering the command falls through to the allow default and executes without prompting. Verified while triaging unbound-force issue #569: the refine-bash-permissions rule set covered glued method flags (`-XPOST*`) but gated the `-f`/`-F` data flags only in spaced form. Write flag-gate patterns as the flag prefix followed immediately by the wildcard (`-f*`, `-F*`, no space) so spaced, glued, and glued-with-equals forms all match.
