package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ghost-tab-tui",
	Short: "Interactive TUI components for Ghost Tab",
	Long:  "Provides terminal UI components for Ghost Tab project selector, AI tool picker, and settings menu.",
}

func init() {
	// Subcommands will be added here
}
