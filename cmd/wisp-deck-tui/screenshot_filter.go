package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/x/term"
	"github.com/creack/pty"
	"github.com/spf13/cobra"

	"github.com/jackuait/wisp-deck/internal/screenshotfilter"
)

// screenshot-filter runs a child command in a PTY and transparently proxies the
// terminal to it, except it rewrites a dropped screencaptureui temp-screenshot
// path (delivered as a bracketed paste) to a stable copy before the child reads
// it. This makes the literal drag-and-drop of a screenshot into the AI pane work
// even though macOS deletes the original temp file moments after the drop.
var screenshotFilterCmd = &cobra.Command{
	Use:                "screenshot-filter -- command [args...]",
	Short:              "Run a command in a PTY, rewriting dropped screenshot temp paths to stable copies",
	DisableFlagParsing: true,
	SilenceUsage:       true,
	SilenceErrors:      true,
	RunE:               runScreenshotFilter,
}

func init() { rootCmd.AddCommand(screenshotFilterCmd) }

func runScreenshotFilter(_ *cobra.Command, args []string) error {
	for len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		return errors.New("usage: wisp-deck-tui screenshot-filter -- <command> [args...]")
	}

	// Non-interactive (no tty on stdin): run the child transparently — the filter
	// only matters for a live terminal drop.
	if !term.IsTerminal(os.Stdin.Fd()) {
		c := exec.Command(args[0], args[1:]...)
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				os.Exit(ee.ExitCode())
			}
			return err
		}
		return nil
	}

	c := exec.Command(args[0], args[1:]...)
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	defer func() { _ = ptmx.Close() }()

	// Keep the child PTY sized to our terminal.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			_ = pty.InheritSize(os.Stdin, ptmx)
		}
	}()
	ch <- syscall.SIGWINCH
	defer signal.Stop(ch)

	// Raw mode so every byte (control sequences, raw keys) passes through to the
	// child unchanged; the child sets its own modes on the inner PTY.
	oldState, rawErr := term.MakeRaw(os.Stdin.Fd())
	restore := func() {
		if rawErr == nil {
			_ = term.Restore(os.Stdin.Fd(), oldState)
		}
	}

	// stdin -> filter -> child
	go func() {
		filt := screenshotfilter.New()
		b := make([]byte, 32*1024)
		for {
			n, rerr := os.Stdin.Read(b)
			if n > 0 {
				if out := filt.Process(b[:n]); len(out) > 0 {
					_, _ = ptmx.Write(out)
				}
			}
			if rerr != nil {
				return
			}
		}
	}()

	// child -> stdout (returns when the child exits and the PTY closes)
	_, _ = io.Copy(os.Stdout, ptmx)
	werr := c.Wait()
	restore()
	var ee *exec.ExitError
	if errors.As(werr, &ee) {
		os.Exit(ee.ExitCode())
	}
	return werr
}
