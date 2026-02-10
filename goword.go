// Package goword provides a pure Go library for reading and writing
// Microsoft Word documents (.docx / Office Open XML format).
//
// It is inspired by PHPWord and provides similar capabilities:
//   - Create documents with sections, paragraphs, tables, images, etc.
//   - Set document properties (title, creator, subject, etc.)
//   - Define font and paragraph styles (bold, italic, alignment, etc.)
//   - Insert headers, footers, page breaks, text breaks
//   - Insert hyperlinks, lists, footnotes, endnotes
//   - Insert images (local files or byte data)
//   - Insert tables with cell spanning (gridSpan, vMerge)
//   - Insert table of contents, bookmarks, comments
//   - Read existing .docx files
//   - Write .docx files
package goword

import (
	"github.com/VantageDataChat/GoWord/common"
	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/ooxml"
	"github.com/VantageDataChat/GoWord/style"
)

// Version is the current version of the GoWord library.
const Version = "1.0.0"

func init() {
	// Register OOXML reader/writer with the document package.
	document.RegisterIO(ooxml.Read, ooxml.ReadFromBytes, ooxml.Save, ooxml.WriteToBytes)
}

// New creates a new empty Word document.
func New() *document.Document {
	return document.New()
}

// Open reads an existing .docx file and returns a Document.
func Open(path string) (*document.Document, error) {
	return document.Open(path)
}

// OpenFromBytes reads a .docx from byte data and returns a Document.
func OpenFromBytes(data []byte) (*document.Document, error) {
	return document.OpenFromBytes(data)
}

// Re-export commonly used types for convenience.
type (
	Document       = document.Document
	Section        = document.Section
	Paragraph      = document.Paragraph
	TextRun        = document.TextRun
	Table          = document.Table
	Row            = document.Row
	Cell           = document.Cell
	Image          = document.Image
	Hyperlink      = document.Hyperlink
	Header         = document.Header
	Footer         = document.Footer
	Footnote       = document.Footnote
	Endnote        = document.Endnote
	ListItem       = document.ListItem
	Bookmark       = document.Bookmark
	Comment        = document.Comment
	FontStyle      = style.FontStyle
	ParagraphStyle = style.ParagraphStyle
	TableStyle     = style.TableStyle
	DocProperties  = common.DocProperties
	NumberingStyle = document.NumberingStyle
	NumberingLevel = document.NumberingLevel
	Line           = document.Line
	WatermarkText    = document.WatermarkText
	WatermarkPicture = document.WatermarkPicture
	FormField        = document.FormField
	FormFieldType    = document.FormFieldType
	TabStop          = style.TabStop
)

// Form field type constants re-exported for convenience.
const (
	FormFieldTypeText     = document.FormFieldTypeText
	FormFieldTypeCheckBox = document.FormFieldTypeCheckBox
	FormFieldTypeDropDown = document.FormFieldTypeDropDown
)

// Field code constants re-exported for convenience.
const (
	FieldCurrentPage   = document.FieldCurrentPage
	FieldNumberOfPages = document.FieldNumberOfPages
	FieldDate          = document.FieldDate
	FieldCreateDate    = document.FieldCreateDate
	FieldEditTime      = document.FieldEditTime
	FieldPrintDate     = document.FieldPrintDate
	FieldSaveDate      = document.FieldSaveDate
	FieldTime          = document.FieldTime
	FieldFileName      = document.FieldFileName
	FieldAuthor        = document.FieldAuthor
	FieldTitle         = document.FieldTitle
	FieldSubject       = document.FieldSubject
)
