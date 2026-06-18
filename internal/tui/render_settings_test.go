package tui

import (
	"strings"
	"testing"
)

func TestRenderSettingsBox_hasTabBar(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetActiveTab(TabSettings)
	out := m.renderSettingsBox()
	if !strings.Contains(out, "Settings") || !strings.Contains(out, "Projects") {
		t.Errorf("settings box missing tab bar: %q", out)
	}
	if !strings.Contains(out, "▌") {
		t.Errorf("settings tab should be accented as active")
	}
}

func TestRenderSettingsBox_settingsTabAccented(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetActiveTab(TabSettings)
	out := m.renderSettingsBox()
	// Active tab renders as ▌Settings▐
	if !strings.Contains(out, "▌Settings▐") {
		t.Errorf("active Settings tab should render as ▌Settings▐, got:\n%s", out)
	}
}

func TestRenderSettingsBox_hasChromeStructure(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetActiveTab(TabSettings)
	out := m.renderSettingsBox()
	// Must have top border, title row, tab bar, separator
	for _, glyph := range []string{"╭", "╮", "╰", "╯", "│"} {
		if !strings.Contains(out, glyph) {
			t.Errorf("settings box missing border glyph %q:\n%s", glyph, out)
		}
	}
}

func TestRenderSettingsBox_preservesSettingsRows(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetActiveTab(TabSettings)
	out := m.renderSettingsBox()
	// Core settings items must still appear
	for _, label := range []string{"Ghost Display", "Tab Title", "Sound"} {
		if !strings.Contains(out, label) {
			t.Errorf("settings box missing row %q:\n%s", label, out)
		}
	}
}
