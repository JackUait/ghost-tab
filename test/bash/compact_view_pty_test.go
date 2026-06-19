package bash_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/creack/pty"
)

// Regression: when the user scrolls fast, SGR mouse reports must never leak onto
// the screen as literal text (e.g. "[<65;40;18M"). `read -s` only silences echo
// while it is actively reading; scroll events that arrive during the render gap
// get echoed by the tty's line discipline. The fix disables terminal echo for
// the interactive session (stty -echo). This test drives the REAL loop over a
// pty, fires a burst of wheel-down reports, and asserts none echo back.
func TestCompactView_does_not_echo_mouse_reports(t *testing.T) {
	module := filepath.Join(projectRoot(t), "lib", "compact-view.sh")

	// A repo with a tall modified file so the ledger overflows and scrolls.
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		c := exec.Command("git", append([]string{"-C", dir}, args...)...)
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	writeTempFile(t, dir, "app.txt", "one\n")
	git("add", "app.txt")
	git("commit", "-q", "-m", "init")
	var tall bytes.Buffer
	for i := 0; i < 40; i++ {
		tall.WriteString("changed line\n")
	}
	writeTempFile(t, dir, "app.txt", tall.String())

	cmd := exec.Command("bash", "-c", "source "+module+" && compact_view "+dir)
	env := []string{}
	for _, e := range os.Environ() {
		if len(e) >= 5 && e[:5] == "TMUX=" {
			continue
		}
		env = append(env, e)
	}
	cmd.Env = append(env, "COMPACT_VIEW_INTERVAL=1", "TERM=xterm")

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 12, Cols: 60})
	if err != nil {
		t.Fatalf("start pty: %v", err)
	}
	defer func() { _ = ptmx.Close() }()

	var mu sync.Mutex
	var out bytes.Buffer
	go func() {
		b := make([]byte, 4096)
		for {
			n, err := ptmx.Read(b)
			if n > 0 {
				mu.Lock()
				out.Write(b[:n])
				mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()

	time.Sleep(600 * time.Millisecond) // let the first frame render
	// Burst of wheel-down reports with tiny gaps so many land during a render.
	for i := 0; i < 40; i++ {
		_, _ = ptmx.Write([]byte("\x1b[<65;40;18M"))
		time.Sleep(5 * time.Millisecond)
	}
	time.Sleep(1200 * time.Millisecond)
	_, _ = ptmx.Write([]byte{0x03}) // Ctrl-C
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	got := out.String()
	mu.Unlock()

	leak := regexp.MustCompile(`\[<\d+;\d+;\d+M`)
	if m := leak.FindAllString(got, -1); len(m) > 0 {
		t.Errorf("mouse reports echoed to screen %d time(s) (terminal echo not disabled); first: %q",
			len(m), m[0])
	}
}
