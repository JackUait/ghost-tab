package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/tui"
)

func testTheme() tui.AIToolTheme {
	return tui.ThemeForTool("claude")
}

func TestBranchPicker_SelectBranch(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(tui.BranchPickerModel)

	selected := m.Selected()
	if selected == nil {
		t.Fatal("expected a selected branch")
	}
	if *selected != "feature/auth" {
		t.Errorf("got %q, want %q", *selected, "feature/auth")
	}
}

func TestBranchPicker_Cancel(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(tui.BranchPickerModel)

	if m.Selected() != nil {
		t.Error("expected nil selection on cancel")
	}
}

func TestBranchPicker_EmptyList(t *testing.T) {
	m := tui.NewBranchPicker(nil, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(tui.BranchPickerModel)

	if m.Selected() != nil {
		t.Error("expected nil selection on empty list")
	}
}

func TestBranchPicker_NavigateAndSelect(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(tui.BranchPickerModel)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(tui.BranchPickerModel)

	selected := m.Selected()
	if selected == nil {
		t.Fatal("expected a selected branch")
	}
	if *selected != "fix/cleanup" {
		t.Errorf("got %q, want %q", *selected, "fix/cleanup")
	}
}

func TestBranchPicker_FilterAndSelect(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	// Press '/' to enter filter mode, then type "dev"
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(tui.BranchPickerModel)
	for _, r := range "dev" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(tui.BranchPickerModel)
	}

	// Enter selects the only matching item
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(tui.BranchPickerModel)

	selected := m.Selected()
	if selected == nil {
		t.Fatal("expected a selected branch")
	}
	if *selected != "develop" {
		t.Errorf("got %q, want %q", *selected, "develop")
	}
}

func TestBranchPicker_EscClearsFilter(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	// Press '/' then type "dev" to filter
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(tui.BranchPickerModel)
	for _, r := range "dev" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(tui.BranchPickerModel)
	}

	// First Esc clears filter and exits filter mode (not quit)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(tui.BranchPickerModel)

	// Should not have quit — view should still render
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after clearing filter")
	}

	// All branches should be visible again
	if !strings.Contains(view, "feature/auth") {
		t.Error("expected all branches visible after clearing filter")
	}
}

func TestBranchPicker_ViewHasBoxBorders(t *testing.T) {
	branches := []string{"feature/auth", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	view := m.View()
	if !strings.Contains(view, "\u250c") { // ┌
		t.Error("expected top-left border character")
	}
	if !strings.Contains(view, "\u2518") { // ┘
		t.Error("expected bottom-right border character")
	}
	if !strings.Contains(view, "Select Branch") {
		t.Error("expected title in view")
	}
	if !strings.Contains(view, "feature/auth") {
		t.Error("expected branch name in view")
	}
}

func TestBranchPicker_DeleteMode(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	// Press 'd' to enter delete mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(tui.BranchPickerModel)

	view := m.View()
	// Title should show "· Delete" suffix like project delete mode
	if !strings.Contains(view, "Delete") {
		t.Error("expected title to contain 'Delete'")
	}
	// Help bar should show delete-mode controls
	if !strings.Contains(view, "delete") {
		t.Error("expected help bar with delete action")
	}
	if !strings.Contains(view, "cancel") {
		t.Error("expected help bar with cancel option")
	}
	// Branches should still be listed
	if !strings.Contains(view, "feature/auth") {
		t.Error("expected branches visible in delete mode")
	}
}

func TestBranchPicker_DeleteCancel(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	// Press 'd' then Esc to cancel
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(tui.BranchPickerModel)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(tui.BranchPickerModel)

	// Should be back to normal view — title should not show "Delete"
	view := m.View()
	if strings.Contains(view, "\u00b7 Delete") {
		t.Error("expected delete mode to be dismissed after Esc")
	}
	// All branches should still be present
	if !strings.Contains(view, "feature/auth") {
		t.Error("expected all branches still present after cancel")
	}
}

func TestBranchPicker_DeleteCancelQ(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	// Press 'd' then 'q' to cancel
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(tui.BranchPickerModel)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(tui.BranchPickerModel)

	// Should be back to normal view
	view := m.View()
	if strings.Contains(view, "\u00b7 Delete") {
		t.Error("expected delete mode to be dismissed after 'q'")
	}
	if !strings.Contains(view, "feature/auth") {
		t.Error("expected all branches still present after cancel")
	}
}

func TestBranchPicker_DeleteRemovesBranch(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	// Simulate receiving a successful branchDeletedMsg
	updated, _ := m.Update(tui.BranchDeletedMsg{Branch: "feature/auth", Err: nil})
	m = updated.(tui.BranchPickerModel)

	// Feedback should show "Deleted feature/auth"
	view := m.View()
	if !strings.Contains(view, "Deleted") {
		t.Error("expected success feedback message")
	}
	if !strings.Contains(view, "fix/cleanup") {
		t.Error("expected remaining branches to still be visible")
	}

	// Press any key to clear feedback, then verify branch is gone
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = updated.(tui.BranchPickerModel)

	view = m.View()
	if strings.Contains(view, "feature/auth") {
		t.Error("expected deleted branch to be removed from view")
	}
	if !strings.Contains(view, "fix/cleanup") {
		t.Error("expected remaining branches to still be visible after feedback cleared")
	}
}

func TestBranchPicker_DeleteModeScrolls(t *testing.T) {
	// Many branches in a small terminal — delete mode must scroll
	branches := make([]string, 30)
	for i := range branches {
		branches[i] = "branch-" + string(rune('a'+i%26)) + "-" + strings.Repeat("x", i)
	}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	// Height 12 → visibleItemCount = 12-8 = 4 (delete box has 8 fixed rows)
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 12})
	m = sized.(tui.BranchPickerModel)

	// Enter delete mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(tui.BranchPickerModel)

	view := m.View()
	lineCount := strings.Count(view, "\n") + 1
	// View must not exceed terminal height
	if lineCount > 12 {
		t.Errorf("delete mode view has %d lines, exceeds terminal height 12", lineCount)
	}
	// First branch should be visible
	if !strings.Contains(view, branches[0]) {
		t.Error("expected first branch visible")
	}
	// A branch far down should NOT be visible
	if strings.Contains(view, branches[29]) {
		t.Error("expected branch 29 to be scrolled out of view")
	}
}

func TestBranchPicker_DeleteKeyIgnoredWhileFiltering(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	// Press '/' to enter filter mode, then type "d"
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(tui.BranchPickerModel)

	// Now type 'd' — should add to filter, not trigger delete mode
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(tui.BranchPickerModel)

	view := m.View()
	// Should NOT show delete mode title
	if strings.Contains(view, "\u00b7 Delete") {
		t.Error("'d' should add to filter text, not trigger delete mode")
	}
}

func TestBranchPicker_SlashActivatesFilterMode(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	// Without pressing '/', typing should NOT filter
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = updated.(tui.BranchPickerModel)

	view := m.View()
	// All branches should still be visible (no filtering happened)
	if !strings.Contains(view, "feature/auth") {
		t.Error("expected all branches visible when not in filter mode")
	}
	if !strings.Contains(view, "develop") {
		t.Error("expected all branches visible when not in filter mode")
	}
}

func TestBranchPicker_HelpBarShowsSlash(t *testing.T) {
	branches := []string{"feature/auth", "develop"}
	m := tui.NewBranchPicker(branches, testTheme(), "/tmp/project")

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	view := m.View()
	if !strings.Contains(view, "/ filter") {
		t.Error("expected help bar to show '/ filter'")
	}
}
