package tui

import (
	"strings"
	"testing"
)

func TestRenderStatsBox_hasTabBar(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetActiveTab(TabStats)
	out := m.renderStatsBox()
	if !strings.Contains(out, "Stats") {
		t.Errorf("stats box missing tab bar: %q", out)
	}
}

func TestRenderStatsBox_statsTabAccented(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetActiveTab(TabStats)
	out := m.renderStatsBox()
	if !strings.Contains(out, "▌Stats▐") {
		t.Errorf("active Stats tab should render as ▌Stats▐, got:\n%s", out)
	}
}

func TestRenderStatsBox_hasChromeStructure(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetActiveTab(TabStats)
	out := m.renderStatsBox()
	for _, glyph := range []string{"╭", "╮", "╰", "╯", "│"} {
		if !strings.Contains(out, glyph) {
			t.Errorf("stats box missing border glyph %q:\n%s", glyph, out)
		}
	}
}

// TestRenderStatsBox_SetSizeHonored is the carried-forward size test from Task 2.
// It verifies that SetSize is honored — the output is non-empty, contains "Stats"
// (from the tab bar), and renders within menuContentWidth.
func TestRenderStatsBox_SetSizeHonored(t *testing.T) {
	const w, h = 120, 40
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetSize(w, h)
	m.SetActiveTab(TabStats)
	out := m.renderStatsBox()

	if out == "" {
		t.Fatal("renderStatsBox returned empty string after SetSize")
	}
	if !strings.Contains(out, "Stats") {
		t.Errorf("output missing 'Stats' (tab bar not rendered): %q", out)
	}

	// Each box line must not exceed the box width (menuInnerWidth + 2 borders).
	maxAllowed := menuInnerWidth + 2
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		// Only check lines that are box rows (contain border glyphs).
		if !strings.Contains(line, "│") {
			continue
		}
		if w := visibleWidth(line); w > maxAllowed {
			t.Errorf("box line exceeds menuInnerWidth+2 (%d): width=%d %q", maxAllowed, w, line)
		}
	}
}

func TestRenderStatsBox_containsStatsRows(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.SetActiveTab(TabStats)
	out := m.renderStatsBox()
	// The stats rows section must be present (loading or data row).
	// At minimum "Token" or "usage" or loading text appears.
	if !strings.Contains(out, "Token") && !strings.Contains(out, "usage") && !strings.Contains(out, "Loading") && !strings.Contains(out, "No usage") {
		t.Errorf("stats box missing stats content:\n%s", out)
	}
}

// visibleWidth returns the printable (non-ANSI) width of s using lipgloss.
func visibleWidth(s string) int {
	// Import lipgloss at the top of the file — reuse existing import in package.
	// We just call strings.Count as a rough proxy; lipgloss.Width handles ANSI.
	// Since this is an internal test in package tui, we can call lipgloss directly.
	return len([]rune(stripANSI(s)))
}

// stripANSI removes ANSI escape sequences for width measurement.
func stripANSI(s string) string {
	var out []rune
	inEscape := false
	for _, r := range s {
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		if r == '\x1b' {
			inEscape = true
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
