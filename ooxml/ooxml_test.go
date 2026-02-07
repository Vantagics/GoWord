package ooxml

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

func newTestDoc() *document.Document {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Test content", nil, nil)
	return doc
}

func TestWriteToBytes(t *testing.T) {
	doc := newTestDoc()
	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("empty output")
	}
}

func TestWriteAndReadRoundTrip(t *testing.T) {
	doc := document.New()
	doc.Properties.Title = "Test"
	doc.Properties.Creator = "GoWord"
	sec := doc.AddSection()
	sec.AddText("Hello", nil, nil)
	sec.AddText("World", &style.FontStyle{Bold: true}, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	if doc2.Properties.Title != "Test" {
		t.Errorf("title = %q", doc2.Properties.Title)
	}
	if len(doc2.Sections) == 0 {
		t.Fatal("no sections")
	}
	if len(doc2.Sections[0].Elements) < 2 {
		t.Errorf("elements = %d", len(doc2.Sections[0].Elements))
	}
}

func TestContentTypesHasRequiredEntries(t *testing.T) {
	doc := newTestDoc()
	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	ct := readEntry(t, data, "[Content_Types].xml")
	required := []string{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml",
		"application/vnd.openxmlformats-package.core-properties+xml",
	}
	for _, r := range required {
		if !strings.Contains(ct, r) {
			t.Errorf("missing content type: %s", r)
		}
	}
}

func TestDocumentXMLNamespaces(t *testing.T) {
	doc := newTestDoc()
	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	namespaces := []string{
		`xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"`,
		`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"`,
	}
	for _, ns := range namespaces {
		if !strings.Contains(content, ns) {
			t.Errorf("missing namespace: %s", ns)
		}
	}
}

func TestXMLValidity(t *testing.T) {
	doc := document.New()
	doc.Properties.Title = "XML Test"
	sec := doc.AddSection()
	sec.AddText("Special chars: & < > \" '", nil, nil)
	sec.AddTitle("Heading", 1)
	sec.AddPageBreak()
	sec.AddTextBreak(2)
	sec.AddLink("https://example.com", "Link", nil, nil)

	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("Cell", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	// Validate all XML files in the ZIP
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
				if err.Error() == "EOF" {
					break
				}
				t.Errorf("invalid XML in %s: %v", f.Name, err)
				break
			}
		}
		rc.Close()
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"a & b", "a &amp; b"},
		{"<tag>", "&lt;tag&gt;"},
		{`"quoted"`, "&#34;quoted&#34;"},
	}
	for _, tt := range tests {
		got := escapeXML(tt.input)
		if got != tt.want {
			t.Errorf("escapeXML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParsePreserveText(t *testing.T) {
	parts := parsePreserveText("Page {PAGE} of {NUMPAGES}")
	if len(parts) != 4 {
		t.Fatalf("parts = %d, want 4", len(parts))
	}
	if parts[0].text != "Page " || parts[0].isField {
		t.Errorf("part 0: %+v", parts[0])
	}
	if parts[1].text != "PAGE" || !parts[1].isField {
		t.Errorf("part 1: %+v", parts[1])
	}
	if parts[2].text != " of " || parts[2].isField {
		t.Errorf("part 2: %+v", parts[2])
	}
	if parts[3].text != "NUMPAGES" || !parts[3].isField {
		t.Errorf("part 3: %+v", parts[3])
	}
}

func TestImageExtension(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{"image/png", "png"},
		{"image/jpeg", "jpeg"},
		{"image/gif", "gif"},
		{"image/bmp", "bmp"},
		{"unknown", "png"},
	}
	for _, tt := range tests {
		got := imageExtension(tt.mime)
		if got != tt.want {
			t.Errorf("imageExtension(%q) = %q, want %q", tt.mime, got, tt.want)
		}
	}
}

func TestImageContentType(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{"png", "image/png"},
		{"jpeg", "image/jpeg"},
		{"gif", "image/gif"},
	}
	for _, tt := range tests {
		got := imageContentType(tt.ext)
		if got != tt.want {
			t.Errorf("imageContentType(%q) = %q, want %q", tt.ext, got, tt.want)
		}
	}
}

func TestComplexDocument(t *testing.T) {
	doc := document.New()
	doc.Properties.Title = "Complex Test"
	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultFontSize(11)

	doc.AddFontStyle("bold", style.FontStyle{Bold: true})
	doc.AddParagraphStyle("centered", style.ParagraphStyle{Alignment: style.AlignCenter})

	sec := doc.AddSection()
	sec.AddTitle("Document Title", 0)
	sec.AddTitle("Chapter 1", 1)
	sec.AddText("Normal paragraph.", nil, nil)
	sec.AddTextWithStyle("Styled text", "bold", "centered")

	tr := sec.AddTextRun(nil)
	tr.AddText("Mixed ", nil)
	tr.AddText("bold ", &style.FontStyle{Bold: true})
	tr.AddText("italic", &style.FontStyle{Italic: true})

	sec.AddLink("https://go.dev", "Go Language", nil, nil)
	sec.AddPageBreak()

	tbl := sec.AddTable(&style.TableStyle{Width: 9000, WidthType: "dxa"})
	tbl.Grid = []int{3000, 3000, 3000}
	row := tbl.AddRow(400, &style.RowStyle{IsHeader: true})
	row.AddCell(3000, nil).AddText("Col 1", &style.FontStyle{Bold: true}, nil)
	row.AddCell(3000, nil).AddText("Col 2", &style.FontStyle{Bold: true}, nil)
	row.AddCell(3000, nil).AddText("Col 3", &style.FontStyle{Bold: true}, nil)

	row2 := tbl.AddRow(0, nil)
	row2.AddCell(3000, nil).AddText("A", nil, nil)
	row2.AddCell(3000, nil).AddText("B", nil, nil)
	row2.AddCell(3000, nil).AddText("C", nil, nil)

	header := sec.AddHeader("")
	header.AddText("Document Header", nil, nil)

	footer := sec.AddFooter("")
	footer.AddPreserveText("Page {PAGE}", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	// Verify round-trip
	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	if doc2.Properties.Title != "Complex Test" {
		t.Errorf("title = %q", doc2.Properties.Title)
	}
	if len(doc2.Sections) == 0 {
		t.Fatal("no sections")
	}
}

func TestReadTableFromDocx(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("Cell A", nil, nil)
	row.AddCell(3000, nil).AddText("Cell B", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	// Find the table
	found := false
	for _, elem := range doc2.Sections[0].Elements {
		if tbl, ok := elem.(*document.Table); ok {
			found = true
			if len(tbl.Rows) != 1 {
				t.Errorf("rows = %d", len(tbl.Rows))
			}
			if len(tbl.Rows[0].Cells) != 2 {
				t.Errorf("cells = %d", len(tbl.Rows[0].Cells))
			}
		}
	}
	if !found {
		t.Error("table not found after round trip")
	}
}

func readEntry(t *testing.T, data []byte, name string) string {
	t.Helper()
	zr, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	for _, f := range zr.File {
		if f.Name == name {
			rc, _ := f.Open()
			defer rc.Close()
			var buf bytes.Buffer
			buf.ReadFrom(rc)
			return buf.String()
		}
	}
	t.Fatalf("entry %s not found", name)
	return ""
}
