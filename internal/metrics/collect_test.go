package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --------------------------------------------------------------------------
// stubGHRunner — dispatch-capable GH CLI stub
// --------------------------------------------------------------------------

// stubGHRunner maps gh command prefixes to fixture responses. When Run is
// called, the stub joins the args, then iterates the dispatch table looking
// for a matching prefix. The first match wins. If no match is found, a
// fallback response (or error) is returned.
type stubGHRunner struct {
	// dispatch maps a command-arg prefix to a response pair.
	// The key is matched via strings.HasPrefix on the joined args.
	dispatch map[string]stubResponse

	// fallback is returned when no dispatch entry matches.
	fallback stubResponse
}

type stubResponse struct {
	out []byte
	err error
}

func (s *stubGHRunner) Run(args ...string) ([]byte, error) {
	joined := strings.Join(args, " ")
	for prefix, resp := range s.dispatch {
		if strings.HasPrefix(joined, prefix) {
			return resp.out, resp.err
		}
	}
	return s.fallback.out, s.fallback.err
}

// --------------------------------------------------------------------------
// ParsePeriod tests
// --------------------------------------------------------------------------

func TestParsePeriod_ValidInputs(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"30d", 30 * 24 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
		{"2w", 2 * 7 * 24 * time.Hour},
		{"1h30m", 90 * time.Minute},
		{"24h", 24 * time.Hour},
		{"1d", 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePeriod(tt.input)
			if err != nil {
				t.Fatalf("ParsePeriod(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParsePeriod(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePeriod_InvalidInputs(t *testing.T) {
	tests := []struct {
		input string
	}{
		{""},
		{"abc"},
		{"d"},
		{"w"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			_, err := ParsePeriod(tt.input)
			if err == nil {
				t.Errorf("ParsePeriod(%q) expected error, got nil", tt.input)
			}
		})
	}
}

// --------------------------------------------------------------------------
// CollectGitHub tests
// --------------------------------------------------------------------------

func TestCollectGitHub_SuccessfulCollection(t *testing.T) {
	now := time.Now().UTC()
	created := now.Add(-48 * time.Hour).Format(time.RFC3339)
	merged := now.Add(-24 * time.Hour).Format(time.RFC3339)

	prJSON := fmt.Sprintf(`[
		{"number":1,"title":"feat: add widget","state":"closed","createdAt":%q,"mergedAt":%q,"closedAt":%q},
		{"number":2,"title":"fix: typo","state":"open","createdAt":%q,"mergedAt":null,"closedAt":null}
	]`, created, merged, merged, created)

	ciJSON := `[
		{"status":"completed","conclusion":"success","createdAt":"2026-03-18T10:00:00Z"},
		{"status":"completed","conclusion":"failure","createdAt":"2026-03-18T11:00:00Z"},
		{"status":"completed","conclusion":"success","createdAt":"2026-03-18T12:00:00Z"}
	]`

	issueJSON := `[
		{"number":10,"state":"open","createdAt":"2026-03-18T00:00:00Z","closedAt":null},
		{"number":11,"state":"OPEN","createdAt":"2026-03-18T00:00:00Z","closedAt":null},
		{"number":12,"state":"closed","createdAt":"2026-03-17T00:00:00Z","closedAt":"2026-03-19T00:00:00Z"}
	]`

	runner := &stubGHRunner{
		dispatch: map[string]stubResponse{
			"api repos/test-org/test-repo/pulls":  {out: []byte(prJSON)},
			"run list":                            {out: []byte(ciJSON)},
			"api repos/test-org/test-repo/issues": {out: []byte(issueJSON)},
		},
	}

	coll, err := CollectGitHub(runner, "test-org/test-repo", 30*24*time.Hour)
	if err != nil {
		t.Fatalf("CollectGitHub returned error: %v", err)
	}
	if coll == nil {
		t.Fatal("CollectGitHub returned nil collection")
	}
	if coll.Source != "github" {
		t.Errorf("Source = %q, want %q", coll.Source, "github")
	}

	// Verify PR metrics
	prCount, ok := coll.RawData["pr_count"]
	if !ok {
		t.Fatal("missing pr_count in RawData")
	}
	prCountInt, isInt := prCount.(int)
	if !isInt || prCountInt != 2 {
		t.Errorf("pr_count = %v, want 2", prCount)
	}

	// Verify merge time is computed (one PR was merged)
	avgMerge, ok := coll.RawData["avg_merge_hours"]
	if !ok {
		t.Fatal("missing avg_merge_hours in RawData")
	}
	avgMergeF, isFloat := avgMerge.(float64)
	if !isFloat || avgMergeF <= 0 {
		t.Errorf("avg_merge_hours = %v, want > 0", avgMerge)
	}

	// Verify CI metrics
	ciRuns, ok := coll.RawData["ci_runs"]
	if !ok {
		t.Fatal("missing ci_runs in RawData")
	}
	ciRunsInt, isInt2 := ciRuns.(int)
	if !isInt2 || ciRunsInt != 3 {
		t.Errorf("ci_runs = %v, want 3", ciRuns)
	}
	ciPassRate, ok := coll.RawData["ci_pass_rate"]
	if !ok {
		t.Fatal("missing ci_pass_rate in RawData")
	}
	// 2 success out of 3 = 66.67%
	wantRate := float64(2) / float64(3) * 100
	ciPassRateF, isFloat2 := ciPassRate.(float64)
	if !isFloat2 || ciPassRateF != wantRate {
		t.Errorf("ci_pass_rate = %v, want %v", ciPassRate, wantRate)
	}

	// Verify issue metrics
	issuesOpened, ok := coll.RawData["issues_opened"]
	if !ok {
		t.Fatal("missing issues_opened in RawData")
	}
	issuesOpenedInt, isInt3 := issuesOpened.(int)
	if !isInt3 || issuesOpenedInt != 2 {
		t.Errorf("issues_opened = %v, want 2", issuesOpened)
	}
	issuesClosed, ok := coll.RawData["issues_closed"]
	if !ok {
		t.Fatal("missing issues_closed in RawData")
	}
	issuesClosedInt, isInt4 := issuesClosed.(int)
	if !isInt4 || issuesClosedInt != 1 {
		t.Errorf("issues_closed = %v, want 1", issuesClosed)
	}

	// Verify total data points = PRs + CI runs + issues
	wantDP := 2 + 3 + 3
	if coll.DataPoints != wantDP {
		t.Errorf("DataPoints = %d, want %d", coll.DataPoints, wantDP)
	}
}

func TestCollectGitHub_RunnerError(t *testing.T) {
	runner := &stubGHRunner{
		dispatch: map[string]stubResponse{
			"api repos/": {err: fmt.Errorf("gh auth login required")},
		},
	}

	coll, err := CollectGitHub(runner, "test-org/repo", 7*24*time.Hour)
	if err == nil {
		t.Fatal("expected error from GH runner, got nil")
	}
	if coll != nil {
		t.Errorf("expected nil collection on error, got %+v", coll)
	}
	if !strings.Contains(err.Error(), "collect PRs") {
		t.Errorf("error should mention 'collect PRs', got: %v", err)
	}
}

func TestCollectGitHub_EmptyResponses(t *testing.T) {
	runner := &stubGHRunner{
		dispatch: map[string]stubResponse{
			"api repos/test-org/repo/pulls":  {out: []byte(`[]`)},
			"run list":                       {out: []byte(`[]`)},
			"api repos/test-org/repo/issues": {out: []byte(`[]`)},
		},
	}

	coll, err := CollectGitHub(runner, "test-org/repo", 30*24*time.Hour)
	if err != nil {
		t.Fatalf("CollectGitHub returned error: %v", err)
	}
	if coll == nil {
		t.Fatal("expected non-nil collection even with empty data")
	}
	prCountVal, prOk := coll.RawData["pr_count"].(int)
	if !prOk || prCountVal != 0 {
		t.Errorf("pr_count = %v, want 0", coll.RawData["pr_count"])
	}
	ciRunsVal, ciOk := coll.RawData["ci_runs"].(int)
	_ = ciOk
	if ciRunsVal != 0 {
		t.Errorf("ci_runs = %v, want 0", coll.RawData["ci_runs"])
	}
	if coll.RawData["issue_count"].(int) != 0 {
		t.Errorf("issue_count = %v, want 0", coll.RawData["issue_count"])
	}

	// No merged PRs → avg_merge_hours should be absent
	if _, ok := coll.RawData["avg_merge_hours"]; ok {
		t.Errorf("avg_merge_hours should be absent when no PRs are merged")
	}
	// Empty CI runs → ci_pass_rate should be absent
	if _, ok := coll.RawData["ci_pass_rate"]; ok {
		t.Errorf("ci_pass_rate should be absent when ci_runs is 0")
	}

	if coll.DataPoints != 0 {
		t.Errorf("DataPoints = %d, want 0", coll.DataPoints)
	}
}

func TestCollectGitHub_CIErrorIsNonFatal(t *testing.T) {
	// PR collection succeeds, but CI collection fails.
	// CI errors are non-fatal (the function checks err == nil).
	runner := &stubGHRunner{
		dispatch: map[string]stubResponse{
			"api repos/org/repo/pulls":  {out: []byte(`[]`)},
			"run list":                  {err: fmt.Errorf("gh run list failed")},
			"api repos/org/repo/issues": {out: []byte(`[]`)},
		},
	}

	coll, err := CollectGitHub(runner, "org/repo", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("CollectGitHub returned error: %v", err)
	}
	if coll == nil {
		t.Fatal("expected non-nil collection")
	}
	// CI data should be absent
	if _, ok := coll.RawData["ci_runs"]; ok {
		t.Errorf("ci_runs should not be present when CI collection failed")
	}
}

func TestCollectGitHub_IssueErrorIsNonFatal(t *testing.T) {
	runner := &stubGHRunner{
		dispatch: map[string]stubResponse{
			"api repos/org/repo/pulls":  {out: []byte(`[]`)},
			"run list":                  {out: []byte(`[]`)},
			"api repos/org/repo/issues": {err: fmt.Errorf("403 forbidden")},
		},
	}

	coll, err := CollectGitHub(runner, "org/repo", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("CollectGitHub returned error: %v", err)
	}
	if coll == nil {
		t.Fatal("expected non-nil collection")
	}
	if _, ok := coll.RawData["issue_count"]; ok {
		t.Errorf("issue_count should not be present when issue collection failed")
	}
}

// --------------------------------------------------------------------------
// CollectMutiMind tests
// --------------------------------------------------------------------------

// writeArtifactFile writes a JSON artifact envelope to dir with the given
// artifact type and payload.
func writeArtifactFile(t *testing.T, dir, id, artifactType string, payload interface{}) {
	t.Helper()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	envelope := map[string]interface{}{
		"hero":           "muti-mind",
		"version":        "1.0.0",
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"artifact_type":  artifactType,
		"schema_version": "1.0.0",
		"payload":        json.RawMessage(payloadBytes),
	}

	envelopeBytes, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	filename := fmt.Sprintf("%s-%s.json", id, artifactType)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), envelopeBytes, 0644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
}

func TestCollectMutiMind_ValidArtifacts(t *testing.T) {
	dir := t.TempDir()

	// Write backlog items
	writeArtifactFile(t, dir, "BI-001", "backlog-item", map[string]interface{}{
		"id":       "BI-001",
		"title":    "Add widget",
		"priority": "P1",
		"status":   "ready",
	})
	writeArtifactFile(t, dir, "BI-002", "backlog-item", map[string]interface{}{
		"id":       "BI-002",
		"title":    "Fix bug",
		"priority": "P2",
		"status":   "done",
	})

	// Write acceptance decisions
	writeArtifactFile(t, dir, "AD-001", "acceptance-decision", map[string]interface{}{
		"item_id":  "BI-001",
		"decision": "accept",
	})
	writeArtifactFile(t, dir, "AD-002", "acceptance-decision", map[string]interface{}{
		"item_id":  "BI-002",
		"decision": "reject",
	})

	since := time.Now().Add(-24 * time.Hour)
	coll, err := CollectMutiMind(dir, since)
	if err != nil {
		t.Fatalf("CollectMutiMind returned error: %v", err)
	}
	if coll == nil {
		t.Fatal("expected non-nil collection")
	}
	if coll.Source != "muti-mind" {
		t.Errorf("Source = %q, want %q", coll.Source, "muti-mind")
	}

	// Verify backlog metrics
	bSize, ok := coll.RawData["backlog_size"]
	if !ok {
		t.Fatal("missing backlog_size")
	}
	if bSize.(int) != 2 {
		t.Errorf("backlog_size = %v, want 2", bSize)
	}

	// Verify acceptance metrics
	aCount, ok := coll.RawData["acceptance_count"]
	if !ok {
		t.Fatal("missing acceptance_count")
	}
	if aCount.(int) != 2 {
		t.Errorf("acceptance_count = %v, want 2", aCount)
	}

	// Verify acceptance rate (1 accept out of 2 = 0.5)
	aRate, ok := coll.RawData["acceptance_rate"]
	if !ok {
		t.Fatal("missing acceptance_rate")
	}
	if aRate.(float64) != 0.5 {
		t.Errorf("acceptance_rate = %v, want 0.5", aRate)
	}

	// DataPoints = backlog items + decisions
	if coll.DataPoints != 4 {
		t.Errorf("DataPoints = %d, want 4", coll.DataPoints)
	}
}

func TestCollectMutiMind_NoArtifacts(t *testing.T) {
	dir := t.TempDir()

	since := time.Now().Add(-24 * time.Hour)
	coll, err := CollectMutiMind(dir, since)
	if err != nil {
		t.Fatalf("CollectMutiMind returned error: %v", err)
	}
	if coll != nil {
		t.Errorf("expected nil collection when no artifacts, got %+v", coll)
	}
}

func TestCollectMutiMind_MalformedArtifact(t *testing.T) {
	dir := t.TempDir()

	// Write a valid backlog item
	writeArtifactFile(t, dir, "BI-001", "backlog-item", map[string]interface{}{
		"id":    "BI-001",
		"title": "Good item",
	})

	// Write a file with a valid envelope structure but payload that is
	// valid JSON at the envelope level (a string) but cannot be
	// unmarshalled into map[string]interface{}.
	// The payload is a JSON string "bad" — valid JSON, but CollectMutiMind
	// tries json.Unmarshal into map[string]interface{} which fails.
	malformed := `{
  "hero": "muti-mind",
  "version": "1.0.0",
  "timestamp": "2026-03-20T00:00:00Z",
  "artifact_type": "backlog-item",
  "schema_version": "1.0.0",
  "payload": "this-is-a-string-not-an-object"
}`
	if err := os.WriteFile(filepath.Join(dir, "BI-BAD-backlog-item.json"), []byte(malformed), 0644); err != nil {
		t.Fatalf("write malformed: %v", err)
	}

	since := time.Now().Add(-24 * time.Hour)
	coll, err := CollectMutiMind(dir, since)
	if err != nil {
		t.Fatalf("CollectMutiMind returned error: %v", err)
	}
	if coll == nil {
		t.Fatal("expected non-nil collection (valid item exists)")
	}

	// Only the valid item should be counted; malformed is skipped
	bSize, ok := coll.RawData["backlog_size"]
	if !ok {
		t.Fatal("missing backlog_size")
	}
	if bSize.(int) != 1 {
		t.Errorf("backlog_size = %v, want 1 (malformed should be skipped)", bSize)
	}
}

func TestCollectMutiMind_OnlyBacklogItems(t *testing.T) {
	dir := t.TempDir()

	writeArtifactFile(t, dir, "BI-001", "backlog-item", map[string]interface{}{
		"id":    "BI-001",
		"title": "Solo item",
	})

	since := time.Now().Add(-24 * time.Hour)
	coll, err := CollectMutiMind(dir, since)
	if err != nil {
		t.Fatalf("CollectMutiMind returned error: %v", err)
	}
	if coll == nil {
		t.Fatal("expected non-nil collection")
	}
	if coll.RawData["backlog_size"].(int) != 1 {
		t.Errorf("backlog_size = %v, want 1", coll.RawData["backlog_size"])
	}
	if coll.RawData["acceptance_count"].(int) != 0 {
		t.Errorf("acceptance_count = %v, want 0", coll.RawData["acceptance_count"])
	}
	// No decisions → acceptance_rate should be absent
	if _, ok := coll.RawData["acceptance_rate"]; ok {
		t.Errorf("acceptance_rate should be absent when no decisions exist")
	}
}

// --------------------------------------------------------------------------
// Collector.Collect tests
// --------------------------------------------------------------------------

func TestCollect_AllSourcesSucceed(t *testing.T) {
	storeDir := t.TempDir()
	store := NewStore(storeDir)
	artifactDir := t.TempDir()

	// Set up muti-mind artifacts so the real CollectMutiMind finds them
	writeArtifactFile(t, artifactDir, "BI-001", "backlog-item", map[string]interface{}{
		"id": "BI-001", "title": "Item",
	})

	// Set up gaze artifacts
	writeArtifactFile(t, artifactDir, "QR-001", "quality-report", map[string]interface{}{
		"score": 95,
	})

	// Set up divisor artifacts
	writeArtifactFile(t, artifactDir, "RV-001", "review-verdict", map[string]interface{}{
		"verdict": "approved",
	})

	runner := &stubGHRunner{
		dispatch: map[string]stubResponse{
			"api repos/org/repo/pulls":  {out: []byte(`[]`)},
			"run list":                  {out: []byte(`[]`)},
			"api repos/org/repo/issues": {out: []byte(`[]`)},
		},
	}

	var buf bytes.Buffer
	c := &Collector{
		GHRunner:    runner,
		ArtifactDir: artifactDir,
		Store:       store,
		Stdout:      &buf,
		Now:         func() time.Time { return time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC) },
	}

	err := c.Collect([]string{"all"}, "org/repo", 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "collected from 4/4 sources") {
		t.Errorf("expected 4/4 sources collected, output:\n%s", output)
	}
}

func TestCollect_OneSourceFails_GracefulDegradation(t *testing.T) {
	storeDir := t.TempDir()
	store := NewStore(storeDir)
	artifactDir := t.TempDir()

	// GitHub will fail
	runner := &stubGHRunner{
		dispatch: map[string]stubResponse{
			"api repos/": {err: fmt.Errorf("auth required")},
		},
	}

	// At least muti-mind will have data
	writeArtifactFile(t, artifactDir, "BI-001", "backlog-item", map[string]interface{}{
		"id": "BI-001", "title": "Item",
	})

	var buf bytes.Buffer
	c := &Collector{
		GHRunner:    runner,
		ArtifactDir: artifactDir,
		Store:       store,
		Stdout:      &buf,
		Now:         func() time.Time { return time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC) },
	}

	err := c.Collect([]string{"github", "muti-mind"}, "org/repo", 30*24*time.Hour)
	// Should NOT return error because at least one source succeeded
	if err != nil {
		t.Fatalf("Collect returned error (expected graceful degradation): %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "collected from 1/2 sources") {
		t.Errorf("expected 1/2 sources collected, output:\n%s", output)
	}
	if !strings.Contains(output, "auth required") {
		t.Errorf("expected error message in output, got:\n%s", output)
	}
}

func TestCollect_NoRepoSpecified_GitHubSkipped(t *testing.T) {
	storeDir := t.TempDir()
	store := NewStore(storeDir)
	artifactDir := t.TempDir()

	runner := &stubGHRunner{}

	var buf bytes.Buffer
	c := &Collector{
		GHRunner:    runner,
		ArtifactDir: artifactDir,
		Store:       store,
		Stdout:      &buf,
		Now:         func() time.Time { return time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC) },
	}

	err := c.Collect([]string{"github"}, "", 30*24*time.Hour)
	// No error because github is skipped (not an error, just skipped)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "no repository specified") {
		t.Errorf("expected 'no repository specified' in output, got:\n%s", output)
	}
}

func TestCollect_AllSourcesFail_ReturnsError(t *testing.T) {
	storeDir := t.TempDir()
	store := NewStore(storeDir)
	artifactDir := t.TempDir()

	runner := &stubGHRunner{
		dispatch: map[string]stubResponse{
			"api repos/": {err: fmt.Errorf("network down")},
		},
	}

	var buf bytes.Buffer
	c := &Collector{
		GHRunner:    runner,
		ArtifactDir: artifactDir,
		Store:       store,
		Stdout:      &buf,
		Now:         func() time.Time { return time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC) },
	}

	err := c.Collect([]string{"github"}, "org/repo", 30*24*time.Hour)
	if err == nil {
		t.Fatal("expected error when all sources fail")
	}
	if !strings.Contains(err.Error(), "no sources collected") {
		t.Errorf("expected 'no sources collected' in error, got: %v", err)
	}
}

func TestCollect_UnknownSource(t *testing.T) {
	storeDir := t.TempDir()
	store := NewStore(storeDir)

	var buf bytes.Buffer
	c := &Collector{
		ArtifactDir: t.TempDir(),
		Store:       store,
		Stdout:      &buf,
		Now:         func() time.Time { return time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC) },
	}

	err := c.Collect([]string{"bogus"}, "", 24*time.Hour)
	// Unknown source is skipped, 0/1 collected but no lastErr so no error
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "unknown source") {
		t.Errorf("expected 'unknown source' in output, got:\n%s", output)
	}
}

func TestCollect_AllExpands(t *testing.T) {
	storeDir := t.TempDir()
	store := NewStore(storeDir)
	artifactDir := t.TempDir()

	runner := &stubGHRunner{
		dispatch: map[string]stubResponse{
			"api repos/": {out: []byte(`[]`)},
			"run list":   {out: []byte(`[]`)},
		},
	}

	var buf bytes.Buffer
	c := &Collector{
		GHRunner:    runner,
		ArtifactDir: artifactDir,
		Store:       store,
		Stdout:      &buf,
		Now:         func() time.Time { return time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC) },
	}

	err := c.Collect([]string{"all"}, "org/repo", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	// "all" expands to 4 sources: github, gaze, divisor, muti-mind
	output := buf.String()
	if !strings.Contains(output, "/4 sources") {
		t.Errorf("expected /4 sources in output (all expands to 4), got:\n%s", output)
	}
}

func TestCollect_SourceReturnsNil(t *testing.T) {
	// When a collector returns nil (no artifacts), it should report "no artifacts found"
	storeDir := t.TempDir()
	store := NewStore(storeDir)
	artifactDir := t.TempDir() // empty dir — gaze/divisor/muti-mind all return nil

	var buf bytes.Buffer
	c := &Collector{
		ArtifactDir: artifactDir,
		Store:       store,
		Stdout:      &buf,
		Now:         func() time.Time { return time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC) },
	}

	err := c.Collect([]string{"gaze"}, "", 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "no artifacts found") {
		t.Errorf("expected 'no artifacts found' for empty source, got:\n%s", output)
	}
}
