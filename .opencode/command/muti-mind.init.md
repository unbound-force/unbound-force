---
description: "Initializes the Muti-Mind environment in the current repository"
agent: muti-mind-po
---

# Command: /muti-mind.init

## Description

Initializes the Muti-Mind environment in the current repository by creating the necessary directory structure and default configuration file.

## Usage

```
/muti-mind.init
```

## Instructions

1. Use the `bash` tool to create the directory `.uf/muti-mind/backlog/` if it does not already exist.
2. Use the `write` tool to create the `.uf/muti-mind/config.yaml` file with the default configuration if it does not exist. The default config should be:
   ```yaml
   # Muti-Mind Configuration
   backlog:
     dir: ".uf/muti-mind/backlog"
     default_priority: "P3"
     default_status: "draft"
   github:
     sync_enabled: true
     project_id: ""
   agent:
     persona_file: ".opencode/agents/muti-mind-po.md"
   ```
3. Use the `bash` tool to run the Go backend CLI command `go run cmd/mutimind/main.go init` (if the project has a `cmd/mutimind/main.go` file) to initialize any backend state.
4. Output a success confirmation message.

## Error Handling

- **Already Initialized**: If the directories and configuration already exist, report that Muti-Mind is already initialized.
