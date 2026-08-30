// render writes the auklet to PNG frames with a real alpha channel, so it can
// be composited over a screen recording instead of filmed off a terminal.
//
// this is the same sprite, the same themes and the same mouth track the viewer
// uses -- only the last step differs. colour is chosen last by design, so a
// renderer that is not a terminal costs one file rather than a second bird.
//
//	go run ./cmd/render -text "hello. i am a puffin." -out frames/
//	go run ./cmd/render -still -scale 16 -out .
//
// transparent pixels are alpha 0, not a background colour, so there is no key
// to pull and no fringe to clean up.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/janearc/puffin-auklet/puffin"
	"github.com/janearc/puffin-auklet/themes"
)

func main() {
	var (
		themeName = flag.String("theme", "corvid", "dodo theme name")
		viewName  = flag.String("view", "front", "side or front")
		scale     = flag.Int("scale", 12, "pixels per source pixel")
		text      = flag.String("text", "hello. i am a puffin, and i will be narrating this software.", "narration")
		fps       = flag.Int("fps", 24, "frames per second")
		out       = flag.String("out", ".", "output directory")
		still     = flag.Bool("still", false, "write one frame with the beak shut")
		opaque    = flag.Bool("opaque", false, "fill the background instead of leaving it transparent")
		blink     = flag.Bool("blink", true, "blink on an idle timer")
		seed      = flag.Int64("seed", 1, "blink timing seed; same seed, same take")
	)
	flag.Parse()

	vars, ok := themes.Dodo[*themeName]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown theme %q; have %v\n", *themeName, themes.DodoOrder)
		os.Exit(2)
	}
	theme := vars.Puffin()
	if err := theme.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "theme %s does not validate:\n%v\n", *themeName, err)
		os.Exit(1)
	}
	if !*opaque {
		theme.Background = nil
	}

	var sprite puffin.Sprite
	for _, s := range puffin.Views() {
		if s.Name == *viewName {
			sprite = s
		}
	}
	if sprite.Name == "" {
		fmt.Fprintf(os.Stderr, "unknown view %q\n", *viewName)
		os.Exit(2)
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	track := []int{0}
	if !*still {
		track = puffin.MouthTrack(*text, *fps)
	}

	rng := rand.New(rand.NewSource(*seed))
	nextBlink, blinkFor := 20+rng.Intn(30), 0

	for i, level := range track {
		pose := append(puffin.Pose{}, sprite.Mouth(level)...)
		if *blink && !*still {
			switch {
			case blinkFor > 0:
				blinkFor--
				pose = append(pose, sprite.Blink...)
			case i >= nextBlink:
				blinkFor = *fps / 8
				nextBlink = i + *fps*2 + rng.Intn(*fps*3)
				pose = append(pose, sprite.Blink...)
			}
		}

		img := frame(sprite, theme, pose, *scale)
		name := filepath.Join(*out, fmt.Sprintf("auklet_%05d.png", i))
		if *still {
			name = filepath.Join(*out, "auklet.png")
		}
		f, err := os.Create(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		f.Close()
	}

	w, h := sprite.Size()
	fmt.Printf("%d frame(s), %dx%d px, %s/%s, alpha=%v -> %s\n",
		len(track), w**scale, h**scale, *themeName, sprite.Name, !*opaque, *out)
}

// frame draws one posed bird. every source pixel becomes a scale x scale block:
// nearest neighbour on purpose, because this is pixel art and smoothing it is
// the one thing that would make it look worse.
func frame(s puffin.Sprite, t puffin.Theme, pose puffin.Pose, scale int) image.Image {
	rows := s.Pixels(pose)
	w, h := len(rows[0])*scale, len(rows)*scale
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// role colours are fixed for the whole frame; resolve once
	cache := map[byte]color.RGBA{}
	lookup := func(role byte) color.RGBA {
		if c, ok := cache[role]; ok {
			return c
		}
		var out color.RGBA
		if r, g, b, ok := puffin.RGB(t.RoleColor(role)); ok {
			out = color.RGBA{r, g, b, 255}
		}
		cache[role] = out
		return out
	}

	for y := 0; y < h; y++ {
		src := rows[y/scale]
		for x := 0; x < w; x++ {
			img.Set(x, y, lookup(src[x/scale]))
		}
	}
	return img
}
