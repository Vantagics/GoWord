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
	// RawSectPr is the original w:sectPr element, including its tags, when the
	// section was loaded from a package. The writer reuses it so page borders,
	// document grid, and printer settings survive a round trip.
	RawSectPr string
	// headersDirty is set when headers or footers are added through the API
	// after load, which forces the writer to rebuild sectPr.
	headersDirty bool
}

// Paragraph represents a text paragraph.
type Paragraph struct {
	doc       *Document
	Runs      []*Run
	Style     style.ParagraphStyle
	StyleName string // Named paragraph style reference
	// Inlines is the ordered inline content (runs, hyperlinks, fields,
	// bookmarks, drawings). When non-empty the writer emits Inlines and
	// ignores Runs. Runs is still filled so ExtractText and existing callers
	// see the visible text. API-created paragraphs leave Inlines nil.
	Inlines []Element
	// RawPPr is the original w:pPr element when one was loaded. Direct
	// formatting Word stored on the paragraph mark is preserved on write.
	RawPPr string
}

// Run represents a run of text with uniform formatting.
type Run struct {
	Text      string
	Style     style.FontStyle
	StyleName string // Named font style reference
	Break     bool   // If true, this run contains a break
	// BreakType is the w:br/@w:type value: "textWrapping" (line), "page", or
	// "column". Empty together with Break means a line break.
	BreakType string
	// Tab inserts a tab character before the run text.
	Tab bool
	// RawRPr is the original w:rPr element when the run used properties this
	// library does not model (theme fonts, fit text, east-Asian layout, …).
	RawRPr string
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
	URL        string
	Text       string
	Font       style.FontStyle
	Para       style.ParagraphStyle
	Tooltip    string
	RelID      string // internal relationship ID
	IsInternal bool   // bookmark link
}

// Image represents an embedded image.
type Image struct {
	Source   string // File path or URL
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
	// RawTblPr is the original w:tblPr element when the table was loaded.
	RawTblPr string
}

// Row represents a table row.
type Row struct {
	doc   *Document
	Cells []*Cell
	Style style.RowStyle
	// PrefixXML is raw markup that must precede trPr (tblPrEx).
	PrefixXML string
	// RawTrPr is the original w:trPr element when the row was loaded.
	RawTrPr string
}

// Cell represents a table cell.
type Cell struct {
	doc      *Document
	Elements []Element
	Style    style.CellStyle
	// RawTcPr is the original w:tcPr element when the cell was loaded.
	RawTcPr string
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
	// PartName is the header part file name ("header1.xml") when loaded
	// or assigned by the writer.
	PartName string
	// RawPart is the original header XML. Cleared when the header is edited
	// through AddText/AddImage/AddTable so the writer rebuilds the part.
	RawPart []byte
}

// Footer represents a section footer.
type Footer struct {
	doc      *Document
	Type     string // "default", "first", "even"
	Elements []Element
	PartName string
	RawPart  []byte
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
	// Span, when true, writes only bookmarkStart. Pair it with BookmarkEnd.
	// The zero value keeps the historical point-bookmark behavior (start and
	// end written together).
	Span bool
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
	Font      style.FontStyle
	MinDepth  int
	MaxDepth  int
	TabLeader string // "dot", "hyphen", "underscore", "none"
	TabPos    int    // Tab position in twips
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
	Source   string // File path
	Data     []byte // Raw image data
	MimeType string
	Width    int // Width in points (0 = auto)
	Height   int // Height in points (0 = auto)
	Washout  bool
}

// FormField represents a form field (text input, dropdown, or checkbox).
type FormField struct {
	Type           FormFieldType
	Name           string
	DefaultValue   string
	Value          string
	Enabled        bool
	CalcOnExit     bool
	PossibleValues []string // For dropdown
	MaxLength      int      // For text input (0 = unlimited)
	Font           style.FontStyle
	Para           style.ParagraphStyle
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

// RawXML is markup this library preserves verbatim so a round trip does not
// drop drawings, math, tracked changes, or other structures it does not model.
type RawXML struct {
	XML string
	// Text is the visible text inside XML, used by ExtractText.
	Text string
}

// Field is a Word field (PAGE, TOC, MERGEFIELD, …).
type Field struct {
	Instr  string
	Result string
	Font   style.FontStyle
}

// Sdt is a structured document tag (content control).
type Sdt struct {
	// PrXML is the original w:sdtPr element.
	PrXML string
	// EndPrXML is the original w:sdtEndPr element.
	EndPrXML string
	Elements []Element
}

// CommentRangeStart marks the start of a comment range.
type CommentRangeStart struct {
	ID int
}

// CommentRangeEnd marks the end of a comment range.
type CommentRangeEnd struct {
	ID int
}

// Implement elementType for all elements.
func (p *Paragraph) elementType() string         { return "paragraph" }
func (t *TextRun) elementType() string           { return "textrun" }
func (t *TextBreak) elementType() string         { return "textbreak" }
func (p *PageBreak) elementType() string         { return "pagebreak" }
func (h *Hyperlink) elementType() string         { return "hyperlink" }
func (i *Image) elementType() string             { return "image" }
func (t *Table) elementType() string             { return "table" }
func (l *ListItem) elementType() string          { return "listitem" }
func (f *Footnote) elementType() string          { return "footnote" }
func (e *Endnote) elementType() string           { return "endnote" }
func (b *Bookmark) elementType() string          { return "bookmark" }
func (b *BookmarkEnd) elementType() string       { return "bookmarkend" }
func (c *Comment) elementType() string           { return "comment" }
func (p *PreserveText) elementType() string      { return "preservetext" }
func (t *TOC) elementType() string               { return "toc" }
func (c *CheckBox) elementType() string          { return "checkbox" }
func (l *Line) elementType() string              { return "line" }
func (w *WatermarkText) elementType() string     { return "watermarktext" }
func (w *WatermarkPicture) elementType() string  { return "watermarkpicture" }
func (f *FormField) elementType() string         { return "formfield" }
func (t *Tab) elementType() string               { return "tab" }
func (r *RawXML) elementType() string            { return "rawxml" }
func (f *Field) elementType() string             { return "field" }
func (s *Sdt) elementType() string               { return "sdt" }
func (c *CommentRangeStart) elementType() string { return "commentrangestart" }
func (c *CommentRangeEnd) elementType() string   { return "commentrangeend" }
func (r *Run) elementType() string               { return "run" }
