---
tag: scaffold-sentinels
author: jay-flowers
category: gotcha
created_at: 2026-11-01T23:38:31Z
identity: scaffold-sentinels-20261101T233831-jay-flowers
tier: draft
---

Gating sub-tool delegation on a bare directory sentinel is fragile when any earlier phase of the same command creates that directory. In `uf init`, the embedded starter-constitution asset walk created `.specify/` before `initSubTools()` ran, so the bare `.specify` sentinel was satisfied on every fresh run and `specify init` never executed — an invisible skip with no summary line. The fix (unbound-force fix-init-speckit-sentinel, issue #827's originating change) probes an install-specific artifact (`.specify/scripts`) instead of the directory, passes `--force` for specify 1.1.0 merge semantics (which preserve user-owned files such as the starter constitution), and renders sentinel skips as visible "already configured" sub-tool summary lines so users and documentation can distinguish "skipped: binary missing" from "skipped: already installed". General rule: sentinels should detect an install-specific artifact, not directory existence, and skips should always be visible in command output.
