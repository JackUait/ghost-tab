# Mirror Subscriptions into OpenCode — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When a Wisp Deck subscription is added or changed on the Settings tab, mirror it into OpenCode's global config as a custom provider so the same key/models work in OpenCode too.

**Architecture:** A new `internal/opencodeconfig` package with a pure `MergeSubscriptions` core (rebuilds only `wisp-deck-*` providers in an existing `opencode.json`, preserving everything else) plus a `Sync` IO wrapper. `Sync` is called after every subscription mutation: the CLI `claude-config add/rename/delete`, the inline TUI panel (`WriteAPIKey`/`WriteModelMappings`), and the active-config switch (`persistClaudeConfig`). Base URLs come from a provider-name lookup table in `claudeconfig`.

**Tech Stack:** Go (`encoding/json`, `cobra`), bash (`lib/config-tui.sh`), Go test suite + `test/bash` integration tests.

## Global Constraints

- Go module path: `github.com/jackuait/wisp-deck`.
- TDD, no exceptions: write the failing test, run it red, implement, run it green, commit.
- Run `shellcheck` on any modified `lib/*.sh` before committing that task.
- Run the full suite `./run-tests.sh` before declaring the whole plan done; `git push` to `main` (this repo pushes straight to main — no PR).
- OpenCode provider ownership namespace prefix: `wisp-deck-` (exact string).
- OpenCode `$schema` value: `https://opencode.ai/config.json` (exact string).
- All mirrored providers use `npm: "@ai-sdk/anthropic"` (subscriptions speak the `ANTHROPIC_*` protocol).
- `apiKey` is written inline (plaintext), matching how the token is already stored in the Claude config JSON.
- Sync is best-effort: any IO/parse failure logs a warning and returns without blocking the Claude-side mutation. Never surface a blocking error.

---

### Task 1: Provider base-URL table in `claudeconfig`

**Files:**
- Modify: `internal/claudeconfig/claudeconfig.go` (add after `ProviderModels`, ~line 262)
- Test: `internal/claudeconfig/claudeconfig_test.go` (append)

**Interfaces:**
- Produces: `var ProviderBaseURLs map[string]string`; `func ProviderBaseURL(configName string) string` — returns the Anthropic-compatible gateway base URL for the provider whose key is a substring (case-insensitive) of `configName`, or `""` if none matches.

- [ ] **Step 1: Write the failing test**

Append to `internal/claudeconfig/claudeconfig_test.go`:

```go
func TestProviderBaseURL(t *testing.T) {
	cases := map[string]string{
		"Work GLM zhipu": "https://api.z.ai/api/anthropic",
		"my zhipu plan":  "https://api.z.ai/api/anthropic",
		"ZHIPU upper":    "https://api.z.ai/api/anthropic",
		"unknown vendor": "",
		"":               "",
	}
	for name, want := range cases {
		if got := ProviderBaseURL(name); got != want {
			t.Errorf("ProviderBaseURL(%q) = %q, want %q", name, got, want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/claudeconfig/ -run TestProviderBaseURL -v`
Expected: FAIL — `undefined: ProviderBaseURL`.

- [ ] **Step 3: Write minimal implementation**

In `internal/claudeconfig/claudeconfig.go`, add immediately after the `ProviderModels` var (after line 262):

```go
// ProviderBaseURLs maps a provider key to its Anthropic-compatible gateway base
// URL. Subscriptions store no base URL, so OpenCode mirroring derives it from the
// config name. zhipu (z.ai GLM Coding Plan) is verified; add new providers here as
// their endpoints are confirmed. mimo is intentionally absent until verified, so
// mimo subscriptions are skipped rather than pointed at a guessed endpoint.
var ProviderBaseURLs = map[string]string{
	"zhipu": "https://api.z.ai/api/anthropic",
}

// ProviderBaseURL returns the base URL for the provider whose key appears in the
// config name (case-insensitive), or "" if no known provider matches.
func ProviderBaseURL(configName string) string {
	lower := strings.ToLower(configName)
	for key, url := range ProviderBaseURLs {
		if strings.Contains(lower, key) {
			return url
		}
	}
	return ""
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/claudeconfig/ -run TestProviderBaseURL -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/claudeconfig/claudeconfig.go internal/claudeconfig/claudeconfig_test.go
git commit -m "feat(claudeconfig): add ProviderBaseURL lookup for OpenCode mirroring"
```

---

### Task 2: `opencodeconfig.MergeSubscriptions` pure core

**Files:**
- Create: `internal/opencodeconfig/opencodeconfig.go`
- Test: `internal/opencodeconfig/opencodeconfig_test.go`

**Interfaces:**
- Consumes: nothing from other tasks (pure, self-contained except `claudeconfig.Slugify`).
- Produces:
  - `type Subscription struct { Name, File, APIKey, BaseURL, OpusModel string; Models []string; Active bool }`
  - `func MergeSubscriptions(existing []byte, subs []Subscription) ([]byte, bool)` — returns the new `opencode.json` bytes and `true` on success; returns `(nil, false)` when `existing` is non-empty but not valid JSON (e.g. JSONC with comments), meaning "do not write".
  - `const ProviderPrefix = "wisp-deck-"`

- [ ] **Step 1: Write the failing tests**

Create `internal/opencodeconfig/opencodeconfig_test.go`:

```go
package opencodeconfig

import (
	"encoding/json"
	"strings"
	"testing"
)

func parse(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("result is not valid JSON: %v\n%s", err, b)
	}
	return m
}

func glmSub(active bool) Subscription {
	return Subscription{
		Name:      "Work GLM",
		File:      "work-glm.json",
		APIKey:    "sk-test-123",
		BaseURL:   "https://api.z.ai/api/anthropic",
		OpusModel: "glm-4.6",
		Models:    []string{"glm-4.6", "glm-4.5-air"},
		Active:    active,
	}
}

func TestMerge_creates_provider_from_empty(t *testing.T) {
	out, ok := MergeSubscriptions(nil, []Subscription{glmSub(true)})
	if !ok {
		t.Fatal("expected ok=true")
	}
	m := parse(t, out)
	if m["$schema"] != "https://opencode.ai/config.json" {
		t.Errorf("missing/wrong $schema: %v", m["$schema"])
	}
	if m["model"] != "wisp-deck-work-glm/glm-4.6" {
		t.Errorf("model = %v, want wisp-deck-work-glm/glm-4.6", m["model"])
	}
	prov := m["provider"].(map[string]any)["wisp-deck-work-glm"].(map[string]any)
	if prov["npm"] != "@ai-sdk/anthropic" {
		t.Errorf("npm = %v", prov["npm"])
	}
	if prov["name"] != "Work GLM" {
		t.Errorf("name = %v", prov["name"])
	}
	opts := prov["options"].(map[string]any)
	if opts["baseURL"] != "https://api.z.ai/api/anthropic" || opts["apiKey"] != "sk-test-123" {
		t.Errorf("options = %v", opts)
	}
	models := prov["models"].(map[string]any)
	if _, ok := models["glm-4.6"]; !ok {
		t.Errorf("models missing glm-4.6: %v", models)
	}
	if _, ok := models["glm-4.5-air"]; !ok {
		t.Errorf("models missing glm-4.5-air: %v", models)
	}
}

func TestMerge_preserves_user_keys_and_providers(t *testing.T) {
	existing := []byte(`{
  "theme": "tokyonight",
  "model": "anthropic/claude-sonnet-4-5",
  "provider": { "myown": { "name": "Mine" } }
}`)
	out, ok := MergeSubscriptions(existing, []Subscription{glmSub(false)})
	if !ok {
		t.Fatal("expected ok=true")
	}
	m := parse(t, out)
	if m["theme"] != "tokyonight" {
		t.Errorf("theme lost: %v", m["theme"])
	}
	// No active subscription -> top-level model must be left untouched.
	if m["model"] != "anthropic/claude-sonnet-4-5" {
		t.Errorf("model changed: %v", m["model"])
	}
	prov := m["provider"].(map[string]any)
	if _, ok := prov["myown"]; !ok {
		t.Errorf("user provider 'myown' was dropped: %v", prov)
	}
	if _, ok := prov["wisp-deck-work-glm"]; !ok {
		t.Errorf("wisp-deck provider not added: %v", prov)
	}
}

func TestMerge_rebuild_removes_stale_wisp_deck_providers(t *testing.T) {
	existing := []byte(`{"provider":{"wisp-deck-old":{"name":"Old"},"keep":{"name":"Keep"}}}`)
	out, ok := MergeSubscriptions(existing, []Subscription{glmSub(true)})
	if !ok {
		t.Fatal("expected ok=true")
	}
	prov := parse(t, out)["provider"].(map[string]any)
	if _, ok := prov["wisp-deck-old"]; ok {
		t.Errorf("stale wisp-deck-old not removed: %v", prov)
	}
	if _, ok := prov["keep"]; !ok {
		t.Errorf("user provider 'keep' was dropped: %v", prov)
	}
	if _, ok := prov["wisp-deck-work-glm"]; !ok {
		t.Errorf("current wisp-deck provider missing: %v", prov)
	}
}

func TestMerge_delete_removes_provider(t *testing.T) {
	existing := []byte(`{"provider":{"wisp-deck-work-glm":{"name":"Work GLM"}}}`)
	out, ok := MergeSubscriptions(existing, nil)
	if !ok {
		t.Fatal("expected ok=true")
	}
	m := parse(t, out)
	if p, ok := m["provider"]; ok {
		if pm, _ := p.(map[string]any); len(pm) != 0 {
			t.Errorf("provider should be empty/absent, got: %v", pm)
		}
	}
}

func TestMerge_skips_non_mirrorable(t *testing.T) {
	noKey := glmSub(true)
	noKey.APIKey = ""
	noURL := glmSub(true)
	noURL.BaseURL = ""
	noModels := glmSub(true)
	noModels.Models = nil
	out, ok := MergeSubscriptions(nil, []Subscription{noKey, noURL, noModels})
	if !ok {
		t.Fatal("expected ok=true")
	}
	m := parse(t, out)
	if p, ok := m["provider"]; ok {
		if pm, _ := p.(map[string]any); len(pm) != 0 {
			t.Errorf("no provider should be written, got: %v", pm)
		}
	}
	if _, ok := m["model"]; ok {
		t.Errorf("no default model should be set, got: %v", m["model"])
	}
}

func TestMerge_returns_false_on_jsonc_comments(t *testing.T) {
	existing := []byte("{\n  // a comment\n  \"theme\": \"x\"\n}")
	out, ok := MergeSubscriptions(existing, []Subscription{glmSub(true)})
	if ok || out != nil {
		t.Errorf("expected (nil,false) for JSONC input, got ok=%v out=%s", ok, out)
	}
}

func TestMerge_default_model_falls_back_to_first_model(t *testing.T) {
	s := glmSub(true)
	s.OpusModel = "" // no opus slot mapped
	out, _ := MergeSubscriptions(nil, []Subscription{s})
	if got := parse(t, out)["model"]; got != "wisp-deck-work-glm/glm-4.6" {
		t.Errorf("fallback model = %v, want wisp-deck-work-glm/glm-4.6", got)
	}
}

func TestProviderPrefixConst(t *testing.T) {
	if !strings.HasPrefix("wisp-deck-work-glm", ProviderPrefix) {
		t.Errorf("ProviderPrefix = %q", ProviderPrefix)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/opencodeconfig/ -v`
Expected: FAIL — package/symbols undefined.

- [ ] **Step 3: Write minimal implementation**

Create `internal/opencodeconfig/opencodeconfig.go`:

```go
// Package opencodeconfig mirrors Wisp Deck subscriptions (custom Claude configs)
// into OpenCode's global config as custom providers. MergeSubscriptions is the
// pure core; Sync (sync.go) wires it to disk.
package opencodeconfig

import (
	"encoding/json"
	"strings"

	"github.com/jackuait/wisp-deck/internal/claudeconfig"
)

// ProviderPrefix namespaces every provider Wisp Deck owns in opencode.json, so a
// rebuild can remove and re-add only these without touching user-authored ones.
const ProviderPrefix = "wisp-deck-"

const schemaURL = "https://opencode.ai/config.json"

// Subscription is the resolved view of a Wisp Deck subscription that OpenCode
// needs. BaseURL is pre-resolved by the caller (empty -> not mirrorable).
type Subscription struct {
	Name      string
	File      string
	APIKey    string
	BaseURL   string
	OpusModel string
	Models    []string
	Active    bool
}

// providerID returns the namespaced OpenCode provider id for a subscription.
func (s Subscription) providerID() string {
	slug := claudeconfig.Slugify(strings.TrimSuffix(s.File, ".json"))
	return ProviderPrefix + slug
}

// mirrorable reports whether the subscription has everything a working OpenCode
// provider needs.
func (s Subscription) mirrorable() bool {
	return s.APIKey != "" && s.BaseURL != "" && len(s.Models) > 0
}

// defaultModel returns the "<providerID>/<model>" string for the active provider.
func (s Subscription) defaultModel() string {
	model := s.OpusModel
	if model == "" {
		model = s.Models[0]
	}
	return s.providerID() + "/" + model
}

// MergeSubscriptions rebuilds the wisp-deck-* providers in existing opencode.json
// bytes from subs, preserving every other key. Returns (nil, false) when existing
// is non-empty but not valid JSON (e.g. JSONC), meaning the caller must not write.
func MergeSubscriptions(existing []byte, subs []Subscription) ([]byte, bool) {
	m := map[string]any{}
	if len(strings.TrimSpace(string(existing))) > 0 {
		if err := json.Unmarshal(existing, &m); err != nil {
			return nil, false
		}
	}

	m["$schema"] = schemaURL

	provider, _ := m["provider"].(map[string]any)
	if provider == nil {
		provider = map[string]any{}
	}
	// Drop every provider we own; user-authored ones stay.
	for id := range provider {
		if strings.HasPrefix(id, ProviderPrefix) {
			delete(provider, id)
		}
	}

	for _, s := range subs {
		if !s.mirrorable() {
			continue
		}
		models := map[string]any{}
		for _, id := range s.Models {
			models[id] = map[string]any{"name": id}
		}
		provider[s.providerID()] = map[string]any{
			"npm":  "@ai-sdk/anthropic",
			"name": s.Name,
			"options": map[string]any{
				"baseURL": s.BaseURL,
				"apiKey":  s.APIKey,
			},
			"models": models,
		}
		if s.Active {
			m["model"] = s.defaultModel()
		}
	}

	if len(provider) == 0 {
		delete(m, "provider")
	} else {
		m["provider"] = provider
	}

	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, false
	}
	return append(out, '\n'), true
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/opencodeconfig/ -v`
Expected: PASS (all 8 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/opencodeconfig/opencodeconfig.go internal/opencodeconfig/opencodeconfig_test.go
git commit -m "feat(opencodeconfig): pure MergeSubscriptions core"
```

---

### Task 3: `opencodeconfig.Sync` IO wrapper

**Files:**
- Create: `internal/opencodeconfig/sync.go`
- Test: `internal/opencodeconfig/sync_test.go`

**Interfaces:**
- Consumes: `MergeSubscriptions`, `Subscription` (Task 2); `claudeconfig.Load/GetActive/ReadAPIKey/ReadModelMappings/ModelsForConfig/ProviderBaseURL` (existing + Task 1).
- Produces:
  - `type Inputs struct { ListFile, ConfigsDir, PointerFile, Home string }`
  - `func Sync(in Inputs) error` — best-effort; resolves the OpenCode config path under `Home` (or `os.UserHomeDir()` if `Home==""`), builds subscriptions, merges, writes. Logs and returns nil on recoverable failure.
  - `func ConfigPath(home string) string` — resolves `<XDG_CONFIG_HOME|~/.config>/opencode/<first existing of opencode.jsonc, opencode.json, config.json | opencode.json>`.
  - `func BuildSubscriptions(in Inputs) []Subscription` — pure-ish builder from the wisp-deck config files (exported for direct testing).

- [ ] **Step 1: Write the failing tests**

Create `internal/opencodeconfig/sync_test.go`:

```go
package opencodeconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// seed writes a minimal wisp-deck config tree under root and returns Inputs.
func seed(t *testing.T, home string, active string) Inputs {
	t.Helper()
	cfgRoot := filepath.Join(home, ".config", "wisp-deck")
	configsDir := filepath.Join(cfgRoot, "claude-configs")
	if err := os.MkdirAll(configsDir, 0755); err != nil {
		t.Fatal(err)
	}
	listFile := filepath.Join(cfgRoot, "claude-configs.list")
	pointer := filepath.Join(cfgRoot, "claude-config")
	// "Work GLM zhipu" -> base URL resolves; map opus -> glm-4.6.
	os.WriteFile(listFile, []byte("Work GLM zhipu:work.json\n"), 0644)
	cfg := `{"env":{"ANTHROPIC_AUTH_TOKEN":"sk-abc","ANTHROPIC_DEFAULT_OPUS_MODEL":"glm-4.6"}}`
	os.WriteFile(filepath.Join(configsDir, "work.json"), []byte(cfg), 0644)
	if active != "" {
		os.WriteFile(pointer, []byte(active+"\n"), 0644)
	}
	return Inputs{ListFile: listFile, ConfigsDir: configsDir, PointerFile: pointer, Home: home}
}

func TestSync_writes_opencode_config(t *testing.T) {
	home := t.TempDir()
	in := seed(t, home, "work.json")
	if err := Sync(in); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "opencode", "opencode.json"))
	if err != nil {
		t.Fatalf("opencode.json not written: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["model"] != "wisp-deck-work-glm-zhipu/glm-4.6" {
		t.Errorf("model = %v", m["model"])
	}
	prov, ok := m["provider"].(map[string]any)["wisp-deck-work-glm-zhipu"].(map[string]any)
	if !ok {
		t.Fatalf("provider missing: %v", m["provider"])
	}
	if prov["options"].(map[string]any)["baseURL"] != "https://api.z.ai/api/anthropic" {
		t.Errorf("baseURL = %v", prov["options"])
	}
}

func TestSync_respects_xdg_config_home(t *testing.T) {
	home := t.TempDir()
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	in := seed(t, home, "work.json")
	if err := Sync(in); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if _, err := os.Stat(filepath.Join(xdg, "opencode", "opencode.json")); err != nil {
		t.Errorf("expected config under XDG_CONFIG_HOME: %v", err)
	}
}

func TestSync_preserves_existing_user_file(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "") // ensure ~/.config path
	ocDir := filepath.Join(home, ".config", "opencode")
	os.MkdirAll(ocDir, 0755)
	os.WriteFile(filepath.Join(ocDir, "opencode.json"),
		[]byte(`{"theme":"tokyonight","provider":{"mine":{"name":"Mine"}}}`), 0644)
	in := seed(t, home, "work.json")
	if err := Sync(in); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(ocDir, "opencode.json"))
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["theme"] != "tokyonight" {
		t.Errorf("theme lost: %v", m["theme"])
	}
	if _, ok := m["provider"].(map[string]any)["mine"]; !ok {
		t.Errorf("user provider lost: %v", m["provider"])
	}
}

func TestBuildSubscriptions_marks_active_and_resolves(t *testing.T) {
	home := t.TempDir()
	in := seed(t, home, "work.json")
	subs := BuildSubscriptions(in)
	if len(subs) != 1 {
		t.Fatalf("got %d subs, want 1", len(subs))
	}
	s := subs[0]
	if !s.Active {
		t.Errorf("sub should be active")
	}
	if s.BaseURL != "https://api.z.ai/api/anthropic" {
		t.Errorf("baseURL = %q", s.BaseURL)
	}
	if s.OpusModel != "glm-4.6" {
		t.Errorf("opusModel = %q", s.OpusModel)
	}
	if s.APIKey != "sk-abc" {
		t.Errorf("apiKey = %q", s.APIKey)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/opencodeconfig/ -run 'TestSync|TestBuild' -v`
Expected: FAIL — `Sync`, `BuildSubscriptions`, `ConfigPath` undefined.

- [ ] **Step 3: Write minimal implementation**

Create `internal/opencodeconfig/sync.go`:

```go
package opencodeconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackuait/wisp-deck/internal/claudeconfig"
)

// Inputs are the wisp-deck config paths plus the home dir used to locate the
// OpenCode global config.
type Inputs struct {
	ListFile    string
	ConfigsDir  string
	PointerFile string
	Home        string
}

// configFilenames is OpenCode's global-config resolution order; first existing
// wins, else we create opencode.json.
var configFilenames = []string{"opencode.jsonc", "opencode.json", "config.json"}

// ConfigPath returns the OpenCode global config path to write, honoring
// XDG_CONFIG_HOME and OpenCode's filename resolution order.
func ConfigPath(home string) string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(home, ".config")
	}
	dir := filepath.Join(base, "opencode")
	for _, name := range configFilenames {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(dir, "opencode.json")
}

// BuildSubscriptions reads the wisp-deck config files and resolves each config
// into a Subscription (api key, mapped models, base URL, active flag).
func BuildSubscriptions(in Inputs) []Subscription {
	configs := claudeconfig.Load(in.ListFile)
	active := claudeconfig.GetActive(in.PointerFile)
	var subs []Subscription
	for _, c := range configs {
		models := claudeconfig.ModelsForConfig(c.Name)
		idx := claudeconfig.ReadModelMappings(in.ConfigsDir, c.File, models)
		var mapped []string
		seen := map[string]bool{}
		for _, i := range idx {
			if i >= 0 && i < len(models) && !seen[models[i]] {
				seen[models[i]] = true
				mapped = append(mapped, models[i])
			}
		}
		opus := ""
		if idx[0] >= 0 && idx[0] < len(models) {
			opus = models[idx[0]]
		}
		subs = append(subs, Subscription{
			Name:      c.Name,
			File:      c.File,
			APIKey:    claudeconfig.ReadAPIKey(in.ConfigsDir, c.File),
			BaseURL:   claudeconfig.ProviderBaseURL(c.Name),
			OpusModel: opus,
			Models:    mapped,
			Active:    c.File == active,
		})
	}
	return subs
}

// Sync rebuilds the wisp-deck-* providers in OpenCode's global config. It is
// best-effort: recoverable failures log a warning and return nil so a Claude-side
// mutation is never blocked.
func Sync(in Inputs) error {
	home := in.Home
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	if home == "" {
		return nil
	}
	path := ConfigPath(home)

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "wisp-deck: opencode sync: read %s: %v\n", path, err)
		return nil
	}

	subs := BuildSubscriptions(in)
	for _, s := range subs {
		if s.APIKey != "" && len(s.Models) > 0 && s.BaseURL == "" {
			fmt.Fprintf(os.Stderr, "wisp-deck: opencode sync: no base URL for %q; skipped\n", s.Name)
		}
	}

	out, ok := MergeSubscriptions(existing, subs)
	if !ok {
		fmt.Fprintf(os.Stderr, "wisp-deck: opencode sync: %s is not plain JSON; left unchanged\n", path)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "wisp-deck: opencode sync: mkdir: %v\n", err)
		return nil
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "wisp-deck: opencode sync: write %s: %v\n", path, err)
		return nil
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/opencodeconfig/ -v`
Expected: PASS (Task 2 + Task 3 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/opencodeconfig/sync.go internal/opencodeconfig/sync_test.go
git commit -m "feat(opencodeconfig): Sync wrapper resolving paths and building subscriptions"
```

---

### Task 4: Hook Sync into the inline TUI panel and active switch

**Files:**
- Modify: `internal/tui/mainmenu.go` (add `syncOpenCode` helper; call it in `persistClaudeConfig` ~line 924)
- Modify: `internal/tui/claude_config_panel.go` (call after `WriteModelMappings` line 44 and `WriteAPIKey` line 119)
- Test: `internal/tui/opencode_sync_hook_test.go` (create)

**Interfaces:**
- Consumes: `opencodeconfig.Sync`, `opencodeconfig.Inputs` (Task 3); existing model fields `claudeConfigsList`, `claudeConfigsDir`, `claudeConfigFile`.
- Produces: `func (m *MainMenuModel) syncOpenCode()` — builds `Inputs` from the model's config paths and calls `opencodeconfig.Sync`.

- [ ] **Step 1: Write the failing test**

Create `internal/tui/opencode_sync_hook_test.go`:

```go
package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncOpenCode_writes_config_from_model_paths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	cfgRoot := filepath.Join(home, ".config", "wisp-deck")
	configsDir := filepath.Join(cfgRoot, "claude-configs")
	os.MkdirAll(configsDir, 0755)
	os.WriteFile(filepath.Join(cfgRoot, "claude-configs.list"),
		[]byte("Work GLM zhipu:work.json\n"), 0644)
	os.WriteFile(filepath.Join(configsDir, "work.json"),
		[]byte(`{"env":{"ANTHROPIC_AUTH_TOKEN":"sk-abc","ANTHROPIC_DEFAULT_OPUS_MODEL":"glm-4.6"}}`), 0644)
	os.WriteFile(filepath.Join(cfgRoot, "claude-config"), []byte("work.json\n"), 0644)

	m := &MainMenuModel{
		claudeConfigsList: filepath.Join(cfgRoot, "claude-configs.list"),
		claudeConfigsDir:  configsDir,
		claudeConfigFile:  filepath.Join(cfgRoot, "claude-config"),
	}
	t.Setenv("HOME", home)
	m.syncOpenCode()

	if _, err := os.Stat(filepath.Join(home, ".config", "opencode", "opencode.json")); err != nil {
		t.Errorf("syncOpenCode did not write opencode.json: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -run TestSyncOpenCode -v`
Expected: FAIL — `m.syncOpenCode undefined`.

- [ ] **Step 3: Write minimal implementation**

In `internal/tui/mainmenu.go`, confirm the import block includes `"os"` (it does — used at line 1453) and add `"github.com/jackuait/wisp-deck/internal/opencodeconfig"` to the imports. Then add this helper directly after `persistClaudeConfig` (after line 935):

```go
// syncOpenCode mirrors the current subscriptions into OpenCode's global config.
// Best-effort: errors are swallowed inside opencodeconfig.Sync.
func (m *MainMenuModel) syncOpenCode() {
	if m.claudeConfigsList == "" || m.claudeConfigsDir == "" {
		return
	}
	home, _ := os.UserHomeDir()
	_ = opencodeconfig.Sync(opencodeconfig.Inputs{
		ListFile:    m.claudeConfigsList,
		ConfigsDir:  m.claudeConfigsDir,
		PointerFile: m.claudeConfigFile,
		Home:        home,
	})
}
```

Then call it at the end of `persistClaudeConfig`. Replace the body of `persistClaudeConfig` (lines 924-935) so the function ends with `m.syncOpenCode()`:

```go
func (m *MainMenuModel) persistClaudeConfig() {
	if m.claudeConfigFile == "" {
		return
	}
	file := m.CurrentClaudeConfigFile()
	if file == "" {
		_ = os.Remove(m.claudeConfigFile)
		m.syncOpenCode()
		return
	}
	_ = os.MkdirAll(filepath.Dir(m.claudeConfigFile), 0755)
	_ = os.WriteFile(m.claudeConfigFile, []byte(file+"\n"), 0644)
	m.syncOpenCode()
}
```

In `internal/tui/claude_config_panel.go`, add a sync call after each successful mutation. After line 47 (inside the `KeyEnter` case, right after `m.modelMapOpen = false` following a successful `WriteModelMappings`), and after the successful `WriteAPIKey` (after line 122). Concretely, change the `WriteModelMappings` success branch:

```go
		if err := claudeconfig.WriteModelMappings(m.claudeConfigsDir, file, m.modelMap, m.modelMapModels); err != nil {
			m.modelMapErr = err
			return m, nil
		}
		m.syncOpenCode()
		m.modelMapOpen = false
		return m, nil
```

and the `WriteAPIKey` success branch:

```go
			if err := claudeconfig.WriteAPIKey(m.claudeConfigsDir, file, key); err != nil {
				m.modelMapErr = err
				return m, nil
			}
			m.syncOpenCode()
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/ -run TestSyncOpenCode -v`
Expected: PASS.
Then run the whole TUI package to confirm no regressions: `go test ./internal/tui/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/tui/mainmenu.go internal/tui/claude_config_panel.go internal/tui/opencode_sync_hook_test.go
git commit -m "feat(tui): sync OpenCode after API-key/model-map edits and config switch"
```

---

### Task 5: Hook Sync into the CLI mutations (add/rename/delete) + bash

**Files:**
- Modify: `cmd/wisp-deck-tui/claude_config.go` (call Sync after each mutation; add `--pointer` flag to add & rename)
- Modify: `lib/config-tui.sh:34,39` (pass `--pointer "$pointer_file"` to add & rename)
- Test: `test/bash/opencode_sync_test.go` (create)

**Interfaces:**
- Consumes: `opencodeconfig.Sync`, `opencodeconfig.Inputs` (Task 3).
- Produces: CLI `claude-config add|rename|delete` now mirror into OpenCode; `add` and `rename` accept `--pointer`.

- [ ] **Step 1: Write the failing test**

Create `test/bash/opencode_sync_test.go`:

```go
package bash_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildTUI compiles wisp-deck-tui once into a temp dir and returns its path.
func buildTUI(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "wisp-deck-tui")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/wisp-deck-tui")
	cmd.Dir = repoRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

// repoRoot walks up from CWD to the module root (where go.mod lives).
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func TestCLIAdd_then_key_mirrors_into_opencode(t *testing.T) {
	bin := buildTUI(t)
	home := t.TempDir()
	cfgRoot := filepath.Join(home, ".config", "wisp-deck")
	configsDir := filepath.Join(cfgRoot, "claude-configs")
	os.MkdirAll(configsDir, 0755)
	list := filepath.Join(cfgRoot, "claude-configs.list")
	pointer := filepath.Join(cfgRoot, "claude-config")

	env := append(os.Environ(), "HOME="+home, "XDG_CONFIG_HOME=")

	// add a zhipu-named config (so base URL resolves), make it active, give it a key+mapping.
	run := func(args ...string) {
		c := exec.Command(bin, args...)
		c.Env = env
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	run("claude-config", "add", "--list", list, "--dir", configsDir, "--pointer", pointer, "--name", "Work GLM zhipu")
	// add wrote work-glm-zhipu.json; give it a key + opus mapping, mark active.
	cfgFile := filepath.Join(configsDir, "work-glm-zhipu.json")
	os.WriteFile(cfgFile, []byte(`{"env":{"ANTHROPIC_AUTH_TOKEN":"sk-abc","ANTHROPIC_DEFAULT_OPUS_MODEL":"glm-4.6"}}`), 0644)
	os.WriteFile(pointer, []byte("work-glm-zhipu.json\n"), 0644)
	// rename triggers a re-sync that now sees the key+mapping.
	run("claude-config", "rename", "--list", list, "--pointer", pointer, "--file", "work-glm-zhipu.json", "--name", "Work GLM zhipu")

	data, err := os.ReadFile(filepath.Join(home, ".config", "opencode", "opencode.json"))
	if err != nil {
		t.Fatalf("opencode.json not written: %v", err)
	}
	var m map[string]any
	json.Unmarshal(data, &m)
	if !strings.Contains(string(data), "wisp-deck-work-glm-zhipu") {
		t.Errorf("provider not mirrored:\n%s", data)
	}
	if m["model"] != "wisp-deck-work-glm-zhipu/glm-4.6" {
		t.Errorf("model = %v", m["model"])
	}

	// delete removes the provider again.
	run("claude-config", "delete", "--list", list, "--dir", configsDir, "--pointer", pointer, "--file", "work-glm-zhipu.json")
	data, _ = os.ReadFile(filepath.Join(home, ".config", "opencode", "opencode.json"))
	if strings.Contains(string(data), "wisp-deck-work-glm-zhipu") {
		t.Errorf("provider not removed after delete:\n%s", data)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./test/bash/ -run TestCLIAdd_then_key_mirrors -v`
Expected: FAIL — `--pointer` unknown flag on add, and no opencode.json written.

- [ ] **Step 3: Write minimal implementation**

In `cmd/wisp-deck-tui/claude_config.go`, add the import `"github.com/jackuait/wisp-deck/internal/opencodeconfig"`, a shared helper, and `--pointer` flags. Replace the three `RunE` bodies and the `init()` flag block:

```go
func syncOpenCode() {
	if ccList == "" || ccDir == "" {
		return
	}
	home, _ := os.UserHomeDir()
	_ = opencodeconfig.Sync(opencodeconfig.Inputs{
		ListFile:    ccList,
		ConfigsDir:  ccDir,
		PointerFile: ccPointer,
		Home:        home,
	})
}
```

Add (the package will need `"os"` imported):

```go
	RunE: func(cmd *cobra.Command, args []string) error {  // add
		file, err := claudeconfig.Add(ccList, ccDir, ccName)
		if err != nil {
			return err
		}
		syncOpenCode()
		fmt.Fprintln(cmd.OutOrStdout(), file)
		return nil
	},
```

```go
	RunE: func(cmd *cobra.Command, args []string) error {  // rename
		if err := claudeconfig.Rename(ccList, ccFile, ccName); err != nil {
			return err
		}
		syncOpenCode()
		return nil
	},
```

```go
	RunE: func(cmd *cobra.Command, args []string) error {  // delete
		if err := claudeconfig.Delete(ccList, ccDir, ccPointer, ccFile); err != nil {
			return err
		}
		syncOpenCode()
		return nil
	},
```

In `init()`, add `--pointer` to add and rename (delete already has it). Note `ccDir` is needed by `syncOpenCode` for rename too, so add `--dir` to rename:

```go
	claudeConfigAddCmd.Flags().StringVar(&ccPointer, "pointer", "", "Path to active config pointer file")
	claudeConfigRenameCmd.Flags().StringVar(&ccDir, "dir", "", "Path to configs directory")
	claudeConfigRenameCmd.Flags().StringVar(&ccPointer, "pointer", "", "Path to active config pointer file")
```

In `lib/config-tui.sh`, pass the new flags. Line 34 (add) becomes:

```bash
        [ -n "$name" ] && [ "$name" != "null" ] && wisp-deck-tui claude-config add --list "$list_file" --dir "$configs_dir" --pointer "$pointer_file" --name "$name" >/dev/null
```

Line 39 (rename) becomes:

```bash
        [ -n "$file" ] && [ "$file" != "null" ] && [ -n "$name" ] && [ "$name" != "null" ] && wisp-deck-tui claude-config rename --list "$list_file" --dir "$configs_dir" --pointer "$pointer_file" --file "$file" --name "$name"
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./test/bash/ -run TestCLIAdd_then_key_mirrors -v`
Expected: PASS.
Then lint the bash: `shellcheck lib/config-tui.sh`
Expected: no warnings.

- [ ] **Step 5: Commit**

```bash
git add cmd/wisp-deck-tui/claude_config.go lib/config-tui.sh test/bash/opencode_sync_test.go
git commit -m "feat(cli): mirror subscriptions into OpenCode on add/rename/delete"
```

---

### Task 6: Full verification and push

- [ ] **Step 1: Run the full test suite**

Run: `./run-tests.sh`
Expected: all pass.

- [ ] **Step 2: shellcheck all touched/auxiliary scripts**

Run: `shellcheck lib/*.sh lib/terminals/*.sh bin/wisp-deck wrapper.sh`
Expected: no warnings (config-tui.sh is the only one this plan changes).

- [ ] **Step 3: Manual smoke (optional but recommended)**

Build and run a real add+key flow against a scratch HOME, then inspect `~/.config/opencode/opencode.json` to confirm the `wisp-deck-*` provider and `model` appear.

- [ ] **Step 4: Push**

```bash
git pull --rebase
git push
git status   # MUST show "up to date with origin"
```

---

## Self-Review notes

- **Spec coverage:** full provider mirror (Tasks 2-3), re-sync on every change (Tasks 4-5 cover all four mutation entry points + active switch), merge into global config preserving user keys (Task 2 tests 2/3 + Task 3 preserve test), base URL derived from provider name (Task 1), JSONC abort (Task 2 test 6), MiMo skipped (no `mimo` table entry → `ProviderBaseURL` returns "" → non-mirrorable). Out-of-scope items (cost/limit, base-URL UI field, env-ref apiKey, Codex) correctly omitted.
- **Type consistency:** `Subscription{Name,File,APIKey,BaseURL,OpusModel,Models,Active}`, `MergeSubscriptions(existing []byte, subs []Subscription) ([]byte, bool)`, `Inputs{ListFile,ConfigsDir,PointerFile,Home}`, `Sync(Inputs) error`, `BuildSubscriptions(Inputs) []Subscription`, `ConfigPath(string) string`, `ProviderPrefix` const, `claudeconfig.ProviderBaseURL(string) string` — all used consistently across tasks.
- **Refinement vs spec:** spec wrote `Sync(home)`; plan refines to `Sync(Inputs)` since the wisp-deck config paths must be passed in (the model/CLI already hold them). Behavior is unchanged.
