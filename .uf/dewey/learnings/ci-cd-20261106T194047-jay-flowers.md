---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-11-06T19:40:47Z
identity: ci-cd-20261106T194047-jay-flowers
tier: draft
---

actions/setup-go with go-version-file sets GOTOOLCHAIN=local, freezing the job to the go.mod toolchain (e.g. go 1.25.0). A later `go install tool@vX` whose module requires a newer Go (osv-scanner v2.7.0 requires go >= 1.27.0) then fails with a toolchain-resolution error that does not name the cause. Exporting GOTOOLCHAIN=auto in the same step lets the bootstrap Go fetch the required toolchain (downloads are authenticated against the Go checksum database), and documenting the version skew beside the export keeps the workaround discoverable. Verified in ci_security.yml osv_plugin_transitive fix commit 24df5a68.
