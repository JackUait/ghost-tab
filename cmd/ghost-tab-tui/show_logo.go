package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/tui"
	"github.com/jackuait/ghost-tab/internal/util"
	"github.com/spf13/cobra"
)

var showLogoCmd = &cobra.Command{
	Use:   "show-logo",
	Short: "Display animated Ghost Tab logo",
	Long:  "Shows an animated ghost mascot for the current AI tool (no JSON output)",
	RunE:  runShowLogo,
}

func init() {
	rootCmd.AddCommand(showLogoCmd)
}

func runShowLogo(cmd *cobra.Command, args []string) error {
	tui.ApplyTheme(tui.ThemeForTool(aiToolFlag))

	model := tui.NewLogo(aiToolFlag)

	ttyOpts, cleanup, err := util.TUITeaOptions()
	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	defer cleanup()

	opts := append([]tea.ProgramOption{tea.WithAltScreen()}, ttyOpts...)
	p := tea.NewProgram(model, opts...)

	_, err = p.Run()
	return err
}
