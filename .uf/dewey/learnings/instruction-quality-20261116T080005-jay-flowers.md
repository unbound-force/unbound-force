---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-16T08:00:05Z
identity: instruction-quality-20261116T080005-jay-flowers
tier: draft
---

DE-001 extraction gotcha: when a command's rewritten section summarizes a tool's input shape by citing the output schema ("each item carrying the <schema-version> item fields plus X"), the summary silently includes fields the tool derives itself — join outputs like filed_issue_number/filed_issue_url in write_feedback_triage_artifact. Strict (`.strict()`) zod input schemas reject those keys, so the literal reading fails with `invalid_input` and the command's report-and-STOP error rule halts the phase. Extraction prose must name the exclusions explicitly ("do not pass FOO/BAR — the tool derives them from <param>") rather than citing the output schema as the input field list. Symptom to watch for: a command paragraph that is a verbatim lift of a lib doc-comment's first sentence without its clarifying second sentence.
