package document

import (
	"testing"

	"github.com/VantageDataChat/GoWord/style"
)

func TestNewDocument(t *testing.T) {
	doc := New()
	if doc == nil {
		t.Fatal("New() returned nil")
	}
	if doc.DefaultFont.Name != "Arial" {
		t.Errorf("default font name = %q, want Arial", doc.DefaultFont.Name)
	}
	if doc.DefaultFont.Size != 10 {
		t.Errorf("default font size = %f, want 10", doc.DefaultFont.Size)
	}
	if doc.Properties.Creator != "GoWord" {
		t.Errorf("creator = %q, want GoWord", doc.Properties.Creator)
	}
}

func TestAddSection(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	if sec == nil {
		t.Fatal("AddSection returned nil")
	}
	if len(doc.Sections) != 1 {
		t.Errorf("sections count = %d, want 1", len(doc.Sections))
	}
	if sec.Style.PageWidth != 11906 {
		t.Errorf("page width = %d, want 11906 (A4)", sec.Style.PageWidth)
	}
}

func TestSetDefaultFont(t *testing.T) {
	doc := New()
	doc.SetDefaultFontName("Times New Roman")
	doc.SetDefaultFontSize(12)
	if doc.DefaultFont.Name != "Times New Roman" {
		t.Errorf("font name = %q", doc.DefaultFont.Name)
	}
	if doc.DefaultFont.Size != 12 {
		t.Errorf("font size = %f", doc.DefaultFont.Size)
	}
}

func TestNamedStyles(t *testing.T) {
	doc := New()

	fs := style.FontStyle{Name: "Tahoma", Size: 10, Bold: true, Color: "1B2232"}
	doc.AddFontStyle("myFont", fs)

	got, ok := doc.GetFontStyle("myFont")
	if !ok {
		t.Fatal("font style not found")
	}
	if got.Name != "Tahoma" || !got.Bold {
		t.Errorf("font style mismatch: %+v", got)
	}

	ps := style.ParagraphStyle{Alignment: style.AlignCenter}
	doc.AddParagraphStyle("myPara", ps)

	gotP, ok := doc.GetParagraphStyle("myPara")
	if !ok {
		t.Fatal("paragraph style not found")
	}
	if gotP.Alignment != style.AlignCenter {
		t.Errorf("paragraph alignment = %q", gotP.Alignment)
	}

	ts := style.TableStyle{}
	ts.SetAllBorders("single", 6, "006699")
	doc.AddTableStyle("myTable", ts)

	gotT, ok := doc.GetTableStyle("myTable")
	if !ok {
		t.Fatal("table style not found")
	}
	if gotT.BorderTop == nil || gotT.BorderTop.Color != "006699" {
		t.Error("table border not set correctly")
	}
}

func TestAddText(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	p := sec.AddText("Hello World", nil, nil)
	if p == nil {
		t.Fatal("AddText returned nil")
	}
	if len(p.Runs) != 1 {
		t.Fatalf("runs count = %d, want 1", len(p.Runs))
	}
	if p.Runs[0].Text != "Hello World" {
		t.Errorf("text = %q", p.Runs[0].Text)
	}
}

func TestAddTextWithFontStyle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fs := &style.FontStyle{Bold: true, Italic: true, Size: 14}
	p := sec.AddText("Styled text", fs, nil)
	if !p.Runs[0].Style.Bold || !p.Runs[0].Style.Italic {
		t.Error("font style not applied")
	}
}

func TestAddTextRun(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Bold ", &style.FontStyle{Bold: true})
	tr.AddText("Normal", nil)
	if len(tr.Elements) != 2 {
		t.Errorf("textrun elements = %d, want 2", len(tr.Elements))
	}
}

func TestAddTitle(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	p := sec.AddTitle("My Title", 1)
	if p.StyleName != "Heading1" {
		t.Errorf("style name = %q, want Heading1", p.StyleName)
	}

	p2 := sec.AddTitle("Document Title", 0)
	if p2.StyleName != "Title" {
		t.Errorf("style name = %q, want Title", p2.StyleName)
	}
}

func TestAddLink(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	h := sec.AddLink("https://example.com", "Example", nil, nil)
	if h.URL != "https://example.com" {
		t.Errorf("URL = %q", h.URL)
	}
	if h.Text != "Example" {
		t.Errorf("text = %q", h.Text)
	}
}

func TestAddTable(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	c1 := row.AddCell(2000, nil)
	c1.AddText("Cell 1", nil, nil)
	c2 := row.AddCell(2000, nil)
	c2.AddText("Cell 2", nil, nil)

	if len(tbl.Rows) != 1 {
		t.Errorf("rows = %d", len(tbl.Rows))
	}
	if len(tbl.Rows[0].Cells) != 2 {
		t.Errorf("cells = %d", len(tbl.Rows[0].Cells))
	}
}

func TestCellSpan(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cs := &style.CellStyle{GridSpan: 3}
	c := row.AddCell(6000, cs)
	c.AddText("Spanning 3 columns", nil, nil)

	if c.Style.GridSpan != 3 {
		t.Errorf("gridSpan = %d, want 3", c.Style.GridSpan)
	}
}

func TestAddListItem(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	li := sec.AddListItem("Item 1", 0, nil, "", nil)
	if li.Text != "Item 1" {
		t.Errorf("text = %q", li.Text)
	}
}

func TestHeaderFooter(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	h := sec.AddHeader("")
	if h.Type != "default" {
		t.Errorf("header type = %q, want default", h.Type)
	}
	h.AddText("Header text", nil, nil)

	f := sec.AddFooter("")
	f.AddPreserveText("Page {PAGE} of {NUMPAGES}", nil, nil)

	if len(sec.Headers) != 1 {
		t.Errorf("headers = %d", len(sec.Headers))
	}
	if len(sec.Footers) != 1 {
		t.Errorf("footers = %d", len(sec.Footers))
	}
}

func TestFootnoteEndnote(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fn := sec.AddFootnote()
	fn.AddText("Footnote text", nil)
	if fn.ID != 1 {
		t.Errorf("footnote ID = %d, want 1", fn.ID)
	}

	en := sec.AddEndnote()
	en.AddText("Endnote text", nil)
	if en.ID != 1 {
		t.Errorf("endnote ID = %d, want 1", en.ID)
	}
}

func TestAddBookmark(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	bm := sec.AddBookmark("myBookmark")
	if bm.Name != "myBookmark" {
		t.Errorf("bookmark name = %q", bm.Name)
	}
}

func TestAddComment(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	c := sec.AddComment("Author", "This is a comment")
	if c.Author != "Author" {
		t.Errorf("author = %q", c.Author)
	}
	if c.Text != "This is a comment" {
		t.Errorf("text = %q", c.Text)
	}
}

func TestAddCheckBox(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	cb := sec.AddCheckBox("cb1", "Check me", nil, nil)
	if cb.Name != "cb1" {
		t.Errorf("name = %q", cb.Name)
	}
}

func TestAddTOC(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	toc := sec.AddTOC(nil, 1, 3)
	if toc.MinDepth != 1 || toc.MaxDepth != 3 {
		t.Errorf("TOC depth = %d-%d", toc.MinDepth, toc.MaxDepth)
	}
}

func TestNumberingStyle(t *testing.T) {
	doc := New()
	doc.AddNumberingStyle("multilevel", NumberingStyle{
		Type: "multilevel",
		Levels: []NumberingLevel{
			{Format: "decimal", Text: "%1.", Left: 360, Hanging: 360},
			{Format: "upperLetter", Text: "%2.", Left: 720, Hanging: 360},
		},
	})

	ns, ok := doc.GetNumberingStyle("multilevel")
	if !ok {
		t.Fatal("numbering style not found")
	}
	if len(ns.Levels) != 2 {
		t.Errorf("levels = %d, want 2", len(ns.Levels))
	}
}

func TestSectionStyle(t *testing.T) {
	doc := New()
	ss := style.SectionStyle{
		Orientation: style.OrientLandscape,
		PageWidth:   16838,
		PageHeight:  11906,
		MarginTop:   720,
		MarginBottom: 720,
		MarginLeft:  720,
		MarginRight: 720,
	}
	sec := doc.AddSectionWithStyle(ss)
	if sec.Style.Orientation != style.OrientLandscape {
		t.Errorf("orientation = %q", sec.Style.Orientation)
	}
}

func TestTextRunWithLink(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Visit ", nil)
	hl := tr.AddLink("https://go.dev", "Go website", nil)
	tr.AddText(" for more.", nil)

	if hl.URL != "https://go.dev" {
		t.Errorf("URL = %q", hl.URL)
	}
	if len(tr.Elements) != 3 {
		t.Errorf("elements = %d, want 3", len(tr.Elements))
	}
}

func TestCellNestedTable(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cell := row.AddCell(5000, nil)
	nested := cell.AddTable(nil)
	nrow := nested.AddRow(0, nil)
	nc := nrow.AddCell(2500, nil)
	nc.AddText("Nested cell", nil, nil)

	if len(cell.Elements) != 1 {
		t.Errorf("cell elements = %d, want 1", len(cell.Elements))
	}
}
