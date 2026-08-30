package auklet

// a terminal cell can hold more than one pixel. how many depends on which
// block glyphs you are willing to emit:
//
//	Half      1x2   two stacked pixels, the upper-half block. universal.
//	Quadrant  2x2   four pixels. unicode 1.1, universal in practice.
//	Sextant   2x3   six pixels. unicode 13 (2020); needs a font that has them.
//
// more subcells per cell means the same picture fits in fewer cells before it
// starts losing features, which is the whole trick behind drawing small.
//
// every set below is COMPLETE: there is a glyph for every possible arrangement
// of lit subcells, so a cell never has to settle for an approximate shape. it
// only has to settle for two colours, which is the real constraint.
type GlyphSet int

const (
	Half GlyphSet = iota
	Quadrant
	Sextant
)

func (g GlyphSet) String() string {
	switch g {
	case Quadrant:
		return "quadrant"
	case Sextant:
		return "sextant"
	default:
		return "half"
	}
}

// Dims reports how many subcells wide and tall one cell holds.
func (g GlyphSet) Dims() (w, h int) {
	switch g {
	case Quadrant:
		return 2, 2
	case Sextant:
		return 2, 3
	default:
		return 1, 2
	}
}

var (
	halfRunes = [4]rune{' ', '▀', '▄', '█'}
	quadRunes = [16]rune{
		' ', '▘', '▝', '▀', '▖', '▌', '▞', '▛',
		'▗', '▚', '▐', '▜', '▄', '▙', '▟', '█',
	}
)

// Rune maps a bitmask of lit subcells to its glyph. bit 0 is the top-left
// subcell and bits run left to right, then top to bottom.
func (g GlyphSet) Rune(mask int) rune {
	switch g {
	case Quadrant:
		return quadRunes[mask&0xf]
	case Sextant:
		return sextant(mask & 0x3f)
	default:
		return halfRunes[mask&0x3]
	}
}

// the sextant block at U+1FB00 deliberately omits the four arrangements that
// already had characters -- empty, left half, right half, full -- so the run is
// 60 code points long and the index has to skip the two holes in the middle.
func sextant(n int) rune {
	switch n {
	case 0:
		return ' '
	case 21: // bits 0,2,4 -- the left column
		return '▌'
	case 42: // bits 1,3,5 -- the right column
		return '▐'
	case 63:
		return '█'
	}
	idx := n - 1
	if n > 21 {
		idx--
	}
	if n > 42 {
		idx--
	}
	return rune(0x1FB00 + idx)
}
