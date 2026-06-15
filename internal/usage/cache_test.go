package usage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoadCache_missingFileReturnsEmpty(t *testing.T) {
	c := LoadCache(filepath.Join(t.TempDir(), "nope.json"))
	if c == nil || c.Files == nil {
		t.Fatalf("LoadCache = %+v, want non-nil cache with initialized Files", c)
	}
	if len(c.Files) != 0 {
		t.Errorf("Files = %v, want empty", c.Files)
	}
}

func TestLoadCache_corruptFileReturnsEmpty(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cache.json")
	writeFixture(t, filepath.Dir(p), "cache.json", "{not valid json")
	c := LoadCache(p)
	if c == nil || len(c.Files) != 0 {
		t.Fatalf("corrupt cache should rebuild empty, got %+v", c)
	}
}

func TestCache_saveThenLoadRoundTrips(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cache.json")
	c := &Cache{Version: cacheVersion, Files: map[string]fileCacheEntry{
		"/a.jsonl": {
			Meta:   FileMeta{ModTime: time.Unix(1000, 0).UTC(), Size: 42},
			Months: map[string]*MonthlyUsage{"2026-05": {Month: "2026-05", Input: 5}},
		},
	}}
	if err := c.Save(p); err != nil {
		t.Fatal(err)
	}
	got := LoadCache(p)
	entry, ok := got.Files["/a.jsonl"]
	if !ok || entry.Meta.Size != 42 || entry.Months["2026-05"].Input != 5 {
		t.Errorf("round-trip mismatch: %+v", got.Files)
	}
}
