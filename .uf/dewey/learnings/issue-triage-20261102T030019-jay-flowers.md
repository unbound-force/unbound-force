---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:00:19Z
identity: issue-triage-20261102T030019-jay-flowers
tier: draft
---

When triaging a Constitution VI tool-extraction issue that names a config file as its data source, read that file and check three contract properties before approving: vocabulary match (do the issue's example inputs exist as keys/scopes in the named file?), cardinality (does the file map an input to one owner or many?), and prose-vs-criteria consistency (constants-in-code and reads-file-at-runtime are different architectures with different test surfaces — constants need drift-detection tests against the canonical file; runtime reads need missing/malformed-file fixture tests). unbound-force #672 (resolve_domain_owner) failed all three checks against .uf/reviewer-capabilities.yaml: its example domains architecture/testing/operations are not manifest scopes (the manifest vocabulary is security, cli-ux, test-quality, documentation, ci-cd, dependencies, standard), 6 of 7 manifest scopes map to multiple agents while the proposed tool returns a single agent, and the issue prose said code constants while its acceptance criteria said runtime manifest reads. Reading the named artifact at triage catches these contradictions in seconds and prevents acceptance criteria that cannot be satisfied simultaneously.
