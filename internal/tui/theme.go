package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// AIToolTheme defines the color palette for an AI tool's TUI appearance.
type AIToolTheme struct {
	Name          string
	Primary       lipgloss.Color
	Dim           lipgloss.Color
	Bright        lipgloss.Color
	Accent        lipgloss.Color
	Cap           lipgloss.Color
	DarkFeet      lipgloss.Color
	EyeWhite      lipgloss.Color
	EyePupil      lipgloss.Color
	// UIAccent is the chrome accent for popup furniture (the diff pager's border,
	// rule, active tab, icons, title). Kept separate from the ghost-shading colors
	// so the window chrome can be tuned without touching the mascot.
	UIAccent lipgloss.Color
	SleepPrimary  lipgloss.Color
	SleepAccent   lipgloss.Color
	SleepBlush    lipgloss.Color
	SleepDim      lipgloss.Color
	SleepDarkFeet lipgloss.Color
	SleepCap      lipgloss.Color
	Text          lipgloss.Color
}

var themes = map[string]AIToolTheme{
	"claude": {
		Name:          "claude",
		Primary:       lipgloss.Color("209"),
		Dim:           lipgloss.Color("166"),
		Bright:        lipgloss.Color("208"),
		Accent:        lipgloss.Color("220"),
		Cap:           lipgloss.Color("223"),
		DarkFeet:      lipgloss.Color("166"),
		EyeWhite:      lipgloss.Color("255"),
		EyePupil:      lipgloss.Color("232"),
		UIAccent:      lipgloss.Color("208"), // orange — the popup chrome color
		SleepPrimary:  lipgloss.Color("166"),
		SleepAccent:   lipgloss.Color("178"),
		SleepBlush:    lipgloss.Color("168"),
		SleepDim:      lipgloss.Color("130"),
		SleepDarkFeet: lipgloss.Color("94"),
		SleepCap:      lipgloss.Color("180"),
		Text:          lipgloss.Color("223"),
	},
	"opencode": {
		Name:          "opencode",
		Primary:       lipgloss.Color("141"), // #af87ff brand purple — gauge fill, title, eye band
		Dim:           lipgloss.Color("99"),  // #875fff — stats border, ghost mid-body
		Bright:        lipgloss.Color("147"), // #afafff — ghost upper body
		Accent:        lipgloss.Color("61"),  // #5f5faf — ghost lower band
		Cap:           lipgloss.Color("183"), // #dfafff — pale crown rim
		DarkFeet:      lipgloss.Color("60"),  // #5f5f87 — feet + smile
		EyeWhite:      lipgloss.Color("147"),
		EyePupil:      lipgloss.Color("235"), // near-black pupils
		UIAccent:      lipgloss.Color("141"), // purple — the popup chrome color
		SleepPrimary:  lipgloss.Color("103"), // #8787af — dim body
		SleepAccent:   lipgloss.Color("61"),  // #5f5faf — dim lower band
		SleepBlush:    lipgloss.Color("139"), // #af87af — mauve cheeks
		SleepDim:      lipgloss.Color("60"),  // #5f5f87
		SleepDarkFeet: lipgloss.Color("236"), // dim feet
		SleepCap:      lipgloss.Color("146"), // #afafd7 — dim rim
		Text:          lipgloss.Color("189"), // #d7d7ff
	},
}

// presetThemes maps a user-selectable theme name (chosen in the Settings menu)
// to its full palette. "orange" and "purple" reuse the per-tool palettes above;
// the rest are dedicated hues. Each ramp is intentionally fairly uniform in body
// tone (light crown → mid body → dark feet) so it reads well on BOTH ghost
// shapes — the claude and opencode mascots assign the Bright/Accent fields with
// opposite light/dark intent, so a high-contrast ramp would break on one of them.
var presetThemes = map[string]AIToolTheme{
	"orange": themes["claude"],
	"purple": themes["opencode"],
	"green": {
		Name:          "green",
		Primary:       lipgloss.Color("78"),  // #5fd787 — readable green: title, gauge, eye band
		Dim:           lipgloss.Color("35"),  // #00af5f — stats border, mid body
		Bright:        lipgloss.Color("114"), // #87d787 — upper body
		Accent:        lipgloss.Color("29"),  // #00875f — lower band
		Cap:           lipgloss.Color("157"), // #afffaf — pale crown
		DarkFeet:      lipgloss.Color("22"),  // #005f00 — feet + smile
		EyeWhite:      lipgloss.Color("194"), // #d7ffd7 — light eyes (claude ghost)
		EyePupil:      lipgloss.Color("235"),
		UIAccent:      lipgloss.Color("78"),
		SleepPrimary:  lipgloss.Color("65"),  // #5f875f — dim body
		SleepAccent:   lipgloss.Color("29"),
		SleepBlush:    lipgloss.Color("174"), // #d78787 — rosy cheeks
		SleepDim:      lipgloss.Color("22"),
		SleepDarkFeet: lipgloss.Color("236"),
		SleepCap:      lipgloss.Color("151"), // #afd7af — dim crown
		Text:          lipgloss.Color("194"),
	},
	"blue": {
		Name:          "blue",
		Primary:       lipgloss.Color("75"),  // #5fafff — sky blue
		Dim:           lipgloss.Color("32"),  // #0087d7
		Bright:        lipgloss.Color("117"), // #87d7ff — upper body
		Accent:        lipgloss.Color("25"),  // #005faf — lower band
		Cap:           lipgloss.Color("153"), // #afd7ff — pale crown
		DarkFeet:      lipgloss.Color("24"),  // #005f87 — feet
		EyeWhite:      lipgloss.Color("195"), // #d7ffff
		EyePupil:      lipgloss.Color("235"),
		UIAccent:      lipgloss.Color("75"),
		SleepPrimary:  lipgloss.Color("67"),  // #5f87af — dim body
		SleepAccent:   lipgloss.Color("25"),
		SleepBlush:    lipgloss.Color("174"),
		SleepDim:      lipgloss.Color("24"),
		SleepDarkFeet: lipgloss.Color("236"),
		SleepCap:      lipgloss.Color("153"),
		Text:          lipgloss.Color("195"),
	},
	"rose": {
		Name:          "rose",
		Primary:       lipgloss.Color("211"), // #ff87af — rose pink
		Dim:           lipgloss.Color("168"), // #d75f87 — stats border, mid body
		Bright:        lipgloss.Color("218"), // #ffafd7 — upper body
		Accent:        lipgloss.Color("125"), // #af005f — lower band
		Cap:           lipgloss.Color("225"), // #ffd7ff — pale crown
		DarkFeet:      lipgloss.Color("89"),  // #87005f — feet
		EyeWhite:      lipgloss.Color("224"), // #ffd7d7
		EyePupil:      lipgloss.Color("235"),
		UIAccent:      lipgloss.Color("211"),
		SleepPrimary:  lipgloss.Color("132"), // #af5f87 — dim body
		SleepAccent:   lipgloss.Color("96"),  // #875f87 — muted dim band (sleeping)
		SleepBlush:    lipgloss.Color("174"),
		SleepDim:      lipgloss.Color("95"),  // #875f5f
		SleepDarkFeet: lipgloss.Color("236"),
		SleepCap:      lipgloss.Color("182"), // #d7afd7 — dim crown
		Text:          lipgloss.Color("225"),
	},
	"cyan": {
		Name:          "cyan",
		Primary:       lipgloss.Color("80"),  // #5fd7d7 — cyan
		Dim:           lipgloss.Color("37"),  // #00afaf
		Bright:        lipgloss.Color("123"), // #87ffff — upper body
		Accent:        lipgloss.Color("30"),  // #008787 — lower band
		Cap:           lipgloss.Color("159"), // #afffff — pale crown
		DarkFeet:      lipgloss.Color("23"),  // #005f5f — feet
		EyeWhite:      lipgloss.Color("195"), // #d7ffff
		EyePupil:      lipgloss.Color("235"),
		UIAccent:      lipgloss.Color("80"),
		SleepPrimary:  lipgloss.Color("66"),  // #5f8787 — dim body
		SleepAccent:   lipgloss.Color("30"),
		SleepBlush:    lipgloss.Color("174"),
		SleepDim:      lipgloss.Color("23"),
		SleepDarkFeet: lipgloss.Color("236"),
		SleepCap:      lipgloss.Color("152"), // #afd7d7 — dim crown
		Text:          lipgloss.Color("195"),
	},
}

// ThemePresets is the ordered list of theme choices shown in the Settings menu.
// "auto" follows the active AI tool (claude → orange, opencode → purple); every
// other entry forces that fixed palette regardless of the tool.
var ThemePresets = []string{"auto", "orange", "purple", "green", "blue", "rose", "cyan"}

// ResolveTheme returns the palette to use for the given AI tool and user theme
// preference. A named preset wins; "auto" (or empty/unknown) follows the tool.
func ResolveTheme(tool, pref string) AIToolTheme {
	if theme, ok := presetThemes[pref]; ok {
		return theme
	}
	return ThemeForTool(tool)
}

// currentTheme is the palette last applied via ApplyTheme. Components that need
// the full palette (not just the package-level title/selected styles) read this.
// Defaults to claude so rendering works even before ApplyTheme is called (tests).
var currentTheme = themes["claude"]

// ThemeForTool returns the color theme for the given AI tool.
// Unknown tools fall back to the claude theme.
func ThemeForTool(tool string) AIToolTheme {
	if theme, ok := themes[tool]; ok {
		return theme
	}
	return themes["claude"]
}

// AnsiFromThemeColor converts a lipgloss.Color (ANSI 256 string) to an
// ANSI escape sequence. This bridges lipgloss theme colors with raw
// escape-code rendering used by ghost ASCII art.
func AnsiFromThemeColor(c lipgloss.Color) string {
	return fmt.Sprintf("\033[38;5;%sm", string(c))
}

// ApplyTheme updates the package-level styles (titleStyle, selectedItemStyle,
// questionStyle) to use the given theme's Primary color. Call this before
// creating any TUI model so that all components reflect the AI tool's colors.
func ApplyTheme(theme AIToolTheme) {
	currentTheme = theme
	titleStyle = lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	selectedItemStyle = lipgloss.NewStyle().Foreground(theme.Primary)
	questionStyle = lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	applyDiffChrome(theme.UIAccent)
}
