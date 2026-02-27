package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/models"
)

func TestTerminalSelector_EnterOnUninstalled_SetsInstallRequest(t *testing.T) {
	terminals := []models.Terminal{
		{Name: "ghostty", DisplayName: "Ghostty", CaskName: "ghostty", Installed: true},
		{Name: "kitty", DisplayName: "kitty", CaskName: "kitty", Installed: false},
	}
	m := NewTerminalSelector(terminals, "")

	// Move cursor to kitty (index 1)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(TerminalSelectorModel)

	if m.Cursor() != 1 {
		t.Fatalf("expected cursor at 1, got %d", m.Cursor())
	}

	// Press Enter on uninstalled terminal
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(TerminalSelectorModel)

	if m.InstallRequest() != "kitty" {
		t.Errorf("expected InstallRequest %q, got %q", "kitty", m.InstallRequest())
	}
	if m.InstallRequestCask() != "kitty" {
		t.Errorf("expected InstallRequestCask %q, got %q", "kitty", m.InstallRequestCask())
	}
	if m.Selected() != nil {
		t.Error("expected Selected() to be nil for install request")
	}
	if cmd == nil {
		t.Error("expected tea.Quit command, got nil")
	}
}

func TestTerminalSelector_EnterOnInstalled_SetsSelected(t *testing.T) {
	terminals := []models.Terminal{
		{Name: "ghostty", DisplayName: "Ghostty", CaskName: "ghostty", Installed: true},
		{Name: "kitty", DisplayName: "kitty", CaskName: "kitty", Installed: false},
	}
	m := NewTerminalSelector(terminals, "")

	// Press Enter on installed terminal (cursor at 0)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(TerminalSelectorModel)

	if m.Selected() == nil {
		t.Fatal("expected Selected() to be non-nil")
	}
	if m.Selected().Name != "ghostty" {
		t.Errorf("expected selected terminal %q, got %q", "ghostty", m.Selected().Name)
	}
	if m.InstallRequest() != "" {
		t.Errorf("expected empty InstallRequest, got %q", m.InstallRequest())
	}
	if cmd == nil {
		t.Error("expected tea.Quit command, got nil")
	}
}

func TestTerminalSelector_Escape_Cancels(t *testing.T) {
	terminals := []models.Terminal{
		{Name: "ghostty", DisplayName: "Ghostty", CaskName: "ghostty", Installed: true},
	}
	m := NewTerminalSelector(terminals, "")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(TerminalSelectorModel)

	if m.Selected() != nil {
		t.Error("expected Selected() to be nil on escape")
	}
	if m.InstallRequest() != "" {
		t.Errorf("expected empty InstallRequest on escape, got %q", m.InstallRequest())
	}
	if cmd == nil {
		t.Error("expected tea.Quit command, got nil")
	}
}

func TestTerminalSelector_IKey_TriggersInstall(t *testing.T) {
	terminals := []models.Terminal{
		{Name: "ghostty", DisplayName: "Ghostty", CaskName: "ghostty", Installed: true},
		{Name: "wezterm", DisplayName: "WezTerm", CaskName: "wezterm", Installed: false},
	}
	m := NewTerminalSelector(terminals, "")

	// Move to uninstalled terminal
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(TerminalSelectorModel)

	// Press 'i' to install
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updated.(TerminalSelectorModel)

	if m.InstallRequest() != "wezterm" {
		t.Errorf("expected InstallRequest %q, got %q", "wezterm", m.InstallRequest())
	}
	if m.InstallRequestCask() != "wezterm" {
		t.Errorf("expected InstallRequestCask %q, got %q", "wezterm", m.InstallRequestCask())
	}
	if cmd == nil {
		t.Error("expected tea.Quit command, got nil")
	}
}
