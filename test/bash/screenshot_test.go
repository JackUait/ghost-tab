package bash_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// gt_latest_screenshot <dir> prints the newest image file in dir.
func TestLatestScreenshot_returns_newest_image(t *testing.T) {
	dir := t.TempDir()
	older := writeTempFile(t, dir, "Screen Shot old.png", "old")
	newer := writeTempFile(t, dir, "Screen Shot new.png", "new")
	// Force mtimes: older is older, newer is newer.
	now := time.Now()
	os.Chtimes(older, now.Add(-2*time.Minute), now.Add(-2*time.Minute))
	os.Chtimes(newer, now, now)

	out, code := runBashFunc(t, "lib/screenshot.sh", "gt_latest_screenshot",
		[]string{dir}, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Screen Shot new.png")
	assertNotContains(t, out, "Screen Shot old.png")
}

// Non-image files must be ignored.
func TestLatestScreenshot_ignores_non_images(t *testing.T) {
	dir := t.TempDir()
	img := writeTempFile(t, dir, "shot.png", "img")
	txt := writeTempFile(t, dir, "notes.txt", "text")
	now := time.Now()
	os.Chtimes(img, now.Add(-1*time.Minute), now.Add(-1*time.Minute))
	os.Chtimes(txt, now, now) // newer, but not an image

	out, code := runBashFunc(t, "lib/screenshot.sh", "gt_latest_screenshot",
		[]string{dir}, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "shot.png")
	assertNotContains(t, out, "notes.txt")
}

// Empty / no-image dir returns non-zero so the binding can no-op.
func TestLatestScreenshot_no_images_returns_error(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "notes.txt", "text")
	out, code := runBashFunc(t, "lib/screenshot.sh", "gt_latest_screenshot",
		[]string{dir}, nil)
	if code == 0 {
		t.Errorf("expected non-zero exit when no images present, got 0; out=%q", out)
	}
}

// gt_screenshot_dir uses the macOS screencapture location when set.
func TestScreenshotDir_uses_configured_location(t *testing.T) {
	dir := t.TempDir()
	shotDir := filepath.Join(dir, "Shots")
	if err := os.MkdirAll(shotDir, 0755); err != nil {
		t.Fatal(err)
	}
	bin := mockCommand(t, dir, "defaults", `echo "`+shotDir+`"`)
	env := buildEnv(t, []string{bin})
	out, code := runBashFunc(t, "lib/screenshot.sh", "gt_screenshot_dir", nil, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, shotDir)
}

// gt_screenshot_dir falls back to ~/Desktop when location is unset.
func TestScreenshotDir_defaults_to_desktop(t *testing.T) {
	home := t.TempDir()
	desktop := filepath.Join(home, "Desktop")
	if err := os.MkdirAll(desktop, 0755); err != nil {
		t.Fatal(err)
	}
	// defaults read prints nothing and exits non-zero (key absent).
	bin := mockCommand(t, home, "defaults", `exit 1`)
	env := buildEnv(t, []string{bin}, "HOME="+home)
	out, code := runBashFunc(t, "lib/screenshot.sh", "gt_screenshot_dir", nil, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, desktop)
}
