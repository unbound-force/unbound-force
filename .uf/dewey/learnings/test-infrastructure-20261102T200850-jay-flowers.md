---
tag: test-infrastructure
author: jay-flowers
category: pattern
created_at: 2026-11-02T20:08:50Z
identity: test-infrastructure-20261102T200850-jay-flowers
tier: draft
---

Safety properties implemented as error-shape conversions in production adapters are not covered by faked-dependency unit tests — the fake bypasses exactly the conversion logic the property depends on. In the preflight-project-checks review, uf_preflight_declared_checks' runProcessNoShell maps execFile's killed/signal error to timed_out + exit 124, but every timeout test preset ProcessResult.timed_out, so the real conversion had zero coverage (found 2026-11). Pair each documented safety property (timeout cap, output bounding, shell-less exec) with a real-process round-trip case: for timeouts, a sleep script with timeout_seconds: 1 asserting timed_out, exit code 124, and the rendered annotation.
