package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/models"
)

// --- config menu (top-level "Ghost Tab Configuration") ---

func TestConfigMenu_geometryMatchesRender(t *testing.T) {
	m := NewConfigMenu(ConfigMenuOptions{})
	m.width = 80
	m.cursor = 1
	lines := strings.Split(m.View(), "\n")
	markerRow := -1
	for i, l := range lines {
		if strings.Contains(l, "▸") { // ▸ cursor marker
			markerRow = i
			break
		}
	}
	if markerRow < 0 {
		t.Fatal("could not find cursor marker in rendered config menu")
	}
	if got := m.configMenuItemAt(markerRow); got != 1 {
		t.Errorf("configMenuItemAt(%d) = %d, want 1 (cursor row mismatch)", markerRow, got)
	}
}

func TestConfigMenu_clickSelectsItem(t *testing.T) {
	m := NewConfigMenu(ConfigMenuOptions{})
	m.width = 80
	msg := tea.MouseMsg{X: 5, Y: 2, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}
	upd, cmd := m.Update(msg)
	got := upd.(ConfigMenuModel)
	if got.Selected() == nil || got.Selected().Action != "manage-claude-configs" {
		t.Fatalf("click on first item should select it, got %v", got.Selected())
	}
	if cmd == nil {
		t.Error("selecting an item should emit a quit command")
	}
}

func TestConfigMenu_hoverMovesCursor(t *testing.T) {
	m := NewConfigMenu(ConfigMenuOptions{})
	m.width = 80
	upd, _ := m.Update(tea.MouseMsg{X: 5, Y: 5, Action: tea.MouseActionMotion})
	got := upd.(ConfigMenuModel)
	if got.Cursor() != 1 {
		t.Errorf("hover moved cursor to %d, want 1", got.Cursor())
	}
}

// --- multi-select AI tools ---

func TestMultiSelect_clickTogglesAndConfirm(t *testing.T) {
	tools := []models.AITool{{Name: "claude", Installed: true}, {Name: "opencode"}}
	m := NewMultiSelect(tools)
	if m.Checked()[1] {
		t.Fatal("precondition: opencode should start unchecked")
	}
	// Click opencode (tool index 1, screen row 3) to toggle it on.
	upd, _ := m.Update(tea.MouseMsg{X: 6, Y: 3, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	m = upd.(MultiSelectModel)
	if !m.Checked()[1] {
		t.Fatal("clicking the opencode row should check it")
	}
	// Click the Confirm button.
	upd, cmd := m.Update(tea.MouseMsg{X: 3, Y: m.multiSelectConfirmRow(), Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	got := upd.(MultiSelectModel)
	if got.Result() == nil || !got.Result().Confirmed {
		t.Fatalf("clicking Confirm should produce a confirmed result, got %v", got.Result())
	}
	if cmd == nil {
		t.Error("Confirm should emit a quit command")
	}
}

func TestMultiSelect_hoverMovesCursor(t *testing.T) {
	tools := []models.AITool{{Name: "claude", Installed: true}, {Name: "opencode"}}
	m := NewMultiSelect(tools)
	upd, _ := m.Update(tea.MouseMsg{X: 6, Y: 3, Action: tea.MouseActionMotion})
	got := upd.(MultiSelectModel)
	if got.Cursor() != 1 {
		t.Errorf("hover moved cursor to %d, want 1", got.Cursor())
	}
}

// --- claude config management menu ---

func TestClaudeConfigMenu_clickAddRowStartsAdd(t *testing.T) {
	configs := []ClaudeConfig{{Name: "Work", File: "work.json"}, {Name: "Personal", File: "personal.json"}}
	m := NewClaudeConfigMenu(configs)
	// Add row is the last list index (= len(configs)), at screen row 2+len.
	addRow := 2 + m.addRowIndex()
	upd, _ := m.Update(tea.MouseMsg{X: 5, Y: addRow, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	got := upd.(ClaudeConfigMenuModel)
	if got.mode != ccmAddInput {
		t.Errorf("clicking the Add row should start add-input mode, got mode %v", got.mode)
	}
}

func TestClaudeConfigMenu_hoverMovesCursor(t *testing.T) {
	configs := []ClaudeConfig{{Name: "Work", File: "work.json"}, {Name: "Personal", File: "personal.json"}}
	m := NewClaudeConfigMenu(configs)
	upd, _ := m.Update(tea.MouseMsg{X: 5, Y: 3, Action: tea.MouseActionMotion}) // second config (row 3)
	got := upd.(ClaudeConfigMenuModel)
	if got.cursor != 1 {
		t.Errorf("hover moved cursor to %d, want 1", got.cursor)
	}
}

// --- confirm dialog ---

func TestConfirmDialog_clickYesAndNo(t *testing.T) {
	row := NewConfirmDialog("Delete this?").confirmButtonRow()

	yes := NewConfirmDialog("Delete this?")
	upd, cmd := yes.Update(tea.MouseMsg{X: 2, Y: row, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	got := upd.(ConfirmDialogModel)
	if !got.Confirmed {
		t.Error("clicking Yes should set Confirmed = true")
	}
	if cmd == nil {
		t.Error("clicking Yes should emit a quit command")
	}

	no := NewConfirmDialog("Delete this?")
	// No button starts after "[ Yes ]" + the gap.
	noX := len(confirmYesLabel) + len(confirmGap) + 1
	upd2, cmd2 := no.Update(tea.MouseMsg{X: noX, Y: row, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	got2 := upd2.(ConfirmDialogModel)
	if got2.Confirmed {
		t.Error("clicking No should leave Confirmed = false")
	}
	if cmd2 == nil {
		t.Error("clicking No should still emit a quit command")
	}
}

func TestConfirmDialog_hoverHighlightsButton(t *testing.T) {
	m := NewConfirmDialog("Delete this?")
	upd, _ := m.Update(tea.MouseMsg{X: 2, Y: m.confirmButtonRow(), Action: tea.MouseActionMotion})
	got := upd.(ConfirmDialogModel)
	if got.btnHover != 1 {
		t.Errorf("hovering Yes should set btnHover = 1, got %d", got.btnHover)
	}
}
