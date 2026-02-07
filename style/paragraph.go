package style

// Alignment constants for paragraph alignment.
const (
	AlignLeft    = "left"
	AlignCenter  = "center"
	AlignRight   = "right"
	AlignBoth    = "both" // Justified
	AlignDistribute = "distribute"
)

// ParagraphStyle represents paragraph-level formatting.
type ParagraphStyle struct {
	Alignment     string // AlignLeft, AlignCenter, AlignRight, AlignBoth
	SpaceBefore   int    // Space before paragraph in twips
	SpaceAfter    int    // Space after paragraph in twips
	LineSpacing   int    // Line spacing in twips (240 = single)
	LineRule      string // "auto", "exact", "atLeast"
	Indent        int    // Left indent in twips
	IndentRight   int    // Right indent in twips
	FirstLine     int    // First line indent in twips
	Hanging       int    // Hanging indent in twips
	KeepNext      bool   // Keep with next paragraph
	KeepLines     bool   // Keep lines together
	PageBreakBefore bool // Page break before paragraph
	WidowControl  bool   // Widow/orphan control
	TabStops      []TabStop
	Borders       *ParagraphBorders
	Shading       *Shading
	NumStyleName  string // Numbering style name for lists
	NumLevel      int    // Numbering level for lists
}

// TabStop defines a tab stop position and type.
type TabStop struct {
	Position int    // Position in twips
	Type     string // "left", "center", "right", "decimal"
	Leader   string // "none", "dot", "hyphen", "underscore"
}

// ParagraphBorders defines borders around a paragraph.
type ParagraphBorders struct {
	Top    *Border
	Bottom *Border
	Left   *Border
	Right  *Border
}

// Border defines a single border.
type Border struct {
	Style string // "single", "double", "dashed", "dotted", "none"
	Size  int    // Border width in eighths of a point
	Color string // Hex color
	Space int    // Space between border and content in points
}

// Shading defines background shading.
type Shading struct {
	Fill    string // Background fill color (hex)
	Color   string // Pattern color (hex)
	Pattern string // "clear", "solid", "horzStripe", etc.
}

// DefaultParagraphStyle returns a default paragraph style.
func DefaultParagraphStyle() ParagraphStyle {
	return ParagraphStyle{
		Alignment:    AlignLeft,
		WidowControl: true,
	}
}
