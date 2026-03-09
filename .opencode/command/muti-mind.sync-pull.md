---
description: "Pulls GitHub Issues into the local backlog"
agent: muti-mind-po
---

# Command: /muti-mind.sync-pull

## Description

Pulls recent updates from GitHub Issues and updates the local backlog items.

## Usage

```
/muti-mind.sync-pull
```

## Instructions

1. Use the `bash` tool to invoke the Go backend:
   ```bash
   go run cmd/mutimind/main.go sync-pull
   ```
2. Output the results returned by the backend.
