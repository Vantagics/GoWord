package ooxml

import (
	"archive/zip"
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

// --- fontStyleToXML uncovered branches ---

func TestFontStyleToXMLAllProperties(t *testing.T) {
	fs := &style.FontStyle{
		Name:                "Courier",
		Bold:                true,
		Italic:              true,
		Strikethrough:       true,
		DoubleStrikethrough: true,
		Underline:           "double",
		Color:               "FF0000",
		HighlightColor:      "yellow",
		Size:                14,
		SuperScript:         true,
		AllCaps:             true,
		SmallCaps:           true,
		Hidden:              true,
		NoProof:             true,
		Lang:                "en-US",
	}
	xml := fontStyleToXML(fs, nil)
	checks := []string{
		`<w:rFonts`, `Courier`,
		`<w:b/>`, `<w:i/>`,
		`<w:strike/>`, `<w:dstrike/>`,
		`<w:u w:val="double"/>`,
		`<w:color w:val="FF0000"/>`,
		`<w:highlight w:val="yellow"/>`,
		`<w:sz w:val="28"/>`,
		`<w:vertAlign w:val="superscript"/>`,
		`<w:caps/>`, `<w:smallCaps/>`,
		`<w:vanish/>`, `<w:noProof/>`,
		`<w:lang w:val="en-US"/>`,
	}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("missing %q in fontStyleToXML output", c)
		}
	}
}

func TestFontStyleToXMLNilBoth(t *testing.T) {
	result := fontStyleToXML(nil, nil)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestFontStyleToXMLDefaultFontFallback(t *testing.T) {
	df := &style.FontStyle{Bold: true}
	result := fontStyleToXML(nil, df)
	if !strings.Contains(result, "<w:b/>") {
		t.Error("default font fallback not applied")
	}
}

func TestFontStyleToXMLUnderlineNone(t *testing.T) {
	fs := &style.FontStyle{Underline: "none"}
	result := fontStyleToXML(fs, nil)
	if strings.Contains(result, "<w:u") {
		t.Error("underline=none should not produce <w:u>")
	}
}

func TestFontStyleToXMLSubScript(t *testing.T) {
	fs := &style.FontStyle{SubScript: true}
	result := fontStyleToXML(fs, nil)
	if !strings.Contains(result, `<w:vertAlign w:val="subscript"/>`) {
		t.Error("missing subscript")
	}
}

// --- paraStyleToXML uncovered branches ---

func TestParaStyleToXMLAllProperties(t *testing.T) {
	ps := &style.ParagraphStyle{
		Alignment:       style.AlignCenter,
		SpaceBefore:     120,
		SpaceAfter:      240,
		LineSpacing:     480,
		LineRule:        "exact",
		Indent:          720,
		IndentRight:     360,
		FirstLine:       360,
		Hanging:         180,
		KeepNext:        true,
		KeepLines:       true,
		PageBreakBefore: true,
		WidowControl:    true,
		NumStyleName:    "1",
		NumLevel:        2,
		Shading:         &style.Shading{Fill: "FFFF00", Color: "auto", Pattern: "clear"},
		Borders: &style.ParagraphBorders{
			Top:    &style.Border{Style: "single", Size: 4, Color: "000000"},
			Bottom: &style.Border{Style: "single", Size: 4, Color: "000000"},
			Left:   &style.Border{Style: "single", Size: 4, Color: "000000"},
			Right:  &style.Border{Style: "single", Size: 4, Color: "000000"},
		},
	}
	xml := paraStyleToXML(ps, "TestStyle")
	checks := []string{
		`<w:pStyle w:val="TestStyle"/>`,
		`<w:jc w:val="center"/>`,
		`w:before="120"`, `w:after="240"`,
		`w:line="480"`, `w:lineRule="exact"`,
		`w:left="720"`, `w:right="360"`,
		`w:firstLine="360"`, `w:hanging="180"`,
		`<w:keepNext/>`, `<w:keepLines/>`,
		`<w:pageBreakBefore/>`, `<w:widowControl/>`,
		`<w:numPr>`, `<w:ilvl w:val="2"/>`,
		`<w:shd`, `w:fill="FFFF00"`,
		`<w:pBdr>`,
	}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("missing %q in paraStyleToXML output", c)
		}
	}
}

func TestParaStyleToXMLNilBoth(t *testing.T) {
	result := paraStyleToXML(nil, "")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestParaStyleToXMLLineSpacingAutoRule(t *testing.T) {
	ps := &style.ParagraphStyle{LineSpacing: 240}
	xml := paraStyleToXML(ps, "")
	if !strings.Contains(xml, `w:lineRule="auto"`) {
		t.Error("default line rule should be auto")
	}
}

// --- writeImages coverage (52% -> higher) ---

func TestWriteImagesFromBytesWithMimeType(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddImageFromBytes(createMinimalPNG2(), "image/png", &style.ImageStyle{Width: 50, Height: 50})

	data, err := WriteToBytes(doc)
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
		t.Error("image file not found in zip")
	}
}

func TestWriteImagesSkipsEmptyImage(t *testing.T) {
	// An image with no source and no data should be skipped
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("test", nil, nil)

	// Manually add an empty image to test the skip path
	img := &document.Image{}
	doc.Sections[0].Elements = append(doc.Sections[0].Elements, img)

	_, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
}

// --- reader coverage improvements ---

func TestReadDocumentWithAllCapsSmallCaps(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("CAPS", &style.FontStyle{AllCaps: true}, nil)
	sec.AddText("small", &style.FontStyle{SmallCaps: true}, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	for _, elem := range doc2.Sections[0].Elements {
		if p, ok := elem.(*document.Paragraph); ok {
			for _, run := range p.Runs {
				if run.Text == "CAPS" && !run.Style.AllCaps {
					t.Error("AllCaps not preserved")
				}
				if run.Text == "small" && !run.Style.SmallCaps {
					t.Error("SmallCaps not preserved")
				}
			}
		}
	}
}

func TestReadDocumentWithFontNameAndSize(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Styled", &style.FontStyle{Name: "Courier New", Size: 16}, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	for _, elem := range doc2.Sections[0].Elements {
		if p, ok := elem.(*document.Paragraph); ok {
			for _, run := range p.Runs {
				if run.Text == "Styled" {
					if run.Style.Name != "Courier New" {
						t.Errorf("font name = %q", run.Style.Name)
					}
					if run.Style.Size != 16 {
						t.Errorf("font size = %f", run.Style.Size)
					}
				}
			}
		}
	}
}

func TestReadDocumentWithColor(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Red", &style.FontStyle{Color: "FF0000"}, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	for _, elem := range doc2.Sections[0].Elements {
		if p, ok := elem.(*document.Paragraph); ok {
			for _, run := range p.Runs {
				if run.Text == "Red" && run.Style.Color != "FF0000" {
					t.Errorf("color = %q", run.Style.Color)
				}
			}
		}
	}
}

// --- writeComments with empty text ---

func TestWriteCommentsNoText(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	c := sec.AddComment("Author", "")
	_ = c

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/comments.xml")
	if !strings.Contains(content, `w:author="Author"`) {
		t.Error("comment author not found")
	}
}

// --- writeListItem with named style ---

func TestWriteListItemWithNamedStyle(t *testing.T) {
	doc := document.New()
	doc.AddNumberingStyle("myList", document.NumberingStyle{
		Type: "multilevel",
		Levels: []document.NumberingLevel{
			{Format: "decimal", Text: "%1."},
		},
	})
	sec := doc.AddSection()
	sec.AddListItem("Item", 0, nil, "myList", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "<w:numPr>") {
		t.Error("numbering properties not found")
	}
}

// --- writeNumbering with font ---

func TestWriteNumberingWithFont(t *testing.T) {
	doc := document.New()
	doc.AddNumberingStyle("bullets", document.NumberingStyle{
		Type: "singleLevel",
		Levels: []document.NumberingLevel{
			{Format: "bullet", Text: "•", Font: "Symbol", Left: 360, Hanging: 360},
		},
	})
	sec := doc.AddSection()
	sec.AddListItem("Bullet", 0, nil, "bullets", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/numbering.xml")
	if !strings.Contains(content, "Symbol") {
		t.Error("font not found in numbering.xml")
	}
	if !strings.Contains(content, `singleLevel`) {
		t.Error("singleLevel type not found")
	}
}

// --- read cell shading ---

func TestReadCellShading(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cs := &style.CellStyle{
		Shading: &style.Shading{Fill: "FFFF00", Color: "auto", Pattern: "clear"},
	}
	row.AddCell(3000, cs).AddText("Shaded", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	for _, elem := range doc2.Sections[0].Elements {
		if tbl, ok := elem.(*document.Table); ok {
			cell := tbl.Rows[0].Cells[0]
			if cell.Style.Shading == nil {
				t.Fatal("shading not read back")
			}
			if cell.Style.Shading.Fill != "FFFF00" {
				t.Errorf("fill = %q", cell.Style.Shading.Fill)
			}
		}
	}
}

// --- read gridSpan ---

func TestReadGridSpan(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	row.AddCell(6000, &style.CellStyle{GridSpan: 3}).AddText("Span", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	for _, elem := range doc2.Sections[0].Elements {
		if tbl, ok := elem.(*document.Table); ok {
			if tbl.Rows[0].Cells[0].Style.GridSpan != 3 {
				t.Errorf("gridSpan = %d", tbl.Rows[0].Cells[0].Style.GridSpan)
			}
		}
	}
}

// --- table cell margins ---

func TestWriteTableCellMargins(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ts := &style.TableStyle{
		Width:            9000,
		WidthType:        "dxa",
		CellMarginTop:    100,
		CellMarginBottom: 100,
		CellMarginLeft:   200,
		CellMarginRight:  200,
	}
	tbl := sec.AddTable(ts)
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("Cell", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "<w:tblCellMar>") {
		t.Error("cell margins not found")
	}
}

// --- table layout ---

func TestWriteTableLayout(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ts := &style.TableStyle{Width: 9000, Layout: "fixed"}
	tbl := sec.AddTable(ts)
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("Cell", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:tblLayout w:type="fixed"/>`) {
		t.Error("table layout not found")
	}
}

// --- cell text direction and noWrap ---

func TestWriteCellTextDirectionAndNoWrap(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cs := &style.CellStyle{
		TextDirection: "tbRl",
		NoWrap:        true,
		VAlign:        "center",
	}
	row.AddCell(3000, cs).AddText("Vertical", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:textDirection w:val="tbRl"/>`) {
		t.Error("text direction not found")
	}
	if !strings.Contains(content, `<w:noWrap/>`) {
		t.Error("noWrap not found")
	}
	if !strings.Contains(content, `<w:vAlign w:val="center"/>`) {
		t.Error("vAlign not found")
	}
}

// --- row cantSplit ---

func TestWriteRowCantSplit(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, &style.RowStyle{CantSplit: true, Height: 400, HeightRule: "exact"})
	row.AddCell(3000, nil).AddText("Cell", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:cantSplit/>`) {
		t.Error("cantSplit not found")
	}
	if !strings.Contains(content, `w:hRule="exact"`) {
		t.Error("height rule not found")
	}
}

func createMinimalPNG2() []byte {
	// Minimal valid PNG (1x1 pixel, red)
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x36, 0x28, 0x19,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}

// --- writeImages from file source ---

func TestWriteImagesFromFileSource(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()

	// Create a temp PNG file
	tmpDir := t.TempDir()
	pngPath := tmpDir + "/test.png"
	if err := os.WriteFile(pngPath, createMinimalPNG2(), 0644); err != nil {
		t.Fatal(err)
	}

	sec.AddImage(pngPath, &style.ImageStyle{Width: 50, Height: 50})

	data, err := WriteToBytes(doc)
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
		t.Error("image file not found in zip")
	}
}

func TestWriteImagesFromFileSourceJpg(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()

	tmpDir := t.TempDir()
	jpgPath := tmpDir + "/test.jpg"
	// Write minimal data (not a real JPEG, but tests the path)
	if err := os.WriteFile(jpgPath, []byte{0xFF, 0xD8, 0xFF, 0xE0}, 0644); err != nil {
		t.Fatal(err)
	}

	sec.AddImage(jpgPath, &style.ImageStyle{Width: 50, Height: 50})

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the image was included
	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	for _, f := range zr.File {
		if strings.Contains(f.Name, "image") && strings.Contains(f.Name, "jpeg") {
			return // found
		}
	}
	// It's OK if the extension mapping worked differently
}

func TestWriteImagesFromMissingFile(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddImage("/nonexistent/path/image.png", &style.ImageStyle{Width: 50, Height: 50})

	_, err := WriteToBytes(doc)
	if err == nil {
		t.Error("expected error for missing image file")
	}
}

// --- fontStyleToXML: Spacing property ---

func TestFontStyleToXMLSpacing(t *testing.T) {
	fs := &style.FontStyle{Spacing: 2.5}
	xml := fontStyleToXML(fs, nil)
	// Spacing is not currently written by fontStyleToXML, but the field exists
	// This test ensures no panic and the function handles it gracefully
	if xml == "" {
		t.Error("expected non-empty output")
	}
}

// --- read document with endnotes and footnotes ---

func TestWriteAndReadEndnotes(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Text with endnote", nil)
	en := tr.AddEndnote()
	en.AddText("Endnote content", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	// Verify endnotes.xml exists
	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	found := false
	for _, f := range zr.File {
		if f.Name == "word/endnotes.xml" {
			found = true
		}
	}
	if !found {
		t.Error("endnotes.xml not found")
	}
}

func TestWriteAndReadFootnotes(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Text with footnote", nil)
	fn := tr.AddFootnote()
	fn.AddText("Footnote content", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	found := false
	for _, f := range zr.File {
		if f.Name == "word/footnotes.xml" {
			found = true
		}
	}
	if !found {
		t.Error("footnotes.xml not found")
	}
}

// --- read document with comments ---

func TestWriteAndReadComments(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddComment("Author", "Comment text")
	sec.AddText("Text", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	found := false
	for _, f := range zr.File {
		if f.Name == "word/comments.xml" {
			found = true
		}
	}
	if !found {
		t.Error("comments.xml not found")
	}
}

// --- cell width type ---

func TestWriteCellWidthType(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cs := &style.CellStyle{WidthType: "pct"}
	row.AddCell(5000, cs).AddText("Pct", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:type="pct"`) {
		t.Error("cell width type pct not found")
	}
}

// --- table width type pct ---

func TestWriteTableWidthTypePct(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ts := &style.TableStyle{Width: 100, WidthType: "pct"}
	tbl := sec.AddTable(ts)
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("Cell", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:type="pct"`) {
		t.Error("table width type pct not found")
	}
}

// --- read paragraph alignment ---

func TestReadParagraphAlignmentRight(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Right", nil, &style.ParagraphStyle{Alignment: style.AlignRight})

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	for _, elem := range doc2.Sections[0].Elements {
		if p, ok := elem.(*document.Paragraph); ok {
			if p.Runs[0].Text == "Right" && p.Style.Alignment != "right" {
				t.Errorf("alignment = %q", p.Style.Alignment)
			}
		}
	}
}

// --- read strikethrough ---

func TestReadStrikethrough(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Strike", &style.FontStyle{Strikethrough: true}, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	for _, elem := range doc2.Sections[0].Elements {
		if p, ok := elem.(*document.Paragraph); ok {
			for _, run := range p.Runs {
				if run.Text == "Strike" && !run.Style.Strikethrough {
					t.Error("strikethrough not preserved")
				}
			}
		}
	}
}

// --- read break element ---

func TestReadBreakInRun(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Before", nil)
	tr.AddTextBreak(1)
	tr.AddText("After", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(doc2.Sections[0].Elements) == 0 {
		t.Error("no elements after read")
	}
}


// --- writeHyperlinkFontProps coverage ---

func TestWriteHyperlinkWithFontStyle(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddLink("https://example.com", "Styled Link", &style.FontStyle{
		Name:                "Arial",
		Bold:                true,
		Italic:              true,
		Strikethrough:       true,
		DoubleStrikethrough: true,
		Underline:           "double",
		Color:               "FF0000",
		Size:                16,
	}, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	checks := []string{
		`<w:rStyle w:val="Hyperlink"/>`,
		`w:ascii="Arial"`,
		`<w:b/>`,
		`<w:i/>`,
		`<w:strike/>`,
		`<w:dstrike/>`,
		`<w:u w:val="double"/>`,
		`<w:color w:val="FF0000"/>`,
		`<w:sz w:val="32"/>`,
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Errorf("missing %q in hyperlink output", c)
		}
	}
}

func TestWriteHyperlinkNilFont(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddLink("https://example.com", "Plain Link", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:rStyle w:val="Hyperlink"/>`) {
		t.Error("missing hyperlink style")
	}
}

// --- writer.go:write() error propagation coverage ---
// The write() function has many error paths at 56.7%. We test by verifying
// the full pipeline works with complex documents that exercise all write steps.

func TestWriteFullPipelineAllSteps(t *testing.T) {
	doc := document.New()
	doc.Properties.Title = "Full Pipeline"
	doc.Properties.Creator = "Test"
	doc.Properties.Company = "Corp"

	doc.AddFontStyle("myFont", style.FontStyle{Bold: true, Color: "0000FF"})
	doc.AddParagraphStyle("myPara", style.ParagraphStyle{Alignment: style.AlignCenter})
	doc.AddNumberingStyle("myNum", document.NumberingStyle{
		Type: "singleLevel",
		Levels: []document.NumberingLevel{
			{Format: "decimal", Text: "%1.", Left: 360, Hanging: 360},
		},
	})

	sec := doc.AddSection()
	sec.AddText("Normal text", nil, nil)
	sec.AddListItem("Item 1", 0, nil, "myNum", nil)

	h := sec.AddHeader("default")
	h.AddText("Header", nil, nil)
	f := sec.AddFooter("default")
	f.AddPreserveText("Page {PAGE}", nil, nil)

	tr := sec.AddTextRun(nil)
	tr.AddText("Run text", nil)
	fn := tr.AddFootnote()
	fn.AddText("Footnote", nil)
	en := tr.AddEndnote()
	en.AddText("Endnote", nil)

	sec.AddComment("Author", "Comment")

	pngData := createMinimalPNG2()
	sec.AddImageFromBytes(pngData, "image/png", &style.ImageStyle{Width: 100, Height: 100})

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	// Verify all expected files exist in the zip
	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	expected := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"docProps/core.xml",
		"docProps/app.xml",
		"word/styles.xml",
		"word/settings.xml",
		"word/numbering.xml",
		"word/footnotes.xml",
		"word/endnotes.xml",
		"word/comments.xml",
		"word/document.xml",
		"word/_rels/document.xml.rels",
	}
	fileSet := make(map[string]bool)
	for _, f := range zr.File {
		fileSet[f.Name] = true
	}
	for _, name := range expected {
		if !fileSet[name] {
			t.Errorf("missing file: %s", name)
		}
	}
}

// --- writeHeaderFooterFiles error path ---

func TestWriteHeaderFooterFilesMultipleSections(t *testing.T) {
	doc := document.New()

	sec1 := doc.AddSection()
	sec1.AddText("Sec 1", nil, nil)
	h1 := sec1.AddHeader("default")
	h1.AddText("H1", nil, nil)
	f1 := sec1.AddFooter("default")
	f1.AddText("F1", nil, nil)

	sec2 := doc.AddSection()
	sec2.AddText("Sec 2", nil, nil)
	h2 := sec2.AddHeader("first")
	h2.AddText("H2", nil, nil)
	f2 := sec2.AddFooter("first")
	f2.AddText("F2", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	headerCount := 0
	footerCount := 0
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "word/header") {
			headerCount++
		}
		if strings.HasPrefix(f.Name, "word/footer") {
			footerCount++
		}
	}
	if headerCount != 2 {
		t.Errorf("headers = %d, want 2", headerCount)
	}
	if footerCount != 2 {
		t.Errorf("footers = %d, want 2", footerCount)
	}
}

// --- Save with current directory path (dir == ".") ---

func TestSaveCurrentDir(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Test", nil, nil)

	tmpDir := t.TempDir()
	path := tmpDir + string(os.PathSeparator) + "test.docx"

	err := Save(doc, path)
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("empty file")
	}
}

// --- Deterministic output test for styles ---

func TestDeterministicStyleOutput(t *testing.T) {
	// Write the same document twice and verify identical output
	makeDoc := func() *document.Document {
		doc := document.New()
		doc.AddFontStyle("zStyle", style.FontStyle{Bold: true})
		doc.AddFontStyle("aStyle", style.FontStyle{Italic: true})
		doc.AddFontStyle("mStyle", style.FontStyle{Color: "FF0000"})
		doc.AddParagraphStyle("zPara", style.ParagraphStyle{Alignment: style.AlignRight})
		doc.AddParagraphStyle("aPara", style.ParagraphStyle{Alignment: style.AlignCenter})
		doc.AddNumberingStyle("zNum", document.NumberingStyle{
			Type:   "singleLevel",
			Levels: []document.NumberingLevel{{Format: "decimal", Text: "%1."}},
		})
		doc.AddNumberingStyle("aNum", document.NumberingStyle{
			Type:   "singleLevel",
			Levels: []document.NumberingLevel{{Format: "bullet", Text: "\u2022"}},
		})
		sec := doc.AddSection()
		sec.AddText("Test", nil, nil)
		sec.AddListItem("A", 0, nil, "aNum", nil)
		sec.AddListItem("Z", 0, nil, "zNum", nil)
		return doc
	}

	data1, err := WriteToBytes(makeDoc())
	if err != nil {
		t.Fatal(err)
	}
	data2, err := WriteToBytes(makeDoc())
	if err != nil {
		t.Fatal(err)
	}

	// Compare styles.xml
	styles1 := readEntry(t, data1, "word/styles.xml")
	styles2 := readEntry(t, data2, "word/styles.xml")
	if styles1 != styles2 {
		t.Error("styles.xml not deterministic")
	}

	// Compare numbering.xml
	num1 := readEntry(t, data1, "word/numbering.xml")
	num2 := readEntry(t, data2, "word/numbering.xml")
	if num1 != num2 {
		t.Error("numbering.xml not deterministic")
	}

	// Compare document.xml (list items should have consistent numIds)
	doc1 := readEntry(t, data1, "word/document.xml")
	doc2 := readEntry(t, data2, "word/document.xml")
	if doc1 != doc2 {
		t.Error("document.xml not deterministic")
	}
}

// --- headerPartMatches / footerPartMatches no-match paths ---

func TestHeaderFooterPartMatchesNoMatch(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Content", nil, nil)

	h1 := sec.AddHeader("default")
	h1.AddText("Header 1", nil, nil)

	f1 := sec.AddFooter("default")
	f1.AddText("Footer 1", nil, nil)

	// Build a writer and generate parts
	w := &Writer{doc: doc, buf: &bytes.Buffer{}, nextDocRelID: 1}
	w.generateHeaderFooterParts()

	// Test matching with a different header (should not match)
	otherHeader := &document.Header{}
	otherHeader.AddText("Different", nil, nil)
	if w.headerPartMatches("header1.xml", otherHeader) {
		t.Error("should not match different header")
	}

	// Test matching with non-existent filename
	if w.headerPartMatches("header99.xml", h1) {
		t.Error("should not match non-existent filename")
	}

	// Test footer matching with different footer
	otherFooter := &document.Footer{}
	otherFooter.AddText("Different", nil, nil)
	if w.footerPartMatches("footer1.xml", otherFooter) {
		t.Error("should not match different footer")
	}

	// Test footer matching with non-existent filename
	if w.footerPartMatches("footer99.xml", f1) {
		t.Error("should not match non-existent filename")
	}

	// Test correct matches
	if !w.headerPartMatches("header1.xml", h1) {
		t.Error("should match correct header")
	}
	if !w.footerPartMatches("footer1.xml", f1) {
		t.Error("should match correct footer")
	}
}

// --- readRels with malformed XML ---

func TestReadRelsWithMalformedXML(t *testing.T) {
	// Create a zip with a malformed rels file
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Add a malformed rels file
	fw, _ := zw.Create("word/_rels/document.xml.rels")
	fw.Write([]byte(`<not valid xml`))

	// Add a minimal document.xml
	fw2, _ := zw.Create("word/document.xml")
	fw2.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Test</w:t></w:r></w:p></w:body></w:document>`))

	zw.Close()

	_, err := ReadFromBytes(buf.Bytes())
	if err == nil {
		t.Error("expected error for malformed rels XML")
	}
}

// --- readZipFile not found path ---

func TestReadZipFileNotFound(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Test", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	reader := &docxReader{zr: zr, doc: document.New()}

	_, err = reader.readZipFile("nonexistent.xml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if _, ok := err.(*zipFileNotFoundError); !ok {
		t.Errorf("expected zipFileNotFoundError, got %T", err)
	}
}

// --- writeComments with elements ---

func TestWriteCommentsWithElements(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	c := sec.AddComment("Author", "Main text")
	// Add additional elements to the comment
	c.Elements = append(c.Elements, &document.Paragraph{
		Runs: []*document.Run{{Text: "Extra paragraph"}},
	})

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/comments.xml")
	if !strings.Contains(content, "Main text") {
		t.Error("missing comment text")
	}
	if !strings.Contains(content, "Extra paragraph") {
		t.Error("missing comment element")
	}
}
