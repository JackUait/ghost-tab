package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// The PLAN switcher row sits under the AGENT picker on the Projects tab. The
// header chrome is shared across tabs, so the PLAN row must render on Settings
// and Stats too — otherwise the header jumps between tabs and the focus ring
// lands on an invisible stop.
func TestPlanRow_rendersOnSettingsTab(t *testing.T) {
	m := subTestMenu("claude")
	m.SetActiveTab(TabSettings)
	out := stripAnsi(m.renderSettingsBox())
	if !strings.Contains(out, "PLAN") {
		t.Errorf("settings box should carry the PLAN switcher row:\n%s", out)
	}
}

func TestPlanRow_rendersOnStatsTab(t *testing.T) {
	m := subTestMenu("claude")
	m.SetActiveTab(TabStats)
	out := stripAnsi(m.renderStatsBox())
	if !strings.Contains(out, "PLAN") {
		t.Errorf("stats box should carry the PLAN switcher row:\n%s", out)
	}
}

// The PLAN row sits between the AGENT title row and the tab bar, matching the
// Projects layout, so the chrome lines up identically across every tab.
func TestPlanRow_sitsAboveTabBarOnSettings(t *testing.T) {
	m := subTestMenu("claude")
	m.SetActiveTab(TabSettings)
	lines := strings.Split(stripAnsi(m.renderSettingsBox()), "\n")
	planIdx, agentIdx, tabIdx := -1, -1, -1
	for i, l := range lines {
		if strings.Contains(l, "AGENT") {
			agentIdx = i
		}
		if strings.Contains(l, "PLAN") {
			planIdx = i
		}
		// The tab bar is the row that lists the section names together.
		if tabIdx == -1 && strings.Contains(l, "Projects") && strings.Contains(l, "Stats") {
			tabIdx = i
		}
	}
	if agentIdx < 0 || planIdx < 0 || tabIdx < 0 {
		t.Fatalf("missing rows: agent=%d plan=%d tab=%d\n%s", agentIdx, planIdx, tabIdx, strings.Join(lines, "\n"))
	}
	if !(agentIdx < planIdx && planIdx < tabIdx) {
		t.Errorf("expected AGENT < PLAN < tab bar, got agent=%d plan=%d tab=%d", agentIdx, planIdx, tabIdx)
	}
}

// Now that the PLAN row renders on every tab, its focus stop must be reachable
// on Settings and Stats too (when a keyed config exists).
func TestPlanRow_focusReachableOnAllTabs(t *testing.T) {
	for _, tab := range []MenuTab{TabSettings, TabStats} {
		m := subFocusMenu(t, "claude", true)
		m.SetActiveTab(tab)
		if !m.subscriptionFocusable() {
			t.Errorf("tab %v: subscription should be focusable now its row renders on every tab", tab)
		}

		// Down from AI must stop on the subscription row.
		m.SetFocus(FocusAI)
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if m.Focus() != FocusSubscription {
			t.Errorf("tab %v: Down from AI = %v, want FocusSubscription", tab, m.Focus())
		}

		// Up from the tab bar must stop on the subscription row.
		m.SetFocus(FocusTabs)
		m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if m.Focus() != FocusSubscription {
			t.Errorf("tab %v: Up from tabs = %v, want FocusSubscription", tab, m.Focus())
		}
	}
}
