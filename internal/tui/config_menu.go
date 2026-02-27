package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ConfigMenuItem struct {
	ItemTitle   string
	ItemDesc    string
	Action      string
}

func (i ConfigMenuItem) Title() string       { return i.ItemTitle }
func (i ConfigMenuItem) Description() string { return i.ItemDesc }
func (i ConfigMenuItem) FilterValue() string { return i.ItemTitle }

type ConfigMenuModel struct {
	list     list.Model
	selected *ConfigMenuItem
	quitting bool
}

func GetConfigMenuItems() []ConfigMenuItem {
	return []ConfigMenuItem{
		{ItemTitle: "Manage Terminals", ItemDesc: "Add, remove, or switch terminal emulator", Action: "manage-terminals"},
		{ItemTitle: "Manage Projects", ItemDesc: "Add or remove projects", Action: "manage-projects"},
		{ItemTitle: "Select AI Tools", ItemDesc: "Choose default AI tools", Action: "select-ai-tools"},
		{ItemTitle: "Display Settings", ItemDesc: "Ghost display, tab title, and more", Action: "display-settings"},
		{ItemTitle: "Reinstall / Update", ItemDesc: "Re-run the installer", Action: "reinstall"},
		{ItemTitle: "Quit", ItemDesc: "Exit configuration", Action: "quit"},
	}
}

func NewConfigMenu() ConfigMenuModel {
	menuItems := GetConfigMenuItems()
	items := make([]list.Item, len(menuItems))
	for i, item := range menuItems {
		items[i] = item
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Ghost Tab Configuration"
	l.Styles.Title = titleStyle

	return ConfigMenuModel{
		list: l,
	}
}

func (m ConfigMenuModel) Init() tea.Cmd {
	return nil
}

func (m ConfigMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.selected = &ConfigMenuItem{Action: "quit"}
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if item, ok := m.list.SelectedItem().(ConfigMenuItem); ok {
				m.selected = &item
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ConfigMenuModel) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

func (m ConfigMenuModel) Selected() *ConfigMenuItem {
	return m.selected
}
