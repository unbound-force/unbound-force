---
tag: learning-hygiene
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:59:32Z
identity: learning-hygiene-20261102T055932-jay-flowers
tier: draft
---

Learning artifacts captured during spec review or retrospective phases cite planned artifact names (test functions, task ids, code paths) that implementation later renames — citations rot at rename time because learnings are write-once files. Case: the scaffold-drift learning in init-override-stubs-gitignore-sync (2026-11-02) cited "TestRun_GitignoreBlockTracksRepositoryFile, task 5.2", but the shipped change named the drift test TestGitignoreBlock_RepositoryManagedRegionMatchesCanonical (task 2.3, byte-for-byte region equality via the shared extractManagedRegion helper) and pinned TestPluginPackageInputs_GitignoreTracksCanonicalInputs to the opencodeGitignore shared constant (task 5.2). Review technique: when a diff adds .uf/dewey/learnings/*.md files, grep every cited identifier (test names, function names, task numbers, file:line) against the final working tree — a citation matching nothing is a retrieval dead-end for every future semantic search, and an understated assertion description quietly weakens the standard the change established. The spec-authoring discipline "verify quotes byte-exact before publishing" extends verbatim to learning citations, but must be applied at review time, after implementation renaming has happened, not at capture time.
