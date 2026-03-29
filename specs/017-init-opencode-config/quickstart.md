# Quickstart: Init OpenCode Config

## Verify After Implementation

```bash
# Fresh repo test
cd $(mktemp -d)
uf init
cat opencode.json
# Should contain mcp.dewey (if dewey installed)
# and plugin array (if .hive/ exists)

# Idempotent test
uf init
# Should report "already configured"

# Force test
uf init --force
# Should report "overwritten"

# Doctor test
uf doctor
# MCP Server Config group should show dewey binary
```

## Expected opencode.json Output

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "type": "local",
      "command": [
        "dewey",
        "serve",
        "--vault",
        "."
      ],
      "enabled": true
    }
  },
  "plugin": [
    "opencode-swarm-plugin"
  ]
}
```
