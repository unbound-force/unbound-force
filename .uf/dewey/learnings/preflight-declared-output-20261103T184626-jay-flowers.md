---
tag: preflight-declared-output
author: jay-flowers
category: pattern
created_at: 2026-11-03T18:46:26Z
identity: preflight-declared-output-20261103T184626-jay-flowers
tier: draft
---

A tool that captures a field is not enough to get it to a bash-denied divisor. The skill's report template has to give that field a place, or the evidence never reaches `pre_flight_results`. uf_preflight_declared_checks returned `checks[].output` from the start, but the pre-flight Phase 5 template only had the Execution Results table. The skill said, in prose, to "include each non-PASS check's output", so the model had to choose placement, fencing and trimming (a DE-003 violation). It also dropped advisory checks that exit 0, whose output (bash-impact IMPACT rows) is the evidence. The fix pattern is to render a second string in the tool (`rendered_output`) next to the table rows (`rendered`), give it its own named report section, and have the skill insert it verbatim. For multi-line untrusted output in markdown, use a fence of max(3, longest backtick run + 1) backticks and a `text` info string, so no output line can close the block and inject headings or a fake verdict. This is the fenced-block twin of the codeSpan rule for table cells. Filter on "executed (exit_code not null) with non-whitespace output", not on status. Learned on branch opsx/preflight-declared-output on 2026-11-03, while wiring fullsend-ai's divisor-fullsend to judge bash-env-contract finding lines.
