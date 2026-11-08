---
tag: resilience
author: jay-flowers
category: gotcha
created_at: 2026-11-08T17:16:50Z
identity: resilience-20261108T171650-jay-flowers
tier: draft
---

Deferring commits of generated artifacts to a periodic sweep (e.g., Dewey learnings staying untracked until a daily learnings PR, PR #993) removes git's durability guarantee for the entire sweep window: git clean -fd, worktree removal/rotation, and re-clone silently and permanently destroy the pending files with no failure signal, and a gitignored sidecar copy on the same checkout (graph.db) is equally non-durable. Risk tables for such deferral policies must cover the destruction vector, not just delivery latency; mitigation options are sweeping at workflow boundaries (retrospective/finale) instead of only on a timer, or warning against destructive git operations while pending artifacts exist.
