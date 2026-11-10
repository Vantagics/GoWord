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
	Orientation     string // OrientPortrait or OrientLandscape
	PageWidth       int    // Page width in twips
	PageHeight      int    // Page height in twips
	MarginTop       int    // Top margin in twips
	MarginBottom    int    // Bottom margin in twips
	MarginLeft      int    // Left margin in twips
	MarginRight     int    // Right margin in twips
	HeaderHeight    int    // Header distance from top in twips
	FooterHeight    int    // Footer distance from bottom in twips
	ColumnCount     int    // Number of columns
	ColumnSpacing   int    // Space between columns in twips
	PageNumberStart *int   // Starting page number (nil = continue)
	// PageNumberFormat is the w:pgNumType/@w:fmt value, for example "decimal"
	// or "upperRoman". Empty means the writer omits the format.
	PageNumberFormat string
	BreakType        string // "nextPage", "continuous", "evenPage", "oddPage"
	// TitlePg is true when the section has a distinct first-page header/footer.
	TitlePg bool
	// Gutter is the gutter margin in twips (0 = omit).
	Gutter int
	// LinePitch is the document grid line pitch in twips (w:docGrid).
	// Zero means the section has no grid.
	LinePitch int
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
