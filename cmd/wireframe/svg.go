package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/janearc/puffin-auklet/canvas"
)

// SVG, because a terminal scene is not shareable and a picture is.
//
// Jane wanted something to send a friend. Terminal output needs a terminal;
// an ANSI capture needs one too, and neither survives a phone. An SVG renders
// anywhere and needs nothing installed.
//
// BLOCK GLYPHS ARE DRAWN AS RECTANGLES, not as text. A sprite at quadrant
// resolution is almost entirely block characters, and text rendering of them
// depends on the viewer's font: gaps appear between cells, weights differ, and
// an eighth-block may not exist at all. Rectangles are exact everywhere and
// need no font. Anything that is not a block still goes out as text, which is
// what the dots in "lights" are.
//
// The cell is 10 by 21, near enough the 1:2.1 a terminal cell actually is that
// the auklet's proportions survive the trip.
const (
	cellW = 10
	cellH = 21
)

// quadrants maps a block glyph to the corners it fills, as (x, y, w, h) in
// halves of a cell. The set is what auklet's Quadrant and Half glyph sets can
// emit; a rune outside it falls through to text.
var quadrants = map[rune][][4]int{
	'█': {{0, 0, 2, 2}},
	'▀': {{0, 0, 2, 1}},
	'▄': {{0, 1, 2, 1}},
	'▌': {{0, 0, 1, 2}},
	'▐': {{1, 0, 1, 2}},
	'▘': {{0, 0, 1, 1}},
	'▝': {{1, 0, 1, 1}},
	'▖': {{0, 1, 1, 1}},
	'▗': {{1, 1, 1, 1}},
	'▚': {{0, 0, 1, 1}, {1, 1, 1, 1}},
	'▞': {{1, 0, 1, 1}, {0, 1, 1, 1}},
	'▛': {{0, 0, 2, 1}, {0, 1, 1, 1}},
	'▜': {{0, 0, 2, 1}, {1, 1, 1, 1}},
	'▙': {{0, 0, 1, 1}, {0, 1, 2, 1}},
	'▟': {{1, 0, 1, 1}, {0, 1, 2, 1}},
}

// hex is a lipgloss colour as #rrggbb, or "" when it is nil -- a nil colour
// means "keep what is underneath", which in an image means draw nothing.
func hex(c color.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, a := c.RGBA()
	if a == 0 {
		return ""
	}
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

// SVG renders a canvas. ground is painted behind everything, so a scene whose
// cells carry no background of their own is not transparent on a white page.
func SVG(c *canvas.Canvas, ground string) string {
	cells := c.Cells()
	w, h := c.Size()
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`,
		w*cellW, h*cellH, w*cellW, h*cellH)
	fmt.Fprintf(&b, `<rect width="100%%" height="100%%" fill="%s"/>`, ground)
	fmt.Fprintf(&b, `<g font-family="ui-monospace,SFMono-Regular,Menlo,Consolas,monospace" font-size="%d" text-anchor="middle">`, cellH*2/3)

	for y, row := range cells {
		for x, cell := range row {
			px, py := x*cellW, y*cellH
			if bg := hex(cell.BG); bg != "" {
				fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`, px, py, cellW, cellH, bg)
			}
			fg := hex(cell.FG)
			if cell.R == 0 || cell.R == ' ' || fg == "" {
				continue
			}
			if parts, ok := quadrants[cell.R]; ok {
				for _, q := range parts {
					fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`,
						px+q[0]*cellW/2, py+q[1]*cellH/2, q[2]*cellW/2, q[3]*cellH/2, fg)
				}
				continue
			}
			fmt.Fprintf(&b, `<text x="%d" y="%d" fill="%s">%s</text>`,
				px+cellW/2, py+cellH*3/4, fg, escape(cell.R))
		}
	}
	b.WriteString(`</g></svg>` + "\n")
	return b.String()
}

// escape keeps XML valid for the handful of runes that would break it.
func escape(r rune) string {
	switch r {
	case '&':
		return "&amp;"
	case '<':
		return "&lt;"
	case '>':
		return "&gt;"
	}
	return string(r)
}

// hexOr is a colour as hex, or a fallback when the theme does not set one.
func hexOr(c color.Color, fallback string) string {
	if h := hex(c); h != "" {
		return h
	}
	return fallback
}
