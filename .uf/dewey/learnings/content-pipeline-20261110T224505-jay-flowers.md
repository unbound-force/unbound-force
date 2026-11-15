---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-10T22:45:05Z
identity: content-pipeline-20261110T224505-jay-flowers
tier: draft
---

Command-template security hardening that changes no observable workflow (heredoc removal in /uf.triage-issue, PR #1024 / #1011; same fix class as #1007's uf.address-feedback heredoc removal) needs no docs/blog/tutorial issue: the sufficient documentation record is a CHANGELOG "Fixed" entry (mechanism, fix pattern, enforcement test, measurable instruction delta per DE-004) plus in-PR updates to the canonical command and its byte-identical scaffold mirror. Discriminator for when a scaffold-mirrored security fix DOES warrant a docs issue: adopter-operational surface. Dependency pins landing in the scaffolded package.json mirror change hero-repo adopters' manifests (effect-security-override got docs issue #1034); template-internal procedure changes alter nothing adopters own, invoke, or observe. Verified on PR #1024: issue #1011 open and correctly referenced, canonical file ends at line 1260 matching the claimed 1250→1260 (+24/−14) delta, zero heredoc operators and zero stale COMMENT_EOF/CHILD_EOF references across both canonical and scaffolded command directories, and no open docs issue tracks either heredoc-removal fix.
