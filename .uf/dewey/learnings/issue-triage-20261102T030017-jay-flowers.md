---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:00:17Z
identity: issue-triage-20261102T030017-jay-flowers
tier: draft
---

Triage lesson from issue #672 (resolve_domain_owner): an issue's acceptance criteria can name a config file as the authoritative source for data that file does not contain. #672 requires the tool to read .uf/reviewer-capabilities.yaml "for authoritative mappings", but that manifest's v1.0.0 schema (additionalProperties: false) carries only agent/capability/scopes — many-to-many eligibility categories (security appears under both divisor-adversary and divisor-sre; documentation under architect, curator, and guard) — while the one-owner-per-dimension mapping the tool's examples promise lives in agent files and specs/019-divisor-council-refinement/data-model.md, and none of the issue's example domains (architecture, operations) exist in the scope enum. The issue also carries two contradictory data-flow statements (code constants inside the tool vs. runtime manifest reads) — a leftover from its struck-through predecessor design. Verify file-reference claims against the named file's actual schema at triage time, and treat narrative/criteria contradictions in issues born from a revised proposal as likely superseded-design leftovers: flag them with current-state facts so the clarify/plan phase picks the design deliberately instead of discovering the gap during implementation.
