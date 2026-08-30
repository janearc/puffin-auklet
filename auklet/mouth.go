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

// MouthTrackFor is MouthTrack fitted to a known duration.
//
// This is the one you want with real speech synthesis. MouthTrack derives its
// length from the TEXT -- how many syllables, how much punctuation -- which is
// a guess at how long the line takes to say. Against actual audio that guess is
// wrong, and a mouth that stops moving two seconds before the voice does is
// worse than one that never moved.
//
// So: ask the synthesiser how long the clip is, pass that here, and the track
// is stretched or compressed to match. The rhythm's PROPORTIONS survive -- the
// pause at the full stop is still the longest thing in the line -- but absolute
// syllable timing does not, and it does not need to. Nobody is reading the beak.
func MouthTrackFor(text string, fps int, seconds float64) []int {
	base := MouthTrack(text, fps)
	if fps < 1 {
		fps = 10
	}
	want := int(seconds*float64(fps) + 0.5)
	if want < 1 || len(base) == 0 {
		return base
	}
	out := make([]int, want)
	for i := range out {
		out[i] = base[i*len(base)/want]
	}
	return out
}
