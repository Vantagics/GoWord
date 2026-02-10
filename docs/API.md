# GoWord API Reference

> Repository: https://github.com/VantageDataChat/GoWord

## Top-Level Functions

```go
import goword "github.com/VantageDataChat/GoWord"
```

| Function | Description |
|----------|-------------|
| `goword.New() *Document` | Create a new empty document |
| `goword.Open(path string) (*Document, error)` | Read a .docx file |
| `goword.OpenFromBytes(data []byte) (*Document, error)` | Read a .docx from bytes |

## Document

```go
type Document struct {
    Properties  DocProperties
    Sections    []*Section
    DefaultFont FontStyle
}
```

### Methods

| Method | Description |
|--------|-------------|
| `Save(path string) error` | Write to .docx file |
| `ToBytes() ([]byte, error)` | Write to byte slice |
| `AddSection() *Section` | Add section with default A4 portrait style |
| `AddSectionWithStyle(s SectionStyle) *Section` | Add section with custom style |
| `SetDefaultFontName(name string)` | Set default font family |
| `SetDefaultFontSize(size float64)` | Set default font size (points) |
| `AddFontStyle(name string, fs FontStyle)` | Register named font style |
| `AddParagraphStyle(name string, ps ParagraphStyle)` | Register named paragraph style |
| `AddTableStyle(name string, ts TableStyle)` | Register named table style |
| `AddNumberingStyle(name string, ns NumberingStyle)` | Register named numbering style |
| `GetFontStyle(name string) (FontStyle, bool)` | Retrieve named font style |
| `GetParagraphStyle(name string) (ParagraphStyle, bool)` | Retrieve named paragraph style |
| `GetTableStyle(name string) (TableStyle, bool)` | Retrieve named table style |
| `GetNumberingStyle(name string) (NumberingStyle, bool)` | Retrieve named numbering style |
| `FontStyles() map[string]FontStyle` | All registered font styles |
| `ParagraphStyles() map[string]ParagraphStyle` | All registered paragraph styles |
| `TableStyles() map[string]TableStyle` | All registered table styles |
| `NumberingStyles() map[string]NumberingStyle` | All registered numbering styles |
| `Footnotes() []*Footnote` | All footnotes |
| `Endnotes() []*Endnote` | All endnotes |
| `Comments() []*Comment` | All comments |
| `Images() []*Image` | All images |
| `Paragraphs() []*Paragraph` | All paragraphs across all sections (flattened) |
| `Tables() []*Table` | All tables across all sections (flattened) |
| `ExtractText() string` | Extract all text content as a single string |
| `RemoveParagraph(p *Paragraph)` | Remove a paragraph from the document |
| `InsertParagraphBefore(p *Paragraph) *Paragraph` | Insert empty paragraph before given one |
| `InsertParagraphAfter(p *Paragraph) *Paragraph` | Insert empty paragraph after given one |
| `AddWatermarkText(text string) *WatermarkText` | Add text watermark |
| `AddWatermarkPicture(src string) *WatermarkPicture` | Add picture watermark from file |
| `AddWatermarkPictureFromBytes(data []byte, mimeType string) *WatermarkPicture` | Add picture watermark from bytes |
| `SetUpdateFieldsOnOpen(b bool)` | Auto-update fields (TOC, page numbers) on open |
| `UpdateFieldsOnOpen() bool` | Check if fields update on open |

## Section

### Methods

| Method | Description |
|--------|-------------|
| `AddText(text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *Paragraph` | Add text paragraph |
| `AddTextWithStyle(text, fontStyleName, paraStyleName string) *Paragraph` | Add text using named styles |
| `AddTextRun(paraStyle *ParagraphStyle) *TextRun` | Add mixed-format paragraph |
| `AddTextBreak(count int)` | Add empty lines |
| `AddPageBreak()` | Add page break |
| `AddLink(url, text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *Hyperlink` | Add hyperlink |
| `AddTitle(text string, depth int) *Paragraph` | Add heading (0=Title, 1–9=Heading1–9) |
| `AddTable(tableStyle *TableStyle) *Table` | Add table |
| `AddTableWithStyle(styleName string) *Table` | Add table with named style |
| `AddImage(src string, imgStyle *ImageStyle) *Image` | Add image from file |
| `AddImageFromBytes(data []byte, mimeType string, imgStyle *ImageStyle) *Image` | Add image from bytes |
| `AddListItem(text string, depth int, fontStyle *FontStyle, listStyleName string, paraStyle *ParagraphStyle) *ListItem` | Add list item |
| `AddFootnote() *Footnote` | Add footnote |
| `AddEndnote() *Endnote` | Add endnote |
| `AddTOC(fontStyle *FontStyle, minDepth, maxDepth int) *TOC` | Add table of contents |
| `AddCheckBox(name, text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *CheckBox` | Add checkbox |
| `AddLine(lineStyle Line) *Line` | Add line shape |
| `AddBookmark(name string) *Bookmark` | Add bookmark |
| `AddComment(author, text string) *Comment` | Add comment |
| `AddFormField(fieldType FormFieldType, name string) *FormField` | Add form field |
| `AddTextInput(name string) *FormField` | Add text input form field |
| `AddDropdownList(name string, values []string) *FormField` | Add dropdown list form field |
| `AddHeader(headerType string) *Header` | Add header ("default", "first", "even") |
| `AddFooter(footerType string) *Footer` | Add footer ("default", "first", "even") |

## TextRun

Mixed-format inline paragraph. Supports combining text, links, images, footnotes, and bookmarks in a single paragraph.

| Method | Description |
|--------|-------------|
| `AddText(text string, fontStyle *FontStyle) *Run` | Add text |
| `AddTextWithStyle(text, styleName string) *Run` | Add text with named style |
| `AddLink(url, text string, fontStyle *FontStyle) *Hyperlink` | Add inline hyperlink |
| `AddTextBreak(count int)` | Add line break |
| `AddImage(src string, imgStyle *ImageStyle) *Image` | Add inline image |
| `AddFootnote() *Footnote` | Add footnote reference |
| `AddEndnote() *Endnote` | Add endnote reference |
| `AddBookmark(name string) *Bookmark` | Add bookmark |
| `AddTab()` | Add tab character |

## Table / Row / Cell

### Table

| Method | Description |
|--------|-------------|
| `AddRow(height int, rowStyle *RowStyle) *Row` | Add row |

Table also has a `Grid []int` field for column widths (twips).

### Row

| Method | Description |
|--------|-------------|
| `AddCell(width int, cellStyle *CellStyle) *Cell` | Add cell |

### Cell

| Method | Description |
|--------|-------------|
| `AddText(text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *Paragraph` | Add text |
| `AddTextRun(paraStyle *ParagraphStyle) *TextRun` | Add mixed-format paragraph |
| `AddImage(src string, imgStyle *ImageStyle) *Image` | Add image |
| `AddTable(tableStyle *TableStyle) *Table` | Add nested table |
| `AddListItem(text string, depth int, fontStyle *FontStyle, listStyleName string) *ListItem` | Add list item |

## Header / Footer

Both `Header` and `Footer` support:

| Method | Description |
|--------|-------------|
| `AddText(text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *Paragraph` | Add text |
| `AddPreserveText(text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *PreserveText` | Add field codes like `{PAGE}`, `{NUMPAGES}` |
| `AddImage(src string, imgStyle *ImageStyle) *Image` | Add image |
| `AddTable(tableStyle *TableStyle) *Table` | Add table |

## Footnote / Endnote

Both `Footnote` and `Endnote` support:

| Method | Description |
|--------|-------------|
| `AddText(text string, fontStyle *FontStyle)` | Add text |
| `AddLink(url, text string) *Hyperlink` | Add hyperlink |
| `AddTextBreak()` | Add line break |


## Style Types

### FontStyle

```go
type FontStyle struct {
    Name                string  // Font family, e.g. "Arial"
    NameEastAsia        string  // East Asian font family, e.g. "SimSun"
    Size                float64 // Size in points
    Bold                bool
    Italic              bool
    Underline           string  // "single", "double", "wave", "dash", "dotted", "none"
    UnderlineColor      string  // Underline color (hex), e.g. "FF0000"
    Strikethrough       bool
    DoubleStrikethrough bool
    SuperScript         bool
    SubScript           bool
    Color               string  // Hex without '#', e.g. "FF0000"
    HighlightColor      string  // e.g. "yellow"
    AllCaps             bool
    SmallCaps           bool
    Hidden              bool
    Spacing             float64 // Character spacing in points
    Kerning             float64 // Font kerning in points (0 = off)
    NoProof             bool
    Lang                string  // e.g. "en-US"
    Emboss              bool    // Embossed text effect
    Shadow              bool    // Shadow text effect
    Imprint             bool    // Imprinted (engraved) text effect
    Outline             bool    // Outline text effect
    TextEffect          string  // "blinkBackground", "lights", "antsBlack", "antsRed", "shimmer", "sparkle"
    RightToLeft         bool    // Right-to-left text direction
}
```

### ParagraphStyle

```go
type ParagraphStyle struct {
    Alignment       string // "left", "center", "right", "both", "distribute"
    SpaceBefore     int    // Twips
    SpaceAfter      int    // Twips
    LineSpacing     int    // Twips (240 = single)
    LineRule        string // "auto", "exact", "atLeast"
    Indent          int    // Left indent (twips)
    IndentRight     int    // Right indent (twips)
    FirstLine       int    // First line indent (twips)
    Hanging         int    // Hanging indent (twips)
    KeepNext        bool
    KeepLines       bool
    PageBreakBefore bool
    WidowControl    bool
    Borders         *ParagraphBorders
    Shading         *Shading
    TabStops        []TabStop // Custom tab stops
}

type TabStop struct {
    Position int    // Position in twips
    Type     string // "left", "center", "right", "decimal", "bar"
    Leader   string // "none", "dot", "hyphen", "underscore"
}
```

### TableStyle

```go
type TableStyle struct {
    Width           int    // Twips (0 = auto)
    WidthType       string // "auto", "dxa", "pct"
    Alignment       string // "left", "center", "right"
    CellMarginTop   int
    CellMarginBottom int
    CellMarginLeft  int
    CellMarginRight int
    CellSpacing     int    // Cell spacing in twips (0 = none)
    BorderTop       *Border
    BorderBottom    *Border
    BorderLeft      *Border
    BorderRight     *Border
    BorderInsideH   *Border
    BorderInsideV   *Border
    Layout          string // "fixed", "autofit"
    Indent          int    // Table indent from leading margin in twips
}

// Convenience method
func (ts *TableStyle) SetAllBorders(style string, size int, color string)
```

### RowStyle

```go
type RowStyle struct {
    Height     int    // Twips
    HeightRule string // "auto", "exact", "atLeast"
    IsHeader   bool   // Repeat as header row
    CantSplit  bool   // Don't split across pages
}
```

### CellStyle

```go
type CellStyle struct {
    Width         int    // Twips
    WidthType     string // "auto", "dxa", "pct"
    GridSpan      int    // Columns to span
    VMerge        string // "restart", "continue", ""
    VAlign        string // "top", "center", "bottom"
    Shading       *Shading
    BorderTop     *Border
    BorderBottom  *Border
    BorderLeft    *Border
    BorderRight   *Border
    TextDirection string // "lrTb", "tbRl", "btLr"
    NoWrap        bool
}
```

### SectionStyle

```go
type SectionStyle struct {
    Orientation     string // "portrait", "landscape"
    PageWidth       int    // Twips
    PageHeight      int    // Twips
    MarginTop       int
    MarginBottom    int
    MarginLeft      int
    MarginRight     int
    HeaderHeight    int
    FooterHeight    int
    ColumnCount     int
    ColumnSpacing   int
    PageNumberStart *int   // nil = continue
    BreakType       string // "nextPage", "continuous", "evenPage", "oddPage"
}

func DefaultSectionStyle() SectionStyle  // A4 portrait, 1-inch margins
func (s *SectionStyle) SetPaperSize(name string) // "A4", "A3", "A5", "Letter", "Legal"
```

### ImageStyle

```go
type ImageStyle struct {
    Width         int    // Points
    Height        int    // Points
    Alignment     string // "left", "center", "right"
    WrappingStyle string // "inline", "behind", "inFrontOf", "square", "tight"
    MarginTop     int
    MarginBottom  int
    MarginLeft    int
    MarginRight   int
}
```

### NumberingStyle

```go
type NumberingStyle struct {
    Type   string           // "multilevel", "singleLevel"
    Levels []NumberingLevel
}

type NumberingLevel struct {
    Format  string // "decimal", "upperLetter", "lowerLetter", "upperRoman", "lowerRoman", "bullet"
    Text    string // e.g. "%1." or "\u2022"
    Left    int    // Left indent (twips)
    Hanging int    // Hanging indent (twips)
    TabPos  int    // Tab position (twips)
    Font    string // Font name for bullet character
}
```

### DocProperties

```go
type DocProperties struct {
    Title          string
    Subject        string
    Creator        string
    Keywords       string
    Description    string
    Category       string
    LastModifiedBy string
    Company        string
    Manager        string
    Created        time.Time
    Modified       time.Time
    Revision       int
}
```

## Unit Conversion Utilities

```go
import "github.com/VantageDataChat/GoWord/common"
```

| Function | Description |
|----------|-------------|
| `common.InchToTwip(inches float64) int` | Inches → twips |
| `common.CmToTwip(cm float64) int` | Centimeters → twips |
| `common.PointToTwip(pt float64) int` | Points → twips |
| `common.TwipToInch(twip int) float64` | Twips → inches |
| `common.TwipToCm(twip int) float64` | Twips → centimeters |
| `common.TwipToPoint(twip int) float64` | Twips → points |
| `common.EmuToTwip(emu int64) int` | EMU → twips |
| `common.TwipToEmu(twip int) int64` | Twips → EMU |
| `common.PointToHalfPoint(pt float64) int` | Points → half-points (font sizes) |
| `common.HalfPointToPoint(hp int) float64` | Half-points → points |
| `common.EmuToPixel(emu int64) int` | EMU → pixels (96 DPI) |
| `common.PixelToEmu(px int) int64` | Pixels → EMU (96 DPI) |

## Watermark Types

### WatermarkText

```go
type WatermarkText struct {
    Text   string  // Watermark text
    Font   string  // Font family, e.g. "Calibri"
    Size   float64 // Font size in points (0 = auto)
    Color  string  // Hex color, e.g. "C0C0C0"
    Bold   bool
    Italic bool
}
```

### WatermarkPicture

```go
type WatermarkPicture struct {
    Source   string // File path
    Data     []byte // Raw image data
    MimeType string
    Width    int    // Width in points (0 = auto)
    Height   int    // Height in points (0 = auto)
    Washout  bool   // Apply washout effect
}
```

## Form Field Types

```go
type FormField struct {
    Type           FormFieldType
    Name           string
    DefaultValue   string
    Value          string
    Enabled        bool
    CalcOnExit     bool
    PossibleValues []string // For dropdown
    MaxLength      int      // For text input (0 = unlimited)
    Font           FontStyle
    Para           ParagraphStyle
}

const (
    FormFieldTypeText     FormFieldType = iota // Text input
    FormFieldTypeCheckBox                      // Checkbox
    FormFieldTypeDropDown                      // Dropdown list
)
```

## Field Code Constants

```go
const (
    FieldCurrentPage   = "PAGE"
    FieldNumberOfPages = "NUMPAGES"
    FieldDate          = "DATE"
    FieldCreateDate    = "CREATEDATE"
    FieldEditTime      = "EDITTIME"
    FieldPrintDate     = "PRINTDATE"
    FieldSaveDate      = "SAVEDATE"
    FieldTime          = "TIME"
    FieldFileName      = "FILENAME"
    FieldAuthor        = "AUTHOR"
    FieldTitle         = "TITLE"
    FieldSubject       = "SUBJECT"
)
```

## Element Types

All section elements implement the `Element` interface. Type-switch to identify:

```go
for _, elem := range section.Elements {
    switch e := elem.(type) {
    case *goword.Paragraph:
        for _, run := range e.Runs {
            fmt.Println(run.Text)
        }
    case *goword.Table:
        for _, row := range e.Rows {
            for _, cell := range row.Cells {
                // ...
            }
        }
    case *goword.Image:
        fmt.Println(e.Source, e.Style.Width, e.Style.Height)
    case *goword.Hyperlink:
        fmt.Println(e.URL, e.Text)
    // ... TextRun, TextBreak, PageBreak, ListItem, Footnote, Endnote,
    //     Bookmark, BookmarkEnd, Comment, PreserveText, TOC, CheckBox, Line,
    //     WatermarkText, WatermarkPicture, FormField, Tab
    }
}
```
