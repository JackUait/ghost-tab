package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/models"
)

// MainMenuResult represents the JSON output when the main menu exits.
type MainMenuResult struct {
	Action string `json:"action"`
	Name   string `json:"name,omitempty"`
	Path   string `json:"path,omitempty"`
	AITool string `json:"ai_tool"`
}

// MenuLayout describes how the ghost and menu are arranged at a given terminal size.
type MenuLayout struct {
	GhostPosition string // "side", "above", "hidden"
	MenuWidth     int    // Always 48
	MenuHeight    int    // Calculated from items
	FirstItemRow  int    // Row offset of first item within menu box
}

// actionNames maps action item offsets to their action strings.
var actionNames = []string{"add-project", "delete-project", "open-once", "plain-terminal"}

// MainMenuModel is the Bubbletea model for the unified main menu.
type MainMenuModel struct {
	projects      []models.Project
	aiTools       []string
	selectedAI    int
	selectedItem  int
	ghostDisplay  string
	ghostSleeping bool
	bobStep       int
	sleepTimer    int
	width         int
	height        int
	theme         AIToolTheme
	quitting      bool
	result        *MainMenuResult
}

// NewMainMenu creates a new main menu model.
func NewMainMenu(projects []models.Project, aiTools []string, currentAI string, ghostDisplay string) *MainMenuModel {
	selectedAI := 0
	for i, tool := range aiTools {
		if tool == currentAI {
			selectedAI = i
			break
		}
	}

	return &MainMenuModel{
		projects:     projects,
		aiTools:      aiTools,
		selectedAI:   selectedAI,
		selectedItem: 0,
		ghostDisplay: ghostDisplay,
		theme:        ThemeForTool(currentAI),
	}
}

// SelectedItem returns the currently selected item index.
func (m *MainMenuModel) SelectedItem() int {
	return m.selectedItem
}

// TotalItems returns the total number of selectable items (projects + 4 actions).
func (m *MainMenuModel) TotalItems() int {
	return len(m.projects) + len(actionNames)
}

// CurrentAITool returns the name of the currently selected AI tool.
func (m *MainMenuModel) CurrentAITool() string {
	if len(m.aiTools) == 0 {
		return ""
	}
	return m.aiTools[m.selectedAI]
}

// CycleAITool cycles the AI tool selection forward ("next") or backward ("prev").
func (m *MainMenuModel) CycleAITool(direction string) {
	n := len(m.aiTools)
	if n <= 1 {
		return
	}
	if direction == "next" {
		m.selectedAI = (m.selectedAI + 1) % n
	} else {
		m.selectedAI = (m.selectedAI - 1 + n) % n
	}
	m.theme = ThemeForTool(m.aiTools[m.selectedAI])
}

// MoveUp moves the selection up by one, wrapping around.
func (m *MainMenuModel) MoveUp() {
	total := m.TotalItems()
	m.selectedItem = (m.selectedItem - 1 + total) % total
}

// MoveDown moves the selection down by one, wrapping around.
func (m *MainMenuModel) MoveDown() {
	total := m.TotalItems()
	m.selectedItem = (m.selectedItem + 1) % total
}

// JumpTo jumps to the given 1-indexed project number.
// Does nothing if n is out of range or beyond the number of projects.
func (m *MainMenuModel) JumpTo(n int) {
	if n < 1 || n > len(m.projects) {
		return
	}
	m.selectedItem = n - 1
}

// SetSize updates the stored terminal dimensions.
func (m *MainMenuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GhostDisplay returns the ghost display mode.
func (m *MainMenuModel) GhostDisplay() string {
	return m.ghostDisplay
}

// Result returns the menu result, or nil if the menu has not exited.
func (m *MainMenuModel) Result() *MainMenuResult {
	return m.result
}

// CalculateLayout determines how the ghost and menu should be arranged.
func (m *MainMenuModel) CalculateLayout(width, height int) MenuLayout {
	numProjects := len(m.projects)
	numSeparators := 0
	if numProjects > 0 {
		numSeparators = 1
	}
	totalItems := m.TotalItems()
	menuHeight := 7 + (totalItems * 2) + numSeparators
	menuWidth := 48

	ghostPosition := "hidden"
	// Side layout: width >= 48 + 3 + 28 + 3 = 82
	if width >= menuWidth+3+28+3 {
		ghostPosition = "side"
	} else if height >= menuHeight+15+2 {
		// Above layout: enough vertical space for ghost (15 lines) + gap (2)
		ghostPosition = "above"
	}

	return MenuLayout{
		GhostPosition: ghostPosition,
		MenuWidth:     menuWidth,
		MenuHeight:    menuHeight,
		FirstItemRow:  0,
	}
}

// selectCurrent produces a result for the currently selected item.
func (m *MainMenuModel) selectCurrent() {
	idx := m.selectedItem
	numProjects := len(m.projects)

	if idx < numProjects {
		m.result = &MainMenuResult{
			Action: "select-project",
			Name:   m.projects[idx].Name,
			Path:   m.projects[idx].Path,
			AITool: m.CurrentAITool(),
		}
	} else {
		actionIdx := idx - numProjects
		if actionIdx < len(actionNames) {
			m.result = &MainMenuResult{
				Action: actionNames[actionIdx],
				AITool: m.CurrentAITool(),
			}
		}
	}
	m.quitting = true
}

// setActionResult produces a result for the given action name.
func (m *MainMenuModel) setActionResult(action string) {
	m.result = &MainMenuResult{
		Action: action,
		AITool: m.CurrentAITool(),
	}
	m.quitting = true
}

// Init implements tea.Model. Returns nil for now (Task 5 will add tick commands).
func (m *MainMenuModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. Handles key bindings and window resize.
func (m *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.MoveUp()
			return m, nil
		case tea.KeyDown:
			m.MoveDown()
			return m, nil
		case tea.KeyLeft:
			m.CycleAITool("prev")
			return m, nil
		case tea.KeyRight:
			m.CycleAITool("next")
			return m, nil
		case tea.KeyEnter:
			m.selectCurrent()
			return m, tea.Quit
		case tea.KeyEsc:
			m.setActionResult("quit")
			return m, tea.Quit
		case tea.KeyCtrlC:
			m.setActionResult("quit")
			return m, tea.Quit
		case tea.KeyRunes:
			if len(msg.Runes) == 1 {
				return m.handleRune(msg.Runes[0])
			}
		}
	}

	return m, nil
}

// handleRune processes a single rune keypress.
func (m *MainMenuModel) handleRune(r rune) (tea.Model, tea.Cmd) {
	switch r {
	case 'j':
		m.MoveDown()
		return m, nil
	case 'k':
		m.MoveUp()
		return m, nil
	case 'a', 'A':
		m.setActionResult("add-project")
		return m, tea.Quit
	case 'd', 'D':
		m.setActionResult("delete-project")
		return m, tea.Quit
	case 'o', 'O':
		m.setActionResult("open-once")
		return m, tea.Quit
	case 'p', 'P':
		m.setActionResult("plain-terminal")
		return m, tea.Quit
	case 's', 'S':
		m.setActionResult("settings")
		return m, tea.Quit
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		n := int(r - '0')
		m.JumpTo(n)
		return m, nil
	}
	return m, nil
}

// View implements tea.Model. Placeholder for now (Task 4 will implement full rendering).
func (m *MainMenuModel) View() string {
	if m.quitting {
		return ""
	}
	return "Main Menu"
}
