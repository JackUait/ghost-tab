package tui_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/tui"
)

func TestConfirmModel(t *testing.T) {
	model := tui.NewConfirmDialog("Delete project?")

	if model.Message != "Delete project?" {
		t.Errorf("Expected message 'Delete project?', got %q", model.Message)
	}

	if model.Confirmed {
		t.Error("Expected confirmed to be false initially")
	}
}

func TestPathSuggestions_DirectoryCompletion(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "project-a"), 0755)
	os.MkdirAll(filepath.Join(dir, "project-b"), 0755)
	os.MkdirAll(filepath.Join(dir, "other"), 0755)

	suggestions := tui.GetPathSuggestions(dir + "/proj")
	if len(suggestions) != 2 {
		t.Errorf("expected 2 suggestions for 'proj', got %d: %v", len(suggestions), suggestions)
	}
}

func TestPathSuggestions_MaxEight(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 12; i++ {
		os.MkdirAll(filepath.Join(dir, fmt.Sprintf("dir%02d", i)), 0755)
	}

	suggestions := tui.GetPathSuggestions(dir + "/")
	if len(suggestions) > 8 {
		t.Errorf("expected max 8 suggestions, got %d", len(suggestions))
	}
}

func TestPathSuggestions_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "MyProject"), 0755)

	suggestions := tui.GetPathSuggestions(dir + "/myp")
	if len(suggestions) != 1 {
		t.Errorf("expected 1 case-insensitive match, got %d: %v", len(suggestions), suggestions)
	}
}

func TestPathSuggestions_EmptyInput(t *testing.T) {
	suggestions := tui.GetPathSuggestions("")
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for empty input, got %d", len(suggestions))
	}
}

func TestPathSuggestions_NonexistentDir(t *testing.T) {
	suggestions := tui.GetPathSuggestions("/nonexistent/path/foo")
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for nonexistent dir, got %d", len(suggestions))
	}
}

func TestPathSuggestions_TrailingSlash(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	suggestions := tui.GetPathSuggestions(dir + "/")
	if len(suggestions) < 1 {
		t.Errorf("expected at least 1 suggestion for trailing slash, got %d", len(suggestions))
	}
	// All suggestions should end with /
	for _, s := range suggestions {
		if !strings.HasSuffix(s, "/") {
			t.Errorf("suggestion %q should end with /", s)
		}
	}
}

func TestConfirmDialog_YConfirms(t *testing.T) {
	m := tui.NewConfirmDialog("Delete?")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	result := updated.(tui.ConfirmDialogModel)
	if !result.Confirmed {
		t.Error("'y' should confirm")
	}
}

func TestConfirmDialog_UpperYConfirms(t *testing.T) {
	m := tui.NewConfirmDialog("Delete?")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})
	result := updated.(tui.ConfirmDialogModel)
	if !result.Confirmed {
		t.Error("'Y' should confirm")
	}
}

func TestConfirmDialog_NDenies(t *testing.T) {
	m := tui.NewConfirmDialog("Delete?")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	result := updated.(tui.ConfirmDialogModel)
	if result.Confirmed {
		t.Error("'n' should deny")
	}
}

func TestConfirmDialog_EscDenies(t *testing.T) {
	m := tui.NewConfirmDialog("Delete?")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := updated.(tui.ConfirmDialogModel)
	if result.Confirmed {
		t.Error("Esc should deny")
	}
}

func TestConfirmDialog_CtrlCDenies(t *testing.T) {
	m := tui.NewConfirmDialog("Delete?")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := updated.(tui.ConfirmDialogModel)
	if result.Confirmed {
		t.Error("Ctrl+C should deny")
	}
}

func TestConfirmDialog_ViewShowsMessage(t *testing.T) {
	m := tui.NewConfirmDialog("Are you sure?")
	view := m.View()
	if !strings.Contains(view, "Are you sure?") {
		t.Error("view should contain the message")
	}
	if !strings.Contains(view, "y/n") {
		t.Error("view should contain y/n hint")
	}
}

func TestProjectInput_New(t *testing.T) {
	m := tui.NewProjectInput()
	if m.Confirmed() {
		t.Error("Should not be confirmed initially")
	}
	if m.Name() != "" {
		t.Errorf("Name should be empty, got %q", m.Name())
	}
}

func TestProjectInput_InitReturnsBlink(t *testing.T) {
	m := tui.NewProjectInput()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return blink command")
	}
}

func TestProjectInput_CtrlCCancels(t *testing.T) {
	m := tui.NewProjectInput()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("Ctrl+C should return quit command")
	}
	result := updated.(tui.ProjectInputModel)
	if result.Confirmed() {
		t.Error("Ctrl+C should not confirm")
	}
}

func TestProjectInput_EscCancels(t *testing.T) {
	m := tui.NewProjectInput()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("Esc should return quit command")
	}
	result := updated.(tui.ProjectInputModel)
	if result.Confirmed() {
		t.Error("Esc should not confirm")
	}
}

func TestProjectInput_ViewContainsLabels(t *testing.T) {
	m := tui.NewProjectInput()
	view := m.View()
	if !strings.Contains(view, "Project Name") {
		t.Error("View should contain 'Project Name'")
	}
	if !strings.Contains(view, "Project Path") {
		t.Error("View should contain 'Project Path'")
	}
}

func TestProjectInput_ViewEmptyAfterQuit(t *testing.T) {
	m := tui.NewProjectInput()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := updated.(tui.ProjectInputModel)
	if result.View() != "" {
		t.Error("View should be empty after quitting")
	}
}

func TestProjectInput_EnterWithEmptyNameShowsError(t *testing.T) {
	m := tui.NewProjectInput()
	// Press enter without typing anything (name is empty)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("Should not quit with empty name")
	}
	result := updated.(tui.ProjectInputModel)
	view := result.View()
	if !strings.Contains(view, "Error") {
		t.Error("Should show error for empty name")
	}
}

func TestProjectInput_EnterWithNameAdvancesToPath(t *testing.T) {
	m := tui.NewProjectInput()
	// Type a project name
	var model tea.Model = m
	for _, r := range "myproject" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	// Press enter to advance to path field
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("Enter with valid name should return blink command for path field")
	}
	result := updated.(tui.ProjectInputModel)
	if result.Confirmed() {
		t.Error("Should not be confirmed yet (still on path step)")
	}
	if result.Name() != "myproject" {
		t.Errorf("Expected name 'myproject', got %q", result.Name())
	}
}

func TestProjectInput_EnterWithEmptyPathShowsError(t *testing.T) {
	m := tui.NewProjectInput()
	// Type a project name and advance
	var model tea.Model = m
	for _, r := range "test" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Now on path field, press enter with empty path
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("Should not quit with empty path")
	}
	result := updated.(tui.ProjectInputModel)
	view := result.View()
	if !strings.Contains(view, "Error") {
		t.Error("Should show error for empty path")
	}
}

func TestProjectInput_FullFlowConfirms(t *testing.T) {
	dir := t.TempDir()
	m := tui.NewProjectInput()
	// Type a project name
	var model tea.Model = m
	for _, r := range "testproj" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	// Advance to path
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Type a valid path
	for _, r := range dir {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	// Dismiss autocomplete before confirming
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	// Confirm
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("Should return quit command on confirmation")
	}
	result := updated.(tui.ProjectInputModel)
	if !result.Confirmed() {
		t.Error("Should be confirmed after entering valid name and path")
	}
	if result.Name() != "testproj" {
		t.Errorf("Expected name 'testproj', got %q", result.Name())
	}
	if result.Path() != dir {
		t.Errorf("Expected path %q, got %q", dir, result.Path())
	}
}

func TestProjectInput_InvalidPathShowsError(t *testing.T) {
	m := tui.NewProjectInput()
	var model tea.Model = m
	for _, r := range "proj" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Type a nonexistent path
	for _, r := range "/nonexistent/path/xyz" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("Should not quit with invalid path")
	}
	result := updated.(tui.ProjectInputModel)
	view := result.View()
	if !strings.Contains(view, "Error") {
		t.Error("Should show error for invalid path")
	}
}

func TestProjectInput_WindowSizeMsg(t *testing.T) {
	m := tui.NewProjectInput()
	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd != nil {
		t.Error("WindowSizeMsg should return nil cmd")
	}
	_ = updated
}
