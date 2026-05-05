---
name: workflow-list
description: List all workflows
---

# /workflow list

List all hero lifecycle workflows.

## Usage

```
/workflow list [--status active|completed|all]
```

Default: shows all workflows.

## Behavior

When this command is invoked:

1. **Read all JSON files** from `.uf/workflows/`.

2. **Parse each workflow** and extract: workflow_id, feature_branch, status, started_at.

3. **Filter by status** if the `--status` flag is provided:
   - `active` — only active workflows
   - `completed` — only completed workflows
   - `all` — all workflows (default)

4. **Sort by started_at** descending (most recent first).

5. **Display as a table** with columns: ID, Branch, Status, Started.

6. **If no workflows exist**, report "No workflows found. Run /workflow start to begin."

## Output Format

```
Workflows:
  ID                                    Branch              Status     Started
  wf-feat-health-check-20260320T143000  feat/health-check   active     2h ago
  wf-feat-login-20260319T090000         feat/login          completed  1d ago
  wf-feat-signup-20260318T100000        feat/signup         escalated  2d ago
```
