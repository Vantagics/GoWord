package document

import (
	"errors"
	"testing"

	"github.com/VantageDataChat/GoWord/style"
)

// --- io.go: cover the registered-func paths for OpenFromBytes, Save, ToBytes ---

func TestOpenFromBytesWithRegistration(t *testing.T) {
	origOpenBytes := openFromBytesFunc
	defer func() { openFromBytesFunc = origOpenBytes }()

	openFromBytesFunc = func(data []byte) (*Document, error) {
		return New(), nil
	}
	doc, err := OpenFromBytes([]byte{})
	if err != nil {
		t.Fatal(err)
	}
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
}

func TestSaveWithRegistration(t *testing.T) {
	origSave := saveFunc
	defer func() { saveFunc = origSave }()

	saveFunc = func(d *Document, path string) error {
		return nil
	}
	doc := New()
	if err := doc.Save("test.docx"); err != nil {
		t.Fatal(err)
	}
}

func TestToBytesWithRegistration(t *testing.T) {
	origToBytes := writeToBytesFunc
	defer func() { writeToBytesFunc = origToBytes }()

	writeToBytesFunc = func(d *Document) ([]byte, error) {
		return []byte("test"), nil
	}
	doc := New()
	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test" {
		t.Errorf("data = %q", data)
	}
}

// --- io.go: cover nil-func error paths with errors.Is ---

func TestOpenFromBytesWithoutRegistration(t *testing.T) {
	origOpenBytes := openFromBytesFunc
	defer func() { openFromBytesFunc = origOpenBytes }()

	openFromBytesFunc = nil
	_, err := OpenFromBytes([]byte{})
	if !errors.Is(err, ErrNoReader) {
		t.Errorf("expected ErrNoReader, got %v", err)
	}
}

func TestSaveWithoutRegistration(t *testing.T) {
	origSave := saveFunc
	defer func() { saveFunc = origSave }()

	saveFunc = nil
	doc := New()
	err := doc.Save("test.docx")
	if !errors.Is(err, ErrNoWriter) {
		t.Errorf("expected ErrNoWriter, got %v", err)
	}
}

func TestToBytesWithoutRegistration(t *testing.T) {
	origToBytes := writeToBytesFunc
	defer func() { writeToBytesFunc = origToBytes }()

	writeToBytesFunc = nil
	doc := New()
	_, err := doc.ToBytes()
	if !errors.Is(err, ErrNoWriter) {
		t.Errorf("expected ErrNoWriter, got %v", err)
	}
}

// --- headerfooter.go: cover Header.AddText with both styles ---

func TestHeaderAddTextWithBothStyles(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	h := sec.AddHeader("")
	fs := &style.FontStyle{Bold: true}
	ps := &style.ParagraphStyle{Alignment: style.AlignCenter}
	p := h.AddText("Header", fs, ps)
	if !p.Runs[0].Style.Bold {
		t.Error("font style not applied")
	}
	if p.Style.Alignment != style.AlignCenter {
		t.Error("para style not applied")
	}
}

// --- headerfooter.go: cover Footer.AddImage with style ---

func TestFooterAddImageWithStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	f := sec.AddFooter("")
	is := &style.ImageStyle{Width: 100, Height: 50}
	img := f.AddImage("logo.png", is)
	if img.Style.Width != 100 {
		t.Errorf("width = %d", img.Style.Width)
	}
}

// --- headerfooter.go: cover Footer.AddTable with style ---

func TestFooterAddTableWithStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	f := sec.AddFooter("")
	ts := &style.TableStyle{Width: 5000}
	tbl := f.AddTable(ts)
	if tbl.Style.Width != 5000 {
		t.Errorf("width = %d", tbl.Style.Width)
	}
}

// --- headerfooter.go: cover Footnote.AddTextBreak ---

func TestFootnoteAddTextBreak(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fn := sec.AddFootnote()
	fn.AddText("Line 1", nil)
	fn.AddTextBreak()
	fn.AddText("Line 2", nil)
	if len(fn.Elements) != 3 {
		t.Errorf("elements = %d, want 3", len(fn.Elements))
	}
}

// --- headerfooter.go: cover Endnote.AddTextBreak ---

func TestEndnoteAddTextBreakDirect(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	en := sec.AddEndnote()
	en.AddText("Line 1", nil)
	en.AddTextBreak()
	en.AddText("Line 2", nil)
	if len(en.Elements) != 3 {
		t.Errorf("elements = %d, want 3", len(en.Elements))
	}
}

// --- Endnotes accessor ---

func TestEndnotesAccessor(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddEndnote()
	if len(doc.Endnotes()) != 1 {
		t.Errorf("endnotes = %d", len(doc.Endnotes()))
	}
}

// --- Images accessor ---

func TestImagesAccessor(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddImage("test.png", nil)
	if len(doc.Images()) != 1 {
		t.Errorf("images = %d", len(doc.Images()))
	}
}

// --- Multiple sections ---

func TestMultipleSections(t *testing.T) {
	doc := New()
	s1 := doc.AddSection()
	s1.AddText("Section 1", nil, nil)
	s2 := doc.AddSection()
	s2.AddText("Section 2", nil, nil)
	if len(doc.Sections) != 2 {
		t.Errorf("sections = %d", len(doc.Sections))
	}
}

// --- AddSectionWithStyle ---

func TestAddSectionWithCustomStyle(t *testing.T) {
	doc := New()
	ss := style.SectionStyle{
		Orientation: style.OrientLandscape,
		PageWidth:   16838,
		PageHeight:  11906,
	}
	sec := doc.AddSectionWithStyle(ss)
	if sec.Style.Orientation != style.OrientLandscape {
		t.Errorf("orientation = %q", sec.Style.Orientation)
	}
}

// --- GetFontStyle not found ---

func TestGetFontStyleNotFound(t *testing.T) {
	doc := New()
	_, ok := doc.GetFontStyle("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

// --- GetParagraphStyle not found ---

func TestGetParagraphStyleNotFound(t *testing.T) {
	doc := New()
	_, ok := doc.GetParagraphStyle("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

// --- GetTableStyle not found ---

func TestGetTableStyleNotFound(t *testing.T) {
	doc := New()
	_, ok := doc.GetTableStyle("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

// --- GetNumberingStyle not found ---

func TestGetNumberingStyleNotFound(t *testing.T) {
	doc := New()
	_, ok := doc.GetNumberingStyle("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}
