package main

import (
	"fmt"

	"github.com/jackuait/ghost-tab/internal/claudeaccount"
	"github.com/spf13/cobra"
)

var (
	caList       string
	caAccountDir string
	caPointer    string
	caDir        string
	caLabel      string
)

var claudeAccountCmd = &cobra.Command{
	Use:   "claude-account",
	Short: "Create and remove native Claude login accounts",
	Long:  "Mutation commands for native Claude logins (each isolated by its own CLAUDE_CONFIG_DIR); the single source of truth shared by the menu and the add-login flow",
}

var claudeAccountAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Register a new Claude account and print its dir name",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := claudeaccount.Add(caList, caAccountDir, caLabel)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), dir)
		return nil
	},
}

var claudeAccountRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a Claude account and clear the pointer if it was active",
	RunE: func(cmd *cobra.Command, args []string) error {
		return claudeaccount.Remove(caList, caAccountDir, caPointer, caDir)
	},
}

func init() {
	claudeAccountAddCmd.Flags().StringVar(&caList, "list", "", "Path to accounts list (label:dir)")
	claudeAccountAddCmd.Flags().StringVar(&caAccountDir, "accounts-dir", "", "Path to accounts directory")
	claudeAccountAddCmd.Flags().StringVar(&caLabel, "label", "", "Display label for the new account")

	claudeAccountRemoveCmd.Flags().StringVar(&caList, "list", "", "Path to accounts list (label:dir)")
	claudeAccountRemoveCmd.Flags().StringVar(&caAccountDir, "accounts-dir", "", "Path to accounts directory")
	claudeAccountRemoveCmd.Flags().StringVar(&caPointer, "pointer", "", "Path to active account pointer file")
	claudeAccountRemoveCmd.Flags().StringVar(&caDir, "dir", "", "Dir name of the account to remove")

	claudeAccountCmd.AddCommand(claudeAccountAddCmd, claudeAccountRemoveCmd)
	rootCmd.AddCommand(claudeAccountCmd)
}
