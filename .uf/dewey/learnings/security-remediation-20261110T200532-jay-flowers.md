---
tag: security-remediation
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:05:32Z
identity: security-remediation-20261110T200532-jay-flowers
tier: draft
---

When a pinned-override set grows, every enumeration site must move in the same change: both package manifests (canonical + scaffold mirror), the provenance record and its pinning test, the runbook's advisory paragraph, provenance list, rollback note, and the manual verification command. PR #1026 (effect 4.0.0-beta.103 override for GHSA-chp3-j5qr-2vfr) updated every site except the runbook's `npm ls zod toml` confirmation step, leaving the documented check unable to confirm the new override. The verification command is the easily-missed site because it reads as an example rather than an enumeration; when adding an override, grep the runbook for every mention of the existing override names and update the command and its expected-versions sentence to enumerate the full override set.
