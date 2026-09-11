---
description: Convert existing tasks into actionable, dependency-ordered GitHub issues for the feature based on available design artifacts.
tools: ['github/github-mcp-server/issue_write']
---
<!-- scaffolded by uf vdev -->

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Outline

1. Run `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks` from repo root and parse FEATURE_DIR and AVAILABLE_DOCS list. All paths must be absolute. For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").
1. From the executed script, extract the path to **tasks**.
1. Get the Git remote by running:

```bash
git config --get remote.origin.url
```

> [!CAUTION]
> ONLY PROCEED TO NEXT STEPS IF THE REMOTE IS A GITHUB URL

1. For each task in the list, use the GitHub MCP server to create a new issue in the repository that is representative of the Git remote.

> [!CAUTION]
> UNDER NO CIRCUMSTANCES EVER CREATE ISSUES IN REPOSITORIES THAT DO NOT MATCH THE REMOTE URL

## Guardrails

- This command creates **GitHub issues via** the MCP API.
  It does NOT write local files.
- Issues MUST only be created in the repository matching
  the current Git remote. NEVER create issues in
  unrelated repositories.
- **Content portability**: Issue titles and bodies MUST
  use language appropriate to the target repository's
  language and project type. Read README.md or AGENTS.md
  to determine the project type. MUST NOT reference
  tools or concepts specific to a different repository
  type (e.g., "organization-configuration repository",
  "Peribolos", "safe-settings") unless the target
  repository actually uses them. See convention pack
  rules CP-001 through CP-003.
- Do NOT modify source code, spec artifacts, or any
  local files.
