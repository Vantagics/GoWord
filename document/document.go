// Package document provides the core document model for goword.
package document

import (
	"github.com/VantageDataChat/GoWord/common"
	"github.com/VantageDataChat/GoWord/style"
)

// Document represents a Word document.
type Document struct {
	Properties  common.DocProperties
	Sections    []*Section
	DefaultFont style.FontStyle

	// Named styles
	fontStyles      map[string]style.FontStyle
	paragraphStyles map[string]style.ParagraphStyle
	tableStyles     map[string]style.TableStyle
	numberingStyles map[string]NumberingStyle

	// Relationships and media
	images    []*Image
	footnotes []*Footnote
	endnotes  []*Endnote
	comments  []*Comment

	// Watermarks
	watermarkText    *WatermarkText
	watermarkPicture *WatermarkPicture

	// Settings
	updateFieldsOnOpen bool

	// Internal counters
	nextRelID    int
	nextImageID  int
	nextFootnoteID int
	nextEndnoteID  int
	nextCommentID  int
	nextBookmarkID int
}

// NumberingStyle defines a custom numbering/list style.
type NumberingStyle struct {
	Type   string           // "multilevel", "singleLevel"
	Levels []NumberingLevel
}

// NumberingLevel defines formatting for one level of a numbering style.
type NumberingLevel struct {
	Format   string // "decimal", "upperLetter", "lowerLetter", "upperRoman", "lowerRoman", "bullet"
	Text     string // e.g. "%1." or "%1.%2."
	Left     int    // Left indent in twips
	Hanging  int    // Hanging indent in twips
	TabPos   int    // Tab position in twips
	Font     string // Font name for bullet character
}

// New creates a new empty Document with default settings.
func New() *Document {
	return &Document{
		Properties:      common.NewDocProperties(),
		DefaultFont:     style.DefaultFont(),
		fontStyles:      make(map[string]style.FontStyle),
		paragraphStyles: make(map[string]style.ParagraphStyle),
		tableStyles:     make(map[string]style.TableStyle),
		numberingStyles: make(map[string]NumberingStyle),
		nextRelID:       1,
		nextImageID:     1,
		nextFootnoteID:  1,
		nextEndnoteID:   1,
		nextCommentID:   1,
		nextBookmarkID:  1,
	}
}
