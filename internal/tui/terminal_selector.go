package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/models"
)

type terminalItem struct {
	terminal models.Terminal
}

func (i terminalItem) Title() string       { return i.terminal.String() }
func (i terminalItem) Description() string { return i.terminal.Name }
func (i terminalItem) FilterValue() string { return i.terminal.Name }

// TerminalSelectorModel is a Bubbletea model for selecting a terminal emulator.
type TerminalSelectorModel struct {
	list      list.Model
	terminals []models.Terminal
	selected  *models.Terminal
	quitting  bool
}

// NewTerminalSelector creates a new terminal selector with the given terminals.
func NewTerminalSelector(terminals []models.Terminal) TerminalSelectorModel {
	items := make([]list.Item, len(terminals))
	for i, t := range terminals {
		items[i] = terminalItem{terminal: t}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Terminal"
	l.Styles.Title = titleStyle

	return TerminalSelectorModel{
		list:      l,
		terminals: terminals,
	}
}

func (m TerminalSelectorModel) Init() tea.Cmd {
	return nil
}

func (m TerminalSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if item, ok := m.list.SelectedItem().(terminalItem); ok {
				if item.terminal.Installed {
					m.selected = &item.terminal
				}
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m TerminalSelectorModel) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

// Selected returns the selected terminal, or nil if none was selected.
func (m TerminalSelectorModel) Selected() *models.Terminal {
	return m.selected
}
