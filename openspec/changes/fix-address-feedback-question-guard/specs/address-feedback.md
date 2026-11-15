## MODIFIED Requirements

### Requirement: Per-item triage decision presentation (question-tool imperative)

Previously: Section 3.2 Author Decision and Section 3.4 Triage Summary instructed the agent to "Present the **question tool**" at each decision point, with a visibility directive that printed text and question tool MUST present the same details and options. For follow-up decisions (Modify / Reject / Ask) the agent was told to "Use **question tool** … to collect" the follow-up.

The requirement is narrowed to an explicit imperative: at every per-item triage decision point (Section 3.2 initial vote, Section 3.2 follow-up open-ended collection, Section 3.4 triage-summary confirmation, and Phase 4 mandatory gates where a question tool call is required), the command template MUST instruct the agent to **call** the `question` tool and MUST forbid writing the tool invocation as text (e.g., `<invoke name="question">` pseudo-XML) in the reply. The visibility directive (printed text and question tool MUST present the same details and options) is retained unchanged.

#### Scenario: Per-item decision instructs tool call plainly

- **GIVEN** the agent is in Section 3.2 presenting item N of M for triage
- **WHEN** the agent reaches the decision point
- **THEN** the command text the agent is executing MUST contain an explicit instruction to call the `question` tool (not merely "Present") and MUST state not to write `<invoke name="question">` as text, while still requiring the printed item display and the question options to contain the same details

#### Scenario: Follow-up collection instructs tool call

- **GIVEN** the author chose Modify, Reject, or Ask for an item (Section 3.2 Decision table)
- **WHEN** the agent collects the follow-up (alternative approach, reasoning, or clarification question)
- **THEN** the command text MUST instruct the agent to call the `question` tool with open-ended, no-preset-options input, and MUST forbid emitting the call as text in the reply

#### Scenario: Triage-summary and Phase 4 gates retain tool-call imperative

- **GIVEN** the agent is at the Section 3.4 triage-summary confirmation or a Phase 4 mandatory confirmation gate
- **WHEN** the agent presents the confirmation question
- **THEN** the command text MUST contain the same explicit call-the-tool imperative and must not rely on the model to infer a tool call from surrounding prose

## ADDED Requirements

(none — wording tightening only; verification surface is the command template text itself)

## REMOVED Requirements

(none)
