package tui

import (
	"testing"
)

func TestMainMenu_TKeySwitchesToStatsTab(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "") // *MainMenuModel
	_, cmd := m.handleRune('t')
	if cmd != nil {
		t.Errorf("'t' should not emit a navigation cmd (switches tab, no PushScreenMsg), got %v", cmd)
	}
	if m.ActiveTab() != TabStats {
		t.Errorf("after 't' tab = %v, want TabStats", m.ActiveTab())
	}
}

func TestMainMenu_hasStatsActionLabel(t *testing.T) {
	found := false
	for _, a := range actionLabels {
		if a.shortcut == "T" && a.label == "Stats" {
			found = true
		}
	}
	if !found {
		t.Errorf("actionLabels missing {T, Stats}: %+v", actionLabels)
	}
}
