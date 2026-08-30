// shot renders one composed screen to stdout with no tty, so the viewer can be
// checked by eye or in ci without launching it.
//
//	go run ./cmd/shot [theme] [backdrop] [rows] [cutout 0|1] [dx] [dy]
//
// dx and dy offset the auklet from centred, in cells, so an out-of-view bird
// can be rendered deliberately.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/janearc/puffin-auklet/scene"
)

func main() {
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)

	o, name, note := scene.Demo(os.Args[1:])
	fmt.Fprintf(os.Stderr, "%s (%s)  %dx%d %s %s  visible=%v\n",
		name, note, o.Rows, o.Rows, o.Glyphs,
		scene.BackdropNames[o.Backdrop%len(scene.BackdropNames)], scene.Visible(o))
	if err := o.Theme.Theme.Validate(); err != nil {
		for _, l := range strings.Split(err.Error(), "\n") {
			fmt.Fprintln(os.Stderr, "  FAIL:", l)
		}
	} else {
		fmt.Fprintln(os.Stderr, "  PASS")
	}
	fmt.Println(scene.Build(o).String())
}
