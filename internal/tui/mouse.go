package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// This file holds the mouse layer for the main menu: hover (motion) and click
// (press) support for every interactive element so the interface is fully
// usable with the pointer as well as the keyboard.
//
// Coordinates: HitTest works in *box-relative* cells — (0,0) is the menu box's
// top-left "╭". The Update handler converts absolute pointer coordinates using
// menuOriginX/menuOriginY (recomputed every View). Keeping HitTest pure makes
// the row/column math directly unit-testable without rendering.
//
// Hover follows the focus ring: moving the pointer over a region moves keyboard
// focus (and the body/settings cursor) there, so every region reuses its
// existing focused appearance — no parallel highlight styling required. The one
// exception is the tab bar, whose highlight tracks the *active* tab, so a
// hovered-but-inactive tab is marked via hoverTab. The motion handler does only
// pure Go work (no subprocess), keeping the all-motion event stream cheap.

// mouseRegion identifies which interactive element a pointer coordinate hit.
type mouseRegion int

const (
	regionNone mouseRegion = iota
	regionAccount
	regionAI
	regionSubscription
	regionTab
	regionBody
	regionSettings
)

// hitTarget is the element under a given box-relative coordinate.
type hitTarget struct {
	region mouseRegion
	index  int  // tab index, body flat-item index, or settings row index
	prev   bool // switcher rows: the pointer fell on the "previous"/left side
}

// menuBoxWidth is the fixed rendered width of the menu box (border + interior +
// padding + border): 1 + menuInnerWidth + 1.
const menuBoxWidth = menuInnerWidth + 2

// accountRowIndex returns the box-relative row of the LOGIN switcher, or -1 when
// it is not shown.
func (m *MainMenuModel) accountRowIndex() int {
	if m.accountRowCount() > 0 {
		return 1 // directly under the top border
	}
	return -1
}

// titleRowIndex returns the box-relative row of the AGENT/title row.
func (m *MainMenuModel) titleRowIndex() int {
	return 1 + m.accountRowCount()
}

// subscriptionRowIndex returns the box-relative row of the PLAN switcher, or -1
// when it is not shown.
func (m *MainMenuModel) subscriptionRowIndex() int {
	if m.subscriptionRowCount() > 0 {
		return m.titleRowIndex() + 1
	}
	return -1
}

// tabBarRowIndex returns the box-relative row of the Projects · Settings · Stats
// tab bar. Layout: top(0) → [account] → title → [subscription] → spacer → tabs.
func (m *MainMenuModel) tabBarRowIndex() int {
	return 3 + m.accountRowCount() + m.subscriptionRowCount()
}

// firstSettingsItemRow returns the box-relative row of settings item 0. After
// the tab bar comes the separator, a blank row, then the items.
func (m *MainMenuModel) firstSettingsItemRow() int {
	return m.tabBarRowIndex() + 3
}

// tabHitRanges returns the [start, end) box-relative column span of each tab
// label, mirroring renderTabBar's layout: a leading "│ " (cols 0,1) then each
// padded label (width = label+2) joined by a two-space separator.
func tabHitRanges() [][2]int {
	ranges := make([][2]int, len(menuTabLabels))
	col := 2 // skip the left border + the leading space
	for i, label := range menuTabLabels {
		w := lipgloss.Width(label) + 2
		ranges[i] = [2]int{col, col + w}
		col += w + 2 // two-space separator between tabs
	}
	return ranges
}

// switcherName returns the value label rendered in a switcher row, used to find
// the midpoint that separates the ◂ (prev) side from the ▸ (next) side.
func (m *MainMenuModel) switcherName(region mouseRegion) string {
	switch region {
	case regionAI:
		return AIToolDisplayName(m.CurrentAITool())
	case regionAccount:
		return m.CurrentClaudeAccountLabel()
	case regionSubscription:
		return m.CurrentClaudeConfigName()
	}
	return ""
}

// switcherPrev reports whether a box-relative X on a switcher row falls on the
// "previous" (left/◂) side. Every switcher caption is padded to width 6
// ("AGENT ", "LOGIN ", "PLAN  "), so the value name starts at column 10
// (col 0 border, col 1 space, cols 2..7 caption, col 8 ◂, col 9 space). The
// value's own midpoint cleanly divides the ◂ side from the ▸ side.
func (m *MainMenuModel) switcherPrev(boxX int, region mouseRegion) bool {
	const nameStartCol = 10
	mid := nameStartCol + lipgloss.Width(m.switcherName(region))/2
	return boxX < mid
}

// HitTest maps a box-relative coordinate to the interactive element under it.
func (m *MainMenuModel) HitTest(boxX, boxY int) hitTarget {
	// Switcher rows (only clickable when there is actually something to switch).
	if boxY == m.accountRowIndex() {
		return hitTarget{region: regionAccount, prev: m.switcherPrev(boxX, regionAccount)}
	}
	if boxY == m.titleRowIndex() && len(m.aiTools) > 1 {
		return hitTarget{region: regionAI, prev: m.switcherPrev(boxX, regionAI)}
	}
	if boxY == m.subscriptionRowIndex() && m.subscriptionFocusable() {
		return hitTarget{region: regionSubscription, prev: m.switcherPrev(boxX, regionSubscription)}
	}

	// Tab bar.
	if boxY == m.tabBarRowIndex() {
		for i, r := range tabHitRanges() {
			if boxX >= r[0] && boxX < r[1] {
				return hitTarget{region: regionTab, index: i}
			}
		}
		return hitTarget{region: regionNone}
	}

	// Tab body.
	switch m.activeTab {
	case TabProjects:
		if item := m.MapRowToItem(boxY); item >= 0 {
			return hitTarget{region: regionBody, index: item}
		}
	case TabSettings:
		if !m.settingsInputMode {
			if idx := m.mapRowToSettingsItem(boxY); idx >= 0 {
				return hitTarget{region: regionSettings, index: idx}
			}
		}
	}

	return hitTarget{region: regionNone}
}

// mapRowToSettingsItem maps a box-relative row to a settings item index, or -1.
func (m *MainMenuModel) mapRowToSettingsItem(boxY int) int {
	first := m.firstSettingsItemRow()
	idx := boxY - first
	if idx >= 0 && idx < m.settingsItemCount() {
		return idx
	}
	return -1
}

// handleMouse routes a mouse event to hover (motion), click (left press), or
// wheel scrolling. Overlay/input modes own all input, so the menu's hit-testing
// is suppressed while one is open.
func (m *MainMenuModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.modelMapOpen || m.accountMenuOpen || m.settingsInputMode ||
		m.inputMode != "" || m.deleteMode || m.staleConfirmIdx >= 0 {
		return m, nil
	}

	switch msg.Button {
	case tea.MouseButtonWheelDown:
		m.scrollBody(1)
		return m, nil
	case tea.MouseButtonWheelUp:
		m.scrollBody(-1)
		return m, nil
	}

	target := m.HitTest(msg.X-m.menuOriginX, msg.Y-m.menuOriginY)

	switch msg.Action {
	case tea.MouseActionMotion:
		m.applyHover(target)
		return m, nil
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft {
			m.applyHover(target) // sync focus/cursor even without a prior motion
			return m.clickTarget(target)
		}
	}
	return m, nil
}

// applyHover records what the pointer is over so the renderers can highlight it.
// Hover is a *separate* visual layer: it never moves keyboard focus or the
// selection cursor (that would hijack keyboard state and risk accidental
// activation). hoverTab is mirrored for the tab bar, whose own highlight tracks
// the active tab and so needs a distinct marker for a hovered-but-inactive tab.
func (m *MainMenuModel) applyHover(t hitTarget) {
	m.hover = t
	if t.region == regionTab {
		m.hoverTab = t.index
	} else {
		m.hoverTab = -1
	}
}

// isHovered reports whether the pointer is currently over the given region.
func (m *MainMenuModel) isHovered(r mouseRegion) bool {
	return m.hover.region == r
}

// clickTarget activates the element under a left-click, mirroring its keyboard
// action. Switchers cycle (and take focus so the highlight persists); the
// project body keeps its select-then-activate behavior so a stray click never
// launches a project; settings rows act immediately (toggles are reversible).
func (m *MainMenuModel) clickTarget(t hitTarget) (tea.Model, tea.Cmd) {
	switch t.region {
	case regionTab:
		m.activeTab = MenuTab(t.index)
		m.focus = FocusTabs
		m.hoverTab = -1
		if m.activeTab == TabStats {
			return m, m.ensureStatsLoad()
		}
		return m, nil
	case regionAI:
		m.focus = FocusAI
		m.CycleAITool(directionFor(t.prev))
		return m, nil
	case regionAccount:
		m.focus = FocusAccount
		m.CycleAccount(directionFor(t.prev))
		return m, nil
	case regionSubscription:
		m.focus = FocusSubscription
		m.CycleMainSubscription(directionFor(t.prev))
		return m, nil
	case regionBody:
		m.focus = FocusBody
		if m.selectedItem == t.index {
			// Clicking the already-selected row activates it (double-click-like).
			if cmd := m.selectCurrent(); cmd != nil {
				return m, cmd
			}
			return m, tea.Quit
		}
		m.selectedItem = t.index
		return m, nil
	case regionSettings:
		return m.clickSettings(t.index)
	}
	return m, nil
}

// clickSettings activates a settings row: the edit/manage rows open their flow,
// every other row cycles its value (same as the → key).
func (m *MainMenuModel) clickSettings(idx int) (tea.Model, tea.Cmd) {
	m.settingsSelected = idx
	m.focus = FocusBody
	loginIdx := m.settingsItemCount() - 1
	switch idx {
	case 4: // Default projects dir → inline edit
		return m.settingsEnter()
	case loginIdx: // Login → account management
		return m.settingsEnter()
	default:
		m.settingsValueRight()
		return m, nil
	}
}

// scrollBody moves the active tab's body cursor by delta (+1 down, -1 up), the
// mouse-wheel equivalent of ↑/↓ within the body.
func (m *MainMenuModel) scrollBody(delta int) {
	m.focus = FocusBody
	switch m.activeTab {
	case TabProjects:
		if delta > 0 {
			if m.selectedItem < m.TotalItems()-1 {
				m.MoveDown()
			}
		} else if m.selectedItem > 0 {
			m.MoveUp()
		}
	case TabSettings:
		if delta > 0 {
			if m.settingsSelected < m.settingsItemCount()-1 {
				m.settingsSelected++
			}
		} else if m.settingsSelected > 0 {
			m.settingsSelected--
		}
	case TabStats:
		if delta > 0 {
			m.statsScrollDown()
		} else if m.statsOffset > 0 {
			m.statsOffset--
		}
	}
}

// directionFor maps the prev/next flag to the "prev"/"next" strings the Cycle*
// helpers expect.
func directionFor(prev bool) string {
	if prev {
		return "prev"
	}
	return "next"
}
