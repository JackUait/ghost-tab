package main

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/models"
	"github.com/jackuait/ghost-tab/internal/tui"
	"github.com/jackuait/ghost-tab/internal/util"
	"github.com/spf13/cobra"
)

var currentTerminalFlag string

var selectTerminalCmd = &cobra.Command{
	Use:   "select-terminal",
	Short: "Interactive terminal emulator selector",
	Long:  "Shows available terminal emulators and returns selected terminal as JSON",
	RunE:  runSelectTerminal,
}

func init() {
	selectTerminalCmd.Flags().StringVar(&currentTerminalFlag, "current", "", "Currently selected terminal name")
	rootCmd.AddCommand(selectTerminalCmd)
}

func runSelectTerminal(cmd *cobra.Command, args []string) error {
	tui.ApplyTheme(tui.ThemeForTool(aiToolFlag))

	terminals := models.DetectTerminals()

	model := tui.NewTerminalSelector(terminals, currentTerminalFlag)

	ttyOpts, cleanup, err := util.TUITeaOptions()
	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	defer cleanup()

	opts := append([]tea.ProgramOption{tea.WithAltScreen()}, ttyOpts...)
	p := tea.NewProgram(model, opts...)

	finalModel, runErr := p.Run()

	m := finalModel.(tui.TerminalSelectorModel)
	selected := m.Selected()
	installReq := m.InstallRequest()
	installReqCask := m.InstallRequestCask()

	// If the TUI errored but the user completed an action (install/select),
	// output the result anyway — bubbletea may error during cleanup even
	// after a successful interaction.
	if runErr != nil && installReq == "" && selected == nil {
		return fmt.Errorf("failed to run TUI: %w", runErr)
	}

	var result map[string]interface{}
	if installReq != "" {
		result = map[string]interface{}{
			"action":   "install",
			"terminal": installReq,
			"cask":     installReqCask,
			"selected": false,
		}
	} else if selected != nil {
		result = map[string]interface{}{
			"terminal": selected.Name,
			"selected": true,
		}
	} else {
		result = map[string]interface{}{"selected": false}
	}

	jsonOutput, _ := json.Marshal(result)
	fmt.Println(string(jsonOutput))

	return nil
}
