// bandersnatch is the workbench for the auklet sprite: dodo's themes, arbitrary size,
// cutout transparency, a window to move it around in and out of, and the poses
// it needs to present to camera.
//
//	f                 cycle focus: auklet / window / view
//	arrows, hjkl      move whatever has focus      shift+arrows  by 5
//	+ / -             sprite size                  g             glyph set
//	[ ] { }           window width / height
//	y                 side view / front view
//	t                 narrate                      m             the intro
//	s                 slide out of sight, or back  e             blink now
//	a                 idle animation on/off          o             attention on/off
//	tab               theme                        b             backdrop
//	c                 cutout                       v             validator
//	r                 reset                        q             quit
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/janearc/puffin-auklet/auklet"
	"github.com/janearc/puffin-auklet/scene"
	"github.com/janearc/puffin-auklet/themes"
)

const (
	hudRows = 3
	fps     = 10
)

// the line is only here so there is something to say. any text works.
const sample = "hello. i am a puffin, and i will be narrating this software."

type focus int

const (
	focusSprite focus = iota
	focusWindow
	focusView
)

func (f focus) String() string {
	switch f {
	case focusWindow:
		return "window"
	case focusView:
		return "view"
	default:
		return "auklet"
	}
}

type model struct {
	theme    int
	backdrop int
	glyphs   auklet.GlyphSet
	view     int
	rows     int
	sx, sy   int // sprite, in WORLD coordinates -- never clamped
	restY    int // where the bird sits when it is on stage
	win      scene.Window
	focus    focus
	cutout   bool
	showVal  bool
	w, h     int
	ready    bool
	quitting bool

	animate  bool
	blinking bool
	frame    int
	next     int

	watch  bool  // look at the middle of the window
	hidden bool  // slid away
	track  []int // narration, one mouth level per frame
	pos    int
	roar   int // frame within the intro, or -1
}

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Second/fps, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m model) Init() tea.Cmd { return tick() }

func (m *model) sprite() auklet.Sprite { return auklet.Views()[m.view] }

func (m *model) reset() {
	screenH := m.h - hudRows
	m.win = scene.Window{
		W: max(12, m.w*2/3), H: max(6, screenH-6),
		ScreenX: 3, ScreenY: 2,
	}
	if m.win.W > m.w-6 {
		m.win.W = m.w - 6
	}
	m.rows = min(m.sprite().RowsToFit(m.win.W, m.win.H), 22)
	m.sx = m.win.WorldX + (m.win.W-m.sprite().ColsFor(m.rows))/2
	m.restY = m.win.WorldY + (m.win.H-m.rows)/2
	m.sy = m.restY
	m.hidden, m.roar = false, -1
	m.track, m.pos = nil, 0
}

// the window is kept on screen; the bird is not. that asymmetry is the point.
func (m *model) clampWindow() {
	m.win.W = clamp(m.win.W, 4, m.w-2)
	m.win.H = clamp(m.win.H, 3, m.h-hudRows-2)
	m.win.ScreenX = clamp(m.win.ScreenX, 1, m.w-m.win.W-1)
	m.win.ScreenY = clamp(m.win.ScreenY, 1, m.h-hudRows-m.win.H-1)
}

// offstage is far enough below the window that no part of the bird shows.
func (m model) offstage() int { return m.win.WorldY + m.win.H + 2 }

// glide eases toward a target rather than snapping, because an entrance that
// arrives instantly is not an entrance.
func glide(cur, want int) int {
	d := want - cur
	if d == 0 {
		return cur
	}
	step := d / 3
	if step == 0 {
		if d > 0 {
			step = 1
		} else {
			step = -1
		}
	}
	return cur + step
}

// the intro: rise into frame, then three calls, then settle. the shape is
// stolen wholesale and it is the right shape.
var introScript = []struct {
	until int
	mouth int
}{
	{12, 0}, {19, 3}, {24, 0}, {31, 3}, {36, 0}, {43, 3}, {50, 0},
}

func (m *model) talkOnly() {
	// only the front view has a beak that opens; presenting in profile would
	// mean swinging a mandible away from a head that was never drawn.
	for i, s := range auklet.Views() {
		if s.Name == "front" {
			m.view = i
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.frame++

		if m.animate {
			switch {
			case m.blinking && m.frame >= m.next+2:
				m.blinking = false
				m.next = m.frame + 25 + rand.Intn(35)
			case !m.blinking && m.frame >= m.next:
				m.blinking = true
			}
		}
		if len(m.track) > 0 {
			m.pos++
			if m.pos >= len(m.track) {
				m.track, m.pos = nil, 0
			}
		}
		if m.roar >= 0 {
			m.roar++
			if m.roar > introScript[len(introScript)-1].until {
				m.roar = -1
			}
		}

		want := m.restY
		if m.hidden {
			want = m.offstage()
		}
		m.sy = glide(m.sy, want)
		return m, tick()

	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		if !m.ready {
			m.reset()
			m.ready = true
		}
		m.clampWindow()

	case tea.KeyMsg:
		step := 1
		k := msg.String()
		if strings.HasPrefix(k, "shift+") {
			step = 5
			k = strings.TrimPrefix(k, "shift+")
		}
		dx, dy := 0, 0
		switch k {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "left", "h":
			dx = -step
		case "right", "l":
			dx = step
		case "up", "k":
			dy = -step
		case "down", "j":
			dy = step
		case "f":
			m.focus = (m.focus + 1) % 3
		case "y":
			m.view = (m.view + 1) % len(auklet.Views())
			m.rows = min(m.rows, m.sprite().RowsToFit(m.win.W, m.win.H))
		case "t":
			m.talkOnly()
			m.hidden = false
			m.track, m.pos = auklet.MouthTrack(sample, fps), 0
		case "m":
			m.talkOnly()
			m.hidden = false
			m.sy = m.offstage()
			m.roar = 0
			m.track, m.pos = nil, 0
		case "s":
			m.hidden = !m.hidden
		case "tab":
			m.theme = (m.theme + 1) % len(themes.All)
		case "backtab":
			m.theme = (m.theme - 1 + len(themes.All)) % len(themes.All)
		case "+", "=":
			m.rows++
		case "-", "_":
			m.rows--
		case "g":
			m.glyphs = (m.glyphs + 1) % 3
		case "[":
			m.win.W--
		case "]":
			m.win.W++
		case "{":
			m.win.H--
		case "}":
			m.win.H++
		case "a":
			m.animate = !m.animate
			m.next = m.frame + 5
		case "o":
			m.watch = !m.watch
		case "e":
			m.blinking = !m.blinking
		case "c":
			m.cutout = !m.cutout
		case "b":
			m.backdrop = (m.backdrop + 1) % len(scene.BackdropNames)
		case "v":
			m.showVal = !m.showVal
		case "r":
			m.reset()
		}

		switch m.focus {
		case focusSprite:
			m.sx += dx
			m.restY += dy
			m.sy += dy
		case focusWindow:
			m.win.ScreenX += dx
			m.win.ScreenY += dy
		case focusView:
			m.win.WorldX += dx
			m.win.WorldY += dy
		}

		m.rows = clamp(m.rows, 3, 40)
		m.clampWindow()
	}
	return m, nil
}

// pose is the whole performance, resolved for this frame. the intro outranks
// narration, narration outranks idle blinking, and a shut beak is the default.
func (m model) pose() auklet.Pose {
	sp := m.sprite()
	var p auklet.Pose

	switch {
	case m.roar >= 0:
		for _, step := range introScript {
			if m.roar <= step.until {
				p = append(p, sp.Mouth(step.mouth)...)
				break
			}
		}
	case len(m.track) > 0:
		p = append(p, sp.Mouth(m.track[m.pos])...)
	}

	if m.blinking {
		p = append(p, sp.Blink...)
	}
	return p
}

// what the bird is attending to. in the workbench that is the middle of the
// window, so walking the bird into a corner leaves it still watching the room.
// in an application it would be the selected row, the pane that just changed,
// the thing that went red.
func (m model) attention() (int, int) {
	return m.win.WorldX + m.win.W/2, m.win.WorldY + m.win.H/2
}

func (m model) state() string {
	switch {
	case m.roar >= 0:
		return "intro"
	case len(m.track) > 0:
		return "narrating"
	case m.hidden:
		return "offstage"
	default:
		return "idle"
	}
}

func (m model) baseOpts() scene.Opts {
	return scene.Opts{
		Theme: themes.All[m.theme], Backdrop: m.backdrop,
		Glyphs: m.glyphs, Sprite: m.sprite(), Rows: m.rows,
		SpriteX: m.sx, SpriteY: m.sy, Win: m.win,
		Cutout: m.cutout, W: m.w, H: m.h - hudRows,
	}
}

func (m model) opts() scene.Opts {
	o := m.baseOpts()
	o.Pose = m.pose()
	if m.watch {
		tx, ty := m.attention()
		o.Pose = append(o.Pose, scene.GazeAt(o, tx, ty)...)
	}
	return o
}

func (m model) View() string {
	if m.quitting || !m.ready {
		return ""
	}
	o := m.opts()
	cur := themes.All[m.theme]

	err := cur.Theme.Validate()
	status := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("PASS")
	if err != nil {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).
			Render(fmt.Sprintf("FAIL %d", strings.Count(err.Error(), "\n")+1))
	}

	cut := "opaque"
	if m.cutout {
		cut = "cutout"
	}
	where := ""
	if !scene.Visible(o) {
		where = " OUT OF VIEW"
	}

	bold := lipgloss.NewStyle().Bold(true)
	faint := lipgloss.NewStyle().Faint(true)

	hud := fmt.Sprintf("%s %s  %s  %s %dx%d %s %s  %s%s",
		bold.Render(cur.Name), status,
		bold.Render("["+m.focus.String()+"]"),
		m.sprite().Name, m.sprite().ColsFor(m.rows), m.rows, m.glyphs, cut,
		bold.Render(m.state()), faint.Render(where))

	help := faint.Render("f focus  y view  t narrate  m intro  s slide  o attention  +/- size  g glyphs  tab theme  b backdrop  c cutout  r reset  q quit")
	if m.showVal && err != nil {
		help = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).
			Render(strings.ReplaceAll(err.Error(), "\n", "  |  "))
	}

	// every line is clipped to the terminal's width. reported from the puffin
	// integration: a footer longer than the window silently sets the width of
	// everything composed with it, and a 200-column terminal renders 251
	// columns. the canvas cannot do this to us -- it is sized in cells -- but
	// the HUD is a plain string and would.
	fit := func(s string) string { return ansi.Truncate(s, m.w, "") }
	return scene.Build(o).String() + "\n" + fit(hud) + "\n" + fit(help)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func main() {
	p := tea.NewProgram(model{
		glyphs: auklet.Quadrant, cutout: true, backdrop: 2, showVal: true,
		animate: true, next: 20, roar: -1, watch: true,
	}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
