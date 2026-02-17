package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type branchItem struct {
	name string
}

func (i branchItem) Title() string       { return i.name }
func (i branchItem) Description() string { return "" }
func (i branchItem) FilterValue() string { return i.name }

// BranchPickerModel lets the user pick a branch from a filterable list.
type BranchPickerModel struct {
	list     list.Model
	selected *string
	quitting bool
}

// NewBranchPicker creates a branch picker with the given branch names.
func NewBranchPicker(branches []string) BranchPickerModel {
	items := make([]list.Item, len(branches))
	for i, b := range branches {
		items[i] = branchItem{name: b}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Branch"
	l.Styles.Title = titleStyle

	return BranchPickerModel{list: l}
}

func (m BranchPickerModel) Init() tea.Cmd {
	return nil
}

func (m BranchPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter:
			if item, ok := m.list.SelectedItem().(branchItem); ok {
				name := item.name
				m.selected = &name
			}
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m BranchPickerModel) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

// Selected returns the selected branch name, or nil if cancelled.
func (m BranchPickerModel) Selected() *string {
	return m.selected
}
