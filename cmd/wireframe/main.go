// wireframe fills a daffy layout with auklet backdrops.
//
// daffy draws and this renders. That separation is the point: daffy will not
// run code for you, and it should not -- a drawing tool that executes what is
// in a document is a drawing tool you cannot open a stranger's file with. So
// the document says WHERE, by name, and this says WITH WHAT.
//
// The seam it uses was already there. daffy's `export json` writes the regions
// and the theme as data -- "for whatever consumes them next", in its own words
// -- and scene.Backdrop takes a canvas rather than a position, so it has never
// known where on screen it is. Carve a child canvas per region, fill it, blit
// it back. Canvas.Blit already clips at every edge and skips empty cells, so
// transparency and overlap come for free.
//
// Usage:
//
//	wireframe -layout roster.json                  # fills from each region's notes
//	wireframe -layout roster.json -fill outer=lights -fill berth=gopher
//	wireframe -layout roster.json -frame 7         # a later phase of the twinkle
//	wireframe -layout roster.json -animate         # run it; ctrl-c to stop
//	wireframe -layout roster.json -list            # what the file contains
//	wireframe --age                                # what this binary is
//	wireframe -layout roster.json -svg scene.svg   # a picture, for anywhere
//
// A fill is a BACKDROP name or a CHARACTER name. Backdrops fill the region;
// characters stand in it, scaled to fit and centred, with their transparent
// parts showing whatever is underneath -- which is what a "berth" is for.
// Backdrop names are tried first; today the two sets do not overlap.
//
// A region is filled if -fill names it, else if its notes carry `fill`. A
// region with neither is left alone rather than guessed at, because an
// unfilled region and a region filled with the wrong thing look identical on
// a first run and only one of them is honest.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/janearc/puffin-auklet/auklet"
	"github.com/janearc/puffin-auklet/buildinfo"
	"github.com/janearc/puffin-auklet/canvas"
	"github.com/janearc/puffin-auklet/scene"
	"github.com/janearc/puffin-auklet/themes"
)

// Layout is daffy's `export json`, in the fields this needs.
//
// Declared here rather than imported so that this tool does not depend on
// daffy's package: the JSON is the published contract between them, and a
// shared struct would make two repositories move together for no gain.
// Unknown fields are ignored, which is what lets daffy add to it without
// breaking this.
type Layout struct {
	Name    string   `json:"name"`
	Cols    int      `json:"cols"`
	Rows    int      `json:"rows"`
	Theme   string   `json:"theme"`
	Regions []Region `json:"regions"`
}

// Region is one named area. Cells is [x0, y0, x1, y1] in daffy's order.
type Region struct {
	Name  string            `json:"name"`
	Kind  string            `json:"kind"`
	Cells [4]int            `json:"cells"`
	Shape bool              `json:"shape"`
	Notes map[string]string `json:"notes"`
}

// Size is the region's width and height in cells, never negative.
func (r Region) Size() (w, h int) {
	w, h = r.Cells[2]-r.Cells[0], r.Cells[3]-r.Cells[1]
	return max(w, 0), max(h, 0)
}

// fills collects repeated -fill name=backdrop flags.
type fills map[string]string

func (f fills) String() string { return "" }

func (f fills) Set(v string) error {
	name, back, ok := strings.Cut(v, "=")
	if !ok {
		return fmt.Errorf("want name=backdrop, got %q", v)
	}
	f[name] = back
	return nil
}

func main() {
	path := flag.String("layout", "", "a daffy `export json` file")
	frame := flag.Int("frame", 0, "the animation frame to draw; backdrops phase on it")
	theme := flag.String("theme", "", "auklet theme to draw with; the layout's own name is tried first")
	list := flag.Bool("list", false, "describe the layout and exit")
	animate := flag.Bool("animate", false, "redraw continuously; ctrl-c to stop")
	fps := flag.Int("fps", 8, "frames a second with -animate")
	age := flag.Bool("age", false, "say what this binary is and when it was built, then exit")
	svg := flag.String("svg", "", "write the scene to an SVG `file` instead of the terminal")
	fill := fills{}
	flag.Var(fill, "fill", "region=backdrop, repeatable")
	flag.Parse()

	// answered before -layout is required, so the question can be asked of a
	// binary you have not decided how to use yet
	if *age {
		fmt.Print(buildinfo.Read().Age())
		return
	}

	if *path == "" {
		fmt.Fprintln(os.Stderr, "wireframe: -layout is required")
		fmt.Fprintln(os.Stderr, "backdrops:", strings.Join(scene.BackdropNames, " "))
		os.Exit(2)
	}
	l, err := load(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "wireframe:", err)
		os.Exit(1)
	}
	if *list {
		describe(os.Stdout, l, fill)
		return
	}
	t, used := pickTheme(l.Theme, *theme)
	// an SVG is the shareable form: it needs no terminal, no font with block
	// glyphs, and nothing installed to look at it
	if *svg != "" {
		// lipgloss degrades to no colour when stdout is not a terminal, and a
		// degraded colour resolves to black -- which is correct for a pipe and
		// useless for a picture. An SVG has no terminal to detect, so the
		// profile is forced, the way cmd/shot does for the same reason.
		lipgloss.SetColorProfile(termenv.TrueColor)
		c, warns := build(l, fill, t, *frame)
		if err := os.WriteFile(*svg, []byte(SVG(c, hexOr(t.Background, "#12121a"))), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "wireframe:", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%s  %dx%d  theme %s  frame %d  -> %s\n", l.Name, l.Cols, l.Rows, used, *frame, *svg)
		for _, w := range warns {
			fmt.Fprintln(os.Stderr, "  "+w)
		}
		return
	}
	if *animate {
		run(l, fill, t, *frame, *fps, used)
		return
	}
	out, warns := render(l, fill, t, *frame)
	fmt.Print(out)
	// the theme actually used is named on stderr, because a layout asking for
	// a theme this tree does not have is a real difference between the picture
	// daffy showed and the one here, and silence would hide it
	fmt.Fprintf(os.Stderr, "%s  %dx%d  theme %s  frame %d\n", l.Name, l.Cols, l.Rows, used, *frame)
	for _, w := range warns {
		fmt.Fprintln(os.Stderr, "  "+w)
	}
}

// load reads and validates a layout.
func load(path string) (*Layout, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var l Layout
	if err := json.Unmarshal(body, &l); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if l.Cols <= 0 || l.Rows <= 0 {
		return nil, fmt.Errorf("%s: layout is %dx%d", path, l.Cols, l.Rows)
	}
	return &l, nil
}

// pickTheme honours the layout's theme name when this tree has one by that
// name, and says which it settled on either way.
func pickTheme(want, override string) (auklet.Theme, string) {
	for _, name := range []string{override, want} {
		if name == "" {
			continue
		}
		for _, n := range themes.All {
			if strings.EqualFold(n.Name, name) {
				return n.Theme, n.Name
			}
		}
	}
	return themes.All[0].Theme, themes.All[0].Name + " (fallback)"
}

// render draws every region that has a fill, and reports what it could not do
// rather than dropping it.
func render(l *Layout, fill fills, t auklet.Theme, frame int) (string, []string) {
	c, warns := build(l, fill, t, frame)
	return c.String() + "\n", warns
}

// build composes the scene and hands back the canvas, so a caller that wants
// pixels rather than escape codes has something to walk.
func build(l *Layout, fill fills, t auklet.Theme, frame int) (*canvas.Canvas, []string) {
	var warns []string
	root := canvas.New(l.Cols, l.Rows)

	for _, r := range l.Regions {
		name := fillFor(r, fill)
		if name == "" {
			continue
		}
		w, h := r.Size()
		if w == 0 || h == 0 {
			warns = append(warns, fmt.Sprintf("region %q is %dx%d, nothing drawn", r.Name, w, h))
			continue
		}
		// A region asking for a fill and one cell across is almost certainly a
		// collapsed drag rather than an intention: no backdrop is legible in a
		// single cell, and daffy today gives no way to see a stray one or to
		// remove it except by hitting it, which at one cell is nearly
		// impossible. Found in the wild -- wireframe-2.daffy carried a 1x1
		// named "ground", kind "backdrop", where the name and kind said what
		// was meant and the geometry said the drag had collapsed. Drawn
		// anyway, because refusing to render what the file says is worse than
		// saying it looks wrong.
		if w < 2 || h < 2 {
			warns = append(warns, fmt.Sprintf(
				"region %q is %dx%d and asks for %q; a fill that small is usually a collapsed drag",
				r.Name, w, h, name))
		}
		// a child canvas the size of the region. Neither a backdrop nor a
		// character learns where this sits, which is why the same code works
		// as a full-screen field and as a nineteen-cell berth.
		child := canvas.New(w, h)
		switch {
		case scene.Backdrop(child, name, t, nil, frame):
			// filled
		case stand(child, name, t):
			// a character stands here
		default:
			warns = append(warns, fmt.Sprintf("region %q wants %q, which is neither a backdrop (%s) nor a character (%s)",
				r.Name, name, strings.Join(scene.BackdropNames, " "), strings.Join(characterNames(), " ")))
			continue
		}
		root.BlitCanvas(child, r.Cells[0], r.Cells[1])
		if r.Shape {
			// daffy says this region has a pixel mask and the JSON does not
			// carry it. Drawn as its bounding box, and said out loud, because
			// a rectangle where a shape was meant is exactly the kind of wrong
			// that looks fine.
			warns = append(warns, fmt.Sprintf("region %q has a mask daffy did not export; drawn as its box", r.Name))
		}
	}
	return root, warns
}

// fillFor is the backdrop a region should get: the flag first, then the
// region's own `fill` note, then nothing.
func fillFor(r Region, fill fills) string {
	if v, ok := fill[r.Name]; ok {
		return v
	}
	return r.Notes["fill"]
}

// describe prints what the file contains without drawing it, so a wireframe
// can be inspected before it is filled.
func describe(w *os.File, l *Layout, fill fills) {
	fmt.Fprintf(w, "%s  %dx%d  theme %s  %d region(s)\n", l.Name, l.Cols, l.Rows, l.Theme, len(l.Regions))
	rs := append([]Region(nil), l.Regions...)
	sort.Slice(rs, func(i, j int) bool { return rs[i].Name < rs[j].Name })
	for _, r := range rs {
		cw, ch := r.Size()
		f := fillFor(r, fill)
		if f == "" {
			f = "-"
		}
		shape := ""
		if r.Shape {
			shape = "  (masked)"
		}
		fmt.Fprintf(w, "  %-16s %-10s %3d,%-3d %3dx%-3d  fill %s%s\n",
			r.Name, r.Kind, r.Cells[0], r.Cells[1], cw, ch, f, shape)
	}
	fmt.Fprintln(w, "backdrops:", strings.Join(scene.BackdropNames, " "))
}

// stand puts a character in the canvas, scaled to fit and centred, and reports
// whether it knew the name.
//
// The sprite carries no background of its own -- the theme's is cleared first,
// the way scene.Build does for a cutout -- so its transparent cells show
// whatever the region is sitting on. A gopher in a berth over a field of bulbs
// should have bulbs behind it, not a rectangle of theme colour.
//
// The auklet scales ITSELF: RowsToFit picks the tallest that fits the region
// and ColsFor derives the width from the art's own aspect. Nothing here
// decides how big a character should be, which is the same division the
// backdrops use -- the region says how much room there is and the thing in it
// works out the rest.
func stand(c *canvas.Canvas, name string, t auklet.Theme) bool {
	var views []auklet.Sprite
	for _, ch := range auklet.Characters() {
		if strings.EqualFold(ch.Name, name) {
			views = ch.Views
			break
		}
	}
	if len(views) == 0 {
		return false
	}
	sp := views[0]
	w, h := c.Size()

	t.Background = nil // see through the character, not around it
	rows := sp.RowsToFit(w, h)
	cols := sp.ColsFor(rows)
	// centred, and never negative when the art is wider than the berth
	c.Blit(sp.CellsAt(t, auklet.Quadrant, cols, rows, nil), max((w-cols)/2, 0), max((h-rows)/2, 0))
	return true
}

// characterNames is the roster, for the message when a fill matches nothing.
func characterNames() []string {
	out := make([]string, 0, len(auklet.Characters()))
	for _, ch := range auklet.Characters() {
		out = append(out, ch.Name)
	}
	return out
}

// run redraws until interrupted.
//
// The alternate screen, so the scene has the terminal to itself and the shell
// comes back untouched -- a loop that scrolls a scene up the scrollback is
// unreadable and leaves a mess behind it. The cursor is hidden for the same
// reason: it would sit in the middle of the picture, blinking.
//
// Restored on SIGINT rather than only on a clean return, because the way this
// ends is ctrl-c and a tool that leaves a terminal without its cursor is a
// tool people stop running.
func run(l *Layout, fill fills, t auklet.Theme, start, fps int, theme string) {
	if fps < 1 {
		fps = 1
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	restore := func() { fmt.Print("\x1b[?25h\x1b[?1049l") }
	fmt.Print("\x1b[?1049h\x1b[?25l")
	defer restore()

	// the warnings go out once, before the alternate screen swallows them,
	// rather than every frame
	if _, warns := render(l, fill, t, start); len(warns) > 0 {
		restore()
		for _, w := range warns {
			fmt.Fprintln(os.Stderr, "  "+w)
		}
		fmt.Print("\x1b[?1049h\x1b[?25l")
	}

	tick := time.NewTicker(time.Second / time.Duration(fps))
	defer tick.Stop()
	for frame := start; ; frame++ {
		out, _ := render(l, fill, t, frame)
		// home rather than clear: repainting every cell every frame with no
		// erase in between is what keeps it from flickering
		fmt.Print("\x1b[H" + out)
		fmt.Printf("%s  %dx%d  theme %s  frame %d  ctrl-c to stop", l.Name, l.Cols, l.Rows, theme, frame)
		select {
		case <-sig:
			return
		case <-tick.C:
		}
	}
}
