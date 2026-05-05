---
description: "Updates an existing backlog item"
agent: muti-mind-po
---

# Command: /muti-mind.backlog-update

## Description

Updates an existing backlog item's metadata.

## Usage

```
/muti-mind.backlog-update <item_id> [--priority <P1-P5>] [--status <status>] [--sprint <sprint>]
```

### Arguments

- `item_id` (required): The ID of the backlog item (e.g., `BI-001`)
- `--priority` (optional): The new priority level
- `--status` (optional): The new status
- `--sprint` (optional): The new sprint assignment

## Instructions

1. Use the `bash` tool to invoke the Go backend to update the backlog item:
   ```bash
   go run cmd/mutimind/main.go update <item_id> --priority "$PRIORITY" --status "$STATUS" --sprint "$SPRINT"
   ```
2. Output the success message returned by the backend.
