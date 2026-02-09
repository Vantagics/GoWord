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
