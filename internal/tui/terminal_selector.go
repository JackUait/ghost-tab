package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackuait/ghost-tab/internal/models"
)

// TerminalSelectorModel is a Bubbletea model for selecting a terminal emulator.
type TerminalSelectorModel struct {
	terminals      []models.Terminal
	current        string
	cursor         int
	selected       *models.Terminal
	installRequest string
	quitting       bool
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

func (m TerminalSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
			if m.cursor < len(m.terminals) && m.terminals[m.cursor].Installed {
				t := m.terminals[m.cursor]
				m.selected = &t
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

	var b strings.Builder

	b.WriteString(titleStyle.Render("Select Terminal"))
	b.WriteString("\n\n")

	installedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	notInstalledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	currentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))

	for i, t := range m.terminals {
		// Cursor indicator
		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render("  ❯ "))
		} else {
			b.WriteString("    ")
		}

		// Terminal display name
		name := t.DisplayName
		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render(name))
		} else {
			b.WriteString(name)
		}

		// Status indicators (right-aligned with padding)
		padding := 28 - len(name)
		if padding < 2 {
			padding = 2
		}
		b.WriteString(strings.Repeat(" ", padding))

		if t.Installed {
			b.WriteString(installedStyle.Render("✓ installed"))
			// Current marker
			if m.current != "" && t.Name == m.current {
				b.WriteString("   ")
				b.WriteString(currentStyle.Render("★ current"))
			}
		} else {
			b.WriteString(notInstalledStyle.Render("○ not installed"))
		}

		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("  ↑↓ navigate  Enter select  i install  Esc cancel"))

	return b.String()
}

// Selected returns the selected terminal, or nil if none was selected.
func (m TerminalSelectorModel) Selected() *models.Terminal {
	return m.selected
}

