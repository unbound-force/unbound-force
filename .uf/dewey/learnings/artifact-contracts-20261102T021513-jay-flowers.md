---
tag: artifact-contracts
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:15:13Z
identity: artifact-contracts-20261102T021513-jay-flowers
tier: draft
---

The /uf.unleash summary artifact (.uf/artifacts/unleash-summary/<sanitized-branch>.json) mixes field-naming conventions: snake_case for change_name, feature_dir, review_verdicts, learnings_stored and camelCase for the demo fields whatWasBuilt, howToVerify, testResults. Issue #739 (Increment 2 of #716) labels the body "camelCase", which misdescribes the first group. When documenting this contract (producer write_unleash_summary in uf-workflow, consumer uf_finale_pr_context in uf-finale), preserve the exact implemented field names from the tool schema — the consumer parses literal keys, so normalizing names to a prose label would document a contract nothing reads. The artifact additionally carries a provenance metadata block (artifact_type, schema_version, branch, timestamp) per org constitution Principle III, which the issue's eight-field list omits; contract documentation should cover both layers.
