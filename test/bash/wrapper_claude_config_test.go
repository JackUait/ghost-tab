package bash_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrapperResolvesActiveConfigForClaude(t *testing.T) {
	root := projectRoot(t)
	libPath := filepath.Join(root, "lib/claude-configs.sh")
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "claude-configs")
	_ = os.MkdirAll(cfgDir, 0o755)
	writeTempFile(t, cfgDir, "work.json", "{}")
	writeTempFile(t, dir, "claude-config", "work.json")

	script := `
source ` + libPath + `
SELECTED_AI_TOOL=claude
GT_CONFIG_DIR="` + dir + `"
GHOST_TAB_CLAUDE_SETTINGS=""
if [ "$SELECTED_AI_TOOL" = "claude" ]; then
  GHOST_TAB_CLAUDE_SETTINGS="$(resolve_claude_config_path "$GT_CONFIG_DIR/claude-configs" "$GT_CONFIG_DIR/claude-config")"
fi
echo "RESULT=$GHOST_TAB_CLAUDE_SETTINGS"
`
	out, code := runBashSnippet(t, script, nil)
	assertExitCode(t, code, 0)
	if !strings.Contains(out, "RESULT="+filepath.Join(cfgDir, "work.json")) {
		t.Fatalf("got %q", out)
	}
}

func TestWrapperNoConfigWhenStandard(t *testing.T) {
	root := projectRoot(t)
	libPath := filepath.Join(root, "lib/claude-configs.sh")
	dir := t.TempDir()
	script := `
source ` + libPath + `
SELECTED_AI_TOOL=claude
GT_CONFIG_DIR="` + dir + `"
GHOST_TAB_CLAUDE_SETTINGS="$(resolve_claude_config_path "$GT_CONFIG_DIR/claude-configs" "$GT_CONFIG_DIR/claude-config")"
echo "RESULT=[$GHOST_TAB_CLAUDE_SETTINGS]"
`
	out, code := runBashSnippet(t, script, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "RESULT=[]")
}
