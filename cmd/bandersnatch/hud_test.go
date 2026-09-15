package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/janearc/puffin-auklet/auklet"
	"github.com/janearc/puffin-auklet/scene"
)

// The HUD names every setting a key cycles, and the backdrop is one of them.
//
// It was the only cycling setting the HUD did not report. cmd/shot has printed
// the backdrop name all along, which left the interactive tool -- the one where
// you are actually pressing `b` -- as the single place the answer was missing.
// Jane, on the field of twinkling bulbs: "bandersnatch is not telling me the
// name of the background i'm looking at so i have no idea how they attained
// that."
//
// Every name is checked rather than one, because the bug was an absent field
// and an absent field passes a test that only ever looks at index zero.
func TestHUDNamesTheBackdrop(t *testing.T) {
	for i, want := range scene.BackdropNames {
		m := model{
			glyphs: auklet.Quadrant, cutout: true, backdrop: i, showVal: true,
			animate: true, next: 20, roar: -1, watch: true,
			ready: true, w: 200, h: 60, rows: 22,
		}
		hud := ansi.Strip(m.View())
		if !strings.Contains(hud, want) {
			t.Errorf("backdrop %d is %q and the HUD does not say so:\n%s", i, want, firstLine(hud))
		}
	}
}

// firstLine keeps a failure readable: the view is a whole screen.
func firstLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			return l
		}
	}
	return ""
}
