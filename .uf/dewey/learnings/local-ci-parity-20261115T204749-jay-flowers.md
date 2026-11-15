---
tag: local-ci-parity
author: jay-flowers
category: pattern
created_at: 2026-11-15T20:47:49Z
identity: local-ci-parity-20261115T204749-jay-flowers
tier: draft
---

Full local CI parity for unbound-force pre-flight gates (verified 2026-11-15 on branch quick/extend-heredoc-ban-contract-test-scan-to): `make check` covers go vet plus golangci-lint, go test -race, go build, the scope-based coverage gate (whose coverage prerequisite also produces coverage.out), and the TypeScript plugin suite — but ci_local.yml additionally runs a global 80% coverage ratchet (`go tool cover -func=coverage.out` total) and `bash council-review-action/test/test-pipeline.sh` (91 assertions) as separate jobs, so run both explicitly after make check. ci_crapload.yml consumes its analysis via a `uses:` reusable workflow (org-infra, not locally executable), ci_eval.yml needs ENMAAS_KEY, and the security/scheduled workflows are CI-runtime-context commands — none belong in the local pre-flight set. When make check stops early at a classified pre-existing lint failure, run the remaining aggregate targets directly (`make test build coverage-gate plugin-test`) so every check still executes, and use the produced coverage.out for the global ratchet.
