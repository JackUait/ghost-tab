package tui

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jackuait/ghost-tab/internal/models"
	"github.com/muesli/termenv"
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

func TestMenuBox_AIToolHasTrailingSpace(t *testing.T) {
	m := newTestMenu()
	box := m.renderMenuBox()
	lines := strings.Split(box, "\n")
	if len(lines) < 2 {
		t.Fatal("renderMenuBox produced fewer than 2 lines")
	}
	raw := stripAnsi(lines[1]) // title row
	// AI tool selector should have a trailing space before the right border │
	// e.g. "◂ Claude Code ▸ │" not "◂ Claude Code ▸│"
	if !strings.HasSuffix(raw, "▸ │") {
		t.Errorf("expected AI tool selector to have trailing space before border, got: %q", raw)
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

func TestMenuBox_HelpTextFitsWithinBorders(t *testing.T) {
	// With multiple AI tools + worktrees, the help text is at its longest:
	// "↑↓ navigate ←→ AI tool S settings w worktrees ⏎ select"
	// This must fit within the container borders (all lines same width).
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
	lines := strings.Split(raw, "\n")

	if len(lines) < 3 {
		t.Fatal("renderMenuBox produced fewer than 3 lines")
	}

	// The top border defines the expected width of every row
	borderWidth := len([]rune(lines[0]))
	for i, line := range lines {
		lineWidth := len([]rune(line))
		if lineWidth != borderWidth {
			t.Errorf("line %d width %d != border width %d: %q", i, lineWidth, borderWidth, line)
		}
	}
}

func TestMenuBox_UnselectedProjectUsesNeutralColors(t *testing.T) {
	// Force color output so lipgloss emits ANSI codes in tests.
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	projects := []models.Project{
		{Name: "selected-proj", Path: "/tmp/selected"},
		{Name: "unselected-proj", Path: "/tmp/unselected"},
	}
	m := NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	m.width = 100
	m.height = 40
	// Item 0 is selected by default, so item 1 (unselected-proj) should use neutral colors
	box := m.renderMenuBox()

	// The unselected project name should use neutral text color (252), not theme.Text (223)
	// ANSI 256-color format: \033[38;5;COLORm
	if !strings.Contains(box, "\x1b[38;5;252m") {
		t.Error("expected neutral text color (252) for unselected project name")
	}
	// The unselected project path should use neutral dim color (245), not theme.Dim (166)
	if !strings.Contains(box, "\x1b[38;5;245m") {
		t.Error("expected neutral dim color (245) for unselected project path/number")
	}
}

func TestMenuBox_SelectedProjectUsesThemeColor(t *testing.T) {
	// Force color output so lipgloss emits ANSI codes in tests.
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	projects := []models.Project{
		{Name: "selected-proj", Path: "/tmp/selected"},
	}
	m := NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	m.width = 100
	m.height = 40
	box := m.renderMenuBox()

	// Selected project should use theme.Primary (209) not neutral colors
	if !strings.Contains(box, "\x1b[38;5;209m") {
		t.Error("expected theme primary color (209) for selected project")
	}
}

func TestMenuBox_UnselectedActionUsesNeutralColors(t *testing.T) {
	// Force color output so lipgloss emits ANSI codes in tests.
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	projects := []models.Project{
		{Name: "proj", Path: "/tmp/proj"},
	}
	m := NewMainMenu(projects, []string{"claude"}, "claude", "animated")
	m.width = 100
	m.height = 40
	// Select first project (item 0), so action items are unselected
	box := m.renderMenuBox()

	// Find action item lines in raw output — they should contain neutral colors
	// Action shortcuts and labels should use 245 and 252 respectively
	lines := strings.Split(box, "\n")
	found245 := false
	found252 := false
	for _, line := range lines {
		raw := stripAnsi(line)
		if strings.Contains(raw, "Add new project") {
			if strings.Contains(line, "\x1b[38;5;245m") {
				found245 = true
			}
			if strings.Contains(line, "\x1b[38;5;252m") {
				found252 = true
			}
		}
	}
	if !found245 {
		t.Error("expected neutral dim color (245) for unselected action shortcut")
	}
	if !found252 {
		t.Error("expected neutral text color (252) for unselected action label")
	}
}

func TestMenuBox_BordersStillUseThemeDim(t *testing.T) {
	// Force color output so lipgloss emits ANSI codes in tests.
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	m := newTestMenu()
	box := m.renderMenuBox()
	lines := strings.Split(box, "\n")

	// Top border (first line) should use theme.Dim (166), not neutral gray
	if len(lines) < 1 {
		t.Fatal("no lines in rendered box")
	}
	if !strings.Contains(lines[0], "\x1b[38;5;166m") {
		t.Error("expected theme dim color (166) for box border")
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
