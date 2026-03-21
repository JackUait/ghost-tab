# Add-Project UX Polish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Polish the add-project UX: consolidate two inconsistent flows into one two-field inline form with reactive name derivation, add a configurable default projects root directory setting, and warn on stale project paths instead of failing silently.

**Architecture:** All new add-project logic lives in `MainMenuModel` in `internal/tui/mainmenu.go`. The standalone `ghost-tab-tui add-project` subcommand and its bash wrapper are deleted. A `Stale bool` field is added to `models.Project` and populated in `LoadProjects`. A new `projects-root` config file stores the default path pre-fill.

**Tech Stack:** Go 1.21+, Bubbletea, Lipgloss, Cobra; Bash; Go test framework.

**Spec:** `docs/superpowers/specs/2026-03-21-add-project-ux-polish-design.md`

---

## File Map

| File | Action | Purpose |
|---|---|---|
| `lib/projects.sh` | Modify | Add `get_projects_root` / `set_projects_root` |
| `test/bash/projects_test.go` | Modify | Tests for new bash helpers |
| `internal/tui/projectfile.go` | Modify | Add `IsDuplicateName` |
| `test/internal/tui/projectfile_test.go` | Modify | Tests for `IsDuplicateName` |
| `internal/models/project.go` | Modify | Add `Stale bool` to `Project`; populate in `LoadProjects` |
| `internal/tui/autocomplete.go` | Modify | Add `GetPathSuggestions` (migrated from `input.go`) |
| `test/internal/tui/input_test.go` | Modify | Remove only `TestProjectInput_*` tests (dead code); keep all others |
| `internal/tui/input.go` | **Delete** | `ProjectInputModel` and `GetPathSuggestions` moved out |
| `internal/models/project_test.go` | Create | Tests for `LoadProjects` stale field (models layer) |
| `cmd/ghost-tab-tui/add_project.go` | **Delete** | Standalone subcommand deleted |
| `cmd/ghost-tab-tui/root.go` | Modify | Remove `add-project` subcommand registration |
| `lib/project-actions-tui.sh` | **Delete** | Bash wrapper for deleted subcommand |
| `wrapper.sh` | Modify | Remove `project-actions-tui` from lib array |
| `test/bash/entrypoints_test.go` | Modify | Remove `project-actions-tui.sh` from expected-lib-files list |
| `test/bash/ai_select_test.go` | Modify | Delete two tests that call `add_project_interactive` |
| `internal/tui/mainmenu.go` | Modify | (1) Two-field add-project form. (2) Projects root pre-fill. (3) Settings panel item. (4) Stale path UI. |
| `test/internal/tui/mainmenu_test.go` | Modify | Tests for all mainmenu.go changes |

---

## Task 1: Bash helpers — get_projects_root / set_projects_root

**Files:**
- Modify: `lib/projects.sh`
- Modify: `test/bash/projects_test.go`

The config file `~/.config/ghost-tab/projects-root` stores a single absolute path (tilde-expanded on write). `get_projects_root` reads it; `set_projects_root` writes (or removes) it.

- [ ] **Step 1: Write the failing tests**

Add to `test/bash/projects_test.go`:

```go
func TestGetProjectsRoot_AbsentFile(t *testing.T) {
	dir := t.TempDir()
	out, code := runBashFunc(t, "lib/projects.sh", "get_projects_root",
		[]string{filepath.Join(dir, "projects-root")}, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty output when file absent, got %q", out)
	}
}

func TestGetProjectsRoot_ReturnsStoredPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects-root")
	os.WriteFile(file, []byte("/Users/jack/Projects\n"), 0644)
	out, code := runBashFunc(t, "lib/projects.sh", "get_projects_root",
		[]string{file}, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "/Users/jack/Projects" {
		t.Errorf("got %q", out)
	}
}

func TestSetProjectsRoot_WritesExpandedPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects-root")
	home := os.Getenv("HOME")
	// Pass a tilde path; expect it to be expanded on write
	_, code := runBashFunc(t, "lib/projects.sh", "set_projects_root",
		[]string{file, "~/Projects"}, nil)
	assertExitCode(t, code, 0)
	data, _ := os.ReadFile(file)
	got := strings.TrimSpace(string(data))
	want := home + "/Projects"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSetProjectsRoot_EmptyArgRemovesFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects-root")
	os.WriteFile(file, []byte("/Users/jack/Projects\n"), 0644)
	_, code := runBashFunc(t, "lib/projects.sh", "set_projects_root",
		[]string{file, ""}, nil)
	assertExitCode(t, code, 0)
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Error("expected file to be removed, but it still exists")
	}
}
```

- [ ] **Step 2: Run tests — expect FAIL**

```bash
go test ./test/bash/... -run "TestGetProjectsRoot|TestSetProjectsRoot" -v
```

Expected: FAIL — functions not defined.

- [ ] **Step 3: Implement in lib/projects.sh**

Add to `lib/projects.sh` after `path_expand()`:

```bash
# get_projects_root <file> — prints the stored projects root, or empty string.
get_projects_root() {
  local file="$1"
  [ -f "$file" ] || return 0
  cat "$file"
}

# set_projects_root <file> <path> — writes tilde-expanded path; removes file if path is empty.
set_projects_root() {
  local file="$1"
  local path="$2"
  if [ -z "$path" ]; then
    rm -f "$file"
    return 0
  fi
  local expanded
  expanded="$(path_expand "$path")"
  printf '%s\n' "$expanded" > "$file"
}
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
go test ./test/bash/... -run "TestGetProjectsRoot|TestSetProjectsRoot" -v
```

- [ ] **Step 5: Run shellcheck**

```bash
shellcheck lib/projects.sh
```

- [ ] **Step 6: Commit**

```bash
git add lib/projects.sh test/bash/projects_test.go
git commit -m "feat: add get_projects_root / set_projects_root bash helpers"
```

---

## Task 2: IsDuplicateName in projectfile.go

**Files:**
- Modify: `internal/tui/projectfile.go`
- Modify: `test/internal/tui/projectfile_test.go`

- [ ] **Step 1: Write the failing test**

Add to `test/internal/tui/projectfile_test.go`:

```go
func TestIsDuplicateName(t *testing.T) {
	projects := []models.Project{
		{Name: "ghost-tab", Path: "/path/a"},
		{Name: "web", Path: "/path/b"},
	}

	tests := []struct {
		name string
		want bool
	}{
		{"ghost-tab", true},
		{"web", true},
		{"api", false},
		{"Ghost-Tab", false}, // case-sensitive
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tui.IsDuplicateName(tt.name, projects)
			if got != tt.want {
				t.Errorf("IsDuplicateName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
go test ./test/internal/tui/... -run TestIsDuplicateName -v
```

- [ ] **Step 3: Implement in projectfile.go**

Add after `IsDuplicateProject`:

```go
// IsDuplicateName checks if a project name already exists in the project list (exact match).
func IsDuplicateName(name string, projects []models.Project) bool {
	for _, p := range projects {
		if p.Name == name {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run test — expect PASS**

```bash
go test ./test/internal/tui/... -run TestIsDuplicateName -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/tui/projectfile.go test/internal/tui/projectfile_test.go
git commit -m "feat: add IsDuplicateName to projectfile"
```

---

## Task 3: Stale bool in LoadProjects

**Files:**
- Modify: `internal/models/project.go`
- Create: `internal/models/project_test.go`

`LoadProjects` already populates `Name` and `Path`. Add `Stale bool` and set it via `os.Stat`. Paths in the file are always absolute (expanded on write), so `os.Stat(path)` works directly.

Tests live next to the model, not in the tui test directory.

- [ ] **Step 1: Write the failing tests**

Create `internal/models/project_test.go`:

```go
package models_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jackuait/ghost-tab/internal/models"
)

func TestLoadProjects_StaleField_ExistingPath(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "myproject")
	os.MkdirAll(realDir, 0755)
	file := filepath.Join(dir, "projects")
	os.WriteFile(file, []byte("myproject:"+realDir+"\n"), 0644)

	projects, err := models.LoadProjects(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Stale {
		t.Error("expected Stale=false for existing path")
	}
}

func TestLoadProjects_StaleField_MissingPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects")
	os.WriteFile(file, []byte("ghost:/nonexistent/path/xyz\n"), 0644)

	projects, err := models.LoadProjects(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if !projects[0].Stale {
		t.Error("expected Stale=true for missing path")
	}
}
```

- [ ] **Step 2: Run tests — expect FAIL**

```bash
go test ./test/internal/tui/... -run "TestLoadProjects_StaleField" -v
```

Expected: compile error — `Stale` field doesn't exist yet.

- [ ] **Step 3: Add Stale field and populate in LoadProjects**

In `internal/models/project.go`, update the `Project` struct:

```go
type Project struct {
	Name      string
	Path      string
	Worktrees []Worktree
	Stale     bool
}
```

In `LoadProjects`, update the `append` call:

```go
path := strings.TrimSpace(parts[1])
_, statErr := os.Stat(path)
projects = append(projects, Project{
    Name:  strings.TrimSpace(parts[0]),
    Path:  path,
    Stale: statErr != nil,
})
```

Add `"os"` to the imports if not already present.

- [ ] **Step 4: Run tests — expect PASS**

```bash
go test ./test/internal/tui/... -run "TestLoadProjects_StaleField" -v
```

- [ ] **Step 5: Run full suite to check nothing broke**

```bash
go test ./... -count=1
```

- [ ] **Step 6: Commit**

```bash
git add internal/models/project.go internal/models/project_test.go
git commit -m "feat: add Stale field to Project, populated in LoadProjects"
```

---

## Task 4: Delete dead code — input.go, add_project.go, project-actions-tui.sh

This task removes the standalone add-project subcommand and all dead code. `input.go` contains `ProjectInputModel` (dead) and `GetPathSuggestions` (live, used by `input_test.go`). The migration order is: move `GetPathSuggestions` to `autocomplete.go` first, remove only the `TestProjectInput_*` tests from `input_test.go`, then delete `input.go`.

**Files:**
- Modify: `internal/tui/autocomplete.go` — add `GetPathSuggestions`
- Modify: `test/internal/tui/input_test.go` — remove only `TestProjectInput_*` tests
- Delete: `internal/tui/input.go`
- Delete: `cmd/ghost-tab-tui/add_project.go`
- Delete: `lib/project-actions-tui.sh`
- Modify: `cmd/ghost-tab-tui/root.go` — remove `addProjectCmd`
- Modify: `wrapper.sh` — remove `project-actions-tui` from lib array
- Modify: `test/bash/entrypoints_test.go` — remove `project-actions-tui.sh` from expected list
- Modify: `test/bash/ai_select_test.go` — delete two tests that source the deleted lib

- [ ] **Step 1: Migrate GetPathSuggestions to autocomplete.go**

`GetPathSuggestions` in `input.go` is a thin wrapper around `PathSuggestionProvider`. Add it to `internal/tui/autocomplete.go`:

```go
// GetPathSuggestions is a convenience wrapper around PathSuggestionProvider.
// Returns nil for empty input (unlike PathSuggestionProvider which defaults to ~/).
func GetPathSuggestions(input string) []string {
	if input == "" {
		return nil
	}
	return PathSuggestionProvider(8)(input)
}
```

- [ ] **Step 2: Verify build still passes**

```bash
go build ./...
```

- [ ] **Step 3: Remove TestProjectInput_* tests from input_test.go**

Delete only the test functions that start with `TestProjectInput_` from `test/internal/tui/input_test.go`. Keep `TestConfirmModel`, `TestPathSuggestions_*`, and `TestConfirmDialog_*` — these test live code.

Run to confirm remaining tests still pass:
```bash
go test ./test/internal/tui/... -run "TestConfirmModel|TestPathSuggestions|TestConfirmDialog" -v
```

- [ ] **Step 4: Delete dead Go files**

```bash
rm internal/tui/input.go cmd/ghost-tab-tui/add_project.go
```

- [ ] **Step 5: Remove subcommand registration from root.go**

In `cmd/ghost-tab-tui/root.go`, find and remove the line that registers `addProjectCmd`:

```bash
grep -n "addProject\|add-project\|add_project" cmd/ghost-tab-tui/root.go
```

Remove the `rootCmd.AddCommand(addProjectCmd)` line (and any `var addProjectCmd` declaration if it exists in root.go rather than the deleted file).

- [ ] **Step 6: Verify build**

```bash
go build ./...
```

Fix any compile errors (import cycles, missing references).

- [ ] **Step 7: Delete bash wrapper**

```bash
rm lib/project-actions-tui.sh
```

- [ ] **Step 8: Remove from wrapper.sh lib array**

In `wrapper.sh` line 42, remove `project-actions-tui` from the `_gt_libs` array:

```bash
_gt_libs=(ai-tools projects process input tui menu-tui project-actions tmux-session settings-json notification-setup tab-title-watcher)
```

- [ ] **Step 9: Update entrypoints_test.go**

In `test/bash/entrypoints_test.go` line 579, remove `"project-actions-tui.sh"` from the expected lib files string slice.

- [ ] **Step 10: Delete add_project_interactive tests in ai_select_test.go**

Delete the two test functions in `test/bash/ai_select_test.go` that call `add_project_interactive`:
- `TestProjectActions_AddProjectInteractiveReturnsSuccess` (around line 450)
- `TestProjectActions_AddProjectInteractiveReturnsFailureWhenCancelled` (around line 510)

- [ ] **Step 11: Run full test suite — expect PASS**

```bash
go test ./... -count=1
```

- [ ] **Step 12: Run shellcheck on wrapper.sh**

```bash
shellcheck wrapper.sh
```

- [ ] **Step 13: Commit**

```bash
git add -u internal/tui/autocomplete.go
git commit -m "feat: delete standalone add-project subcommand and bash wrapper"
```

---

## Task 5: Two-field add-project form in mainmenu.go

**Files:**
- Modify: `internal/tui/mainmenu.go`
- Modify: `test/internal/tui/mainmenu_test.go`

The current inline add-project mode has a single path field. This task upgrades it to a two-field form: path (focused first) + name (auto-derived, editable). The upgrade modifies `MainMenuModel` struct fields and the `enterInputMode` / `updateInputMode` / `submitInputMode` / `renderInputBox` methods.

### New fields to add to MainMenuModel struct

Add after the existing `inputErr error` field (around line 172):

```go
// Two-field add-project form state
nameInput      textinput.Model
nameTouched    bool  // user manually edited name; disable auto-derive
nameErr        error
nameWarnShown  bool  // true after first Enter on duplicate name; second Enter confirms
inputFocusPath bool  // true = path field focused, false = name field focused
```

### enterInputMode changes

The existing `enterInputMode("add-project")` creates the path input. Add name input initialization:

```go
func (m *MainMenuModel) enterInputMode(mode string) (tea.Model, tea.Cmd) {
	m.inputMode = mode
	m.inputErr = nil
	m.nameErr = nil
	m.nameTouched = false
	m.nameWarnShown = false
	m.inputFocusPath = true

	ti := textinput.New()
	ti.Placeholder = "Project path (e.g., ~/code/project)"
	ti.Width = menuContentWidth - 11
	ti.Focus()
	m.pathInput = ti

	ni := textinput.New()
	ni.Placeholder = "Project name"
	ni.Width = menuContentWidth - 11
	m.nameInput = ni

	m.autocomplete = NewAutocomplete(PathSuggestionProvider(8), 8)
	return m, textinput.Blink
}
```

### exitInputMode changes

```go
func (m *MainMenuModel) exitInputMode() {
	m.inputMode = ""
	m.inputErr = nil
	m.nameErr = nil
	m.nameTouched = false
	m.nameWarnShown = false
	m.inputFocusPath = true
	m.pathInput.Blur()
	m.nameInput.Blur()
	m.autocomplete.Dismiss()
}
```

### updateInputMode changes

Replace the existing `updateInputMode`. Key rules:
- While `inputFocusPath`: path field is active; Tab/Enter on autocomplete accepts suggestion; Enter without autocomplete validates path and advances to name field; Esc cancels.
- While `!inputFocusPath` (name field active): Shift+Tab returns to path; Esc returns to path (clears soft-warn); Enter validates name and submits.
- After path changes and `!nameTouched`: auto-update `nameInput` value to `filepath.Base(expanded)`.

```go
func (m *MainMenuModel) updateInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputFocusPath {
		return m.updateInputModePath(msg)
	}
	return m.updateInputModeName(msg)
}

func (m *MainMenuModel) updateInputModePath(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		if m.autocomplete.ShowSuggestions() {
			m.autocomplete.Dismiss()
			return m, nil
		}
		m.exitInputMode()
		return m, nil
	case tea.KeyCtrlC:
		m.exitInputMode()
		m.setActionResult("quit")
		return m, tea.Quit
	case tea.KeyUp:
		if m.autocomplete.ShowSuggestions() {
			m.autocomplete.MoveUp()
			return m, nil
		}
	case tea.KeyDown:
		if m.autocomplete.ShowSuggestions() {
			m.autocomplete.MoveDown()
			return m, nil
		}
	case tea.KeyTab:
		if m.autocomplete.ShowSuggestions() && len(m.autocomplete.Suggestions()) > 0 {
			accepted := m.autocomplete.AcceptSelected()
			m.pathInput.SetValue(accepted)
			m.autocomplete.SetInput(accepted)
			m.autocomplete.RefreshSuggestions()
			m.maybeAutoDeriveName()
			return m, nil
		}
		// No suggestions: advance to name field
		return m.advanceToNameField()
	case tea.KeyEnter:
		if m.autocomplete.ShowSuggestions() && len(m.autocomplete.Suggestions()) > 0 {
			accepted := m.autocomplete.AcceptSelected()
			m.pathInput.SetValue(accepted)
			m.autocomplete.SetInput(accepted)
			m.autocomplete.RefreshSuggestions()
			m.maybeAutoDeriveName()
			return m, nil
		}
		return m.advanceToNameField()
	}

	var cmd tea.Cmd
	m.pathInput, cmd = m.pathInput.Update(msg)
	current := m.pathInput.Value()
	if current != "" {
		m.autocomplete.SetInput(current)
		m.autocomplete.RefreshSuggestions()
	} else {
		m.autocomplete.Dismiss()
	}
	m.maybeAutoDeriveName()
	return m, cmd
}

// advanceToNameField validates path and moves focus to name field.
func (m *MainMenuModel) advanceToNameField() (tea.Model, tea.Cmd) {
	path := strings.TrimSpace(m.pathInput.Value())
	if path == "" {
		m.inputErr = fmt.Errorf("project path cannot be empty")
		return m, nil
	}
	if err := util.ValidatePath(path); err != nil {
		m.inputErr = err
		return m, nil
	}
	m.inputErr = nil
	m.inputFocusPath = false
	m.pathInput.Blur()
	m.nameInput.Focus()
	m.maybeAutoDeriveName()
	return m, textinput.Blink
}

// maybeAutoDeriveName sets nameInput value from path basename if user hasn't touched it.
func (m *MainMenuModel) maybeAutoDeriveName() {
	if m.nameTouched {
		return
	}
	path := strings.TrimSpace(m.pathInput.Value())
	if path == "" {
		return
	}
	expanded := filepath.Clean(util.ExpandPath(path))
	base := filepath.Base(expanded)
	if base != "" && base != "." && base != "/" {
		m.nameInput.SetValue(base)
	}
}

func (m *MainMenuModel) updateInputModeName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyShiftTab:
		// Return to path field; clear soft-warn
		m.nameWarnShown = false
		m.nameErr = nil
		m.inputFocusPath = true
		m.nameInput.Blur()
		m.pathInput.Focus()
		return m, textinput.Blink
	case tea.KeyCtrlC:
		m.exitInputMode()
		m.setActionResult("quit")
		return m, tea.Quit
	case tea.KeyEnter:
		return m.submitInputMode()
	}

	var cmd tea.Cmd
	prev := m.nameInput.Value()
	m.nameInput, cmd = m.nameInput.Update(msg)
	if m.nameInput.Value() != prev {
		m.nameTouched = true
		m.nameWarnShown = false // reset warn if user edits after warn
		m.nameErr = nil
	}
	return m, cmd
}
```

### submitInputMode changes

```go
func (m *MainMenuModel) submitInputMode() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.nameInput.Value())
	if name == "" {
		m.nameErr = fmt.Errorf("project name cannot be empty")
		return m, nil
	}

	path := strings.TrimSpace(m.pathInput.Value())
	expanded := filepath.Clean(util.ExpandPath(path))

	if IsDuplicateProject(expanded, m.projects) {
		m.nameErr = fmt.Errorf("Project already exists")
		return m, nil
	}

	if IsDuplicateName(name, m.projects) {
		if !m.nameWarnShown {
			m.nameWarnShown = true
			m.nameErr = fmt.Errorf("A project named '%s' already exists — press Enter again to add anyway", name)
			return m, nil
		}
		// Second Enter: user confirmed, proceed
	}
	m.nameWarnShown = false

	if err := AppendProject(name, expanded, m.projectsFile); err != nil {
		m.nameErr = fmt.Errorf("Failed to save: %v", err)
		return m, nil
	}

	projects, _ := models.LoadProjects(m.projectsFile)
	models.PopulateWorktrees(projects)
	m.projects = projects
	m.expandedWorktrees = make(map[int]bool)

	m.exitInputMode()
	m.setFeedback("Added "+name, "success")
	return m, nil
}
```

### renderInputBox changes

Add a name field row after the path field row:

```go
// After path field rows, before error row:
nameLabel := "  Name: "
nameView := m.nameInput.View()
nameContent := nameLabel + nameView
if !m.nameTouched {
    nameContent += hintStyle.Render(" (auto)")
}
namePadding := menuContentWidth - lipgloss.Width(nameContent)
if namePadding < 0 {
    namePadding = 0
}
lines = append(lines, leftBorder+nameContent+strings.Repeat(" ", namePadding)+rightBorder)

if m.nameErr != nil {
    errMsg := errorStyle.Render(m.nameErr.Error())
    // ... same pattern as inputErr rendering
}
```

Update help row to show Shift+Tab hint when on name field:
```go
if !m.inputFocusPath {
    helpContent = helpStyle.Render("⇧Tab back") + sep + helpStyle.Render("⏎ confirm") + sep + helpStyle.Render("Esc back")
}
```

---

Now the tests:

- [ ] **Step 1: Write failing tests**

Add to `test/internal/tui/mainmenu_test.go`:

```go
func TestAddProject_PathFieldFocusedFirst(t *testing.T) {
	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	model, _ := m.EnterInputModeForTest("add-project"), nil
	_ = model
	// inputFocusPath should be true (path is active)
	if !m.InputFocusPath() {
		t.Error("expected path field focused first")
	}
}

func TestAddProject_NameAutoFillsFromPath(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "my-project")
	os.MkdirAll(projDir, 0755)

	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.EnterInputModeForTest("add-project")
	// Simulate typing path and pressing Enter
	m.SetPathInputValue(projDir)
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm := result.(*tui.MainMenuModel)
	if mm.NameInputValue() != "my-project" {
		t.Errorf("expected name 'my-project', got %q", mm.NameInputValue())
	}
}

func TestAddProject_NameAutoDeriveLockOnEdit(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "my-project")
	os.MkdirAll(projDir, 0755)

	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.EnterInputModeForTest("add-project")
	m.SetPathInputValue(projDir)
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm := result.(*tui.MainMenuModel)
	// User edits name
	mm.SetNameInputValue("custom-name")
	mm.SetNameTouched(true)
	// Changing path should not overwrite custom name
	mm.SetPathInputValue(filepath.Join(dir, "other"))
	mm.TriggerAutoDeriveName()
	if mm.NameInputValue() != "custom-name" {
		t.Errorf("expected locked name 'custom-name', got %q", mm.NameInputValue())
	}
}

func TestAddProject_ShiftTabReturnsToCopyField(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "proj")
	os.MkdirAll(projDir, 0755)

	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.EnterInputModeForTest("add-project")
	m.SetPathInputValue(projDir)
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // advance to name
	mm := result.(*tui.MainMenuModel)
	result2, _ := mm.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	mm2 := result2.(*tui.MainMenuModel)
	if !mm2.InputFocusPath() {
		t.Error("expected Shift+Tab to return to path field")
	}
}

func TestAddProject_DuplicateNameSoftWarn(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "existing")
	os.MkdirAll(projDir, 0755)
	projDir2 := filepath.Join(dir, "new-path")
	os.MkdirAll(projDir2, 0755)

	existing := []models.Project{{Name: "existing", Path: projDir}}
	m := tui.NewMainMenu(existing, []string{"claude"}, "claude", "animated")
	m.SetProjectsFile(filepath.Join(dir, "projects"))
	m.EnterInputModeForTest("add-project")
	m.SetPathInputValue(projDir2)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // advance to name
	m.SetNameInputValue("existing")
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // first Enter: warn
	mm := result.(*tui.MainMenuModel)
	if mm.NameErr() == nil {
		t.Error("expected soft-warn error on first Enter with duplicate name")
	}
	// Second Enter: confirm
	result2, _ := mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm2 := result2.(*tui.MainMenuModel)
	if mm2.InInputMode() {
		t.Error("expected input mode exited after second Enter")
	}
}

func TestAddProject_EscFromNameClearsSoftWarn(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "proj")
	os.MkdirAll(projDir, 0755)

	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.EnterInputModeForTest("add-project")
	m.SetPathInputValue(projDir)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // advance to name
	m.SetNameWarnShown(true)
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm := result.(*tui.MainMenuModel)
	if mm.NameWarnShown() {
		t.Error("expected soft-warn cleared after Esc from name field")
	}
	if !mm.InputFocusPath() {
		t.Error("expected focus returned to path field after Esc from name")
	}
}
```

**Important:** The existing `EnterInputModeForTest` (line 828 in `mainmenu.go`) only sets `m.inputMode = mode` — it does not call `enterInputMode()` and therefore does not initialize `pathInput`, `nameInput`, `inputFocusPath`, etc. Update it to call the real `enterInputMode`:

```go
func (m *MainMenuModel) EnterInputModeForTest(mode string) {
	m.enterInputMode(mode)
}
```

Add these exported accessor/mutator methods to `MainMenuModel` for tests (following the existing pattern of `InputMode`, `InInputMode`, etc.):

```go
func (m *MainMenuModel) InputFocusPath() bool        { return m.inputFocusPath }
func (m *MainMenuModel) NameInputValue() string       { return m.nameInput.Value() }
func (m *MainMenuModel) SetPathInputValue(v string)   { m.pathInput.SetValue(v) }
func (m *MainMenuModel) SetNameInputValue(v string)   { m.nameInput.SetValue(v) }
func (m *MainMenuModel) SetNameTouched(v bool)        { m.nameTouched = v }
func (m *MainMenuModel) SetNameWarnShown(v bool)      { m.nameWarnShown = v }
func (m *MainMenuModel) NameErr() error               { return m.nameErr }
func (m *MainMenuModel) NameWarnShown() bool          { return m.nameWarnShown }
func (m *MainMenuModel) TriggerAutoDeriveName()       { m.maybeAutoDeriveName() }
```

- [ ] **Step 2: Run tests — expect FAIL**

```bash
go test ./test/internal/tui/... -run "TestAddProject_" -v
```

- [ ] **Step 3: Implement (struct fields + enterInputMode + exitInputMode + updateInputMode + submitInputMode + renderInputBox)**

Follow the code provided above in the design section of this task.

- [ ] **Step 4: Run tests — expect PASS**

```bash
go test ./test/internal/tui/... -run "TestAddProject_" -v
```

- [ ] **Step 5: Run full suite**

```bash
go test ./... -count=1
```

- [ ] **Step 6: Commit**

```bash
git add internal/tui/mainmenu.go test/internal/tui/mainmenu_test.go
git commit -m "feat: upgrade add-project to two-field form with reactive name derivation"
```

---

## Task 6: Pre-fill path with default projects root

**Files:**
- Modify: `internal/tui/mainmenu.go`
- Modify: `test/internal/tui/mainmenu_test.go`

When `enterInputMode("add-project")` is called, read `projectsRootFile` and pre-fill the path input.

### New field and setter

Add to `MainMenuModel` struct:
```go
projectsRootFile string
```

Add setter (following `SetProjectsFile` pattern):
```go
func (m *MainMenuModel) SetProjectsRootFile(path string) { m.projectsRootFile = path }
```

### Read helper

Add a package-level function (not exported — used only within mainmenu.go):
```go
func readProjectsRoot(file string) string {
	if file == "" {
		return ""
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
```

### Update enterInputMode

After `ti.Focus()`, add:
```go
if root := readProjectsRoot(m.projectsRootFile); root != "" {
    prefill := root
    if !strings.HasSuffix(prefill, "/") {
        prefill += "/"
    }
    ti.SetValue(prefill)
    ti.CursorEnd()
}
```

- [ ] **Step 1: Write failing tests**

```go
func TestAddProject_PreFillsPathWithProjectsRoot(t *testing.T) {
	dir := t.TempDir()
	rootFile := filepath.Join(dir, "projects-root")
	os.WriteFile(rootFile, []byte(dir+"\n"), 0644)

	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.SetProjectsRootFile(rootFile)
	m.EnterInputModeForTest("add-project")

	if !strings.HasPrefix(m.PathInputValue(), dir) {
		t.Errorf("expected path pre-filled with %q, got %q", dir, m.PathInputValue())
	}
}

func TestAddProject_NoPreFillWhenRootFileAbsent(t *testing.T) {
	dir := t.TempDir()
	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.SetProjectsRootFile(filepath.Join(dir, "missing-file"))
	m.EnterInputModeForTest("add-project")

	if m.PathInputValue() != "" {
		t.Errorf("expected empty path when root file absent, got %q", m.PathInputValue())
	}
}
```

Add `PathInputValue()` accessor to `MainMenuModel`:
```go
func (m *MainMenuModel) PathInputValue() string { return m.pathInput.Value() }
```

- [ ] **Step 2: Run tests — expect FAIL**

```bash
go test ./test/internal/tui/... -run "TestAddProject_PreFill|TestAddProject_NoPreFill" -v
```

- [ ] **Step 3: Implement**

Add `projectsRootFile` field, `SetProjectsRootFile`, `readProjectsRoot`, and update `enterInputMode`.

- [ ] **Step 4: Run tests — expect PASS**

```bash
go test ./test/internal/tui/... -run "TestAddProject_PreFill|TestAddProject_NoPreFill" -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/tui/mainmenu.go test/internal/tui/mainmenu_test.go
git commit -m "feat: pre-fill add-project path with default projects root"
```

---

## Task 7: Settings panel — default projects directory

**Files:**
- Modify: `internal/tui/mainmenu.go`
- Modify: `test/internal/tui/mainmenu_test.go`

Add a fourth settings item "Default projects dir" that opens a text input when Enter is pressed (not left/right like the other items). Uses a new `settingsInputMode` sub-state within the settings panel.

### New fields on MainMenuModel

```go
settingsInputMode bool
settingsInput     textinput.Model
settingsInputErr  error
projectsRoot      string // current value loaded from projectsRootFile; "" = not set
```

### Load projectsRoot on init

In the caller that sets up `MainMenuModel` (check `cmd/ghost-tab-tui/main_menu.go` or wherever `SetProjectsRootFile` is called), also read and set `projectsRoot`. OR load it lazily in `renderSettingsBox`.

Simpler approach: add a `LoadProjectsRoot()` method that reads the file and stores in `m.projectsRoot`:
```go
func (m *MainMenuModel) LoadProjectsRoot() {
	m.projectsRoot = readProjectsRoot(m.projectsRootFile)
}
```

Call this from wherever other file-based settings are loaded (e.g., after `SetProjectsRootFile` is called, or at startup).

### updateSettings changes

Update `n = 3` → `n = 4` at **all** occurrences. There are 5 sites:
- `case tea.KeyUp: const n = 3` (line 1356)
- `case tea.KeyDown: const n = 3` (line 1360) — **and** the literal `% 3` on line 1361 which does NOT use `n` and must also be changed to `% 4`
- `case 'j': ... % 3` (line 1388) — uses literal `3`, not `n`
- `case 'k': const n = 3` (line 1391)

Grep to find all: `grep -n "% 3\|const n = 3" internal/tui/mainmenu.go`

Add case 3 under `tea.KeyEnter`:
```go
case 3:
    // Open text input for projects root
    m.settingsInputMode = true
    si := textinput.New()
    si.Placeholder = "e.g., ~/Projects"
    si.Width = menuContentWidth - 11
    si.SetValue(m.projectsRoot)
    si.Focus()
    m.settingsInput = si
    m.settingsInputErr = nil
    return m, textinput.Blink
```

Add guard at top of `updateSettings` to route to settings input handler:
```go
if m.settingsInputMode {
    return m.updateSettingsInput(msg)
}
```

### WantsEsc update

The existing `WantsEsc()` method returns `m.settingsMode || m.inputMode != "" || m.deleteMode`. When `settingsInputMode` is true, Esc should stay inside the settings flow rather than triggering the app's double-Esc quit guard. Update:

```go
func (m *MainMenuModel) WantsEsc() bool {
    return m.settingsMode || m.inputMode != "" || m.deleteMode || m.settingsInputMode
}
```

### updateSettingsInput

```go
func (m *MainMenuModel) updateSettingsInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.settingsInputMode = false
		m.settingsInput.Blur()
		return m, nil
	case tea.KeyEnter:
		val := strings.TrimSpace(m.settingsInput.Value())
		if val != "" {
			expanded := util.ExpandPath(val)
			if _, err := os.Stat(expanded); err != nil {
				m.settingsInputErr = fmt.Errorf("Directory not found")
				return m, nil
			}
			if err := os.WriteFile(m.projectsRootFile, []byte(expanded+"\n"), 0644); err != nil {
				m.settingsInputErr = fmt.Errorf("Failed to save: %v", err)
				return m, nil
			}
			m.projectsRoot = expanded
		} else {
			os.Remove(m.projectsRootFile)
			m.projectsRoot = ""
		}
		m.settingsInputMode = false
		m.settingsInput.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.settingsInput, cmd = m.settingsInput.Update(msg)
	m.settingsInputErr = nil
	return m, cmd
}
```

### renderSettingsBox changes

1. Add item 3 ("Default projects dir") after the Sound item:
```go
// Default projects dir item
rootLabel := "Default projects dir"
var rootState string
if m.projectsRoot != "" {
    rootState = "[" + shortenHomePath(m.projectsRoot) + "]"
} else {
    rootState = "[(not set)]"
}
rootColor := lipgloss.Color("241")
if m.projectsRoot != "" {
    rootColor = lipgloss.Color("114")
}
rootStyle := lipgloss.NewStyle().Foreground(rootColor)
lines = append(lines, m.renderSettingsItem(3, rootLabel, rootState, rootStyle, primaryBoldStyle, leftBorder, rightBorder))
```

2. When `m.settingsInputMode`, show the text input in place of the items (or append it after item 3):
```go
if m.settingsInputMode {
    inputLabel := "  Path: "
    inputContent := inputLabel + m.settingsInput.View()
    // ... render same pattern as renderInputBox path row
    if m.settingsInputErr != nil {
        // render error row
    }
}
```

3. Update help row to show `⏎ edit` for item 3 instead of `← → cycle`.

- [ ] **Step 1: Add stripAnsi helper to test package**

`stripAnsi` is also used by Task 8. Add it once here in `test/internal/tui/mainmenu_test.go`:

```go
import "regexp"
var ansiEscRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)
func stripAnsi(s string) string { return ansiEscRe.ReplaceAllString(s, "") }
```

- [ ] **Step 2: Update the existing SettingsNavigationWraps test**

In `test/internal/tui/mainmenu_test.go` around line 1508, `TestMainMenu_SettingsNavigationWraps` uses `const numItems = 3`. Update to `numItems = 4` (the new fourth item "Default projects dir" shifts the wrap point).

- [ ] **Step 3: Write new failing tests**

```go
func TestSettings_ProjectsRootItem_AppearsInSettingsBox(t *testing.T) {
	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.EnterSettings()
	view := m.View()
	if !strings.Contains(stripAnsi(view), "Default projects dir") {
		t.Error("expected 'Default projects dir' in settings view")
	}
}

func TestSettings_ProjectsRootItem_ShowsNotSet(t *testing.T) {
	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.EnterSettings()
	view := stripAnsi(m.View())
	if !strings.Contains(view, "(not set)") {
		t.Errorf("expected '(not set)' when no root configured, view: %q", view)
	}
}

func TestSettings_ProjectsRootItem_ShowsCurrentValue(t *testing.T) {
	dir := t.TempDir()
	rootFile := filepath.Join(dir, "projects-root")
	os.WriteFile(rootFile, []byte(dir), 0644)

	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.SetProjectsRootFile(rootFile)
	m.LoadProjectsRoot()
	m.EnterSettings()
	view := stripAnsi(m.View())
	if !strings.Contains(view, filepath.Base(dir)) {
		t.Errorf("expected root path in view, got: %q", view)
	}
}

func TestSettings_NavWrapsWithFourItems(t *testing.T) {
	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.EnterSettings()
	// Navigate down 4 times — should wrap back to 0
	for i := 0; i < 4; i++ {
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.SettingsSelected() != 0 {
		t.Errorf("expected settingsSelected=0 after wrapping past 4 items, got %d", m.SettingsSelected())
	}
}
```

Add accessor (the `EnterSettings()` method already exists; use it):
```go
func (m *MainMenuModel) LoadProjectsRoot() { m.projectsRoot = readProjectsRoot(m.projectsRootFile) }
```

(`stripAnsi` was added in Task 7 Step 1; it is already available in this package.)

- [ ] **Step 2: Run tests — expect FAIL**

```bash
go test ./test/internal/tui/... -run "TestSettings_ProjectsRoot|TestSettings_NavWraps" -v
```

- [ ] **Step 3: Implement**

- Add struct fields, `settingsInputMode`, `settingsInput`, `settingsInputErr`, `projectsRoot`
- Update `updateSettings`: guard for `settingsInputMode`, update `n=3→4` in all 5 places, add case 3
- Add `updateSettingsInput`
- Update `renderSettingsBox`: add item 3, render input when `settingsInputMode`

- [ ] **Step 4: Run tests — expect PASS**

```bash
go test ./test/internal/tui/... -run "TestSettings_ProjectsRoot|TestSettings_NavWraps" -v
```

- [ ] **Step 5: Run full suite**

```bash
go test ./... -count=1
```

- [ ] **Step 6: Commit**

```bash
git add internal/tui/mainmenu.go test/internal/tui/mainmenu_test.go
git commit -m "feat: add default projects directory setting to settings panel"
```

---

## Task 8: Stale path UI

**Files:**
- Modify: `internal/tui/mainmenu.go`
- Modify: `test/internal/tui/mainmenu_test.go`

Show `⚠` marker for stale projects in the list. On selection of a stale project, show an inline confirmation instead of launching.

### New field on MainMenuModel

```go
staleConfirmIdx int // project index awaiting stale launch confirmation; -1 = inactive
```

Initialize to `-1` in `NewMainMenu`.

### renderMenuBox changes

When rendering a project row, check `m.projects[i].Stale` and prepend `⚠ ` (in yellow) to the row. Since `renderMenuBox` is large, find the project row rendering section and add:

```go
if project.Stale {
    staleMarker := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("⚠")
    // prepend marker to row prefix
}
```

### Selection handling for stale projects

In the project selection path (where `m.setActionResult("select")` is called), check staleness first:

```go
if m.projects[projectIdx].Stale {
    m.staleConfirmIdx = projectIdx
    return m, nil // show confirmation instead of launching
}
```

Add a stale confirmation rendering in `renderMenuBox` (or as a separate box that overrides the menu — simpler: show it as a feedback-style message at the bottom of the menu box):

```
⚠ Path not found: /deleted/path
  Launch anyway? [y/N]
```

Handle `y`/`Y`/`n`/`N`/Enter/Esc keys when `m.staleConfirmIdx >= 0`:

```go
if m.staleConfirmIdx >= 0 {
    switch msg.String() {
    case "y", "Y":
        // Proceed with launch even though stale
        m.staleConfirmIdx = -1
        return m.selectProject(savedIdx) // or re-trigger the selection
    default:
        m.staleConfirmIdx = -1
        return m, nil
    }
}
```

- [ ] **Step 1: Write failing tests**

```go
func TestStale_MarkerRenderedForStaleProject(t *testing.T) {
	projects := []models.Project{
		{Name: "good", Path: "/exists", Stale: false},
		{Name: "bad", Path: "/nonexistent", Stale: true},
	}
	m := tui.NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	view := stripAnsi(m.View())
	if !strings.Contains(view, "⚠") {
		t.Error("expected ⚠ marker for stale project in view")
	}
}

func TestStale_NoMarkerForHealthyProject(t *testing.T) {
	projects := []models.Project{
		{Name: "good", Path: "/exists", Stale: false},
	}
	m := tui.NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	view := stripAnsi(m.View())
	if strings.Contains(view, "⚠") {
		t.Error("expected no ⚠ marker when project is healthy")
	}
}

func TestStale_SelectionShowsConfirmation(t *testing.T) {
	projects := []models.Project{
		{Name: "bad", Path: "/nonexistent", Stale: true},
	}
	m := tui.NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	// Select the stale project
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm := result.(*tui.MainMenuModel)
	view := stripAnsi(mm.View())
	if !strings.Contains(view, "Launch anyway") {
		t.Errorf("expected stale confirmation prompt, got: %q", view)
	}
}

func TestStale_NKeyAtConfirmationCancels(t *testing.T) {
	projects := []models.Project{
		{Name: "bad", Path: "/nonexistent", Stale: true},
	}
	m := tui.NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // select → confirmation
	mm := result.(*tui.MainMenuModel)
	result2, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	mm2 := result2.(*tui.MainMenuModel)
	if mm2.StaleConfirmIdx() >= 0 {
		t.Error("expected stale confirmation dismissed after n")
	}
	if mm2.Result() != nil {
		t.Error("expected no result after cancelling stale confirmation")
	}
}

func TestStale_EnterKeyAtConfirmationCancels(t *testing.T) {
	// Enter = accept default N = cancel
	projects := []models.Project{{Name: "bad", Path: "/nonexistent", Stale: true}}
	m := tui.NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm := result.(*tui.MainMenuModel)
	result2, _ := mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm2 := result2.(*tui.MainMenuModel)
	if mm2.StaleConfirmIdx() >= 0 {
		t.Error("expected stale confirmation dismissed after Enter (default N)")
	}
}

func TestStale_YKeyAtConfirmationProceeds(t *testing.T) {
	projects := []models.Project{{Name: "bad", Path: "/nonexistent", Stale: true}}
	m := tui.NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // select → confirmation
	mm := result.(*tui.MainMenuModel)
	result2, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	mm2 := result2.(*tui.MainMenuModel)
	if mm2.StaleConfirmIdx() >= 0 {
		t.Error("expected stale confirmation dismissed after y")
	}
	if mm2.Result() == nil {
		t.Error("expected a result (launch) after y at stale confirmation")
	}
}
```

Add accessor:
```go
func (m *MainMenuModel) StaleConfirmIdx() int { return m.staleConfirmIdx }
```

- [ ] **Step 2: Run tests — expect FAIL**

```bash
go test ./test/internal/tui/... -run "TestStale_" -v
```

- [ ] **Step 3: Implement stale UI**

- Add `staleConfirmIdx int` field; initialize to `-1` in `NewMainMenu`
- Update `renderMenuBox` to show `⚠` for stale projects
- Update selection handling to set `staleConfirmIdx` instead of launching
- Add stale confirmation rendering (a few rows at bottom of menu box)
- Add stale confirmation key handling early in `Update`

- [ ] **Step 4: Run tests — expect PASS**

```bash
go test ./test/internal/tui/... -run "TestStale_" -v
```

- [ ] **Step 5: Run full suite**

```bash
go test ./... -count=1
```

- [ ] **Step 6: Commit**

```bash
git add internal/tui/mainmenu.go test/internal/tui/mainmenu_test.go
git commit -m "feat: show stale path warning and launch confirmation in project list"
```

---

## Task 9: Wire up projectsRootFile in the binary entry point

**Files:**
- Modify: `cmd/ghost-tab-tui/main_menu.go` (or wherever `NewMainMenu` is set up and file paths are configured)

Find the place where `SetProjectsFile`, `SetAIToolFile`, `SetSettingsFile` are called and add:

```go
configDir := filepath.Join(os.Getenv("HOME"), ".config", "ghost-tab")
// or use the same XDG_CONFIG_HOME logic already present
m.SetProjectsRootFile(filepath.Join(configDir, "projects-root"))
m.LoadProjectsRoot()
```

- [ ] **Step 1: Find the call site**

```bash
grep -n "SetProjectsFile\|SetAIToolFile\|SetSettingsFile" cmd/ghost-tab-tui/*.go
```

- [ ] **Step 2: Add SetProjectsRootFile and LoadProjectsRoot calls**

Following the exact same pattern as the existing calls.

- [ ] **Step 3: Build the binary**

```bash
go build ./...
```

- [ ] **Step 4: Run full test suite**

```bash
go test ./... -count=1
```

- [ ] **Step 5: Run shellcheck on all modified shell scripts**

```bash
shellcheck lib/projects.sh wrapper.sh
```

- [ ] **Step 6: Commit**

```bash
git add cmd/ghost-tab-tui/
git commit -m "feat: wire projects root file into main menu binary"
```

---

## Final Verification

- [ ] **Run complete test suite**

```bash
./run-tests.sh
```

- [ ] **Run shellcheck on all lib scripts**

```bash
find lib bin -name '*.sh' -exec shellcheck {} + && shellcheck wrapper.sh
```

- [ ] **Verify git status is clean**

```bash
git status
git log --oneline -8
```

- [ ] **Push**

```bash
git pull --rebase && git push
```
