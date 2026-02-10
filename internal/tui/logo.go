package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var logoFrames = []string{
	`
   ____  _               _     _____     _
  / ___|| |__   ___  ___| |_  |_   _|_ _| |__
 | |  _ | '_ \ / _ \/ __| __|   | |/ _  | '_ \
 | |_| || | | | (_) \__ \ |_    | | (_| | |_) |
  \____||_| |_|\___/|___/\__|   |_|\__,_|_.__/
`,
	`
   ____  _               _     _____     _
  / ___|| |__   ___  ___| |_  |_   _|_ _| |__
 | |  _ | '_ \ / _ \/ __| __|   | |/ _  | '_ \
 | |_| || | | | (_) \__ \ |_    | | (_| | |_) |
  \____||_| |_|\___/|___/\__|   |_|\__,_|_.__/

`,
}

type logoTickMsg time.Time

type LogoModel struct {
	frame    int
	quitting bool
}

func NewLogo() LogoModel {
	return LogoModel{}
}

func (m LogoModel) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
			return quitMsg{}
		}),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return logoTickMsg(t)
	})
}

type quitMsg struct{}

func (m LogoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case logoTickMsg:
		m.frame = (m.frame + 1) % len(logoFrames)
		if !m.quitting {
			return m, tickCmd()
		}
		return m, nil

	case quitMsg:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m LogoModel) View() string {
	if m.quitting {
		return ""
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	logo := logoFrames[m.frame]
	return style.Render(strings.TrimSpace(logo))
}
