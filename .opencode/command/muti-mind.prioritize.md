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

1. Retrieve the list of all current backlog items by using the `/muti-mind.backlog-list` command.
2. For each backlog item, read its details (including description and dependencies) using `/muti-mind.backlog-show <item_id>`.
3. Evaluate and compute a priority score for each item based on the scoring engine criteria defined in your persona (Business Value, Risk, Dependency Weight, Urgency, Effort).
4. Rank the items based on the calculated composite score.
5. Apply the new priority levels (P1 to P5) to the backlog items using the `/muti-mind.backlog-update` command based on their new ranking.
6. Generate a final report summarizing the new ranking, displaying the score breakdowns, and providing a brief rationale for the changes.
