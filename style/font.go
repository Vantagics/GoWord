// Package style defines styling types for fonts, paragraphs, tables, etc.
package style

// FontStyle represents character-level formatting.
type FontStyle struct {
	Name          string  // Font family name, e.g. "Arial"
	Size          float64 // Font size in points
	Bold          bool
	Italic        bool
	Underline     string // "single", "double", "wave", "dash", "dotted", "none"
	Strikethrough bool
	DoubleStrikethrough bool
	SuperScript   bool
	SubScript     bool
	Color         string // Hex color without '#', e.g. "FF0000"
	HighlightColor string // Highlight color name, e.g. "yellow"
	AllCaps       bool
	SmallCaps     bool
	Hidden        bool
	Spacing       float64 // Character spacing in points
	NoProof       bool
	Lang          string // Language tag, e.g. "en-US"
}

// DefaultFont returns the default font style.
func DefaultFont() FontStyle {
	return FontStyle{
		Name: "Arial",
		Size: 10,
	}
}
