---
tag: tolerant-gh-runner
author: jay-flowers
category: pattern
created_at: 2026-11-03T10:58:09Z
identity: tolerant-gh-runner-20261103T105809-jay-flowers
tier: draft
---

When one PR lands module-local twins of a tolerant-CLI runner (duplicate implementations of the same external-contract detection logic), enforce conjunctive-guard symmetry across the twins: every marker recognition implemented as a conjunction of verified-shape facts (empty stdout + exit code + marker regex) in one twin should appear identically in the other, unless the design record documents the asymmetry as deliberate. In unbound-force PR #927 (fix-uf-finale-ci-poll-gh-schema), the review-dispatch twin required stdout === "" before honoring the no-checks stderr marker (per FPR-FR-017) while the uf-finale poller honored the marker on any non-array stdout — identical external gh contract, unequal defense-in-depth, with the looser twin being the one whose output gates the finale's CI-watch decision. The asymmetry was self-declared and behavior-identical on all verified gh shapes, so it stayed LOW, but unifying the guard (or recording the asymmetry in the design decision) costs far less than the review noise the divergence generates. Review heuristic: when a diff contains two or more copies of the same marker constant (including scaffold twins), diff the guard conditions around each copy, not just the regex.
