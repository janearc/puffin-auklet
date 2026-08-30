// shot renders one composed scene to stdout with no tty, so the sprite can be
// checked by eye or in ci without launching the viewer.
//
//	go run ./cmd/shot [themeIdx] [backdropIdx] [scale] [cutout 0|1]
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/janearc/puffin-auklet/puffin"
	"github.com/janearc/puffin-auklet/scene"
	"github.com/janearc/puffin-auklet/themes"
)

func argi(i, def int) int {
	if len(os.Args) > i+1 {
		if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
			return v
		}
	}
	return def
}

func main() {
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)

	ti, bi, sc, cut := argi(0, 0), argi(1, 0), argi(2, 1), argi(3, 0)
	cur := themes.All[ti%len(themes.All)]

	w, h := 60, 26
	sw, sh := puffin.SizeAt(sc)
	c := scene.Build(scene.Opts{
		Theme: cur, Backdrop: bi, Scale: sc, Cutout: cut == 1,
		X: (w - sw) / 2, Y: (h - sh) / 2, W: w, H: h,
	})

	fmt.Fprintf(os.Stderr, "%s (%s)  scale=%d  backdrop=%s  cutout=%v\n",
		cur.Name, cur.Note, sc, scene.BackdropNames[bi%len(scene.BackdropNames)], cut == 1)
	if err := cur.Theme.Validate(); err != nil {
		for _, l := range strings.Split(err.Error(), "\n") {
			fmt.Fprintln(os.Stderr, "  FAIL:", l)
		}
	} else {
		fmt.Fprintln(os.Stderr, "  PASS")
	}
	fmt.Println(c.String())
}
