package pricing

import (
	"testing"
)

func testEngine(t *testing.T) *Engine {
	t.Helper()
	e, err := New()
	if err != nil {
		t.Fatalf("creating pricing engine: %v", err)
	}
	return e
}

func TestCalculateCostCents_Anthropic(t *testing.T) {
	e := testEngine(t)

	tests := []struct {
		name     string
		model    string
		usage    TokenUsage
		wantMin  int
		wantMax  int
	}{
		{
			name:  "sonnet basic",
			model: "claude-sonnet-4-6",
			usage: TokenUsage{InputTokens: 1000, OutputTokens: 500},
			// $3/M input = $0.003, $15/M output = $0.0075 → $0.0105 = 1.05¢ → rounds to 2¢
			wantMin: 1, wantMax: 3,
		},
		{
			name:  "haiku cheap",
			model: "claude-haiku-4-5-20251001",
			usage: TokenUsage{InputTokens: 100, OutputTokens: 50},
			// Very small usage → rounds up to at least 1¢
			wantMin: 1, wantMax: 1,
		},
		{
			name:  "zero tokens",
			model: "claude-sonnet-4-6",
			usage: TokenUsage{},
			// No tokens = 0 cost
			wantMin: 0, wantMax: 0,
		},
		{
			name:  "with cache tokens",
			model: "claude-sonnet-4-6",
			usage: TokenUsage{
				InputTokens:      1000,
				OutputTokens:     500,
				CacheReadTokens:  2000,
				CacheWriteTokens: 500,
			},
			wantMin: 1, wantMax: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, err := e.CalculateCostCents("anthropic", tt.model, tt.usage)
			if err != nil {
				t.Fatalf("CalculateCostCents: %v", err)
			}
			if cost < tt.wantMin || cost > tt.wantMax {
				t.Errorf("cost = %d, want [%d, %d]", cost, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateCostCents_UnknownProvider(t *testing.T) {
	e := testEngine(t)

	_, err := e.CalculateCostCents("unknown-provider", "some-model", TokenUsage{InputTokens: 100})
	if err == nil {
		t.Error("expected error for unknown provider, got nil")
	}
}

func TestCalculateCostCents_UnknownModel(t *testing.T) {
	e := testEngine(t)

	_, err := e.CalculateCostCents("anthropic", "nonexistent-model", TokenUsage{InputTokens: 100})
	if err == nil {
		t.Error("expected error for unknown model, got nil")
	}
}

func TestCalculateCostCents_ModelPrefixMatch(t *testing.T) {
	e := testEngine(t)

	// A versioned model name should match the base model via prefix matching.
	cost, err := e.CalculateCostCents("anthropic", "claude-sonnet-4-6-20260301", TokenUsage{
		InputTokens: 1000, OutputTokens: 500,
	})
	if err != nil {
		t.Fatalf("prefix match failed: %v", err)
	}
	if cost < 1 {
		t.Errorf("cost = %d, want >= 1", cost)
	}
}

func TestListModels(t *testing.T) {
	e := testEngine(t)

	models := e.ListModels("anthropic")
	if len(models) == 0 {
		t.Error("ListModels returned empty for anthropic")
	}

	// Unknown provider should return nil.
	unknown := e.ListModels("unknown")
	if unknown != nil {
		t.Errorf("ListModels for unknown = %v, want nil", unknown)
	}
}

func TestCostIsAlwaysNonNegative(t *testing.T) {
	e := testEngine(t)

	// Even with weird inputs, cost should never be negative.
	cost, err := e.CalculateCostCents("anthropic", "claude-sonnet-4-6", TokenUsage{
		InputTokens:  0,
		OutputTokens: 0,
	})
	if err != nil {
		t.Fatalf("CalculateCostCents: %v", err)
	}
	if cost < 0 {
		t.Errorf("cost = %d, want >= 0", cost)
	}
}
