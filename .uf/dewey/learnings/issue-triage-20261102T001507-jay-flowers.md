---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:07Z
identity: issue-triage-20261102T001507-jay-flowers
tier: draft
---

Migration cutover phase issues can bundle two completion horizons: the cutover itself completes as soon as its start-gates hold (config converted, defaults flipped, release tagged, support branch cut), while post-window cleanup — e.g., removing v1 server() implementations "once the support window closes and no v1 consumers remain" — triggers on a 90-day clock with its own verifiable conditions (window elapsed, zero remaining consumers). Tracking the time-deferred cleanup inside the phase issue holds that issue open for the full support window after the primary work is done. Triage should recommend splitting the time-deferred item into its own tracking issue with an explicit trigger gate, extending the established deferred-scope split pattern (learning/issue-triage-20261101T222239-jay-flowers) to time-clock-gated deferrals; the phase issue then closes at cutover completion. Verified triaging #807 (Phase 5 cutover, umbrella #797).
