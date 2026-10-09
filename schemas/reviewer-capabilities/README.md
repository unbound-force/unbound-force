# Reviewer Capabilities Schema

Version 1 defines the closed structural contract for
`.uf/reviewer-capabilities.yaml`. The manifest separates review
eligibility from OpenCode agent frontmatter. Agent uniqueness, the
required known-persona table, and eligibility rules are semantic policy
checks owned by the review dispatch planner.

Positive and negative JSON fixtures exercise the schema. Policy fixtures
cover semantic failures that JSON Schema cannot express, including
duplicate agents, conflicting entries, missing known personas, and YAML
aliases.
