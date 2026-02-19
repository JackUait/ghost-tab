package models

import "os"

// Terminal represents a supported terminal emulator
type Terminal struct {
	Name        string
	DisplayName string
	CaskName    string
	AppName     string // For /Applications check (macOS)
	Installed   bool
}

// String returns display string for terminal
func (t Terminal) String() string {
	if t.Installed {
		return t.DisplayName + " ✓"
	}
	return t.DisplayName + " (not installed)"
}

// SupportedTerminals returns the list of terminals Ghost Tab can configure
func SupportedTerminals() []Terminal {
	return []Terminal{
		{Name: "ghostty", DisplayName: "Ghostty", CaskName: "ghostty", AppName: "Ghostty"},
		{Name: "iterm2", DisplayName: "iTerm2", CaskName: "iterm2", AppName: "iTerm"},
		{Name: "wezterm", DisplayName: "WezTerm", CaskName: "wezterm", AppName: "WezTerm"},
		{Name: "kitty", DisplayName: "kitty", CaskName: "kitty", AppName: "kitty"},
	}
}

// DetectTerminals checks which terminals are installed
func DetectTerminals() []Terminal {
	terminals := SupportedTerminals()
	for i := range terminals {
		terminals[i].Installed = isAppInstalled(terminals[i].AppName)
	}
	return terminals
}

func isAppInstalled(appName string) bool {
	_, err := os.Stat("/Applications/" + appName + ".app")
	return err == nil
}
