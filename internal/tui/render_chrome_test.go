package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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
	// Force a real color profile so lipgloss emits styling escapes; otherwise the
	// label/padding distinction is invisible to string matching.
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetActiveTab(TabSettings)
	_, _, _, lb, rb := m.boxBorders()
	bar := m.renderTabBar(lb, rb)
	// The active tab is bold + underlined on the label only (the surrounding
	// padding spaces are not underlined, so the rule sits tight under the word).
	want := lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true).Underline(true).Render("Settings")
	if !strings.Contains(bar, want) {
		t.Errorf("active tab bar missing bold+underline accent on label: %q", bar)
	}
	padded := lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true).Underline(true).Render(" Settings ")
	if strings.Contains(bar, padded) {
		t.Errorf("active tab underline should not span the padding spaces: %q", bar)
	}
}
