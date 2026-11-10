package ooxml

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/VantageDataChat/GoWord/document"
)

func TestLayoutMatchesLibreOfficePageBreaks(t *testing.T) {
	if _, err := exec.LookPath("soffice"); err != nil {
		t.Skip("soffice not installed")
	}
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed")
	}
	doc := document.New()
	sec := doc.AddSection()
	toc := sec.AddTOC(nil, 1, 3)
	sec.AddText("Page one", nil, nil)
	sec.Elements = append(sec.Elements, &document.PageBreak{})
	sec.AddText("Page two", nil, nil)
	sec.Elements = append(sec.Elements, &document.PageBreak{})
	sec.AddTitle("Chapter", 1)
	doc.UpdateFields()
	if doc.EstimateLayout().PageCount != 3 {
		t.Fatalf("estimate = %d", doc.EstimateLayout().PageCount)
	}
	if !strings.Contains(toc.Result, "Chapter\t3") {
		t.Fatalf("toc = %q", toc.Result)
	}
	dir := t.TempDir()
	in := filepath.Join(dir, "pages.docx")
	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(in, data, 0o644); err != nil {
		t.Fatal(err)
	}
	n, pages, err := libreOfficePageText(in)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("LibreOffice PDF pages = %d, estimate 3\n%v", n, pages)
	}
	if !strings.Contains(pages[2], "Chapter") {
		t.Fatalf("chapter not on page 3: %q", pages[2])
	}
}

func TestFlowingLayoutNearLibreOffice(t *testing.T) {
	if _, err := exec.LookPath("soffice"); err != nil {
		t.Skip("soffice not installed")
	}
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed")
	}
	doc := document.New()
	sec := doc.AddSection()
	const sentence = "The quick brown fox jumps over the lazy dog near the river bank. "
	for i := 0; i < 80; i++ {
		sec.AddText(sentence+sentence, nil, nil)
	}
	est := doc.EstimateLayout().PageCount
	dir := t.TempDir()
	in := filepath.Join(dir, "flow.docx")
	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(in, data, 0o644); err != nil {
		t.Fatal(err)
	}
	n, _, err := libreOfficePageText(in)
	if err != nil {
		t.Fatal(err)
	}
	diff := est - n
	if diff < 0 {
		diff = -diff
	}
	t.Logf("flowing text: estimate %d pages, LibreOffice %d (delta %d)", est, n, est-n)
	if diff > 1 {
		t.Fatalf("estimate %d is more than one page away from LibreOffice %d", est, n)
	}
}

func libreOfficePageText(docx string) (int, []string, error) {
	dir := filepath.Dir(docx)
	pdf := strings.TrimSuffix(docx, filepath.Ext(docx)) + ".pdf"
	cmd := exec.Command("soffice", "--headless", "--norestore", "--nolockcheck",
		"-env:UserInstallation=file:///tmp/lo-fields-"+filepath.Base(dir),
		"--convert-to", "pdf", "--outdir", dir, docx)
	if out, err := cmd.CombinedOutput(); err != nil {
		return 0, nil, errString(out, err)
	}
	txt := pdf + ".txt"
	cmd = exec.Command("pdftotext", "-layout", pdf, txt)
	if out, err := cmd.CombinedOutput(); err != nil {
		return 0, nil, errString(out, err)
	}
	b, err := os.ReadFile(txt)
	if err != nil {
		return 0, nil, err
	}
	parts := strings.Split(string(b), "\f")
	for len(parts) > 0 && strings.TrimSpace(parts[len(parts)-1]) == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		return 0, nil, errString(nil, errString(nil, nil))
	}
	return len(parts), parts, nil
}

type strErr struct{ s string }

func (e strErr) Error() string { return e.s }

func errString(out []byte, err error) error {
	if err == nil {
		return strErr{"pdf conversion produced no pages"}
	}
	return strErr{err.Error() + " " + string(out)}
}

func TestTOCFieldSeparatorAndNestedPageRef(t *testing.T) {
	data, err := os.ReadFile("../testdata/word/libreoffice/TOC_field_b.docx")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	text := doc.ExtractText()
	if !strings.Contains(text, "This is a text\t1") {
		t.Fatalf("extract = %q", text)
	}
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	again, err := ReadFromBytes(out)
	if err != nil {
		t.Fatal(err)
	}
	if again.ExtractText() != text {
		t.Fatalf("round trip extract\n got %q\nwant %q", again.ExtractText(), text)
	}

	refData, err := os.ReadFile("../testdata/word/tika/testWORD_refFields.docx")
	if err != nil {
		t.Fatal(err)
	}
	ref, err := ReadFromBytes(refData)
	if err != nil {
		t.Fatal(err)
	}
	var nested *document.Field
	var walk func([]document.Element)
	walk = func(elems []document.Element) {
		for _, el := range elems {
			switch e := el.(type) {
			case *document.Paragraph:
				for _, in := range e.Inlines {
					f, ok := in.(*document.Field)
					if !ok {
						continue
					}
					for _, c := range f.Content {
						inner, ok := c.(*document.Field)
						if ok && strings.Contains(strings.ToUpper(inner.Instr), "PAGEREF") {
							nested = inner
						}
					}
				}
			case *document.Sdt:
				walk(e.Elements)
			}
		}
	}
	for _, sec := range ref.Sections {
		walk(sec.Elements)
	}
	if nested == nil {
		t.Fatal("nested PAGEREF was not read")
	}
	ref.UpdateFieldsAt(time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC))
	if nested.Result != "1" {
		t.Fatalf("PAGEREF = %q, want 1 (bookmark is on the first sheet)", nested.Result)
	}
}

func TestHyperlinkNestedPageRef(t *testing.T) {
	data, err := os.ReadFile("../testdata/word/libreoffice/PreserveWfieldTOC.docx")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	before := doc.ExtractText()
	var nested *document.Field
	var walk func([]document.Element)
	walk = func(elems []document.Element) {
		for _, el := range elems {
			switch e := el.(type) {
			case *document.Paragraph:
				for _, in := range e.Inlines {
					h, ok := in.(*document.Hyperlink)
					if !ok {
						continue
					}
					for _, c := range h.Content {
						f, ok := c.(*document.Field)
						if ok && strings.Contains(strings.ToUpper(f.Instr), "PAGEREF") {
							nested = f
						}
					}
				}
			case *document.Sdt:
				walk(e.Elements)
			}
		}
	}
	for _, sec := range doc.Sections {
		walk(sec.Elements)
	}
	if nested == nil {
		t.Fatal("PAGEREF inside a hyperlink was not read")
	}
	doc.UpdateFieldsAt(time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC))
	if nested.Result == "" {
		t.Fatal("hyperlink PAGEREF was not evaluated")
	}
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	again, err := ReadFromBytes(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(again.ExtractText(), "Heading 1") || !strings.Contains(before, "Heading 1") {
		t.Fatalf("before %q after %q", before, again.ExtractText())
	}
}
