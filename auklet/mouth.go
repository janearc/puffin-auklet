package auklet

import "strings"

// MouthTrack turns a line of narration into one mouth level per frame.
//
// this is not phoneme-accurate lip sync and does not try to be. it is the same
// trick cheap 2D animation has always used: group the text into rough
// syllables, open on the vowel, tick through the consonants, shut on the
// punctuation. at a puffin's size nobody is reading its beak -- what sells the
// illusion is that the mouth moves WITH the sentence and stops at the full
// stop, and this gets that for forty lines and no audio pipeline.
//
// the output is deterministic: the same text always produces the same track, so
// a recorded take can be reproduced exactly.
func MouthTrack(text string, fps int) []int {
	if fps < 1 {
		fps = 10
	}
	// roughly eight mouth shapes a second is where speech starts reading as
	// speech rather than as chewing.
	unit := fps / 8
	if unit < 1 {
		unit = 1
	}

	var out []int
	hold := func(level, frames int) {
		for i := 0; i < frames; i++ {
			out = append(out, level)
		}
	}

	const vowels = "aeiouy"
	consonants := 0
	flush := func(level int) {
		if consonants > 0 {
			hold(1, unit) // the approach to the vowel
			consonants = 0
		}
		if level >= 0 {
			hold(level, unit+unit/2)
		}
	}

	for _, r := range strings.ToLower(text) {
		switch {
		case r == ' ' || r == '\n' || r == '\t':
			flush(-1)
			hold(0, unit)
		case strings.ContainsRune(".!?", r):
			flush(-1)
			hold(0, unit*4) // a full stop is a real pause; it is most of the rhythm
		case strings.ContainsRune(",;:-", r):
			flush(-1)
			hold(0, unit*2)
		case strings.ContainsRune(vowels, r):
			flush(vowelLevel(r))
		default:
			consonants++
		}
	}
	flush(-1)
	hold(0, unit*2) // settle shut rather than ending mid-syllable
	return out
}

// open vowels get a wider beak than closed ones. it is a coarse rule and it is
// the whole of the "lip sync".
func vowelLevel(r rune) int {
	switch r {
	case 'a', 'o':
		return 3
	case 'e', 'u':
		return 2
	default:
		return 2
	}
}
