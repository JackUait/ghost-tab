package tui

import (
	"testing"
)

func TestGetConfigMenuItems_returns_expected_items(t *testing.T) {
	items := GetConfigMenuItems()

	expectedActions := []string{
		"manage-terminals",
		"manage-projects",
		"select-ai-tools",
		"display-settings",
		"reinstall",
		"quit",
	}

	if len(items) != len(expectedActions) {
		t.Fatalf("got %d items, want %d", len(items), len(expectedActions))
	}

	for i, item := range items {
		if item.Action != expectedActions[i] {
			t.Errorf("item %d: got action %q, want %q", i, item.Action, expectedActions[i])
		}
	}
}

func TestNewConfigMenu_initializes_list(t *testing.T) {
	model := NewConfigMenu()
	if model.Selected() != nil {
		t.Error("expected nil selected before interaction")
	}
}
