---
tag: contract-tests
author: jay-flowers
category: pattern
created_at: 2026-11-16T03:57:15Z
identity: contract-tests-20261116T035715-jay-flowers
tier: draft
---

Layered positive contract tests are the established pattern for pinning command-template choreography in internal/scaffold/command_contract_test.go (precedents: TestCommandContracts_FixCiCommandStructure, and now TestCommandContracts_TriageIssuePostingStructure for issue #1049). Three layers, each catching regression classes the others miss: (1) presence clauses through assertCommandCopies(t, name, true, clauses) — one call pins both the canonical .opencode/commands/<name> and the byte-identical embedded scaffold mirror via loadCommandCopies; (2) count pins (strings.Count on normalizeCommandText output, iterating loadCommandCopies) holding occurrence counts — e.g. a routing rule that must appear at exactly its two sites, a gh posting verb appearing exactly once so a re-added raw invocation fails the count rather than just presence; (3) forbidden fragments (negative clauses) banning direct shell spellings that an earlier extraction removed — e.g. "gh issue comment" and "--method POST" for the triage-issue posting tools — which is the only layer that catches a regression re-adding the old spelling alongside intact tool prose. Fragments must reproduce the command's exact characters modulo lowercase/whitespace collapse (em-dashes and backticks included), so verify each against the live command text. Before landing, mutation-test the pin: temporarily mutate the canonical command (swap the invocation back to the historical shell spelling, drop the tool clauses), run the focused test, confirm it fails, then git checkout restore the command byte-exact. Provenance: branch quick/pin-the-triage-issue-posting-choreograph, /uf.unleash code review APPROVE 0 findings, 2026-11-15.
