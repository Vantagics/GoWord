package ooxml

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

// childOrder returns the local names of the direct children of every element
// whose local name is parent.
func childOrders(xmlText, parent string) [][]string {
	dec := xml.NewDecoder(strings.NewReader(xmlText))
	var stack []string
	var orders [][]string
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if len(stack) > 0 && stack[len(stack)-1] == parent {
				orders[len(orders)-1] = append(orders[len(orders)-1], t.Name.Local)
			}
			if t.Name.Local == parent {
				orders = append(orders, nil)
			}
			stack = append(stack, t.Name.Local)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return orders
}

func inSchemaOrder(got, schema []string) bool {
	i := 0
	for _, name := range got {
		for i < len(schema) && schema[i] != name {
			i++
		}
		if i == len(schema) {
			return false
		}
		i++
	}
	return true
}

func TestGeneratedMarkupFollowsOOXMLOrder(t *testing.T) {
	doc := document.New()
	doc.SetUpdateFieldsOnOpen(true)
	doc.AddNumberingStyle("bullet", document.NumberingStyle{
		Type: "singleLevel",
		Levels: []document.NumberingLevel{{
			Format: "bullet", Text: "•", Left: 720, Hanging: 360,
		}},
	})
	doc.AddNumberingStyle("decimal", document.NumberingStyle{
		Type: "multilevel",
		Levels: []document.NumberingLevel{
			{Format: "decimal", Text: "%1."},
			{Format: "decimal", Text: "%1.%2."},
		},
	})

	sec := doc.AddSection()
	sec.Style.BreakType = "continuous"
	sec.Style.TitlePg = true
	h := sec.AddHeader("first")
	h.AddText("First header", &style.FontStyle{Bold: true, Italic: true, Size: 14, Color: "FF0000", Underline: "single"}, nil)
	sec.AddFooter("even").AddPreserveText("Page {PAGE}", nil, nil)

	sec.AddText("Body", &style.FontStyle{
		Name: "Calibri", Size: 12, Bold: true, Italic: true, Color: "112233",
		Underline: "double", HighlightColor: "yellow", AllCaps: true,
	}, &style.ParagraphStyle{
		Alignment: style.AlignCenter, SpaceBefore: 120, SpaceAfter: 80,
		Indent: 360, FirstLine: 200, KeepNext: true,
		Shading: &style.Shading{Fill: "EEEEEE", Pattern: "clear"},
		Borders: &style.ParagraphBorders{Top: &style.Border{Style: "single", Size: 4, Color: "000000"}},
	})
	sec.AddListItem("Item", 0, nil, "bullet", &style.ParagraphStyle{Alignment: style.AlignBoth})

	ts := &style.TableStyle{Width: 5000, Alignment: "center", CellSpacing: 20, Indent: 100, Layout: "fixed"}
	ts.SetAllBorders("single", 4, "000000")
	ts.CellMarginLeft = 40
	tbl := sec.AddTable(ts)
	tbl.Grid = []int{2500, 2500}
	row := tbl.AddRow(300, &style.RowStyle{IsHeader: true, CantSplit: true})
	cell := row.AddCell(2500, &style.CellStyle{GridSpan: 2, VAlign: "center", NoWrap: true, Shading: &style.Shading{Fill: "DDDDDD", Pattern: "clear"}})
	cell.AddText("Cell", nil, nil)

	sec2 := doc.AddSection()
	sec2.AddText("Second", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	assertWellFormed(t, data)

	docXML := readEntry(t, data, "word/document.xml")
	if !strings.Contains(docXML, `mc:Ignorable="w14 w15 wp14"`) {
		t.Error("document root missing mc:Ignorable")
	}
	rPrOrder := []string{"rStyle", "rFonts", "b", "bCs", "i", "iCs", "caps", "smallCaps", "strike", "dstrike", "outline", "shadow", "emboss", "imprint", "noProof", "vanish", "color", "spacing", "kern", "sz", "szCs", "highlight", "u", "effect", "vertAlign", "rtl", "lang"}
	for _, kids := range childOrders(docXML, "rPr") {
		if !inSchemaOrder(kids, rPrOrder) {
			t.Errorf("rPr out of order: %v", kids)
		}
	}
	pPrOrder := []string{"pStyle", "keepNext", "keepLines", "pageBreakBefore", "widowControl", "numPr", "pBdr", "shd", "tabs", "spacing", "ind", "jc", "sectPr"}
	for _, kids := range childOrders(docXML, "pPr") {
		if !inSchemaOrder(kids, pPrOrder) {
			t.Errorf("pPr out of order: %v", kids)
		}
	}
	tblOrder := []string{"tblStyle", "tblW", "jc", "tblCellSpacing", "tblInd", "tblBorders", "tblLayout", "tblCellMar"}
	for _, kids := range childOrders(docXML, "tblPr") {
		if !inSchemaOrder(kids, tblOrder) {
			t.Errorf("tblPr out of order: %v", kids)
		}
	}
	tcOrder := []string{"tcW", "gridSpan", "vMerge", "tcBorders", "shd", "noWrap", "textDirection", "vAlign"}
	for _, kids := range childOrders(docXML, "tcPr") {
		if !inSchemaOrder(kids, tcOrder) {
			t.Errorf("tcPr out of order: %v", kids)
		}
	}
	trOrder := []string{"cantSplit", "trHeight", "tblHeader"}
	for _, kids := range childOrders(docXML, "trPr") {
		if !inSchemaOrder(kids, trOrder) {
			t.Errorf("trPr out of order: %v", kids)
		}
	}
	sectOrder := []string{"headerReference", "footerReference", "type", "pgSz", "pgMar", "pgNumType", "cols", "titlePg"}
	for _, kids := range childOrders(docXML, "sectPr") {
		if !inSchemaOrder(kids, sectOrder) {
			t.Errorf("sectPr out of order: %v", kids)
		}
	}
	// type must precede pgSz when both are present.
	if strings.Contains(docXML, `<w:type`) {
		if strings.Index(docXML, `<w:type`) > strings.Index(docXML, `<w:pgSz`) {
			t.Error("sectPr type element is after pgSz")
		}
	}

	numXML := readEntry(t, data, "word/numbering.xml")
	if strings.Index(numXML, "</w:abstractNum>") > strings.Index(numXML, "<w:num ") {
		// The first </w:abstractNum> may still be before the first <w:num if we
		// closed each abstractNum before any num. Fail if a num appears before
		// the last abstractNum.
	}
	if i, j := strings.Index(numXML, "<w:num "), strings.LastIndex(numXML, "<w:abstractNum "); i >= 0 && j >= 0 && i < j {
		t.Error("numbering.xml interleaves num before the last abstractNum")
	}

	settings := readEntry(t, data, "word/settings.xml")
	settingsOrder := []string{"zoom", "defaultTabStop", "evenAndOddHeaders", "characterSpacingControl", "updateFields", "compat"}
	for _, kids := range childOrders(settings, "settings") {
		if !inSchemaOrder(kids, settingsOrder) {
			t.Errorf("settings out of order: %v", kids)
		}
	}
	if !strings.Contains(settings, `<w:evenAndOddHeaders`) {
		t.Error("even header did not set evenAndOddHeaders")
	}

	app := readEntry(t, data, "docProps/app.xml")
	if strings.Index(app, "<Application>") > strings.Index(app, "<Manager>") && strings.Contains(app, "<Manager>") {
		t.Error("app.xml Application element is after Manager")
	}
	if _, err := readEntryOK(data, "word/fontTable.xml"); err != nil {
		t.Error("missing fontTable.xml")
	}

	// Round trip keeps both sections and the list text.
	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc2.Sections) != 2 {
		t.Fatalf("sections = %d, want 2", len(doc2.Sections))
	}
	text := doc2.ExtractText()
	if !strings.Contains(text, "Body") || !strings.Contains(text, "Second") || !strings.Contains(text, "Item") {
		t.Fatalf("round trip text = %q", text)
	}
	if len(doc2.Sections[0].Headers) == 0 {
		t.Error("first-page header was not read back")
	}
}

func TestExplicitBoldOffRoundTrip(t *testing.T) {
	// A heading-style run that explicitly turns bold off must not come back bold.
	raw := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<w:body><w:p><w:r><w:rPr><w:b w:val="0"/><w:i/></w:rPr><w:t>Plain</w:t></w:r></w:p>
<w:sectPr><w:pgSz w:w="12240" w:h="15840"/></w:sectPr></w:body></w:document>`
	buf := zipPackage(t, map[string]string{"word/document.xml": raw})
	doc, err := ReadFromBytes(buf)
	if err != nil {
		t.Fatal(err)
	}
	var run *document.Run
	for _, el := range doc.Sections[0].Elements {
		if p, ok := el.(*document.Paragraph); ok && len(p.Runs) > 0 {
			run = p.Runs[0]
		}
	}
	if run == nil {
		t.Fatal("run not read")
	}
	if run.Style.Bold {
		t.Error("w:b w:val=0 was read as bold")
	}
	if !run.Style.Italic {
		t.Error("italic was dropped")
	}
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	xmlText := readEntry(t, out, "word/document.xml")
	if !strings.Contains(xmlText, `w:val="0"`) {
		t.Errorf("explicit bold-off was not written back: %s", xmlText)
	}
}

func TestWordFixturesRoundTrip(t *testing.T) {
	root := filepath.Join("..", "testdata", "word")
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".docx") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 60 {
		t.Fatalf("expected the Word fixture set, found %d files", len(files))
	}
	for _, path := range files {
		rel := strings.TrimPrefix(path, root+string(filepath.Separator))
		t.Run(rel, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			doc, err := ReadFromBytes(data)
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			if len(doc.Sections) == 0 {
				t.Fatal("no sections")
			}
			text := doc.ExtractText()
			out, err := WriteToBytes(doc)
			if err != nil {
				t.Fatalf("write: %v", err)
			}
			assertWellFormed(t, out)
			doc2, err := ReadFromBytes(out)
			if err != nil {
				t.Fatalf("reopen: %v", err)
			}
			if got := doc2.ExtractText(); got != text {
				t.Fatalf("extract text changed\n got %q\nwant %q", truncate(got, 200), truncate(text, 200))
			}
		})
	}
}

func TestWordFixtureFeatures(t *testing.T) {
	open := func(t *testing.T, rel string) *document.Document {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("..", "testdata", "word", rel))
		if err != nil {
			t.Fatal(err)
		}
		doc, err := ReadFromBytes(data)
		if err != nil {
			t.Fatal(err)
		}
		return doc
	}

	t.Run("sections", func(t *testing.T) {
		if n := len(open(t, "poi/Headers.docx").Sections); n != 3 {
			t.Fatalf("Headers.docx sections = %d, want 3", n)
		}
		if n := len(open(t, "python-docx/doc-access-sections.docx").Sections); n != 3 {
			t.Fatalf("doc-access-sections sections = %d, want 3", n)
		}
		if n := len(open(t, "python-docx/sct-section-props.docx").Sections); n != 5 {
			t.Fatalf("sct-section-props sections = %d, want 5", n)
		}
		if n := len(open(t, "python-docx/hdr-header-footer.docx").Sections); n != 2 {
			t.Fatalf("hdr-header-footer sections = %d, want 2", n)
		}
	})

	t.Run("headers", func(t *testing.T) {
		doc := open(t, "poi/Headers.docx")
		var text string
		for _, sec := range doc.Sections {
			for _, h := range sec.Headers {
				for _, el := range h.Elements {
					if p, ok := el.(*document.Paragraph); ok {
						for _, run := range p.Runs {
							text += run.Text
						}
					}
				}
			}
		}
		if !strings.Contains(text, "Section 1") || !strings.Contains(text, "Section 3") {
			t.Fatalf("header text = %q", text)
		}
	})

	t.Run("images", func(t *testing.T) {
		if n := len(open(t, "mammoth/tiny-picture.docx").Images()); n == 0 {
			t.Fatal("tiny-picture.docx: no images")
		}
		if n := len(open(t, "poi/VariousPictures.docx").Images()); n < 2 {
			t.Fatalf("VariousPictures images = %d", n)
		}
	})

	t.Run("hyperlink", func(t *testing.T) {
		doc := open(t, "python-docx/par-hyperlinks.docx")
		var urls []string
		var walk func([]document.Element)
		walk = func(elems []document.Element) {
			for _, el := range elems {
				switch e := el.(type) {
				case *document.Paragraph:
					for _, in := range e.Inlines {
						if h, ok := in.(*document.Hyperlink); ok && h.URL != "" {
							urls = append(urls, h.URL)
						}
					}
				case *document.Table:
					for _, row := range e.Rows {
						for _, cell := range row.Cells {
							walk(cell.Elements)
						}
					}
				}
			}
		}
		for _, sec := range doc.Sections {
			walk(sec.Elements)
		}
		if len(urls) == 0 {
			t.Fatal("no hyperlinks read")
		}
	})

	t.Run("footnotes", func(t *testing.T) {
		doc := open(t, "poi/footnotes.docx")
		if len(doc.Footnotes()) == 0 {
			t.Fatal("no footnotes")
		}
		text := doc.ExtractText()
		if strings.TrimSpace(text) == "" {
			t.Fatal("footnote document produced no text")
		}
	})

	t.Run("fields", func(t *testing.T) {
		doc := open(t, "poi/FldSimple.docx")
		found := false
		var walk func([]document.Element)
		walk = func(elems []document.Element) {
			for _, el := range elems {
				switch e := el.(type) {
				case *document.Paragraph:
					for _, in := range e.Inlines {
						if f, ok := in.(*document.Field); ok && strings.TrimSpace(f.Instr) != "" {
							found = true
						}
					}
				}
			}
		}
		for _, sec := range doc.Sections {
			walk(sec.Elements)
		}
		if !found {
			t.Fatal("fldSimple field was not read")
		}
	})

	t.Run("numbering", func(t *testing.T) {
		doc := open(t, "poi/Numbering.docx")
		n := 0
		for _, p := range doc.Paragraphs() {
			if p.Style.NumStyleName != "" {
				n++
			}
		}
		if n == 0 {
			t.Fatal("no numbering properties read")
		}
	})

	t.Run("strict-units", func(t *testing.T) {
		doc := open(t, "mammoth/strict-format.docx")
		if !strings.Contains(doc.ExtractText(), "Test") {
			t.Fatalf("text = %q", doc.ExtractText())
		}
		w := doc.Sections[0].Style.PageWidth
		// 595.30pt = 11906 twips (A4 width).
		if w < 11900 || w > 11910 {
			t.Fatalf("page width = %d twips, want about 11906", w)
		}
	})

	t.Run("textbox", func(t *testing.T) {
		doc := open(t, "poi/shapes-with-text.docx")
		if !strings.Contains(doc.ExtractText(), "Floating text box") {
			t.Fatalf("text box text missing: %q", doc.ExtractText())
		}
	})

	t.Run("math", func(t *testing.T) {
		doc := open(t, "docx4j/2010-sample1.docx")
		if !strings.Contains(doc.ExtractText(), "2010") {
			t.Fatalf("text = %q", doc.ExtractText())
		}
	})

	t.Run("nested-table", func(t *testing.T) {
		doc := open(t, "python-docx/tbl-having-tables.docx")
		if len(doc.Tables()) < 2 {
			t.Fatalf("tables = %d, want nested tables", len(doc.Tables()))
		}
	})

	t.Run("alternate-document-part", func(t *testing.T) {
		doc := open(t, "docx4j/HelloWordOnline.docx")
		if strings.TrimSpace(doc.ExtractText()) == "" {
			t.Fatal("document22.xml part was not read")
		}
	})
}

func assertWellFormed(t *testing.T, data []byte) {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	sawDoc := false
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			sawDoc = true
		}
		if !strings.HasSuffix(f.Name, ".xml") && !strings.HasSuffix(f.Name, ".rels") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		dec := xml.NewDecoder(rc)
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				rc.Close()
				t.Fatalf("invalid XML in %s: %v", f.Name, err)
			}
		}
		rc.Close()
	}
	if !sawDoc {
		t.Error("package has no word/document.xml")
	}
}

func readEntryOK(data []byte, name string) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()
			b, err := io.ReadAll(rc)
			return string(b), err
		}
	}
	return "", os.ErrNotExist
}

func zipPackage(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		fw, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
