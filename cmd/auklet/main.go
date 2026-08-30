// auklet is a workbench for the puffin sprite: cycle dodo's themes, move
// it, scale it, and toggle the cutout against backdrops that make transparency
// obvious.
//
//	arrows / hjkl     move          shift+arrows  move faster
//	tab / shift+tab   theme         + / -         scale
//	c                 cutout        b             backdrop
//	v                 validator     r             recentre
//	q / esc / ctrl-c  quit
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/janearc/puffin-auklet/puffin"
	"github.com/janearc/puffin-auklet/scene"
	"github.com/janearc/puffin-auklet/themes"
)

const hudRows = 3

type model struct {
	theme    int
	backdrop int
	scale    int
	x, y     int
	cutout   bool
	showVal  bool
	w, h     int
	placed   bool
	quitting bool
}

func (m model) Init() tea.Cmd { return nil }

func (m *model) centre() {
	sw, sh := puffin.SizeAt(m.scale)
	m.x = (m.w - sw) / 2
	m.y = (m.h - hudRows - sh) / 2
	m.clamp()
}

// the sprite may hang off an edge, but never so far that it is entirely gone --
// losing it off-screen with no way back is a bad workbench.
func (m *model) clamp() {
	sw, sh := puffin.SizeAt(m.scale)
	minX, maxX := -sw+4, m.w-4
	minY, maxY := -sh+2, m.h-hudRows-2
	m.x = min(max(m.x, minX), maxX)
	m.y = min(max(m.y, minY), maxY)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		if !m.placed {
			m.centre()
			m.placed = true
		} else {
			m.clamp()
		}
	case tea.KeyMsg:
		step := 1
		k := msg.String()
		if strings.HasPrefix(k, "shift+") {
			step = 5
			k = strings.TrimPrefix(k, "shift+")
		}
		switch k {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "left", "h":
			m.x -= step
		case "right", "l":
			m.x += step
		case "up", "k":
			m.y -= step
		case "down", "j":
			m.y += step
		case "tab", " ":
			m.theme = (m.theme + 1) % len(themes.All)
		case "backtab":
			m.theme = (m.theme - 1 + len(themes.All)) % len(themes.All)
		case "+", "=":
			if m.scale < 4 {
				m.scale++
			}
		case "-", "_":
			if m.scale > 1 {
				m.scale--
			}
		case "c":
			m.cutout = !m.cutout
		case "b":
			m.backdrop = (m.backdrop + 1) % len(scene.BackdropNames)
		case "v":
			m.showVal = !m.showVal
		case "r":
			m.centre()
		}
		m.clamp()
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting || m.w == 0 {
		return ""
	}
	cur := themes.All[m.theme]

	c := scene.Build(scene.Opts{
		Theme: cur, Backdrop: m.backdrop, Scale: m.scale,
		X: m.x, Y: m.y, Cutout: m.cutout,
		W: m.w, H: m.h - hudRows,
	})

	err := cur.Theme.Validate()
	status := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("PASS")
	if err != nil {
		n := strings.Count(err.Error(), "\n") + 1
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).
			Render(fmt.Sprintf("FAIL %d", n))
	}

	cut := "opaque"
	if m.cutout {
		cut = "cutout"
	}
	bold := lipgloss.NewStyle().Bold(true)
	faint := lipgloss.NewStyle().Faint(true)

	hud := fmt.Sprintf("%s  %s  %s  x%d  %s  %s",
		bold.Render(cur.Name), status,
		faint.Render(fmt.Sprintf("%d/%d", m.theme+1, len(themes.All))),
		m.scale, cut, scene.BackdropNames[m.backdrop])

	second := faint.Render("arrows move  tab theme  +/- scale  c cutout  b backdrop  v validator  r recentre  q quit")
	if m.showVal && err != nil {
		second = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).
			Render(strings.ReplaceAll(err.Error(), "\n", "  |  "))
		if len(second) > m.w*3 {
			second = second[:m.w*3]
		}
	}

	return c.String() + "\n" + hud + "\n" + second
}

func main() {
	p := tea.NewProgram(model{scale: 1, cutout: true, backdrop: 2, showVal: true},
		tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
