---
description: "Invokes the Muti-Mind AI persona to score and rank the backlog"
agent: muti-mind-po
---

# Command: /muti-mind.prioritize

## Description

Delegates the task of prioritizing the backlog to the Muti-Mind AI persona. The AI evaluates existing backlog items, computes scores across five dimensions (business value, risk, dependency weight, urgency, and effort), and updates the item priority fields based on the final ranking.

## Usage

```
/muti-mind.prioritize
```

## Instructions

1. Retrieve the list of all current backlog items by using the `dewey_search` or `dewey_find_by_tag` MCP tools (or other Dewey MCP tools). You MUST use a pagination loop or recursive fetching to ensure all items are retrieved.
2. For each backlog item, read its details (including description and dependencies) using `dewey_get_page`. Do NOT use CLI commands to read backlog files.
3. Evaluate and compute a priority score for each item based on the scoring engine criteria defined in your persona (Business Value, Risk, Dependency Weight, Urgency, Effort). Combine explicit YAML dependencies with knowledge graph link traversal to discover implicit relationships.
4. Rank the items based on the calculated composite score.
5. Apply the new priority levels (P1 to P5) to the backlog items using the `bash` tool to run `mutimind update ...` (or `go run cmd/mutimind/main.go update ...`) based on their new ranking.
6. Generate a final report summarizing the new ranking, displaying the score breakdowns, and providing a brief rationale for the changes.
