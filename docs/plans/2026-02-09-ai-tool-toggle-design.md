# AI Tool Toggle Design

Ghost Tab becomes AI-agnostic by letting users switch between Claude Code, Codex CLI (OpenAI), and OpenCode (anomalyco) from the project selector menu.

## Overview

A dedicated toggle row in the project selector lets users cycle between installed AI tools. The choice persists across sessions. During setup, users choose which tools to install.

## Supported tools

| Tool | Install command | Binary | Launch in tmux pane |
|------|----------------|--------|-------------------|
| Claude Code | `curl -fsSL https://claude.ai/install.sh \| bash` | `claude` | `claude` |
| Codex CLI | `brew install --cask codex` | `codex` | `codex --cd "$PROJECT_DIR"` |
| OpenCode | `brew install anomalyco/tap/opencode` | `opencode` | `opencode "$PROJECT_DIR"` |

## Setup flow

After installing core dependencies (tmux, lazygit, broot, Ghostty), a new step asks which AI tools to install using a multi-select prompt. At least one tool must be selected. Already-installed tools are pre-checked. The first selected tool becomes the default.

## Project selector UI

The toggle row sits between the project list and the shortcut legend:

```
  Ghost Tab v1.5.0

  > ~/Projects/my-app
    ~/Projects/other-project

  <  AI: Claude Code  >

  A: Add project  D: Delete project  O: Open once  P: Plain terminal
```

Arrow up/down navigates projects. When focus moves past the last project, it lands on the AI toggle row. Left/right arrows cycle through installed tools. The selection saves to `~/.config/ghostty/ghost-tab-ai-tool` as plain text (`claude`, `codex`, or `opencode`).

If only one tool is installed, the row displays as static text without arrows.

## Wrapper script changes

A dispatch function replaces the hardcoded Claude Code launch:

```bash
launch_ai_tool() {
    local tool="$1" project_dir="$2"
    case "$tool" in
        claude)   tmux send-keys -t "$SESSION:0.1" "claude" Enter ;;
        codex)    tmux send-keys -t "$SESSION:0.1" "codex --cd \"$project_dir\"" Enter ;;
        opencode) tmux send-keys -t "$SESSION:0.1" "opencode \"$project_dir\"" Enter ;;
    esac
}
```

Process cleanup stays generic -- it already walks the process tree from the tmux pane PID, so it handles any tool running in pane 1.

## Status line

Best-effort per tool:

- Claude Code: full status line (git info + context % + memory usage)
- Codex CLI / OpenCode: git info only (repo, branch, staged/unstaged/untracked)

The wrapper checks which tool is active and sources the appropriate status line script.

## Preference storage

File: `~/.config/ghostty/ghost-tab-ai-tool`
Content: single line with `claude`, `codex`, or `opencode`
Created during setup, updated by the toggle.

## Files modified

| File | Changes |
|------|---------|
| `bin/ghost-tab` | New setup step for AI tool installation. Multi-select prompt. Saves initial default. |
| `ghostty/claude-wrapper.sh` | AI toggle row in project selector. Preference read/write. Tool-aware launch function. Conditional status line. |

## Out of scope

- Per-project AI tool selection (global toggle is sufficient)
- Status line parity across tools
- Custom arguments/flags per tool (each tool has its own config system)
