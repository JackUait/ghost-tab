package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/models"
	"github.com/jackuait/ghost-tab/internal/tui"
)

func testTerminals() []models.Terminal {
	return []models.Terminal{
		{Name: "ghostty", DisplayName: "Ghostty", CaskName: "ghostty", Installed: true},
		{Name: "iterm2", DisplayName: "iTerm2", CaskName: "iterm2", Installed: true},
		{Name: "wezterm", DisplayName: "WezTerm", CaskName: "wezterm", Installed: false},
		{Name: "kitty", DisplayName: "kitty", CaskName: "kitty", Installed: false},
	}
}

func TestTerminalSelector_initial_state_has_no_selection(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	if m.Selected() != nil {
		t.Error("expected no selection initially")
	}
	if m.InstallRequest() != "" {
		t.Error("expected no install request initially")
	}
}

func TestTerminalSelector_cursor_starts_at_zero(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	if m.Cursor() != 0 {
		t.Errorf("expected cursor at 0, got %d", m.Cursor())
	}
}

func TestTerminalSelector_enter_selects_installed_terminal(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
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
	var model tea.Model = tui.NewTerminalSelector(testTerminals(), "ghostty")
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := model.(tui.TerminalSelectorModel)
	if result.Selected() != nil {
		t.Error("expected no selection for uninstalled terminal")
	}
}

func TestTerminalSelector_enter_on_uninstalled_triggers_install(t *testing.T) {
	var model tea.Model = tui.NewTerminalSelector(testTerminals(), "ghostty")
	// Move to WezTerm (index 2, uninstalled)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := model.(tui.TerminalSelectorModel)
	if result.InstallRequest() != "wezterm" {
		t.Errorf("expected install request for wezterm, got %q", result.InstallRequest())
	}
	if result.InstallRequestCask() != "wezterm" {
		t.Errorf("expected cask name 'wezterm', got %q", result.InstallRequestCask())
	}
	if cmd == nil {
		t.Error("expected quit command after install request")
	}
}

func TestTerminalSelector_down_wraps_around(t *testing.T) {
	var model tea.Model = tui.NewTerminalSelector(testTerminals(), "")
	for i := 0; i < 4; i++ {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	result := model.(tui.TerminalSelectorModel)
	if result.Cursor() != 0 {
		t.Errorf("expected cursor to wrap to 0, got %d", result.Cursor())
	}
}

func TestTerminalSelector_up_wraps_around(t *testing.T) {
	var model tea.Model = tui.NewTerminalSelector(testTerminals(), "")
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	result := model.(tui.TerminalSelectorModel)
	if result.Cursor() != 3 {
		t.Errorf("expected cursor to wrap to 3, got %d", result.Cursor())
	}
}

func TestTerminalSelector_j_moves_down(t *testing.T) {
	var model tea.Model = tui.NewTerminalSelector(testTerminals(), "")
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	result := model.(tui.TerminalSelectorModel)
	if result.Cursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", result.Cursor())
	}
}

func TestTerminalSelector_k_moves_up(t *testing.T) {
	var model tea.Model = tui.NewTerminalSelector(testTerminals(), "")
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	result := model.(tui.TerminalSelectorModel)
	if result.Cursor() != 0 {
		t.Errorf("expected cursor at 0, got %d", result.Cursor())
	}
}

func TestTerminalSelector_escape_cancels(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
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
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := updated.(tui.TerminalSelectorModel)
	if result.Selected() != nil {
		t.Error("expected no selection after ctrl+c")
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestTerminalSelector_i_on_uninstalled_triggers_install(t *testing.T) {
	var model tea.Model = tui.NewTerminalSelector(testTerminals(), "ghostty")
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	result := model.(tui.TerminalSelectorModel)
	if result.InstallRequest() != "wezterm" {
		t.Errorf("expected install request for wezterm, got %q", result.InstallRequest())
	}
	if cmd == nil {
		t.Error("expected quit command after install request")
	}
}

func TestTerminalSelector_i_on_installed_does_nothing(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	result := updated.(tui.TerminalSelectorModel)
	if result.InstallRequest() != "" {
		t.Error("expected no install request for installed terminal")
	}
	if cmd != nil {
		t.Error("expected no quit command")
	}
}

func TestTerminalSelector_view_shows_title(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	view := m.View()
	if !strings.Contains(view, "Select Terminal") {
		t.Error("view should contain title 'Select Terminal'")
	}
}

func TestTerminalSelector_view_shows_installed_marker(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	view := m.View()
	if !strings.Contains(view, "installed") {
		t.Error("view should show 'installed' for installed terminals")
	}
}

func TestTerminalSelector_view_shows_not_installed_marker(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	view := m.View()
	if !strings.Contains(view, "not installed") {
		t.Error("view should show 'not installed' for uninstalled terminals")
	}
}

func TestTerminalSelector_view_shows_current_marker(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	view := m.View()
	if !strings.Contains(view, "current") {
		t.Error("view should show 'current' marker for current terminal")
	}
}

func TestTerminalSelector_view_no_current_when_empty(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "")
	view := m.View()
	if strings.Contains(view, "current") {
		t.Error("view should not show 'current' when no current terminal set")
	}
}

func TestTerminalSelector_view_shows_cursor(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "")
	view := m.View()
	if !strings.Contains(view, "\u25b8") {
		t.Error("view should show cursor indicator ▸")
	}
}

func TestTerminalSelector_view_has_bordered_box(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	view := m.View()
	// Should have rounded border corners like the config menu
	if !strings.Contains(view, "╭") {
		t.Error("view should have rounded top-left border corner")
	}
	if !strings.Contains(view, "╰") {
		t.Error("view should have rounded bottom-left border corner")
	}
}

func TestTerminalSelector_view_has_title_in_border(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) == 0 {
		t.Fatal("view should not be empty")
	}
	// Title should be overlaid on the top border line
	if !strings.Contains(lines[0], "Select Terminal") {
		t.Error("title 'Select Terminal' should be overlaid on the top border")
	}
}

func TestTerminalSelector_view_has_bullet_separators_in_hint(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	view := m.View()
	if !strings.Contains(view, "\u2022") {
		t.Error("hint bar should use bullet separators like config menu")
	}
}

func TestTerminalSelector_handles_window_size_msg(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "")
	_, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd != nil {
		t.Error("WindowSizeMsg should return nil cmd")
	}
}

func TestTerminalSelector_view_shows_hint_bar(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "")
	view := m.View()
	if !strings.Contains(view, "navigate") {
		t.Error("view should show hint bar with navigation help")
	}
}

func TestTerminalSelector_view_empty_after_quit(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := updated.(tui.TerminalSelectorModel)
	if result.View() != "" {
		t.Error("View should be empty after quitting")
	}
}

func TestTerminalSelector_init_returns_nil(t *testing.T) {
	m := tui.NewTerminalSelector(testTerminals(), "")
	if m.Init() != nil {
		t.Error("Init should return nil")
	}
}

func TestTerminalSelector_enter_on_only_uninstalled(t *testing.T) {
	terminals := []models.Terminal{
		{Name: "wezterm", DisplayName: "WezTerm", CaskName: "wezterm", Installed: false},
	}
	m := tui.NewTerminalSelector(terminals, "")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(tui.TerminalSelectorModel)
	if result.Selected() != nil {
		t.Error("should not select uninstalled terminal")
	}
	if result.InstallRequest() != "wezterm" {
		t.Errorf("expected install request for wezterm, got %q", result.InstallRequest())
	}
	if cmd == nil {
		t.Error("expected quit command for install request")
	}
}

func TestTerminalSelector_hint_shows_install_only_on_uninstalled(t *testing.T) {
	// Cursor on installed terminal (index 0: Ghostty) → should NOT show "Enter install"
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	view := m.View()
	if strings.Contains(view, "Enter install") {
		t.Error("hint bar should not show 'Enter install' when cursor is on installed terminal")
	}
}

func TestTerminalSelector_hint_shows_install_on_uninstalled(t *testing.T) {
	// Cursor on uninstalled terminal (index 2: WezTerm) → should show "Enter install"
	var model tea.Model = tui.NewTerminalSelector(testTerminals(), "ghostty")
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	result := model.(tui.TerminalSelectorModel)
	view := result.View()
	if !strings.Contains(view, "Enter install") {
		t.Error("hint bar should show 'Enter install' when cursor is on uninstalled terminal")
	}
}

func TestTerminalSelector_hint_hides_select_on_uninstalled(t *testing.T) {
	// Cursor on uninstalled terminal → should show "Enter install" not "Enter select"
	var model tea.Model = tui.NewTerminalSelector(testTerminals(), "ghostty")
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	result := model.(tui.TerminalSelectorModel)
	view := result.View()
	if strings.Contains(view, "Enter select") {
		t.Error("hint bar should not show 'Enter select' when cursor is on uninstalled terminal")
	}
}

func TestTerminalSelector_hint_shows_enter_on_installed(t *testing.T) {
	// Cursor on installed terminal → should show "Enter select"
	m := tui.NewTerminalSelector(testTerminals(), "ghostty")
	view := m.View()
	if !strings.Contains(view, "Enter select") {
		t.Error("hint bar should show 'Enter select' when cursor is on installed terminal")
	}
}

func TestTerminalSelector_install_request_includes_cask_name(t *testing.T) {
	terminals := []models.Terminal{
		{Name: "wezterm", DisplayName: "WezTerm", CaskName: "wezterm", Installed: false},
	}
	var model tea.Model = tui.NewTerminalSelector(terminals, "")
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	result := model.(tui.TerminalSelectorModel)
	if result.InstallRequestCask() != "wezterm" {
		t.Errorf("expected cask name 'wezterm', got %q", result.InstallRequestCask())
	}
}

func TestTerminalSelector_select_second_installed(t *testing.T) {
	var model tea.Model = tui.NewTerminalSelector(testTerminals(), "ghostty")
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := model.(tui.TerminalSelectorModel)
	selected := result.Selected()
	if selected == nil {
		t.Fatal("expected a selection")
	}
	if selected.Name != "iterm2" {
		t.Errorf("expected iterm2, got %q", selected.Name)
	}
}
