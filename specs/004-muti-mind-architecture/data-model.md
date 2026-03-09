# Data Model: Muti-Mind Architecture

## 1. Backlog Item

The core entity managed by Muti-Mind, stored as a local Markdown file with YAML frontmatter.

**Storage**: `.muti-mind/backlog/BI-NNN.md`

### Fields (YAML Frontmatter)
- `id` (string): Unique identifier (e.g., `BI-001`).
- `title` (string): Short descriptive title.
- `type` (enum): `epic`, `story`, `task`, `bug`.
- `priority` (enum): `P1`, `P2`, `P3`, `P4`, `P5`.
- `status` (enum): `draft`, `ready`, `in-progress`, `review`, `done`, `cancelled`.
- `sprint` (string, optional): Assigned sprint or iteration.
- `effort_estimate` (string, optional): T-shirt size (XS, S, M, L, XL) or story points.
- `dependencies` (array of strings): List of other BI IDs this item depends on.
- `related_specs` (array of strings): File paths to related specification documents.
- `github_issue_number` (int, optional): The corresponding issue number when synced.
- `created_at` (timestamp): ISO 8601 creation time.
- `modified_at` (timestamp): ISO 8601 last modification time.

### Body (Markdown)
- `Description`: The narrative description of the work.
- `Acceptance Criteria`: Given/When/Then scenarios.

## 2. Priority Score

A computed entity generated during the `/muti-mind.prioritize` action. It is not permanently stored but can be attached to the Backlog Item frontmatter or outputted as an artifact.

### Fields
- `item_id` (string): Reference to the Backlog Item.
- `business_value` (integer 0-10): Derived from user value.
- `risk` (integer 0-10): Assessment of technical or market risk.
- `dependency_weight` (integer): Boost applied if this item blocks others.
- `urgency` (enum): `low`, `medium`, `high`, `critical`.
- `effort` (enum): Copied from the item's `effort_estimate`.
- `composite_score` (integer 0-100): The final calculated rank score.
- `rank` (integer): Position in the ordered backlog.

## 3. Acceptance Decision

A machine-parseable JSON artifact produced when Muti-Mind reviews a Gaze Quality Report via the Review Council. Conforms to the Spec 002 Artifact Envelope.

**Storage**: Transmitted to orchestrator/heroes, or saved in `.muti-mind/decisions/`

### Fields (Payload)
- `item_id` (string): Reference to the origin Backlog Item.
- `decision` (enum): `accept`, `reject`, `conditional`.
- `rationale` (string): Markdown explanation of the decision.
- `criteria_met` (array of strings): List of acceptance criteria that passed.
- `criteria_failed` (array of strings): List of acceptance criteria that failed.
- `gaze_report_ref` (string): Path to the triggering quality report.
- `decided_at` (timestamp): ISO 8601 decision time.

## 4. Sync State

Tracking information for the GitHub synchronization process. Usually maintained in a lightweight local DB or embedded within the Backlog Item frontmatter.

### Fields
- `item_id` (string): Reference to the Backlog Item.
- `github_issue_number` (int): The remote ID.
- `last_synced_at` (timestamp): ISO 8601 sync time.
- `local_hash` (string): Hash of the local file state at sync time.
- `remote_hash` (string): Hash of the remote issue state at sync time.
- `sync_status` (enum): `synced`, `local-modified`, `remote-modified`, `conflict`, `local-only`, `remote-only`.
