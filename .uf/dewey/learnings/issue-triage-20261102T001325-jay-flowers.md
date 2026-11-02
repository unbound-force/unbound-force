---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:13:25Z
identity: issue-triage-20261102T001325-jay-flowers
tier: draft
---

Documentation-companion issues for scaffolded command assets are well-formed when they enumerate the full mirror surface: the canonical .opencode/commands/ file, the embedded mirror under internal/scaffold/assets/, and the scaffold test pins. At triage time this enumeration is checkable against the repo's drift-detection invariant (embedded tool-owned assets MUST match canonical sources, enforced by scaffold_test.go pins). A docs issue naming only the canonical copy predicts an incomplete PR that fails scaffold drift tests in CI, so the enumeration itself is a triage-validity signal.
