---
tag: scaffold-content-gates
author: jay-flowers
category: reference
created_at: 2026-11-02T04:53:30Z
identity: scaffold-content-gates-20261102T045330-jay-flowers
tier: draft
---

Scaffold content-gate tests in internal/scaffold/scaffold_test.go (TestQuestionToolPresentationContractGates, TestConsolidateDispatchFinalizationContractGates) read only the embedded mirror assets via assetContent("opencode/commands/<name>") — not the live .opencode/commands/ files. Canonical-source coverage comes transitively: TestEmbeddedAssets_MatchSource walks the full internal/scaffold/assets/ tree and enforces byte-identity against the repo-root canonical files, so embedded-asset pins cover both copies. The dual-read pattern (live canonical + embedded mirror) lives in command_contract_test.go's assertCommandCopies. When a design doc or task describes a new content gate, name the actual read path (embedded mirror + drift-test transitivity vs. assertCommandCopies dual read) — describing an embedded-only gate as reading canonical sources creates record-versus-code drift even though the enforced contract is identical.
