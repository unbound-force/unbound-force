---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-07T17:22:15Z
identity: ci-cd-20261107T172215-jay-flowers
tier: draft
---

Triage of toolchain-pinning feedback on CI vulnerability-scanner jobs should separate build-vehicle reproducibility from scan-outcome reproducibility. A scanner querying a live advisory database (osv-scanner, Dependabot-style scans) is non-reproducible by design — advisories appear between runs — so pinning the Go/npm toolchain that builds the scanner does not make the job's outcome reproducible; the determinism levers that matter are the scanner version pin and the scanned input. In Go, GOTOOLCHAIN=auto under a version-pinned `go install module@vX.Y.Z` derives its toolchain requirement from the pinned module's go directive, so the selection is bounded by the reviewed pin and cannot drift independently of it; an explicit GOTOOLCHAIN pin instead adds a second coupled value to co-bump with the scanner pin. When downloads are checksum-verified and the trade-off is documented at the decision point (inline rationale comment plus change-level risk register), AUTHOR-DECIDES is the proportionate resolution for such LOW feedback.
