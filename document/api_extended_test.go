package document

import (
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/style"
)

func TestExtractText(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Hello World", nil, nil)
	sec.AddText("Second line", nil, nil)

	text := doc.ExtractText()
	if !strings.Contains(text, "Hello World") {
		t.Error("missing 'Hello World'")
	}
	if !strings.Contains(text, "Second line") {
		t.Error("missing 'Second line'")
	}
}

func TestExtractTextWithTable(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	row.AddCell(3000, nil).AddText("A", nil, nil)
	row.AddCell(3000, nil).AddText("B", nil, nil)

	text := doc.ExtractText()
	if !strings.Contains(text, "A") || !strings.Contains(text, "B") {
		t.Error("missing table cell text")
	}
}

func TestExtractTextWithHyperlink(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddLink("https://example.com", "Click here", nil, nil)

	text := doc.ExtractText()
	if !strings.Contains(text, "Click here") {
		t.Error("missing hyperlink text")
	}
}

func TestExtractTextWithListItem(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddListItem("Item 1", 0, nil, "", nil)

	text := doc.ExtractText()
	if !strings.Contains(text, "Item 1") {
		t.Error("missing list item text")
	}
}

func TestExtractTextWithCheckBox(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddCheckBox("cb", "Check me", nil, nil)

	text := doc.ExtractText()
	if !strings.Contains(text, "Check me") {
		t.Error("missing checkbox text")
	}
}

func TestExtractTextWithTextBreak(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Before", nil, nil)
	sec.AddTextBreak(2)
	sec.AddText("After", nil, nil)

	text := doc.ExtractText()
	if !strings.Contains(text, "Before") || !strings.Contains(text, "After") {
		t.Error("missing text around break")
	}
}

func TestExtractTextWithTextRun(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Part1 ", nil)
	tr.AddText("Part2", nil)

	text := doc.ExtractText()
	if !strings.Contains(text, "Part1") || !strings.Contains(text, "Part2") {
		t.Error("missing textrun text")
	}
}

func TestParagraphs(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("P1", nil, nil)
	sec.AddText("P2", nil, nil)
	sec.AddTable(nil) // not a paragraph

	paras := doc.Paragraphs()
	if len(paras) != 2 {
		t.Errorf("Paragraphs() = %d, want 2", len(paras))
	}
}

func TestTables(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("text", nil, nil)
	sec.AddTable(nil)
	sec.AddTable(nil)

	tables := doc.Tables()
	if len(tables) != 2 {
		t.Errorf("Tables() = %d, want 2", len(tables))
	}
}

func TestRemoveParagraph(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	p1 := sec.AddText("Keep", nil, nil)
	p2 := sec.AddText("Remove", nil, nil)
	_ = p1

	doc.RemoveParagraph(p2)
	if len(sec.Elements) != 1 {
		t.Errorf("elements = %d, want 1", len(sec.Elements))
	}
}

func TestInsertParagraphBefore(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	p1 := sec.AddText("Original", nil, nil)

	np := doc.InsertParagraphBefore(p1)
	if np == nil {
		t.Fatal("InsertParagraphBefore returned nil")
	}
	if len(sec.Elements) != 2 {
		t.Errorf("elements = %d, want 2", len(sec.Elements))
	}
	if sec.Elements[0] != np {
		t.Error("new paragraph not at position 0")
	}
}

func TestInsertParagraphAfter(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	p1 := sec.AddText("Original", nil, nil)

	np := doc.InsertParagraphAfter(p1)
	if np == nil {
		t.Fatal("InsertParagraphAfter returned nil")
	}
	if len(sec.Elements) != 2 {
		t.Errorf("elements = %d, want 2", len(sec.Elements))
	}
	if sec.Elements[1] != np {
		t.Error("new paragraph not at position 1")
	}
}

func TestInsertParagraphBeforeNotFound(t *testing.T) {
	doc := New()
	doc.AddSection()
	fake := &Paragraph{}
	np := doc.InsertParagraphBefore(fake)
	if np != nil {
		t.Error("expected nil for non-existent paragraph")
	}
}

func TestInsertParagraphAfterNotFound(t *testing.T) {
	doc := New()
	doc.AddSection()
	fake := &Paragraph{}
	np := doc.InsertParagraphAfter(fake)
	if np != nil {
		t.Error("expected nil for non-existent paragraph")
	}
}

func TestWatermarkText(t *testing.T) {
	doc := New()
	wm := doc.AddWatermarkText("DRAFT")
	wm.Bold = true
	wm.Color = "FF0000"

	if doc.WatermarkTextValue() == nil {
		t.Fatal("watermark text not set")
	}
	if doc.WatermarkTextValue().Text != "DRAFT" {
		t.Error("wrong watermark text")
	}
	if doc.WatermarkPictureValue() != nil {
		t.Error("picture watermark should be nil")
	}
}

func TestWatermarkPicture(t *testing.T) {
	doc := New()
	wm := doc.AddWatermarkPicture("logo.png")
	wm.Width = 200
	wm.Height = 200

	if doc.WatermarkPictureValue() == nil {
		t.Fatal("watermark picture not set")
	}
	if doc.WatermarkTextValue() != nil {
		t.Error("text watermark should be nil")
	}
}

func TestWatermarkPictureFromBytes(t *testing.T) {
	doc := New()
	wm := doc.AddWatermarkPictureFromBytes([]byte{0x89, 0x50}, "image/png")
	wm.Washout = false

	if doc.WatermarkPictureValue() == nil {
		t.Fatal("watermark picture not set")
	}
}

func TestWatermarkMutualExclusion(t *testing.T) {
	doc := New()
	doc.AddWatermarkText("DRAFT")
	doc.AddWatermarkPicture("logo.png")

	if doc.WatermarkTextValue() != nil {
		t.Error("text watermark should be cleared when picture is set")
	}
	if doc.WatermarkPictureValue() == nil {
		t.Error("picture watermark should be set")
	}
}

func TestUpdateFieldsOnOpen(t *testing.T) {
	doc := New()
	if doc.UpdateFieldsOnOpen() {
		t.Error("default should be false")
	}
	doc.SetUpdateFieldsOnOpen(true)
	if !doc.UpdateFieldsOnOpen() {
		t.Error("should be true after setting")
	}
}

func TestFormFieldTextInput(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ff := sec.AddTextInput("name")
	ff.DefaultValue = "John"
	ff.MaxLength = 50

	if ff.Type != FormFieldTypeText {
		t.Error("wrong type")
	}
	if ff.Name != "name" {
		t.Error("wrong name")
	}
}

func TestFormFieldDropDown(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ff := sec.AddDropdownList("color", []string{"Red", "Green", "Blue"})
	ff.DefaultValue = "Green"

	if ff.Type != FormFieldTypeDropDown {
		t.Error("wrong type")
	}
	if len(ff.PossibleValues) != 3 {
		t.Errorf("values = %d, want 3", len(ff.PossibleValues))
	}
}

func TestFormFieldGeneric(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ff := sec.AddFormField(FormFieldTypeCheckBox, "agree")
	ff.Value = "true"

	if ff.Type != FormFieldTypeCheckBox {
		t.Error("wrong type")
	}
}

func TestTextRunAddTab(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Before", nil)
	tr.AddTab()
	tr.AddText("After", nil)

	if len(tr.Elements) != 3 {
		t.Errorf("elements = %d, want 3", len(tr.Elements))
	}
	if _, ok := tr.Elements[1].(*Tab); !ok {
		t.Error("second element should be Tab")
	}
}

func TestFieldConstants(t *testing.T) {
	if FieldCurrentPage != "PAGE" {
		t.Error("FieldCurrentPage")
	}
	if FieldNumberOfPages != "NUMPAGES" {
		t.Error("FieldNumberOfPages")
	}
	if FieldDate != "DATE" {
		t.Error("FieldDate")
	}
	if FieldCreateDate != "CREATEDATE" {
		t.Error("FieldCreateDate")
	}
	if FieldEditTime != "EDITTIME" {
		t.Error("FieldEditTime")
	}
	if FieldPrintDate != "PRINTDATE" {
		t.Error("FieldPrintDate")
	}
	if FieldSaveDate != "SAVEDATE" {
		t.Error("FieldSaveDate")
	}
	if FieldTime != "TIME" {
		t.Error("FieldTime")
	}
	if FieldFileName != "FILENAME" {
		t.Error("FieldFileName")
	}
	if FieldAuthor != "AUTHOR" {
		t.Error("FieldAuthor")
	}
	if FieldTitle != "TITLE" {
		t.Error("FieldTitle")
	}
	if FieldSubject != "SUBJECT" {
		t.Error("FieldSubject")
	}
}

func TestAllocPublicMethods(t *testing.T) {
	doc := New()
	rid := doc.AllocRelIDPublic()
	if rid == "" {
		t.Error("empty rel ID")
	}
	iid := doc.AllocImageIDPublic()
	if iid == 0 {
		t.Error("zero image ID")
	}
}

func TestAppendImage(t *testing.T) {
	doc := New()
	img := &Image{Source: "test.png", RelID: "rId1", ID: 1}
	doc.AppendImage(img)
	if len(doc.Images()) != 1 {
		t.Error("image not appended")
	}
}

func TestNewElementTypes(t *testing.T) {
	// Test elementType() for new types
	wt := &WatermarkText{Text: "DRAFT"}
	if wt.elementType() != "watermarktext" {
		t.Error("WatermarkText elementType")
	}
	wp := &WatermarkPicture{Source: "logo.png"}
	if wp.elementType() != "watermarkpicture" {
		t.Error("WatermarkPicture elementType")
	}
	ff := &FormField{Name: "test"}
	if ff.elementType() != "formfield" {
		t.Error("FormField elementType")
	}
	tab := &Tab{}
	if tab.elementType() != "tab" {
		t.Error("Tab elementType")
	}
}

func TestFontStyleNewProperties(t *testing.T) {
	fs := style.FontStyle{
		Emboss:        true,
		Shadow:        true,
		Imprint:       true,
		Outline:       true,
		TextEffect:    "shimmer",
		RightToLeft:   true,
		NameEastAsia:  "SimSun",
		UnderlineColor: "FF0000",
		Kerning:       12,
	}
	if !fs.Emboss || !fs.Shadow || !fs.Imprint || !fs.Outline {
		t.Error("font style properties not set")
	}
	if fs.TextEffect != "shimmer" {
		t.Error("TextEffect")
	}
	if fs.RightToLeft != true {
		t.Error("RightToLeft")
	}
	if fs.NameEastAsia != "SimSun" {
		t.Error("NameEastAsia")
	}
}

func TestTableStyleNewProperties(t *testing.T) {
	ts := style.TableStyle{
		CellSpacing: 20,
		Indent:      100,
	}
	if ts.CellSpacing != 20 {
		t.Error("CellSpacing")
	}
	if ts.Indent != 100 {
		t.Error("Indent")
	}
}
