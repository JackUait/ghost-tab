package usage

import (
	"math"
	"testing"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestModelCostUSD_tiers(t *testing.T) {
	// 1M input + 1M output on opus = $5 + $25 = $30
	usd, priced := ModelCostUSD(ModelUsage{Model: "claude-opus-4-7", Input: 1_000_000, Output: 1_000_000})
	if !priced || !approx(usd, 30) {
		t.Errorf("opus = %v priced=%v, want 30/true", usd, priced)
	}
	// sonnet 1M in + 1M out = $3 + $15 = $18
	usd, _ = ModelCostUSD(ModelUsage{Model: "claude-sonnet-4-6", Input: 1_000_000, Output: 1_000_000})
	if !approx(usd, 18) {
		t.Errorf("sonnet = %v, want 18", usd)
	}
	// haiku 1M in = $1
	usd, _ = ModelCostUSD(ModelUsage{Model: "claude-haiku-4-5", Input: 1_000_000})
	if !approx(usd, 1) {
		t.Errorf("haiku = %v, want 1", usd)
	}
	// fable 1M in + 1M out = $10 + $50 = $60
	usd, _ = ModelCostUSD(ModelUsage{Model: "claude-fable-5", Input: 1_000_000, Output: 1_000_000})
	if !approx(usd, 60) {
		t.Errorf("fable = %v, want 60", usd)
	}
}

func TestModelCostUSD_cacheMultipliers(t *testing.T) {
	// opus input rate $5/MTok. cache-write 1.25x = $6.25/MTok; cache-read 0.1x = $0.50/MTok
	usd, _ := ModelCostUSD(ModelUsage{Model: "claude-opus-4-7", CacheWrite: 1_000_000})
	if !approx(usd, 6.25) {
		t.Errorf("cacheWrite = %v, want 6.25", usd)
	}
	usd, _ = ModelCostUSD(ModelUsage{Model: "claude-opus-4-7", CacheRead: 1_000_000})
	if !approx(usd, 0.5) {
		t.Errorf("cacheRead = %v, want 0.5", usd)
	}
}

func TestModelCostUSD_prefixMatchAndUnknown(t *testing.T) {
	usd, priced := ModelCostUSD(ModelUsage{Model: "claude-haiku-4-5-20251001", Input: 1_000_000})
	if !priced || !approx(usd, 1) {
		t.Errorf("suffixed haiku = %v priced=%v, want 1/true", usd, priced)
	}
	if _, priced := ModelCostUSD(ModelUsage{Model: "gpt-4o", Input: 1_000_000}); priced {
		t.Errorf("unknown model should be unpriced")
	}
}

func TestMonthlyCostUSD_sumAndFlag(t *testing.T) {
	mu := MonthlyUsage{Models: []ModelUsage{
		{Model: "claude-opus-4-7", Input: 1_000_000}, // $5
		{Model: "mystery", Input: 1_000_000},         // unpriced
	}}
	usd, allPriced := mu.CostUSD()
	if !approx(usd, 5) {
		t.Errorf("month cost = %v, want 5 (priced only)", usd)
	}
	if allPriced {
		t.Errorf("allPriced should be false when a model is unpriced")
	}
}
