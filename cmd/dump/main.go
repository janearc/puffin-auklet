// dump prints a composed screen as one line per cell: "x y codepoint fg bg".
// it exists so a renderer can be verified without parsing escape sequences,
// which is a second, flakier implementation of the thing under test.
//
// the rune is a decimal code point, not a quoted character: a quoted space
// contains a space and silently shifts every field after it.
//
//	go run ./cmd/dump [theme] [backdrop] [rows] [cutout 0|1] [dx] [dy]
package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"

	"github.com/janearc/puffin-auklet/scene"
)

func hex(c lipgloss.TerminalColor) string {
	if c == nil {
		return "-"
	}
	if v, ok := c.(lipgloss.Color); ok {
		return string(v)
	}
	return "?"
}

func main() {
	o, _, _ := scene.Demo(os.Args[1:])
	c := scene.Build(o)
	w, h := c.Size()
	fmt.Printf("# %d %d 1 2\n", w, h)
	for y, row := range c.Cells() {
		for x, cell := range row {
			r := cell.R
			if r == 0 {
				r = ' '
			}
			fmt.Printf("%d %d %d %s %s\n", x, y, r, hex(cell.FG), hex(cell.BG))
		}
	}
}
