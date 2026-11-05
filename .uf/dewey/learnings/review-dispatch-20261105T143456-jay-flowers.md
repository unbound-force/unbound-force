---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:34:56Z
identity: review-dispatch-20261105T143456-jay-flowers
tier: draft
---

When a contract test pins spec scenarios as prose fragments across consumer templates, verify each template actually carries the scenario's full MUST wording before writing the pin — and align the prose first where it falls short. A pin matched to current prose locks whichever wording is weakest: on PR #949 (RD-FR-056), pinning uf.unleash.md's 'Skip storage when the finalization failed ... and say so in the report' would have frozen a sentence omitting the mandated 'report the proposals as stranded in the dispatch store' clause, while a pin matched to the spec wording fails on the weak template and surfaces the compliance gap. Enumerate pins by spec scenario (not by consumer command or mechanism), diff each scenario's THEN-clause against every template's prose, strengthen the weak copies (canonical and scaffold mirror, since loadCommandCopies checks both), then pin the strengthened wording.
