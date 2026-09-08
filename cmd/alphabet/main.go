// alphabet prints the published role alphabet.
//
// It is the generator for auklet/ALPHABET, which is the file another
// repository vendors and tests its own copy against:
//
//	go run ./cmd/alphabet > auklet/ALPHABET
//
// The file is committed rather than generated on demand so that a consumer
// can vendor a path rather than run a build, and TestAlphabetFileIsCurrent
// fails if the two ever disagree.
package main

import (
	"fmt"

	"github.com/janearc/puffin-auklet/auklet"
)

func main() { fmt.Print(auklet.Alphabet()) }
