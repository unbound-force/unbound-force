---
tag: address-feedback
author: jay-flowers
category: gotcha
created_at: 2026-11-16T15:42:01Z
identity: address-feedback-20261116T154201-jay-flowers
tier: draft
as_of: "unbound-force@c2dfd4b8f15f9ae106c599aef9a0418bf075dd93"
---

Editing .opencode/commands/uf.address-feedback.md prose (quick branch quick/address-feedback-skip-branch-terminal-state, 2026-11-16, issue #1029) is constrained by three test pins invisible in the file itself: TestQuestionToolPresentationContractGates counts exact option-label strings per file — a new paragraph that quotes the Section 4.5 gate label "No -- skip posting" pushes the per-label occurrence total off its pin, so paraphrase as "skip-posting path" instead; TestCommandContracts_Feedback* asserts verbatim clause strings on BOTH the live command and its internal/scaffold/assets mirror (matched after lowercase normalization), so pinned sentences like "it files each file item's issue before its reply (one retry; a file item whose issue still fails is not posted or resolved)" must survive edits unchanged; and TestEmbeddedAssets_MatchSource enforces byte-identical mirroring of the command into internal/scaffold/assets/opencode/commands/. The working sequence: read the pins in internal/scaffold/command_contract_test.go before authoring, avoid the forbidden legacy envelope strings (the old hand-written 4.7 JSON template, "scan existing files for the highest round number", "after the artifact is written, call clear_progress"), mirror with cp, then run go test ./internal/scaffold/... — the label-count failure surfaces only at test time, never in lint.
