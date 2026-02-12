package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/tui"
)

func TestConfirmDialog_InitialState(t *testing.T) {
	m := tui.NewConfirmDialog("Delete project?")

	if m.Message != "Delete project?" {
		t.Errorf("Expected message %q, got %q", "Delete project?", m.Message)
	}
	if m.Confirmed {
		t.Error("Expected Confirmed to be false initially")
	}
}

func TestConfirmDialog_Init(t *testing.T) {
	m := tui.NewConfirmDialog("Delete?")
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestConfirmDialog_ConfirmKeys(t *testing.T) {
	keys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'y'}},
		{Type: tea.KeyRunes, Runes: []rune{'Y'}},
	}

	for _, key := range keys {
		t.Run(key.String(), func(t *testing.T) {
			m := tui.NewConfirmDialog("Confirm?")
			updated, cmd := m.Update(key)

			result := updated.(tui.ConfirmDialogModel)
			if !result.Confirmed {
				t.Error("Expected Confirmed=true after pressing " + key.String())
			}
			if cmd == nil {
				t.Error("Expected tea.Quit command after pressing " + key.String())
			}
		})
	}
}

func TestConfirmDialog_RejectKeys(t *testing.T) {
	keys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'n'}},
		{Type: tea.KeyRunes, Runes: []rune{'N'}},
		{Type: tea.KeyCtrlC},
		{Type: tea.KeyEscape},
	}

	for _, key := range keys {
		t.Run(key.String(), func(t *testing.T) {
			m := tui.NewConfirmDialog("Cancel?")
			updated, cmd := m.Update(key)

			result := updated.(tui.ConfirmDialogModel)
			if result.Confirmed {
				t.Error("Expected Confirmed=false after pressing " + key.String())
			}
			if cmd == nil {
				t.Error("Expected tea.Quit command after pressing " + key.String())
			}
		})
	}
}

func TestConfirmDialog_IgnoredKeys(t *testing.T) {
	keys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'a'}},
		{Type: tea.KeySpace},
		{Type: tea.KeyEnter},
	}

	for _, key := range keys {
		t.Run(key.String(), func(t *testing.T) {
			m := tui.NewConfirmDialog("Ignore?")
			_, cmd := m.Update(key)

			if cmd != nil {
				t.Error("Expected no command for ignored key " + key.String())
			}
		})
	}
}

func TestConfirmDialog_View(t *testing.T) {
	m := tui.NewConfirmDialog("Are you sure?")
	view := m.View()

	if !strings.Contains(view, "Are you sure?") {
		t.Errorf("View should contain message, got %q", view)
	}
	if !strings.Contains(view, "y/n") {
		t.Errorf("View should contain y/n hint, got %q", view)
	}
}

func TestConfirmDialog_ViewAfterQuit(t *testing.T) {
	m := tui.NewConfirmDialog("Quit?")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	result := updated.(tui.ConfirmDialogModel)
	view := result.View()

	if view != "" {
		t.Errorf("View after quitting should be empty, got %q", view)
	}
}
