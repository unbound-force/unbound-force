---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-04T20:08:35Z
identity: supply-chain-20261104T200835-jay-flowers
tier: draft
---

A repo that stops tracking its package-lock.json can keep transitive-dependency vulnerability scanning without reintroducing the tracked-lock conflict: run `npm install --package-lock-only --ignore-scripts` in CI (generates a lock without installing anything or executing scripts), scan it with osv-scanner, then discard it — the generated lock is covered by the same .gitignore line that stopped tracking it, so it can never be committed by accident. This is the compensating control for the OSV-Scanner/Dependa  [… truncated at capture; final clause unrecoverable — recorded 2026-11-07].
