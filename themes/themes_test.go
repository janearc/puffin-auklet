package themes

import (
	"strings"
	"testing"
)

// every dodo theme must render a legible bird. this is the gate that stops a
// theme change in dodo silently breaking the sprite.
func TestDodoThemesValidate(t *testing.T) {
	for _, name := range DodoOrder {
		if err := Dodo[name].Auklet().Validate(); err != nil {
			t.Errorf("%s:\n  %s", name, strings.ReplaceAll(err.Error(), "\n", "\n  "))
		}
	}
}
