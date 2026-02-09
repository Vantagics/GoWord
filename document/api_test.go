package document

import (
	"errors"
	"testing"

	"github.com/VantageDataChat/GoWord/style"
)

// --- io.go error paths ---

func TestOpenWithoutRegistration(t *testing.T) {
	// Save current funcs and restore after test
	origOpen := openFunc
	origOpenBytes := openFromBytesFunc
	origSave := saveFunc
	origToBytes := writeToBytesFunc
	defer func() {
		openFunc = origOpen
		openFromBytesFunc = origOpenBytes
		saveFunc = origSave
		writeToBytesFunc = origToBytes
	}()

	openFunc = nil
	openFromBytesFunc = nil
	saveFunc = nil
	writeToBytesFunc = nil

	_, err := Open("test.docx")
	if err == nil {
		t.Error("expected error from Open without registration")
	}
	if !errors.Is(err, ErrNoReader) {
		t.Errorf("error = %q", err.Error())
	}

	_, err = OpenFromBytes([]byte{})
	if err == nil {
		t.Error("expected error from OpenFromBytes without registration")
	}

	doc := New()
	err = doc.Save("test.docx")
	if err == nil {
		t.Error("expected error from Save without registration")
	}
	if !errors.Is(err, ErrNoWriter) {
		t.Errorf("error = %q", err.Error())
	}

	_, err = doc.ToBytes()
	if err == nil {
		t.Error("expected error from ToBytes without registration")
	}
}

func TestRegisterIO(t *testing.T) {
	called := false
	RegisterIO(
		func(s string) (*Document, error) { called = true; return New(), nil },
		func(b []byte) (*Document, error) { return New(), nil },
		func(d *Document, s string) error { return nil },
		func(d *Document) ([]byte, error) { return []byte{}, nil },
	)
	_, _ = Open("test.docx")
	if !called {
		t.Error("registered open func not called")
	}
}

func TestSentinelErrors(t *testing.T) {
	if ErrNoReader == nil {
		t.Error("ErrNoReader is nil")
	}
	if ErrNoWriter == nil {
		t.Error("ErrNoWriter is nil")
	}
	if ErrNoReader.Error() == "" {
		t.Error("ErrNoReader has empty message")
	}
	if ErrNoWriter.Error() == "" {
		t.Error("ErrNoWriter has empty message")
	}
}

// --- TextRun coverage ---

func TestTextRunAddTextWithStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	run := tr.AddTextWithStyle("styled", "myStyle")
	if run.StyleName != "myStyle" {
		t.Errorf("styleName = %q", run.StyleName)
	}
}

func TestTextRunAddImage(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	img := tr.AddImage("test.png", &style.ImageStyle{Width: 50, Height: 50})
	if img.Source != "test.png" {
		t.Errorf("source = %q", img.Source)
	}
	if len(doc.Images()) != 1 {
		t.Errorf("images = %d", len(doc.Images()))
	}
}

func TestTextRunAddEndnote(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	en := tr.AddEndnote()
	en.AddText("Endnote in textrun", nil)
	if en.ID != 1 {
		t.Errorf("endnote ID = %d", en.ID)
	}
	if len(doc.Endnotes()) != 1 {
		t.Errorf("endnotes = %d", len(doc.Endnotes()))
	}
}

func TestTextRunAddBookmark(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	bm := tr.AddBookmark("trBookmark")
	if bm.Name != "trBookmark" {
		t.Errorf("name = %q", bm.Name)
	}
}

func TestTextRunAddTextBreak(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddTextBreak(0) // should default to 1
	if len(tr.Elements) != 1 {
		t.Errorf("elements = %d", len(tr.Elements))
	}
}

func TestTextRunAddLink(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	hl := tr.AddLink("https://go.dev", "Go", &style.FontStyle{Bold: true})
	if hl.URL != "https://go.dev" {
		t.Errorf("URL = %q", hl.URL)
	}
	if !hl.Font.Bold {
		t.Error("font not bold")
	}
}

// --- Header/Footer coverage ---

func TestHeaderAddImage(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	h := sec.AddHeader("")
	img := h.AddImage("logo.png", &style.ImageStyle{Width: 100, Height: 50})
	if img.Source != "logo.png" {
		t.Errorf("source = %q", img.Source)
	}
}

func TestHeaderAddTable(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	h := sec.AddHeader("")
	tbl := h.AddTable(&style.TableStyle{Width: 5000})
	if tbl.Style.Width != 5000 {
		t.Errorf("width = %d", tbl.Style.Width)
	}
}

func TestHeaderAddPreserveTextWithStyles(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	h := sec.AddHeader("")
	pt := h.AddPreserveText("Page {PAGE}", &style.FontStyle{Bold: true}, &style.ParagraphStyle{Alignment: style.AlignCenter})
	if !pt.Font.Bold {
		t.Error("font not bold")
	}
	if pt.Para.Alignment != style.AlignCenter {
		t.Errorf("alignment = %q", pt.Para.Alignment)
	}
}

func TestFooterAddImage(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	f := sec.AddFooter("")
	img := f.AddImage("footer.png", nil)
	if img.Source != "footer.png" {
		t.Errorf("source = %q", img.Source)
	}
}

func TestFooterAddTable(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	f := sec.AddFooter("")
	tbl := f.AddTable(nil)
	if tbl == nil {
		t.Fatal("table is nil")
	}
}

func TestFooterAddText(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	f := sec.AddFooter("")
	p := f.AddText("Footer text", &style.FontStyle{Size: 8}, &style.ParagraphStyle{Alignment: style.AlignCenter})
	if len(p.Runs) != 1 {
		t.Errorf("runs = %d", len(p.Runs))
	}
	if p.Runs[0].Style.Size != 8 {
		t.Errorf("size = %f", p.Runs[0].Style.Size)
	}
}

func TestFooterAddPreserveText(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	f := sec.AddFooter("")
	pt := f.AddPreserveText("{PAGE}/{NUMPAGES}", &style.FontStyle{Size: 9}, &style.ParagraphStyle{Alignment: style.AlignRight})
	if pt.Text != "{PAGE}/{NUMPAGES}" {
		t.Errorf("text = %q", pt.Text)
	}
}

// --- Footnote/Endnote coverage ---

func TestFootnoteAddLink(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fn := sec.AddFootnote()
	fn.AddText("See ", nil)
	hl := fn.AddLink("https://example.com", "example")
	fn.AddTextBreak()
	fn.AddText("More text", &style.FontStyle{Italic: true})

	if hl.URL != "https://example.com" {
		t.Errorf("URL = %q", hl.URL)
	}
	if len(fn.Elements) != 4 {
		t.Errorf("elements = %d, want 4", len(fn.Elements))
	}
}

func TestEndnoteAddTextWithStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	en := sec.AddEndnote()
	en.AddText("Styled endnote", &style.FontStyle{Bold: true, Color: "FF0000"})
	if len(en.Elements) != 1 {
		t.Errorf("elements = %d", len(en.Elements))
	}
}

func TestEndnoteAddLink(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	en := sec.AddEndnote()
	en.AddText("See ", nil)
	h := en.AddLink("https://example.com", "Example")
	if h.URL != "https://example.com" {
		t.Errorf("url = %q", h.URL)
	}
	if h.Text != "Example" {
		t.Errorf("text = %q", h.Text)
	}
	if len(en.Elements) != 2 {
		t.Errorf("elements = %d, want 2", len(en.Elements))
	}
}

func TestEndnoteAddTextBreak(t *testing.T) {
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

// --- Cell coverage ---

func TestCellAddTextRun(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cell := row.AddCell(3000, nil)
	tr := cell.AddTextRun(&style.ParagraphStyle{Alignment: style.AlignCenter})
	tr.AddText("In cell", nil)
	if len(cell.Elements) != 1 {
		t.Errorf("elements = %d", len(cell.Elements))
	}
}

func TestCellAddImage(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cell := row.AddCell(3000, nil)
	img := cell.AddImage("cell.png", &style.ImageStyle{Width: 80, Height: 60})
	if img.Source != "cell.png" {
		t.Errorf("source = %q", img.Source)
	}
}

func TestCellAddListItem(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cell := row.AddCell(3000, nil)
	li := cell.AddListItem("Cell list item", 0, &style.FontStyle{Bold: true}, "bullets")
	if li.Text != "Cell list item" {
		t.Errorf("text = %q", li.Text)
	}
	if !li.Font.Bold {
		t.Error("font not bold")
	}
}

// --- Section coverage ---

func TestSectionAddImageFromBytes(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	data := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header bytes
	img := sec.AddImageFromBytes(data, "image/png", &style.ImageStyle{Width: 200, Height: 150})
	if img.MimeType != "image/png" {
		t.Errorf("mimeType = %q", img.MimeType)
	}
	if len(img.Data) != 4 {
		t.Errorf("data len = %d", len(img.Data))
	}
}

func TestSectionAddTextWithStyle(t *testing.T) {
	doc := New()
	doc.AddFontStyle("myBold", style.FontStyle{Bold: true})
	doc.AddParagraphStyle("myCentered", style.ParagraphStyle{Alignment: style.AlignCenter})
	sec := doc.AddSection()
	p := sec.AddTextWithStyle("Styled", "myBold", "myCentered")
	if p.StyleName != "myCentered" {
		t.Errorf("styleName = %q", p.StyleName)
	}
	if p.Runs[0].StyleName != "myBold" {
		t.Errorf("run styleName = %q", p.Runs[0].StyleName)
	}
}

func TestSectionAddTableWithStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTableWithStyle("myTableStyle")
	if tbl.StyleName != "myTableStyle" {
		t.Errorf("styleName = %q", tbl.StyleName)
	}
}

func TestSectionAddLinkWithStyles(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	hl := sec.AddLink("https://go.dev", "Go", &style.FontStyle{Bold: true}, &style.ParagraphStyle{Alignment: style.AlignCenter})
	if !hl.Font.Bold {
		t.Error("font not bold")
	}
}

func TestSectionAddListItemWithStyles(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	li := sec.AddListItem("Item", 1, &style.FontStyle{Italic: true}, "numbered", &style.ParagraphStyle{SpaceAfter: 100})
	if li.Depth != 1 {
		t.Errorf("depth = %d", li.Depth)
	}
	if !li.Font.Italic {
		t.Error("font not italic")
	}
	if li.Para.SpaceAfter != 100 {
		t.Errorf("spaceAfter = %d", li.Para.SpaceAfter)
	}
}

func TestSectionAddCheckBoxWithStyles(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	cb := sec.AddCheckBox("cb", "Check", &style.FontStyle{Size: 12}, &style.ParagraphStyle{Alignment: style.AlignLeft})
	if cb.Font.Size != 12 {
		t.Errorf("size = %f", cb.Font.Size)
	}
}

func TestSectionAddTOCWithFont(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	toc := sec.AddTOC(&style.FontStyle{Size: 11}, 1, 6)
	if toc.Font.Size != 11 {
		t.Errorf("size = %f", toc.Font.Size)
	}
}

func TestSectionAddHeaderTypes(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	h1 := sec.AddHeader("first")
	h2 := sec.AddHeader("even")
	if h1.Type != "first" {
		t.Errorf("type = %q", h1.Type)
	}
	if h2.Type != "even" {
		t.Errorf("type = %q", h2.Type)
	}
}

// --- Table row style coverage ---

func TestTableRowWithStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, &style.RowStyle{Height: 500, HeightRule: "exact", IsHeader: true, CantSplit: true})
	if row.Style.Height != 500 {
		t.Errorf("height = %d", row.Style.Height)
	}
	if !row.Style.IsHeader {
		t.Error("not header")
	}
	if !row.Style.CantSplit {
		t.Error("not cantSplit")
	}
}

// --- Element type coverage ---

func TestElementTypes(t *testing.T) {
	tests := []struct {
		elem Element
		want string
	}{
		{&Paragraph{}, "paragraph"},
		{&TextRun{}, "textrun"},
		{&TextBreak{}, "textbreak"},
		{&PageBreak{}, "pagebreak"},
		{&Hyperlink{}, "hyperlink"},
		{&Image{}, "image"},
		{&Table{}, "table"},
		{&ListItem{}, "listitem"},
		{&Footnote{}, "footnote"},
		{&Endnote{}, "endnote"},
		{&Bookmark{}, "bookmark"},
		{&BookmarkEnd{}, "bookmarkend"},
		{&Comment{}, "comment"},
		{&PreserveText{}, "preservetext"},
		{&TOC{}, "toc"},
		{&CheckBox{}, "checkbox"},
		{&Line{}, "line"},
	}
	for _, tt := range tests {
		if got := tt.elem.elementType(); got != tt.want {
			t.Errorf("%T.elementType() = %q, want %q", tt.elem, got, tt.want)
		}
	}
}

// --- Alloc ID coverage ---

func TestAllocIDs(t *testing.T) {
	doc := New()
	r1 := doc.allocRelID()
	r2 := doc.allocRelID()
	if r1 == r2 {
		t.Error("duplicate rel IDs")
	}

	i1 := doc.allocImageID()
	i2 := doc.allocImageID()
	if i1 == i2 {
		t.Error("duplicate image IDs")
	}

	f1 := doc.allocFootnoteID()
	f2 := doc.allocFootnoteID()
	if f1 == f2 {
		t.Error("duplicate footnote IDs")
	}

	e1 := doc.allocEndnoteID()
	e2 := doc.allocEndnoteID()
	if e1 == e2 {
		t.Error("duplicate endnote IDs")
	}

	c1 := doc.allocCommentID()
	c2 := doc.allocCommentID()
	if c1 == c2 {
		t.Error("duplicate comment IDs")
	}

	b1 := doc.allocBookmarkID()
	b2 := doc.allocBookmarkID()
	if b1 == b2 {
		t.Error("duplicate bookmark IDs")
	}
}

func TestItoa(t *testing.T) {
	if itoa(42) != "42" {
		t.Errorf("itoa(42) = %q", itoa(42))
	}
	if itoa(0) != "0" {
		t.Errorf("itoa(0) = %q", itoa(0))
	}
}

// --- Accessor methods coverage ---

func TestFontStylesAccessor(t *testing.T) {
	doc := New()
	doc.AddFontStyle("a", style.FontStyle{Bold: true})
	doc.AddFontStyle("b", style.FontStyle{Italic: true})
	fs := doc.FontStyles()
	if len(fs) != 2 {
		t.Errorf("fontStyles = %d", len(fs))
	}
}

func TestParagraphStylesAccessor(t *testing.T) {
	doc := New()
	doc.AddParagraphStyle("p1", style.ParagraphStyle{Alignment: style.AlignCenter})
	ps := doc.ParagraphStyles()
	if len(ps) != 1 {
		t.Errorf("paragraphStyles = %d", len(ps))
	}
}

func TestTableStylesAccessor(t *testing.T) {
	doc := New()
	doc.AddTableStyle("t1", style.TableStyle{Width: 5000})
	ts := doc.TableStyles()
	if len(ts) != 1 {
		t.Errorf("tableStyles = %d", len(ts))
	}
}

func TestNumberingStylesAccessor(t *testing.T) {
	doc := New()
	doc.AddNumberingStyle("n1", NumberingStyle{Type: "singleLevel"})
	ns := doc.NumberingStyles()
	if len(ns) != 1 {
		t.Errorf("numberingStyles = %d", len(ns))
	}
}

func TestFootnotesAccessor(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddFootnote()
	if len(doc.Footnotes()) != 1 {
		t.Errorf("footnotes = %d", len(doc.Footnotes()))
	}
}

func TestCommentsAccessor(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddComment("Author", "Text")
	if len(doc.Comments()) != 1 {
		t.Errorf("comments = %d", len(doc.Comments()))
	}
}

// --- Section methods that were 0% ---

func TestSectionAddTextBreakDirect(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddTextBreak(3)
	if len(sec.Elements) != 1 {
		t.Errorf("elements = %d", len(sec.Elements))
	}
	tb := sec.Elements[0].(*TextBreak)
	if tb.Count != 3 {
		t.Errorf("count = %d", tb.Count)
	}
}

func TestSectionAddTextBreakMinimum(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddTextBreak(0) // should become 1
	tb := sec.Elements[0].(*TextBreak)
	if tb.Count != 1 {
		t.Errorf("count = %d, want 1", tb.Count)
	}
}

func TestSectionAddPageBreakDirect(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddPageBreak()
	if len(sec.Elements) != 1 {
		t.Errorf("elements = %d", len(sec.Elements))
	}
	if _, ok := sec.Elements[0].(*PageBreak); !ok {
		t.Error("not a PageBreak")
	}
}

func TestSectionAddImageDirect(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	img := sec.AddImage("test.png", &style.ImageStyle{Width: 100, Height: 100})
	if img.Source != "test.png" {
		t.Errorf("source = %q", img.Source)
	}
	if img.Style.Width != 100 {
		t.Errorf("width = %d", img.Style.Width)
	}
	if len(doc.Images()) != 1 {
		t.Errorf("images = %d", len(doc.Images()))
	}
}

func TestSectionAddImageNoStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	img := sec.AddImage("test.jpg", nil)
	if img.Source != "test.jpg" {
		t.Errorf("source = %q", img.Source)
	}
}

func TestSectionAddLineDirect(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	l := sec.AddLine(Line{Width: 100, Height: 0, Weight: 2, Color: "FF0000"})
	if l.Width != 100 {
		t.Errorf("width = %d", l.Width)
	}
	if l.Color != "FF0000" {
		t.Errorf("color = %q", l.Color)
	}
}

func TestTextRunAddFootnoteDirect(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	fn := tr.AddFootnote()
	fn.AddText("Footnote", nil)
	if fn.ID != 1 {
		t.Errorf("ID = %d", fn.ID)
	}
	if len(doc.Footnotes()) != 1 {
		t.Errorf("footnotes = %d", len(doc.Footnotes()))
	}
}

// --- Section AddText/AddTextRun nil style branches ---

func TestSectionAddTextNilFont(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	p := sec.AddText("No font", nil, nil)
	if p.Runs[0].Style.Bold {
		t.Error("should not be bold")
	}
}

func TestSectionAddTextWithFont(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	p := sec.AddText("With font", &style.FontStyle{Bold: true}, &style.ParagraphStyle{Alignment: style.AlignRight})
	if !p.Runs[0].Style.Bold {
		t.Error("should be bold")
	}
	if p.Style.Alignment != style.AlignRight {
		t.Errorf("alignment = %q", p.Style.Alignment)
	}
}

func TestSectionAddTextRunWithStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ps := &style.ParagraphStyle{Alignment: style.AlignCenter}
	tr := sec.AddTextRun(ps)
	if tr.Style.Alignment != style.AlignCenter {
		t.Errorf("alignment = %q", tr.Style.Alignment)
	}
}

func TestSectionAddTableWithTableStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ts := &style.TableStyle{Width: 8000}
	tbl := sec.AddTable(ts)
	if tbl.Style.Width != 8000 {
		t.Errorf("width = %d", tbl.Style.Width)
	}
}

// --- Cell AddText nil branches ---

func TestCellAddTextNilStyles(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cell := row.AddCell(3000, nil)
	p := cell.AddText("Plain", nil, nil)
	if p.Runs[0].Text != "Plain" {
		t.Errorf("text = %q", p.Runs[0].Text)
	}
}

func TestCellAddTextWithStyles(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cell := row.AddCell(3000, nil)
	p := cell.AddText("Styled", &style.FontStyle{Bold: true}, &style.ParagraphStyle{Alignment: style.AlignCenter})
	if !p.Runs[0].Style.Bold {
		t.Error("not bold")
	}
	if p.Style.Alignment != style.AlignCenter {
		t.Errorf("alignment = %q", p.Style.Alignment)
	}
}

// --- Cell AddTable with style ---

func TestCellAddTableWithStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cell := row.AddCell(5000, nil)
	nested := cell.AddTable(&style.TableStyle{Width: 4000})
	if nested.Style.Width != 4000 {
		t.Errorf("width = %d", nested.Style.Width)
	}
}

// --- Header AddText nil branches ---

func TestHeaderAddTextNilStyles(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	h := sec.AddHeader("")
	p := h.AddText("Header", nil, nil)
	if p.Runs[0].Text != "Header" {
		t.Errorf("text = %q", p.Runs[0].Text)
	}
}

// --- Footer AddImage nil style ---

func TestFooterAddImageNilStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	f := sec.AddFooter("")
	img := f.AddImage("img.png", nil)
	if img.Source != "img.png" {
		t.Errorf("source = %q", img.Source)
	}
}

// --- io.go OpenFromBytes/Save/ToBytes error paths ---

func TestOpenFromBytesError(t *testing.T) {
	origFunc := openFromBytesFunc
	defer func() { openFromBytesFunc = origFunc }()

	openFromBytesFunc = nil
	_, err := OpenFromBytes([]byte{})
	if err == nil {
		t.Error("expected error")
	}
}

func TestSaveError(t *testing.T) {
	origFunc := saveFunc
	defer func() { saveFunc = origFunc }()

	saveFunc = nil
	doc := New()
	err := doc.Save("test.docx")
	if err == nil {
		t.Error("expected error")
	}
}

func TestToBytesError(t *testing.T) {
	origFunc := writeToBytesFunc
	defer func() { writeToBytesFunc = origFunc }()

	writeToBytesFunc = nil
	doc := New()
	_, err := doc.ToBytes()
	if err == nil {
		t.Error("expected error")
	}
}
