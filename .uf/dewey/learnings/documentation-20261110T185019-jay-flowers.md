---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-10T18:50:19Z
identity: documentation-20261110T185019-jay-flowers
tier: draft
---

A convention introduced as code (contract test) plus in-template prose is not a documented convention. The heredoc ban for command templates (TestCommandContracts_RejectShellHeredocPayloads, unbound-force #1011 / PR #1024) runs only in the meta repo's Go suite against .opencode/commands/ and internal/scaffold/assets/opencode/commands/; uf init deploys the templates to hero repos without that guard, so outside the meta repo the ban's only carrier is the prose inside the single template that was fixed. Precedent for the missing half: the clarification-marker grammar (mechanically enforced) is still documented in AGENTS.md Coding Conventions, and convention additions DE-007/DE-008 each shipped with docs issues (#781, #791). When a change adds a test-enforced authoring rule, check whether AGENTS.md or the governing convention pack also needs the one-line record.
