# Cross-Tool Feature Migration Design

Ghost Tab extends Claude Code-specific quality-of-life features (sound notification, tab spinner, status line) to Codex CLI and OpenCode using each tool's native configuration system.

## Scope

**Covered:**
- Codex CLI: sound notification, tab spinner (self-resetting workaround), status line (via `config.toml`)
- OpenCode: sound notification, tab spinner (via global TypeScript plugin)

**Skipped:**
- Copilot CLI: no global hooks location, no idle event — blocked upstream (#1128, #1157)
- OpenCode status line: no native API — blocked upstream (#8619)

## Feature support matrix

| Feature | Claude Code | Codex CLI | OpenCode | Copilot CLI |
|---------|------------|-----------|----------|-------------|
| Sound | Yes | Yes | Yes | No |
| Spinner | Yes | Yes | Yes | No |
| Status line setup | Yes (ccstatusline) | Yes (config.toml) | No | No |

## Codex CLI configuration

During setup, when the user selects Codex CLI, Ghost Tab creates or merges into `~/.codex/config.toml`.

### Sound notification

The `notify` top-level key runs a command fire-and-forget after each agent turn completes:

```toml
notify = ["afplay", "/System/Library/Sounds/Bottle.aiff"]
```

### Status line

The `tui.status_line` key replaces the default footer with selected items that mirror Claude Code's status line (model, git branch, context usage):

```toml
[tui]
status_line = ["model-with-reasoning", "git-branch", "context-remaining", "used-tokens"]
```

### Tab spinner

Codex has no "user submitted prompt" hook, so the spinner can't be stopped cleanly. Instead, the `notify` command points to a wrapper script that kills any running spinner before playing the sound. The spinner starts after the sound, then gets killed on the next `notify` call (next agent turn).

```toml
notify = ["bash", "~/.config/ghost-tab/codex-notify.sh"]
```

The script (`~/.config/ghost-tab/codex-notify.sh`):
1. Kill any existing spinner process (via PID file)
2. Play the notification sound (if enabled)
3. Start the spinner in background (if enabled)

If the user declined both sound and spinner during setup, `notify` is omitted entirely.

### Existing config handling

If `~/.codex/config.toml` already exists, Ghost Tab merges settings using Python's `tomllib`/`tomli_w`. Existing keys are preserved.

## OpenCode plugin

During setup, when the user selects OpenCode, Ghost Tab creates a global plugin at `~/.config/opencode/plugins/ghost-tab.ts`. Global plugins are auto-discovered — no registration needed.

### Sound notification

On `session.idle`, play the sound via `spawn` (non-blocking):

```typescript
if (event.type === "session.idle") {
  spawn("afplay", ["/System/Library/Sounds/Bottle.aiff"], { stdio: "ignore" })
}
```

### Tab spinner

On `session.idle`, start the spinner. On `session.status` with `type: "busy"`, stop it:

```typescript
if (event.type === "session.idle") {
  // kill previous spinner, start new one
}
if (event.type === "session.status" && event.properties?.status?.type === "busy") {
  // kill spinner, restore title
}
```

The spinner uses a PID file approach matching Claude Code's existing spinner scripts.

### Feature flags

The plugin reads `~/.config/ghost-tab/opencode-features.json` to know which features are enabled:

```json
{ "sound": true, "spinner": true }
```

Written during setup based on user choices. If both are disabled, the plugin file is not created.

## Setup flow changes

The sound and spinner prompts move outside the Claude-only gate and become available to all tools that support them. If the user selected only Copilot CLI, the prompts are skipped with a message.

If the user selected a mix of tools, features are configured for each supported tool independently.

## Files modified

| File | Changes |
|------|---------|
| `bin/ghost-tab` | Move sound/spinner prompts outside Claude-only gate. Add Codex config.toml creation/merge. Add OpenCode plugin creation. Add feature-support check per tool. Add Codex status line config section. |

## New files created during setup (on user's machine)

| File | When |
|------|------|
| `~/.codex/config.toml` | Codex selected + sound/spinner/status line enabled |
| `~/.config/ghost-tab/codex-notify.sh` | Codex selected + sound or spinner enabled |
| `~/.config/opencode/plugins/ghost-tab.ts` | OpenCode selected + sound or spinner enabled |
| `~/.config/ghost-tab/opencode-features.json` | OpenCode selected + sound or spinner enabled |

## Out of scope

- Copilot CLI features (blocked upstream)
- OpenCode status line (blocked upstream)
- Per-project configuration (global-only, matching current approach)
