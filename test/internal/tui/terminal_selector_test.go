package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/models"
	"github.com/jackuait/ghost-tab/internal/tui"
)

func testTerminals() []models.Terminal {
	return []models.Terminal{
		{Name: "ghostty", DisplayName: "Ghostty", Installed: true},
		{Name: "iterm2", DisplayName: "iTerm2", Installed: true},
		{Name: "wezterm", DisplayName: "WezTerm", Installed: false},
		{Name: "kitty", DisplayName: "kitty", Installed: true},
	}
}

func TestTerminalSelector_initial_state_has_no_selection(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals())
	if m.Selected() != nil {
		t.Error("expected no selection initially")
	}
}

func TestTerminalSelector_enter_selects_installed_terminal(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(tui.TerminalSelectorModel)
	selected := result.Selected()
	if selected == nil {
		t.Fatal("expected a selection")
	}
	if selected.Name != "ghostty" {
		t.Errorf("expected ghostty, got %q", selected.Name)
	}
}

func TestTerminalSelector_enter_does_not_select_uninstalled(t *testing.T) {
	var model tea.Model = tui.NewTerminalSelector(testTerminals())
	// Move down to wezterm (index 2, not installed)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := model.(tui.TerminalSelectorModel)
	if result.Selected() != nil {
		t.Error("expected no selection for uninstalled terminal")
	}
}

func TestTerminalSelector_escape_cancels(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals())
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := updated.(tui.TerminalSelectorModel)
	if result.Selected() != nil {
		t.Error("expected no selection after escape")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestTerminalSelector_ctrl_c_cancels(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals())
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := updated.(tui.TerminalSelectorModel)
	if result.Selected() != nil {
		t.Error("expected no selection after ctrl+c")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestTerminalSelector_init_returns_nil(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals())
	if m.Init() != nil {
		t.Error("Init should return nil")
	}
}

func TestTerminalSelector_window_size_msg(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals())
	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd != nil {
		t.Error("WindowSizeMsg should return nil cmd")
	}
	_ = updated // should not panic
}

func TestTerminalSelector_view_non_empty(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals())
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := updated.(tui.TerminalSelectorModel).View()
	if view == "" {
		t.Error("View should not be empty before quitting")
	}
}

func TestTerminalSelector_view_empty_after_quit(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := updated.(tui.TerminalSelectorModel)
	if result.View() != "" {
		t.Error("View should be empty after quitting")
	}
}

func TestTerminalSelector_enter_on_only_uninstalled(t *testing.T) {
	terminals := []models.Terminal{
		{Name: "wezterm", DisplayName: "WezTerm", Installed: false},
	}
	m := tui.NewTerminalSelector(terminals)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(tui.TerminalSelectorModel)
	if result.Selected() != nil {
		t.Error("should not select uninstalled terminal")
	}
}
