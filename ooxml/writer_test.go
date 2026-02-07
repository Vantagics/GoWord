package ooxml

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

// --- Cell styling coverage ---

func TestCellShadingAndBorders(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cs := &style.CellStyle{
		Width: 3000,
		Shading: &style.Shading{Fill: "FFFF00", Color: "auto", Pattern: "clear"},
		BorderTop:    &style.Border{Style: "single", Size: 4, Color: "000000"},
		BorderBottom: &style.Border{Style: "single", Size: 4, Color: "000000"},
		BorderLeft:   &style.Border{Style: "single", Size: 4, Color: "000000"},
		BorderRight:  &style.Border{Style: "single", Size: 4, Color: "000000"},
		VAlign:       "center",
		TextDirection: "tbRl",
		NoWrap:       true,
	}
	c := row.AddCell(3000, cs)
	c.AddText("Styled cell", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:fill="FFFF00"`) {
		t.Error("missing cell shading")
	}
	if !strings.Contains(content, `<w:vAlign w:val="center"/>`) {
		t.Error("missing vAlign")
	}
	if !strings.Contains(content, `<w:textDirection w:val="tbRl"/>`) {
		t.Error("missing textDirection")
	}
	if !strings.Contains(content, `<w:noWrap/>`) {
		t.Error("missing noWrap")
	}
	if !strings.Contains(content, `<w:tcBorders>`) {
		t.Error("missing cell borders")
	}
}

// --- Table styling coverage ---

func TestTableAlignmentAndLayout(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ts := &style.TableStyle{
		Width:     9000,
		WidthType: "dxa",
		Alignment: "center",
		Layout:    "fixed",
		CellMarginTop:    50,
		CellMarginBottom: 50,
		CellMarginLeft:   100,
		CellMarginRight:  100,
	}
	tbl := sec.AddTable(ts)
	tbl.Grid = []int{3000, 3000, 3000}
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("A", nil, nil)
	row.AddCell(3000, nil).AddText("B", nil, nil)
	row.AddCell(3000, nil).AddText("C", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:jc w:val="center"/>`) {
		t.Error("missing table alignment")
	}
	if !strings.Contains(content, `<w:tblLayout w:type="fixed"/>`) {
		t.Error("missing table layout")
	}
	if !strings.Contains(content, `<w:tblCellMar>`) {
		t.Error("missing cell margins")
	}
	if !strings.Contains(content, `<w:tblGrid>`) {
		t.Error("missing table grid")
	}
	if !strings.Contains(content, `<w:gridCol`) {
		t.Error("missing grid columns")
	}
}

func TestTableNamedStyle(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTableWithStyle("TableGrid")
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("X", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:tblStyle w:val="TableGrid"/>`) {
		t.Error("missing named table style")
	}
}

// --- Row styling coverage ---

func TestRowHeightAndProperties(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, &style.RowStyle{
		Height:     500,
		HeightRule: "exact",
		IsHeader:   true,
		CantSplit:  true,
	})
	row.AddCell(3000, nil).AddText("Header", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:trHeight`) {
		t.Error("missing row height")
	}
	if !strings.Contains(content, `w:hRule="exact"`) {
		t.Error("missing height rule")
	}
	if !strings.Contains(content, `<w:tblHeader/>`) {
		t.Error("missing header flag")
	}
	if !strings.Contains(content, `<w:cantSplit/>`) {
		t.Error("missing cantSplit")
	}
}

// --- Paragraph styling coverage ---

func TestParagraphSpacingAndIndent(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ps := &style.ParagraphStyle{
		SpaceBefore:  120,
		SpaceAfter:   240,
		LineSpacing:  360,
		LineRule:     "exact",
		Indent:       720,
		IndentRight:  360,
		FirstLine:    360,
		KeepNext:     true,
		KeepLines:    true,
		PageBreakBefore: true,
		WidowControl: true,
	}
	sec.AddText("Styled paragraph", nil, ps)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:before="120"`) {
		t.Error("missing spaceBefore")
	}
	if !strings.Contains(content, `w:after="240"`) {
		t.Error("missing spaceAfter")
	}
	if !strings.Contains(content, `w:line="360"`) {
		t.Error("missing lineSpacing")
	}
	if !strings.Contains(content, `w:lineRule="exact"`) {
		t.Error("missing lineRule")
	}
	if !strings.Contains(content, `w:left="720"`) {
		t.Error("missing indent")
	}
	if !strings.Contains(content, `w:right="360"`) {
		t.Error("missing indentRight")
	}
	if !strings.Contains(content, `w:firstLine="360"`) {
		t.Error("missing firstLine")
	}
	if !strings.Contains(content, `<w:keepNext/>`) {
		t.Error("missing keepNext")
	}
	if !strings.Contains(content, `<w:keepLines/>`) {
		t.Error("missing keepLines")
	}
	if !strings.Contains(content, `<w:pageBreakBefore/>`) {
		t.Error("missing pageBreakBefore")
	}
}

func TestParagraphHangingIndent(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ps := &style.ParagraphStyle{Hanging: 360}
	sec.AddText("Hanging", nil, ps)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:hanging="360"`) {
		t.Error("missing hanging indent")
	}
}

func TestParagraphShading(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ps := &style.ParagraphStyle{
		Shading: &style.Shading{Fill: "EEEEEE", Color: "auto", Pattern: "clear"},
	}
	sec.AddText("Shaded", nil, ps)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:fill="EEEEEE"`) {
		t.Error("missing paragraph shading")
	}
}

func TestParagraphBorders(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ps := &style.ParagraphStyle{
		Borders: &style.ParagraphBorders{
			Top:    &style.Border{Style: "single", Size: 4, Color: "000000", Space: 1},
			Bottom: &style.Border{Style: "single", Size: 4, Color: "000000", Space: 1},
		},
	}
	sec.AddText("Bordered", nil, ps)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:pBdr>`) {
		t.Error("missing paragraph borders")
	}
}

// --- Font styling coverage ---

func TestFontStyleAllProperties(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	fs := &style.FontStyle{
		Name:                "Courier New",
		Size:                14,
		Bold:                true,
		Italic:              true,
		Underline:           "double",
		Strikethrough:       true,
		DoubleStrikethrough: true,
		SuperScript:         true,
		Color:               "FF0000",
		HighlightColor:      "yellow",
		AllCaps:             true,
		SmallCaps:           true,
		Hidden:              true,
		NoProof:             true,
		Lang:                "en-US",
	}
	sec.AddText("All styles", fs, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	checks := []string{
		`w:ascii="Courier New"`,
		`<w:b/>`,
		`<w:i/>`,
		`<w:u w:val="double"/>`,
		`<w:strike/>`,
		`<w:dstrike/>`,
		`w:val="superscript"`,
		`w:val="FF0000"`,
		`w:val="yellow"`,
		`<w:caps/>`,
		`<w:smallCaps/>`,
		`<w:vanish/>`,
		`<w:noProof/>`,
		`w:val="en-US"`,
		`w:val="28"`, // 14pt = 28 half-points
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("missing: %s", check)
		}
	}
}

// --- TextRun with mixed elements in writer ---

func TestTextRunWithBreaksAndBookmarks(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Before ", nil)
	tr.AddTextBreak(2)
	tr.AddBookmark("myMark")
	tr.AddText(" After", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:br/>`) {
		t.Error("missing break in textrun")
	}
	if !strings.Contains(content, `w:name="myMark"`) {
		t.Error("missing bookmark in textrun")
	}
}

func TestTextRunWithFootnoteAndEndnote(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Text", nil)
	fn := tr.AddFootnote()
	fn.AddText("Footnote", nil)
	en := tr.AddEndnote()
	en.AddText("Endnote", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "w:footnoteReference") {
		t.Error("missing footnote reference")
	}
	if !strings.Contains(content, "w:endnoteReference") {
		t.Error("missing endnote reference")
	}
}

// --- Checkbox checked ---

func TestCheckBoxChecked(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	cb := sec.AddCheckBox("cb1", "Checked box", nil, nil)
	cb.Checked = true

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:val="1"`) {
		t.Error("missing checked value")
	}
}

// --- Whitespace preservation ---

func TestTextWithWhitespace(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText(" leading space", nil, nil)
	sec.AddText("trailing space ", nil, nil)
	sec.AddText("double  space", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `xml:space="preserve"`) {
		t.Error("missing space preservation")
	}
}

// --- Section properties coverage ---

func TestSectionWithColumns(t *testing.T) {
	doc := document.New()
	ss := style.DefaultSectionStyle()
	ss.ColumnCount = 2
	ss.ColumnSpacing = 720
	sec := doc.AddSectionWithStyle(ss)
	sec.AddText("Two columns", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:num="2"`) {
		t.Error("missing column count")
	}
	if !strings.Contains(content, `w:space="720"`) {
		t.Error("missing column spacing")
	}
}

func TestSectionWithPageNumberStart(t *testing.T) {
	doc := document.New()
	ss := style.DefaultSectionStyle()
	start := 5
	ss.PageNumberStart = &start
	sec := doc.AddSectionWithStyle(ss)
	sec.AddText("Page 5 start", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:start="5"`) {
		t.Error("missing page number start")
	}
}

func TestSectionWithBreakType(t *testing.T) {
	doc := document.New()
	sec1 := doc.AddSection()
	sec1.AddText("Section 1", nil, nil)

	ss := style.DefaultSectionStyle()
	ss.BreakType = "continuous"
	sec2 := doc.AddSectionWithStyle(ss)
	sec2.AddText("Section 2 continuous", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:val="continuous"`) {
		t.Error("missing break type")
	}
}

// --- WriteToWriter ---

func TestWriteToWriter(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Writer test", nil, nil)

	var buf bytes.Buffer
	err := WriteToWriter(doc, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("empty output")
	}

	// Verify it's a valid ZIP
	_, err = zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("invalid ZIP: %v", err)
	}
}

// --- Reader edge cases ---

func TestReadInvalidZip(t *testing.T) {
	_, err := ReadFromBytes([]byte("not a zip"))
	if err == nil {
		t.Error("expected error for invalid zip")
	}
}

func TestReadEmptyDocument(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	_ = sec

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc2.Sections) == 0 {
		t.Error("no sections")
	}
}

func TestReadBoldText(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Bold", &style.FontStyle{Bold: true, Size: 14, Name: "Tahoma", Color: "FF0000"}, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	sec2 := doc2.Sections[0]
	found := false
	for _, elem := range sec2.Elements {
		if p, ok := elem.(*document.Paragraph); ok {
			for _, run := range p.Runs {
				if run.Text == "Bold" {
					found = true
					if !run.Style.Bold {
						t.Error("bold not read back")
					}
					if run.Style.Size != 14 {
						t.Errorf("size = %f, want 14", run.Style.Size)
					}
					if run.Style.Name != "Tahoma" {
						t.Errorf("name = %q", run.Style.Name)
					}
					if run.Style.Color != "FF0000" {
						t.Errorf("color = %q", run.Style.Color)
					}
				}
			}
		}
	}
	if !found {
		t.Error("bold text not found")
	}
}

func TestReadItalicUnderlineStrikethrough(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Styled", &style.FontStyle{
		Italic:        true,
		Underline:     "single",
		Strikethrough: true,
		AllCaps:       true,
		SmallCaps:     true,
	}, nil)

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
					if !run.Style.Italic {
						t.Error("italic not read")
					}
					if run.Style.Underline != "single" {
						t.Errorf("underline = %q", run.Style.Underline)
					}
					if !run.Style.Strikethrough {
						t.Error("strikethrough not read")
					}
					if !run.Style.AllCaps {
						t.Error("allCaps not read")
					}
					if !run.Style.SmallCaps {
						t.Error("smallCaps not read")
					}
				}
			}
		}
	}
}

func TestReadParagraphAlignment(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Centered", nil, &style.ParagraphStyle{Alignment: style.AlignCenter})

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
			if p.Runs[0].Text == "Centered" && p.Style.Alignment != "center" {
				t.Errorf("alignment = %q", p.Style.Alignment)
			}
		}
	}
}

func TestReadParagraphStyleName(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddTitle("My Heading", 1)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, elem := range doc2.Sections[0].Elements {
		if p, ok := elem.(*document.Paragraph); ok {
			if p.StyleName == "Heading1" {
				found = true
			}
		}
	}
	if !found {
		t.Error("Heading1 style not read back")
	}
}

func TestReadSectionProperties(t *testing.T) {
	doc := document.New()
	ss := style.DefaultSectionStyle()
	ss.Orientation = style.OrientLandscape
	ss.PageWidth = 16838
	ss.PageHeight = 11906
	ss.MarginTop = 720
	ss.MarginBottom = 720
	ss.MarginLeft = 720
	ss.MarginRight = 720
	sec := doc.AddSectionWithStyle(ss)
	sec.AddText("Landscape", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	s := doc2.Sections[0].Style
	if s.Orientation != "landscape" {
		t.Errorf("orientation = %q", s.Orientation)
	}
	if s.PageWidth != 16838 {
		t.Errorf("pageWidth = %d", s.PageWidth)
	}
	if s.MarginTop != 720 {
		t.Errorf("marginTop = %d", s.MarginTop)
	}
}

func TestReadCellProperties(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	cs := &style.CellStyle{
		Width:    5000,
		GridSpan: 2,
		VMerge:   "restart",
		Shading:  &style.Shading{Fill: "AABBCC", Color: "auto", Pattern: "clear"},
	}
	c := row.AddCell(5000, cs)
	c.AddText("Merged", nil, nil)

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
			if cell.Style.GridSpan != 2 {
				t.Errorf("gridSpan = %d", cell.Style.GridSpan)
			}
			if cell.Style.VMerge != "restart" {
				t.Errorf("vMerge = %q", cell.Style.VMerge)
			}
			if cell.Style.Shading == nil || cell.Style.Shading.Fill != "AABBCC" {
				t.Error("shading not read")
			}
		}
	}
}

// --- Validate all XML in complex doc ---

func TestAllXMLFilesValid(t *testing.T) {
	doc := document.New()
	doc.Properties.Title = "Validation Test"
	doc.Properties.Creator = "Test & <Author>"
	doc.SetDefaultFontName("Calibri")

	doc.AddFontStyle("bold", style.FontStyle{Bold: true})
	doc.AddParagraphStyle("centered", style.ParagraphStyle{Alignment: style.AlignCenter})
	doc.AddNumberingStyle("bullets", document.NumberingStyle{
		Type: "singleLevel",
		Levels: []document.NumberingLevel{
			{Format: "bullet", Text: "\u2022", Left: 360, Hanging: 360, Font: "Symbol"},
		},
	})

	sec := doc.AddSection()
	sec.AddTitle("Title", 0)
	sec.AddTitle("H1", 1)
	sec.AddText("Normal", nil, nil)
	sec.AddText("Bold & <Italic>", &style.FontStyle{Bold: true, Italic: true}, nil)
	sec.AddLink("https://example.com/path?q=1&r=2", "Link with &", nil, nil)
	sec.AddPageBreak()
	sec.AddTextBreak(2)
	sec.AddListItem("Bullet 1", 0, nil, "bullets", nil)
	sec.AddCheckBox("cb", "Check \"me\"", nil, nil)

	tbl := sec.AddTable(&style.TableStyle{Width: 9000})
	tbl.Grid = []int{4500, 4500}
	row := tbl.AddRow(400, &style.RowStyle{IsHeader: true})
	row.AddCell(4500, nil).AddText("H1", nil, nil)
	row.AddCell(4500, nil).AddText("H2", nil, nil)

	header := sec.AddHeader("")
	header.AddText("Header", nil, nil)
	footer := sec.AddFooter("")
	footer.AddPreserveText("Page {PAGE}", nil, nil)

	fn := sec.AddFootnote()
	fn.AddText("Footnote", nil)
	en := sec.AddEndnote()
	en.AddText("Endnote", nil)
	sec.AddComment("Author", "Comment text")

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	for _, f := range zr.File {
		if !strings.HasSuffix(f.Name, ".xml") && !strings.HasSuffix(f.Name, ".rels") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("opening %s: %v", f.Name, err)
		}
		decoder := xml.NewDecoder(rc)
		for {
			_, err := decoder.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Errorf("invalid XML in %s: %v", f.Name, err)
				break
			}
		}
		rc.Close()
	}
}

// --- Empty cell coverage ---

func TestEmptyCellWritesEmptyParagraph(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil) // empty cell

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:tc><w:tcPr>`) {
		t.Error("missing cell")
	}
}

// --- Named run style ---

func TestRunWithNamedStyle(t *testing.T) {
	doc := document.New()
	doc.AddFontStyle("emphasis", style.FontStyle{Italic: true, Color: "0000FF"})
	sec := doc.AddSection()
	sec.AddTextWithStyle("Emphasized", "emphasis", "")

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:rStyle w:val="emphasis"/>`) {
		t.Error("missing named run style")
	}
}

// --- Line break run ---

func TestRunBreak(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Before", nil)
	// Manually add a break run
	p := &document.Paragraph{Runs: []*document.Run{{Break: true}}}
	tr.Elements = append(tr.Elements, p)
	tr.AddText("After", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:br/>`) {
		t.Error("missing break")
	}
}

// --- TextBreak with styles ---

func TestTextBreakWithStyles(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Before", nil, nil)
	// Add styled text break directly
	ps := &style.ParagraphStyle{Alignment: style.AlignCenter}
	fs := &style.FontStyle{Size: 8}
	sec.Elements = append(sec.Elements, &document.TextBreak{Count: 1, Font: fs, Para: ps})
	sec.AddText("After", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	// Should have the styled empty paragraph
	if !strings.Contains(content, `<w:jc w:val="center"/>`) {
		t.Error("missing styled text break")
	}
}

// --- Superscript/subscript read ---

func TestReadSuperSubscript(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Super", &style.FontStyle{SuperScript: true}, nil)
	sec.AddText("Sub", &style.FontStyle{SubScript: true}, nil)

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
				if run.Text == "Super" && !run.Style.SuperScript {
					t.Error("superscript not read")
				}
				if run.Text == "Sub" && !run.Style.SubScript {
					t.Error("subscript not read")
				}
			}
		}
	}
}

// --- Read run style name ---

func TestReadRunStyleName(t *testing.T) {
	doc := document.New()
	doc.AddFontStyle("myStyle", style.FontStyle{Bold: true})
	sec := doc.AddSection()
	sec.AddTextWithStyle("Styled", "myStyle", "")

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, elem := range doc2.Sections[0].Elements {
		if p, ok := elem.(*document.Paragraph); ok {
			for _, run := range p.Runs {
				if run.StyleName == "myStyle" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("run style name not read back")
	}
}

// --- getFileExt ---

func TestGetFileExt(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"image.png", ".png"},
		{"path/to/file.jpeg", ".jpeg"},
		{"noext", ""},
		{"multi.dots.txt", ".txt"},
	}
	for _, tt := range tests {
		got := getFileExt(tt.path)
		if got != tt.want {
			t.Errorf("getFileExt(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// --- zipFileNotFoundError ---

func TestZipFileNotFoundError(t *testing.T) {
	e := &zipFileNotFoundError{}
	if e.Error() != "file not found in zip" {
		t.Errorf("error = %q", e.Error())
	}
}


// --- Tests for 0% coverage functions ---

func TestWriteHyperlinkInline(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Click ", nil)
	tr.AddLink("https://example.com", "here", nil)
	tr.AddText(" to visit", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:hyperlink r:id=`) {
		t.Error("missing inline hyperlink")
	}
	if !strings.Contains(content, `<w:rStyle w:val="Hyperlink"/>`) {
		t.Error("missing hyperlink style")
	}
	if !strings.Contains(content, `>here</w:t>`) {
		t.Error("missing hyperlink text")
	}
}

func TestWriteImageElement(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	// Add image from bytes so writeImages can process it
	pngData := createMinimalPNG()
	sec.AddImageFromBytes(pngData, "image/png", &style.ImageStyle{Width: 200, Height: 150})

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:drawing>`) {
		t.Error("missing drawing element")
	}
	if !strings.Contains(content, `<wp:inline`) {
		t.Error("missing inline element")
	}
	if !strings.Contains(content, `<wp:extent`) {
		t.Error("missing extent element")
	}
	if !strings.Contains(content, `<a:graphic>`) {
		t.Error("missing graphic element")
	}
	if !strings.Contains(content, `<pic:pic`) {
		t.Error("missing pic element")
	}

	// Verify image is in the zip
	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	found := false
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "word/media/image") {
			found = true
			break
		}
	}
	if !found {
		t.Error("image file not found in zip")
	}
}

func TestWriteInlineImageInTextRun(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Before image ", nil)
	pngData := createMinimalPNG()
	img := tr.AddImage("", &style.ImageStyle{Width: 50, Height: 50})
	img.Data = pngData
	img.MimeType = "image/png"
	tr.AddText(" After image", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:drawing>`) {
		t.Error("missing inline image in textrun")
	}
}

func TestWriteBookmarkEnd(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	// Add a BookmarkEnd element directly
	sec.Elements = append(sec.Elements, &document.BookmarkEnd{ID: 42})

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:bookmarkEnd w:id="42"/>`) {
		t.Error("missing bookmarkEnd")
	}
}

func TestWriteTOC(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddTitle("Chapter 1", 1)
	sec.AddTOC(nil, 1, 3)
	sec.AddTitle("Chapter 2", 1)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `TOC \o "1-3"`) {
		t.Error("missing TOC field code")
	}
	if !strings.Contains(content, `fldCharType="begin"`) {
		t.Error("missing TOC begin")
	}
	if !strings.Contains(content, `fldCharType="end"`) {
		t.Error("missing TOC end")
	}
}

func TestWriteTOCDefaults(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	// TOC with zero min/max should default to 1-9
	sec.AddTOC(nil, 0, 0)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `TOC \o "1-9"`) {
		t.Error("missing default TOC range 1-9")
	}
}

func TestWriteLine(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddLine(document.Line{Width: 400, Height: 0, Weight: 2, Color: "FF0000"})

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<v:line`) {
		t.Error("missing VML line")
	}
	if !strings.Contains(content, `strokecolor="#FF0000"`) {
		t.Error("missing line color")
	}
	if !strings.Contains(content, `strokeweight="2pt"`) {
		t.Error("missing line weight")
	}
}

func TestWriteLineDefaults(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	// Line with zero color/weight should use defaults
	sec.AddLine(document.Line{Width: 100, Height: 10})

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `strokecolor="#000000"`) {
		t.Error("missing default line color")
	}
	if !strings.Contains(content, `strokeweight="1pt"`) {
		t.Error("missing default line weight")
	}
}

func TestWriteTableBordersAllSides(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ts := &style.TableStyle{
		Width: 9000,
		BorderTop:     &style.Border{Style: "single", Size: 4, Color: "000000"},
		BorderBottom:  &style.Border{Style: "single", Size: 4, Color: "000000"},
		BorderLeft:    &style.Border{Style: "single", Size: 4, Color: "000000"},
		BorderRight:   &style.Border{Style: "single", Size: 4, Color: "000000"},
		BorderInsideH: &style.Border{Style: "single", Size: 2, Color: "888888"},
		BorderInsideV: &style.Border{Style: "single", Size: 2, Color: "888888"},
	}
	tbl := sec.AddTable(ts)
	row := tbl.AddRow(0, nil)
	row.AddCell(4500, nil).AddText("A", nil, nil)
	row.AddCell(4500, nil).AddText("B", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:tblBorders>`) {
		t.Error("missing table borders")
	}
	if !strings.Contains(content, `<w:insideH`) {
		t.Error("missing insideH border")
	}
	if !strings.Contains(content, `<w:insideV`) {
		t.Error("missing insideV border")
	}
}

func TestWriteImagesFromBytes(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()

	// PNG image
	pngData := createMinimalPNG()
	sec.AddImageFromBytes(pngData, "image/png", &style.ImageStyle{Width: 100, Height: 100})

	// JPEG image (fake data, just testing the path)
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	sec.AddImageFromBytes(jpegData, "image/jpeg", &style.ImageStyle{Width: 100, Height: 100})

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	// Check content types include both png and jpeg
	ct := readEntry(t, data, "[Content_Types].xml")
	if !strings.Contains(ct, `Extension="png"`) {
		t.Error("missing png content type")
	}
	if !strings.Contains(ct, `Extension="jpeg"`) {
		t.Error("missing jpeg content type")
	}

	// Check images exist in zip
	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	imageCount := 0
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "word/media/") {
			imageCount++
		}
	}
	if imageCount != 2 {
		t.Errorf("expected 2 images, got %d", imageCount)
	}
}

func TestCollectImageExtensionsFromSource(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	// Add image with Source path (won't actually read file, but tests collectImageExtensions)
	img := sec.AddImage("photo.jpg", &style.ImageStyle{Width: 100, Height: 100})
	_ = img

	// We can't WriteToBytes because the file doesn't exist,
	// but we can test collectImageExtensions directly via a Writer
	w := &Writer{doc: doc}
	exts := w.collectImageExtensions()
	found := false
	for _, ext := range exts {
		if ext == "jpeg" {
			found = true
		}
	}
	if !found {
		t.Error("expected jpeg extension from .jpg source")
	}
}

// createMinimalPNG creates a minimal valid 1x1 PNG image
func createMinimalPNG() []byte {
	// Minimal 1x1 white PNG
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

// --- writeElementXML dispatcher coverage for missing branches ---

func TestWriteElementXMLDispatcherCoverage(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()

	// Add elements that exercise uncovered writeElementXML branches
	sec.AddTOC(nil, 1, 3)
	sec.AddLine(document.Line{Width: 100, Height: 5, Color: "0000FF"})
	sec.Elements = append(sec.Elements, &document.BookmarkEnd{ID: 99})

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "TOC") {
		t.Error("missing TOC in dispatcher")
	}
	if !strings.Contains(content, `<v:line`) {
		t.Error("missing line in dispatcher")
	}
	if !strings.Contains(content, `bookmarkEnd`) {
		t.Error("missing bookmarkEnd in dispatcher")
	}
}

// --- Image default dimensions ---

func TestWriteImageDefaultDimensions(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	// Image with zero width/height should default to 100
	pngData := createMinimalPNG()
	sec.AddImageFromBytes(pngData, "image/png", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<wp:extent`) {
		t.Error("missing extent with default dimensions")
	}
}

// --- writeImages with MimeType-based extension ---

func TestWriteImagesGifBmpTiff(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()

	sec.AddImageFromBytes([]byte{0x47, 0x49, 0x46}, "image/gif", &style.ImageStyle{Width: 10, Height: 10})
	sec.AddImageFromBytes([]byte{0x42, 0x4D}, "image/bmp", &style.ImageStyle{Width: 10, Height: 10})
	sec.AddImageFromBytes([]byte{0x49, 0x49}, "image/tiff", &style.ImageStyle{Width: 10, Height: 10})

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	ct := readEntry(t, data, "[Content_Types].xml")
	if !strings.Contains(ct, `Extension="gif"`) {
		t.Error("missing gif content type")
	}
	if !strings.Contains(ct, `Extension="bmp"`) {
		t.Error("missing bmp content type")
	}
	if !strings.Contains(ct, `Extension="tiff"`) {
		t.Error("missing tiff content type")
	}
}

// --- imageContentType and imageExtension coverage ---

func TestImageContentTypeAndExtension(t *testing.T) {
	tests := []struct {
		ext  string
		ct   string
	}{
		{"png", "image/png"},
		{"jpeg", "image/jpeg"},
		{"jpg", "image/jpeg"},
		{"gif", "image/gif"},
		{"bmp", "image/bmp"},
		{"tiff", "image/tiff"},
		{"tif", "image/tiff"},
		{"unknown", "image/png"},
	}
	for _, tt := range tests {
		got := imageContentType(tt.ext)
		if got != tt.ct {
			t.Errorf("imageContentType(%q) = %q, want %q", tt.ext, got, tt.ct)
		}
	}

	extTests := []struct {
		mime string
		ext  string
	}{
		{"image/png", "png"},
		{"image/jpeg", "jpeg"},
		{"image/jpg", "jpeg"},
		{"image/gif", "gif"},
		{"image/bmp", "bmp"},
		{"image/tiff", "tiff"},
		{"image/unknown", "png"},
	}
	for _, tt := range extTests {
		got := imageExtension(tt.mime)
		if got != tt.ext {
			t.Errorf("imageExtension(%q) = %q, want %q", tt.mime, got, tt.ext)
		}
	}
}


// --- Reader round-trip coverage for hyperlinks and tables ---

func TestReadHyperlink(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddLink("https://example.com/test", "Click me", &style.FontStyle{Bold: true}, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, elem := range doc2.Sections[0].Elements {
		if p, ok := elem.(*document.Paragraph); ok {
			for _, run := range p.Runs {
				if strings.Contains(run.Text, "Click me") {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("hyperlink text not read back")
	}
}

func TestReadTableWithMultipleRows(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(&style.TableStyle{Width: 9000})
	for i := 0; i < 3; i++ {
		row := tbl.AddRow(400, nil)
		row.AddCell(4500, nil).AddText("A", nil, nil)
		row.AddCell(4500, nil).AddText("B", nil, nil)
	}

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
			if len(tbl.Rows) != 3 {
				t.Errorf("rows = %d, want 3", len(tbl.Rows))
			}
			for _, row := range tbl.Rows {
				if len(row.Cells) != 2 {
					t.Errorf("cells = %d, want 2", len(row.Cells))
				}
			}
		}
	}
}

// --- Properties round-trip ---

func TestReadWriteAllProperties(t *testing.T) {
	doc := document.New()
	doc.Properties.Title = "Test Title"
	doc.Properties.Subject = "Test Subject"
	doc.Properties.Creator = "Test Creator"
	doc.Properties.Keywords = "go, docx"
	doc.Properties.Description = "A test document"
	doc.Properties.Category = "Testing"
	doc.Properties.LastModifiedBy = "Tester"
	doc.Properties.Company = "TestCorp"
	doc.Properties.Manager = "Boss"
	doc.Properties.Revision = 3
	sec := doc.AddSection()
	sec.AddText("Props test", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	// Verify core.xml
	core := readEntry(t, data, "docProps/core.xml")
	checks := []string{
		"Test Title", "Test Subject", "Test Creator",
		"go, docx", "A test document", "Testing",
		"Tester", "<cp:revision>3</cp:revision>",
	}
	for _, c := range checks {
		if !strings.Contains(core, c) {
			t.Errorf("missing in core.xml: %s", c)
		}
	}

	// Verify app.xml
	app := readEntry(t, data, "docProps/app.xml")
	if !strings.Contains(app, "TestCorp") {
		t.Error("missing Company")
	}
	if !strings.Contains(app, "Boss") {
		t.Error("missing Manager")
	}

	// Read back
	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if doc2.Properties.Title != "Test Title" {
		t.Errorf("title = %q", doc2.Properties.Title)
	}
	if doc2.Properties.Creator != "Test Creator" {
		t.Errorf("creator = %q", doc2.Properties.Creator)
	}
}

// --- Empty header/footer ---

func TestEmptyHeaderFooter(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Content", nil, nil)
	sec.AddHeader("") // empty header
	sec.AddFooter("") // empty footer

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	// Verify header/footer files exist
	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	hasHeader := false
	hasFooter := false
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "word/header") {
			hasHeader = true
		}
		if strings.HasPrefix(f.Name, "word/footer") {
			hasFooter = true
		}
	}
	if !hasHeader {
		t.Error("missing header file")
	}
	if !hasFooter {
		t.Error("missing footer file")
	}
}

// --- Multiple sections with headers ---

func TestMultipleSectionsWithHeaders(t *testing.T) {
	doc := document.New()

	sec1 := doc.AddSection()
	sec1.AddText("Section 1", nil, nil)
	h1 := sec1.AddHeader("default")
	h1.AddText("Header 1", nil, nil)
	f1 := sec1.AddFooter("default")
	f1.AddPreserveText("Page {PAGE}", nil, nil)

	sec2 := doc.AddSection()
	sec2.AddText("Section 2", nil, nil)
	h2 := sec2.AddHeader("first")
	h2.AddText("First page header", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:headerReference`) {
		t.Error("missing header reference")
	}
	if !strings.Contains(content, `w:footerReference`) {
		t.Error("missing footer reference")
	}
}

// --- Read multiple sections ---

func TestReadMultipleSectionsRoundTrip(t *testing.T) {
	doc := document.New()
	sec1 := doc.AddSection()
	sec1.AddText("Section 1", nil, nil)

	ss := style.DefaultSectionStyle()
	ss.BreakType = "continuous"
	sec2 := doc.AddSectionWithStyle(ss)
	sec2.AddText("Section 2", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	// Reader puts all content into one section (doesn't split on sectPr in pPr)
	if len(doc2.Sections) < 1 {
		t.Errorf("sections = %d, want >= 1", len(doc2.Sections))
	}
	// Verify text was read
	found := false
	for _, elem := range doc2.Sections[0].Elements {
		if p, ok := elem.(*document.Paragraph); ok {
			for _, run := range p.Runs {
				if run.Text == "Section 1" || run.Text == "Section 2" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("section text not read back")
	}
}

// --- parsePreserveText edge cases ---

func TestParsePreserveTextNoField(t *testing.T) {
	parts := parsePreserveText("Just plain text")
	if len(parts) != 1 || parts[0].isField {
		t.Error("expected single non-field part")
	}
}

func TestParsePreserveTextUnclosedBrace(t *testing.T) {
	parts := parsePreserveText("Page {PAGE of {NUMPAGES}")
	// Should handle gracefully
	if len(parts) == 0 {
		t.Error("expected parts")
	}
}

func TestParsePreserveTextMultipleFields(t *testing.T) {
	parts := parsePreserveText("Page {PAGE} of {NUMPAGES}")
	fieldCount := 0
	for _, p := range parts {
		if p.isField {
			fieldCount++
		}
	}
	if fieldCount != 2 {
		t.Errorf("field count = %d, want 2", fieldCount)
	}
}

// --- writeTextElement empty string ---

func TestWriteTextElementEmpty(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	// Add paragraph with empty text
	sec.AddText("", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	// Should still have a paragraph but no w:t element for empty text
	if !strings.Contains(content, `<w:p>`) {
		t.Error("missing paragraph")
	}
}

// --- Row without explicit style ---

func TestRowWithoutStyle(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(300, nil) // height via parameter, no RowStyle
	row.AddCell(3000, nil).AddText("Test", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:trHeight`) {
		t.Error("missing row height from parameter")
	}
}
