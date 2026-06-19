package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// boxBorderColor is the color of the outer box border. Idle, it is a neutral
// gray so the chrome recedes and the accent is free to mark only the selected
// row. When a focus region outside the body is active the border brightens to
// Primary so the user can see which box "owns" the keyboard at a glance.
func (m *MainMenuModel) boxBorderColor() lipgloss.Color {
	if m.focus != FocusBody {
		return m.theme.Primary
	}
	return lipgloss.Color("240") // neutral gray (matches the rest of the grays)
}

// boxBorders returns the rounded-box border strings shared by every tab body.
func (m *MainMenuModel) boxBorders() (top, separator, bottom, leftBorder, rightBorder string) {
	borderStyle := lipgloss.NewStyle().Foreground(m.boxBorderColor())
	// Inner separators stay a touch dimmer than the outer border so the box
	// reads as one frame rather than a stack of equally-weighted rules.
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	hLine := strings.Repeat("─", menuInnerWidth)
	top = borderStyle.Render("╭" + hLine + "╮")
	separator = sepStyle.Render("├" + hLine + "┤")
	bottom = borderStyle.Render("╰" + hLine + "╯")
	leftBorder = borderStyle.Render("│")
	rightBorder = strings.Repeat(" ", menuPadding) + borderStyle.Render("│")
	return
}

// menuTabLabels is the ordered list of top-level tab labels.
var menuTabLabels = []string{"Projects", "Settings", "Stats"}

// renderTabBar renders the Projects · Settings · Stats row. The active tab is
// wrapped in block accents and styled bold; inactive tabs are dimmed.
func (m *MainMenuModel) renderTabBar(leftBorder, rightBorder string) string {
	// When the tab bar holds focus, the active tab brightens to signal that ←/→
	// will switch sections; otherwise it stays the dimmer Primary.
	activeColor := m.theme.Primary
	if m.focus == FocusTabs {
		activeColor = m.theme.Bright
	}
	// Active tab: bold + underlined, so it reads as a tab rather than the old
	// ▌label▐ block glyphs that looked like a render artifact. Inactive tabs are
	// neutral gray and recede. Both keep the same " label " width so the row math
	// is unchanged.
	activeStyle := lipgloss.NewStyle().Foreground(activeColor).Bold(true).Underline(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	var parts []string
	for i, label := range menuTabLabels {
		if MenuTab(i) == m.activeTab {
			parts = append(parts, activeStyle.Render(" "+label+" "))
		} else {
			parts = append(parts, inactiveStyle.Render(" "+label+" "))
		}
	}
	content := strings.Join(parts, "  ")
	gap := menuContentWidth - lipgloss.Width(content) - 1
	if gap < 0 {
		gap = 0
	}
	return leftBorder + " " + content + strings.Repeat(" ", gap) + rightBorder
}
