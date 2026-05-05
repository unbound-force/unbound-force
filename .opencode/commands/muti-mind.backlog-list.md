---
description: "Lists current backlog items"
agent: muti-mind-po
---

# Command: /muti-mind.backlog-list

## Description

Lists current backlog items in the `.uf/muti-mind/backlog/` directory, sorted by priority.

## Usage

```
/muti-mind.backlog-list [--status <status>] [--sprint <sprint>]
```

### Arguments

- `--status` (optional): Filter items by a specific status (e.g., `draft`, `ready`, `done`)
- `--sprint` (optional): Filter items by a specific sprint

## Instructions

1. Use the `bash` tool to invoke the Go backend to list the backlog items:
   ```bash
   go run cmd/mutimind/main.go list --status "$STATUS" --sprint "$SPRINT"
   ```
2. Output the results returned by the backend.
