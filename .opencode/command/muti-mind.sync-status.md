---
description: "Reports on the synchronization state"
agent: muti-mind-po
---

# Command: /muti-mind.sync-status

## Description

Displays the current synchronization status of all local backlog items, showing whether they are mapped to a GitHub issue.

## Usage

```
/muti-mind.sync-status
```

## Instructions

1. Use the `bash` tool to invoke the Go backend:
   ```bash
   go run cmd/mutimind/main.go sync-status
   ```
2. Output the results returned by the backend.
