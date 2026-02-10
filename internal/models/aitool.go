package models

import (
	"os/exec"
)

// AITool represents an AI coding assistant
type AITool struct {
	Name      string
	Command   string
	Installed bool
}

// String returns display string for AI tool
func (t AITool) String() string {
	displayName := t.Name
	switch t.Name {
	case "claude":
		displayName = "Claude Code"
	case "codex":
		displayName = "Codex CLI"
	case "copilot":
		displayName = "Copilot CLI"
	case "opencode":
		displayName = "OpenCode"
	}

	if t.Installed {
		return displayName + " ✓"
	}
	return displayName + " (not installed)"
}

// DetectAITools checks which AI tools are installed
func DetectAITools() []AITool {
	tools := []AITool{
		{Name: "claude", Command: "claude"},
		{Name: "codex", Command: "codex"},
		{Name: "copilot", Command: "gh copilot"},
		{Name: "opencode", Command: "opencode"},
	}

	for i := range tools {
		tools[i].Installed = isCommandAvailable(tools[i].Command)
	}

	return tools
}

func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
