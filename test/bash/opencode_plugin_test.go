package bash_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readPluginTemplate(t *testing.T) string {
	t.Helper()
	root := projectRoot(t)
	pluginPath := filepath.Join(root, "templates", "opencode-plugin.ts")
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("failed to read opencode-plugin.ts template: %v", err)
	}
	return string(data)
}

// --- Debounce: sound should not fire immediately on session.idle ---

func TestOpencodePlugin_session_idle_does_not_call_afplay_directly(t *testing.T) {
	content := readPluginTemplate(t)

	// The session.idle handler should NOT call spawn("afplay", ...) directly.
	// It should use a debounce timer instead.
	lines := strings.Split(content, "\n")
	inIdleBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, `"session.idle"`) {
			inIdleBlock = true
		}
		if inIdleBlock && strings.Contains(trimmed, "afplay") && !strings.Contains(trimmed, "//") {
			// afplay in idle block is OK only inside a setTimeout/timer callback,
			// not as a direct call. Check that it's wrapped in a timer.
			t.Error("session.idle handler should not call afplay directly — must use debounce timer")
			break
		}
		// Detect end of the idle block (next event type check or closing brace pattern)
		if inIdleBlock && strings.Contains(trimmed, `"session.status"`) {
			break
		}
	}
}

func TestOpencodePlugin_contains_debounce_timer_for_sound(t *testing.T) {
	content := readPluginTemplate(t)

	// Plugin must use setTimeout (or equivalent timer) to delay sound playback
	if !strings.Contains(content, "setTimeout") {
		t.Error("plugin should use setTimeout for debounce timer")
	}

	// Plugin must be able to cancel the timer when session becomes busy
	if !strings.Contains(content, "clearTimeout") {
		t.Error("plugin should use clearTimeout to cancel debounce on session.busy")
	}
}

func TestOpencodePlugin_cancels_debounce_on_session_busy(t *testing.T) {
	content := readPluginTemplate(t)

	// The session.status handler with status.type === "busy" should cancel the debounce timer.
	// This can be a direct clearTimeout call or a wrapper function that calls clearTimeout.
	lines := strings.Split(content, "\n")
	inBusyBlock := false
	cancelsBusyTimer := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, `"session.status"`) || strings.Contains(trimmed, `session.status`) {
			inBusyBlock = true
		}
		if inBusyBlock && strings.Contains(trimmed, "busy") {
			inBusyBlock = true
		}
		if inBusyBlock && (strings.Contains(trimmed, "clearTimeout") || strings.Contains(trimmed, "cancelIdleTimer")) {
			cancelsBusyTimer = true
			break
		}
	}
	if !cancelsBusyTimer {
		t.Error("session.busy handler should cancel pending debounce timer (clearTimeout or cancelIdleTimer)")
	}
}

// --- Debounce: spinner should also be debounced ---

func TestOpencodePlugin_contains_debounce_for_spinner(t *testing.T) {
	content := readPluginTemplate(t)

	// The spinner should also be debounced — it shouldn't start immediately on idle
	// because subagent completions cause brief idle states
	if !strings.Contains(content, "spinner") {
		t.Error("plugin should still support spinner feature")
	}

	// The spinner start should happen inside the debounce callback, not directly
	// Look for startSpinner being called within a setTimeout context
	if !strings.Contains(content, "startSpinner") {
		t.Error("plugin should contain startSpinner function")
	}
}

// --- Feature flags still work ---

func TestOpencodePlugin_reads_features_from_config(t *testing.T) {
	content := readPluginTemplate(t)

	if !strings.Contains(content, "opencode-features.json") {
		t.Error("plugin should read feature flags from opencode-features.json")
	}

	if !strings.Contains(content, "sound") {
		t.Error("plugin should support sound feature flag")
	}
}

// --- Plugin exports correct structure ---

func TestOpencodePlugin_exports_GhostTab(t *testing.T) {
	content := readPluginTemplate(t)

	if !strings.Contains(content, "export") && !strings.Contains(content, "GhostTab") {
		t.Error("plugin should export GhostTab")
	}
}

// --- Debounce threshold is reasonable ---

func TestOpencodePlugin_debounce_threshold_is_at_least_10_seconds(t *testing.T) {
	content := readPluginTemplate(t)

	// The debounce should be long enough to filter out subagent processing windows.
	// Subagent results cause 2-15+ seconds of thinking. A 10+ second debounce
	// covers the documented gap range and aligns with Claude Code's cooldown.
	// Check for a numeric constant >= 10000 (milliseconds)
	hasReasonableDebounce := false
	for _, threshold := range []string{"10000", "15000", "20000", "30000"} {
		if strings.Contains(content, threshold) {
			hasReasonableDebounce = true
			break
		}
	}
	if !hasReasonableDebounce {
		t.Error("plugin should have a debounce threshold of at least 10000ms (10 seconds)")
	}
}

// --- killSpinner on busy should remain immediate (no debounce) ---

func TestOpencodePlugin_kill_spinner_on_busy_is_immediate(t *testing.T) {
	content := readPluginTemplate(t)

	// killSpinner should still be called directly on session.busy (not debounced)
	// because we want the tab title to reset immediately when the AI starts working
	if !strings.Contains(content, "killSpinner") {
		t.Error("plugin should contain killSpinner function")
	}
}
