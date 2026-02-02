# Setup Script Design

## Overview

Interactive setup script (`setup.sh`) that automates installation and configuration of vibecode-editor for new users.

## Flow

1. **Check OS** — verify macOS, exit with message on other platforms
2. **Install Homebrew** — if missing, install via official script
3. **Install dependencies** — `tmux`, `lazygit`, `broot` via `brew install`, skip already-installed
4. **Check Claude Code** — verify `claude` in PATH, if not prompt to install (`npm install -g @anthropic-ai/claude-code`) and exit with instructions (requires auth)
5. **Check Ghostty** — verify installed, warn if not but continue setup
6. **Copy wrapper script** — copy `claude-wrapper.sh` to `~/.config/ghostty/`, make executable
7. **Handle Ghostty config** — if existing config, ask user: merge or backup+replace. If no config, copy it
8. **Interactive project input** — loop asking for name/path pairs, write to `~/.config/vibecode-editor/projects`
9. **Print summary** — what was installed/configured, next steps

## Ghostty Config Merge Logic

- Existing config with `command =` line: replace that line with wrapper path
- Existing config without `command =`: append it
- Backup+replace: copy to `config.backup.<timestamp>`

## Project Input Loop

- Prompt for name and path pairs in a loop
- Expand `~` to `$HOME` when validating
- Warn but allow if path doesn't exist yet
- Skip entirely if user declines first prompt

## Output Style

- Green (`✓`) for success
- Yellow (`!`) for warnings
- Red (`✗`) for errors
- Blue for info/prompts
- ANSI color codes with fallback to plain text

## Error Handling

- Each step prints status with colored prefix
- Brew install failure: warn and continue
- Missing Ghostty: warn, still set up config files
