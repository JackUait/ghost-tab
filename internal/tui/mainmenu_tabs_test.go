package tui

import "testing"

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
