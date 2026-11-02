---
tag: spec-authoring
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:08:57Z
identity: spec-authoring-20261101T220857-jay-flowers
tier: draft
---

OpenSpec spec-delta numbering headers that cite sibling capabilities should use paths that resolve at review time. Unarchived changes keep their capability deltas at openspec/changes/<change>/specs/<capability>/spec.md — they only move to openspec/specs/<capability>/spec.md at archive time — and the numbered Speckit tree (specs/NNN-*/) never holds them. In the model-matrix-scaffold-deployment review (issue #889), the delta's allocation header cited specs/model-matrix/spec.md and specs/unleash/spec.md, neither of which existed; the requirements actually lived in the unarchived slice-1 change directory. When reviewing or authoring allocation headers, verify sibling citations with three globs — specs/*/spec.md, openspec/specs/*/spec.md, openspec/changes/*/specs/*/spec.md — and cite the path that actually resolves (or note it will resolve post-archive).
