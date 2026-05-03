## 1. Update Task Template

- [x] 1.1 Update `openspec/schemas/unbound-force/templates/tasks.md`:
  add `[P]` marker examples showing both parallel and
  sequential tasks in the same group, with a comment
  explaining the `[P]` convention (tasks touching
  different files with no dependencies)

## 2. Update Schema Instructions

- [x] 2.1 Update `openspec/schemas/unbound-force/schema.yaml`:
  expand the `tasks` artifact `instruction` field to
  include guidance on when to add `[P]` markers
  (different files, no inter-task dependencies) and
  when NOT to add them (same file, sequential
  dependency)

## 3. Documentation

- [x] 3.1 [P] Update AGENTS.md: document the `[P]`
  marker alignment between Speckit and OpenSpec task
  formats in the "Specification Framework" section
- [x] 3.2 [P] Update AGENTS.md "Recent Changes" with
  this change summary

## 4. Verification

- [x] 4.1 Run `go build ./...` — verify clean build
  (no Go changes but confirms no regressions)
- [x] 4.2 Verify the template renders correctly as
  Markdown
<!-- spec-review: passed -->
<!-- code-review: passed -->
<!-- scaffolded by uf vdev -->
