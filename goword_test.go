package goword

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/style"
)

func TestCreateAndSaveDocument(t *testing.T) {
	doc := New()
	doc.Properties.Title = "Test Document"
	doc.Properties.Creator = "GoWord Test"
	doc.Properties.Subject = "Testing"

	sec := doc.AddSection()
	sec.AddText("Hello World!", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("empty output")
	}

	// Verify it's a valid ZIP
	_, err = zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("invalid ZIP: %v", err)
	}
}

func TestDocxContainsRequiredParts(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Test", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	required := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"word/document.xml",
		"word/styles.xml",
		"word/settings.xml",
		"word/_rels/document.xml.rels",
		"docProps/core.xml",
		"docProps/app.xml",
	}

	files := map[string]bool{}
	for _, f := range zr.File {
		files[f.Name] = true
	}

	for _, name := range required {
		if !files[name] {
			t.Errorf("missing required part: %s", name)
		}
	}
}

func TestDocumentXMLContainsText(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Hello GoWord!", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "Hello GoWord!") {
		t.Error("document.xml does not contain expected text")
	}
}

func TestBoldItalicText(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fs := &style.FontStyle{Bold: true, Italic: true, Size: 14, Color: "FF0000"}
	sec.AddText("Styled", fs, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "<w:b/>") {
		t.Error("missing bold")
	}
	if !strings.Contains(content, "<w:i/>") {
		t.Error("missing italic")
	}
	if !strings.Contains(content, `w:val="FF0000"`) {
		t.Error("missing color")
	}
	if !strings.Contains(content, `w:val="28"`) { // 14pt = 28 half-points
		t.Error("missing font size")
	}
}

func TestParagraphAlignment(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ps := &style.ParagraphStyle{Alignment: style.AlignCenter}
	sec.AddText("Centered", nil, ps)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:jc w:val="center"/>`) {
		t.Error("missing center alignment")
	}
}

func TestTextRun(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Normal ", nil)
	tr.AddText("Bold", &style.FontStyle{Bold: true})

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "Normal ") {
		t.Error("missing normal text")
	}
	if !strings.Contains(content, "Bold") {
		t.Error("missing bold text")
	}
}

func TestHeadings(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddTitle("Heading 1", 1)
	sec.AddTitle("Heading 2", 2)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:val="Heading1"`) {
		t.Error("missing Heading1 style")
	}
	if !strings.Contains(content, `w:val="Heading2"`) {
		t.Error("missing Heading2 style")
	}
}

func TestHyperlink(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddLink("https://example.com", "Click here", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "w:hyperlink") {
		t.Error("missing hyperlink element")
	}
	if !strings.Contains(content, "Click here") {
		t.Error("missing hyperlink text")
	}

	// Check rels
	rels := readZipEntry(t, data, "word/_rels/document.xml.rels")
	if !strings.Contains(rels, "https://example.com") {
		t.Error("missing hyperlink in rels")
	}
}

func TestTable(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ts := &style.TableStyle{}
	ts.SetAllBorders("single", 4, "000000")
	tbl := sec.AddTable(ts)
	row := tbl.AddRow(0, nil)
	c1 := row.AddCell(3000, nil)
	c1.AddText("A", nil, nil)
	c2 := row.AddCell(3000, nil)
	c2.AddText("B", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "<w:tbl>") {
		t.Error("missing table")
	}
	if !strings.Contains(content, "<w:tr>") {
		t.Error("missing table row")
	}
	if !strings.Contains(content, "<w:tc>") {
		t.Error("missing table cell")
	}
}

func TestTableCellSpan(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cs := &style.CellStyle{GridSpan: 3}
	c := row.AddCell(9000, cs)
	c.AddText("Spanning", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:gridSpan w:val="3"/>`) {
		t.Error("missing gridSpan")
	}
}

func TestPageBreak(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Before", nil, nil)
	sec.AddPageBreak()
	sec.AddText("After", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:type="page"`) {
		t.Error("missing page break")
	}
}

func TestTextBreak(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Before", nil, nil)
	sec.AddTextBreak(2)
	sec.AddText("After", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	// Should have empty paragraphs
	if strings.Count(content, "<w:p>") < 4 { // Before + 2 breaks + After
		t.Error("expected at least 4 paragraphs")
	}
}

func TestHeaderFooter(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Body text", nil, nil)

	header := sec.AddHeader("")
	header.AddText("Header text", nil, nil)

	footer := sec.AddFooter("")
	footer.AddPreserveText("Page {PAGE} of {NUMPAGES}", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// Check header file exists
	headerContent := readZipEntry(t, data, "word/header1.xml")
	if !strings.Contains(headerContent, "Header text") {
		t.Error("header missing text")
	}

	// Check footer file exists
	footerContent := readZipEntry(t, data, "word/footer1.xml")
	if !strings.Contains(footerContent, "PAGE") {
		t.Error("footer missing PAGE field")
	}
}

func TestFootnotes(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Text with footnote", nil)
	fn := tr.AddFootnote()
	fn.AddText("This is a footnote.", nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// Check footnotes.xml exists
	fnContent := readZipEntry(t, data, "word/footnotes.xml")
	if !strings.Contains(fnContent, "This is a footnote.") {
		t.Error("footnotes.xml missing text")
	}

	// Check content types
	ct := readZipEntry(t, data, "[Content_Types].xml")
	if !strings.Contains(ct, "footnotes") {
		t.Error("content types missing footnotes")
	}
}

func TestEndnotes(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	en := sec.AddEndnote()
	en.AddText("Endnote text.", nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	enContent := readZipEntry(t, data, "word/endnotes.xml")
	if !strings.Contains(enContent, "Endnote text.") {
		t.Error("endnotes.xml missing text")
	}
}

func TestComments(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddComment("John", "Great work!")

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	cmContent := readZipEntry(t, data, "word/comments.xml")
	if !strings.Contains(cmContent, "Great work!") {
		t.Error("comments.xml missing text")
	}
	if !strings.Contains(cmContent, `w:author="John"`) {
		t.Error("comments.xml missing author")
	}
}

func TestCoreProperties(t *testing.T) {
	doc := New()
	doc.Properties.Title = "My Title"
	doc.Properties.Subject = "My Subject"
	doc.Properties.Creator = "Test Author"
	doc.Properties.Keywords = "go, word, test"
	doc.Properties.Description = "A test document"
	doc.Properties.Category = "Testing"

	sec := doc.AddSection()
	sec.AddText("Content", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	core := readZipEntry(t, data, "docProps/core.xml")
	if !strings.Contains(core, "My Title") {
		t.Error("missing title")
	}
	if !strings.Contains(core, "Test Author") {
		t.Error("missing creator")
	}
	if !strings.Contains(core, "go, word, test") {
		t.Error("missing keywords")
	}
}

func TestStylesXML(t *testing.T) {
	doc := New()
	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultFontSize(11)
	doc.AddFontStyle("myBold", style.FontStyle{Bold: true, Size: 12})

	sec := doc.AddSection()
	sec.AddText("Test", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	styles := readZipEntry(t, data, "word/styles.xml")
	if !strings.Contains(styles, "Calibri") {
		t.Error("missing default font")
	}
	if !strings.Contains(styles, `w:styleId="myBold"`) {
		t.Error("missing custom font style")
	}
	// Check heading styles exist
	if !strings.Contains(styles, `w:styleId="Heading1"`) {
		t.Error("missing Heading1 style")
	}
}

func TestTOC(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddTitle("Chapter 1", 1)
	sec.AddTOC(nil, 1, 3)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "TOC") {
		t.Error("missing TOC field")
	}
}

func TestCheckBox(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddCheckBox("cb1", "Accept terms", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "FORMCHECKBOX") {
		t.Error("missing checkbox field")
	}
	if !strings.Contains(content, "Accept terms") {
		t.Error("missing checkbox text")
	}
}

func TestLandscapeSection(t *testing.T) {
	doc := New()
	ss := style.DefaultSectionStyle()
	ss.Orientation = style.OrientLandscape
	ss.SetPaperSize("A4")
	sec := doc.AddSectionWithStyle(ss)
	sec.AddText("Landscape", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:orient="landscape"`) {
		t.Error("missing landscape orientation")
	}
}

func TestMultipleSections(t *testing.T) {
	doc := New()
	sec1 := doc.AddSection()
	sec1.AddText("Section 1", nil, nil)
	sec2 := doc.AddSection()
	sec2.AddText("Section 2", nil, nil)

	if len(doc.Sections) != 2 {
		t.Errorf("sections = %d, want 2", len(doc.Sections))
	}

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "Section 1") || !strings.Contains(content, "Section 2") {
		t.Error("missing section content")
	}
	// Should have sectPr elements
	if strings.Count(content, "w:sectPr") < 2 {
		t.Error("expected at least 2 sectPr elements")
	}
}

func TestDocumentXMLIsValidXML(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Test & <special> \"chars\"", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")

	// Verify it's valid XML
	decoder := xml.NewDecoder(strings.NewReader(content))
	for {
		_, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("invalid XML: %v", err)
		}
	}
}

func TestSpecialCharactersEscaped(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Tom & Jerry <friends> \"forever\"", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if strings.Contains(content, "Tom & Jerry") {
		t.Error("ampersand not escaped")
	}
	if !strings.Contains(content, "Tom &amp; Jerry") {
		t.Error("missing escaped ampersand")
	}
}

func TestRoundTrip(t *testing.T) {
	// Create a document
	doc := New()
	doc.Properties.Title = "Round Trip Test"
	doc.Properties.Creator = "GoWord"
	sec := doc.AddSection()
	sec.AddText("Hello World", nil, nil)
	sec.AddText("Bold text", &style.FontStyle{Bold: true}, nil)

	// Write to bytes
	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// Read back
	doc2, err := OpenFromBytes(data)
	if err != nil {
		t.Fatalf("OpenFromBytes failed: %v", err)
	}

	if doc2.Properties.Title != "Round Trip Test" {
		t.Errorf("title = %q, want 'Round Trip Test'", doc2.Properties.Title)
	}

	if len(doc2.Sections) == 0 {
		t.Fatal("no sections after round trip")
	}

	// Check we got some paragraphs back
	sec2 := doc2.Sections[0]
	if len(sec2.Elements) < 2 {
		t.Errorf("elements = %d, want >= 2", len(sec2.Elements))
	}
}

func TestEmptyDocument(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	_ = sec // empty section

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Fatal("empty output for empty document")
	}

	// Should still be valid ZIP
	_, err = zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("invalid ZIP: %v", err)
	}
}

func TestNumberingInDocument(t *testing.T) {
	doc := New()
	doc.AddNumberingStyle("bullets", NumberingStyle{
		Type: "singleLevel",
		Levels: []NumberingLevel{
			{Format: "bullet", Text: "\u2022", Left: 360, Hanging: 360, Font: "Symbol"},
		},
	})

	sec := doc.AddSection()
	sec.AddListItem("Item 1", 0, nil, "bullets", nil)
	sec.AddListItem("Item 2", 0, nil, "bullets", nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// Check numbering.xml exists
	numContent := readZipEntry(t, data, "word/numbering.xml")
	if numContent == "" {
		t.Error("numbering.xml not found")
	}
	if !strings.Contains(numContent, "w:abstractNum") {
		t.Error("missing abstractNum")
	}
}

func TestVMerge(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)

	row1 := tbl.AddRow(0, nil)
	c1 := row1.AddCell(3000, &style.CellStyle{VMerge: "restart"})
	c1.AddText("Merged", nil, nil)
	row1.AddCell(3000, nil).AddText("B1", nil, nil)

	row2 := tbl.AddRow(0, nil)
	row2.AddCell(3000, &style.CellStyle{VMerge: "continue"})
	row2.AddCell(3000, nil).AddText("B2", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:vMerge w:val="restart"/>`) {
		t.Error("missing vMerge restart")
	}
	if !strings.Contains(content, `<w:vMerge/>`) {
		t.Error("missing vMerge continue")
	}
}

func TestLine(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddLine(Line{Width: 200, Height: 0, Weight: 2, Color: "FF0000"})

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "v:line") {
		t.Error("missing line element")
	}
}

// Helper to read a zip entry as string.
func readZipEntry(t *testing.T, data []byte, name string) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("invalid ZIP: %v", err)
	}
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("opening %s: %v", name, err)
			}
			defer rc.Close()
			var buf bytes.Buffer
			buf.ReadFrom(rc)
			return buf.String()
		}
	}
	return ""
}
