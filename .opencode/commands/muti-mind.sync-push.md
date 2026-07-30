---
description: "Pushes local backlog items to GitHub Issues"
agent: muti-mind-po
---

# Command: /muti-mind.sync-push

## Description

Pushes local backlog items to GitHub Issues using the Go backend, which relies on the `gh` CLI. If an item doesn't have an associated GitHub Issue, it will be created. If it does, the existing issue will be updated.

## Usage

```
/muti-mind.sync-push [item_id]
```

### Arguments

- `item_id` (optional): If provided, only pushes the specified backlog item (e.g., `BI-001`). If omitted, pushes all items.

## Instructions

1. **Preview**: Use the `bash` tool to invoke the Go
   backend in dry-run mode to see what will be synced:
   ```bash
   go run cmd/mutimind/main.go sync-push --dry-run [item_id]
   ```

2. **Present preview**: Display the dry-run output to
   the user so they can see what items will be created
   or updated on GitHub.

3. **Confirmation gate**: If the preview shows items
   pending sync (i.e., the output is NOT "No items
   pending sync"), use the **AskUserQuestion tool** to
   request explicit confirmation before proceeding:
   - Options:
     `["Confirm -- sync to GitHub", "Abort -- cancel sync"]`
   - If the user selects **"Confirm -- sync to GitHub"**:
     proceed to step 4.
   - If the user selects **"Abort -- cancel sync"**:
     inform the user that the sync was cancelled and
     stop. Do NOT invoke the Go backend.

4. **Execute**: If the user confirmed, use the `bash`
   tool to invoke the Go backend without `--dry-run`:
   ```bash
   go run cmd/mutimind/main.go sync-push [item_id]
   ```

5. **Output results**: Display the results returned by
   the backend.

6. **No items pending**: If the dry-run preview reports
   "No items pending sync", skip the confirmation step
   and inform the user that there is nothing to sync.
