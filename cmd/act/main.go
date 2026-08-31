// act composes any character onto the same windowed, themed canvas
// bandersnatch does -- scene.Opts/scene.Build, unchanged -- and prints it
// straight to the shell. No TUI, no alt-screen, no keypresses: pick who,
// where on the canvas, and what they're doing, and it prints.
//
// A held pose is one frame:
//
//	go run ./cmd/act -character gopher -pose astronaut
//
// An emote actually plays, in place, in the terminal -- each frame clears
// and redraws, so a hop is a hop, not a description of one:
//
//	go run ./cmd/act -character lamp -emote hop
//
// Position and window size make it a canvas rather than just a sprite
// print: -sx/-sy move the sprite off centre, -win-w/-win-h resize the
// window around it, the same two independent things bandersnatch lets you
// drag with focus and arrows.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/janearc/puffin-auklet/auklet"
	"github.com/janearc/puffin-auklet/scene"
	"github.com/janearc/puffin-auklet/themes"
)

func main() {
	character := flag.String("character", "auklet", "who's on the canvas")
	view := flag.String("view", "front", "which of their views")
	themeName := flag.String("theme", "", "theme name; empty means the character's own default")
	backdrop := flag.Int("backdrop", 2, "backdrop index")
	rows := flag.Int("rows", 16, "sprite height in cells")
	winW := flag.Int("win-w", 46, "canvas window width, in cells")
	winH := flag.Int("win-h", 22, "canvas window height, in cells")
	cutout := flag.Bool("cutout", true, "cutout transparency")
	sx := flag.Int("sx", 0, "sprite position, cells off centre")
	sy := flag.Int("sy", 0, "sprite position, cells off centre")
	pose := flag.String("pose", "", "hold a named pose")
	emote := flag.String("emote", "", "play a named emote -- animates in place, in this terminal")
	fps := flag.Int("fps", 12, "playback rate for -emote")
	flag.Parse()

	var sprite auklet.Sprite
	found := false
	var charNames, viewNames []string
	for _, cs := range auklet.Characters() {
		charNames = append(charNames, cs.Name)
		if cs.Name != *character {
			continue
		}
		for _, s := range cs.Views {
			viewNames = append(viewNames, s.Name)
			if s.Name == *view {
				sprite, found = s, true
			}
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "unknown character/view %q/%q; characters: %v, views here: %v\n",
			*character, *view, charNames, viewNames)
		os.Exit(2)
	}

	tn := *themeName
	if tn == "" {
		tn = "corvid"
		for _, t := range themes.All {
			if t.Name == *character {
				tn = *character
			}
		}
	}
	var theme themes.Named
	themeFound := false
	var themeNames []string
	for _, t := range themes.All {
		themeNames = append(themeNames, t.Name)
		if t.Name == tn {
			theme, themeFound = t, true
		}
	}
	if !themeFound {
		fmt.Fprintf(os.Stderr, "unknown theme %q; have %v\n", tn, themeNames)
		os.Exit(2)
	}

	win := scene.Window{W: *winW, H: *winH, ScreenX: 3, ScreenY: 2}
	base := scene.Opts{
		Theme: theme, Backdrop: *backdrop, Glyphs: auklet.Quadrant, Sprite: sprite, Rows: *rows,
		Win: win, Cutout: *cutout, W: *winW + 8, H: *winH + 8,
	}
	place := func(o *scene.Opts, rows, dx, dy int) {
		o.Rows = rows
		o.SpriteX = win.WorldX + (win.W-sprite.ColsFor(rows))/2 + *sx + dx
		o.SpriteY = win.WorldY + (win.H-rows)/2 + *sy + dy
	}

	switch {
	case *emote != "":
		em, ok := sprite.Emote(*emote)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown emote %q for %s/%s; have %v\n", *emote, *character, *view, sprite.Emotes())
			os.Exit(2)
		}
		step := time.Second / time.Duration(*fps)
		reps := 1
		if em.Loop {
			reps = 3 // a loop plays a few cycles rather than forever unattended
		}
		total := em.Length() * time.Duration(reps)
		for t := time.Duration(0); t < total; t += step {
			fr, ok := em.At(t % em.Length())
			if !ok {
				break
			}
			o := base
			place(&o, auklet.ScaleRows(*rows, fr.Transform.Factor()), fr.Transform.DX, fr.Transform.DY)
			o.Pose = fr.Pose
			o.Frame = int(t / step)
			fmt.Print("\x1b[H\x1b[2J")
			fmt.Println(scene.Build(o).String())
			time.Sleep(step)
		}
	case *pose != "":
		p, ok := sprite.NamedPose(*pose)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown pose %q for %s/%s; have %v\n", *pose, *character, *view, sprite.PoseNames())
			os.Exit(2)
		}
		o := base
		place(&o, *rows, 0, 0)
		o.Pose = p
		fmt.Println(scene.Build(o).String())
	default:
		o := base
		place(&o, *rows, 0, 0)
		fmt.Println(scene.Build(o).String())
	}
}
