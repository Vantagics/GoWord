package ooxml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

func TestSaveToFile(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("File save test", nil, nil)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.docx")

	err := Save(doc, path)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("empty file")
	}

	// Read it back
	doc2, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(doc2.Sections) == 0 {
		t.Error("no sections after read")
	}
}

func TestSaveToSubdirectory(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Subdir test", nil, nil)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "sub", "dir", "test.docx")

	err := Save(doc, path)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestReadNonExistentFile(t *testing.T) {
	_, err := Read("nonexistent.docx")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadBreakElement(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tr := sec.AddTextRun(nil)
	tr.AddText("Before", nil)
	tr.AddTextBreak(1)
	tr.AddText("After", nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	// Should have read back the paragraph
	if len(doc2.Sections[0].Elements) == 0 {
		t.Error("no elements")
	}
}

func TestReadMultipleTables(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()

	tbl1 := sec.AddTable(nil)
	r1 := tbl1.AddRow(0, nil)
	r1.AddCell(3000, nil).AddText("Table 1", nil, nil)

	sec.AddText("Between tables", nil, nil)

	tbl2 := sec.AddTable(nil)
	r2 := tbl2.AddRow(0, nil)
	r2.AddCell(3000, nil).AddText("Table 2", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	tableCount := 0
	for _, elem := range doc2.Sections[0].Elements {
		if _, ok := elem.(*document.Table); ok {
			tableCount++
		}
	}
	if tableCount != 2 {
		t.Errorf("tables = %d, want 2", tableCount)
	}
}

func TestReadCellWidth(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)
	row := tbl.AddRow(0, nil)
	row.AddCell(5000, nil).AddText("Wide", nil, nil)

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
			if tbl.Rows[0].Cells[0].Style.Width != 5000 {
				t.Errorf("width = %d, want 5000", tbl.Rows[0].Cells[0].Style.Width)
			}
		}
	}
}

func TestReadVMergeContinue(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	tbl := sec.AddTable(nil)

	row1 := tbl.AddRow(0, nil)
	row1.AddCell(3000, &style.CellStyle{VMerge: "restart"}).AddText("Start", nil, nil)

	row2 := tbl.AddRow(0, nil)
	row2.AddCell(3000, &style.CellStyle{VMerge: "continue"})

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
			if len(tbl.Rows) != 2 {
				t.Fatalf("rows = %d", len(tbl.Rows))
			}
			if tbl.Rows[0].Cells[0].Style.VMerge != "restart" {
				t.Errorf("row0 vMerge = %q", tbl.Rows[0].Cells[0].Style.VMerge)
			}
			if tbl.Rows[1].Cells[0].Style.VMerge != "continue" {
				t.Errorf("row1 vMerge = %q", tbl.Rows[1].Cells[0].Style.VMerge)
			}
		}
	}
}

func TestReadSectionColumns(t *testing.T) {
	doc := document.New()
	ss := style.DefaultSectionStyle()
	ss.ColumnCount = 3
	sec := doc.AddSectionWithStyle(ss)
	sec.AddText("Three columns", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	if doc2.Sections[0].Style.ColumnCount != 3 {
		t.Errorf("columns = %d", doc2.Sections[0].Style.ColumnCount)
	}
}

func TestReadDocumentProperties(t *testing.T) {
	doc := document.New()
	doc.Properties.Title = "Props Test"
	doc.Properties.Subject = "Subject"
	doc.Properties.Creator = "Creator"
	doc.Properties.Keywords = "key1, key2"
	doc.Properties.Description = "Desc"
	doc.Properties.Category = "Cat"
	doc.Properties.LastModifiedBy = "Modifier"
	doc.Properties.Revision = 5

	sec := doc.AddSection()
	sec.AddText("Test", nil, nil)

	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2, err := ReadFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	p := doc2.Properties
	if p.Title != "Props Test" {
		t.Errorf("title = %q", p.Title)
	}
	if p.Subject != "Subject" {
		t.Errorf("subject = %q", p.Subject)
	}
	if p.Creator != "Creator" {
		t.Errorf("creator = %q", p.Creator)
	}
	if p.Keywords != "key1, key2" {
		t.Errorf("keywords = %q", p.Keywords)
	}
	if p.Description != "Desc" {
		t.Errorf("description = %q", p.Description)
	}
	if p.Category != "Cat" {
		t.Errorf("category = %q", p.Category)
	}
	if p.LastModifiedBy != "Modifier" {
		t.Errorf("lastModifiedBy = %q", p.LastModifiedBy)
	}
}

func TestReadUnderlineType(t *testing.T) {
	doc := document.New()
	sec := doc.AddSection()
	sec.AddText("Wave underline", &style.FontStyle{Underline: "wave"}, nil)

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
				if run.Text == "Wave underline" {
					if run.Style.Underline != "wave" {
						t.Errorf("underline = %q, want wave", run.Style.Underline)
					}
				}
			}
		}
	}
}
