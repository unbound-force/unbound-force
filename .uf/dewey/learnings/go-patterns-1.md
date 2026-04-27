---
tag: go-patterns
category: pattern
created_at: 2026-04-27T01:57:07Z
identity: go-patterns-1
tier: draft
---

When changing a Go constructor's return signature from *T to (*T, error), the change is mechanical when all callers already return (T, error). The pattern for existing tests is: replace `prov := newFoo()` with `prov, err := newFoo()` followed by `if err != nil { t.Fatalf(...) }`. Use `provErr` as the variable name when the test later needs a separate `err` variable for the function being tested (e.g., TestVertexProvider_Start_GcloudFails tests Start() error separately). Finding all call sites is easy -- the compiler will report them as type errors. Always add dedicated tests for both the new error path AND the preserved success path to lock down regression protection.
