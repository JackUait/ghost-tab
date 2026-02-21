package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackuait/ghost-tab/internal/models"
)

// BranchDeletedMsg is sent after an async branch deletion completes.
type BranchDeletedMsg struct {
	Branch string
	Err    error
}

// BranchPickerModel lets the user pick a branch from a filterable list.
// It renders with box-drawing borders matching the main menu style.
type BranchPickerModel struct {
	allBranches    []string
	filtered       []string
	filtering      bool   // whether filter mode is active (activated by '/')
	filterText     string
	cursor         int // index in filtered list
	offset         int // scroll offset for visible window
	selected       *string
	quitting       bool
	width          int
	height         int
	theme          AIToolTheme
	projectPath    string
	deleteMode     bool
	deleteSelected int
	deleteOffset   int
	feedback       string
	feedbackIsErr  bool
}

// NewBranchPicker creates a branch picker with the given branch names, theme, and project path.
func NewBranchPicker(branches []string, theme AIToolTheme, projectPath string) BranchPickerModel {
	filtered := make([]string, len(branches))
	copy(filtered, branches)

	return BranchPickerModel{
		allBranches: branches,
		filtered:    filtered,
		theme:       theme,
		projectPath: projectPath,
	}
}

func (m BranchPickerModel) Init() tea.Cmd {
	return nil
}

func (m BranchPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.clampScroll()
		return m, nil

	case BranchDeletedMsg:
		m.deleteMode = false
		if msg.Err != nil {
			m.feedback = msg.Err.Error()
			m.feedbackIsErr = true
		} else {
			m.feedback = "Deleted " + msg.Branch
			m.feedbackIsErr = false
			m.removeBranch(msg.Branch)
		}
		return m, nil

	case tea.KeyMsg:
		// Clear feedback on any keypress after deletion
		if m.feedback != "" {
			m.feedback = ""
			m.feedbackIsErr = false
			return m, nil
		}

		if m.deleteMode {
			return m.updateDeleteMode(msg)
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.filtering {
				m.filtering = false
				m.filterText = ""
				m.applyFilter()
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				name := m.filtered[m.cursor]
				m.selected = &name
			}
			m.quitting = true
			return m, tea.Quit

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
				m.clampScroll()
			}
			return m, nil

		case tea.KeyDown:
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				m.clampScroll()
			}
			return m, nil

		case tea.KeyBackspace:
			if m.filtering && len(m.filterText) > 0 {
				m.filterText = m.filterText[:len(m.filterText)-1]
				m.applyFilter()
			}
			return m, nil

		case tea.KeyRunes:
			r := msg.Runes[0]
			// '/' activates filter mode
			if !m.filtering && r == '/' {
				m.filtering = true
				return m, nil
			}
			// When not filtering, handle command keys
			if !m.filtering {
				if r == 'k' {
					if m.cursor > 0 {
						m.cursor--
						m.clampScroll()
					}
					return m, nil
				}
				if r == 'j' {
					if m.cursor < len(m.filtered)-1 {
						m.cursor++
						m.clampScroll()
					}
					return m, nil
				}
				if r == 'd' && len(m.filtered) > 0 {
					m.deleteMode = true
					m.deleteSelected = 0
					return m, nil
				}
				return m, nil
			}
			// In filter mode, add to filter text
			m.filterText += string(msg.Runes)
			m.applyFilter()
			return m, nil
		}
	}

	return m, nil
}

func (m BranchPickerModel) updateDeleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.deleteMode = false
		return m, nil

	case tea.KeyUp:
		if m.deleteSelected > 0 {
			m.deleteSelected--
		} else {
			m.deleteSelected = len(m.allBranches) - 1
		}
		m.clampDeleteScroll()
		return m, nil

	case tea.KeyDown:
		if m.deleteSelected < len(m.allBranches)-1 {
			m.deleteSelected++
		} else {
			m.deleteSelected = 0
		}
		m.clampDeleteScroll()
		return m, nil

	case tea.KeyEnter:
		if m.deleteSelected < len(m.allBranches) {
			branch := m.allBranches[m.deleteSelected]
			projectPath := m.projectPath
			return m, func() tea.Msg {
				err := models.DeleteBranch(projectPath, branch)
				return BranchDeletedMsg{Branch: branch, Err: err}
			}
		}
		return m, nil

	case tea.KeyRunes:
		r := msg.Runes[0]
		switch {
		case r == 'q' || r == 'Q':
			m.deleteMode = false
			return m, nil
		case r == 'k':
			if m.deleteSelected > 0 {
				m.deleteSelected--
			} else {
				m.deleteSelected = len(m.allBranches) - 1
			}
			m.clampDeleteScroll()
			return m, nil
		case r == 'j':
			if m.deleteSelected < len(m.allBranches)-1 {
				m.deleteSelected++
			} else {
				m.deleteSelected = 0
			}
			m.clampDeleteScroll()
			return m, nil
		}
	}
	return m, nil
}

func (m *BranchPickerModel) removeBranch(branch string) {
	var newAll []string
	for _, b := range m.allBranches {
		if b != branch {
			newAll = append(newAll, b)
		}
	}
	m.allBranches = newAll
	m.applyFilter()
	if m.cursor >= len(m.filtered) && m.cursor > 0 {
		m.cursor = len(m.filtered) - 1
	}
	m.clampScroll()
}

func (m *BranchPickerModel) applyFilter() {
	if m.filterText == "" {
		m.filtered = make([]string, len(m.allBranches))
		copy(m.filtered, m.allBranches)
	} else {
		lower := strings.ToLower(m.filterText)
		m.filtered = nil
		for _, b := range m.allBranches {
			if strings.Contains(strings.ToLower(b), lower) {
				m.filtered = append(m.filtered, b)
			}
		}
	}
	m.cursor = 0
	m.offset = 0
}

func (m *BranchPickerModel) clampScroll() {
	visible := m.visibleItemCount()
	if visible <= 0 {
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

// visibleItemCount returns how many branch items fit in the select box.
// Box layout: top border, title row, separator, filter row, empty row,
// ... items ..., empty row, separator, help row, bottom border.
// That's 9 fixed rows. Items get the rest.
func (m BranchPickerModel) visibleItemCount() int {
	count := m.height - 9
	if count < 1 {
		count = 1
	}
	return count
}

// deleteVisibleCount returns how many branch items fit in the delete box.
// Box layout: top border, title row, separator, empty row,
// ... items ..., empty row, separator, help row, bottom border.
// That's 8 fixed rows. Items get the rest.
func (m BranchPickerModel) deleteVisibleCount() int {
	count := m.height - 8
	if count < 1 {
		count = 1
	}
	return count
}

func (m *BranchPickerModel) clampDeleteScroll() {
	visible := m.deleteVisibleCount()
	if visible <= 0 {
		return
	}
	if m.deleteSelected < m.deleteOffset {
		m.deleteOffset = m.deleteSelected
	}
	if m.deleteSelected >= m.deleteOffset+visible {
		m.deleteOffset = m.deleteSelected - visible + 1
	}
}

func (m BranchPickerModel) View() string {
	if m.quitting {
		return ""
	}

	if m.deleteMode {
		return m.renderDeleteBox()
	}

	return m.renderSelectBox()
}

func (m BranchPickerModel) renderSelectBox() string {
	dimStyle := lipgloss.NewStyle().Foreground(m.theme.Dim)
	primaryStyle := lipgloss.NewStyle().Foreground(m.theme.Primary)
	primaryBoldStyle := lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true)
	textStyle := lipgloss.NewStyle().Foreground(m.theme.Text)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("247"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76"))

	hLine := strings.Repeat("\u2500", menuInnerWidth)
	topBorder := dimStyle.Render("\u250c" + hLine + "\u2510")
	separator := dimStyle.Render("\u251c" + hLine + "\u2524")
	bottomBorder := dimStyle.Render("\u2514" + hLine + "\u2518")
	leftBorder := dimStyle.Render("\u2502") + strings.Repeat(" ", menuPadding)
	rightBorder := strings.Repeat(" ", menuPadding) + dimStyle.Render("\u2502")

	emptyRow := leftBorder + strings.Repeat(" ", menuContentWidth) + rightBorder

	var lines []string

	// Top border
	lines = append(lines, topBorder)

	// Title row
	title := primaryBoldStyle.Render("\u2b21  Select Branch")
	titlePadding := menuContentWidth - lipgloss.Width(title) - 1
	if titlePadding < 0 {
		titlePadding = 0
	}
	lines = append(lines, leftBorder+" "+title+strings.Repeat(" ", titlePadding)+rightBorder)

	// Separator
	lines = append(lines, separator)

	// Filter row
	var filterContent string
	if m.filtering {
		filterPrompt := dimStyle.Render("/")
		if m.filterText == "" {
			filterContent = "  " + filterPrompt + " " + dimStyle.Render("type to filter...") + dimStyle.Render("│")
		} else {
			filterContent = "  " + filterPrompt + " " + textStyle.Render(m.filterText) + dimStyle.Render("│")
		}
	} else {
		filterContent = "  " + dimStyle.Render("/ to filter")
	}
	filterPadding := menuContentWidth - lipgloss.Width(filterContent)
	if filterPadding < 0 {
		filterPadding = 0
	}
	lines = append(lines, leftBorder+filterContent+strings.Repeat(" ", filterPadding)+rightBorder)

	// Empty line before items
	lines = append(lines, emptyRow)

	// Branch items
	visible := m.visibleItemCount()
	if len(m.filtered) == 0 {
		noItems := "  " + dimStyle.Render("No matching branches")
		noItemsPadding := menuContentWidth - lipgloss.Width(noItems)
		if noItemsPadding < 0 {
			noItemsPadding = 0
		}
		lines = append(lines, leftBorder+noItems+strings.Repeat(" ", noItemsPadding)+rightBorder)
	} else {
		end := m.offset + visible
		if end > len(m.filtered) {
			end = len(m.filtered)
		}
		for i := m.offset; i < end; i++ {
			branch := m.filtered[i]
			selected := i == m.cursor

			truncBranch := TruncateMiddle(branch, menuContentWidth-7)

			var row string
			if selected {
				marker := primaryBoldStyle.Render("\u258e")
				branchText := primaryBoldStyle.Render(truncBranch)
				content := "  " + marker + " " + branchText
				padding := menuContentWidth - lipgloss.Width(content)
				if padding < 0 {
					padding = 0
				}
				row = leftBorder + content + strings.Repeat(" ", padding) + rightBorder
			} else {
				branchText := primaryStyle.Render(truncBranch)
				content := "    " + branchText
				padding := menuContentWidth - lipgloss.Width(content)
				if padding < 0 {
					padding = 0
				}
				row = leftBorder + content + strings.Repeat(" ", padding) + rightBorder
			}
			lines = append(lines, row)
		}
	}

	// Empty line after items
	lines = append(lines, emptyRow)

	// Separator before help
	lines = append(lines, separator)

	// Help row
	var helpContent string
	if m.feedback != "" {
		if m.feedbackIsErr {
			helpContent = errorStyle.Render(m.feedback)
		} else {
			helpContent = successStyle.Render(m.feedback)
		}
	} else {
		helpText := "\u2191/k up \u00b7 \u2193/j down \u00b7 enter select \u00b7 / filter \u00b7 d delete \u00b7 esc back"
		helpContent = helpStyle.Render(helpText)
	}
	helpPadding := menuContentWidth - lipgloss.Width(helpContent) - 1
	if helpPadding < 0 {
		helpPadding = 0
	}
	lines = append(lines, leftBorder+" "+helpContent+strings.Repeat(" ", helpPadding)+rightBorder)

	// Bottom border
	lines = append(lines, bottomBorder)

	return m.centerBox(lines)
}

func (m BranchPickerModel) renderDeleteBox() string {
	dimStyle := lipgloss.NewStyle().Foreground(m.theme.Dim)
	primaryBoldStyle := lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("247"))
	deleteHighlight := lipgloss.NewStyle().Background(lipgloss.Color("196")).Foreground(lipgloss.Color("15"))

	hLine := strings.Repeat("\u2500", menuInnerWidth)
	topBorder := dimStyle.Render("\u250c" + hLine + "\u2510")
	separator := dimStyle.Render("\u251c" + hLine + "\u2524")
	bottomBorder := dimStyle.Render("\u2514" + hLine + "\u2518")
	leftBorder := dimStyle.Render("\u2502") + strings.Repeat(" ", menuPadding)
	rightBorder := strings.Repeat(" ", menuPadding) + dimStyle.Render("\u2502")
	emptyRow := leftBorder + strings.Repeat(" ", menuContentWidth) + rightBorder

	var lines []string

	lines = append(lines, topBorder)

	// Title row with "· Delete" suffix
	title := primaryBoldStyle.Render("\u2b21  Select Branch")
	titleContent := title + " " + dimStyle.Render("\u00b7 Delete")
	titlePadding := menuContentWidth - lipgloss.Width(titleContent) - 1
	if titlePadding < 0 {
		titlePadding = 0
	}
	lines = append(lines, leftBorder+" "+titleContent+strings.Repeat(" ", titlePadding)+rightBorder)
	lines = append(lines, separator)
	lines = append(lines, emptyRow)

	// Branch items — selected in red highlight, others dimmed
	visible := m.deleteVisibleCount()
	end := m.deleteOffset + visible
	if end > len(m.allBranches) {
		end = len(m.allBranches)
	}
	for i := m.deleteOffset; i < end; i++ {
		branch := m.allBranches[i]
		selected := m.deleteSelected == i
		truncBranch := TruncateMiddle(branch, menuContentWidth-6)

		var row string
		if selected {
			nameText := deleteHighlight.Render(" " + truncBranch + " ")
			content := "  " + nameText
			padding := menuContentWidth - lipgloss.Width(content)
			if padding < 0 {
				padding = 0
			}
			row = leftBorder + content + strings.Repeat(" ", padding) + rightBorder
		} else {
			nameText := dimStyle.Render(truncBranch)
			content := "    " + nameText
			padding := menuContentWidth - lipgloss.Width(content)
			if padding < 0 {
				padding = 0
			}
			row = leftBorder + content + strings.Repeat(" ", padding) + rightBorder
		}
		lines = append(lines, row)
	}

	lines = append(lines, emptyRow)
	lines = append(lines, separator)

	// Help row
	helpText := "\u2191\u2193 navigate  \u23ce delete  Q cancel"
	helpContent := helpStyle.Render(helpText)
	helpPadding := menuContentWidth - lipgloss.Width(helpContent) - 1
	if helpPadding < 0 {
		helpPadding = 0
	}
	lines = append(lines, leftBorder+" "+helpContent+strings.Repeat(" ", helpPadding)+rightBorder)
	lines = append(lines, bottomBorder)

	return m.centerBox(lines)
}

func (m BranchPickerModel) centerBox(lines []string) string {
	box := strings.Join(lines, "\n")

	// Center horizontally
	if m.width > 0 {
		boxWidth := menuInnerWidth + 2
		leftPad := (m.width - boxWidth) / 2
		if leftPad > 0 {
			padStr := strings.Repeat(" ", leftPad)
			padded := make([]string, len(lines))
			for i, line := range lines {
				padded[i] = padStr + line
			}
			box = strings.Join(padded, "\n")
		}
	}

	// Center vertically
	if m.height > 0 {
		boxLines := strings.Count(box, "\n") + 1
		topPad := (m.height - boxLines) / 2
		if topPad > 0 {
			box = strings.Repeat("\n", topPad) + box
		}
	}

	return box
}

// Selected returns the selected branch name, or nil if cancelled.
func (m BranchPickerModel) Selected() *string {
	return m.selected
}
