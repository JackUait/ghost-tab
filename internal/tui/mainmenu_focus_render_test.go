package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHelpRow_focusAI(t *testing.T) {
	m := focusTestMenu()
	m.SetFocus(FocusAI)
	help := stripAnsi(m.renderHelpRow())
	if !strings.Contains(help, "switch agent") {
		t.Errorf("agent-focus help should say 'switch agent', got %q", help)
	}
	if strings.Contains(help, "switch AI") {
		t.Errorf("agent-focus help should no longer say 'switch AI', got %q", help)
	}
	if !strings.Contains(help, "sections") {
		t.Errorf("agent-focus help should point down to sections, got %q", help)
	}
}

func TestHelpRow_focusTabs(t *testing.T) {
	m := focusTestMenu()
	m.SetFocus(FocusTabs)
	help := stripAnsi(m.renderHelpRow())
	if !strings.Contains(help, "section") {
		t.Errorf("tabs-focus help should mention switching section, got %q", help)
	}
}

func TestHelpRow_focusBodyProjects(t *testing.T) {
	m := focusTestMenu()
	// default focus body, projects tab
	help := stripAnsi(m.renderHelpRow())
	if !strings.Contains(help, "open") {
		t.Errorf("projects-body help should mention open, got %q", help)
	}
	if !strings.Contains(help, "sections") {
		t.Errorf("projects-body help should advertise ↑ to sections, got %q", help)
	}
}

func TestHelpRow_focusBodySettings(t *testing.T) {
	m := focusTestMenu()
	m.EnterSettings()
	help := stripAnsi(m.renderHelpRow())
	if !strings.Contains(help, "change") {
		t.Errorf("settings-body help should mention change, got %q", help)
	}
}

func TestHelpRow_focusBodyStats(t *testing.T) {
	m := focusTestMenu()
	m.SetActiveTab(TabStats)
	m.SetFocus(FocusBody)
	help := stripAnsi(m.renderHelpRow())
	if !strings.Contains(help, "scroll") {
		t.Errorf("stats-body help should mention scroll, got %q", help)
	}
}

// TestHelpRow_spansFullBoxWidth guards centering when no ghost is shown. The
// help row sits below the box and is centered by lipgloss.Place against the
// widest content line (the box border, menuBoxWidth wide). If the help line is
// narrower than the box, Place gives it extra left margin (short/2), shifting
// the hints right of box-center. Padding the line to the full box width keeps
// short==0 so the hints stay centered relative to the box.
func TestHelpRow_spansFullBoxWidth(t *testing.T) {
	m := focusTestMenu()
	m.SetFocus(FocusTabs)
	help := stripAnsi(m.renderHelpRow())
	if got := lipgloss.Width(help); got != menuBoxWidth {
		t.Errorf("help row width = %d, want %d (full box width so it centers with the box)", got, menuBoxWidth)
	}
	// The hints must be visually centered: leading and trailing padding equal
	// (within one column for odd remainders).
	trimmed := strings.TrimRight(help, " ")
	leftPad := lipgloss.Width(help) - lipgloss.Width(strings.TrimLeft(help, " "))
	rightPad := lipgloss.Width(help) - lipgloss.Width(trimmed)
	if diff := leftPad - rightPad; diff < -1 || diff > 1 {
		t.Errorf("help row not centered: leftPad=%d rightPad=%d", leftPad, rightPad)
	}
}

// TestView_helpCentersOnBox_noGhost renders the whole menu with the ghost
// disabled and verifies the footer hints sit centered over the box, not shifted
// to one side. This is the exact scenario the user reported.
func TestView_helpCentersOnBox_noGhost(t *testing.T) {
	m := focusTestMenu()
	m.ghostDisplay = "none"
	m.SetFocus(FocusTabs)

	lines := strings.Split(stripAnsi(m.View()), "\n")

	// The box border lines are the widest; find one to locate the box's columns.
	var boxLeft, boxRight int
	for _, l := range lines {
		trimmed := strings.TrimRight(l, " ")
		if strings.Contains(trimmed, "─────") {
			boxLeft = lipgloss.Width(l) - lipgloss.Width(strings.TrimLeft(l, " "))
			boxRight = lipgloss.Width(trimmed)
			break
		}
	}
	if boxRight == 0 {
		t.Fatal("could not locate box border in rendered view")
	}

	// Find the footer hint line (the one mentioning the tab-focus hint).
	var hint string
	for _, l := range lines {
		if strings.Contains(l, "switch section") {
			hint = l
			break
		}
	}
	if hint == "" {
		t.Fatal("could not locate footer hint line in rendered view")
	}

	hintLeft := lipgloss.Width(hint) - lipgloss.Width(strings.TrimLeft(hint, " "))
	hintRight := lipgloss.Width(strings.TrimRight(hint, " "))
	boxCenter := float64(boxLeft+boxRight) / 2
	hintCenter := float64(hintLeft+hintRight) / 2
	if diff := hintCenter - boxCenter; diff < -1 || diff > 1 {
		t.Errorf("hint not centered on box: boxCenter=%.1f hintCenter=%.1f (diff %.1f)", boxCenter, hintCenter, diff)
	}
}
