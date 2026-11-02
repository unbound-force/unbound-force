---
tag: managed-region
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:35:35Z
identity: managed-region-20261102T013535-jay-flowers
tier: draft
---

Marker/terminator-delimited managed regions in user-editable files (gitignore, exclude files, generated config headers) must specify line-equality normalization and delimiter inertness in their contracts. Two concrete requirements: (1) Substring marker detection (e.g. strings.Contains) tolerates CRLF line endings, but strict line equality does not — on a CRLF-normalized file the marker line carries a trailing \r, fails equality, and the tool takes the append path, permanently accumulating duplicate managed blocks because the CRLF original never matches. Specify trailing-\r trimming or LF normalization for delimiter comparison. (2) Any sentinel/terminator line placed in a user-consumed format like .git/info/exclude must be a comment line — a non-comment sentinel acts as a live ignore pattern and gets counted by line-parsing consumers (e.g. stealthExcludedPaths parsing all non-comment lines). Case: the gitignoreBlock managed region in internal/scaffold/scaffold.go used substring detection, and its replacement spec's "first line equal to the marker" contract left the normalization unspecified, admitting two implementations with observably different CRLF behavior.
