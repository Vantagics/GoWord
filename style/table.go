package style

// TableStyle represents table-level formatting.
type TableStyle struct {
	Width       int    // Table width in twips (0 = auto)
	WidthType   string // "auto", "dxa" (twips), "pct" (percentage)
	Alignment   string // "left", "center", "right"
	CellMarginTop    int // Default cell margin top in twips
	CellMarginBottom int
	CellMarginLeft   int
	CellMarginRight  int
	CellSpacing int    // Cell spacing in twips (0 = none)
	BorderTop    *Border
	BorderBottom *Border
	BorderLeft   *Border
	BorderRight  *Border
	BorderInsideH *Border
	BorderInsideV *Border
	Layout       string // "fixed", "autofit"
	Indent       int    // Table indent from leading margin in twips
}

// RowStyle represents table row formatting.
type RowStyle struct {
	Height     int    // Row height in twips
	HeightRule string // "auto", "exact", "atLeast"
	IsHeader   bool   // Repeat as header row
	CantSplit  bool   // Don't split row across pages
}

// CellStyle represents table cell formatting.
type CellStyle struct {
	Width     int    // Cell width in twips
	WidthType string // "auto", "dxa", "pct"
	GridSpan  int    // Number of columns to span
	VMerge    string // "restart" to start merge, "continue" to continue, "" for none
	VAlign    string // "top", "center", "bottom"
	Shading   *Shading
	BorderTop    *Border
	BorderBottom *Border
	BorderLeft   *Border
	BorderRight  *Border
	TextDirection string // "lrTb", "tbRl", "btLr"
	NoWrap    bool
}

// SetAllBorders is a convenience method to set all borders at once.
// Each border gets its own copy so mutating one doesn't affect the others.
func (ts *TableStyle) SetAllBorders(bstyle string, size int, color string) {
	mk := func() *Border { return &Border{Style: bstyle, Size: size, Color: color} }
	ts.BorderTop = mk()
	ts.BorderBottom = mk()
	ts.BorderLeft = mk()
	ts.BorderRight = mk()
	ts.BorderInsideH = mk()
	ts.BorderInsideV = mk()
}
