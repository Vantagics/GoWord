package style

import "testing"

func TestSetPaperSizeAllTypes(t *testing.T) {
	for name, size := range PaperSizes {
		s := DefaultSectionStyle()
		s.SetPaperSize(name)
		if s.PageWidth != size[0] || s.PageHeight != size[1] {
			t.Errorf("%s: got %dx%d, want %dx%d", name, s.PageWidth, s.PageHeight, size[0], size[1])
		}
	}
}

func TestSetPaperSizeLandscapeAllTypes(t *testing.T) {
	for name, size := range PaperSizes {
		s := DefaultSectionStyle()
		s.Orientation = OrientLandscape
		s.SetPaperSize(name)
		if s.PageWidth != size[1] || s.PageHeight != size[0] {
			t.Errorf("%s landscape: got %dx%d, want %dx%d", name, s.PageWidth, s.PageHeight, size[1], size[0])
		}
	}
}

func TestSetPaperSizeUnknown(t *testing.T) {
	s := DefaultSectionStyle()
	origW, origH := s.PageWidth, s.PageHeight
	s.SetPaperSize("Unknown")
	if s.PageWidth != origW || s.PageHeight != origH {
		t.Error("unknown paper size should not change dimensions")
	}
}

func TestSetAllBordersIndependence(t *testing.T) {
	ts := TableStyle{}
	ts.SetAllBorders("single", 6, "FF0000")

	// Mutating one border should not affect others
	ts.BorderTop.Color = "00FF00"
	if ts.BorderBottom.Color != "FF0000" {
		t.Error("mutating BorderTop affected BorderBottom — shared pointer bug")
	}
	if ts.BorderLeft.Color != "FF0000" {
		t.Error("mutating BorderTop affected BorderLeft — shared pointer bug")
	}
}

func TestDistributeAlignment(t *testing.T) {
	if AlignDistribute != "distribute" {
		t.Errorf("AlignDistribute = %q", AlignDistribute)
	}
}

func TestOrientationConstants(t *testing.T) {
	if OrientPortrait != "portrait" {
		t.Errorf("OrientPortrait = %q", OrientPortrait)
	}
	if OrientLandscape != "landscape" {
		t.Errorf("OrientLandscape = %q", OrientLandscape)
	}
}

func TestListStyleConstants(t *testing.T) {
	expected := map[int]int{
		ListBulletFilled:  0,
		ListBulletEmpty:   1,
		ListNumberDecimal: 2,
		ListNumberUpper:   3,
		ListNumberLower:   4,
		ListAlphaUpper:    5,
		ListAlphaLower:    6,
	}
	for got, want := range expected {
		if got != want {
			t.Errorf("list constant %d != %d", got, want)
		}
	}
}

func TestDefaultSectionStyleMargins(t *testing.T) {
	s := DefaultSectionStyle()
	if s.MarginBottom != 1440 {
		t.Errorf("marginBottom = %d", s.MarginBottom)
	}
	if s.MarginLeft != 1440 {
		t.Errorf("marginLeft = %d", s.MarginLeft)
	}
	if s.MarginRight != 1440 {
		t.Errorf("marginRight = %d", s.MarginRight)
	}
	if s.HeaderHeight != 720 {
		t.Errorf("headerHeight = %d", s.HeaderHeight)
	}
	if s.FooterHeight != 720 {
		t.Errorf("footerHeight = %d", s.FooterHeight)
	}
	if s.ColumnCount != 1 {
		t.Errorf("columnCount = %d", s.ColumnCount)
	}
}

func TestImageStyleDefaults(t *testing.T) {
	is := ImageStyle{}
	if is.Width != 0 || is.Height != 0 {
		t.Error("default image style should have zero dimensions")
	}
}

func TestCellStyleDefaults(t *testing.T) {
	cs := CellStyle{}
	if cs.GridSpan != 0 || cs.VMerge != "" || cs.VAlign != "" {
		t.Error("default cell style should have zero values")
	}
}

func TestRowStyleDefaults(t *testing.T) {
	rs := RowStyle{}
	if rs.Height != 0 || rs.IsHeader || rs.CantSplit {
		t.Error("default row style should have zero values")
	}
}
