package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	questionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	hintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// Clickable button labels for the confirm dialog (mouse parity for y/n).
const (
	confirmYesLabel = "[ Yes ]"
	confirmNoLabel  = "[ No ]"
	confirmGap      = "   " // spacing between the two buttons
)

type ConfirmDialogModel struct {
	Message   string
	Confirmed bool
	quitting  bool
	btnHover  int // 0 = none, 1 = Yes, 2 = No
}

func NewConfirmDialog(message string) ConfirmDialogModel {
	return ConfirmDialogModel{
		Message: message,
	}
}

func (m ConfirmDialogModel) Init() tea.Cmd {
	return nil
}

// confirmButtonRow returns the screen row of the Yes/No buttons. The view
// renders at the origin (AltScreen): the message, a blank line, then the buttons.
func (m ConfirmDialogModel) confirmButtonRow() int {
	return strings.Count(m.Message, "\n") + 2
}

// handleMouse gives the dialog pointer parity: hover highlights a button, and a
// left-click on Yes/No answers the prompt (the y/n keys).
func (m ConfirmDialogModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	hit := 0 // 1 = Yes, 2 = No
	if msg.Y == m.confirmButtonRow() {
		yesEnd := lipgloss.Width(confirmYesLabel)
		noStart := yesEnd + len(confirmGap)
		noEnd := noStart + lipgloss.Width(confirmNoLabel)
		switch {
		case msg.X >= 0 && msg.X < yesEnd:
			hit = 1
		case msg.X >= noStart && msg.X < noEnd:
			hit = 2
		}
	}
	switch msg.Action {
	case tea.MouseActionMotion:
		m.btnHover = hit
		return m, nil
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft && hit != 0 {
			m.Confirmed = hit == 1
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ConfirmDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEscape:
			m.Confirmed = false
			m.quitting = true
			return m, tea.Quit
		case tea.KeyRunes:
			if len(msg.Runes) == 1 {
				r := TranslateRune(msg.Runes[0])
				switch r {
				case 'y', 'Y':
					m.Confirmed = true
					m.quitting = true
					return m, tea.Quit
				case 'n', 'N':
					m.Confirmed = false
					m.quitting = true
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

func (m ConfirmDialogModel) View() string {
	if m.quitting {
		return ""
	}

	// Render the two buttons; the hovered one is reversed so it reads as the
	// pointer target. y/n still work from the keyboard.
	yesStyle, noStyle := hintStyle, hintStyle
	if m.btnHover == 1 {
		yesStyle = questionStyle.Copy().Reverse(true)
	}
	if m.btnHover == 2 {
		noStyle = questionStyle.Copy().Reverse(true)
	}
	// The buttons sit first on the row so their click columns stay fixed; the
	// "y/n" keyboard hint trails to the right.
	buttons := yesStyle.Render(confirmYesLabel) + confirmGap + noStyle.Render(confirmNoLabel) +
		"   " + hintStyle.Render("y/n")

	return fmt.Sprintf(
		"%s\n\n%s",
		questionStyle.Render(m.Message),
		buttons,
	)
}
