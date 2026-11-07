---
tag: convention-enforcement
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:40:48Z
identity: convention-enforcement-20261107T184048-jay-flowers
tier: draft
---

When a change hand-syncs a duplicated value across surfaces, enumerate every surface carrying the value and give each enforceable pair a parity guard or a single source — the guard belongs to the change that creates or multiplies the surfaces, not only to the pair that triggered reviewer feedback. Precedent: PR #957 feedback fixes added TestPluginTestNpmPin_TwinsCISecurityPin for the Makefile↔ci_security.yml npm-pin pair while the same diff hand-synced npm@11.6.2 into four remediation constants (scaffold/doctor, D2 twin commitment), a new copyable docs recipe carrying the npm/osv-scanner/GOTOOLCHAIN pins, and an amended spec-delta scenario — none of which the twin test reads. A future pin bump therefore fails loudly on two surfaces and drifts silently on consumer-shipped remediation guidance and the recipe hero repos copy verbatim. Cheapest remedies: extend the twin test to parse the doc and both constant sites (findProjectRoot precedent), or extract the four constants into one shared internal package per the D2 extraction clause once the strings carry conditional guidance.
