package usage

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// opencodeMessage is the subset of an OpenCode stored message we read. OpenCode
// writes one JSON object per message under
// <data>/storage/message/<sessionID>/msg_*.json. Only assistant messages carry
// token usage; user messages and tool parts do not. Note the nested cache shape
// (tokens.cache.read / .write) and that cost is frequently 0 in storage, so we
// reprice from tokens via pricing.go rather than trusting it.
type opencodeMessage struct {
	Role    string `json:"role"`
	ModelID string `json:"modelID"`
	Time    struct {
		Created int64 `json:"created"`
	} `json:"time"`
	Tokens *struct {
		Input     int64 `json:"input"`
		Output    int64 `json:"output"`
		Reasoning int64 `json:"reasoning"`
		Cache     struct {
			Read  int64 `json:"read"`
			Write int64 `json:"write"`
		} `json:"cache"`
	} `json:"tokens"`
}

// millisThreshold separates epoch-millisecond timestamps from epoch-second ones.
// 1e12 ms is ~2001; any plausible "now" expressed in seconds (~1.7e9) is far
// below it, so the split is unambiguous for real session data.
const millisThreshold = 1_000_000_000_000

// monthFromEpoch converts an OpenCode time.created value to a YYYY-MM bucket.
// OpenCode stores JS epoch milliseconds, but some older/custom providers have
// been observed storing seconds, so sub-threshold values are read as seconds to
// avoid bucketing them into 1970.
func monthFromEpoch(v int64) string {
	var t time.Time
	if v >= millisThreshold {
		t = time.UnixMilli(v).UTC()
	} else {
		t = time.Unix(v, 0).UTC()
	}
	return t.Format("2006-01")
}

// ParseOpenCodeMessage reads a single OpenCode message JSON file and aggregates
// its token usage by month and model, matching ParseFile's contract so the
// Aggregate cache can treat both sources uniformly. Non-assistant messages,
// messages without token usage or a timestamp, zero-token turns, and malformed
// files are skipped (returning an empty map, not an error). A message with no
// model id is attributed to "unknown". Reasoning tokens are folded into Output
// since they are billed at the output rate.
func ParseOpenCodeMessage(path string) (map[string]*MonthlyUsage, FileMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, FileMeta{}, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, FileMeta{}, err
	}
	meta := FileMeta{ModTime: info.ModTime(), Size: info.Size()}

	empty := map[string]*MonthlyUsage{}
	data, err := io.ReadAll(io.LimitReader(f, maxLineBytes))
	if err != nil {
		return nil, meta, err
	}
	var msg opencodeMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return empty, meta, nil // skip non-JSON / unexpected shape
	}
	if msg.Role != "assistant" || msg.Tokens == nil || msg.Time.Created <= 0 {
		return empty, meta, nil
	}

	model := msg.ModelID
	if model == "" {
		model = "unknown"
	}
	byModel := map[string]*ModelUsage{}
	// OpenCode does not split cache writes by TTL, so the whole cache.write total
	// is charged at the 5-minute rate (CacheWrite1h stays 0) via the flat path.
	addCounts(byModel, model, usageCounts{
		Input:      msg.Tokens.Input,
		Output:     msg.Tokens.Output + msg.Tokens.Reasoning,
		CacheWrite: msg.Tokens.Cache.Write,
		CacheRead:  msg.Tokens.Cache.Read,
	})

	month := monthFromEpoch(msg.Time.Created)
	months := map[string]*MonthlyUsage{}
	if mu := buildMonthly(month, byModel); mu != nil {
		months[month] = mu
	}
	return months, meta, nil
}
