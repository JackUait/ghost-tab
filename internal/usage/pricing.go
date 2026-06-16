package usage

import "strings"

type modelRate struct{ inPerMTok, outPerMTok float64 }

// modelRates holds published Anthropic prices per 1,000,000 tokens, matched by
// model-id prefix so date-suffixed ids (e.g. claude-haiku-4-5-20251001) resolve.
var modelRates = map[string]modelRate{
	"claude-opus-4-5":   {5, 25},
	"claude-opus-4-6":   {5, 25},
	"claude-opus-4-7":   {5, 25},
	"claude-opus-4-8":   {5, 25},
	"claude-sonnet-4-5": {3, 15},
	"claude-sonnet-4-6": {3, 15},
	"claude-haiku-4-5":  {1, 5},
	"claude-fable-5":    {10, 50},
	"claude-mythos-5":   {10, 50},
}

const (
	// cacheWriteMult prices cache creation relative to the input rate. Our data
	// only has the combined cache_creation total, so it is charged at the 5-minute
	// TTL rate (1.25x) — a documented approximation that ignores any 1h-TTL writes.
	cacheWriteMult = 1.25
	// cacheReadMult prices cache reads relative to the input rate.
	cacheReadMult = 0.10
)

func rateFor(model string) (modelRate, bool) {
	for prefix, r := range modelRates {
		if strings.HasPrefix(model, prefix) {
			return r, true
		}
	}
	return modelRate{}, false
}

// ModelCostUSD returns the estimated USD cost for a model's usage and whether the
// model had a pricing entry. Input and output use the model's rates; cache-write
// is 1.25x the input rate and cache-read is 0.1x.
func ModelCostUSD(m ModelUsage) (float64, bool) {
	r, ok := rateFor(m.Model)
	if !ok {
		return 0, false
	}
	inRate := r.inPerMTok / 1_000_000
	outRate := r.outPerMTok / 1_000_000
	usd := float64(m.Input)*inRate +
		float64(m.Output)*outRate +
		float64(m.CacheWrite)*cacheWriteMult*inRate +
		float64(m.CacheRead)*cacheReadMult*inRate
	return usd, true
}

// CostUSD sums the cost of every priced model in the month. allPriced is false if
// any model in the month lacked a pricing entry.
func (mu MonthlyUsage) CostUSD() (float64, bool) {
	var total float64
	allPriced := true
	for _, m := range mu.Models {
		usd, priced := ModelCostUSD(m)
		if !priced {
			allPriced = false
			continue
		}
		total += usd
	}
	return total, allPriced
}
