---
description: "Creates a new backlog item"
agent: muti-mind-po
---

# Command: /muti-mind.backlog-add

## Description

Creates a new backlog item in the `.uf/muti-mind/backlog/` directory.

## Usage

```
/muti-mind.backlog-add --type <type> --title "<title>" [--priority <P1-P5>] [--description "<desc>"]
```

### Arguments

- `--type` (required): `epic|story|task|bug`
- `--title` (required): A short descriptive title
- `--priority` (optional): `P1`, `P2`, `P3`, `P4`, `P5`
- `--description` (optional): Detailed description and acceptance criteria

## Instructions

1. Use the `bash` tool to invoke the Go backend to create the backlog item:
   ```bash
   go run cmd/mutimind/main.go add --type "$TYPE" --title "$TITLE" --priority "$PRIORITY" --description "$DESCRIPTION"
   ```
2. Parse the output to get the generated `BI-NNN` ID.
3. Return a success message with the new ID.
