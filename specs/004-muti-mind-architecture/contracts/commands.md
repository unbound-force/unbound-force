# Muti-Mind Command Contracts

The primary interface for Muti-Mind is a set of OpenCode commands. These act as the execution layer for the AI agents and Swarm orchestration.

## 1. Backlog Management

### `/muti-mind.init`
Initializes the Muti-Mind environment in the current repository.
- **Action**: Creates `.muti-mind/backlog/` directory and `.muti-mind/config.yaml`.
- **Output**: Success confirmation message.

### `/muti-mind.backlog-add`
Creates a new backlog item.
- **Arguments**: 
  - `--type` (required): `epic|story|task|bug`
  - `--title` (required): String
  - `--priority` (optional): `P1-P5`
  - `--description` (optional): String
- **Action**: Generates a new `BI-NNN.md` file in the backlog directory.
- **Output**: ID of the newly created item.

### `/muti-mind.backlog-list`
Lists current backlog items.
- **Arguments**:
  - `--status` (optional): Filter by status
  - `--sprint` (optional): Filter by sprint
- **Output**: Formatted list/table of items sorted by priority (P1 first).

### `/muti-mind.backlog-update`
Updates an existing backlog item.
- **Arguments**:
  - `[item_id]` (required): e.g., `BI-003`
  - `--priority` (optional): New priority
  - `--sprint` (optional): New sprint assignment
  - `--status` (optional): New status
- **Action**: Modifies the YAML frontmatter of the target item.
- **Output**: Success confirmation.

### `/muti-mind.backlog-show`
Displays full details of an item.
- **Arguments**:
  - `[item_id]` (required): e.g., `BI-003`
- **Output**: Rendered markdown content and frontmatter details.

## 2. GitHub Synchronization

### `/muti-mind.sync-push`
Pushes local backlog items to GitHub Issues.
- **Arguments**:
  - `[item_id]` (optional): Specific item to push, otherwise pushes all ready/un-synced items.
- **Action**: Creates/updates GitHub Issues, matching labels to type/priority. Updates local `github_issue_number`.

### `/muti-mind.sync-pull`
Pulls GitHub Issues into the local backlog.
- **Action**: Creates/updates local `BI-NNN.md` files from GitHub issues.

### `/muti-mind.sync-status`
Reports on the synchronization state.
- **Output**: Lists items in sync, modified locally, modified remotely, or in conflict.

## 3. AI Capabilities (Agent Delegated)

### `/muti-mind.prioritize`
Invokes the Muti-Mind AI persona to score and rank the backlog.
- **Action**: Agents read the backlog files, compute scores across 5 dimensions, and update the priority fields.
- **Output**: A report showing the new ranking and the dimension breakdown for the scores.

### `/muti-mind.generate-stories`
Invokes the Muti-Mind AI persona to break down a high-level goal.
- **Arguments**:
  - `[goal_description]` (required): A prompt describing the feature need.
- **Action**: Agent generates fully-formed user stories (with acceptance criteria).
- **Output**: Proposed stories presented to the user for approval before being saved via `backlog-add`.
