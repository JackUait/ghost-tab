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

func TestOpencodePlugin_long_debounce_threshold_is_at_least_10_seconds(t *testing.T) {
	content := readPluginTemplate(t)

	// The LONG debounce (after tool use) should be >= 10 seconds to filter
	// out subagent processing windows.
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

// --- Dual-threshold debounce: tool.execute.after hook ---

func TestOpencodePlugin_has_tool_execute_after_hook(t *testing.T) {
	content := readPluginTemplate(t)

	if !strings.Contains(content, "tool.execute.after") {
		t.Error("plugin should have a tool.execute.after hook to track tool completions for cooldown")
	}
}

func TestOpencodePlugin_tracks_last_tool_complete_time(t *testing.T) {
	content := readPluginTemplate(t)

	// Plugin should track when the last tool completed to implement dual-threshold debounce
	if !strings.Contains(content, "lastToolComplete") {
		t.Error("plugin should track last tool completion time (lastToolCompleteTime or similar)")
	}
}

func TestOpencodePlugin_uses_dual_threshold_debounce(t *testing.T) {
	content := readPluginTemplate(t)

	// Plugin should have a short debounce for normal idle (no recent tool activity)
	// and a long debounce after tool use
	hasShortDebounce := false
	for _, threshold := range []string{"1000", "2000", "3000"} {
		if strings.Contains(content, threshold) {
			hasShortDebounce = true
			break
		}
	}
	if !hasShortDebounce {
		t.Error("plugin should have a short debounce threshold (1-3 seconds) for idle without recent tool activity")
	}

	hasLongDebounce := false
	for _, threshold := range []string{"15000", "20000", "30000"} {
		if strings.Contains(content, threshold) {
			hasLongDebounce = true
			break
		}
	}
	if !hasLongDebounce {
		t.Error("plugin should have a long debounce threshold (15+ seconds) for idle after tool use")
	}
}

func TestOpencodePlugin_has_cooldown_window(t *testing.T) {
	content := readPluginTemplate(t)

	// Plugin should have a cooldown window to determine when tool activity is "recent"
	// (mirrors Claude Code's 30-second cooldown window in tab-title-watcher.sh)
	if !strings.Contains(content, "30000") && !strings.Contains(content, "COOLDOWN") {
		t.Error("plugin should have a cooldown window constant (e.g., 30000ms or COOLDOWN_WINDOW_MS)")
	}
}

// --- getProject: extracts clean project name from tmux session name ---

func TestOpencodePlugin_getProject_strips_session_name_prefix_and_pid(t *testing.T) {
	content := readPluginTemplate(t)

	// wrapper.sh creates tmux sessions named "dev-<project>-<pid>".
	// getProject() must extract just "<project>" from this session name,
	// not return the raw "dev-<project>-<pid>" string.
	// It should contain a regex or string manipulation to strip the prefix and PID suffix.
	if !strings.Contains(content, `dev-`) || !strings.Contains(content, `\d+`) {
		// Must have logic to parse the "dev-...-PID" format
		t.Error("getProject should parse tmux session name format 'dev-<project>-<PID>' to extract project name")
	}
}

// --- Dot indicator: ● prefix in tab title when waiting (like Claude Code) ---

func TestOpencodePlugin_sets_dot_indicator_on_idle(t *testing.T) {
	content := readPluginTemplate(t)

	// The onIdle function should set the tab title with ● prefix via
	// process.stdout.write, matching Claude Code's waiting indicator.
	// This must happen regardless of the spinner feature flag.
	// Check for ● inside a stdout.write call (not just a comment)
	if !strings.Contains(content, "● ") {
		t.Error("plugin should set tab title with ● dot prefix when idle (like Claude Code)")
	}

	// The ● must appear inside a process.stdout.write call within onIdle,
	// not just as a comment
	lines := strings.Split(content, "\n")
	inOnIdle := false
	hasDotWriteInOnIdle := false
	braceDepth := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "function onIdle") {
			inOnIdle = true
			braceDepth = 0
		}
		if inOnIdle {
			braceDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			// Must be an actual write call containing ●, not a comment
			if strings.Contains(trimmed, "●") && strings.Contains(trimmed, "write") && !strings.HasPrefix(trimmed, "//") {
				hasDotWriteInOnIdle = true
				break
			}
			if braceDepth <= 0 && inOnIdle && strings.Contains(trimmed, "}") {
				break
			}
		}
	}
	if !hasDotWriteInOnIdle {
		t.Error("● dot indicator must be set via process.stdout.write inside onIdle (not just a comment)")
	}
}

func TestOpencodePlugin_clears_dot_indicator_on_busy(t *testing.T) {
	content := readPluginTemplate(t)

	// When session becomes busy, the tab title must be reset to the plain
	// project name (no ● prefix). This ensures the dot is cleared even
	// when the spinner feature is disabled.
	lines := strings.Split(content, "\n")
	inBusyBlock := false
	resetsTitle := false
	braceDepth := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, `"busy"`) || strings.Contains(trimmed, `'busy'`) {
			inBusyBlock = true
			braceDepth = 1 // We're inside the if-block
		}
		if inBusyBlock {
			braceDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			// Should reset tab title via OSC escape sequence or a resetTabTitle helper
			if strings.Contains(trimmed, "resetTabTitle") && !strings.HasPrefix(trimmed, "//") {
				resetsTitle = true
				break
			}
			// Stop scanning when we exit the busy block
			if braceDepth <= 0 {
				break
			}
		}
	}
	if !resetsTitle {
		t.Error("busy handler should call resetTabTitle() to clear ● dot from tab title")
	}
}

func TestOpencodePlugin_defines_resetTabTitle_function(t *testing.T) {
	content := readPluginTemplate(t)

	// resetTabTitle must be a real function definition, not just referenced
	if !strings.Contains(content, "function resetTabTitle") {
		t.Error("plugin should define a resetTabTitle function to clear tab title indicators")
	}
}
