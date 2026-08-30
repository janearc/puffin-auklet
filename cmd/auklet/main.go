// auklet is a workbench for the puffin sprite: dodo's themes, arbitrary size,
// cutout transparency, and a window to move it around in and out of.
//
//	f                 cycle focus: auklet / window / view
//	a                 animation on/off             e             blink now
//	arrows, hjkl      move whatever has focus      shift+arrows  by 5
//	+ / -             sprite size                  g             glyph set
//	[ ] { }           window width / height
//	tab               theme                        b             backdrop
//	c                 cutout                       v             validator
//	r                 reset                        q             quit
//
// the auklet's position is in world space and is never clamped. move it off the
// window and it keeps its coordinates; the border grows a pointer showing which
// way it went.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/janearc/puffin-auklet/puffin"
	"github.com/janearc/puffin-auklet/scene"
	"github.com/janearc/puffin-auklet/themes"
)

const hudRows = 3

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
	glyphs   puffin.GlyphSet
	rows     int
	sx, sy   int // sprite, in WORLD coordinates -- never clamped
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
}

type tickMsg struct{}

// 100ms is slow enough to cost nothing and fast enough that a two-frame blink
// looks like a blink rather than a glitch.
func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m model) Init() tea.Cmd { return tick() }

func (m *model) reset() {
	screenH := m.h - hudRows
	m.win = scene.Window{
		W: max(12, m.w*2/3), H: max(6, screenH-6),
		ScreenX: 3, ScreenY: 2,
	}
	if m.win.W > m.w-6 {
		m.win.W = m.w - 6
	}
	m.rows = min(puffin.RowsToFit(m.win.W, m.win.H), 22)
	m.sx = m.win.WorldX + (m.win.W-puffin.ColsFor(m.rows))/2
	m.sy = m.win.WorldY + (m.win.H-m.rows)/2
}

// the window is kept on screen; the bird is not. that asymmetry is the point.
func (m *model) clampWindow() {
	m.win.W = clamp(m.win.W, 4, m.w-2)
	m.win.H = clamp(m.win.H, 3, m.h-hudRows-2)
	m.win.ScreenX = clamp(m.win.ScreenX, 1, m.w-m.win.W-1)
	m.win.ScreenY = clamp(m.win.ScreenY, 1, m.h-hudRows-m.win.H-1)
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
		case "c":
			m.cutout = !m.cutout
		case "b":
			m.backdrop = (m.backdrop + 1) % len(scene.BackdropNames)
		case "a":
			m.animate = !m.animate
			m.next = m.frame + 5
		case "e":
			m.blinking = !m.blinking
		case "v":
			m.showVal = !m.showVal
		case "r":
			m.reset()
		}

		switch m.focus {
		case focusSprite:
			m.sx += dx
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

func (m model) opts() scene.Opts {
	return scene.Opts{
		Theme: themes.All[m.theme], Backdrop: m.backdrop,
		Glyphs: m.glyphs, Rows: m.rows,
		SpriteX: m.sx, SpriteY: m.sy, Win: m.win,
		Cutout: m.cutout, W: m.w, H: m.h - hudRows,
		Pose: m.pose(),
	}
}

func (m model) pose() puffin.Pose {
	if m.blinking {
		return puffin.Pose{puffin.Blink}
	}
	return nil
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
	where := "in view"
	if !scene.Visible(o) {
		where = "OUT OF VIEW"
	}

	bold := lipgloss.NewStyle().Bold(true)
	faint := lipgloss.NewStyle().Faint(true)

	hud := fmt.Sprintf("%s %s  %s  %dx%d %s %s  auklet@%d,%d %s  view@%d,%d  win %dx%d",
		bold.Render(cur.Name), status,
		bold.Render("["+m.focus.String()+"]"),
		puffin.ColsFor(m.rows), m.rows, m.glyphs, cut,
		m.sx, m.sy, faint.Render(where),
		m.win.WorldX, m.win.WorldY, m.win.W, m.win.H)

	help := faint.Render("f focus  arrows move  +/- size  g glyphs  [ ] { } window  tab theme  b backdrop  c cutout  a animate  r reset  q quit")
	if m.showVal && err != nil {
		help = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).
			Render(strings.ReplaceAll(err.Error(), "\n", "  |  "))
	}

	return scene.Build(o).String() + "\n" + hud + "\n" + help
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
		glyphs: puffin.Quadrant, cutout: true, backdrop: 2, showVal: true,
		animate: true, next: 20,
	}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
