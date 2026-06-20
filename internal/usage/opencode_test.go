package usage

import (
	"fmt"
	"testing"
	"time"
)

// ocMsg renders an OpenCode message JSON file body. created is epoch millis.
func ocMsg(role, model string, created int64, in, out, reasoning, cacheRead, cacheWrite int64) string {
	return fmt.Sprintf(
		`{"role":%q,"modelID":%q,"providerID":"anthropic","cost":0,`+
			`"tokens":{"input":%d,"output":%d,"reasoning":%d,"cache":{"read":%d,"write":%d}},`+
			`"time":{"created":%d}}`,
		role, model, in, out, reasoning, cacheRead, cacheWrite, created)
}

func TestParseOpenCodeMessage_assistantTokensByMonth(t *testing.T) {
	dir := t.TempDir()
	created := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC).UnixMilli()
	// reasoning (3) is billed as output, so Output = 20 + 3 = 23.
	p := writeFixture(t, dir, "msg_a.json", ocMsg("assistant", "claude-opus-4-8", created, 10, 20, 3, 1, 5))
	months, meta, err := ParseOpenCodeMessage(p)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Size == 0 {
		t.Errorf("meta.Size = 0, want non-zero")
	}
	m := months["2026-05"]
	if m == nil {
		t.Fatalf("no May usage: %+v", months)
	}
	if m.Input != 10 || m.Output != 23 || m.CacheWrite != 5 || m.CacheRead != 1 {
		t.Errorf("month = %+v, want input 10 output 23 cacheW 5 cacheR 1", m)
	}
	if len(m.Models) != 1 {
		t.Fatalf("models = %d, want 1", len(m.Models))
	}
	md := m.Models[0]
	if md.Model != "claude-opus-4-8" || md.Input != 10 || md.Output != 23 ||
		md.CacheRead != 1 || md.CacheWrite != 5 || md.CacheWrite1h != 0 {
		t.Errorf("model = %+v, want opus in10 out23 cr1 cw5 cw1h0", md)
	}
}

func TestParseOpenCodeMessage_skipsNonAssistant(t *testing.T) {
	dir := t.TempDir()
	created := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	p := writeFixture(t, dir, "msg_u.json", ocMsg("user", "claude-opus-4-8", created, 100, 0, 0, 0, 0))
	months, _, err := ParseOpenCodeMessage(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(months) != 0 {
		t.Errorf("months = %+v, want empty (user role skipped)", months)
	}
}

func TestParseOpenCodeMessage_skipsMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	p := writeFixture(t, dir, "msg_bad.json", "this is not json")
	months, _, err := ParseOpenCodeMessage(p)
	if err != nil {
		t.Fatalf("malformed file should not error, got %v", err)
	}
	if len(months) != 0 {
		t.Errorf("months = %+v, want empty for malformed json", months)
	}
}

func TestParseOpenCodeMessage_missingModelIsUnknown(t *testing.T) {
	dir := t.TempDir()
	created := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	p := writeFixture(t, dir, "msg_a.json", ocMsg("assistant", "", created, 10, 0, 0, 0, 0))
	months, _, _ := ParseOpenCodeMessage(p)
	if months["2026-05"] == nil || months["2026-05"].Models[0].Model != "unknown" {
		t.Errorf("missing model not attributed to unknown: %+v", months)
	}
}

func TestParseOpenCodeMessage_dropsZeroTokenMessage(t *testing.T) {
	dir := t.TempDir()
	created := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	// An assistant message with no token usage (e.g. an aborted turn) must not
	// create an empty month row.
	p := writeFixture(t, dir, "msg_a.json", ocMsg("assistant", "claude-opus-4-8", created, 0, 0, 0, 0, 0))
	months, _, _ := ParseOpenCodeMessage(p)
	if len(months) != 0 {
		t.Errorf("months = %+v, want empty (zero-token message dropped)", months)
	}
}

func TestParseOpenCodeMessage_skipsMissingTimestamp(t *testing.T) {
	dir := t.TempDir()
	// created == 0 means no completion time; cannot bucket by month, so skip.
	p := writeFixture(t, dir, "msg_a.json", ocMsg("assistant", "claude-opus-4-8", 0, 10, 0, 0, 0, 0))
	months, _, _ := ParseOpenCodeMessage(p)
	if len(months) != 0 {
		t.Errorf("months = %+v, want empty (missing timestamp skipped)", months)
	}
}

func TestMonthFromEpoch_handlesMillisAndSeconds(t *testing.T) {
	ts := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	if got := monthFromEpoch(ts.UnixMilli()); got != "2026-05" {
		t.Errorf("millis -> %q, want 2026-05", got)
	}
	// Older/custom providers may store seconds; values below the millis threshold
	// are interpreted as seconds rather than landing in 1970.
	if got := monthFromEpoch(ts.Unix()); got != "2026-05" {
		t.Errorf("seconds -> %q, want 2026-05", got)
	}
}
