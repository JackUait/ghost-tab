package screenshotfilter

import (
	"bytes"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A bracketed paste whose content is an ephemeral screencaptureui temp screenshot
// must be copied to the stable dir and the path rewritten to the stable copy.
func TestRewriteScreenshotPath_ephemeral_copies_and_rewrites(t *testing.T) {
	root := t.TempDir()
	tempDir := filepath.Join(root, "TemporaryItems", "NSIRD_screencaptureui_ABC")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(tempDir, "Screenshot 2026-06-19 at 2.12.29 PM.png")
	if err := os.WriteFile(src, []byte("PNGDATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	stash := filepath.Join(root, "stash")
	t.Setenv("GT_SCREENSHOT_STASH_DIR", stash)

	// Ghostty escapes spaces as "\ "; the filter must unescape to find the file.
	escaped := strings.ReplaceAll(src, " ", `\ `)
	out := string(RewriteScreenshotPath([]byte(escaped)))

	if !strings.HasPrefix(out, stash) {
		t.Fatalf("rewritten path %q should live under stash %q", out, stash)
	}
	if strings.Contains(out, " ") {
		t.Errorf("stable path should be space-free (no escaping needed): %q", out)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("stable copy unreadable: %v", err)
	}
	if string(data) != "PNGDATA" {
		t.Errorf("stable copy content = %q, want PNGDATA", data)
	}
	if !strings.HasSuffix(strings.ToLower(out), ".png") {
		t.Errorf("stable path must keep image extension: %q", out)
	}
}

// A normal (non-ephemeral) path must pass through unchanged — the filter only
// touches the doomed temp-file case.
func TestRewriteScreenshotPath_nonephemeral_passthrough(t *testing.T) {
	in := `/Users/x/Desktop/Screenshot.png`
	out := string(RewriteScreenshotPath([]byte(in)))
	if out != in {
		t.Errorf("non-ephemeral path changed: got %q want %q", out, in)
	}
}

// An ephemeral-looking path whose file is already gone must pass through
// unchanged (no worse than today).
func TestRewriteScreenshotPath_missing_file_passthrough(t *testing.T) {
	in := `/var/folders/x/T/TemporaryItems/NSIRD_screencaptureui_X/gone.png`
	out := string(RewriteScreenshotPath([]byte(in)))
	if out != in {
		t.Errorf("missing-file path changed: got %q want %q", out, in)
	}
}

// A Finder/Desktop drag in Ghostty delivers a bracketed paste of a percent-encoded
// file:// URL, NOT a plain path. Claude Code attaches a plain filesystem path but
// never a file:// URL (proven empirically), so the filter must decode the URL to the
// real local path. The name carries a U+202F narrow no-break space (how macOS names
// screenshots) plus regular spaces, so this also proves %E2%80%AF and %20 both decode.
func TestRewriteScreenshotPath_fileurl_existing_decodes_to_plain_path(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "Screenshot 2026-06-20 at 1.38.03 PM.png")
	if err := os.WriteFile(src, []byte("PNGDATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	urlStr := (&url.URL{Scheme: "file", Path: src}).String()
	if !strings.HasPrefix(urlStr, "file://") || !strings.Contains(urlStr, "%20") {
		t.Fatalf("test setup: expected a percent-encoded file:// URL, got %q", urlStr)
	}

	out := string(RewriteScreenshotPath([]byte(urlStr)))

	if strings.HasPrefix(out, "file://") {
		t.Fatalf("file:// URL must be decoded (Claude won't attach a URL): got %q", out)
	}
	if out != src {
		t.Errorf("decoded path = %q, want the real local path %q", out, src)
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("decoded path must resolve to the real file: %v", err)
	}
}

// A file:// URL whose file does not exist must pass through unchanged (no worse than
// today): we only rewrite when we can resolve a real local image.
func TestRewriteScreenshotPath_fileurl_missing_passthrough(t *testing.T) {
	in := "file:///no/such/dir/Screenshot%20gone.png"
	out := string(RewriteScreenshotPath([]byte(in)))
	if out != in {
		t.Errorf("missing file:// must pass through unchanged: got %q want %q", out, in)
	}
}

// A file:// URL pointing at an ephemeral screencaptureui temp file (which macOS
// deletes moments after the drop) must be copied to the stable, space-free dir and
// the path rewritten to the stable copy — same contract as the plain-path ephemeral
// case, but reached via a file:// URL.
func TestRewriteScreenshotPath_fileurl_ephemeral_copies_to_stable(t *testing.T) {
	root := t.TempDir()
	tempDir := filepath.Join(root, "TemporaryItems", "NSIRD_screencaptureui_ABC")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(tempDir, "Screenshot 2026-06-20 at 2.04.15 PM.png")
	if err := os.WriteFile(src, []byte("PNGDATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	stash := filepath.Join(root, "stash")
	t.Setenv("GT_SCREENSHOT_STASH_DIR", stash)

	urlStr := (&url.URL{Scheme: "file", Path: src}).String()
	out := string(RewriteScreenshotPath([]byte(urlStr)))

	if !strings.HasPrefix(out, stash) {
		t.Fatalf("ephemeral file:// should copy to stash %q; got %q", stash, out)
	}
	if strings.Contains(out, " ") {
		t.Errorf("stable path should be space-free: %q", out)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("stable copy unreadable: %v", err)
	}
	if string(data) != "PNGDATA" {
		t.Errorf("stable copy content = %q, want PNGDATA", data)
	}
}

// Claude can't attach video files (proven: a plain-path .mov stays literal text), so
// a dropped video must be turned into image frames. isVideoPath gates that.
func TestIsVideoPath(t *testing.T) {
	cases := map[string]bool{
		"/x/clip.mov": true, "/x/clip.MOV": true, "/x/a.mp4": true,
		"/x/a.m4v": true, "/x/a.webm": true, "/x/a.mkv": true,
		"/x/shot.png": false, "/x/doc.pdf": false, "/x/a.txt": false, "/x/noext": false,
	}
	for p, want := range cases {
		if got := isVideoPath(p); got != want {
			t.Errorf("isVideoPath(%q) = %v, want %v", p, got, want)
		}
	}
}

// Claude attaches multiple images only as SEPARATE bracketed pastes (proven via a
// live TUI: newline-joined paths in one paste do NOT attach; N separate pastes give
// [Image #1..N]). framesToPayload joins frames with "201~ 200~" so Filter.Process's
// single 200~/201~ wrap splits them into one paste per frame.
func TestFramesToPayload_splits_into_separate_pastes(t *testing.T) {
	got := string(framesToPayload([]string{"/s/a.png", "/s/b.png", "/s/c.png"}))
	want := "/s/a.png" + pasteEnd + pasteStart + "/s/b.png" + pasteEnd + pasteStart + "/s/c.png"
	if got != want {
		t.Errorf("framesToPayload = %q, want %q", got, want)
	}
}

// End-to-end through Filter: a Rewrite that returns framesToPayload must emerge from
// Process as N independent bracketed pastes (the shape Claude attaches as N images).
func TestFilter_video_payload_emerges_as_separate_pastes(t *testing.T) {
	f := &Filter{Rewrite: func([]byte) []byte {
		return framesToPayload([]string{"/s/a.png", "/s/b.png"})
	}}
	got := string(f.Process([]byte(pasteStart + "/whatever/clip.mov" + pasteEnd)))
	want := pasteStart + "/s/a.png" + pasteEnd + pasteStart + "/s/b.png" + pasteEnd
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

// A dropped video (delivered as a file:// URL, like a Finder drag) is decoded,
// frames are extracted with ffmpeg into the stable space-free dir, and the paste is
// rewritten to those frame paths as separate pastes. Requires ffmpeg.
func TestRewriteScreenshotPath_video_extracts_frames(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	root := t.TempDir()
	stash := filepath.Join(root, "stash")
	t.Setenv("GT_SCREENSHOT_STASH_DIR", stash)
	t.Setenv("GT_VIDEO_MAX_FRAMES", "3")

	src := filepath.Join(root, "Screen Recording test.mov")
	gen := exec.Command("ffmpeg", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "testsrc=duration=2:size=320x240:rate=10",
		"-pix_fmt", "yuv420p", src)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("could not generate test video: %v\n%s", err, out)
	}

	urlStr := (&url.URL{Scheme: "file", Path: src}).String()
	out := string(RewriteScreenshotPath([]byte(urlStr)))

	if strings.HasPrefix(out, "file://") || strings.Contains(out, ".mov") {
		t.Fatalf("video URL must be replaced with frame images, got %q", out)
	}
	frames := strings.Split(out, pasteEnd+pasteStart)
	if len(frames) < 2 {
		t.Fatalf("expected multiple frames, got %d: %q", len(frames), out)
	}
	if len(frames) > 3 {
		t.Errorf("GT_VIDEO_MAX_FRAMES=3 must cap frame count, got %d", len(frames))
	}
	for _, fr := range frames {
		if !strings.HasPrefix(fr, stash) {
			t.Errorf("frame %q should live under stash %q", fr, stash)
		}
		if strings.Contains(fr, " ") {
			t.Errorf("frame path should be space-free: %q", fr)
		}
		if !strings.HasSuffix(strings.ToLower(fr), ".png") {
			t.Errorf("frame should be a png: %q", fr)
		}
		if _, err := os.Stat(fr); err != nil {
			t.Errorf("frame file missing: %v", err)
		}
	}
}

// A file with a video extension that ffmpeg can't decode (here: empty) must pass
// through unchanged — never worse than handing the drop straight to Claude.
func TestRewriteScreenshotPath_unreadable_video_passthrough(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	dir := t.TempDir()
	t.Setenv("GT_SCREENSHOT_STASH_DIR", filepath.Join(dir, "stash"))
	src := filepath.Join(dir, "broken.mov")
	if err := os.WriteFile(src, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	in := src
	out := string(RewriteScreenshotPath([]byte(in)))
	if out != in {
		t.Errorf("undecodable video must pass through unchanged: got %q want %q", out, in)
	}
}

func upper(b []byte) []byte { return bytes.ToUpper(b) }

// Normal keystrokes pass through untouched.
func TestFilter_passthrough(t *testing.T) {
	f := &Filter{Rewrite: upper}
	if got := f.Process([]byte("hello")); string(got) != "hello" {
		t.Errorf("got %q want hello", got)
	}
}

// A whole bracketed paste is rewritten via Rewrite, markers preserved.
func TestFilter_rewrites_bracketed_paste(t *testing.T) {
	f := &Filter{Rewrite: upper}
	got := f.Process([]byte("\x1b[200~hi there\x1b[201~"))
	want := "\x1b[200~HI THERE\x1b[201~"
	if string(got) != want {
		t.Errorf("got %q want %q", got, want)
	}
}

// The start marker split across two reads must not corrupt the stream.
func TestFilter_split_start_marker(t *testing.T) {
	f := &Filter{Rewrite: upper}
	var out []byte
	out = append(out, f.Process([]byte("ab\x1b[20"))...) // partial start marker held back
	out = append(out, f.Process([]byte("0~hi\x1b[201~"))...)
	want := "ab\x1b[200~HI\x1b[201~"
	if string(out) != want {
		t.Errorf("got %q want %q", out, want)
	}
}

// The end marker split across two reads must still rewrite correctly.
func TestFilter_split_end_marker(t *testing.T) {
	f := &Filter{Rewrite: upper}
	var out []byte
	out = append(out, f.Process([]byte("\x1b[200~hi\x1b[20"))...)
	out = append(out, f.Process([]byte("1~rest"))...)
	want := "\x1b[200~HI\x1b[201~rest"
	if string(out) != want {
		t.Errorf("got %q want %q", out, want)
	}
}

// Text before and after a paste is preserved.
func TestFilter_text_around_paste(t *testing.T) {
	f := &Filter{Rewrite: upper}
	got := f.Process([]byte("x\x1b[200~hi\x1b[201~y"))
	want := "x\x1b[200~HI\x1b[201~y"
	if string(got) != want {
		t.Errorf("got %q want %q", got, want)
	}
}
