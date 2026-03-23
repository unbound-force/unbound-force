## 1. Update Indicator Function

- [x] 1.1 Replace `formatIndicator` in `internal/doctor/format.go`: change colored indicators from `✓`/`!`/`✗`/`○` to `✅`/`⚠️ `/`❌`/`⊘`. Update color codes from (2,3,1,8) to (10,11,9,241). Keep `[PASS]`/`[WARN]`/`[FAIL]`/`[INFO]` plain-text fallback unchanged.

## 2. Update Title and Styles

- [x] 2.1 Add `titleStyle` (bold, color 212) and `boxStyle` (rounded border, border color 63, padding 0,1) to `FormatText` in `internal/doctor/format.go`
- [x] 2.2 Replace the header output from `"Unbound Force Doctor\n===================="` to `titleStyle.Render("🩺 Unbound Force Doctor")` in `internal/doctor/format.go`

## 3. Update Fix Hints

- [x] 3.1 Change fix hint formatting in `FormatText` from `"                     Install: %s"` to `dimStyle.Render("     Fix: " + hint)` in `internal/doctor/format.go`. Apply same pattern to `InstallURL` lines.

## 4. Add Boxed Summary

- [x] 4.1 Replace the plain summary line in `FormatText` with a boxed summary using `boxStyle`: `"  ✅ N passed  ⚠️  N warnings  ❌ N failed"` inside the box. After the box, add contextual message: if 0 failures and 0 warnings show `passStyle.Render("🎉 Everything looks good!")`, if failures show `dimStyle.Render("  Run 'uf doctor' after fixes.")`, if warnings only show `dimStyle.Render("  All critical checks passed.")`

## 5. Update Tests

- [x] 5.1 Update `internal/doctor/format_test.go` (or equivalent test assertions in `doctor_test.go`): change expected indicators from `✓`/`✗`/`!`/`○` to `✅`/`❌`/`⚠️`/`⊘` in colored-output assertions. Update title assertion from `"Unbound Force Doctor"` to include `🩺`. Update summary assertion to check for boxed format.

## 6. Verify

- [x] 6.1 Run `go build ./...` to verify compilation
- [x] 6.2 Run `go test -race -count=1 ./internal/doctor/...` to verify doctor tests pass
- [x] 6.3 Run `go test -race -count=1 ./...` to verify full test suite passes
- [x] 6.4 Verify constitution alignment: Observable Quality (JSON unchanged), Testability (format tests updated)
