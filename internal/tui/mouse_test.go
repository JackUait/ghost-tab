package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackuait/ghost-tab/internal/models"
	"github.com/muesli/termenv"
)

// helper: build a model sized and rendered so menuOriginX/Y are populated.
func mouseTestModel(t *testing.T, projects []models.Project, aiTools []string) *MainMenuModel {
	t.Helper()
	m := NewMainMenu(projects, aiTools, aiTools[0], "none")
	m.width = 100
	m.height = 60
	_ = m.View() // populates menuOriginX / menuOriginY
	return m
}

func TestHitTest_tabBar_mapsColumnsToTabs(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.width = 100
	m.height = 60
	row := m.tabBarRowIndex()

	cases := []struct {
		boxX   int
		region mouseRegion
		index  int
	}{
		{2, regionTab, 0},  // Projects start
		{11, regionTab, 0}, // Projects end
		{13, regionNone, 0}, // separator gap
		{14, regionTab, 1}, // Settings start
		{23, regionTab, 1}, // Settings end
		{26, regionTab, 2}, // Stats start
		{32, regionTab, 2}, // Stats end
		{40, regionNone, 0}, // trailing padding
	}
	for _, c := range cases {
		got := m.HitTest(c.boxX, row)
		if got.region != c.region || (c.region == regionTab && got.index != c.index) {
			t.Errorf("HitTest(%d, %d) = {%v,%d}, want {%v,%d}", c.boxX, row, got.region, got.index, c.region, c.index)
		}
	}
}

func TestHitTest_aiRow_directionByX(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude", "opencode"}, "claude", "none")
	m.width = 100
	m.height = 60
	row := m.titleRowIndex()

	left := m.HitTest(8, row) // on the ◂ chevron
	if left.region != regionAI || !left.prev {
		t.Errorf("left click = {%v, prev=%v}, want {regionAI, prev=true}", left.region, left.prev)
	}
	right := m.HitTest(24, row) // past the value name, on ▸ side
	if right.region != regionAI || right.prev {
		t.Errorf("right click = {%v, prev=%v}, want {regionAI, prev=false}", right.region, right.prev)
	}
}

func TestHitTest_aiRow_singleToolNotClickable(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.width = 100
	m.height = 60
	got := m.HitTest(8, m.titleRowIndex())
	if got.region == regionAI {
		t.Errorf("single AI tool should not be an AI click target, got %v", got.region)
	}
}

func TestHitTest_body_mapsToItem(t *testing.T) {
	m := NewMainMenu([]models.Project{{Name: "alpha", Path: "/tmp/a"}}, []string{"claude"}, "claude", "none")
	m.width = 100
	m.height = 60
	// startRow = 6 + subscription(1) + account(0) = 7 → first project name row.
	got := m.HitTest(5, 7)
	if got.region != regionBody || got.index != 0 {
		t.Errorf("HitTest body = {%v,%d}, want {regionBody,0}", got.region, got.index)
	}
}

func TestUpdate_clickTab_switchesActiveTab(t *testing.T) {
	m := mouseTestModel(t, nil, []string{"claude"})
	row := m.tabBarRowIndex()
	// Click "Stats" (tab index 2): boxX within [26,33).
	msg := tea.MouseMsg{
		X:      m.menuOriginX + 28,
		Y:      m.menuOriginY + row,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	updated, _ := m.Update(msg)
	mm := updated.(*MainMenuModel)
	if mm.ActiveTab() != TabStats {
		t.Errorf("after click on Stats tab, activeTab = %v, want TabStats", mm.ActiveTab())
	}
}

func TestUpdate_clickAIChevron_cyclesTool(t *testing.T) {
	m := mouseTestModel(t, nil, []string{"claude", "opencode"})
	if m.CurrentAITool() != "claude" {
		t.Fatalf("precondition: current tool = %q, want claude", m.CurrentAITool())
	}
	// Click the right (▸) side of the AGENT row → next tool.
	msg := tea.MouseMsg{
		X:      m.menuOriginX + 30,
		Y:      m.menuOriginY + m.titleRowIndex(),
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	updated, _ := m.Update(msg)
	mm := updated.(*MainMenuModel)
	if mm.CurrentAITool() != "opencode" {
		t.Errorf("after clicking ▸ on AGENT row, tool = %q, want opencode", mm.CurrentAITool())
	}
}

func TestUpdate_motionOverTab_setsHoverNotFocus(t *testing.T) {
	m := mouseTestModel(t, nil, []string{"claude"})
	startFocus := m.focus
	row := m.tabBarRowIndex()
	msg := tea.MouseMsg{
		X:      m.menuOriginX + 16, // over Settings (tab 1)
		Y:      m.menuOriginY + row,
		Action: tea.MouseActionMotion,
	}
	updated, _ := m.Update(msg)
	mm := updated.(*MainMenuModel)
	if mm.hoverTab != 1 {
		t.Errorf("hoverTab = %d, want 1", mm.hoverTab)
	}
	// Hover is a visual-only layer: it must not steal keyboard focus...
	if mm.focus != startFocus {
		t.Errorf("hover changed focus to %v, want unchanged %v", mm.focus, startFocus)
	}
	// ...nor change the active tab (hover ≠ click).
	if mm.ActiveTab() != TabProjects {
		t.Errorf("hover changed activeTab to %v, want TabProjects", mm.ActiveTab())
	}
}

func TestUpdate_motionOverProject_doesNotMoveSelection(t *testing.T) {
	projects := []models.Project{
		{Name: "a", Path: "/tmp/a"},
		{Name: "b", Path: "/tmp/b"},
	}
	m := mouseTestModel(t, projects, []string{"claude"})
	// Hover the second project's name row (box row 9), selection starts at 0.
	msg := tea.MouseMsg{
		X:      m.menuOriginX + 5,
		Y:      m.menuOriginY + 9,
		Action: tea.MouseActionMotion,
	}
	updated, _ := m.Update(msg)
	mm := updated.(*MainMenuModel)
	if mm.SelectedItem() != 0 {
		t.Errorf("hover moved selection to %d, want 0 (hover must not move the cursor)", mm.SelectedItem())
	}
	if !mm.isHovered(regionBody) || mm.hover.index != 1 {
		t.Errorf("hover = {%v,%d}, want {regionBody,1}", mm.hover.region, mm.hover.index)
	}
}

func TestUpdate_wheelScrollsBodySelection(t *testing.T) {
	projects := []models.Project{
		{Name: "a", Path: "/tmp/a"},
		{Name: "b", Path: "/tmp/b"},
		{Name: "c", Path: "/tmp/c"},
	}
	m := mouseTestModel(t, projects, []string{"claude"})
	start := m.selectedItem
	wheel := tea.MouseMsg{
		X:      m.menuOriginX + 5,
		Y:      m.menuOriginY + 8,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
	}
	updated, _ := m.Update(wheel)
	mm := updated.(*MainMenuModel)
	if mm.selectedItem != start+1 {
		t.Errorf("after wheel down, selectedItem = %d, want %d", mm.selectedItem, start+1)
	}
}

func TestRenderTabBar_highlightsHoveredTab(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.width = 100
	m.height = 60
	_, _, _, lb, rb := m.boxBorders()
	plain := m.renderTabBar(lb, rb)
	m.hoverTab = 2 // Stats
	hovered := m.renderTabBar(lb, rb)
	if plain == hovered {
		t.Errorf("hovering a tab should change the rendered tab bar, but output was identical")
	}
}

func TestRender_hoverHighlights(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	// Two projects: flat items 0=alpha (selected), 1=beta, 2=add-project. Hover
	// targets must be *unselected* rows, since selection has its own treatment.
	projects := []models.Project{{Name: "alpha", Path: "/tmp/a"}, {Name: "beta", Path: "/tmp/b"}}
	cases := []struct {
		name  string
		tab   MenuTab
		hover hitTarget
	}{
		{"AGENT switcher row", TabProjects, hitTarget{region: regionAI}},
		{"project body row", TabProjects, hitTarget{region: regionBody, index: 1}},
		{"add-project row", TabProjects, hitTarget{region: regionBody, index: 2}},
		{"settings row", TabSettings, hitTarget{region: regionSettings, index: 1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := NewMainMenu(projects, []string{"claude", "opencode"}, "claude", "none")
			m.width = 100
			m.height = 60
			m.SetActiveTab(c.tab)
			plain := m.View()
			m.applyHover(c.hover)
			hovered := m.View()
			if plain == hovered {
				t.Errorf("hovering %s should change the rendered output, but it was identical", c.name)
			}
		})
	}
}
