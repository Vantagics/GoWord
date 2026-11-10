package ooxml

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/document"
)

func TestVMLTextBoxRoundTripKeepsWords(t *testing.T) {
	doc, err := Read("../testdata/word/docx4j/vml-textbox.docx")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.ExtractText(), "v:textbox/w:txbxContent") {
		t.Fatalf("extract = %q", doc.ExtractText())
	}
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	xml := readZip(t, out, "word/document.xml")
	if !strings.Contains(xml, "<w:r>") || !strings.Contains(xml, "<w:pict") {
		t.Fatalf("pict was not written inside a run: %s", xml[:min(200, len(xml))])
	}
	if strings.Contains(xml, "<w:p><w:pict") {
		t.Fatal("pict is a direct child of the paragraph")
	}
	doc2, err := ReadFromBytes(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc2.ExtractText(), "v:textbox/w:txbxContent") {
		t.Fatalf("round-trip extract = %q", doc2.ExtractText())
	}
}

func TestCommentAlternateContentStripped(t *testing.T) {
	doc, err := Read("../testdata/word/docx4j/2010-mcAlternateContent.docx")
	if err != nil {
		t.Fatal(err)
	}
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	comments := readZip(t, out, "word/comments.xml")
	if strings.Contains(comments, "<mc:AlternateContent") || strings.Contains(comments, "<AlternateContent") {
		t.Fatal("comments.xml still contains an AlternateContent element")
	}
	if !strings.Contains(comments, "The AlternateContent will be in this comment.") {
		t.Fatalf("comment text dropped: %s", comments)
	}
}

func TestRubyBaseAndReading(t *testing.T) {
	doc, err := Read("../testdata/word/poi/61470.docx")
	if err != nil {
		t.Fatal(err)
	}
	text := doc.ExtractText()
	if !strings.Contains(text, "東京") {
		t.Fatalf("base missing: %q", text)
	}
	if strings.Contains(text, "とうきょう") {
		t.Fatalf("reading leaked into ExtractText: %q", text)
	}
	rubies := doc.Rubies()
	if len(rubies) != 1 || rubies[0].Base != "東京" || rubies[0].Reading != "とうきょう" {
		t.Fatalf("rubies = %+v", rubies)
	}
	opt := document.ExtractDefaults()
	opt.Ruby = "both"
	both := doc.ExtractTextWith(opt)
	if !strings.Contains(both, "東京") || !strings.Contains(both, "とうきょう") {
		t.Fatalf("both = %q", both)
	}
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	xml := readZip(t, out, "word/document.xml")
	if !strings.Contains(xml, "<w:ruby") || !strings.Contains(xml, "とうきょう") {
		t.Fatal("ruby markup was not preserved")
	}
}

func TestNumberingOverrides(t *testing.T) {
	doc, err := Read("../testdata/word/poi/NumberingWOverrides.docx")
	if err != nil {
		t.Fatal(err)
	}
	text := doc.ExtractText()
	for _, want := range []string{"1 1", "A B", "B C", "A A", "B B", "2 4"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %q", want, text)
		}
	}
	if strings.Contains(text, "1 B") || strings.Contains(text, "2 C") {
		t.Fatalf("abstract format used instead of the override: %q", text)
	}
}

func TestNestedMoveText(t *testing.T) {
	doc, err := Read("../testdata/word/poi/Tika-792.docx")
	if err != nil {
		t.Fatal(err)
	}
	text := doc.ExtractText()
	if !strings.Contains(text, "s.") || !strings.Contains(text, "b") {
		t.Fatalf("extract = %q", text)
	}
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	xml := readZip(t, out, "word/document.xml")
	if !strings.Contains(xml, "moveFrom") || !strings.Contains(xml, "moveTo") {
		t.Fatal("move markup was not preserved")
	}
}

func TestExtractOptions(t *testing.T) {
	doc, err := Read("../testdata/word/docx4j/2010-mcAlternateContent.docx")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.ExtractText(), "The AlternateContent will be in this comment.") {
		t.Fatal("ExtractText dropped the comment body")
	}
	bare := doc.ExtractTextWith(document.ExtractOptions{})
	if strings.Contains(bare, "The AlternateContent will be in this comment.") {
		t.Fatalf("zero options still include comments: %q", bare)
	}
	if !strings.Contains(bare, "Hello") {
		t.Fatalf("body missing: %q", bare)
	}

	notes, err := Read("../testdata/word/poi/footnotes.docx")
	if err != nil {
		t.Fatal(err)
	}
	full := notes.ExtractText()
	noNotes := notes.ExtractTextWith(document.ExtractOptions{})
	if len(strings.TrimSpace(noNotes)) >= len(strings.TrimSpace(full)) {
		t.Fatalf("omitting note bodies did not shorten text\nfull=%q\nnone=%q", full, noNotes)
	}
	sep := document.ExtractDefaults()
	sep.NoteAnchor = "separate"
	separated := notes.ExtractTextWith(sep)
	if strings.Contains(full, "prostoy1") && !strings.Contains(separated, "prostoy 1") {
		t.Fatalf("separate anchor = %q", separated)
	}
}

func TestOMMLToLatex(t *testing.T) {
	doc, err := Read("../testdata/word/phpword/reader-formula.docx")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(doc.ExtractText(), `\frac`) {
		t.Fatalf("default extract became TeX: %q", doc.ExtractText())
	}
	opt := document.ExtractDefaults()
	opt.MathFormat = "latex"
	got := doc.ExtractTextWith(opt)
	if !strings.Contains(got, `\frac{`) {
		t.Fatalf("latex = %q", got)
	}
	sample, err := Read("../testdata/word/docx4j/2010-sample1.docx")
	if err != nil {
		t.Fatal(err)
	}
	opt = document.ExtractDefaults()
	opt.MathFormat = "latex"
	sampleTex := sample.ExtractTextWith(opt)
	if !strings.Contains(sampleTex, `\frac{dy}{dx}`) || !strings.Contains(sampleTex, `\sqrt`) {
		t.Fatalf("sample latex = %q", sampleTex)
	}
}

func TestStyleLinkedNumbering(t *testing.T) {
	doc, err := Read("../testdata/word/python-docx/num-having-numbering-part.docx")
	if err != nil {
		t.Fatal(err)
	}
	text := doc.ExtractText()
	if !strings.Contains(text, "1.") || !strings.Contains(text, "Paragraph having List Number style.") {
		t.Fatalf("style numbering missing: %q", text)
	}
}

func TestShapeTextsStaySeparateWords(t *testing.T) {
	doc, err := Read("../testdata/word/poi/shapes-with-text.docx")
	if err != nil {
		t.Fatal(err)
	}
	text := doc.ExtractText()
	if !strings.Contains(text, "text box\n") {
		t.Fatalf("shape text ran together: %q", text)
	}
	if !strings.Contains(text, "A square shape") {
		t.Fatalf("square missing: %q", text)
	}
}

func TestStrictWriteIsOptIn(t *testing.T) {
	doc := document.New()
	doc.AddSection().AddText("Hello", nil, nil)
	trans, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(trans, []byte("purl.oclc.org/ooxml")) {
		t.Fatal("transitional save used a strict namespace")
	}
	doc.SetStrict(true)
	strict, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	docXML := readZip(t, strict, "word/document.xml")
	if !strings.Contains(docXML, "http://purl.oclc.org/ooxml/wordprocessingml/main") {
		t.Fatalf("document namespace: %s", docXML[:min(400, len(docXML))])
	}
	rels := readZip(t, strict, "_rels/.rels")
	if !strings.Contains(rels, "http://purl.oclc.org/ooxml/officeDocument/relationships/officeDocument") {
		t.Fatalf("root rels: %s", rels)
	}
	styles := readZip(t, strict, "word/styles.xml")
	if !strings.Contains(styles, "http://purl.oclc.org/ooxml/wordprocessingml/main") {
		t.Fatal("styles stayed transitional")
	}
	again, err := ReadFromBytes(strict)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(again.ExtractText(), "Hello") {
		t.Fatalf("strict reread = %q", again.ExtractText())
	}
}

func readZip(t *testing.T, data []byte, name string) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer rc.Close()
		var b bytes.Buffer
		if _, err := b.ReadFrom(rc); err != nil {
			t.Fatal(err)
		}
		return b.String()
	}
	t.Fatalf("missing %s", name)
	return ""
}
