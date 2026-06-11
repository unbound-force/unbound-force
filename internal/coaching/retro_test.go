package coaching

import (
	"testing"
)

func TestRetroStore_SaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewRetroStore(dir)

	record := &RetroRecord{
		Date:         "2026-03-20",
		Participants: []string{"@dev1", "@dev2"},
		DataPresented: map[string]interface{}{
			"velocity": 8.2,
		},
		PatternsIdentified:   []string{"Reviews taking longer"},
		RootCauses:           []string{"Large PRs"},
		ImprovementProposals: []string{"Smaller PRs"},
		ActionItems: []ActionItem{
			{ID: "AI-001", Description: "Split large PRs", Owner: "@dev1", Deadline: "2026-03-27", Status: "pending", RetrospectiveID: "2026-03-20"},
		},
		Notes: "Good discussion.",
	}

	if err := store.SaveRetro(record); err != nil {
		t.Fatalf("SaveRetro: %v", err)
	}

	loaded, err := store.LoadRetro("2026-03-20")
	if err != nil {
		t.Fatalf("LoadRetro: %v", err)
	}
	if loaded.Date != "2026-03-20" {
		t.Errorf("Date = %q", loaded.Date)
	}
	if len(loaded.ActionItems) != 1 {
		t.Fatalf("ActionItems count = %d", len(loaded.ActionItems))
	}
	if loaded.ActionItems[0].ID != "AI-001" {
		t.Errorf("ActionItem ID = %q", loaded.ActionItems[0].ID)
	}
}

func TestRetroStore_ListRetros(t *testing.T) {
	dir := t.TempDir()
	store := NewRetroStore(dir)

	r1 := &RetroRecord{Date: "2026-03-13"}
	r2 := &RetroRecord{Date: "2026-03-20"}
	if err := store.SaveRetro(r1); err != nil {
		t.Fatalf("SaveRetro r1: %v", err)
	}
	if err := store.SaveRetro(r2); err != nil {
		t.Fatalf("SaveRetro r2: %v", err)
	}

	retros, err := store.ListRetros()
	if err != nil {
		t.Fatalf("ListRetros: %v", err)
	}
	if len(retros) != 2 {
		t.Fatalf("got %d retros, want 2", len(retros))
	}
	if retros[0].Date != "2026-03-20" {
		t.Errorf("first retro = %q, want 2026-03-20 (descending order)", retros[0].Date)
	}
}

func TestRetroStore_StartRetro(t *testing.T) {
	dir := t.TempDir()
	store := NewRetroStore(dir)

	metricsData := map[string]interface{}{
		"velocity":   8.2,
		"cycle_time": 3.5,
	}

	record, err := store.StartRetro("2026-04-01", metricsData)
	if err != nil {
		t.Fatalf("StartRetro: %v", err)
	}
	if record == nil {
		t.Fatal("StartRetro returned nil record")
	}
	if record.Date != "2026-04-01" {
		t.Errorf("Date = %q, want 2026-04-01", record.Date)
	}
	if record.DataPresented == nil {
		t.Fatal("DataPresented is nil")
	}
	if v, ok := record.DataPresented["velocity"]; !ok || v != 8.2 {
		t.Errorf("DataPresented[velocity] = %v, want 8.2", v)
	}
	if v, ok := record.DataPresented["cycle_time"]; !ok || v != 3.5 {
		t.Errorf("DataPresented[cycle_time] = %v, want 3.5", v)
	}
}

func TestRetroStore_StartRetro_NilMetrics(t *testing.T) {
	dir := t.TempDir()
	store := NewRetroStore(dir)

	record, err := store.StartRetro("2026-04-02", nil)
	if err != nil {
		t.Fatalf("StartRetro with nil metrics: %v", err)
	}
	if record == nil {
		t.Fatal("StartRetro returned nil record")
	}
	if record.Date != "2026-04-02" {
		t.Errorf("Date = %q, want 2026-04-02", record.Date)
	}
}

func TestNextActionID(t *testing.T) {
	retros := []RetroRecord{
		{ActionItems: []ActionItem{{ID: "AI-001"}, {ID: "AI-003"}}},
		{ActionItems: []ActionItem{{ID: "AI-002"}}},
	}
	next := NextActionID(retros)
	if next != "AI-004" {
		t.Errorf("NextActionID = %q, want AI-004", next)
	}
}

func TestReviewPreviousActions_StaleDetection(t *testing.T) {
	retros := []RetroRecord{
		{
			ActionItems: []ActionItem{
				{ID: "AI-001", Status: "pending", Deadline: "2020-01-01"},
				{ID: "AI-002", Status: "completed", Deadline: "2020-01-01"},
				{ID: "AI-003", Status: "pending", Deadline: "2099-12-31"},
			},
		},
	}
	items := ReviewPreviousActions(retros)
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2 (excluding completed)", len(items))
	}
	staleFound := false
	for _, ai := range items {
		if ai.ID == "AI-001" && ai.Status == "stale" {
			staleFound = true
		}
	}
	if !staleFound {
		t.Error("AI-001 should be marked stale (past deadline)")
	}
}
