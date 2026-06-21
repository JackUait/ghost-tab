package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncOpenCode_writes_config_from_model_paths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	cfgRoot := filepath.Join(home, ".config", "ghost-tab")
	configsDir := filepath.Join(cfgRoot, "claude-configs")
	os.MkdirAll(configsDir, 0755)
	os.WriteFile(filepath.Join(cfgRoot, "claude-configs.list"),
		[]byte("Work GLM zhipu:work.json\n"), 0644)
	os.WriteFile(filepath.Join(configsDir, "work.json"),
		[]byte(`{"env":{"ANTHROPIC_AUTH_TOKEN":"sk-abc","ANTHROPIC_DEFAULT_OPUS_MODEL":"glm-4.6"}}`), 0644)
	os.WriteFile(filepath.Join(cfgRoot, "claude-config"), []byte("work.json\n"), 0644)

	m := &MainMenuModel{
		claudeConfigsList: filepath.Join(cfgRoot, "claude-configs.list"),
		claudeConfigsDir:  configsDir,
		claudeConfigFile:  filepath.Join(cfgRoot, "claude-config"),
	}
	t.Setenv("HOME", home)
	m.syncOpenCode()

	if _, err := os.Stat(filepath.Join(home, ".config", "opencode", "opencode.json")); err != nil {
		t.Errorf("syncOpenCode did not write opencode.json: %v", err)
	}
}
