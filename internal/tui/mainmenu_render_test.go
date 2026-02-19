package tui

import (
	"os"
	"strings"
	"testing"

	"github.com/jackuait/ghost-tab/internal/models"
)

func newTestMenu() *MainMenuModel {
	projects := []models.Project{
		{Name: "test-proj", Path: "/tmp/test-proj"},
	}
	m := NewMainMenu(projects, []string{"claude", "codex"}, "claude", "animated")
	m.width = 100
	m.height = 40
	return m
}

func TestMenuBox_AIToolRightAligned(t *testing.T) {
	m := newTestMenu()
	box := m.renderMenuBox()
	lines := strings.Split(box, "\n")

	// Title row is the second line (index 1), after top border
	if len(lines) < 2 {
		t.Fatal("renderMenuBox produced fewer than 2 lines")
	}
	titleRow := lines[1]

	// The AI tool display name should appear after Ghost Tab, not immediately adjacent
	// With right-alignment, there should be spaces between "Ghost Tab" and the AI tool
	if !strings.Contains(titleRow, "Ghost Tab") {
		t.Error("title row missing 'Ghost Tab'")
	}
	if !strings.Contains(titleRow, "Claude Code") {
		t.Error("title row missing 'Claude Code'")
	}

	// Verify right-alignment: there should be multiple spaces between Ghost Tab and the ◂ arrow
	// Strip ANSI codes to check raw layout
	raw := stripAnsi(titleRow)
	ghostIdx := strings.Index(raw, "Ghost Tab")
	arrowIdx := strings.Index(raw, "◂")
	if ghostIdx < 0 || arrowIdx < 0 {
		t.Fatal("could not find Ghost Tab or ◂ in stripped title row")
	}
	// With right-alignment, there should be significant padding between the end of
	// "Ghost Tab" and "◂" (more than just a single space)
	gap := raw[ghostIdx+len("Ghost Tab") : arrowIdx]
	if len(strings.TrimSpace(gap)) != 0 {
		t.Errorf("expected only whitespace between Ghost Tab and ◂, got %q", gap)
	}
	if len(gap) < 5 {
		t.Errorf("expected at least 5 chars padding for right-alignment, got %d: %q", len(gap), gap)
	}
}

func TestMenuBox_HelpTextPresent(t *testing.T) {
	m := newTestMenu()
	box := m.renderMenuBox()

	raw := stripAnsi(box)
	// Help text should contain navigation hints
	if !strings.Contains(raw, "navigate") {
		t.Error("help text missing 'navigate'")
	}
	if !strings.Contains(raw, "AI tool") {
		t.Error("help text missing 'AI tool' (expected when multiple AI tools available)")
	}
	if !strings.Contains(raw, "select") {
		t.Error("help text missing 'select'")
	}
}

func TestSettingsBox_StateRightAligned(t *testing.T) {
	m := newTestMenu()
	m.settingsMode = true
	m.tabTitle = "full"
	box := m.renderSettingsBox()
	raw := stripAnsi(box)
	lines := strings.Split(raw, "\n")

	// Find lines containing "Ghost Display" and "Tab Title"
	for _, line := range lines {
		if strings.Contains(line, "Ghost Display") && strings.Contains(line, "[Animated]") {
			// State text should be right-aligned: ends near the right border
			// The line should end with the state text followed by the border character
			trimmed := strings.TrimRight(line, " ")
			idx := strings.Index(trimmed, "[Animated]")
			if idx < 0 {
				t.Fatal("could not find [Animated] in Ghost Display line")
			}
			afterState := trimmed[idx+len("[Animated]"):]
			// After state text, only a small gap + border char should remain
			cleaned := strings.TrimSpace(afterState)
			if cleaned != "│" {
				t.Errorf("expected only border after [Animated], got %q", afterState)
			}
			// Between label and state there should be significant padding
			labelEnd := strings.Index(line, "Ghost Display") + len("Ghost Display")
			gap := line[labelEnd:idx]
			if len(strings.TrimSpace(gap)) != 0 {
				t.Errorf("expected only whitespace between label and state, got %q", gap)
			}
			if len(gap) < 5 {
				t.Errorf("expected at least 5 chars gap for right-alignment, got %d", len(gap))
			}
		}
	}
}

func TestGhostDisplayLabel_AllModes(t *testing.T) {
	tests := []struct {
		mode     string
		expected string
	}{
		{"animated", "Animated"},
		{"static", "Static"},
		{"none", "None"},
		{"custom", "custom"},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			result := ghostDisplayLabel(tt.mode)
			if result != tt.expected {
				t.Errorf("ghostDisplayLabel(%q) = %q, want %q", tt.mode, result, tt.expected)
			}
		})
	}
}

func TestTabTitleLabel_AllModes(t *testing.T) {
	tests := []struct {
		mode     string
		expected string
	}{
		{"full", "Project \u00b7 Tool"},
		{"project", "Project Only"},
		{"other", "other"},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			result := tabTitleLabel(tt.mode)
			if result != tt.expected {
				t.Errorf("tabTitleLabel(%q) = %q, want %q", tt.mode, result, tt.expected)
			}
		})
	}
}

func TestShortenHomePath(t *testing.T) {
	home := os.Getenv("HOME")
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"home prefix", home + "/projects/foo", "~/projects/foo"},
		{"no home prefix", "/usr/local/bin", "/usr/local/bin"},
		{"exact home", home, "~"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shortenHomePath(tt.input)
			if result != tt.expected {
				t.Errorf("shortenHomePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSettingsBox_SoundDisabled(t *testing.T) {
	m := newTestMenu()
	m.SetSoundName("")
	m.EnterSettings()
	box := m.renderSettingsBox()
	if !strings.Contains(box, "Sound") {
		t.Error("settings box missing 'Sound' label")
	}
	if !strings.Contains(box, "Off") {
		t.Error("settings box should show 'Off' when sound disabled")
	}
}

func TestSettingsBox_SoundName(t *testing.T) {
	m := newTestMenu()
	m.SetSoundName("Glass")
	m.EnterSettings()
	box := m.renderSettingsBox()
	if !strings.Contains(box, "Sound") {
		t.Error("settings box missing 'Sound' label")
	}
	if !strings.Contains(box, "Glass") {
		t.Error("settings box should show 'Glass' when sound set to Glass")
	}
}

func TestCycleSoundName(t *testing.T) {
	m := newTestMenu()
	m.SetSoundName("")
	m.CycleSoundName()
	if m.SoundName() != "Basso" {
		t.Errorf("expected 'Basso' after cycling from Off, got %q", m.SoundName())
	}
}

func TestCycleSoundNameReverse(t *testing.T) {
	m := newTestMenu()
	m.SetSoundName("")
	m.CycleSoundNameReverse()
	if m.SoundName() != "Tink" {
		t.Errorf("expected 'Tink' after reverse cycling from Off, got %q", m.SoundName())
	}
}

func TestSoundNameForResult_UnchangedReturnsNil(t *testing.T) {
	m := newTestMenu()
	m.SetSoundName("Bottle")
	result := m.soundNameForResult()
	if result != nil {
		t.Error("expected nil when sound not changed")
	}
}

func TestSoundNameForResult_ChangedReturnsValue(t *testing.T) {
	m := newTestMenu()
	m.SetSoundName("Bottle")
	m.CycleSoundName()
	result := m.soundNameForResult()
	if result == nil {
		t.Fatal("expected non-nil when sound changed")
	}
	if *result != "Frog" {
		t.Errorf("expected 'Frog' after cycling from Bottle, got %q", *result)
	}
}

func TestMenuBox_WorktreeCountIndicator(t *testing.T) {
	projects := []models.Project{
		{
			Name: "ghost-tab",
			Path: "/tmp/ghost-tab",
			Worktrees: []models.Worktree{
				{Path: "/tmp/wt1", Branch: "feature/auth"},
				{Path: "/tmp/wt2", Branch: "fix/bug"},
			},
		},
	}
	m := NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	m.width = 100
	m.height = 40
	box := m.renderMenuBox()
	raw := stripAnsi(box)

	if !strings.Contains(raw, "2 worktrees") {
		t.Errorf("expected '2 worktrees' indicator in menu, got:\n%s", raw)
	}
}

func TestMenuBox_ExpandedWorktreeEntries(t *testing.T) {
	projects := []models.Project{
		{
			Name: "ghost-tab",
			Path: "/tmp/ghost-tab",
			Worktrees: []models.Worktree{
				{Path: "/tmp/wt1", Branch: "feature/auth"},
				{Path: "/tmp/wt2", Branch: "fix/bug"},
			},
		},
	}
	m := NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	m.width = 100
	m.height = 40
	m.expandedWorktrees = map[int]bool{0: true}
	box := m.renderMenuBox()
	raw := stripAnsi(box)

	if !strings.Contains(raw, "feature/auth") {
		t.Errorf("expected 'feature/auth' in expanded menu, got:\n%s", raw)
	}
	if !strings.Contains(raw, "fix/bug") {
		t.Errorf("expected 'fix/bug' in expanded menu, got:\n%s", raw)
	}
}

func TestMenuBox_WorktreeTreeConnectors(t *testing.T) {
	projects := []models.Project{
		{
			Name: "proj",
			Path: "/tmp/proj",
			Worktrees: []models.Worktree{
				{Path: "/tmp/wt1", Branch: "feature/auth"},
				{Path: "/tmp/wt2", Branch: "fix/bug"},
			},
		},
	}
	m := NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	m.width = 100
	m.height = 40
	m.expandedWorktrees = map[int]bool{0: true}
	box := m.renderMenuBox()
	raw := stripAnsi(box)

	// All worktrees use ├─ connector (add-worktree follows)
	if !strings.Contains(raw, "├─ feature/auth") {
		t.Errorf("expected '├─ feature/auth' for worktree, got:\n%s", raw)
	}
	if !strings.Contains(raw, "├─ fix/bug") {
		t.Errorf("expected '├─ fix/bug' for worktree, got:\n%s", raw)
	}
	// Add-worktree item uses └─ connector as last item
	if !strings.Contains(raw, "└─ + Add worktree") {
		t.Errorf("expected '└─ + Add worktree' as last item, got:\n%s", raw)
	}
}

func TestMenuBox_SingleWorktreeUsesEndConnector(t *testing.T) {
	projects := []models.Project{
		{
			Name: "proj",
			Path: "/tmp/proj",
			Worktrees: []models.Worktree{
				{Path: "/tmp/wt1", Branch: "only-branch"},
			},
		},
	}
	m := NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	m.width = 100
	m.height = 40
	m.expandedWorktrees = map[int]bool{0: true}
	box := m.renderMenuBox()
	raw := stripAnsi(box)

	// Single worktree uses ├─ (add-worktree follows as └─)
	if !strings.Contains(raw, "├─ only-branch") {
		t.Errorf("expected '├─ only-branch' for single worktree, got:\n%s", raw)
	}
	// Add-worktree item uses └─ connector as last item
	if !strings.Contains(raw, "└─ + Add worktree") {
		t.Errorf("expected '└─ + Add worktree' as last item, got:\n%s", raw)
	}
}

func TestMenuBox_WorktreeShowsPath(t *testing.T) {
	projects := []models.Project{
		{
			Name: "proj",
			Path: "/tmp/proj",
			Worktrees: []models.Worktree{
				{Path: "/home/jack/wt/feature-auth", Branch: "feature/auth"},
			},
		},
	}
	m := NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	m.width = 100
	m.height = 40
	m.expandedWorktrees = map[int]bool{0: true}
	box := m.renderMenuBox()
	raw := stripAnsi(box)

	// Worktree entry should show the shortened path on a second line
	if !strings.Contains(raw, "wt/feature-auth") {
		t.Errorf("expected worktree path in expanded menu, got:\n%s", raw)
	}
}

func TestMenuBox_NoIndicatorWithoutWorktrees(t *testing.T) {
	projects := []models.Project{
		{Name: "simple", Path: "/tmp/simple"},
	}
	m := NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	m.width = 100
	m.height = 40
	box := m.renderMenuBox()
	raw := stripAnsi(box)

	if strings.Contains(raw, "worktree") {
		t.Errorf("expected no worktree indicator for project without worktrees, got:\n%s", raw)
	}
}

func TestMenuBox_HelpTextIncludesWorktreeKey(t *testing.T) {
	projects := []models.Project{
		{
			Name: "proj",
			Path: "/tmp/proj",
			Worktrees: []models.Worktree{
				{Path: "/tmp/wt", Branch: "feature"},
			},
		},
	}
	m := NewMainMenu(projects, []string{"claude", "codex"}, "claude", "animated")
	m.width = 100
	m.height = 40
	box := m.renderMenuBox()
	raw := stripAnsi(box)

	if !strings.Contains(raw, "w worktrees") && !strings.Contains(raw, "W worktrees") {
		t.Errorf("expected help text to mention 'w worktrees', got:\n%s", raw)
	}
}

// stripAnsi removes ANSI escape sequences from a string.
func stripAnsi(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			// Skip until 'm' (end of ANSI sequence)
			for i < len(s) && s[i] != 'm' {
				i++
			}
			if i < len(s) {
				i++ // skip the 'm'
			}
			continue
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}
