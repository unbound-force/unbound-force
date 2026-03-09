# Muti-Mind Quickstart

## 1. Initialization

Initialize the Muti-Mind backlog structure in your repository:

```bash
/muti-mind.init
```
This creates the `.muti-mind/backlog/` directory where all items will be stored as Markdown files.

## 2. Managing the Backlog

You can ask the Muti-Mind agent to generate stories for you:

```bash
/muti-mind.generate-stories "Users need to be able to export their data in CSV format"
```

Or you can add items directly:

```bash
/muti-mind.backlog-add --type story --title "CSV Data Export" --priority P2
```

List your current backlog:

```bash
/muti-mind.backlog-list
```

## 3. Prioritization

If your backlog is growing, you can ask the Muti-Mind agent to re-score and prioritize it based on business value, risk, and dependencies:

```bash
/muti-mind.prioritize
```

## 4. GitHub Sync

Keep your local MD cache in sync with the repository's GitHub Issues:

```bash
# Push local changes to GitHub
/muti-mind.sync-push

# Pull new issues from GitHub
/muti-mind.sync-pull

# Check sync status
/muti-mind.sync-status
```
