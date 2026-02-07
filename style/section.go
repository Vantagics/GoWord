package style

// Orientation constants.
const (
	OrientPortrait  = "portrait"
	OrientLandscape = "landscape"
)

// PaperSize constants in twips (width x height).
var PaperSizes = map[string][2]int{
	"A4":     {11906, 16838},
	"A3":     {16838, 23811},
	"A5":     {8391, 11906},
	"Letter": {12240, 15840},
	"Legal":  {12240, 20160},
}

// SectionStyle represents section-level formatting.
type SectionStyle struct {
	Orientation  string // OrientPortrait or OrientLandscape
	PageWidth    int    // Page width in twips
	PageHeight   int    // Page height in twips
	MarginTop    int    // Top margin in twips
	MarginBottom int    // Bottom margin in twips
	MarginLeft   int    // Left margin in twips
	MarginRight  int    // Right margin in twips
	HeaderHeight int    // Header distance from top in twips
	FooterHeight int    // Footer distance from bottom in twips
	ColumnCount  int    // Number of columns
	ColumnSpacing int   // Space between columns in twips
	PageNumberStart *int // Starting page number (nil = continue)
	BreakType    string // "nextPage", "continuous", "evenPage", "oddPage"
}

// DefaultSectionStyle returns A4 portrait with standard margins.
func DefaultSectionStyle() SectionStyle {
	return SectionStyle{
		Orientation:  OrientPortrait,
		PageWidth:    11906, // A4
		PageHeight:   16838,
		MarginTop:    1440, // 1 inch
		MarginBottom: 1440,
		MarginLeft:   1440,
		MarginRight:  1440,
		HeaderHeight: 720,
		FooterHeight: 720,
		ColumnCount:  1,
	}
}

// SetPaperSize sets the page dimensions from a named paper size.
func (s *SectionStyle) SetPaperSize(name string) {
	if size, ok := PaperSizes[name]; ok {
		if s.Orientation == OrientLandscape {
			s.PageWidth = size[1]
			s.PageHeight = size[0]
		} else {
			s.PageWidth = size[0]
			s.PageHeight = size[1]
		}
	}
}
