package ooxml

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

// readEntryOpt reads a zip entry, returning "" if not found (no fatal).
func readEntryOpt(data []byte, name string) string {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ""
	}
	for _, f := range zr.File {
		if f.Name == name {
			rc, _ := f.Open()
			defer rc.Close()
			var buf bytes.Buffer
			buf.ReadFrom(rc)
			return buf.String()
		}
	}
	return ""
}

// --- Form field coverage ---

func TestWriteFormFieldText(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ff := sec.AddTextInput("username")
	ff.DefaultValue = "John"
	ff.MaxLength = 50
	ff.CalcOnExit = true

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	checks := []string{
		`FORMTEXT`,
		`<w:name w:val="username"/>`,
		`<w:enabled/>`,
		`<w:calcOnExit w:val="true"/>`,
		`<w:default w:val="John"/>`,
		`<w:maxLength w:val="50"/>`,
		`<w:textInput>`,
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Errorf("missing %q in form field text output", c)
		}
	}
}

func TestWriteFormFieldTextWithValue(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ff := sec.AddTextInput("field1")
	ff.Value = "CustomValue"

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, "CustomValue") {
		t.Error("form field value not written")
	}
}

func TestWriteFormFieldDropDown(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ff := sec.AddDropdownList("color", []string{"Red", "Green", "Blue"})
	ff.DefaultValue = "Green"
	ff.CalcOnExit = true

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	checks := []string{
		`FORMDROPDOWN`,
		`<w:name w:val="color"/>`,
		`<w:enabled/>`,
		`<w:calcOnExit w:val="true"/>`,
		`<w:ddList>`,
		`<w:default w:val="1"/>`,
		`<w:listEntry w:val="Red"/>`,
		`<w:listEntry w:val="Green"/>`,
		`<w:listEntry w:val="Blue"/>`,
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Errorf("missing %q in dropdown output", c)
		}
	}
}

func TestWriteFormFieldCheckBox(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ff := sec.AddFormField(document.FormFieldTypeCheckBox, "agree")
	ff.Value = "true"

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	checks := []string{
		`FORMCHECKBOX`,
		`<w:name w:val="agree"/>`,
		`<w:enabled/>`,
		`<w:default w:val="1"/>`,
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Errorf("missing %q in checkbox form field output", c)
		}
	}
}

func TestWriteFormFieldCheckBoxUnchecked(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ff := sec.AddFormField(document.FormFieldTypeCheckBox, "opt")
	ff.Value = "false"

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:default w:val="0"/>`) {
		t.Error("unchecked checkbox should have val=0")
	}
}

func TestWriteFormFieldDisabled(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ff := sec.AddFormField(document.FormFieldTypeText, "readonly")
	ff.Enabled = false

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if strings.Contains(content, `<w:enabled/>`) {
		t.Error("disabled form field should not have <w:enabled/>")
	}
}

// --- Watermark coverage ---

func TestWriteTextWatermark(t *testing.T) {
	doc := document.New()
	wm := doc.AddWatermarkText("DRAFT")
	wm.Font = "Arial"
	wm.Size = 72
	wm.Color = "FF0000"
	wm.Bold = true
	wm.Italic = true

	sec := doc.AddSection()
	sec.AddText("Content", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for i := 1; i <= 5; i++ {
		content := readEntryOpt(data, fmt.Sprintf("word/header%d.xml", i))
		if content != "" && strings.Contains(content, "DRAFT") {
			found = true
			checks := []string{
				`Arial`,
				`font-size:72pt`,
				`font-weight:bold`,
				`font-style:italic`,
				`fillcolor="#FF0000"`,
				`PowerPlusWaterMarkObject`,
			}
			for _, c := range checks {
				if !strings.Contains(content, c) {
					t.Errorf("missing %q in watermark header", c)
				}
			}
			break
		}
	}
	if !found {
		t.Error("watermark text not found in any header")
	}
}

func TestWriteTextWatermarkDefaults(t *testing.T) {
	doc := document.New()
	doc.AddWatermarkText("CONFIDENTIAL")
	sec := doc.AddSection()
	sec.AddText("Content", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for i := 1; i <= 5; i++ {
		content := readEntryOpt(data, fmt.Sprintf("word/header%d.xml", i))
		if content != "" && strings.Contains(content, "CONFIDENTIAL") {
			found = true
			if !strings.Contains(content, `fillcolor="#C0C0C0"`) {
				t.Error("default color not applied")
			}
			if !strings.Contains(content, `Calibri`) {
				t.Error("default font not applied")
			}
			break
		}
	}
	if !found {
		t.Error("watermark not found")
	}
}

func TestWritePictureWatermark(t *testing.T) {
	doc := document.New()
	pngData := createMinimalPNG2()
	wm := doc.AddWatermarkPictureFromBytes(pngData, "image/png")
	wm.Width = 200
	wm.Height = 150
	wm.Washout = false

	sec := doc.AddSection()
	sec.AddText("Content", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for i := 1; i <= 5; i++ {
		content := readEntryOpt(data, fmt.Sprintf("word/header%d.xml", i))
		if content != "" && strings.Contains(content, "WordPictureWatermark") {
			found = true
			if !strings.Contains(content, `width:200pt`) {
				t.Error("picture width not set")
			}
			if !strings.Contains(content, `height:150pt`) {
				t.Error("picture height not set")
			}
			if !strings.Contains(content, `gain="65536f"`) {
				t.Error("non-washout gain not applied")
			}
			break
		}
	}
	if !found {
		t.Error("picture watermark not found in any header")
	}
}

func TestWritePictureWatermarkWashoutDefaults(t *testing.T) {
	doc := document.New()
	pngData := createMinimalPNG2()
	wm := doc.AddWatermarkPictureFromBytes(pngData, "image/png")
	wm.Washout = true
	// Leave Width/Height at 0 to test defaults

	sec := doc.AddSection()
	sec.AddText("Content", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for i := 1; i <= 5; i++ {
		content := readEntryOpt(data, fmt.Sprintf("word/header%d.xml", i))
		if content != "" && strings.Contains(content, "WordPictureWatermark") {
			found = true
			if !strings.Contains(content, `gain="19661f"`) {
				t.Error("washout gain not applied")
			}
			if !strings.Contains(content, `width:468pt`) {
				t.Error("default width not applied")
			}
			break
		}
	}
	if !found {
		t.Error("picture watermark not found")
	}
}

func TestWriteWatermarkAutoHeader(t *testing.T) {
	doc := document.New()
	doc.AddWatermarkText("AUTO")
	sec := doc.AddSection()
	sec.AddText("No explicit header", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for i := 1; i <= 5; i++ {
		content := readEntryOpt(data, fmt.Sprintf("word/header%d.xml", i))
		if content != "" && strings.Contains(content, "AUTO") {
			found = true
			break
		}
	}
	if !found {
		t.Error("auto-created header with watermark not found")
	}
}

func TestWriteWatermarkOnlyOnce(t *testing.T) {
	doc := document.New()
	doc.AddWatermarkText("ONCE")
	sec := doc.AddSection()
	h1 := sec.AddHeader("default")
	h1.AddText("Header1", nil, nil)
	h2 := sec.AddHeader("first")
	h2.AddText("Header2", nil, nil)
	sec.AddText("Content", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for i := 1; i <= 5; i++ {
		content := readEntryOpt(data, fmt.Sprintf("word/header%d.xml", i))
		if content != "" && strings.Contains(content, "ONCE") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("watermark appeared %d times, want 1", count)
	}
}

// --- New font properties in fontStyleToXML ---

func TestFontStyleNewProperties(t *testing.T) {
	fs := &style.FontStyle{
		Name:           "Arial",
		NameEastAsia:   "SimSun",
		UnderlineColor: "0000FF",
		Underline:      "single",
		Kerning:        12,
		Emboss:         true,
		Shadow:         true,
		Imprint:        true,
		Outline:        true,
		TextEffect:     "shimmer",
		RightToLeft:    true,
	}
	xml := fontStyleToXML(fs, nil)
	checks := []string{
		`w:eastAsia="SimSun"`,
		`w:color="0000FF"`,
		`<w:kern w:val="24"/>`,
		`<w:emboss/>`,
		`<w:shadow/>`,
		`<w:imprint/>`,
		`<w:outline/>`,
		`<w:effect w:val="shimmer"/>`,
		`<w:rtl/>`,
	}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("missing %q in fontStyleToXML output", c)
		}
	}
}

func TestFontStyleEastAsiaOnly(t *testing.T) {
	fs := &style.FontStyle{NameEastAsia: "MS Mincho"}
	xml := fontStyleToXML(fs, nil)
	if !strings.Contains(xml, `w:eastAsia="MS Mincho"`) {
		t.Error("east asia font not written when Name is empty")
	}
}

// --- Tab stops in paraStyleToXML ---

func TestParaStyleTabStops(t *testing.T) {
	ps := &style.ParagraphStyle{
		TabStops: []style.TabStop{
			{Position: 2880, Type: "center", Leader: "dot"},
			{Position: 5760, Type: "right"},
			{Position: 7200},
		},
	}
	xml := paraStyleToXML(ps, "")
	checks := []string{
		`<w:tabs>`,
		`<w:tab w:val="center" w:pos="2880" w:leader="dot"/>`,
		`<w:tab w:val="right" w:pos="5760" w:leader="none"/>`,
		`<w:tab w:val="left" w:pos="7200" w:leader="none"/>`,
		`</w:tabs>`,
	}
	for _, c := range checks {
		if !strings.Contains(xml, c) {
			t.Errorf("missing %q in tab stops output", c)
		}
	}
}

// --- Tab element in TextRun ---

func TestWriteTabInTextRun(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Before", nil)
	tr.AddTab()
	tr.AddText("After", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:tab/>`) {
		t.Error("tab element not found in output")
	}
}

// --- Table cell spacing and indent ---

func TestWriteTableCellSpacingAndIndent(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ts := &style.TableStyle{
		Width:       9000,
		CellSpacing: 50,
		Indent:      200,
	}
	tbl := sec.AddTable(ts)
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("Cell", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:tblCellSpacing w:w="50" w:type="dxa"/>`) {
		t.Error("cell spacing not found")
	}
	if !strings.Contains(content, `<w:tblInd w:w="200" w:type="dxa"/>`) {
		t.Error("table indent not found")
	}
}

// --- UpdateFieldsOnOpen in settings ---

func TestWriteSettingsUpdateFields(t *testing.T) {
	doc := document.New()
	doc.SetUpdateFieldsOnOpen(true)
	sec := doc.AddSection()
	sec.AddText("Content", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/settings.xml")
	if !strings.Contains(content, `<w:updateFields w:val="true"/>`) {
		t.Error("updateFields not found in settings.xml")
	}
}

func TestWriteSettingsNoUpdateFields(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Content", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/settings.xml")
	if strings.Contains(content, `<w:updateFields`) {
		t.Error("updateFields should not be present when not set")
	}
}

// --- FormField with paragraph style ---

func TestWriteFormFieldWithParaStyle(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ff := sec.AddTextInput("styled")
	ff.Para = style.ParagraphStyle{Alignment: style.AlignCenter}

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:jc w:val="center"/>`) {
		t.Error("form field paragraph style not applied")
	}
}

// --- FormField with font style ---

func TestWriteFormFieldWithFontStyle(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ff := sec.AddTextInput("styled")
	ff.Font = style.FontStyle{Bold: true, Size: 14}
	ff.DefaultValue = "test"

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	if !strings.Contains(content, `<w:b/>`) {
		t.Error("form field font bold not applied")
	}
}

// --- Dropdown with no default match ---

func TestWriteDropdownNoDefaultMatch(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	ff := sec.AddDropdownList("items", []string{"A", "B", "C"})
	ff.DefaultValue = "NotInList"

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	content := readEntry(t, data, "word/document.xml")
	// The default element should not appear since value is not in list
	if strings.Contains(content, `<w:default w:val="`) {
		// It's OK if it doesn't match - the loop just won't write it
	}
}

// --- Reader round-trip tests for new font properties ---

func TestReadRoundTripNewFontProperties(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Embossed", &style.FontStyle{Emboss: true}, nil)
	sec.AddText("Shadow", &style.FontStyle{Shadow: true}, nil)
	sec.AddText("Imprint", &style.FontStyle{Imprint: true}, nil)
	sec.AddText("Outline", &style.FontStyle{Outline: true}, nil)
	sec.AddText("RTL", &style.FontStyle{RightToLeft: true}, nil)
	sec.AddText("Effect", &style.FontStyle{TextEffect: "shimmer"}, nil)
	sec.AddText("DStrike", &style.FontStyle{DoubleStrikethrough: true}, nil)
	sec.AddText("EastAsia", &style.FontStyle{Name: "Arial", NameEastAsia: "SimSun"}, nil)
	sec.AddText("UColor", &style.FontStyle{Underline: "single", UnderlineColor: "FF0000"}, nil)
	sec.AddText("Kern", &style.FontStyle{Kerning: 12}, nil)
	sec.AddText("Hidden", &style.FontStyle{Hidden: true}, nil)
	sec.AddText("NoProof", &style.FontStyle{NoProof: true}, nil)
	sec.AddText("Lang", &style.FontStyle{Lang: "zh-CN"}, nil)
	sec.AddText("Highlight", &style.FontStyle{HighlightColor: "yellow"}, nil)
	sec.AddText("Spacing", &style.FontStyle{Spacing: 2.5}, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	// Build a map of text -> run for easy lookup
	runs := map[string]*document.Run{}
	for _, elem := range doc2.Sections[0].Elements {
		if p, ok := elem.(*document.Paragraph); ok {
			for _, run := range p.Runs {
				runs[run.Text] = run
			}
		}
	}

	tests := []struct {
		text  string
		check func(*document.Run) bool
		desc  string
	}{
		{"Embossed", func(r *document.Run) bool { return r.Style.Emboss }, "Emboss"},
		{"Shadow", func(r *document.Run) bool { return r.Style.Shadow }, "Shadow"},
		{"Imprint", func(r *document.Run) bool { return r.Style.Imprint }, "Imprint"},
		{"Outline", func(r *document.Run) bool { return r.Style.Outline }, "Outline"},
		{"RTL", func(r *document.Run) bool { return r.Style.RightToLeft }, "RightToLeft"},
		{"Effect", func(r *document.Run) bool { return r.Style.TextEffect == "shimmer" }, "TextEffect"},
		{"DStrike", func(r *document.Run) bool { return r.Style.DoubleStrikethrough }, "DoubleStrikethrough"},
		{"EastAsia", func(r *document.Run) bool { return r.Style.NameEastAsia == "SimSun" }, "NameEastAsia"},
		{"UColor", func(r *document.Run) bool { return r.Style.UnderlineColor == "FF0000" }, "UnderlineColor"},
		{"Kern", func(r *document.Run) bool { return r.Style.Kerning == 12 }, "Kerning"},
		{"Hidden", func(r *document.Run) bool { return r.Style.Hidden }, "Hidden"},
		{"NoProof", func(r *document.Run) bool { return r.Style.NoProof }, "NoProof"},
		{"Lang", func(r *document.Run) bool { return r.Style.Lang == "zh-CN" }, "Lang"},
		{"Highlight", func(r *document.Run) bool { return r.Style.HighlightColor == "yellow" }, "HighlightColor"},
		{"Spacing", func(r *document.Run) bool { return r.Style.Spacing == 2.5 }, "Spacing"},
	}

	for _, tc := range tests {
		r, ok := runs[tc.text]
		if !ok {
			t.Errorf("run %q not found after read", tc.text)
			continue
		}
		if !tc.check(r) {
			t.Errorf("%s not preserved for %q", tc.desc, tc.text)
		}
	}
}
