package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderStatsRows renders the stats content as box rows using leftBorder/rightBorder,
// mirroring renderSettingsItem's label-left/value-right layout.
func (m *MainMenuModel) renderStatsRows(leftBorder, rightBorder string) []string {
	primaryBoldStyle := lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(m.theme.Dim)
	numStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	faint := lipgloss.NewStyle().Faint(true)

	// Fetch data from the stats model (loading state — we render what we have).
	s := NewStatsModel()

	emptyRow := leftBorder + strings.Repeat(" ", menuContentWidth) + rightBorder

	// Helper: render a label-left / value-right row inside the box.
	itemRow := func(label, value string, labelStyle, valStyle lipgloss.Style) string {
		labelRendered := labelStyle.Render(label)
		valRendered := valStyle.Render(value)
		prefix := "    " + labelRendered
		gap := menuContentWidth - lipgloss.Width(prefix) - lipgloss.Width(valRendered) - 1
		if gap < 1 {
			gap = 1
		}
		return leftBorder + prefix + strings.Repeat(" ", gap) + valRendered + " " + rightBorder
	}

	var rows []string
	rows = append(rows, emptyRow)

	if s.loading {
		loadingText := primaryBoldStyle.Render("Loading token usage…")
		gap := menuContentWidth - lipgloss.Width(loadingText) - 2
		if gap < 0 {
			gap = 0
		}
		rows = append(rows, leftBorder+"  "+loadingText+strings.Repeat(" ", gap)+rightBorder)
		rows = append(rows, emptyRow)
		hintText := faint.Render("Usage data is read from ~/.claude/usage/")
		hintGap := menuContentWidth - lipgloss.Width(hintText) - 2
		if hintGap < 0 {
			hintGap = 0
		}
		rows = append(rows, leftBorder+"  "+hintText+strings.Repeat(" ", hintGap)+rightBorder)
		rows = append(rows, emptyRow)
		return rows
	}

	if s.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
		errText := errStyle.Render("Error: " + s.err.Error())
		gap := menuContentWidth - lipgloss.Width(errText) - 2
		if gap < 0 {
			gap = 0
		}
		rows = append(rows, leftBorder+"  "+errText+strings.Repeat(" ", gap)+rightBorder)
		rows = append(rows, emptyRow)
		return rows
	}

	if len(s.months) == 0 {
		noDataText := muted.Render("No usage data found yet.")
		gap := menuContentWidth - lipgloss.Width(noDataText) - 2
		if gap < 0 {
			gap = 0
		}
		rows = append(rows, leftBorder+"  "+noDataText+strings.Repeat(" ", gap)+rightBorder)
		rows = append(rows, emptyRow)
		hintText := faint.Render("Usage data is read from ~/.claude/usage/")
		hintGap := menuContentWidth - lipgloss.Width(hintText) - 2
		if hintGap < 0 {
			hintGap = 0
		}
		rows = append(rows, leftBorder+"  "+hintText+strings.Repeat(" ", hintGap)+rightBorder)
		rows = append(rows, emptyRow)
		return rows
	}

	// Column header row.
	header := lipgloss.NewStyle().Foreground(m.theme.Dim).Bold(true)
	hdr := "    " + header.Render(fmt.Sprintf("%-8s %8s %8s %8s %8s %9s",
		"Month", "Input", "Output", "Cache W", "Cache R", "Total"))
	hdrGap := menuContentWidth - lipgloss.Width(hdr)
	if hdrGap < 0 {
		hdrGap = 0
	}
	rows = append(rows, leftBorder+hdr+strings.Repeat(" ", hdrGap)+rightBorder)

	// Separator
	sepRow := leftBorder + strings.Repeat(" ", menuContentWidth) + rightBorder
	rows = append(rows, sepRow)

	// Month rows (all months, not windowed — this is a summary panel not a scroller).
	grandTotal := statsGrandTotal(s.months)
	allTotal := grandTotal.Total()
	if allTotal < 1 {
		allTotal = 1
	}
	for _, mu := range s.months {
		frac := float64(mu.Total()) / float64(allTotal)
		pct := int(frac*100 + 0.5)
		monthCost, allPriced := mu.CostUSD()
		costStr := dollarFmt(monthCost)
		if !allPriced {
			costStr = "~" + costStr
		}

		dataLine := "    " + numStyle.Render(fmt.Sprintf("%-8s %8s %8s %8s %8s",
			mu.Month,
			humanizeTokens(mu.Input),
			humanizeTokens(mu.Output),
			humanizeTokens(mu.CacheWrite),
			humanizeTokens(mu.CacheRead))) + " " +
			primaryBoldStyle.Render(fmt.Sprintf("%9s", humanizeTokens(mu.Total())))
		dataGap := menuContentWidth - lipgloss.Width(dataLine)
		if dataGap < 0 {
			dataGap = 0
		}
		rows = append(rows, leftBorder+dataLine+strings.Repeat(" ", dataGap)+rightBorder)

		// Bar + percent + cost on line below the data.
		gaugeStr := statsGauge(frac, lipgloss.NewStyle().Foreground(m.theme.Primary), dimStyle)
		barLine := "    " + gaugeStr + " " + faint.Render(fmt.Sprintf("%3d%%", pct))
		costPad := menuContentWidth - lipgloss.Width(barLine) - lipgloss.Width(costStr) - 1
		if costPad < 1 {
			costPad = 1
		}
		barRow := leftBorder + barLine + strings.Repeat(" ", costPad) + primaryBoldStyle.Render(costStr) + " " + rightBorder
		rows = append(rows, barRow)
	}

	// Grand total row.
	rows = append(rows, sepRow)
	g := grandTotal
	grandCost := 0.0
	grandAllPriced := true
	for _, mu := range s.months {
		c, ap := mu.CostUSD()
		grandCost += c
		if !ap {
			grandAllPriced = false
		}
	}
	grandCostStr := dollarFmt(grandCost)
	if !grandAllPriced {
		grandCostStr = "~" + grandCostStr
	}

	rows = append(rows, itemRow("Total", humanizeTokens(g.Total()), primaryBoldStyle, primaryBoldStyle))
	rows = append(rows, itemRow("Est. cost", grandCostStr, header, primaryBoldStyle))
	rows = append(rows, emptyRow)

	return rows
}

// renderStatsBox renders the Stats tab: shared chrome (top border + title row +
// tab bar + separator) followed by stats content rows + bottom border + help row.
func (m *MainMenuModel) renderStatsBox() string {
	top, separator, bottom, leftBorder, rightBorder := m.boxBorders()

	var lines []string
	lines = append(lines, top)
	lines = append(lines, m.renderTitleRow(leftBorder, rightBorder))
	lines = append(lines, m.renderTabBar(leftBorder, rightBorder))
	lines = append(lines, separator)
	lines = append(lines, m.renderStatsRows(leftBorder, rightBorder)...)
	lines = append(lines, bottom)
	lines = append(lines, m.renderHelpRow())
	return strings.Join(lines, "\n")
}
