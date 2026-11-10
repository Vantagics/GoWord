package ooxml

import (
	"bytes"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

func TestWrappedImageAndFloatingTable(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	img := sec.AddImageFromBytes([]byte{0x89, 0x50, 0x4e, 0x47, 0, 0, 0, 0}, "image/png", &style.ImageStyle{
		Width: 40, Height: 40, WrappingStyle: "square", Alignment: "right",
	})
	if img == nil {
		t.Fatal("no image")
	}
	tbl := sec.AddTable(nil)
	tbl.Style.FloatXSpec = "center"
	tbl.Style.FloatYSpec = "top"
	tbl.Style.Width = 4000
	row := tbl.AddRow(0, nil)
	row.AddCell(2000, nil).AddText("cell", nil, nil)

	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	xml := readZip(t, out, "word/document.xml")
	for _, want := range []string{"<wp:anchor", "<wp:wrapSquare", `w:tblpXSpec="center"`, `w:tblpYSpec="top"`} {
		if !strings.Contains(xml, want) {
			t.Errorf("missing %s", want)
		}
	}
}

func TestUpdateFieldsWritesHeadingText(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddTOC(nil, 1, 3)
	sec.AddTitle("Chapter One", 1)
	sec.AddText("body", nil, nil)
	doc.UpdateFields()
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	settings := readZip(t, out, "word/settings.xml")
	if !strings.Contains(settings, "updateFields") {
		t.Fatal("updateFields not set")
	}
	xml := readZip(t, out, "word/document.xml")
	if !strings.Contains(xml, "Chapter One") {
		t.Fatalf("toc result missing heading: %s", xml)
	}
}

func TestGlossaryAndEmbeddedFontRoundTrip(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Body", nil, nil)
	doc.AddBuildingBlock("Disclaimer", "Not legal advice")
	font := minimalTTF()
	if _, err := doc.EmbedFont("GoWord Test", font); err != nil {
		t.Fatal(err)
	}
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte("word/glossary/document.xml")) && !strings.Contains(readZip(t, out, "[Content_Types].xml"), "glossary") {
		t.Fatal("glossary content type missing")
	}
	gloss := readZip(t, out, "word/glossary/document.xml")
	if !strings.Contains(gloss, "Disclaimer") || !strings.Contains(gloss, "Not legal advice") {
		t.Fatalf("glossary = %s", gloss)
	}
	again, err := ReadFromBytes(out)
	if err != nil {
		t.Fatal(err)
	}
	blocks := again.BuildingBlocks()
	if len(blocks) != 1 || blocks[0].Name != "Disclaimer" || !strings.Contains(blocks[0].Text, "Not legal advice") {
		t.Fatalf("blocks = %+v", blocks)
	}
	fonts := again.EmbeddedFonts()
	if len(fonts) != 1 || fonts[0].Name != "GoWord Test" {
		t.Fatalf("fonts = %+v", fonts)
	}
	if !bytes.Equal(fonts[0].Data, font) {
		t.Fatal("deobfuscated font does not match")
	}
}

func TestCommentAndFootnotePackage(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("See note", nil, nil)
	fn := sec.AddFootnote()
	fn.AddText("The note", nil)
	c := sec.AddComment("Ada", "A comment")
	if c == nil || fn == nil {
		t.Fatal("api returned nil")
	}
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readZip(t, out, "word/footnotes.xml"), "The note") {
		t.Fatal("footnote body missing")
	}
	if !strings.Contains(readZip(t, out, "word/comments.xml"), "A comment") {
		t.Fatal("comment body missing")
	}
	if !strings.Contains(readZip(t, out, "word/document.xml"), "footnoteReference") {
		t.Fatal("footnote reference missing")
	}
}

// minimalTTF is a small TrueType font with one empty glyph. It is large
// enough to obfuscate and structurally valid enough for a package round trip.
func minimalTTF() []byte {
	// A hand-built sfnt with the required tables and zeroed glyphs.
	// Checksums are computed so the head table adjustment is consistent.
	return buildMinimalTTF()
}
