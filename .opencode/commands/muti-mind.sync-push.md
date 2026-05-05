---
description: "Pushes local backlog items to GitHub Issues"
agent: muti-mind-po
---

# Command: /muti-mind.sync-push

## Description

Pushes local backlog items to GitHub Issues using the Go backend, which relies on the `gh` CLI. If an item doesn't have an associated GitHub Issue, it will be created. If it does, the existing issue will be updated.

## Usage

```
/muti-mind.sync-push [item_id]
```

### Arguments

- `item_id` (optional): If provided, only pushes the specified backlog item (e.g., `BI-001`). If omitted, pushes all items.

## Instructions

1. Use the `bash` tool to invoke the Go backend:
   ```bash
   go run cmd/mutimind/main.go sync-push [item_id]
   ```
2. Output the results returned by the backend.
