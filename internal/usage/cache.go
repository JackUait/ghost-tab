package usage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const cacheVersion = 1

// fileCacheEntry stores one transcript file's identity and its parsed months.
type fileCacheEntry struct {
	Meta   FileMeta                 `json:"meta"`
	Months map[string]*MonthlyUsage `json:"months"`
}

// Cache is the persisted incremental-parse state keyed by absolute file path.
type Cache struct {
	Version int                       `json:"version"`
	Files   map[string]fileCacheEntry `json:"files"`
}

// LoadCache reads the cache file. A missing or corrupt cache returns an empty,
// usable cache (so the stats screen can always rebuild from scratch) — never nil.
func LoadCache(path string) *Cache {
	empty := &Cache{Version: cacheVersion, Files: map[string]fileCacheEntry{}}
	data, err := os.ReadFile(path)
	if err != nil {
		return empty
	}
	var c Cache
	if err := json.Unmarshal(data, &c); err != nil || c.Version != cacheVersion || c.Files == nil {
		return empty
	}
	return &c
}

// Save writes the cache atomically (temp file + rename).
func (c *Cache) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
