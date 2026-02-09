package document

import (
	"strconv"

	"github.com/VantageDataChat/GoWord/style"
)

// AddSection adds a new section to the document with default style.
func (d *Document) AddSection() *Section {
	return d.AddSectionWithStyle(style.DefaultSectionStyle())
}

// AddSectionWithStyle adds a new section with the given style.
func (d *Document) AddSectionWithStyle(s style.SectionStyle) *Section {
	sec := &Section{
		doc:   d,
		Style: s,
	}
	d.Sections = append(d.Sections, sec)
	return sec
}

// SetDefaultFontName sets the default font name.
func (d *Document) SetDefaultFontName(name string) {
	d.DefaultFont.Name = name
}

// SetDefaultFontSize sets the default font size in points.
func (d *Document) SetDefaultFontSize(size float64) {
	d.DefaultFont.Size = size
}

// AddFontStyle registers a named font style.
func (d *Document) AddFontStyle(name string, fs style.FontStyle) {
	d.fontStyles[name] = fs
}

// AddParagraphStyle registers a named paragraph style.
func (d *Document) AddParagraphStyle(name string, ps style.ParagraphStyle) {
	d.paragraphStyles[name] = ps
}

// AddTableStyle registers a named table style.
func (d *Document) AddTableStyle(name string, ts style.TableStyle) {
	d.tableStyles[name] = ts
}

// AddNumberingStyle registers a named numbering style for lists.
func (d *Document) AddNumberingStyle(name string, ns NumberingStyle) {
	d.numberingStyles[name] = ns
}

// GetFontStyle returns a named font style.
func (d *Document) GetFontStyle(name string) (style.FontStyle, bool) {
	fs, ok := d.fontStyles[name]
	return fs, ok
}

// GetParagraphStyle returns a named paragraph style.
func (d *Document) GetParagraphStyle(name string) (style.ParagraphStyle, bool) {
	ps, ok := d.paragraphStyles[name]
	return ps, ok
}

// GetTableStyle returns a named table style.
func (d *Document) GetTableStyle(name string) (style.TableStyle, bool) {
	ts, ok := d.tableStyles[name]
	return ts, ok
}

// GetNumberingStyle returns a named numbering style.
func (d *Document) GetNumberingStyle(name string) (NumberingStyle, bool) {
	ns, ok := d.numberingStyles[name]
	return ns, ok
}

// FontStyles returns all registered font styles.
func (d *Document) FontStyles() map[string]style.FontStyle {
	return d.fontStyles
}

// ParagraphStyles returns all registered paragraph styles.
func (d *Document) ParagraphStyles() map[string]style.ParagraphStyle {
	return d.paragraphStyles
}

// TableStyles returns all registered table styles.
func (d *Document) TableStyles() map[string]style.TableStyle {
	return d.tableStyles
}

// NumberingStyles returns all registered numbering styles.
func (d *Document) NumberingStyles() map[string]NumberingStyle {
	return d.numberingStyles
}

// Footnotes returns all footnotes in the document.
func (d *Document) Footnotes() []*Footnote {
	return d.footnotes
}

// Endnotes returns all endnotes in the document.
func (d *Document) Endnotes() []*Endnote {
	return d.endnotes
}

// Comments returns all comments in the document.
func (d *Document) Comments() []*Comment {
	return d.comments
}

// Images returns all images in the document.
func (d *Document) Images() []*Image {
	return d.images
}

// allocRelID returns the next available relationship ID.
func (d *Document) allocRelID() string {
	id := d.nextRelID
	d.nextRelID++
	return "rId" + itoa(id)
}

func (d *Document) allocImageID() int {
	id := d.nextImageID
	d.nextImageID++
	return id
}

func (d *Document) allocFootnoteID() int {
	id := d.nextFootnoteID
	d.nextFootnoteID++
	return id
}

func (d *Document) allocEndnoteID() int {
	id := d.nextEndnoteID
	d.nextEndnoteID++
	return id
}

func (d *Document) allocCommentID() int {
	id := d.nextCommentID
	d.nextCommentID++
	return id
}

func (d *Document) allocBookmarkID() int {
	id := d.nextBookmarkID
	d.nextBookmarkID++
	return id
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
