package usage

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
)

// MonthlyUsage holds token counts for a single YYYY-MM bucket.
type MonthlyUsage struct {
	Month      string `json:"month"`
	Input      int64  `json:"input"`
	Output     int64  `json:"output"`
	CacheWrite int64  `json:"cache_write"`
	CacheRead  int64  `json:"cache_read"`
}

// Total returns the sum of all token columns.
func (m MonthlyUsage) Total() int64 {
	return m.Input + m.Output + m.CacheWrite + m.CacheRead
}

// FileMeta captures the on-disk identity used for incremental caching.
type FileMeta struct {
	ModTime time.Time `json:"mod_time"`
	Size    int64     `json:"size"`
}

// maxLineBytes bounds a single transcript line (some carry large embedded content).
const maxLineBytes = 50 * 1024 * 1024

type transcriptRecord struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		ID    string `json:"id"`
		Usage *struct {
			Input      int64 `json:"input_tokens"`
			Output     int64 `json:"output_tokens"`
			CacheWrite int64 `json:"cache_creation_input_tokens"`
			CacheRead  int64 `json:"cache_read_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// ParseFile reads a single .jsonl transcript and aggregates token usage by month.
// Non-assistant records, records without usage, and malformed lines are skipped.
// Assistant records are deduped by message.id within this file.
func ParseFile(path string) (map[string]*MonthlyUsage, FileMeta, error) {
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

	months := map[string]*MonthlyUsage{}
	seen := map[string]bool{}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), maxLineBytes)
	for sc.Scan() {
		line := sc.Bytes()
		var rec transcriptRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			continue
		}
		if rec.Type != "assistant" || rec.Message.Usage == nil {
			continue
		}
		if len(rec.Timestamp) < 7 {
			continue
		}
		if id := rec.Message.ID; id != "" {
			if seen[id] {
				continue
			}
			seen[id] = true
		}
		month := rec.Timestamp[:7]
		mu := months[month]
		if mu == nil {
			mu = &MonthlyUsage{Month: month}
			months[month] = mu
		}
		u := rec.Message.Usage
		mu.Input += u.Input
		mu.Output += u.Output
		mu.CacheWrite += u.CacheWrite
		mu.CacheRead += u.CacheRead
	}
	if err := sc.Err(); err != nil {
		return nil, meta, err
	}
	return months, meta, nil
}
