---
tag: ci
category: gotcha
created_at: 2026-05-04T02:04:28Z
identity: ci-1
tier: draft
---

CRAP score regressions in CI are typically caused by cross-platform coverage divergence, not floating-point precision differences. IEEE 754 guarantees bit-identical arithmetic results for the CRAP formula on both ARM64 and x86_64, but the coverage percentages fed into the formula differ because platform-conditional code paths (macOS vs Linux branches, tool availability via LookPath, SELinux detection) exercise different lines on each platform. The systemic fix is twofold: (1) generate the CRAP baseline ON the CI runner via a post-merge workflow so comparisons are always same-platform, and (2) add epsilon tolerance (ε=0.5) to regression detection to absorb minor fluctuations from Go toolchain updates or non-deterministic coverage measurement. The epsilon tolerance was implemented as a local evaluate job in ci_crapload.yml since the upstream reusable workflow (complytime/org-infra) does not support a regression-epsilon input.
