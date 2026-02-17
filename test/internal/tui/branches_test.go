package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/tui"
)

func TestBranchPicker_SelectBranch(t *testing.T) {
	branches := []string{"feature/auth", "fix/cleanup", "develop"}
	m := tui.NewBranchPicker(branches)

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
	m := tui.NewBranchPicker(branches)

	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = sized.(tui.BranchPickerModel)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(tui.BranchPickerModel)

	if m.Selected() != nil {
		t.Error("expected nil selection on cancel")
	}
}

func TestBranchPicker_EmptyList(t *testing.T) {
	m := tui.NewBranchPicker(nil)

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
	m := tui.NewBranchPicker(branches)

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
