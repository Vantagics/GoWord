package ooxml

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/VantageDataChat/GoWord/document"
)

func TestGraphicTextOption(t *testing.T) {
	chart, err := Read("../testdata/word/powertools/HC043-Chart.docx")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(chart.ExtractText(), "Series 1") {
		t.Fatal("default extract included chart labels")
	}
	opt := document.ExtractDefaults()
	opt.IncludeGraphicText = true
	got := chart.ExtractTextWith(opt)
	for _, want := range []string{"Series 1", "Category 1", "Series 2"} {
		if !strings.Contains(got, want) {
			t.Errorf("chart text missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "4.3") {
		t.Fatalf("numeric cache leaked into graphic text: %q", got)
	}

	art, err := Read("../testdata/word/powertools/HC052-SmartArt.docx")
	if err != nil {
		t.Fatal(err)
	}
	opt = document.ExtractDefaults()
	opt.IncludeGraphicText = true
	artText := art.ExtractTextWith(opt)
	for _, want := range []string{"AB", "CD", "IJ"} {
		if !strings.Contains(artText, want) {
			t.Errorf("smartart missing %q in %q", want, artText)
		}
	}
	// The diagram data part and the drawing copy the same labels.
	if strings.Count(artText, "AB") != 1 {
		t.Fatalf("smartart label duplicated: %q", artText)
	}
}

func TestFormattingRevisionsRoundTrip(t *testing.T) {
	cases := []struct {
		path string
		mark string
	}{
		{"../testdata/word/powertools/RP024-ParagraphMark-rPr-Change.docx", "rPrChange"},
		{"../testdata/word/powertools/RP025-Paragraph-Props-Change.docx", "pPrChange"},
		{"../testdata/word/powertools/RP031-Table-Prop-Change.docx", "tblPrChange"},
	}
	for _, tc := range cases {
		doc, err := Read(tc.path)
		if err != nil {
			t.Fatal(err)
		}
		out, err := WriteToBytes(doc)
		if err != nil {
			t.Fatal(err)
		}
		xml := readZip(t, out, "word/document.xml")
		if !strings.Contains(xml, tc.mark) {
			t.Errorf("%s: rewrite dropped %s", tc.path, tc.mark)
		}
	}
}

func TestUnclosedFieldText(t *testing.T) {
	doc, err := Read("../testdata/word/tika/testWORD_unclosedFields.docx")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.ExtractText(), "junk") {
		t.Fatalf("missing field text: %q", doc.ExtractText())
	}
}

func TestPreservedPrefixRoundTrip(t *testing.T) {
	doc, err := Read("../testdata/word/powertools/RP051-Arabic.docx")
	if err != nil {
		t.Fatal(err)
	}
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	xmlText := readZip(t, out, "word/document.xml")
	if !strings.Contains(xmlText, `xmlns:pt14="http://powertools.codeplex.com/2011"`) {
		t.Fatal("pt14 namespace was dropped")
	}
	if !strings.Contains(xmlText, "pt14:Unid") {
		t.Fatal("pt14 attribute was dropped")
	}
	dec := xml.NewDecoder(strings.NewReader(xmlText))
	for {
		if _, err := dec.Token(); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
	}
}

func TestLargeDocumentRoundTrip(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	const n = 2000
	for i := 0; i < n; i++ {
		sec.AddText("paragraph", nil, nil)
	}
	start := time.Now()
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > 15*time.Second {
		t.Fatalf("writing %d paragraphs took %s", n, time.Since(start))
	}
	again, err := ReadFromBytes(out)
	if err != nil {
		t.Fatal(err)
	}
	text := again.ExtractText()
	if strings.Count(text, "paragraph") != n {
		t.Fatalf("paragraphs = %d, want %d", strings.Count(text, "paragraph"), n)
	}
}
