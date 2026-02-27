package bash_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func iterm2AdapterSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	installPath := filepath.Join(root, "lib", "install.sh")
	adapterPath := filepath.Join(root, "lib", "terminals", "iterm2.sh")
	return fmt.Sprintf("source %q && source %q && source %q && %s",
		tuiPath, installPath, adapterPath, body)
}

func TestIterm2Adapter_get_config_path_returns_dynamic_profile(t *testing.T) {
	snippet := iterm2AdapterSnippet(t, `terminal_get_config_path`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	got := strings.TrimSpace(out)
	home := os.Getenv("HOME")
	expected := home + "/Library/Application Support/iTerm2/DynamicProfiles/ghost-tab.json"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestIterm2Adapter_get_wrapper_path(t *testing.T) {
	snippet := iterm2AdapterSnippet(t, `terminal_get_wrapper_path`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	got := strings.TrimSpace(out)
	home := os.Getenv("HOME")
	expected := home + "/.config/ghost-tab/wrapper.sh"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestIterm2Adapter_install_calls_ensure_cask(t *testing.T) {
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "Applications", "iTerm.app")
	os.MkdirAll(appDir, 0755)

	snippet := iterm2AdapterSnippet(t, fmt.Sprintf(
		`APPLICATIONS_DIR=%q terminal_install`, filepath.Join(tmpDir, "Applications")))
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "iTerm found")
}

func TestIterm2Adapter_setup_config_creates_json_file(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "DynamicProfiles", "ghost-tab.json")
	wrapperPath := "/path/to/wrapper.sh"

	snippet := iterm2AdapterSnippet(t,
		fmt.Sprintf(`terminal_setup_config %q %q`, profilePath, wrapperPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		t.Fatal("expected dynamic profile JSON to be created")
	}
}

func TestIterm2Adapter_setup_config_json_has_correct_profile(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "DynamicProfiles", "ghost-tab.json")
	wrapperPath := "/test/wrapper.sh"

	snippet := iterm2AdapterSnippet(t,
		fmt.Sprintf(`terminal_setup_config %q %q`, profilePath, wrapperPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("failed to read profile JSON: %v", err)
	}

	var profile map[string]interface{}
	if err := json.Unmarshal(data, &profile); err != nil {
		t.Fatalf("invalid JSON: %v\ncontent: %s", err, string(data))
	}

	profiles, ok := profile["Profiles"].([]interface{})
	if !ok || len(profiles) == 0 {
		t.Fatal("expected Profiles array with at least one entry")
	}

	p := profiles[0].(map[string]interface{})
	if p["Name"] != "Ghost Tab" {
		t.Errorf("Name = %q, want %q", p["Name"], "Ghost Tab")
	}
	if p["Guid"] != "ghost-tab-profile" {
		t.Errorf("Guid = %q, want %q", p["Guid"], "ghost-tab-profile")
	}
	if p["Custom Command"] != "Yes" {
		t.Errorf("Custom Command = %q, want %q", p["Custom Command"], "Yes")
	}
	if p["Command"] != wrapperPath {
		t.Errorf("Command = %q, want %q", p["Command"], wrapperPath)
	}
}

func TestIterm2Adapter_setup_config_overwrites_existing(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "DynamicProfiles", "ghost-tab.json")
	os.MkdirAll(filepath.Dir(profilePath), 0755)
	os.WriteFile(profilePath, []byte(`{"Profiles":[{"Name":"Old"}]}`), 0644)

	wrapperPath := "/new/wrapper.sh"
	snippet := iterm2AdapterSnippet(t,
		fmt.Sprintf(`terminal_setup_config %q %q`, profilePath, wrapperPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("failed to read profile: %v", err)
	}

	var profile map[string]interface{}
	json.Unmarshal(data, &profile)
	profiles := profile["Profiles"].([]interface{})
	p := profiles[0].(map[string]interface{})
	if p["Command"] != wrapperPath {
		t.Errorf("Command = %q, want %q", p["Command"], wrapperPath)
	}
}

func TestIterm2Adapter_cleanup_config_removes_json(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "ghost-tab.json")
	os.WriteFile(profilePath, []byte(`{"Profiles":[]}`), 0644)

	snippet := iterm2AdapterSnippet(t,
		fmt.Sprintf(`terminal_cleanup_config %q`, profilePath))
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Removed Ghost Tab profile")

	if _, err := os.Stat(profilePath); !os.IsNotExist(err) {
		t.Error("expected dynamic profile JSON to be removed")
	}
}

func TestIterm2Adapter_cleanup_config_noop_if_missing(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "nonexistent.json")

	snippet := iterm2AdapterSnippet(t,
		fmt.Sprintf(`terminal_cleanup_config %q`, profilePath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
}
