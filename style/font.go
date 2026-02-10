// Package style defines styling types for fonts, paragraphs, tables, etc.
package style

// FontStyle represents character-level formatting.
type FontStyle struct {
	Name          string  // Font family name, e.g. "Arial"
	NameEastAsia  string  // East Asian font family name
	Size          float64 // Font size in points
	Bold          bool
	Italic        bool
	Underline     string // "single", "double", "wave", "dash", "dotted", "none"
	UnderlineColor string // Underline color (hex), e.g. "FF0000"
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
	Kerning       float64 // Font kerning in points (0 = off)
	NoProof       bool
	Lang          string // Language tag, e.g. "en-US"
	Emboss        bool   // Embossed text effect
	Shadow        bool   // Shadow text effect
	Imprint       bool   // Imprinted (engraved) text effect
	Outline       bool   // Outline text effect
	TextEffect    string // Text animation effect: "blinkBackground", "lights", "antsBlack", "antsRed", "shimmer", "sparkle"
	RightToLeft   bool   // Right-to-left text direction
}

// DefaultFont returns the default font style.
func DefaultFont() FontStyle {
	return FontStyle{
		Name: "Arial",
		Size: 10,
	}
}
