package tui_test

import (
	"strings"
	"testing"

	"github.com/jackuait/ghost-tab/internal/tui"
)

func TestZzzAnimation_InitialFrame(t *testing.T) {
	z := tui.NewZzzAnimation()
	if z.Frame() != 0 {
		t.Errorf("initial frame should be 0, got %d", z.Frame())
	}
}

func TestZzzAnimation_TickAdvancesFrame(t *testing.T) {
	z := tui.NewZzzAnimation()
	z.Tick()
	if z.Frame() != 1 {
		t.Errorf("after tick: expected frame 1, got %d", z.Frame())
	}
}

func TestZzzAnimation_FrameWraps(t *testing.T) {
	z := tui.NewZzzAnimation()
	totalFrames := z.TotalFrames()
	for i := 0; i < totalFrames; i++ {
		z.Tick()
	}
	if z.Frame() != 0 {
		t.Errorf("after %d ticks: expected frame 0 (wrapped), got %d", totalFrames, z.Frame())
	}
}

func TestZzzAnimation_FramesDiffer(t *testing.T) {
	z := tui.NewZzzAnimation()
	frame0 := z.View()
	z.Tick()
	frame1 := z.View()

	if frame0 == frame1 {
		t.Error("consecutive Zzz frames should differ")
	}
}

func TestZzzAnimation_Reset(t *testing.T) {
	z := tui.NewZzzAnimation()
	z.Tick()
	z.Tick()
	z.Reset()
	if z.Frame() != 0 {
		t.Errorf("after reset: expected frame 0, got %d", z.Frame())
	}
}

func TestZzzAnimation_ViewContainsZ(t *testing.T) {
	z := tui.NewZzzAnimation()
	view := z.View()
	if view == "" {
		t.Error("Zzz view should not be empty")
	}
}

func TestZzzAnimation_ViewColored(t *testing.T) {
	z := tui.NewZzzAnimation()
	colored := z.ViewColored("\033[35m")
	if colored == "" {
		t.Error("ViewColored should not be empty")
	}
	if !strings.Contains(colored, "\033[35m") {
		t.Error("ViewColored should contain the color code")
	}
	if !strings.Contains(colored, "\033[0m") {
		t.Error("ViewColored should contain reset code")
	}
}

func TestZzzAnimation_ViewAtAllFrames(t *testing.T) {
	z := tui.NewZzzAnimation()
	for i := 0; i < z.TotalFrames(); i++ {
		view := z.View()
		if view == "" {
			t.Errorf("Frame %d: View should not be empty", i)
		}
		z.Tick()
	}
}

func TestZzzAnimation_ViewColoredAtAllFrames(t *testing.T) {
	z := tui.NewZzzAnimation()
	for i := 0; i < z.TotalFrames(); i++ {
		colored := z.ViewColored("\033[36m")
		if colored == "" {
			t.Errorf("Frame %d: ViewColored should not be empty", i)
		}
		z.Tick()
	}
}
