package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/tui"
)

func TestConfigMenuItems(t *testing.T) {
	items := tui.GetConfigMenuItems()

	expectedCount := 6

	if len(items) != expectedCount {
		t.Errorf("Expected %d menu items, got %d", expectedCount, len(items))
	}

	if items[0].Action != "manage-terminals" {
		t.Errorf("Expected first action 'manage-terminals', got %q", items[0].Action)
	}
}

func TestConfigMenu_New(t *testing.T) {
	m := tui.NewConfigMenu()
	if m.Selected() != nil {
		t.Error("Selected should be nil initially")
	}
}

func TestConfigMenu_InitReturnsNil(t *testing.T) {
	m := tui.NewConfigMenu()
	if m.Init() != nil {
		t.Error("Init should return nil")
	}
}

func TestConfigMenu_EnterSelectsItem(t *testing.T) {
	m := tui.NewConfigMenu()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("Enter should return quit command")
	}
	result := updated.(tui.ConfigMenuModel)
	if result.Selected() == nil {
		t.Fatal("Enter should select current item")
	}
	if result.Selected().Action != "manage-terminals" {
		t.Errorf("Expected first item 'manage-terminals', got %q", result.Selected().Action)
	}
}

func TestConfigMenu_EscSelectsQuit(t *testing.T) {
	m := tui.NewConfigMenu()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("Esc should return quit command")
	}
	result := updated.(tui.ConfigMenuModel)
	if result.Selected() == nil {
		t.Fatal("Esc should set selected to quit action")
	}
	if result.Selected().Action != "quit" {
		t.Errorf("Expected quit action, got %q", result.Selected().Action)
	}
}

func TestConfigMenu_CtrlCSelectsQuit(t *testing.T) {
	m := tui.NewConfigMenu()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("Ctrl+C should return quit command")
	}
	result := updated.(tui.ConfigMenuModel)
	if result.Selected() == nil {
		t.Fatal("Ctrl+C should set selected to quit action")
	}
	if result.Selected().Action != "quit" {
		t.Errorf("Expected quit action, got %q", result.Selected().Action)
	}
}

func TestConfigMenu_WindowSizeMsg(t *testing.T) {
	m := tui.NewConfigMenu()
	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd != nil {
		t.Error("WindowSizeMsg should return nil cmd")
	}
	_ = updated
}

func TestConfigMenu_ViewNonEmpty(t *testing.T) {
	m := tui.NewConfigMenu()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := updated.(tui.ConfigMenuModel).View()
	if view == "" {
		t.Error("View should not be empty before quitting")
	}
}

func TestConfigMenu_ViewEmptyAfterQuit(t *testing.T) {
	m := tui.NewConfigMenu()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := updated.(tui.ConfigMenuModel)
	if result.View() != "" {
		t.Error("View should be empty after quitting")
	}
}
