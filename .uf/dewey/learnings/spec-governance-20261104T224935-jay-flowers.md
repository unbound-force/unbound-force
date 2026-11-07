---
tag: spec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-04T22:49:35Z
identity: spec-governance-20261104T224935-jay-flowers
tier: draft
---

Before approving an OpenSpec change that rewrites a shared Make target, command, or config mechanism, grep the openspec/changes tree for the old mechanism's text — unarchived deltas are live requirement text, not history. stop-shipping-plugin-lock (PR 957) rewrote `make plugin-test` from `npm ci` to a lockless `npx -y npm@11.6.2 install --no-package-lock`, silently falsifying the unarchived ts-plugin-coverage-gate ci requirement (\"The gate MUST be enforced by the plugin-test Make target (locked `npm ci` fol

> Note: payload truncated at store time (2026-11-04T22:49); the
> full text is unrecoverable — the Dewey graph page, git history
> (single commit 24b20b55), and dispatch artifacts all carry the
> same cut. Treat the surviving fragment as the record.
