package auklet

import (
	"strings"
	"testing"
)

// The open mouth is a HOLE, not a black wedge.
//
// It was drawn in 'E', the pupil, and the theme said so out loud: "the eye,
// and the open mouth's interior". On an arcade cabinet that is right --
// the cabinet is black, so a black mouth and an absent one are the same
// picture. Off a cabinet they stop being the same picture. Jane, looking at
// Ms Pac-Man on a vaporwave page: a black wedge stuck to a yellow disc.
//
// '_' resolves to Background, so it is cabinet-black in the workbench and
// transparent wherever Background is nil. It cannot be '.', because '.'
// means "leave what is beneath" when an overlay composites and the thing
// beneath the mouth is the disc -- the mouth would simply not open.
func TestMsPacManMouthIsAHole(t *testing.T) {
	sp := mspacmanFront(t)
	e, ok := sp.Emote("chomp")
	if !ok {
		t.Fatal("mspacman has no chomp")
	}

	sawOpen := false
	for _, f := range e.Frames {
		grid := sp.Pixels(f.Pose)
		holes, pupils := count(grid, '_'), count(grid, 'E')
		if holes == 0 {
			continue // a closed frame
		}
		sawOpen = true
		// the eye survives every opening: it is a pupil in a ring, and it
		// is the one thing on this character that SHOULD be pupil-coloured
		if pupils == 0 {
			t.Errorf("an open frame has no pupil left; the eye was blanked along with the mouth:\n%s",
				strings.Join(grid, "\n"))
		}
		if !hasEyeInRing(grid) {
			t.Errorf("the pupil is no longer inside its ring:\n%s", strings.Join(grid, "\n"))
		}
	}
	if !sawOpen {
		t.Fatal("no frame of chomp opens the mouth")
	}
}

// A hole has to hold its edge. roleWeight decides which role wins a cell
// when several source pixels share one, and a hole with no weight loses
// every contested cell to the disc -- the mouth closes by half a pixel at
// exactly the sizes the corner uses.
func TestHoleHoldsItsEdge(t *testing.T) {
	if roleWeight['_'] < roleWeight['K'] {
		t.Errorf("a hole (%v) is lighter than the disc around it (%v); its edge will be eaten",
			roleWeight['_'], roleWeight['K'])
	}
}

// '_' must resolve to the field, which is what makes it black on a cabinet
// and transparent in a cutout without choosing between them.
func TestHoleResolvesToTheField(t *testing.T) {
	th := DefaultTheme()
	if got := th.colorFor('_'); got != th.Background {
		t.Errorf("colorFor('_') = %v, want the field %v", got, th.Background)
	}
	th.Background = nil
	if got := th.colorFor('_'); got != nil {
		t.Errorf("with a nil field, colorFor('_') = %v, want nil (transparent)", got)
	}
}

func mspacmanFront(t *testing.T) Sprite {
	t.Helper()
	for _, c := range Characters() {
		if c.Name != "mspacman" {
			continue
		}
		for _, v := range c.Views {
			if v.Name == "front" {
				return v
			}
		}
	}
	t.Fatal("no mspacman/front")
	return Sprite{}
}

func count(grid []string, role byte) int {
	n := 0
	for _, r := range grid {
		for i := 0; i < len(r); i++ {
			if r[i] == role {
				n++
			}
		}
	}
	return n
}

// hasEyeInRing looks for a pupil with ring pixels on both sides of it,
// which is the shape rather than merely the presence.
func hasEyeInRing(grid []string) bool {
	for _, r := range grid {
		if i := strings.Index(r, "E"); i > 0 && i+1 < len(r) {
			j := i
			for j < len(r) && r[j] == 'E' {
				j++
			}
			if r[i-1] == 'X' && j < len(r) && r[j] == 'X' {
				return true
			}
		}
	}
	return false
}
