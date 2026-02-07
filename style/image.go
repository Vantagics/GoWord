package style

// ImageStyle represents image formatting.
type ImageStyle struct {
	Width         int    // Width in points
	Height        int    // Height in points
	Alignment     string // "left", "center", "right"
	WrappingStyle string // "inline", "behind", "inFrontOf", "square", "tight"
	MarginTop     int    // Margin top in points
	MarginBottom  int    // Margin bottom in points
	MarginLeft    int    // Margin left in points
	MarginRight   int    // Margin right in points
}

// ListStyle represents list item formatting.
type ListStyle struct {
	Type  int // List type constant
	Level int // Nesting level (0-based)
}

// List type constants.
const (
	ListBulletFilled  = 0
	ListBulletEmpty   = 1
	ListNumberDecimal = 2
	ListNumberUpper   = 3
	ListNumberLower   = 4
	ListAlphaUpper    = 5
	ListAlphaLower    = 6
)
