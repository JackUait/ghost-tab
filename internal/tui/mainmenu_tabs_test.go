package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestActiveTab_defaultsToProjects(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	if m.ActiveTab() != TabProjects {
		t.Errorf("default tab = %v, want TabProjects", m.ActiveTab())
	}
}

func TestSetActiveTab(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetActiveTab(TabStats)
	if m.ActiveTab() != TabStats {
		t.Errorf("after SetActiveTab(TabStats) = %v, want TabStats", m.ActiveTab())
	}
}

func TestCycleTab_wraps(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.CycleTab("next") // Projects -> Settings
	if m.ActiveTab() != TabSettings {
		t.Fatalf("next from Projects = %v, want Settings", m.ActiveTab())
	}
	m.CycleTab("next") // Settings -> Stats
	m.CycleTab("next") // Stats -> Projects (wrap)
	if m.ActiveTab() != TabProjects {
		t.Fatalf("3x next = %v, want Projects (wrap)", m.ActiveTab())
	}
	m.CycleTab("prev") // Projects -> Stats (wrap back)
	if m.ActiveTab() != TabStats {
		t.Fatalf("prev from Projects = %v, want Stats (wrap)", m.ActiveTab())
	}
}

func TestHandleRune_sSwitchesToSettingsTab(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.handleRune('s')
	if m.ActiveTab() != TabSettings {
		t.Errorf("after 's' tab = %v, want Settings", m.ActiveTab())
	}
}

func TestHandleRune_tSwitchesToStatsTab(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	_, cmd := m.handleRune('t')
	if m.ActiveTab() != TabStats {
		t.Errorf("after 't' tab = %v, want Stats", m.ActiveTab())
	}
	// 't' may return a load cmd but must not push a new screen.
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(PushScreenMsg); ok {
			t.Errorf("'t' should not emit a PushScreenMsg, got %T", msg)
		}
	}
}

func TestUpdate_tabKeyCycles(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.ActiveTab() != TabSettings {
		t.Errorf("Tab from Projects = %v, want Settings", m.ActiveTab())
	}
	m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.ActiveTab() != TabProjects {
		t.Errorf("Shift+Tab back = %v, want Projects", m.ActiveTab())
	}
}
