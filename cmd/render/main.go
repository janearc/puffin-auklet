// render writes a narrating auklet out as frames.
//
// three formats, and the default is the one that keeps the bird made of
// terminal cells:
//
//	cast   an asciicast v2 recording. one file, plays in a terminal, and the
//	       existing tools turn it into a gif or an mp4. this is the one you
//	       want for a video about terminal software.
//	ansi   one .ans file per frame, if you would rather drive the timing
//	       yourself.
//	png    pixel frames with a real alpha channel, for compositing the bird
//	       over footage as an image rather than as text.
//
//	go run ./cmd/render -text "hello. i am a auklet." -out auklet.cast
//	go run ./cmd/render -format png -still -scale 16 -out .
//
// in cast and ansi the background is simply never set, so the terminal's own
// background shows through and the bird composites onto whatever is behind it.
// in png that same transparency becomes alpha 0 -- no key to pull, no fringe.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/janearc/puffin-auklet/auklet"
	"github.com/janearc/puffin-auklet/canvas"
	"github.com/janearc/puffin-auklet/themes"
)

func main() {
	var (
		themeName = flag.String("theme", "corvid", "dodo theme name")
		viewName  = flag.String("view", "front", "side or front")
		format    = flag.String("format", "cast", "cast, ansi or png")
		rows      = flag.Int("rows", 22, "sprite height in terminal cells (cast, ansi)")
		glyphs    = flag.String("glyphs", "quadrant", "half, quadrant or sextant (cast, ansi)")
		scale     = flag.Int("scale", 12, "pixels per source pixel (png only)")
		text      = flag.String("text", "hello. i am a puffin, and i will be narrating this software.", "narration")
		fps       = flag.Int("fps", 24, "frames per second")
		out       = flag.String("out", "auklet.cast", "output file, or directory for ansi and png")
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
	theme := vars.Auklet()
	if err := theme.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "theme %s does not validate:\n%v\n", *themeName, err)
		os.Exit(1)
	}
	if !*opaque {
		theme.Background = nil
	}

	var sprite auklet.Sprite
	for _, s := range auklet.Views() {
		if s.Name == *viewName {
			sprite = s
		}
	}
	if sprite.Name == "" {
		fmt.Fprintf(os.Stderr, "unknown view %q\n", *viewName)
		os.Exit(2)
	}

	gs := auklet.Quadrant
	switch *glyphs {
	case "half":
		gs = auklet.Half
	case "sextant":
		gs = auklet.Sextant
	case "quadrant":
	default:
		fmt.Fprintf(os.Stderr, "unknown glyph set %q\n", *glyphs)
		os.Exit(2)
	}

	track := []int{0}
	if !*still {
		track = auklet.MouthTrack(*text, *fps)
	}

	rng := rand.New(rand.NewSource(*seed))
	nextBlink, blinkFor := 20+rng.Intn(30), 0

	poses := make([]auklet.Pose, len(track))
	for i, level := range track {
		p := append(auklet.Pose{}, sprite.Mouth(level)...)
		if *blink && !*still {
			switch {
			case blinkFor > 0:
				blinkFor--
				p = append(p, sprite.Blink...)
			case i >= nextBlink:
				blinkFor = *fps / 8
				nextBlink = i + *fps*2 + rng.Intn(*fps*3)
				p = append(p, sprite.Blink...)
			}
		}
		poses[i] = p
	}

	switch *format {
	case "cast":
		if err := writeCast(*out, sprite, theme, gs, *rows, poses, *fps); err != nil {
			fail(err)
		}
		fmt.Printf("%d frames, %dx%d cells, %s/%s/%s -> %s\n",
			len(poses), sprite.ColsFor(*rows), *rows, *themeName, sprite.Name, gs, *out)

	case "ansi":
		if err := os.MkdirAll(*out, 0o755); err != nil {
			fail(err)
		}
		for i, p := range poses {
			cells := sprite.CellsAt(theme, gs, sprite.ColsFor(*rows), *rows, p)
			name := filepath.Join(*out, fmt.Sprintf("auklet_%05d.ans", i))
			if *still {
				name = filepath.Join(*out, "auklet.ans")
			}
			if err := os.WriteFile(name, []byte(canvas.Render(cells)+"\n"), 0o644); err != nil {
				fail(err)
			}
		}
		fmt.Printf("%d frames, %dx%d cells, %s/%s/%s -> %s\n",
			len(poses), sprite.ColsFor(*rows), *rows, *themeName, sprite.Name, gs, *out)

	case "png":
		if err := os.MkdirAll(*out, 0o755); err != nil {
			fail(err)
		}
		for i, p := range poses {
			name := filepath.Join(*out, fmt.Sprintf("auklet_%05d.png", i))
			if *still {
				name = filepath.Join(*out, "auklet.png")
			}
			f, err := os.Create(name)
			if err != nil {
				fail(err)
			}
			if err := png.Encode(f, frame(sprite, theme, p, *scale)); err != nil {
				f.Close()
				fail(err)
			}
			f.Close()
		}
		w, h := sprite.Size()
		fmt.Printf("%d frames, %dx%d px, %s/%s, alpha=%v -> %s\n",
			len(poses), w**scale, h**scale, *themeName, sprite.Name, !*opaque, *out)

	default:
		fmt.Fprintf(os.Stderr, "unknown format %q; want cast, ansi or png\n", *format)
		os.Exit(2)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// writeCast emits asciicast v2: a JSON header, then one [time, "o", data] event
// per frame. each frame homes the cursor and repaints rather than clearing, so
// a player never shows a blank screen between frames.
func writeCast(path string, s auklet.Sprite, t auklet.Theme, gs auklet.GlyphSet,
	rows int, poses []auklet.Pose, fps int) error {

	// a file is not a terminal, so nothing will infer truecolor for us
	lipgloss.SetColorProfile(termenv.TrueColor)

	cols := s.ColsFor(rows)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	header := map[string]any{
		"version":   2,
		"width":     cols,
		"height":    rows,
		"timestamp": time.Now().Unix(),
		"env":       map[string]string{"TERM": "xterm-256color"},
	}
	h, err := json.Marshal(header)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(f, "%s\n", h); err != nil {
		return err
	}

	for i, p := range poses {
		cells := s.CellsAt(t, gs, cols, rows, p)
		body := strings.ReplaceAll(canvas.Render(cells), "\n", "\r\n")

		data := "\u001b[H" + body
		if i == 0 {
			data = "\u001b[2J\u001b[?25l" + data // clear once, and hide the cursor
		}

		ev, err := json.Marshal([]any{float64(i) / float64(fps), "o", data})
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(f, "%s\n", ev); err != nil {
			return err
		}
	}
	return nil
}

// frame draws one posed bird. every source pixel becomes a scale x scale block:
// nearest neighbour on purpose, because this is pixel art and smoothing it is
// the one thing that would make it look worse.
func frame(s auklet.Sprite, t auklet.Theme, pose auklet.Pose, scale int) image.Image {
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
		if r, g, b, ok := auklet.RGB(t.RoleColor(role)); ok {
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
