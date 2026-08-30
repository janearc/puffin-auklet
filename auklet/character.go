package auklet

// A CharacterSet names one character's turn views. Views() and GopherViews()
// existed first, each hardcoded to its own bird; this is the seam that lets
// a caller -- the workbench, a renderer -- pick between characters without
// hardcoding either one by name.
type CharacterSet struct {
	Name  string
	Views []Sprite
}

// Characters lists every character this package ships. Registering a third
// one is adding a line here, not touching the two that already exist.
func Characters() []CharacterSet {
	return []CharacterSet{
		{Name: "auklet", Views: Views()},
		{Name: "gopher", Views: GopherViews()},
		{Name: "lamp", Views: LampViews()},
	}
}
