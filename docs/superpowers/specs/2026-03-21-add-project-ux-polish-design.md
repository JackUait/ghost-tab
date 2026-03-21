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

**Storage:** `~/.config/ghost-tab/projects-root` — a single line containing an absolute (tilde-expanded) path. Follows the existing pattern of `~/.config/ghost-tab/ai-tool` and `~/.config/ghost-tab/terminal`. Tilde is expanded on write so the stored value is always an absolute path (e.g. `/Users/jack/Projects`, never `~/Projects`).

**Bash module:** `lib/settings-json.sh` is not used for this; the value is read/written as a plain file, consistent with other single-value settings. Two new helpers in `lib/projects.sh`:
- `get_projects_root` — reads the file; prints empty string if absent
- `set_projects_root <path>` — tilde-expands and writes the path; with an empty argument, removes the file

**UI — settings panel:** The settings panel is rendered in `renderSettingsBox()` inside `MainMenuModel` in `internal/tui/mainmenu.go`. It currently shows three items (Ghost Display, Tab Title, Sound) using left/right arrows to cycle values inline. The new item works differently (requires free-text input), so it uses a distinct interaction.

**Implementation note:** `updateSettings` in `mainmenu.go` currently hardcodes `n = 3` (the number of settings items) in five places for cursor wrapping, plus constructs exactly three items in `renderSettingsBox`. Adding the fourth item requires updating every one of these sites to `n = 4`. The implementer must search for all occurrences of this constant before completing the task.

- The item is listed as: `Default projects dir   ~/Projects` (right-aligned state text shows the expanded path, or `(not set)` when unset)
- **Enter** (not left/right arrows) opens an inline text input within the settings panel, pre-filled with the current value
- The input accepts any directory path with the same validation as the add-project path field (must exist and be a directory)
- **Esc** from the input returns to the settings panel (does not close the main menu)
- **Enter** on the input with a valid path saves and returns to the settings panel with updated state text
- **Enter** on an empty input removes the setting and returns to the settings panel with `(not set)` state text

---

### 2. Consolidated Add-Project Form

The standalone `ghost-tab-tui add-project` subcommand is deleted. All add-project logic moves into the main menu inline form. `cmd/ghost-tab-tui/add_project.go`, `internal/tui/input.go` (`ProjectInputModel`), and `lib/project-actions-tui.sh` are deleted; all callers are updated. The `add-project` subcommand registration is removed from `cmd/ghost-tab-tui/root.go`.

The main menu inline form is upgraded to a two-field form implemented entirely within `MainMenuModel` in `internal/tui/mainmenu.go` (specifically `enterInputMode`, `updateInputMode`, `submitInputMode`, and `renderInputBox`).

#### Fields

```
Add New Project

Path:  ~/Projects/█
Name:  ghost-tab  (auto)
```

**Path field** (focused first):
- Pre-filled with the default projects root directory if one is configured; otherwise empty.
- Cursor placed at the end of the pre-fill. User types the subdirectory name.
- Autocomplete (`PathSuggestionProvider`) operates from the current value.
- Tab on the path field: if autocomplete suggestions are visible, Tab accepts the highlighted suggestion (existing behaviour); only if no suggestions are visible does Tab advance focus to the name field.
- User may backspace to remove the pre-fill entirely and type any absolute or `~/`-relative path.

**Name field**:
- Starts empty. Auto-updates to `filepath.Base(expandedPath)` whenever the path changes, as long as the user has not manually edited the name. A dim `(auto)` suffix is shown while in auto-derive mode.
- Once the user edits the name field directly, auto-derive is locked off for the rest of that add session; the `(auto)` suffix disappears.
- Navigating away from the name field (Shift+Tab or Esc to path field) always clears any active soft-warn state so the warning does not persist if the user returns.
- Validation: must be non-empty after trim.

#### Navigation

| Key | Path field | Name field |
|-----|-----------|-----------|
| Tab | Accept autocomplete suggestion (if visible); otherwise move to name field | — |
| Shift+Tab | — | Return to path field; clears soft-warn state |
| Enter | Validate path; move to name field | Validate name; confirm (or show soft-warn on first duplicate-name Enter) |
| Esc | Cancel entire add | Return to path field; clears soft-warn state |
| Ctrl+C | Cancel | Cancel |

#### Validation (unified)

| Check | On | Behaviour |
|---|---|---|
| Path non-empty | Enter on path | Error inline: `"project path cannot be empty"` |
| Path exists and is directory | Enter on path | Error inline: `"path does not exist: {path}"` / `"path is not a directory: {path}"` |
| Duplicate path | Enter on name | Blocked: `"Project already exists"` |
| Name non-empty | Enter on name | Error inline: `"project name cannot be empty"` |
| Duplicate name | First Enter on name | Soft-warn: `"A project named 'X' already exists — press Enter again to add anyway"`. Second Enter confirms and saves. Navigating away clears the soft-warn state. |

---

### 3. Stale Path Warnings

When the main menu loads (or reloads) its project list, it tags each project as stale or not. The staleness check runs once at load time (not on every render tick) by calling `os.Stat` on each project's path from within `models.LoadProjects()` in `internal/models/project.go`. A `Stale bool` field on `models.Project` carries the result.

**In the project list**, stale projects display a warning marker rendered in muted yellow. The project name remains at normal brightness:

```
  ghost-tab        /Users/jack/Packages/ghost-tab
⚠ old-project      /Users/jack/deleted-dir
  web              /Users/jack/code/web
```

Tests for the stale marker strip ANSI escape codes before asserting on the rendered string (using `stripAnsi` or equivalent in the test package).

**On selection** of a stale project, `MainMenuModel` in `internal/tui/mainmenu.go` shows an inline confirmation instead of launching:

```
⚠ Path not found: /Users/jack/deleted-dir
  Launch anyway? [y/N]
```

Default is `N`. The user must press `y` explicitly to proceed (covers temporarily unmounted volumes). Pressing `n`, Enter, or any other key cancels and returns to the project list.

**No automatic removal.** The user removes stale projects via the existing delete flow.

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
| `internal/tui/mainmenu.go` | (1) Settings panel: new "Default projects dir" item with Enter-to-edit input. (2) Add-project: upgrade inline form to two-field (path + reactive name), pre-fill, Tab/Shift+Tab nav, unified validation, soft-warn. (3) Stale: staleness check on load, stale marker render, stale launch confirmation. |
| `internal/tui/projectfile.go` | Add `IsDuplicateName` alongside existing `IsDuplicateProject` |
| `internal/models/project.go` | Add `Stale bool` field; populate via `os.Stat` in `LoadProjects` |
| `cmd/ghost-tab-tui/add_project.go` | **Delete** |
| `internal/tui/input.go` | **Delete** `ProjectInputModel` (dead code once subcommand is gone). Move `GetPathSuggestions` and its tests to `internal/tui/autocomplete.go` / `test/internal/tui/autocomplete_test.go` before deleting the file. |
| `lib/project-actions-tui.sh` | **Delete** (callers updated — see below) |
| `cmd/ghost-tab-tui/root.go` | Remove `add-project` subcommand registration |
| `wrapper.sh` | Remove `project-actions-tui` from the lib sources array |
| `test/bash/entrypoints_test.go` | Remove `project-actions-tui.sh` from the expected-lib-files list |
| `test/bash/ai_select_test.go` | Delete the two tests that call `add_project_interactive` (subject no longer exists) |
| `test/bash/projects_test.go` | Tests for `get_projects_root` / `set_projects_root` |
| `test/internal/tui/mainmenu_test.go` | Tests for two-field form, settings input, stale confirmation |
| `test/internal/tui/projectfile_test.go` | Tests for `IsDuplicateName` |

---

## Testing

All changes follow the project's iron rule: test first, watch fail, then implement.

**New test cases required:**

*Bash (projects.sh):*
- `get_projects_root` returns empty string when `projects-root` file absent
- `get_projects_root` returns stored path when file present
- `set_projects_root` writes tilde-expanded absolute path (not tilde form) to file
- `set_projects_root` with empty argument removes the file

*Go (mainmenu — add-project form):*
- Path field pre-filled with default root when provided
- Name auto-fills from path basename while in auto-derive mode
- Name auto-derive locks when user edits name field directly; subsequent path changes do not overwrite name
- Tab with autocomplete visible accepts suggestion (does not advance focus)
- Tab without autocomplete advances focus to name field
- Shift+Tab from name field returns focus to path field
- Esc from name field returns focus to path field
- Duplicate name first Enter shows soft-warn message
- Duplicate name second Enter confirms and saves
- Navigating away from name field (Shift+Tab / Esc) clears soft-warn state
- `IsDuplicateName` returns true for exact name match, false otherwise

*Go (models — stale detection):*
- `LoadProjects` sets `Stale: true` when project path does not exist on disk
- `LoadProjects` sets `Stale: false` when project path exists

*Go (mainmenu — stale UX):*
- Stale project list renders `⚠` marker (tested with ANSI codes stripped)
- Non-stale project list does not render `⚠` marker
- Selecting stale project shows launch confirmation prompt
- Pressing `n` at stale confirmation cancels and returns to project list
- Pressing Enter at stale confirmation (default N) cancels and returns to project list
- Pressing `y` at stale confirmation proceeds to launch
