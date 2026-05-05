---
description: "Invokes the Muti-Mind AI persona to generate user stories from a high-level goal"
agent: muti-mind-po
---

# Command: /muti-mind.generate-stories

## Description

Takes a high-level goal or feature description and delegates it to the Muti-Mind AI persona. The AI generates well-formed user stories (with Given/When/Then criteria), presents them for user approval, and then adds them to the backlog.

## Usage

```
/muti-mind.generate-stories "<goal_description>"
```

### Arguments

- `goal_description` (required): A prompt or paragraph describing the feature or goal you want to break down into stories.

## Instructions

1. Use the provided `goal_description` to draft a set of user stories. 
2. Ensure each story includes a title, priority, narrative description, and `Given/When/Then` acceptance criteria.
3. Present the drafted stories to the user for interactive review. **Wait for the user to confirm** before proceeding.
4. Once the user approves the stories, use the `bash` tool to invoke `/muti-mind.backlog-add` (or directly use `go run cmd/mutimind/main.go add`) for each approved story.
5. Output a final summary of the created backlog items.
