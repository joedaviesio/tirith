package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func sampleCall(id, model, tag, user, env string, input, output, cost int) *APICall {
	return &APICall{
		ID:           id,
		Timestamp:    time.Now().UTC(),
		Provider:     "anthropic",
		Model:        model,
		InputTokens:  input,
		OutputTokens: output,
		CostCents:    cost,
		LatencyMs:    150,
		StatusCode:   200,
		Streaming:    false,
		Tag:          tag,
		UserTag:      user,
		Environment:  env,
		Endpoint:     "/v1/messages",
	}
}

func TestInsertAndGetSummary(t *testing.T) {
	s := testStore(t)

	if err := s.Insert(sampleCall("1", "claude-sonnet-4-6", "feat", "alice", "prod", 100, 50, 5)); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := s.Insert(sampleCall("2", "claude-haiku-4-5-20251001", "feat", "bob", "prod", 200, 80, 2)); err != nil {
		t.Fatalf("insert: %v", err)
	}

	sum, err := s.GetSummary(ReportFilter{})
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if sum.TotalCalls != 2 {
		t.Errorf("TotalCalls = %d, want 2", sum.TotalCalls)
	}
	if sum.TotalCostCents != 7 {
		t.Errorf("TotalCostCents = %d, want 7", sum.TotalCostCents)
	}
	if sum.TotalInput != 300 {
		t.Errorf("TotalInput = %d, want 300", sum.TotalInput)
	}
	if sum.TotalOutput != 130 {
		t.Errorf("TotalOutput = %d, want 130", sum.TotalOutput)
	}
}

func TestGetSummaryEmpty(t *testing.T) {
	s := testStore(t)

	sum, err := s.GetSummary(ReportFilter{})
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if sum.TotalCalls != 0 {
		t.Errorf("TotalCalls = %d, want 0", sum.TotalCalls)
	}
	if sum.TotalCostCents != 0 {
		t.Errorf("TotalCostCents = %d, want 0", sum.TotalCostCents)
	}
}

func TestFilterBySince(t *testing.T) {
	s := testStore(t)

	old := &APICall{
		ID: "old", Timestamp: time.Now().UTC().Add(-48 * time.Hour),
		Provider: "anthropic", Model: "claude-sonnet-4-6",
		InputTokens: 100, OutputTokens: 50, CostCents: 5,
		LatencyMs: 100, StatusCode: 200, Environment: "default",
	}
	recent := sampleCall("new", "claude-sonnet-4-6", "", "", "default", 100, 50, 3)

	_ = s.Insert(old)
	_ = s.Insert(recent)

	sum, err := s.GetSummary(ReportFilter{Since: time.Now().UTC().Add(-1 * time.Hour)})
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if sum.TotalCalls != 1 {
		t.Errorf("TotalCalls = %d, want 1 (only recent)", sum.TotalCalls)
	}
}

func TestFilterByTag(t *testing.T) {
	s := testStore(t)

	_ = s.Insert(sampleCall("1", "claude-sonnet-4-6", "scanner", "alice", "prod", 100, 50, 5))
	_ = s.Insert(sampleCall("2", "claude-sonnet-4-6", "chat", "alice", "prod", 100, 50, 3))

	sum, err := s.GetSummary(ReportFilter{Tag: "scanner"})
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if sum.TotalCalls != 1 {
		t.Errorf("TotalCalls = %d, want 1", sum.TotalCalls)
	}
	if sum.TotalCostCents != 5 {
		t.Errorf("TotalCostCents = %d, want 5", sum.TotalCostCents)
	}
}

func TestFilterByModel(t *testing.T) {
	s := testStore(t)

	_ = s.Insert(sampleCall("1", "claude-sonnet-4-6", "", "", "default", 100, 50, 5))
	_ = s.Insert(sampleCall("2", "claude-haiku-4-5-20251001", "", "", "default", 200, 80, 2))

	sum, err := s.GetSummary(ReportFilter{Model: "claude-haiku-4-5-20251001"})
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if sum.TotalCalls != 1 {
		t.Errorf("TotalCalls = %d, want 1", sum.TotalCalls)
	}
}

func TestGetByModel(t *testing.T) {
	s := testStore(t)

	_ = s.Insert(sampleCall("1", "claude-sonnet-4-6", "", "", "default", 100, 50, 5))
	_ = s.Insert(sampleCall("2", "claude-sonnet-4-6", "", "", "default", 100, 50, 3))
	_ = s.Insert(sampleCall("3", "claude-haiku-4-5-20251001", "", "", "default", 200, 80, 1))

	models, err := s.GetByModel(ReportFilter{})
	if err != nil {
		t.Fatalf("GetByModel: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("len(models) = %d, want 2", len(models))
	}
	// Sorted by cost DESC — sonnet (8) first.
	if models[0].Model != "claude-sonnet-4-6" {
		t.Errorf("first model = %s, want claude-sonnet-4-6", models[0].Model)
	}
	if models[0].TotalCalls != 2 {
		t.Errorf("sonnet calls = %d, want 2", models[0].TotalCalls)
	}
}

func TestGetByTag(t *testing.T) {
	s := testStore(t)

	_ = s.Insert(sampleCall("1", "claude-sonnet-4-6", "scanner", "", "default", 100, 50, 5))
	_ = s.Insert(sampleCall("2", "claude-sonnet-4-6", "scanner", "", "default", 100, 50, 3))
	_ = s.Insert(sampleCall("3", "claude-sonnet-4-6", "chat", "", "default", 100, 50, 1))

	tags, err := s.GetByTag(ReportFilter{})
	if err != nil {
		t.Fatalf("GetByTag: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
	if tags[0].Tag != "scanner" {
		t.Errorf("first tag = %s, want scanner", tags[0].Tag)
	}
	if tags[0].TotalCostCents != 8 {
		t.Errorf("scanner cost = %d, want 8", tags[0].TotalCostCents)
	}
}

func TestGetByUser(t *testing.T) {
	s := testStore(t)

	_ = s.Insert(sampleCall("1", "claude-sonnet-4-6", "", "alice", "default", 100, 50, 5))
	_ = s.Insert(sampleCall("2", "claude-sonnet-4-6", "", "bob", "default", 100, 50, 3))

	users, err := s.GetByUser(ReportFilter{})
	if err != nil {
		t.Fatalf("GetByUser: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("len(users) = %d, want 2", len(users))
	}
}

func TestGetDailySpend(t *testing.T) {
	s := testStore(t)

	_ = s.Insert(sampleCall("1", "claude-sonnet-4-6", "", "", "default", 100, 50, 5))
	_ = s.Insert(sampleCall("2", "claude-sonnet-4-6", "", "", "default", 100, 50, 3))

	daily, err := s.GetDailySpend(ReportFilter{})
	if err != nil {
		t.Fatalf("GetDailySpend: %v", err)
	}
	if len(daily) != 1 {
		t.Fatalf("len(daily) = %d, want 1 (same day)", len(daily))
	}
	if daily[0].TotalCostCents != 8 {
		t.Errorf("daily cost = %d, want 8", daily[0].TotalCostCents)
	}
}

func TestGetDailyByModel(t *testing.T) {
	s := testStore(t)

	_ = s.Insert(sampleCall("1", "claude-sonnet-4-6", "", "", "default", 100, 50, 5))
	_ = s.Insert(sampleCall("2", "claude-haiku-4-5-20251001", "", "", "default", 200, 80, 2))

	daily, err := s.GetDailyByModel(ReportFilter{})
	if err != nil {
		t.Fatalf("GetDailyByModel: %v", err)
	}
	if len(daily) != 2 {
		t.Fatalf("len(daily) = %d, want 2 (two models same day)", len(daily))
	}
}

func TestGetRecentCalls(t *testing.T) {
	s := testStore(t)

	for i := 0; i < 5; i++ {
		c := sampleCall(
			fmt.Sprintf("call-%d", i), "claude-sonnet-4-6", "", "", "default", 100, 50, 5,
		)
		_ = s.Insert(c)
	}

	calls, err := s.GetRecentCalls(3)
	if err != nil {
		t.Fatalf("GetRecentCalls: %v", err)
	}
	if len(calls) != 3 {
		t.Errorf("len(calls) = %d, want 3", len(calls))
	}
}

func TestGetRecentCallsEmpty(t *testing.T) {
	s := testStore(t)

	calls, err := s.GetRecentCalls(10)
	if err != nil {
		t.Fatalf("GetRecentCalls: %v", err)
	}
	if len(calls) != 0 {
		t.Errorf("len(calls) = %d, want 0", len(calls))
	}
}

func TestDuplicateIDRejected(t *testing.T) {
	s := testStore(t)

	c := sampleCall("dup", "claude-sonnet-4-6", "", "", "default", 100, 50, 5)
	if err := s.Insert(c); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if err := s.Insert(c); err == nil {
		t.Error("expected error on duplicate ID, got nil")
	}
}

func TestDatabaseFileCreated(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	s.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}

func TestTimezoneOffset(t *testing.T) {
	s := testStore(t)

	_ = s.Insert(sampleCall("1", "claude-sonnet-4-6", "", "", "default", 100, 50, 5))

	// With a timezone offset, daily spend should still return one row.
	daily, err := s.GetDailySpend(ReportFilter{TZOffset: "+12:00"})
	if err != nil {
		t.Fatalf("GetDailySpend with TZ: %v", err)
	}
	if len(daily) == 0 {
		t.Error("expected at least 1 daily row with timezone offset")
	}
}

