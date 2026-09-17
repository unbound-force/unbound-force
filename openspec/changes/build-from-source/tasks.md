<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Add Build from Source subsections

- [x] 1.1 Add a "### Build from Source" subsection to `README.md`
  in the `## Getting Started` section after the Fedora/RHEL
  subsection (line ~44) and before the QUICKSTART.md reference
  paragraph. Include: prerequisites (Go 1.25+, make), clone
  command, `make install`, note that it creates both
  `unbound-force` and `uf` symlink in `$GOPATH/bin`, and
  verification (`uf version`). Note that `$GOPATH/bin` must be
  on `$PATH`.
- [x] 1.2 [P] Add a "### Build from Source" subsection to
  `QUICKSTART.md` in the `## Install` section after the
  "Fedora / RHEL (dnf -- minimal)" subsection (line ~104) and
  before "## For Project Maintainers". Include the same content
  as the README subsection.

## 2. Verification

- [x] 2.1 Run `make install` from repo root and verify `uf version`
  outputs version information.
- [x] 2.2 Verify consistency: compare the new subsection structure
  against gaze (`README.md:44-50`), dewey (`README.md:184-190`),
  and replicator (`README.md:42-50`) build-from-source sections --
  same pattern: heading, code block with clone + build, prereqs
  note.
- [x] 2.3 Verify Composability First alignment: the documented
  instructions allow building unbound-force independently without
  Homebrew, dnf, or any external distribution channel.
<!-- scaffolded by uf vdev -->
