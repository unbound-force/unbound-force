---
description: "Sync GitHub Projects"
agent: muti-mind-po
---

# Command: /muti-mind.sync-project

## Description

Synchronizes the backlog with GitHub Projects (beta).

## Usage

```
/muti-mind.sync-project
```

## Instructions

1. Use the `bash` tool to invoke the Go backend:
   ```bash
   go run cmd/mutimind/main.go sync-project
   ```
2. Output the results returned by the backend.
