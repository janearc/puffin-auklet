package puffin

// the head-on bird. generated as shapes, then trimmed of its blank margin --
// see the note in sprite.go for why this view exists and what it costs.
//
// 24 by 44 source pixels.
var frontArt = []string{
	".........KKKKKK.........",
	"......KKKKKKKKKKKK......",
	".....KKKKKKKKKKKKKK.....",
	"....KKKKKKKKKKKKKKKK....",
	"...KKKKKKKKKKKKKKKKKK...",
	"..KKKKKKKKKKKKKKKKKKKK..",
	".KKKKWKKKKKKKKKKKKWKKKK.",
	".KKWWWWWKKKKKKKKWWWWWKK.",
	".KWWDDDWWKKKKKKWWDDDWWK.",
	"KWWWXXXWWBBBBBBWWXXXWWWK",
	"KWWXEEEXWBBBBBBWXEEEXWWK",
	"KWWXEEEXWBBBBBWWXEEEXWWK",
	"KWWWXXXWWWBBBBWWWXXXWWWK",
	"KWWWWWWWWWYYYYWWWWWWWWWK",
	"KWWWWWWWWWYYYYWWWWWWWWWK",
	".WWWWWWWWWRRRRWWWWWWWWW.",
	".WWWWWWWWWRRRRWWWWWWWWW.",
	".KWWWWWWWKRRRRKWWWWWWWK.",
	"..KWWWWWKKRRRRKKWWWWWK..",
	"...KKWKKKKRRRRKKKKWKK...",
	"....KKKKKKKRRKKKKKKK....",
	".....KKKKKKRRKKKKKK.....",
	"......KKKKKRRKKKKK......",
	"......KKKKKRRKKKKK......",
	"....KKKKKKKKKKKKKKKK....",
	"...KKKKKKKKKKKKKKKKKK...",
	"..VVKKKKKKKKKKKKKKKKVV..",
	".VVVKKWWWWWWWWWWWWKKVVV.",
	".VVVKWWWWWWWWWWWWWWKVVV.",
	"VVVVWWWWWWWWWWWWWWWWVVVV",
	"VVVVWWWWWWWWWWWWWWWWVVVV",
	"VVVWWWWWWWWWWWWWWWWWWVVV",
	"VVVWWWWWWWWWWWWWWWWWWVVV",
	"VVVWWWWWWWWWWWWWWWWWWVVV",
	"VVVWWWWWWWWWWWWWWWWWWVVV",
	".VVVWWWWWWWWWWWWWWWWVVV.",
	".VVVWWWWWWWWWWWWWWWWVVV.",
	"..VVKWWWWWWWWWWWWWWKVV..",
	"...KKKWWWWWWWWWWWWKKK...",
	"....KKKWWWWWWWWWWKKK....",
	"......KKKWWWWWWKKK......",
	".....OO.KKKKKKKK.OO.....",
	"..OOOOOOOOO...OOOOOOOOO.",
	"...OOOOOOO.....OOOOOOO..",
}

// frontBlink closes both eyes. two patches, not one: the eyes are far apart and
// a single overlay spanning them would repaint the beak between.
var frontBlink = Pose{
	{Name: "blink-l", OX: 3, OY: 9, Art: []string{
		"WWWWW",
		"WWWWW",
		"DDDDD",
		"WWWWW",
	}},
	{Name: "blink-r", OX: 16, OY: 9, Art: []string{
		"WWWWW",
		"WWWWW",
		"DDDDD",
		"WWWWW",
	}},
}
