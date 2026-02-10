package document

import (
	"time"

	"github.com/VantageDataChat/GoWord/style"
)

// Element is the interface for all document elements.
type Element interface {
	elementType() string
}

// Field code constants for use with PreserveText.
const (
	FieldCurrentPage   = "PAGE"
	FieldNumberOfPages = "NUMPAGES"
	FieldDate          = "DATE"
	FieldCreateDate    = "CREATEDATE"
	FieldEditTime      = "EDITTIME"
	FieldPrintDate     = "PRINTDATE"
	FieldSaveDate      = "SAVEDATE"
	FieldTime          = "TIME"
	FieldTOC           = "TOC"
	FieldFileName      = "FILENAME"
	FieldAuthor        = "AUTHOR"
	FieldTitle         = "TITLE"
	FieldSubject       = "SUBJECT"
)

// Section represents a document section.
type Section struct {
	doc      *Document
	Style    style.SectionStyle
	Elements []Element
	Headers  []*Header // up to 3: default, first, even
	Footers  []*Footer // up to 3: default, first, even
}

// Paragraph represents a text paragraph.
type Paragraph struct {
	doc       *Document
	Runs      []*Run
	Style     style.ParagraphStyle
	StyleName string // Named paragraph style reference
}

// Run represents a run of text with uniform formatting.
type Run struct {
	Text      string
	Style     style.FontStyle
	StyleName string // Named font style reference
	Break     bool   // If true, this is a line break
}

// TextRun is a complex paragraph containing mixed-format runs and inline elements.
type TextRun struct {
	doc       *Document
	Elements  []Element
	Style     style.ParagraphStyle
	StyleName string
}

// TextBreak represents an empty line / text break.
type TextBreak struct {
	Count int
	Font  *style.FontStyle
	Para  *style.ParagraphStyle
}

// PageBreak represents a page break element.
type PageBreak struct{}

// Hyperlink represents a hyperlink element.
type Hyperlink struct {
	URL       string
	Text      string
	Font      style.FontStyle
	Para      style.ParagraphStyle
	Tooltip   string
	RelID     string // internal relationship ID
	IsInternal bool  // bookmark link
}

// Image represents an embedded image.
type Image struct {
	Source    string // File path or URL
	Data     []byte // Raw image data (alternative to Source)
	MimeType string // e.g. "image/png"
	Style    style.ImageStyle
	RelID    string
	ID       int
	Name     string // Internal name like "image1.png"
}

// Table represents a table element.
type Table struct {
	doc       *Document
	Rows      []*Row
	Style     style.TableStyle
	StyleName string
	Grid      []int // Column widths in twips
}

// Row represents a table row.
type Row struct {
	doc   *Document
	Cells []*Cell
	Style style.RowStyle
}

// Cell represents a table cell.
type Cell struct {
	doc      *Document
	Elements []Element
	Style    style.CellStyle
}

// ListItem represents a list item.
type ListItem struct {
	Text      string
	Depth     int
	Font      style.FontStyle
	ListStyle string // Named numbering style or type constant
	Para      style.ParagraphStyle
}

// Header represents a section header.
type Header struct {
	doc      *Document
	Type     string // "default", "first", "even"
	Elements []Element
}

// Footer represents a section footer.
type Footer struct {
	doc      *Document
	Type     string // "default", "first", "even"
	Elements []Element
}

// Footnote represents a footnote.
type Footnote struct {
	doc      *Document
	ID       int
	Elements []Element
	RefRun   *Run // The reference run in the main text
}

// Endnote represents an endnote.
type Endnote struct {
	doc      *Document
	ID       int
	Elements []Element
	RefRun   *Run
}

// Bookmark represents a bookmark (anchor).
type Bookmark struct {
	ID   int
	Name string
}

// BookmarkEnd marks the end of a bookmark.
type BookmarkEnd struct {
	ID int
}

// Comment represents a document comment.
type Comment struct {
	doc      *Document
	ID       int
	Author   string
	Date     time.Time
	Text     string
	Elements []Element
}

// PreserveText is used in headers/footers for page numbers etc.
type PreserveText struct {
	Text string // e.g. "Page {PAGE} of {NUMPAGES}"
	Font style.FontStyle
	Para style.ParagraphStyle
}

// TOC represents a Table of Contents.
type TOC struct {
	Font     style.FontStyle
	MinDepth int
	MaxDepth int
	TabLeader string // "dot", "hyphen", "underscore", "none"
	TabPos   int     // Tab position in twips
}

// CheckBox represents a checkbox form field.
type CheckBox struct {
	Name    string
	Text    string
	Checked bool
	Font    style.FontStyle
	Para    style.ParagraphStyle
}

// Line represents a drawing line shape.
type Line struct {
	Width  int    // Width in points
	Height int    // Height in points
	Weight int    // Line weight in twips
	Color  string // Hex color
	Dash   string // "solid", "dash", "dot", etc.
}

// WatermarkText represents a text watermark.
type WatermarkText struct {
	Text   string
	Font   string  // Font family, e.g. "Calibri"
	Size   float64 // Font size in points (0 = auto)
	Color  string  // Hex color, e.g. "C0C0C0"
	Bold   bool
	Italic bool
}

// WatermarkPicture represents a picture watermark.
type WatermarkPicture struct {
	Source  string // File path
	Data    []byte // Raw image data
	MimeType string
	Width   int // Width in points (0 = auto)
	Height  int // Height in points (0 = auto)
	Washout bool
}

// FormField represents a form field (text input, dropdown, or checkbox).
type FormField struct {
	Type          FormFieldType
	Name          string
	DefaultValue  string
	Value         string
	Enabled       bool
	CalcOnExit    bool
	PossibleValues []string // For dropdown
	MaxLength     int       // For text input (0 = unlimited)
	Font          style.FontStyle
	Para          style.ParagraphStyle
}

// FormFieldType is the type of form field.
type FormFieldType int

const (
	FormFieldTypeText     FormFieldType = iota // Text input
	FormFieldTypeCheckBox                      // Checkbox (use CheckBox element instead for simple cases)
	FormFieldTypeDropDown                      // Dropdown list
)

// Tab represents a tab character in a run.
type Tab struct{}

// Implement elementType for all elements.
func (p *Paragraph) elementType() string    { return "paragraph" }
func (t *TextRun) elementType() string      { return "textrun" }
func (t *TextBreak) elementType() string    { return "textbreak" }
func (p *PageBreak) elementType() string    { return "pagebreak" }
func (h *Hyperlink) elementType() string    { return "hyperlink" }
func (i *Image) elementType() string        { return "image" }
func (t *Table) elementType() string        { return "table" }
func (l *ListItem) elementType() string     { return "listitem" }
func (f *Footnote) elementType() string     { return "footnote" }
func (e *Endnote) elementType() string      { return "endnote" }
func (b *Bookmark) elementType() string     { return "bookmark" }
func (b *BookmarkEnd) elementType() string  { return "bookmarkend" }
func (c *Comment) elementType() string      { return "comment" }
func (p *PreserveText) elementType() string { return "preservetext" }
func (t *TOC) elementType() string          { return "toc" }
func (c *CheckBox) elementType() string     { return "checkbox" }
func (l *Line) elementType() string         { return "line" }
func (w *WatermarkText) elementType() string    { return "watermarktext" }
func (w *WatermarkPicture) elementType() string { return "watermarkpicture" }
func (f *FormField) elementType() string        { return "formfield" }
func (t *Tab) elementType() string              { return "tab" }
