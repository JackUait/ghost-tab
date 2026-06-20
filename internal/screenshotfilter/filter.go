// Package screenshotfilter rewrites a dragged macOS screenshot's path mid-stream
// so the literal drag-and-drop into the AI pane works.
//
// When a screenshot's floating thumbnail is dragged into the terminal, the drop
// delivers (as a bracketed paste) the path to the screencaptureui *temp* file in
// .../TemporaryItems/NSIRD_screencaptureui_*/. macOS deletes that temp file the
// moment the thumbnail finalizes, so by the time the AI tool reads the path the
// file is gone and nothing attaches. This filter sits in the AI tool's input
// stream, spots such a paste, copies the file to a stable location while it
// still exists, and rewrites the pasted path to the stable copy. Everything else
// passes through untouched, so it is transparent when no screenshot is dropped.
package screenshotfilter

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	pasteStart = "\x1b[200~"
	pasteEnd   = "\x1b[201~"
	// maxPaste guards against buffering an unterminated paste forever; past this
	// the buffer is flushed through as-is.
	maxPaste = 1 << 16
)

// Filter is a streaming byte filter for terminal input. Feed it bytes via
// Process; it forwards everything unchanged except a bracketed paste, whose
// inner content it passes through Rewrite.
type Filter struct {
	inPaste bool
	buf     []byte
	// Rewrite maps a bracketed paste's inner content to its replacement.
	Rewrite func([]byte) []byte
}

// New returns a Filter that rewrites ephemeral screenshot paths.
func New() *Filter { return &Filter{Rewrite: RewriteScreenshotPath} }

// Process consumes input bytes and returns the bytes to forward downstream.
// Bytes may be held back (when a marker is split across reads, or while a paste
// body is still arriving) and emitted on a later call.
func (f *Filter) Process(p []byte) []byte {
	f.buf = append(f.buf, p...)
	var out []byte
	for {
		if !f.inPaste {
			if i := bytes.Index(f.buf, []byte(pasteStart)); i >= 0 {
				out = append(out, f.buf[:i]...)
				f.buf = append([]byte(nil), f.buf[i+len(pasteStart):]...)
				f.inPaste = true
				continue
			}
			// No full start marker: emit everything except a possible trailing
			// partial of the marker (it may complete on the next read).
			keep := partialSuffix(f.buf, []byte(pasteStart))
			out = append(out, f.buf[:len(f.buf)-keep]...)
			f.buf = append([]byte(nil), f.buf[len(f.buf)-keep:]...)
			return out
		}
		if j := bytes.Index(f.buf, []byte(pasteEnd)); j >= 0 {
			out = append(out, []byte(pasteStart)...)
			out = append(out, f.Rewrite(f.buf[:j])...)
			out = append(out, []byte(pasteEnd)...)
			f.buf = append([]byte(nil), f.buf[j+len(pasteEnd):]...)
			f.inPaste = false
			continue
		}
		// End marker not seen yet; keep buffering unless it grows pathological.
		if len(f.buf) > maxPaste {
			out = append(out, []byte(pasteStart)...)
			out = append(out, f.buf...)
			f.buf = nil
			f.inPaste = false
		}
		return out
	}
}

// partialSuffix returns the length of the longest suffix of b that is a proper
// prefix of needle, so a marker split across reads isn't emitted prematurely.
func partialSuffix(b, needle []byte) int {
	max := len(needle) - 1
	if max > len(b) {
		max = len(b)
	}
	for n := max; n > 0; n-- {
		if bytes.Equal(b[len(b)-n:], needle[:n]) {
			return n
		}
	}
	return 0
}

// RewriteScreenshotPath rewrites a bracketed paste's content into a path Claude
// will attach as an image. Two drop shapes reach here:
//   - Finder/Desktop drags deliver a percent-encoded file:// URL. Claude attaches a
//     plain filesystem path but NEVER a file:// URL, so we decode the URL first.
//   - Floating-thumbnail drags deliver a plain (often backslash-escaped) path.
//
// In both cases an ephemeral screencaptureui temp file is copied to a stable
// location (macOS deletes it moments after the drop); a persistent file is handed
// over as its plain path. A dropped video (which Claude cannot attach at all) is
// turned into image frames via ffmpeg. Anything we can't resolve to a real local
// image or video — a normal non-media paste, or a temp path whose file is already
// gone — is returned unchanged, so the filter is never worse than passing the drop
// straight through.
func RewriteScreenshotPath(content []byte) []byte {
	path, ok := fileURLToPath(string(content))
	if !ok {
		path = unescape(string(content))
	}
	if payload := videoFramesPayload(path); payload != nil {
		return payload
	}
	return resolveLocalImage(path, content)
}

// resolveLocalImage returns the bytes to hand Claude for a dragged image at `path`,
// or `orig` unchanged when it can't resolve a real, on-disk image file.
func resolveLocalImage(path string, orig []byte) []byte {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || !isImagePath(path) {
		return orig
	}
	if isEphemeralScreenshot(path) {
		if stable, err := copyToStable(path); err == nil {
			return []byte(stable)
		}
		return orig
	}
	return []byte(path)
}

// fileURLToPath converts a dragged file:// URL (Finder/Desktop drags deliver these,
// percent-encoded) to a local filesystem path. Returns ok=false when content is not
// a file:// URL, so the caller falls back to plain-path handling.
func fileURLToPath(content string) (string, bool) {
	s := strings.TrimSpace(content)
	if !strings.HasPrefix(s, "file://") {
		return "", false
	}
	u, err := url.Parse(s)
	if err != nil || u.Path == "" {
		return "", false
	}
	return u.Path, true
}

// unescape undoes shell backslash-escaping (Ghostty escapes spaces as "\ ").
func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func isEphemeralScreenshot(path string) bool {
	return strings.Contains(path, "/TemporaryItems/") &&
		strings.Contains(path, "screencaptureui") && isImagePath(path)
}

func isImagePath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return true
	}
	return false
}

func isVideoPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mov", ".mp4", ".m4v", ".webm", ".mkv", ".avi":
		return true
	}
	return false
}

// videoFramesPayload returns a rewrite payload of extracted image frames when `path`
// is an existing video file (which Claude cannot attach directly), or nil when it is
// not a video or frames can't be produced — in which case the caller falls back to
// normal image handling / passthrough.
func videoFramesPayload(path string) []byte {
	if !isVideoPath(path) {
		return nil
	}
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return nil
	}
	frames, err := extractVideoFrames(path)
	if err != nil || len(frames) == 0 {
		return nil
	}
	return framesToPayload(frames)
}

// framesToPayload renders frame paths as back-to-back bracketed pastes. Filter.Process
// wraps a rewrite result in a single 200~/201~ pair, so joining frames with
// "201~ 200~" splits the wrapped output into one paste per frame — Claude attaches
// multiple images only as separate pastes (proven against a live TUI).
func framesToPayload(frames []string) []byte {
	return []byte(strings.Join(frames, pasteEnd+pasteStart))
}

// maxVideoFrames caps how many evenly-spaced frames are pulled from a dropped video
// (each becomes one attached image, so this bounds context cost). Overridable via
// GT_VIDEO_MAX_FRAMES.
const maxVideoFrames = 8

func videoFrameCap() int {
	if v := os.Getenv("GT_VIDEO_MAX_FRAMES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return maxVideoFrames
}

// extractVideoFrames pulls up to videoFrameCap() evenly-spaced frames from src into
// the stable dir via ffmpeg and returns their (space-free) paths in order. Frames are
// spread across the whole clip by deriving an fps of frames/duration from ffprobe.
func extractVideoFrames(src string) ([]string, error) {
	dir := StableDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	n := videoFrameCap()
	prefix := fmt.Sprintf("gt-frame-%d-", time.Now().UnixNano())
	pattern := filepath.Join(dir, prefix+"%03d.png")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffmpeg", "-nostdin", "-loglevel", "error", "-y",
		"-i", src, "-vf", "fps="+frameRate(src, n), "-frames:v", strconv.Itoa(n), pattern)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	frames, err := filepath.Glob(filepath.Join(dir, prefix+"*.png"))
	if err != nil || len(frames) == 0 {
		return nil, fmt.Errorf("no frames extracted from %s", src)
	}
	sort.Strings(frames)
	return frames, nil
}

// frameRate returns the ffmpeg fps that yields ~n frames across the clip's full
// duration (n/duration). Falls back to "1" when the duration can't be probed.
func frameRate(src string, n int) string {
	out, err := exec.Command("ffprobe", "-v", "error", "-show_entries",
		"format=duration", "-of", "default=nw=1:nk=1", src).Output()
	if err == nil {
		if d, perr := strconv.ParseFloat(strings.TrimSpace(string(out)), 64); perr == nil && d > 0 {
			return strconv.FormatFloat(float64(n)/d, 'f', 6, 64)
		}
	}
	return "1"
}

// StableDir is where ephemeral screenshots are copied. Matches the bash side's
// gt_stable_screenshot_dir; overridable via GT_SCREENSHOT_STASH_DIR.
func StableDir() string {
	if d := os.Getenv("GT_SCREENSHOT_STASH_DIR"); d != "" {
		return d
	}
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(base, "ghost-tab", "screenshots")
}

func copyToStable(src string) (string, error) {
	dir := StableDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	// Space-free name so the rewritten path needs no shell escaping downstream.
	dest := filepath.Join(dir, fmt.Sprintf("gt-shot-%d%s",
		time.Now().UnixNano(), strings.ToLower(filepath.Ext(src))))
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return "", err
	}
	return dest, nil
}
