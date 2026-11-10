---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:46:34Z
identity: instruction-quality-20261108T184634-jay-flowers
tier: draft
---

A quoted heredoc written to a temp file and passed via `gh api --input` satisfies shell-injection prevention (temp file + --input, never on the command line) but does NOT guarantee payload validity: when the consuming API expects JSON, substituting external text (finding summaries, review comments, issue titles) into a JSON string literal breaks on any embedded double quote, newline, or backslash, and the API returns 400. JSON-bound payloads need jq to compose the document (jq -n --arg / --rawfile writing the temp file, preserving the temp-file + --input shape) or gh's own field encoding (gh api -f). Every outward-facing mutation step in instruction prose also needs an explicit failure outcome that preserves protocol invariants — e.g., never post a reply that links an issue that failed to create; retry once, then stop and report.
