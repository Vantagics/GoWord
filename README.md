# GoWord

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Zero Dependencies](https://img.shields.io/badge/dependencies-zero-brightgreen.svg)]()

A pure Go library for reading and writing Microsoft Word documents (.docx / Office Open XML). Inspired by [PHPWord](https://github.com/PHPOffice/PHPWord), with zero external dependencies.

## Features

- Create and read `.docx` files
- Document properties (title, creator, subject, keywords, etc.)
- Sections with page size, orientation, margins, columns, page numbering
- Paragraphs with rich font styles (bold, italic, underline, color, size, etc.)
- Paragraph formatting (alignment, spacing, indent, borders, shading)
- Headings (Title, Heading1–9)
- Hyperlinks
- Tables with cell spanning (gridSpan, vMerge), borders, shading, nested tables
- Lists with custom numbering styles (bullets, decimal, roman, etc.)
- Images from file path or byte data (PNG, JPEG, GIF, BMP, TIFF)
- Headers and footers with preserve text (page numbers)
- Footnotes and endnotes
- Comments, bookmarks, table of contents
- Checkboxes, line shapes
- Watermarks (text and picture)
- Form fields (text input, dropdown, checkbox)
- Tab stops and tab characters
- Named styles (font, paragraph, table, numbering)
- TextRun for mixed-format inline content
- ExtractText for full document text extraction
- Document-level operations (Paragraphs, Tables, InsertParagraph, RemoveParagraph)
- Field code constants (PAGE, NUMPAGES, DATE, AUTHOR, etc.)
- UpdateFieldsOnOpen setting
- Unit conversion utilities (twip, cm, inch, pt, EMU, pixel)

## Installation

```bash
go get github.com/VantageDataChat/GoWord
```

## Quick Start

```go
package main

import (
    "log"

    goword "github.com/VantageDataChat/GoWord"
    "github.com/VantageDataChat/GoWord/style"
)

func main() {
    doc := goword.New()
    doc.Properties.Title = "My Document"
    doc.Properties.Creator = "GoWord"

    sec := doc.AddSection()
    sec.AddTitle("Hello GoWord", 1)
    sec.AddText("This is a paragraph with bold text.",
        &style.FontStyle{Bold: true, Size: 12, Color: "333333"}, nil)

    if err := doc.Save("hello.docx"); err != nil {
        log.Fatal(err)
    }
}
```

## Reading a Document

```go
doc, err := goword.Open("existing.docx")
if err != nil {
    log.Fatal(err)
}
for _, sec := range doc.Sections {
    for _, elem := range sec.Elements {
        // Process elements...
    }
}
```


## Examples

### Tables

```go
sec := doc.AddSection()
ts := &style.TableStyle{Width: 9000, Alignment: "center"}
ts.SetAllBorders("single", 4, "000000")
tbl := sec.AddTable(ts)
tbl.Grid = []int{3000, 3000, 3000}

row := tbl.AddRow(0, &style.RowStyle{IsHeader: true})
row.AddCell(3000, nil).AddText("Name", &style.FontStyle{Bold: true}, nil)
row.AddCell(3000, nil).AddText("Age", &style.FontStyle{Bold: true}, nil)
row.AddCell(3000, nil).AddText("City", &style.FontStyle{Bold: true}, nil)

row2 := tbl.AddRow(0, nil)
row2.AddCell(3000, nil).AddText("Alice", nil, nil)
row2.AddCell(3000, nil).AddText("30", nil, nil)
row2.AddCell(3000, nil).AddText("Beijing", nil, nil)
```

### Images

```go
// From file
sec.AddImage("photo.png", &style.ImageStyle{Width: 200, Height: 150})

// From bytes
sec.AddImageFromBytes(pngData, "image/png", &style.ImageStyle{Width: 100, Height: 100})
```

### Headers, Footers & Page Numbers

```go
header := sec.AddHeader("")
header.AddText("Company Name", &style.FontStyle{Bold: true}, nil)

footer := sec.AddFooter("")
footer.AddPreserveText("Page {PAGE} of {NUMPAGES}", nil,
    &style.ParagraphStyle{Alignment: style.AlignCenter})
```

### Lists

```go
doc.AddNumberingStyle("bullets", goword.NumberingStyle{
    Type: "singleLevel",
    Levels: []goword.NumberingLevel{
        {Format: "bullet", Text: "\u2022", Left: 360, Hanging: 360, Font: "Symbol"},
    },
})
sec.AddListItem("First item", 0, nil, "bullets", nil)
sec.AddListItem("Second item", 0, nil, "bullets", nil)
```

### TextRun (Mixed Inline Content)

```go
tr := sec.AddTextRun(nil)
tr.AddText("Normal text ", nil)
tr.AddText("bold text ", &style.FontStyle{Bold: true})
tr.AddLink("https://example.com", "a link", nil)
fn := tr.AddFootnote()
fn.AddText("Footnote content", nil)
```

### Footnotes & Endnotes

```go
fn := sec.AddFootnote()
fn.AddText("See reference.", nil)
fn.AddLink("https://example.com", "Source")

en := sec.AddEndnote()
en.AddText("End note text.", nil)
en.AddLink("https://example.com", "Reference")
en.AddTextBreak()
en.AddText("Second line.", nil)
```

### Sections & Page Layout

```go
// Landscape A4
ss := style.DefaultSectionStyle()
ss.Orientation = style.OrientLandscape
ss.PageWidth = 16838
ss.PageHeight = 11906
sec := doc.AddSectionWithStyle(ss)

// Two columns
ss2 := style.DefaultSectionStyle()
ss2.ColumnCount = 2
ss2.ColumnSpacing = 720
```

### Named Styles

```go
doc.AddFontStyle("emphasis", style.FontStyle{Italic: true, Color: "0000FF"})
doc.AddParagraphStyle("centered", style.ParagraphStyle{Alignment: style.AlignCenter})

sec.AddTextWithStyle("Styled text", "emphasis", "centered")
```

### Watermarks

```go
// Text watermark
wm := doc.AddWatermarkText("DRAFT")
wm.Font = "Arial"
wm.Color = "C0C0C0"
wm.Bold = true

// Picture watermark from bytes
wm2 := doc.AddWatermarkPictureFromBytes(pngData, "image/png")
wm2.Washout = true
```

### Form Fields

```go
// Text input
ff := sec.AddTextInput("username")
ff.DefaultValue = "Enter name"
ff.MaxLength = 50

// Dropdown list
dd := sec.AddDropdownList("color", []string{"Red", "Green", "Blue"})
dd.DefaultValue = "Green"

// Checkbox form field
cb := sec.AddFormField(goword.FormFieldTypeCheckBox, "agree")
cb.Value = "true"
```

### Tab Stops

```go
ps := &style.ParagraphStyle{
    TabStops: []style.TabStop{
        {Position: 2880, Type: "center", Leader: "dot"},
        {Position: 5760, Type: "right"},
    },
}
tr := sec.AddTextRun(ps)
tr.AddText("Left", nil)
tr.AddTab()
tr.AddText("Center", nil)
tr.AddTab()
tr.AddText("Right", nil)
```

### Extract Text

```go
doc, _ := goword.Open("document.docx")
text := doc.ExtractText()
fmt.Println(text)
```

### Document-Level Operations

```go
// Get all paragraphs/tables across sections
paragraphs := doc.Paragraphs()
tables := doc.Tables()

// Insert/remove paragraphs
newP := doc.InsertParagraphAfter(paragraphs[0])
doc.RemoveParagraph(paragraphs[1])

// Auto-update fields on open
doc.SetUpdateFieldsOnOpen(true)
```

## Project Structure

```
github.com/VantageDataChat/GoWord
├── goword.go              # Top-level API: New(), Open(), type aliases
├── common/
│   ├── properties.go      # DocProperties
│   └── units.go           # Unit conversions (twip, cm, inch, pt, EMU, pixel)
├── style/
│   ├── font.go            # FontStyle
│   ├── paragraph.go       # ParagraphStyle, Border, Shading
│   ├── table.go           # TableStyle, RowStyle, CellStyle
│   ├── section.go         # SectionStyle, paper sizes
│   └── image.go           # ImageStyle
├── document/
│   ├── document.go        # Document struct, NumberingStyle
│   ├── document_api.go    # Document methods (AddSection, AddStyle, etc.)
│   ├── elements.go        # All element types (Paragraph, Table, Image, etc.)
│   ├── section.go         # Section.Add* methods
│   ├── table.go           # Table/Row/Cell methods
│   ├── textrun.go         # TextRun.Add* methods
│   ├── headerfooter.go    # Header/Footer/Footnote/Endnote methods
│   └── io.go              # Save/Open with pluggable backend
└── ooxml/
    ├── reader.go          # Read/ReadFromBytes
    ├── writer.go          # Save/WriteToBytes/WriteToWriter
    └── ...                # OOXML XML generation and parsing
```

## Testing

```bash
go test ./... -race
```

374 tests, all passing with race detection. Coverage:

| Package    | Coverage |
|------------|----------|
| `common`   | 100%     |
| `style`    | 100%     |
| `document` | 100%     |
| `ooxml`    | 96.1%    |

## License

MIT
