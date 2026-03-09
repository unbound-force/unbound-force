---
description: "Bidirectional sync including interactive conflict detection"
agent: muti-mind-po
---

# Command: /muti-mind.sync

## Description

Performs a full bidirectional sync with GitHub issues.

## Usage

```
/muti-mind.sync
```

## Instructions

1. Use the `bash` tool to invoke the Go backend:
   ```bash
   go run cmd/mutimind/main.go sync
   ```
2. Output the results returned by the backend.
