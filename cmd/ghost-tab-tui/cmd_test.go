package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCmd_HasAIToolFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("ai-tool")
	if flag == nil {
		t.Fatal("Expected --ai-tool persistent flag to be registered")
	}
	if flag.DefValue != "claude" {
		t.Errorf("Expected default value %q, got %q", "claude", flag.DefValue)
	}
}

func TestRootCmd_SubcommandRegistered(t *testing.T) {
	subcommands := []string{
		"confirm",
		"show-logo",
		"select-project",
		"select-ai-tool",
		"add-project",
		"settings-menu",
		"main-menu",
		"multi-select-ai-tool",
	}

	for _, name := range subcommands {
		t.Run(name, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{name})
			if err != nil {
				t.Fatalf("Failed to find subcommand %q: %v", name, err)
			}
			if cmd.Name() != name {
				t.Errorf("Expected command name %q, got %q", name, cmd.Name())
			}
		})
	}
}

func TestConfirmCmd_RequiresArg(t *testing.T) {
	cmd, _, _ := rootCmd.Find([]string{"confirm"})
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided to confirm")
	}
}

func TestConfirmCmd_AcceptsOneArg(t *testing.T) {
	cmd, _, _ := rootCmd.Find([]string{"confirm"})
	err := cmd.Args(cmd, []string{"Delete?"})
	if err != nil {
		t.Errorf("Expected no error with 1 arg, got: %v", err)
	}
}

func TestSelectProjectCmd_HasProjectsFileFlag(t *testing.T) {
	cmd, _, _ := rootCmd.Find([]string{"select-project"})
	flag := cmd.Flags().Lookup("projects-file")
	if flag == nil {
		t.Fatal("Expected --projects-file flag on select-project")
	}
}

func TestMainMenuCmd_HasAllFlags(t *testing.T) {
	cmd, _, _ := rootCmd.Find([]string{"main-menu"})

	flags := []struct {
		name     string
		defValue string
	}{
		{"projects-file", ""},
		{"ai-tool", "claude"},
		{"ai-tools", "claude"},
		{"ghost-display", "animated"},
		{"tab-title", "full"},
		{"update-version", ""},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(f.name)
			if flag == nil {
				t.Fatalf("Expected --%s flag", f.name)
			}
			if flag.DefValue != f.defValue {
				t.Errorf("Expected default %q, got %q", f.defValue, flag.DefValue)
			}
		})
	}
}

func TestRunSelectProject_EmptyProjectsFile(t *testing.T) {
	// Create an empty projects file
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "projects")
	os.WriteFile(emptyFile, []byte(""), 0644)

	// Reset flag and execute
	rootCmd.SetArgs([]string{"select-project", "--projects-file", emptyFile})

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := rootCmd.Execute()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("Expected no error for empty projects, got: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, `"selected":false`) {
		t.Errorf("Expected {\"selected\":false} for empty projects, got: %s", output)
	}
}

func TestRunSelectProject_MissingFile(t *testing.T) {
	rootCmd.SetArgs([]string{"select-project", "--projects-file", "/nonexistent/path/projects"})
	err := rootCmd.Execute()

	if err == nil {
		t.Error("Expected error for missing projects file")
	}
}
