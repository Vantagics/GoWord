package goword

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/style"
)

func TestWatermarkTextInDocx(t *testing.T) {
	doc := New()
	wm := doc.AddWatermarkText("CONFIDENTIAL")
	wm.Bold = true
	wm.Color = "FF0000"
	wm.Font = "Arial"
	wm.Size = 72

	sec := doc.AddSection()
	sec.AddText("Document with watermark", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// Check that a header was auto-created with watermark
	headerContent := readZipEntry(t, data, "word/header1.xml")
	if !strings.Contains(headerContent, "CONFIDENTIAL") {
		t.Error("watermark text not found in header")
	}
	if !strings.Contains(headerContent, "PowerPlusWaterMarkObject") {
		t.Error("watermark shape not found")
	}
}

func TestWatermarkTextWithExistingHeader(t *testing.T) {
	doc := New()
	doc.AddWatermarkText("DRAFT")

	sec := doc.AddSection()
	sec.AddText("Body", nil, nil)
	h := sec.AddHeader("")
	h.AddText("My Header", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	headerContent := readZipEntry(t, data, "word/header1.xml")
	if !strings.Contains(headerContent, "DRAFT") {
		t.Error("watermark not in header")
	}
	if !strings.Contains(headerContent, "My Header") {
		t.Error("header text missing")
	}
}

func TestWatermarkPictureInDocx(t *testing.T) {
	doc := New()
	pngData := createMinimalPNG()
	wm := doc.AddWatermarkPictureFromBytes(pngData, "image/png")
	wm.Width = 300
	wm.Height = 300
	wm.Washout = true

	sec := doc.AddSection()
	sec.AddText("Document with picture watermark", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	headerContent := readZipEntry(t, data, "word/header1.xml")
	if !strings.Contains(headerContent, "WordPictureWatermark") {
		t.Error("picture watermark not found in header")
	}
}

func TestFormFieldTextInputInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ff := sec.AddTextInput("fullname")
	ff.DefaultValue = "Enter name"
	ff.MaxLength = 100

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "FORMTEXT") {
		t.Error("missing FORMTEXT field")
	}
	if !strings.Contains(content, `w:val="fullname"`) {
		t.Error("missing field name")
	}
}

func TestFormFieldDropDownInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ff := sec.AddDropdownList("color", []string{"Red", "Green", "Blue"})
	ff.DefaultValue = "Green"

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "FORMDROPDOWN") {
		t.Error("missing FORMDROPDOWN field")
	}
	if !strings.Contains(content, `w:val="Red"`) {
		t.Error("missing dropdown value Red")
	}
	if !strings.Contains(content, `w:val="Green"`) {
		t.Error("missing dropdown value Green")
	}
	if !strings.Contains(content, `w:val="Blue"`) {
		t.Error("missing dropdown value Blue")
	}
}

func TestFormFieldCheckBoxInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ff := sec.AddFormField(FormFieldTypeCheckBox, "agree")
	ff.Value = "1"

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "FORMCHECKBOX") {
		t.Error("missing FORMCHECKBOX field")
	}
}

func TestTabStopsInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ps := &style.ParagraphStyle{
		TabStops: []style.TabStop{
			{Position: 2880, Type: "left", Leader: "none"},
			{Position: 5760, Type: "center", Leader: "dot"},
			{Position: 8640, Type: "right", Leader: "hyphen"},
		},
	}
	tr := sec.AddTextRun(ps)
	tr.AddText("Col1", nil)
	tr.AddTab()
	tr.AddText("Col2", nil)
	tr.AddTab()
	tr.AddText("Col3", nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "<w:tabs>") {
		t.Error("missing tabs element")
	}
	if !strings.Contains(content, `w:pos="2880"`) {
		t.Error("missing tab stop position")
	}
	if !strings.Contains(content, `w:leader="dot"`) {
		t.Error("missing tab leader")
	}
	if !strings.Contains(content, "<w:tab/>") {
		t.Error("missing tab character")
	}
}

func TestFontStyleEmbossShadowInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fs := &style.FontStyle{
		Bold:    true,
		Emboss:  true,
		Shadow:  true,
		Imprint: true,
		Outline: true,
		Size:    14,
	}
	sec.AddText("Styled text", fs, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "<w:emboss/>") {
		t.Error("missing emboss")
	}
	if !strings.Contains(content, "<w:shadow/>") {
		t.Error("missing shadow")
	}
	if !strings.Contains(content, "<w:imprint/>") {
		t.Error("missing imprint")
	}
	if !strings.Contains(content, "<w:outline/>") {
		t.Error("missing outline")
	}
}

func TestFontStyleTextEffectInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fs := &style.FontStyle{TextEffect: "shimmer"}
	sec.AddText("Shimmer", fs, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:effect w:val="shimmer"/>`) {
		t.Error("missing text effect")
	}
}

func TestFontStyleRTLInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fs := &style.FontStyle{RightToLeft: true}
	sec.AddText("RTL text", fs, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "<w:rtl/>") {
		t.Error("missing rtl")
	}
}

func TestFontStyleEastAsiaFontInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fs := &style.FontStyle{Name: "Arial", NameEastAsia: "SimSun"}
	sec.AddText("East Asia", fs, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:eastAsia="SimSun"`) {
		t.Error("missing east asia font")
	}
}

func TestFontStyleEastAsiaOnlyInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fs := &style.FontStyle{NameEastAsia: "SimSun"}
	sec.AddText("East Asia Only", fs, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:eastAsia="SimSun"`) {
		t.Error("missing east asia font")
	}
}

func TestFontStyleUnderlineColorInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fs := &style.FontStyle{Underline: "single", UnderlineColor: "FF0000"}
	sec.AddText("Underlined", fs, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:color="FF0000"`) {
		t.Error("missing underline color")
	}
}

func TestFontStyleKerningInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fs := &style.FontStyle{Kerning: 12}
	sec.AddText("Kerned", fs, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:kern w:val="24"/>`) { // 12pt * 2 = 24 half-points
		t.Error("missing kerning")
	}
}

func TestFontStyleSpacingInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	fs := &style.FontStyle{Spacing: 2}
	sec.AddText("Spaced", fs, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:spacing w:val="40"/>`) { // 2pt * 20 = 40 twips
		t.Error("missing character spacing")
	}
}

func TestUpdateFieldsOnOpenInDocx(t *testing.T) {
	doc := New()
	doc.SetUpdateFieldsOnOpen(true)
	sec := doc.AddSection()
	sec.AddText("Test", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	settings := readZipEntry(t, data, "word/settings.xml")
	if !strings.Contains(settings, `<w:updateFields w:val="true"/>`) {
		t.Error("missing updateFields setting")
	}
}

func TestUpdateFieldsOnOpenDefaultInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Test", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	settings := readZipEntry(t, data, "word/settings.xml")
	if strings.Contains(settings, "updateFields") {
		t.Error("updateFields should not be present by default")
	}
}

func TestTableCellSpacingInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ts := &style.TableStyle{CellSpacing: 20}
	tbl := sec.AddTable(ts)
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("A", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:tblCellSpacing w:w="20" w:type="dxa"/>`) {
		t.Error("missing cell spacing")
	}
}

func TestTableIndentInDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ts := &style.TableStyle{Indent: 720}
	tbl := sec.AddTable(ts)
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("A", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:tblInd w:w="720" w:type="dxa"/>`) {
		t.Error("missing table indent")
	}
}

func TestExtractTextAPI(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Hello", nil, nil)
	sec.AddText("World", nil, nil)

	text := doc.ExtractText()
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) < 2 {
		t.Errorf("expected at least 2 lines, got %d", len(lines))
	}
}

func TestTabStopDefaultValues(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ps := &style.ParagraphStyle{
		TabStops: []style.TabStop{
			{Position: 1440}, // no Type or Leader specified
		},
	}
	sec.AddText("Tab test", nil, ps)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `w:val="left"`) {
		t.Error("missing default tab type")
	}
	if !strings.Contains(content, `w:leader="none"`) {
		t.Error("missing default tab leader")
	}
}

func TestFieldCodesInPreserveText(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	footer := sec.AddFooter("")
	footer.AddPreserveText("{DATE} - Page {PAGE}", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	content := readZipEntry(t, data, "word/footer1.xml")
	if !strings.Contains(content, "DATE") {
		t.Error("missing DATE field")
	}
	if !strings.Contains(content, "PAGE") {
		t.Error("missing PAGE field")
	}
}

func TestFullFeatureNewAPIs(t *testing.T) {
	doc := New()
	doc.SetUpdateFieldsOnOpen(true)
	doc.AddWatermarkText("DRAFT")

	sec := doc.AddSection()
	sec.AddText("Document with new features", nil, nil)

	// Form fields
	sec.AddTextInput("name").DefaultValue = "John"
	sec.AddDropdownList("dept", []string{"Engineering", "Sales", "HR"})

	// Tab stops
	ps := &style.ParagraphStyle{
		TabStops: []style.TabStop{
			{Position: 2880, Type: "left"},
			{Position: 5760, Type: "right", Leader: "dot"},
		},
	}
	tr := sec.AddTextRun(ps)
	tr.AddText("Item", nil)
	tr.AddTab()
	tr.AddText("100", nil)

	// Rich font styles
	sec.AddText("Embossed", &style.FontStyle{Emboss: true}, nil)
	sec.AddText("Shadow", &style.FontStyle{Shadow: true}, nil)

	// Table with spacing
	ts := &style.TableStyle{CellSpacing: 10, Indent: 360}
	tbl := sec.AddTable(ts)
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("Cell", nil, nil)

	data, err := doc.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// Verify ZIP is valid
	_, err = zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("invalid ZIP: %v", err)
	}

	// Verify key content
	content := readZipEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "FORMTEXT") {
		t.Error("missing text input form field")
	}
	if !strings.Contains(content, "FORMDROPDOWN") {
		t.Error("missing dropdown form field")
	}
	if !strings.Contains(content, "<w:emboss/>") {
		t.Error("missing emboss")
	}

	settings := readZipEntry(t, data, "word/settings.xml")
	if !strings.Contains(settings, "updateFields") {
		t.Error("missing updateFields")
	}

	header := readZipEntry(t, data, "word/header1.xml")
	if !strings.Contains(header, "DRAFT") {
		t.Error("missing watermark")
	}
}
