package style

import "testing"

func TestDefaultFont(t *testing.T) {
	f := DefaultFont()
	if f.Name != "Arial" {
		t.Errorf("name = %q, want Arial", f.Name)
	}
	if f.Size != 10 {
		t.Errorf("size = %f, want 10", f.Size)
	}
}

func TestDefaultParagraphStyle(t *testing.T) {
	p := DefaultParagraphStyle()
	if p.Alignment != AlignLeft {
		t.Errorf("alignment = %q, want left", p.Alignment)
	}
	if !p.WidowControl {
		t.Error("widowControl should be true by default")
	}
}

func TestDefaultSectionStyle(t *testing.T) {
	s := DefaultSectionStyle()
	if s.Orientation != OrientPortrait {
		t.Errorf("orientation = %q", s.Orientation)
	}
	if s.PageWidth != 11906 {
		t.Errorf("pageWidth = %d", s.PageWidth)
	}
	if s.MarginTop != 1440 {
		t.Errorf("marginTop = %d", s.MarginTop)
	}
}

func TestSetPaperSize(t *testing.T) {
	s := DefaultSectionStyle()
	s.SetPaperSize("Letter")
	if s.PageWidth != 12240 {
		t.Errorf("Letter width = %d, want 12240", s.PageWidth)
	}
	if s.PageHeight != 15840 {
		t.Errorf("Letter height = %d, want 15840", s.PageHeight)
	}

	s.Orientation = OrientLandscape
	s.SetPaperSize("Letter")
	if s.PageWidth != 15840 {
		t.Errorf("Landscape Letter width = %d, want 15840", s.PageWidth)
	}
}

func TestTableStyleSetAllBorders(t *testing.T) {
	ts := TableStyle{}
	ts.SetAllBorders("single", 6, "FF0000")

	borders := []*Border{ts.BorderTop, ts.BorderBottom, ts.BorderLeft, ts.BorderRight, ts.BorderInsideH, ts.BorderInsideV}
	for i, b := range borders {
		if b == nil {
			t.Errorf("border %d is nil", i)
			continue
		}
		if b.Style != "single" || b.Size != 6 || b.Color != "FF0000" {
			t.Errorf("border %d = %+v", i, b)
		}
	}
}

func TestAlignmentConstants(t *testing.T) {
	if AlignLeft != "left" {
		t.Error("AlignLeft")
	}
	if AlignCenter != "center" {
		t.Error("AlignCenter")
	}
	if AlignRight != "right" {
		t.Error("AlignRight")
	}
	if AlignBoth != "both" {
		t.Error("AlignBoth")
	}
}

func TestListConstants(t *testing.T) {
	if ListBulletFilled != 0 {
		t.Error("ListBulletFilled")
	}
	if ListNumberDecimal != 2 {
		t.Error("ListNumberDecimal")
	}
}
