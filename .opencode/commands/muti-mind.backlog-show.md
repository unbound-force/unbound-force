---
description: "Displays full details of a backlog item"
agent: muti-mind-po
---

# Command: /muti-mind.backlog-show

## Description

Displays the full details of a specific backlog item, including its YAML frontmatter properties and Markdown body.

## Usage

```
/muti-mind.backlog-show <item_id>
```

### Arguments

- `item_id` (required): The ID of the backlog item (e.g., `BI-001`)

## Instructions

1. Use the `bash` tool to invoke the Go backend to show the backlog item details:
   ```bash
   go run cmd/mutimind/main.go show <item_id>
   ```
2. Output the details returned by the backend.
