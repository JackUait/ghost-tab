package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackuait/ghost-tab/internal/models"
)

// TerminalSelectorModel is a Bubbletea model for selecting a terminal emulator.
type TerminalSelectorModel struct {
	terminals          []models.Terminal
	current            string
	cursor             int
	selected           *models.Terminal
	installRequest     string
	installRequestCask string
	quitting           bool
	width              int
}

// NewTerminalSelector creates a new terminal selector.
// current is the name of the currently saved terminal preference (may be empty).
func NewTerminalSelector(terminals []models.Terminal, current string) TerminalSelectorModel {
	return TerminalSelectorModel{
		terminals: terminals,
		current:   current,
	}
}

func (m TerminalSelectorModel) Init() tea.Cmd {
	return nil
}

// Cursor returns the current cursor position.
func (m TerminalSelectorModel) Cursor() int {
	return m.cursor
}

// InstallRequest returns the terminal name requested for install, or empty string.
func (m TerminalSelectorModel) InstallRequest() string {
	return m.installRequest
}

// InstallRequestCask returns the cask name for the install request, or empty string.
func (m TerminalSelectorModel) InstallRequestCask() string {
	return m.installRequestCask
}

func (m TerminalSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEscape:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyUp:
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.terminals) - 1
			}
			return m, nil

		case tea.KeyDown:
			m.cursor++
			if m.cursor >= len(m.terminals) {
				m.cursor = 0
			}
			return m, nil

		case tea.KeyEnter:
			if m.cursor < len(m.terminals) {
				if m.terminals[m.cursor].Installed {
					t := m.terminals[m.cursor]
					m.selected = &t
					m.quitting = true
					return m, tea.Quit
				}
				// Trigger install for uninstalled terminals
				m.installRequest = m.terminals[m.cursor].Name
				m.installRequestCask = m.terminals[m.cursor].CaskName
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil

		case tea.KeyRunes:
			if len(msg.Runes) == 1 {
				r := TranslateRune(msg.Runes[0])
				switch r {
				case 'j':
					m.cursor++
					if m.cursor >= len(m.terminals) {
						m.cursor = 0
					}
					return m, nil
				case 'k':
					m.cursor--
					if m.cursor < 0 {
						m.cursor = len(m.terminals) - 1
					}
					return m, nil
				case 'i':
					if m.cursor < len(m.terminals) && !m.terminals[m.cursor].Installed {
						m.installRequest = m.terminals[m.cursor].Name
						m.installRequestCask = m.terminals[m.cursor].CaskName
						m.quitting = true
						return m, tea.Quit
					}
					return m, nil
				}
			}
		}
	}

	return m, nil
}

func (m TerminalSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	borderColor := titleStyle.GetForeground()

	boxWidth := 56
	if m.width > 0 && m.width < boxWidth {
		boxWidth = m.width
	}

	innerWidth := boxWidth - 4 // border + padding
	if innerWidth < 20 {
		innerWidth = 20
	}

	installedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	notInstalledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	currentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))

	var content strings.Builder

	for i, t := range m.terminals {
		// Cursor + terminal name
		var line string
		if i == m.cursor {
			line = selectedItemStyle.Render(fmt.Sprintf(" %s %s", "\u25b8", t.DisplayName))
		} else {
			line = fmt.Sprintf("   %s", t.DisplayName)
		}

		// Status right-aligned
		var statusStr string
		if t.Installed {
			statusStr = installedStyle.Render("✓ installed")
			if m.current != "" && t.Name == m.current {
				statusStr += "  " + currentStyle.Render("★ current")
			}
		} else {
			statusStr = notInstalledStyle.Render("○ not installed")
		}

		titleVisible := lipgloss.Width(line)
		statusVisible := lipgloss.Width(statusStr)
		gap := innerWidth - titleVisible - statusVisible
		if gap < 2 {
			gap = 2
		}
		line += strings.Repeat(" ", gap) + statusStr

		content.WriteString(line)
		if i < len(m.terminals)-1 {
			content.WriteString("\n")
		}
	}

	// Hint text
	content.WriteString("\n\n")
	hint := " \u2191/\u2193 navigate \u2022 "
	if m.cursor < len(m.terminals) && m.terminals[m.cursor].Installed {
		hint += "Enter select \u2022 "
	} else {
		hint += "Enter install \u2022 "
	}
	hint += "Esc cancel"
	content.WriteString(hintStyle.Render(hint))

	// Bordered box
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 1).
		Width(boxWidth)

	box := borderStyle.Render(content.String())

	// Overlay title on the top border
	title := " Select Terminal "
	titleRendered := lipgloss.NewStyle().
		Foreground(borderColor).
		Bold(true).
		Render(title)

	lines := strings.Split(box, "\n")
	if len(lines) > 0 {
		topRunes := []rune(lines[0])
		titleRunes := []rune(titleRendered)
		insertPos := 2

		if len(topRunes) > insertPos+len(titleRunes) {
			result := make([]rune, 0, len(topRunes))
			result = append(result, topRunes[:insertPos]...)
			result = append(result, titleRunes...)
			result = append(result, topRunes[insertPos+len(titleRunes):]...)
			lines[0] = string(result)
		}
		box = strings.Join(lines, "\n")
	}

	return box
}

// Selected returns the selected terminal, or nil if none was selected.
func (m TerminalSelectorModel) Selected() *models.Terminal {
	return m.selected
}
