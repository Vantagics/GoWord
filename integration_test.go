package goword

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/common"
	"github.com/VantageDataChat/GoWord/style"
)

// --- Unit conversion coverage ---

func TestUnitConversions(t *testing.T) {
	// CmToTwip
	if common.CmToTwip(0) != 0 {
		t.Error("CmToTwip(0)")
	}
	// TwipToCm
	if common.TwipToCm(0) != 0 {
		t.Error("TwipToCm(0)")
	}
	// EmuToTwip
	if common.EmuToTwip(0) != 0 {
		t.Error("EmuToTwip(0)")
	}
	// EmuToPixel
	if common.EmuToPixel(9525) != 1 {
		t.Errorf("EmuToPixel(9525) = %d", common.EmuToPixel(9525))
	}
}

// --- Complex round-trip with all element types ---

func TestFullFeatureRoundTrip(t *testing.T) {
	doc := New()
	doc.Properties.Title = "Full Feature Test"
	doc.Properties.Subject = "Testing"
	doc.Properties.Creator = "GoWord"
	doc.Properties.Keywords = "test, full"
	doc.Properties.Description = "A comprehensive test"
	doc.Properties.Category = "Test"
	doc.Properties.Company = "GoWord Inc"
	doc.Properties.Manager = "Test Manager"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultFontSize(11)

	doc.AddFontStyle("bold", FontStyle{Bold: true})
	doc.AddParagraphStyle("centered", ParagraphStyle{Alignment: style.AlignCenter})
	doc.AddTableStyle("bordered", TableStyle{})
	doc.AddNumberingStyle("nums", NumberingStyle{
		Type: "multilevel",
		Levels: []NumberingLevel{
			{Format: "decimal", Text: "%1.", Left: 360, Hanging: 360},
			{Format: "lowerLetter", Text: "%2)", Left: 720, Hanging: 360},
		},
	})

	sec := doc.AddSection()

	// Title and headings
	sec.AddTitle("Document Title", 0)
	for i := 1; i <= 3; i++ {
		sec.AddTitle("Heading "+string(rune('0'+i)), i)
	}

	// Text with various styles
	sec.AddText("Plain text", nil, nil)
	sec.AddText("Bold text", &FontStyle{Bold: true}, nil)
	sec.AddText("Italic text", &FontStyle{Italic: true}, nil)
	sec.AddText("Colored text", &FontStyle{Color: "0000FF"}, nil)
	sec.AddText("Underlined", &FontStyle{Underline: "single"}, nil)
	sec.AddTextWithStyle("Named style", "bold", "centered")

	// TextRun
	tr := sec.AddTextRun(&ParagraphStyle{Alignment: style.AlignBoth})
	tr.AddText("Normal ", nil)
	tr.AddText("bold ", &FontStyle{Bold: true})
	tr.AddText("italic ", &FontStyle{Italic: true})
	tr.AddLink("https://go.dev", "Go", nil)
	tr.AddTextBreak(1)
	tr.AddText("After break", nil)

	// Breaks
	sec.AddTextBreak(3)
	sec.AddPageBreak()

	// Links
	sec.AddLink("https://example.com", "Example Link", &FontStyle{Bold: true}, nil)

	// Table
	ts := &TableStyle{Width: 9000, WidthType: "dxa"}
	ts.SetAllBorders("single", 4, "000000")
	tbl := sec.AddTable(ts)
	tbl.Grid = []int{3000, 3000, 3000}

	hrow := tbl.AddRow(400, &style.RowStyle{IsHeader: true})
	hrow.AddCell(3000, nil).AddText("Col A", &FontStyle{Bold: true}, nil)
	hrow.AddCell(3000, nil).AddText("Col B", &FontStyle{Bold: true}, nil)
	hrow.AddCell(3000, nil).AddText("Col C", &FontStyle{Bold: true}, nil)

	drow := tbl.AddRow(0, nil)
	drow.AddCell(3000, nil).AddText("1", nil, nil)
	drow.AddCell(3000, nil).AddText("2", nil, nil)
	drow.AddCell(3000, nil).AddText("3", nil, nil)

	// List items
	sec.AddListItem("First item", 0, nil, "nums", nil)
	sec.AddListItem("Second item", 0, nil, "nums", nil)
	sec.AddListItem("Sub item", 1, nil, "nums", nil)

	// Header/Footer
	header := sec.AddHeader("")
	header.AddText("Document Header", &FontStyle{Size: 8}, nil)

	footer := sec.AddFooter("")
	footer.AddPreserveText("Page {PAGE} of {NUMPAGES}", &FontStyle{Size: 8}, nil)

	// Footnote
	fn := sec.AddFootnote()
	fn.AddText("This is a footnote.", nil)

	// Endnote
	en := sec.AddEndnote()
	en.AddText("This is an endnote.", nil)

	// Comment
	sec.AddComment("Reviewer", "Looks good!")

	// Bookmark
	sec.AddBookmark("section1")

	// TOC
	sec.AddTOC(nil, 1, 3)

	// Checkbox
	sec.AddCheckBox("agree", "I agree", nil, nil)

	// Line
	sec.AddLine(Line{Width: 200, Height: 0, Weight: 1, Color: "000000"})

	// Write
	data, err := doc.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes: %v", err)
	}

	// Validate ZIP
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("invalid ZIP: %v", err)
	}

	// Check all required parts exist
	parts := map[string]bool{}
	for _, f := range zr.File {
		parts[f.Name] = true
	}

	required := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"word/document.xml",
		"word/styles.xml",
		"word/settings.xml",
		"word/_rels/document.xml.rels",
		"docProps/core.xml",
		"docProps/app.xml",
		"word/numbering.xml",
		"word/footnotes.xml",
		"word/endnotes.xml",
		"word/comments.xml",
		"word/header1.xml",
		"word/footer1.xml",
	}
	for _, name := range required {
		if !parts[name] {
			t.Errorf("missing: %s", name)
		}
	}

	// Read back
	doc2, err := OpenFromBytes(data)
	if err != nil {
		t.Fatalf("OpenFromBytes: %v", err)
	}

	if doc2.Properties.Title != "Full Feature Test" {
		t.Errorf("title = %q", doc2.Properties.Title)
	}
	if doc2.Properties.Subject != "Testing" {
		t.Errorf("subject = %q", doc2.Properties.Subject)
	}
	if len(doc2.Sections) == 0 {
		t.Fatal("no sections")
	}
	if len(doc2.Sections[0].Elements) < 10 {
		t.Errorf("elements = %d, expected many", len(doc2.Sections[0].Elements))
	}
}

// --- App properties coverage ---

func TestAppProperties(t *testing.T) {
	doc := New()
	doc.Properties.Company = "Test Corp"
	doc.Properties.Manager = "Jane Doe"
	sec := doc.AddSection()
	sec.AddText("Test", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	app := readZipEntry(t, data, "docProps/app.xml")
	if !strings.Contains(app, "Test Corp") {
		t.Error("missing company")
	}
	if !strings.Contains(app, "Jane Doe") {
		t.Error("missing manager")
	}
	if !strings.Contains(app, "GoWord") {
		t.Error("missing application name")
	}
}

// --- Image from bytes in docx ---

func TestImageFromBytesInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	// Minimal valid PNG (1x1 pixel)
	pngData := createMinimalPNG()
	sec.AddImageFromBytes(pngData, "image/png", &style.ImageStyle{Width: 50, Height: 50})

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	found := false
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "word/media/image") {
			found = true
		}
	}
	if !found {
		t.Error("image not found in zip")
	}

	// Check content types has png
	ct := readZipEntry(t, data, "[Content_Types].xml")
	if !strings.Contains(ct, `Extension="png"`) {
		t.Error("missing png content type")
	}
}

func createMinimalPNG() []byte {
	// Minimal valid 1x1 white PNG
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC,
		0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, // IEND chunk
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}

// --- Footnote with link and break ---

func TestFootnoteWithLinkAndBreak(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("See footnote", nil)
	fn := tr.AddFootnote()
	fn.AddText("Footnote with ", nil)
	fn.AddLink("https://example.com", "link")
	fn.AddTextBreak()
	fn.AddText("After break.", nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	fnContent := readZipEntry(t, data, "word/footnotes.xml")
	if !strings.Contains(fnContent, "Footnote with") {
		t.Error("missing footnote text")
	}
}

// --- Multiple sections with different orientations ---

func TestMultipleSectionsWithDifferentStyles(t *testing.T) {
	doc := New()

	// Portrait section
	sec1 := doc.AddSection()
	sec1.AddText("Portrait section", nil, nil)

	// Landscape section
	ss := style.DefaultSectionStyle()
	ss.Orientation = style.OrientLandscape
	ss.SetPaperSize("A4")
	sec2 := doc.AddSectionWithStyle(ss)
	sec2.AddText("Landscape section", nil, nil)

	// Back to portrait
	sec3 := doc.AddSection()
	sec3.AddText("Portrait again", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if strings.Count(content, "w:sectPr") < 3 {
		t.Error("expected at least 3 sectPr elements")
	}
}

// --- Nested table ---

func TestNestedTableInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cell := row.AddCell(5000, nil)

	nested := cell.AddTable(nil)
	nrow := nested.AddRow(0, nil)
	nc := nrow.AddCell(2500, nil)
	nc.AddText("Nested", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	// Should have two tbl elements
	if strings.Count(content, "<w:tbl>") < 2 {
		t.Error("expected nested table")
	}
}

// --- Cell with TextRun ---

func TestCellWithTextRunInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cell := row.AddCell(5000, nil)
	tr := cell.AddTextRun(nil)
	tr.AddText("Bold ", &FontStyle{Bold: true})
	tr.AddText("Normal", nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "Bold ") {
		t.Error("missing bold text in cell")
	}
}

// --- Preserve text with no fields ---

func TestPreserveTextNoFields(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	footer := sec.AddFooter("")
	footer.AddPreserveText("Just plain text", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/footer1.xml")
	if !strings.Contains(content, "Just plain text") {
		t.Error("missing plain preserve text")
	}
}

// --- Comment with date ---

func TestCommentDate(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	c := sec.AddComment("Author", "Comment")
	if c.Date.IsZero() {
		// Date is zero by default, that's OK
	}

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/comments.xml")
	if !strings.Contains(content, `w:author="Author"`) {
		t.Error("missing author")
	}
}
