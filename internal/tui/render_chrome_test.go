package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderTabBar_showsAllTabs(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	_, _, _, lb, rb := m.boxBorders()
	bar := m.renderTabBar(lb, rb)
	for _, label := range []string{"Projects", "Settings", "Stats"} {
		if !strings.Contains(bar, label) {
			t.Errorf("tab bar missing %q: %q", label, bar)
		}
	}
}

func TestRenderTabBar_activeTabAccented(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetActiveTab(TabSettings)
	_, _, _, lb, rb := m.boxBorders()
	bar := m.renderTabBar(lb, rb)
	// The active tab is bold + underlined (no block-glyph artifacts).
	want := lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true).Underline(true).Render(" Settings ")
	if !strings.Contains(bar, want) {
		t.Errorf("active tab bar missing bold+underline accent: %q", bar)
	}
}
