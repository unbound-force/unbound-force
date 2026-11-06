---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-05T16:21:12Z
identity: supply-chain-20261105T162112-jay-flowers
tier: draft
---

A name+version pin resolved through the package manager's own mechanism (e.g., `npx -y npm@11.6.2` in a Makefile) sits inside npm's built-in integrity verification (tarball checked against registry-served dist.integrity), so a content-hash MUST scoped to downloads outside the package manager's verification does not bind — a missing recorded dist.integrity is a documentation gap on an accepted risk, not a rule violation. When strengthening such a pin, distinguish content-anchored (hash recorded beside the pin for after-the-fact audit; goes silently stale on pin bumps because nothing checks it) from content-verified (an enforcement step compares the recorded hash at run time, e.g., `npm view <pkg>@<ver> dist.integrity` against the recorded value, or a throwaway `npm install --package-lock-only` scan that captures resolved+integrity for the tree). Prefer the enforced form; a bare comment hash implies verification nothing performs.
