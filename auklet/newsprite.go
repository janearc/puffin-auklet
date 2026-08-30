package auklet

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// NewSprite builds a sprite from a role grid.
//
// This is the way in for art that did not ship with the package. The grid is
// validated rather than trusted: a stencil with a stray character or a ragged
// row fails here, loudly, instead of rendering a hole three sizes later.
//
// Options attach poses:
//
//	s, err := auklet.NewSprite("gopher", rows,
//	    auklet.WithBlink(blinkPose),
//	    auklet.WithMouths(shut, ajar, open))
func NewSprite(name string, art []string, opts ...SpriteOption) (Sprite, error) {
	if strings.TrimSpace(name) == "" {
		return Sprite{}, fmt.Errorf("sprite needs a name")
	}
	if len(art) == 0 {
		return Sprite{}, fmt.Errorf("%s: art is empty", name)
	}

	w := len(art[0])
	if w == 0 {
		return Sprite{}, fmt.Errorf("%s: first row is empty", name)
	}
	for y, row := range art {
		if len(row) != w {
			return Sprite{}, fmt.Errorf(
				"%s: row %d is %d wide, row 0 is %d; the grid must be rectangular",
				name, y, len(row), w)
		}
		for x := 0; x < len(row); x++ {
			if !knownRole(row[x]) {
				return Sprite{}, fmt.Errorf(
					"%s: unknown role %q at %d,%d -- see auklet.Roles()",
					name, string(row[x]), x, y)
			}
		}
	}
	// two pixel rows share one terminal cell, so an odd grid loses its last
	// row on render. pad rather than fail: it is always the right fix.
	if len(art)%2 == 1 {
		art = append(append([]string{}, art...), strings.Repeat(string(RoleNone), w))
	}

	s := Sprite{Name: name, art: append([]string{}, art...)}
	for _, o := range opts {
		o(&s)
	}
	s.eyeCache = findEyes(s.art)
	if len(s.Blink) == 0 {
		s.Blink = s.derivedBlink()
	}

	// the stock vocabulary is assembled from whatever poses this sprite turned
	// out to have, and it has to happen for sprites built at RUNTIME too. it
	// once ran only in package init for the two built-ins, which left every
	// emote unavailable to exactly the people the file format exists for.
	// emotes explicitly attached with WithEmote win.
	stock := Sprite{Name: s.Name, art: s.art, Blink: s.Blink, Mouths: s.Mouths, eyeCache: s.eyeCache}
	buildEmotes(&stock)
	for _, n := range stock.emoteOrder {
		if _, taken := s.emotes[n]; !taken {
			WithEmote(stock.emotes[n])(&s)
		}
	}
	return s, nil
}

// A SpriteOption attaches poses to a sprite under construction.
type SpriteOption func(*Sprite)

// WithBlink sets the pose that shuts the eyes.
func WithBlink(p Pose) SpriteOption { return func(s *Sprite) { s.Blink = p } }

// WithMouths sets the mouth shapes, shut first and widening. The first is
// conventionally empty, since the base art is already shut.
func WithMouths(m ...Pose) SpriteOption { return func(s *Sprite) { s.Mouths = m } }

// WithPose attaches a named pose, retrievable with Sprite.NamedPose. Use it for
// anything the package has no opinion about -- a wave, a hat, a sleeping face.
func WithPose(name string, p Pose) SpriteOption {
	return func(s *Sprite) {
		if s.poses == nil {
			s.poses = map[string]Pose{}
		}
		s.poses[name] = p
	}
}

// NamedPose returns a pose attached with WithPose, and whether it existed.
func (s Sprite) NamedPose(name string) (Pose, bool) {
	p, ok := s.poses[name]
	return p, ok
}

// PoseNames lists the named poses, unordered.
func (s Sprite) PoseNames() []string {
	out := make([]string, 0, len(s.poses))
	for k := range s.poses {
		out = append(out, k)
	}
	return out
}

// ParseSprite reads a sprite file.
//
// The format is deliberately a text file rather than Go source, because the
// point of publishing this is that somebody who does not write Go can draw a
// bird, put it in a gist, and have it work. It is line-oriented, diffable, and
// you can edit it in anything:
//
//	# a gopher. lines starting with # are ignored.
//	sprite gopher
//
//	art
//	....KKKK....
//	...KWWWWK...
//
//	pose blink
//	at 9,11
//	WWWWW
//	DDDDD
//	at 22,11
//	WWWWW
//	DDDDD
//
//	pose mouth1
//	at 10,17
//	KKKK
//
// A `pose` may hold several `at` patches, which is how a face with two eyes
// blinks without the overlay between them repainting the nose. Poses named
// `blink` and `mouth0`..`mouth9` are wired to Blink and Mouths; anything else
// is available through NamedPose.
//
// Coordinates are in source pixels, relative to the top-left of `art`.
func ParseSprite(r io.Reader) (Sprite, error) {
	var (
		name    string
		art     []string
		poses   = map[string]Pose{}
		order   []string
		section string // "art", or the pose name
		patch   *Overlay
		lineNo  int
	)

	flushPatch := func() {
		if patch != nil && len(patch.Art) > 0 {
			poses[section] = append(poses[section], *patch)
		}
		patch = nil
	}

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lineNo++
		line := strings.TrimRight(sc.Text(), " \t\r")
		trimmed := strings.TrimSpace(line)

		switch {
		case trimmed == "" || strings.HasPrefix(trimmed, "#"):
			continue

		case strings.HasPrefix(trimmed, "sprite "):
			flushPatch()
			name = strings.TrimSpace(strings.TrimPrefix(trimmed, "sprite "))
			section = ""

		case trimmed == "art":
			flushPatch()
			section = "art"

		case strings.HasPrefix(trimmed, "pose "):
			flushPatch()
			section = strings.TrimSpace(strings.TrimPrefix(trimmed, "pose "))
			if section == "" {
				return Sprite{}, fmt.Errorf("line %d: pose needs a name", lineNo)
			}
			if _, seen := poses[section]; !seen {
				order = append(order, section)
				poses[section] = nil
			}

		case strings.HasPrefix(trimmed, "at "):
			if section == "" || section == "art" {
				return Sprite{}, fmt.Errorf("line %d: `at` outside a pose", lineNo)
			}
			flushPatch()
			x, y, err := parseAt(strings.TrimPrefix(trimmed, "at "))
			if err != nil {
				return Sprite{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			patch = &Overlay{Name: section, OX: x, OY: y}

		default:
			// a row of role codes
			switch {
			case section == "art":
				art = append(art, trimmed)
			case patch != nil:
				patch.Art = append(patch.Art, trimmed)
			case section == "":
				return Sprite{}, fmt.Errorf(
					"line %d: art row before any `art` or `pose` heading", lineNo)
			default:
				return Sprite{}, fmt.Errorf(
					"line %d: rows in pose %q must follow an `at x,y`", lineNo, section)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return Sprite{}, err
	}
	flushPatch()

	if len(art) == 0 {
		return Sprite{}, fmt.Errorf("%s: no `art` section", name)
	}

	var opts []SpriteOption
	if p, ok := poses["blink"]; ok {
		opts = append(opts, WithBlink(p))
	}
	var mouths []Pose
	for i := 0; i < 10; i++ {
		p, ok := poses[fmt.Sprintf("mouth%d", i)]
		if !ok {
			break
		}
		mouths = append(mouths, p)
	}
	if len(mouths) > 0 {
		opts = append(opts, WithMouths(mouths...))
	}
	for _, n := range order {
		if n == "blink" || (strings.HasPrefix(n, "mouth") && len(n) == 6) {
			continue
		}
		opts = append(opts, WithPose(n, poses[n]))
	}

	return NewSprite(name, art, opts...)
}

func parseAt(s string) (x, y int, err error) {
	f := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' })
	if len(f) != 2 {
		return 0, 0, fmt.Errorf("want `at x,y`, got %q", strings.TrimSpace(s))
	}
	if x, err = strconv.Atoi(f[0]); err != nil {
		return 0, 0, fmt.Errorf("bad x in `at`: %w", err)
	}
	if y, err = strconv.Atoi(f[1]); err != nil {
		return 0, 0, fmt.Errorf("bad y in `at`: %w", err)
	}
	return x, y, nil
}

// WriteSprite writes a sprite in the ParseSprite format. Round-tripping is
// exact for the art and for blink and mouth poses.
func WriteSprite(w io.Writer, s Sprite) error {
	b := &strings.Builder{}
	fmt.Fprintf(b, "sprite %s\n\nart\n", s.Name)
	for _, row := range s.art {
		fmt.Fprintf(b, "%s\n", row)
	}

	// an empty pose still needs its heading: mouth0 is empty BY DESIGN, since
	// the base art is already shut, and dropping the heading silently drops the
	// whole mouth series on the way back in.
	writePose := func(name string, p Pose, always bool) {
		if len(p) == 0 && !always {
			return
		}
		fmt.Fprintf(b, "\npose %s\n", name)
		for _, o := range p {
			fmt.Fprintf(b, "at %d,%d\n", o.OX, o.OY)
			for _, row := range o.Art {
				fmt.Fprintf(b, "%s\n", row)
			}
		}
	}
	writePose("blink", s.Blink, false)
	for i, m := range s.Mouths {
		writePose(fmt.Sprintf("mouth%d", i), m, true)
	}
	for _, n := range s.PoseNames() {
		p, _ := s.NamedPose(n)
		writePose(n, p, false)
	}

	_, err := io.WriteString(w, b.String())
	return err
}
