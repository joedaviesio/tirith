package pricing

import (
	"embed"
	"fmt"
	"math"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed pricing.yaml
var embeddedPricing embed.FS

type PricingData struct {
	Version   string                      `yaml:"version"`
	Providers map[string]ProviderPricing  `yaml:"providers"`
}

type ProviderPricing struct {
	Models map[string]ModelPricing `yaml:"models"`
}

type ModelPricing struct {
	InputPerMillion      float64 `yaml:"input_per_million"`
	OutputPerMillion     float64 `yaml:"output_per_million"`
	CacheReadPerMillion  float64 `yaml:"cache_read_per_million"`
	CacheWritePerMillion float64 `yaml:"cache_write_per_million"`
}

type Engine struct {
	data *PricingData
}

type TokenUsage struct {
	InputTokens      int
	OutputTokens     int
	CacheReadTokens  int
	CacheWriteTokens int
}

func New() (*Engine, error) {
	data, err := embeddedPricing.ReadFile("pricing.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading embedded pricing: %w", err)
	}

	var pd PricingData
	if err := yaml.Unmarshal(data, &pd); err != nil {
		return nil, fmt.Errorf("parsing pricing data: %w", err)
	}

	return &Engine{data: &pd}, nil
}

// CalculateCostCents returns the cost in cents (integer) for a given provider, model, and usage.
func (e *Engine) CalculateCostCents(provider, model string, usage TokenUsage) (int, error) {
	providerData, ok := e.data.Providers[strings.ToLower(provider)]
	if !ok {
		return 0, fmt.Errorf("unknown provider: %s", provider)
	}

	modelData, ok := providerData.Models[model]
	if !ok {
		// Try without version suffix (e.g., "claude-sonnet-4-6-20260301" -> "claude-sonnet-4-6")
		for name, md := range providerData.Models {
			if strings.HasPrefix(model, name) {
				modelData = md
				ok = true
				break
			}
		}
		if !ok {
			return 0, fmt.Errorf("unknown model %s for provider %s", model, provider)
		}
	}

	costDollars := float64(usage.InputTokens)*modelData.InputPerMillion/1_000_000 +
		float64(usage.OutputTokens)*modelData.OutputPerMillion/1_000_000 +
		float64(usage.CacheReadTokens)*modelData.CacheReadPerMillion/1_000_000 +
		float64(usage.CacheWriteTokens)*modelData.CacheWritePerMillion/1_000_000

	// Convert to cents and round up to avoid losing fractions.
	costCents := int(math.Ceil(costDollars * 100))
	return costCents, nil
}

func (e *Engine) ListModels(provider string) []string {
	providerData, ok := e.data.Providers[strings.ToLower(provider)]
	if !ok {
		return nil
	}
	models := make([]string, 0, len(providerData.Models))
	for name := range providerData.Models {
		models = append(models, name)
	}
	return models
}
