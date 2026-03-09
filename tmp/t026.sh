#!/bin/bash
set -e

MANIFEST="schemas/hero-manifest/muti-mind-hero.json"
TMP_FILE="tmp/manifest.json"

if [ -f "$MANIFEST" ]; then
    # Add mcp_server to dependencies
    jq '.dependencies = (.dependencies // []) + [{"type": "mcp_server", "name": "graphthulhu", "description": "Required for querying the knowledge graph of backlog items", "required": true}]' "$MANIFEST" > "$TMP_FILE"
    mv "$TMP_FILE" "$MANIFEST"
    echo "Updated $MANIFEST"
else
    echo "$MANIFEST not found"
    exit 1
fi
