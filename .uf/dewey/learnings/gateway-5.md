---
tag: gateway
category: gotcha
created_at: 2026-04-28T00:44:56Z
identity: gateway-5
tier: draft
---

When adding new fields (like tokenExpiry, credExpiry) to provider structs that have existing tests constructing the struct directly, every existing test must be updated to set the new field to a valid value. Zero-value time.Time is in the past, which triggers expiry checks and causes test failures if the expiry check fires before the empty-token check. The fix was to add tokenExpiry: time.Now().Add(30 * time.Minute) to all 8 existing test struct literals. The spec review caught this proactively (Divisor Testing agent task 6.15) by requiring an explicit list of test functions to update, preventing the vague "where applicable" phrasing that could lead to missed updates.
