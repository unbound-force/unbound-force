---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:01:09Z
identity: issue-triage-20261102T020109-jay-flowers
tier: draft
---

Issue triage gotcha: quoted identifiers in issue comments can drift from current repository state. On issue #747 (docs gate for the command-prose-pilot rewrite of /uf.run-pipeline-plan), the follow-up comment cites "two /uf.unleash Step 8 references" in AGENTS.md, but the current AGENTS.md reads "/uf.unleash Step 6, Code Review" at both cited locations (the Review council behavioral rule and the PR Review Commands table row). The actionable fix — converting those references to the name-based house standard the pilot established — remained valid regardless of the quoted number. When triaging documentation follow-ups, locate the referenced content by its described section or role rather than by quoted numbers or literal strings, verify against current repo state, and treat a number mismatch as corroborating evidence when the issue is about exactly that fragility (number-based cross-references breaking under renumbering).
