# Add-Project UX Polish — Design Spec

**Date:** 2026-03-21
**Scope:** Tier A polish of the existing add-project flow. No new metadata fields, no frecency ranking, no scan-directory import.

---

## Problem

The current add-project UX has three concrete problems:

1. **Two inconsistent flows.** The main menu inline mode and the standalone `ghost-tab-tui add-project` subcommand collect different fields, apply different validation, and derive the project name differently. A user who adds a project one way gets different behaviour than the other.

2. **No smart path pre-fill.** Users must type the full path every time. Developers who keep all projects under a common root (e.g. `~/Projects/`) retype that prefix on every add.

3. **Stale paths fail silently.** When a project's directory is deleted or moved, the project stays in the list with no warning. Selecting it fails with a raw shell error (`cd: No such file or directory`) rather than a helpful Ghost Tab message.

---

## Design

### 1. Settings: Default Projects Directory

A new optional setting stores a root directory that is pre-filled into the path field whenever the user adds a project.

**Storage:** `~/.config/ghost-tab/projects-root` — a single line containing an absolute path (tilde-expanded on write). Follows the existing pattern of `~/.config/ghost-tab/ai-tool` and `~/.config/ghost-tab/terminal`.

**UI:** A new action item in the existing settings menu: `"Set default projects directory"`. Selecting it opens an inline text input pre-filled with the current value (or empty if unset). Validation: path must exist and be a directory. Clearing the field and confirming removes the file (no default going forward).

**Bash module:** `lib/settings-json.sh` is not used for this; the value is read/written as a plain file, consistent with other single-value settings. A new helper `get_projects_root` / `set_projects_root` lives in `lib/projects.sh`.

---

### 2. Consolidated Add-Project Form

The standalone `ghost-tab-tui add-project` subcommand and the main menu inline add mode are unified into a single two-field inline form rendered within the main menu.

#### Fields

```
Add New Project

Path:  ~/Projects/█
Name:  ghost-tab  (auto)
```

**Path field** (focused first):
- Pre-filled with the default projects root directory if one is configured; otherwise empty.
- Cursor placed at the end of the pre-fill. User types the subdirectory name.
- Autocomplete (existing `PathSuggestionProvider`) operates from the current value.
- User may backspace to remove the pre-fill entirely and type any absolute or `~/`-relative path.

**Name field**:
- Starts empty.
- Auto-updates to `filepath.Base(expandedPath)` whenever the path changes, as long as the user has not manually edited the name. A dim `(auto)` suffix is shown while in auto-derive mode.
- Once the user edits the name field directly, auto-derive is disabled for the rest of that add session. The `(auto)` suffix disappears.
- Validation: must be non-empty after trim.

#### Navigation

| Key | Path field | Name field |
|-----|-----------|-----------|
| Tab | Move to name field | — |
| Shift+Tab | — | Return to path field |
| Enter | Move to name field (if path valid) | Confirm (if name valid) |
| Esc | Cancel entire add | Return to path field |
| Ctrl+C | Cancel | Cancel |

#### Validation (unified, applied in both entry points)

| Check | On | Behaviour |
|---|---|---|
| Path non-empty | Enter on path | Error inline: `"project path cannot be empty"` |
| Path exists and is directory | Enter on path | Error inline: `"path does not exist: {path}"` / `"path is not a directory: {path}"` |
| Duplicate path | Enter on name | Blocked: `"Project already exists"` |
| Name non-empty | Enter on name | Error inline: `"project name cannot be empty"` |
| Duplicate name | Enter on name | Warning (soft): `"A project named 'X' already exists — press Enter again to add anyway"`. Second Enter confirms. |

#### Removal of standalone subcommand

`cmd/ghost-tab-tui/add_project.go` and `lib/project-actions-tui.sh` are removed. All callers are updated to use the main menu flow. The `ghost-tab-tui add-project` subcommand is deleted.

---

### 3. Stale Path Warnings

When the main menu loads (or reloads) its project list, it checks each project path for existence using `os.Stat`. This check is fast (local filesystem) and happens once per load, not on every render tick.

**In the project list**, projects with non-existent paths display a warning marker:

```
  ghost-tab        /Users/jack/Packages/ghost-tab
⚠ old-project      /Users/jack/deleted-dir
  web              /Users/jack/code/web
```

The `⚠` marker and the path text are rendered in a muted yellow colour. The project name remains at normal brightness to keep the list scannable.

**On selection** of a stale project, instead of launching and failing at `cd` time, Ghost Tab shows an inline confirmation:

```
⚠ Path not found: /Users/jack/deleted-dir
  Launch anyway? [y/N]
```

Default is `N`. The user must press `y` explicitly to proceed. This covers the case of temporarily unmounted volumes.

**No automatic removal.** The user removes stale projects via the existing delete flow. No new "clean up stale projects" command is added.

---

## What Is Not In Scope

- Project metadata (description, tags, AI tool preference per project)
- Frecency-based project ranking
- Scan-directory-and-import mode
- Clipboard detection for path pre-fill
- Recent directories suggestions
- Git repo detection or project-type inference

---

## Files Affected

| File | Change |
|---|---|
| `lib/projects.sh` | Add `get_projects_root` / `set_projects_root` helpers |
| `lib/settings-menu-tui.sh` | Add "Set default projects directory" action item |
| `internal/tui/mainmenu.go` | Upgrade inline add-project to two-field form with pre-fill, reactive name, back-nav, unified validation |
| `internal/tui/input.go` | Extend `ProjectInputModel`: name auto-derive with lock flag, Shift+Tab back-nav, duplicate name soft-warn |
| `internal/tui/projectfile.go` | `IsDuplicateName` check alongside existing `IsDuplicateProject` |
| `internal/tui/projects.go` | Stale path detection on load; stale marker in render; launch confirmation for stale |
| `internal/models/project.go` | Add `Stale bool` field populated at load time |
| `cmd/ghost-tab-tui/add_project.go` | **Delete** |
| `lib/project-actions-tui.sh` | **Delete** (callers updated) |
| `cmd/ghost-tab-tui/root.go` | Remove `add-project` subcommand registration |
| `internal/tui/settings_menu.go` (or equivalent) | Add default projects root setting action |
| `test/bash/projects_test.go` | Tests for `get_projects_root` / `set_projects_root` |
| `test/internal/tui/input_test.go` | Tests for name auto-derive, lock flag, Shift+Tab, duplicate name soft-warn |
| `test/internal/tui/projects_test.go` | Tests for stale detection, stale marker rendering, stale launch confirmation |

---

## Testing

All changes follow the project's iron rule: test first, watch fail, then implement.

**New test cases required:**
- `get_projects_root` returns empty string when file absent
- `get_projects_root` returns stored path when file present
- `set_projects_root` writes file; clears file on empty input
- `ProjectInputModel`: name auto-fills from path basename
- `ProjectInputModel`: name auto-derive locks when user edits name field
- `ProjectInputModel`: Shift+Tab from name returns focus to path
- `ProjectInputModel`: duplicate name shows soft warning, second Enter confirms
- `ProjectInputModel`: path field pre-filled with default root when provided
- `IsDuplicateName`: returns true for exact name match, false otherwise
- Project list: stale projects have `Stale: true` after load when path absent
- Project list: stale marker rendered for stale projects
- Selection of stale project: confirmation prompt shown, `n` cancels, `y` proceeds
