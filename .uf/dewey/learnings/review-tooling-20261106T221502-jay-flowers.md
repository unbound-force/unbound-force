---
tag: review-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-06T22:15:02Z
identity: review-tooling-20261106T221502-jay-flowers
tier: draft
---

Search tools that respect .gitignore (ripgrep-backed grep wrappers) silently skip gitignored workspace-state paths — .uf/feedback/, .uf/progress/, .uf/sibling-clones/, .uf/dewey/cache/ per the root .gitignore — so a no-match grep result is not absence evidence for files under those paths. During the 2026-11-06 iteration-3 review of chore/dewey-learnings-coffee-20261106, a repo-wide grep for the coffee-lineage symbols (PLUGIN_TEST_NPM, osv_plugin_transitive, checkManifestInstallConsistency) returned matches only under .uf/dewey/learnings/, apparently falsifying the lineage notes' claim that .uf/feedback/pr-957/state.json is those symbols' only other local trace; a direct read of that path confirmed the file exists and carries all the claimed references — it is gitignored, hence skipped by the filtered search. Extends review-tooling-20261106T215704 (glob dot-directory false negatives): dot-directory globbing and gitignore filtering are two distinct false-absence mechanisms. When verifying a committed claim that names a specific path as a trace or artifact source, read that path directly before treating a search no-match as a contradiction — and when authoring verification claims, name the search mode (gitignore-filtered vs plain grep -r) so reviewers can reproduce the result.
