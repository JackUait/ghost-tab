# Mirror Ghost Tab subscriptions into OpenCode

**Date:** 2026-06-21
**Status:** Approved design, pending implementation plan

## Problem

A Ghost Tab "subscription" (the PLAN / Subscription row on the Settings tab) is a custom
Claude config: a settings JSON under `~/.config/ghost-tab/claude-configs/` holding an
`ANTHROPIC_AUTH_TOKEN` and model-alias remaps
(`ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL` → provider models such as
`glm-4.6`). Today this configures **only Claude Code** (launched via
`claude --settings <file>`). OpenCode is launched with no config at all
(`npx opencode-ai@latest`), so a subscription the user adds is invisible to OpenCode.

**Goal:** when a subscription is added or changed on the Settings tab, mirror it into
OpenCode so the same provider/key/models are usable there too.

## Decisions (from brainstorming)

1. **Sync scope: full provider mirror.** Each subscription becomes an OpenCode custom
   provider (base URL + API key), its mapped provider models become OpenCode models, and
   the active subscription sets OpenCode's default `model`.
2. **Trigger: re-sync on every change.** Add / edit-key / edit-mappings / rename / delete /
   active-switch all re-run one idempotent sync.
3. **Config ownership: merge into the global config.** Read/merge
   `~/.config/opencode/opencode.json` (XDG-respecting), touching only Ghost Tab's own
   provider blocks and leaving every other key the user has set untouched.
4. **Base URL source: derive from provider name.** A subscription stores no base URL, so
   it is looked up from a provider table keyed off the config name (same substring match as
   `claudeconfig.ModelsForConfig`).

## Background facts (verified against the codebase + OpenCode docs)

- `internal/claudeconfig/claudeconfig.go` is the single source of truth for subscriptions:
  `Load`, `GetActive`, `Add`, `Rename`, `Delete`, `WriteAPIKey`, `WriteModelMappings`,
  `ReadAPIKey`, `ReadModelMappings`, plus `ProviderModels` / `ModelsForConfig`.
  It writes **only** `ANTHROPIC_AUTH_TOKEN` and the four `ANTHROPIC_DEFAULT_*_MODEL` keys —
  **never `ANTHROPIC_BASE_URL`.**
- OpenCode global config: `~/.config/opencode/opencode.json` (respects `XDG_CONFIG_HOME`).
  Resolution order `opencode.jsonc` → `opencode.json` → `config.json`; first found wins.
  Merged (not replaced) against other layers; project config overrides global.
- A custom provider block:
  ```jsonc
  "provider": {
    "<id>": {
      "npm": "@ai-sdk/anthropic",
      "name": "<display name>",
      "options": { "baseURL": "<url>", "apiKey": "<key>" },
      "models": { "<model-id>": { "name": "<model-id>" } }
    }
  }
  ```
- Default model: top-level `"model": "<providerId>/<modelId>"`.
- `apiKey` may be inline, `{env:VAR}`, or `{file:path}`. We use **inline**, consistent with
  the token already living in plaintext in the Claude config JSON.
- OpenCode's JSON parser is JSONC-aware; Go's `encoding/json` is **not**. If the existing
  file contains comments, a strict parse fails — we abort that sync rather than clobber.

## Design

### New package: `internal/opencodeconfig`

A small package with a **pure, fully-unit-tested merge core** and a thin IO wrapper.

#### Pure core (no IO)

```go
// Subscription is the minimal view of a Ghost Tab subscription that OpenCode needs.
type Subscription struct {
    Name     string   // display name -> provider "name"
    File     string   // config filename -> slug -> provider id
    APIKey   string   // ANTHROPIC_AUTH_TOKEN -> options.apiKey
    Models   []string // distinct mapped provider model ids (from the 4 alias slots)
    OpusModel string  // the OPUS-slot model id, used to pick the default model
    Active   bool     // true for the currently-selected subscription
}

// MergeSubscriptions takes the existing opencode.json bytes (may be nil/empty) and the
// current subscription set, and returns the new bytes. It:
//   - removes every provider whose id is prefixed "ghost-tab-"
//   - re-adds one provider per mirrorable subscription
//   - sets top-level "model" to the active subscription's default, only if a custom
//     subscription is active; otherwise leaves "model" untouched
//   - ensures "$schema" is present
//   - preserves all other keys byte-for-value
// Returns (nil, false) — meaning "do not write" — if the existing bytes are non-empty and
// fail to parse as JSON (e.g. JSONC with comments).
func MergeSubscriptions(existing []byte, subs []Subscription) ([]byte, bool)
```

- **Provider id:** `"ghost-tab-" + claudeconfig.Slugify(strings.TrimSuffix(file, ".json"))`.
  The `ghost-tab-` prefix is the ownership namespace: sync removes/re-adds only these,
  never user-authored providers.
- **`npm`:** `@ai-sdk/anthropic` for all entries (subscriptions speak the `ANTHROPIC_*`
  protocol, which is how Claude Code reaches these gateways).
- **`baseURL`:** `ProviderBaseURL(sub.Name + " " + sub.File)` (see table below).
- **`apiKey`:** inline `sub.APIKey`.
- **`models`:** one entry per id in `sub.Models`, value `{ "name": "<id>" }`. Cost/limit
  enrichment is an optional follow-up (see Out of scope).
- **Default model:** for the active subscription, `"<id>/<OpusModel>"`; fall back to the
  first of `Models` if no OPUS slot is set. Not written when the active subscription is
  Standard (no custom subscription active).
- **Mirrorable:** a subscription is mirrored only if it has a non-empty `APIKey`, at least
  one model, and a resolvable base URL. Non-mirrorable subscriptions are skipped (and, for
  an unresolved base URL, a warning is logged by the IO wrapper).

#### Provider base-URL table (in `claudeconfig`, beside `ProviderModels`)

```go
// ProviderBaseURLs maps a provider key to its Anthropic-compatible gateway base URL.
var ProviderBaseURLs = map[string]string{
    "zhipu": "https://api.z.ai/api/anthropic", // verified (z.ai GLM Coding Plan)
    // "mimo": TO CONFIRM — Xiaomi MiMo endpoint not yet verified; until added,
    //         mimo subscriptions are skipped with a warning.
}

// ProviderBaseURL returns the base URL for the provider matching the config name,
// or "" if none matches.
func ProviderBaseURL(configName string) string
```

Keyed by the same substring match `ModelsForConfig` already uses, so naming stays
consistent. zhipu/GLM is the one verified entry; MiMo is added once its endpoint is
confirmed. An unmatched name yields `""` → that subscription is skipped (not pointed at a
guessed endpoint).

#### IO wrapper

```go
// Sync rebuilds the ghost-tab-* providers in OpenCode's global config from the current
// subscription set. Best-effort: a read/parse/write failure logs a warning and returns
// nil so it never blocks a Claude-side mutation.
func Sync(home string) error
```

`Sync`:
1. Resolves the config path: `<XDG_CONFIG_HOME or ~/.config>/opencode/`, picking the first
   existing of `opencode.jsonc`, `opencode.json`, `config.json`; else `opencode.json`.
2. Reads existing bytes (absent → nil).
3. Builds `[]Subscription` from `claudeconfig.Load(listFile)` + `ReadAPIKey` +
   `ReadModelMappings` + the active pointer + `ProviderBaseURL`.
4. Calls `MergeSubscriptions`. If it returns `ok == false` (unparseable existing file),
   logs a warning and returns without writing.
5. Writes the result back to the resolved path (2-space indent, trailing newline),
   creating the directory if needed.

### Hook points

`Sync(home)` is called after every subscription mutation, so "re-sync on every change"
holds regardless of whether the change came from the inline TUI panel or the CLI:

- After `claudeconfig.Add`, `Rename`, `Delete`, `WriteAPIKey`, `WriteModelMappings`
  (added at the call sites in `internal/tui/claude_config_panel.go`,
  `internal/tui/claude_config_menu.go`, and `cmd/ghost-tab-tui/claude_config.go`).
- After the active-config switch (`persistClaudeConfig` in `internal/tui/mainmenu.go`).

A single private helper (e.g. `syncOpenCode()` in the TUI package) resolves `home` and
calls `opencodeconfig.Sync`, so each call site is one line. Sync errors are logged, never
surfaced as a blocking error — mirroring must not break Claude-side editing.

### Edge cases

| Case | Behavior |
| --- | --- |
| `opencode.json` missing | Create it with `$schema` + ghost-tab providers. |
| Existing `opencode.jsonc` / `config.json` | Write back to whichever exists (resolution order). |
| Existing file has comments (JSONC) | `MergeSubscriptions` returns `ok=false`; Sync logs a warning, writes nothing. |
| User-authored providers / other keys present | Preserved; only `ghost-tab-*` providers and (conditionally) `model` change. |
| Delete last subscription / switch to Standard | All `ghost-tab-*` providers removed; `model` left as-is. |
| Subscription with no API key / no models / unresolved base URL | Skipped (no provider written); unresolved base URL logs a warning. |
| MiMo subscription (endpoint not yet in table) | Skipped with warning until `ProviderBaseURLs["mimo"]` is filled. |

## Testing (TDD — tests first, watch them fail, then implement)

### Go unit tests — `internal/opencodeconfig/opencodeconfig_test.go`
1. Merge into nil/empty input → valid `opencode.json` with `$schema`, one `ghost-tab-*`
   provider, correct `npm`/`baseURL`/inline `apiKey`/`models`.
2. Preserve unrelated user `provider` entries and other top-level keys.
3. Idempotent rebuild: a stale `ghost-tab-old` provider is removed; current ones re-added.
4. Delete: subscription dropped from the set → its provider disappears.
5. Default `model` = active subscription's `"<id>/<opusModel>"`; **not** set when no
   subscription is active; left untouched when Standard active.
6. Unparseable existing bytes (`// comment`) → returns `ok=false` (no write).
7. Non-mirrorable subscriptions (no key / no models / no base URL) produce no provider.
8. `claudeconfig.ProviderBaseURL` returns the zhipu URL for a GLM-named config, `""` for
   an unknown provider name.

### Go integration test — `internal/opencodeconfig/sync_test.go`
9. `Sync(home)` end to end against a temp HOME: seed `claude-configs.list` + a config JSON
   (key + mappings) + active pointer → assert `~/.config/opencode/opencode.json` content.
   Includes the XDG_CONFIG_HOME path variant and the "preserve existing user file" variant.

### Bash integration test — `test/bash/` (only if a hook lands in bash)
10. If any `lib/*.sh` / CLI path is touched, assert that running the add/delete flow leaves
    `opencode.json` updated. (Expected to be Go-only; bash test added only if the hook is.)

`shellcheck` is run only if any `lib/*.sh` is modified (not anticipated).

## Out of scope (YAGNI)

- Capturing a per-subscription base URL via a new panel/CLI field (the "Add a base-URL
  field" option). Revisit if provider-name derivation proves too limiting.
- Cost/limit enrichment of the OpenCode `models` map from `internal/usage/pricing.go`
  (would give accurate spend/context display in OpenCode). Easy follow-up: add an exported
  `usage.RateFor` and populate `cost`/`limit`. Left out of v1 to keep the merge core simple.
- `{env:VAR}` / `{file:path}` apiKey indirection. Inline matches existing plaintext storage.
- Configuring Codex (only Claude Code and OpenCode are in scope).
