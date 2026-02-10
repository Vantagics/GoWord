package document

import (
	"strconv"
	"strings"

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

// AllocRelIDPublic returns the next available relationship ID (for use by ooxml package).
func (d *Document) AllocRelIDPublic() string {
	return d.allocRelID()
}

// AllocImageIDPublic returns the next available image ID (for use by ooxml package).
func (d *Document) AllocImageIDPublic() int {
	return d.allocImageID()
}

// AppendImage adds an image to the document's image list (for use by ooxml package).
func (d *Document) AppendImage(img *Image) {
	d.images = append(d.images, img)
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

// AddWatermarkText adds a text watermark to the document.
func (d *Document) AddWatermarkText(text string) *WatermarkText {
	wm := &WatermarkText{
		Text:  text,
		Font:  "Calibri",
		Color: "C0C0C0",
	}
	d.watermarkText = wm
	d.watermarkPicture = nil // only one watermark at a time
	return wm
}

// AddWatermarkPicture adds a picture watermark to the document.
func (d *Document) AddWatermarkPicture(src string) *WatermarkPicture {
	wm := &WatermarkPicture{
		Source:  src,
		Washout: true,
	}
	d.watermarkPicture = wm
	d.watermarkText = nil
	return wm
}

// AddWatermarkPictureFromBytes adds a picture watermark from raw bytes.
func (d *Document) AddWatermarkPictureFromBytes(data []byte, mimeType string) *WatermarkPicture {
	wm := &WatermarkPicture{
		Data:     data,
		MimeType: mimeType,
		Washout:  true,
	}
	d.watermarkPicture = wm
	d.watermarkText = nil
	return wm
}

// WatermarkTextValue returns the current text watermark, or nil.
func (d *Document) WatermarkTextValue() *WatermarkText {
	return d.watermarkText
}

// WatermarkPictureValue returns the current picture watermark, or nil.
func (d *Document) WatermarkPictureValue() *WatermarkPicture {
	return d.watermarkPicture
}

// SetUpdateFieldsOnOpen controls whether fields (TOC, page numbers, etc.)
// are recalculated when the document is opened.
func (d *Document) SetUpdateFieldsOnOpen(b bool) {
	d.updateFieldsOnOpen = b
}

// UpdateFieldsOnOpen returns whether fields are recalculated on open.
func (d *Document) UpdateFieldsOnOpen() bool {
	return d.updateFieldsOnOpen
}

// Paragraphs returns all paragraphs across all sections (flattened).
func (d *Document) Paragraphs() []*Paragraph {
	var result []*Paragraph
	for _, sec := range d.Sections {
		for _, elem := range sec.Elements {
			if p, ok := elem.(*Paragraph); ok {
				result = append(result, p)
			}
		}
	}
	return result
}

// Tables returns all tables across all sections (flattened).
func (d *Document) Tables() []*Table {
	var result []*Table
	for _, sec := range d.Sections {
		for _, elem := range sec.Elements {
			if t, ok := elem.(*Table); ok {
				result = append(result, t)
			}
		}
	}
	return result
}

// ExtractText returns all text content from the document as a single string.
func (d *Document) ExtractText() string {
	var b strings.Builder
	for _, sec := range d.Sections {
		for _, elem := range sec.Elements {
			extractElementText(&b, elem)
		}
	}
	return b.String()
}

func extractElementText(b *strings.Builder, elem Element) {
	switch e := elem.(type) {
	case *Paragraph:
		for _, run := range e.Runs {
			b.WriteString(run.Text)
		}
		b.WriteString("\n")
	case *TextRun:
		for _, child := range e.Elements {
			extractElementText(b, child)
		}
	case *TextBreak:
		for i := 0; i < e.Count; i++ {
			b.WriteString("\n")
		}
	case *Hyperlink:
		b.WriteString(e.Text)
	case *ListItem:
		b.WriteString(e.Text)
		b.WriteString("\n")
	case *Table:
		for _, row := range e.Rows {
			for ci, cell := range row.Cells {
				for _, ce := range cell.Elements {
					extractElementText(b, ce)
				}
				if ci < len(row.Cells)-1 {
					b.WriteString("\t")
				}
			}
		}
	case *CheckBox:
		b.WriteString(e.Text)
	}
}

// RemoveParagraph removes a paragraph from the document.
func (d *Document) RemoveParagraph(p *Paragraph) {
	for _, sec := range d.Sections {
		for i, elem := range sec.Elements {
			if elem == p {
				sec.Elements = append(sec.Elements[:i], sec.Elements[i+1:]...)
				return
			}
		}
	}
}

// InsertParagraphBefore inserts a new empty paragraph before the given paragraph.
func (d *Document) InsertParagraphBefore(relativeTo *Paragraph) *Paragraph {
	for _, sec := range d.Sections {
		for i, elem := range sec.Elements {
			if elem == relativeTo {
				np := &Paragraph{doc: d}
				sec.Elements = append(sec.Elements[:i+1], sec.Elements[i:]...)
				sec.Elements[i] = np
				return np
			}
		}
	}
	return nil
}

// InsertParagraphAfter inserts a new empty paragraph after the given paragraph.
func (d *Document) InsertParagraphAfter(relativeTo *Paragraph) *Paragraph {
	for _, sec := range d.Sections {
		for i, elem := range sec.Elements {
			if elem == relativeTo {
				np := &Paragraph{doc: d}
				sec.Elements = append(sec.Elements[:i+1], append([]Element{np}, sec.Elements[i+1:]...)...)
				return np
			}
		}
	}
	return nil
}
