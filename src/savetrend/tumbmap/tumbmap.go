package tumbmap

type ThumbInfo struct {
	Size  uint
	Shape string
}

var ThumbByName = map[string]ThumbInfo{
	"XL":       {1080, "thumb"},
	"L":        {750, "thumb"},
	"M_square": {480, "square"},
	"S_square": {306, "square"},
}
