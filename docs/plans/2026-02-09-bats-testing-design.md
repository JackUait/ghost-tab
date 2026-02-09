# Bats Testing Design

## Overview

Add comprehensive automated testing to Ghost Tab using bats-core. Extract pure/near-pure functions from the two main scripts into sourceable library files under `lib/`, then test them in isolation.

## Approach

**Extract-and-test pure functions.** The scripts are heavily side-effect-driven (installing packages, writing config files, launching tmux sessions). Rather than mocking all of that, we extract the testable logic into library files and test those directly. The remaining "glue" code stays untested but thin.

## Project Structure

```
ghost-tab/
  lib/
    ai-tools.sh        # ai_tool_display_name, ai_tool_color, cycle_ai_tool, validate_ai_tool
    projects.sh         # parse_project_line, load_projects, path_expand, path_truncate
    process.sh          # kill_tree
    input.sh            # parse_esc_sequence (SGR mouse parsing)
    statusline.sh       # format_memory, parse_cwd_from_json
    setup.sh            # resolve_share_dir
  test/
    test_helper/
      bats-core/        # git submodule
      bats-assert/      # git submodule
      bats-support/     # git submodule
      common.bash       # shared setup (load assert + support)
    ai-tools.bats
    projects.bats
    process.bats
    input.bats
    statusline.bats
    setup.bats
  run-tests.sh
```

## Extracted Functions

### lib/ai-tools.sh

From `claude-wrapper.sh`:

- `ai_tool_display_name <id>` -- converts tool ID to display name ("claude" -> "Claude Code")
- `ai_tool_color <id>` -- returns ANSI color escape for a tool
- `cycle_ai_tool <direction>` -- cycles SELECTED_AI_TOOL through AI_TOOLS_AVAILABLE array (next/prev with wrapping)
- `validate_ai_tool` -- checks SELECTED_AI_TOOL against AI_TOOLS_AVAILABLE, falls back to first available if invalid

### lib/projects.sh

From `claude-wrapper.sh`:

- `parse_project_name <line>` -- extracts name from "name:path" format
- `parse_project_path <line>` -- extracts path from "name:path" format
- `load_projects <file>` -- reads project file, skips blanks and comments, populates array
- `path_expand <path>` -- expands ~ to $HOME
- `path_truncate <path> <max_width>` -- truncates long paths with ... in the middle

### lib/process.sh

From `claude-wrapper.sh`:

- `kill_tree <pid> [signal]` -- recursively kills a process and all its children (depth-first)

### lib/input.sh

From `claude-wrapper.sh`:

- `parse_esc_sequence` -- reads and parses escape sequences from stdin. Returns arrow key direction or "click:ROW" for SGR mouse events. Ignores mouse release and non-left-click.

### lib/statusline.sh

From `bin/ghost-tab` (statusline-wrapper.sh template):

- `format_memory <kb>` -- converts kilobytes to human-readable (e.g., 512000 -> "500M", 1572864 -> "1.5G")
- `parse_cwd_from_json <input>` -- extracts current_dir value from JSON string

### lib/setup.sh

From `bin/ghost-tab`:

- `resolve_share_dir <script_dir>` -- determines SHARE_DIR: Homebrew share path when installed via brew, otherwise relative ../ from script

## Test Cases (~25-30 total)

### ai-tools.bats

- ai_tool_display_name maps each known ID correctly
- ai_tool_display_name passes through unknown input as-is
- ai_tool_color returns correct ANSI escape per known tool
- cycle_ai_tool next wraps from last to first
- cycle_ai_tool prev wraps from first to last
- cycle_ai_tool is a no-op with single tool
- validate_ai_tool keeps valid saved preference
- validate_ai_tool falls back to first available when pref is invalid

### projects.bats

- parse_project_name extracts name from "name:path"
- parse_project_path extracts path from "name:path"
- Handles paths containing colons
- load_projects skips blank lines
- load_projects skips comment lines
- path_expand converts ~/foo to $HOME/foo
- path_expand leaves absolute paths unchanged
- path_truncate returns short paths unchanged
- path_truncate inserts ... in middle for long paths
- path_truncate respects max-width parameter

### process.bats

- kill_tree kills parent and children (real sleep subprocesses)
- kill_tree handles nonexistent PIDs gracefully

### input.bats

- Parses up arrow escape sequence
- Parses down arrow escape sequence
- Parses SGR mouse left click with correct row
- Ignores mouse release events
- Ignores non-left-click buttons

### statusline.bats

- format_memory converts KB to MB display
- format_memory converts KB to GB display with decimal
- parse_cwd_from_json extracts path from valid JSON

### setup.bats

- resolve_share_dir returns brew share path when in brew prefix
- resolve_share_dir falls back to relative path

## Main Script Changes

### claude-wrapper.sh

- Sources lib files from its own directory: `source "$(dirname "$0")/lib/ai-tools.sh"` etc.
- Extracted functions replaced with calls to sourced functions
- Behavior is identical

### bin/ghost-tab

- Copies `lib/` directory to `~/.config/ghostty/lib/` alongside the wrapper script
- Sources `lib/setup.sh` for resolve_share_dir

## Test Infrastructure

- bats-core, bats-assert, bats-support as git submodules under test/test_helper/
- `run-tests.sh` at repo root invokes bats on all test files
- Each test uses bats setup()/teardown() with temp directories for filesystem tests
- No mocking framework -- pure function extraction eliminates the need
- No CI/CD initially -- run locally with `./run-tests.sh`

## Implementation Order

1. Add bats-core git submodules
2. Create lib/ files by extracting functions from main scripts
3. Update main scripts to source lib/
4. Update bin/ghost-tab to copy lib/ during setup
5. Write test files
6. Verify all tests pass and scripts still work
