# Muti-Mind Quickstart

## Installation

Muti-Mind operates as an OpenCode agent backed by a local CLI tool and requires the `graphthulhu` MCP server to be active.

1. Ensure the `unbound` repo is initialized and `graphthulhu` MCP server is running.
2. Ensure you have GitHub CLI (`gh`) authenticated if you intend to sync with GitHub Issues.
3. In OpenCode, install the Muti-Mind persona and command wrappers:
   ```bash
   # (Assuming future opencode plugin/agent install commands)
   # Muti-Mind agents are located in .opencode/agents/muti-mind-po.md
   ```

## Initialization

Initialize a new backlog in your project:

```bash
/muti-mind.init
```
This creates the `.muti-mind/backlog/` directory.

## Managing the Backlog

Add a new item:
```bash
/muti-mind.backlog-add --type story --title "Implement user login" --priority P2
```

View the backlog:
```bash
/muti-mind.backlog-list
```

## AI Prioritization

To let Muti-Mind analyze dependencies via the Knowledge Graph and re-score the backlog:

```bash
/muti-mind.prioritize
```

## GitHub Sync

Push your local backlog to GitHub Issues:

```bash
/muti-mind.sync-push
```

Pull new issues from GitHub into your local backlog:

```bash
/muti-mind.sync-pull
```