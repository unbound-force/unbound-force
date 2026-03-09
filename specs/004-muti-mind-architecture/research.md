# Research: Muti-Mind Architecture

## Decision 1: AI Delegation Strategy
- **Decision**: AI features (priority scoring, story generation) will be delegated to the OpenCode runtime via dedicated commands (`/muti-mind.prioritize`, `/muti-mind.generate-stories`) or agents.
- **Rationale**: Keeps the architecture simple and aligned with the Unbound Force ecosystem. Avoids building a heavy CLI that duplicates OpenCode's LLM management, API keys, and context handling.
- **Alternatives considered**: Building a Go-based CLI that calls OpenAI/Anthropic APIs directly (rejected due to duplication of effort and breaking the OpenCode integration pattern).

## Decision 2: Local Backlog Persistence & Indexing
- **Decision**: Backlog items are stored as individual Markdown files with YAML frontmatter in `.muti-mind/backlog/`. No central index file is used for ordering.
- **Rationale**: Individual MD files are perfect for `graphthulhu` to index as a knowledge graph, enabling semantic search and context retrieval for the Swarm. GitHub remains the source of truth for exact ordering, avoiding local index file drift/conflicts.
- **Alternatives considered**: 
  - A single `backlog.json` file (rejected: poor readability, harder for agents to consume/diff).
  - Local index file (rejected: creates a secondary source of truth that competes with GitHub UI).

## Decision 3: Execution Environment
- **Decision**: Muti-Mind's interface will be exclusively OpenCode commands and agents. A Go binary will only be introduced if an MCP server is strictly required for complex GitHub API interactions that exceed simple bash scripting.
- **Rationale**: Aligns with "Intent-to-Context" and ensures the user stays in the AI environment where the rest of the swarm operates.
- **Alternatives considered**: A standalone CLI binary as the primary interface (rejected: breaks the seamless swarm UX).

## Decision 4: Knowledge Graph Integration
- **Decision**: Muti-Mind OpenCode agent MUST exclusively use `graphthulhu` MCP tools for all read operations, using the CLI only for write/sync operations. It is a hard dependency.
- **Rationale**: Relying on MCP ensures the agent has access to the full dependency graph and semantic search capabilities, adhering to Spec 010.
- **Alternatives considered**: Soft dependency with CLI fallback (rejected because it fractures the architecture and creates inconsistent AI context retrieval).

## Decision 5: Dependency Resolution
- **Decision**: Combine explicit YAML `dependencies[]` with knowledge graph link traversal to build a dependency map.
- **Rationale**: Ensures accurate priority scoring by factoring in both manual constraints and implicit links discovered by the graph.
- **Alternatives considered**: Solely relying on explicit YAML dependencies (rejected because it misses systemic context).

## Decision 6: Edge Case Handling (MCP)
- **Decision**: Implement a pagination loop for handling large backlog queries via MCP and fail fast with clear errors on MCP failure.
- **Rationale**: Graphthulhu limits results to conserve tokens; pagination is necessary for full-backlog operations like reprioritization.
- **Alternatives considered**: Local file reads for full backlog (rejected due to exclusive read constraint on MCP).
