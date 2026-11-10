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
	if !d.loading {
		d.stylesDirty = true
	}
}

// AddParagraphStyle registers a named paragraph style.
func (d *Document) AddParagraphStyle(name string, ps style.ParagraphStyle) {
	d.paragraphStyles[name] = ps
	if !d.loading {
		d.stylesDirty = true
	}
}

// AddTableStyle registers a named table style.
func (d *Document) AddTableStyle(name string, ts style.TableStyle) {
	d.tableStyles[name] = ts
	if !d.loading {
		d.stylesDirty = true
	}
}

// AddNumberingStyle registers a named numbering style for lists.
func (d *Document) AddNumberingStyle(name string, ns NumberingStyle) {
	d.numberingStyles[name] = ns
	if !d.loading {
		d.numberingDirty = true
	}
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
	if !d.loading {
		d.settingsDirty = true
	}
}

// UpdateFieldsOnOpen returns whether fields are recalculated on open.
func (d *Document) UpdateFieldsOnOpen() bool {
	return d.updateFieldsOnOpen
}

// Paragraphs returns all paragraphs across all sections (flattened).
// Paragraphs inside content controls are included. Paragraphs inside tables
// are not, matching the historical behavior of this method.
func (d *Document) Paragraphs() []*Paragraph {
	var result []*Paragraph
	var walk func([]Element)
	walk = func(elems []Element) {
		for _, elem := range elems {
			switch e := elem.(type) {
			case *Paragraph:
				result = append(result, e)
			case *Sdt:
				walk(e.Elements)
			}
		}
	}
	for _, sec := range d.Sections {
		walk(sec.Elements)
	}
	return result
}

// Tables returns all tables across all sections, including nested tables and
// tables inside content controls.
func (d *Document) Tables() []*Table {
	var result []*Table
	var walk func([]Element)
	walk = func(elems []Element) {
		for _, elem := range elems {
			switch e := elem.(type) {
			case *Table:
				result = append(result, e)
				for _, row := range e.Rows {
					for _, cell := range row.Cells {
						walk(cell.Elements)
					}
				}
			case *Sdt:
				walk(e.Elements)
			}
		}
	}
	for _, sec := range d.Sections {
		walk(sec.Elements)
	}
	return result
}

// ExtractOptions controls ExtractTextWith.
//
// The zero value omits headers, footers, note bodies, and comment bodies.
// ExtractDefaults matches ExtractText: headers, footers, note bodies, and
// comment bodies are all included, math stays in linear form, and ruby
// contributes its base text. Note bodies are appended after the body and
// the header/footer text. Comment bodies are appended last.
type ExtractOptions struct {
	IncludeComments   bool
	IncludeNoteBodies bool
	IncludeHeaders    bool
	// MathFormat is "" or "linear" for the linearized equation, or "latex"
	// for an OMML-to-LaTeX conversion.
	MathFormat string
	// Ruby is "" or "base" (the default), "reading", or "both".
	Ruby string
	// NoteAnchor is "" or "glue" (default: the note number is appended to the
	// preceding word, which is what LibreOffice writes) or "separate" (a space
	// before the number, which is what pandoc's plain text does).
	NoteAnchor string
}

// ExtractDefaults is the option set used by ExtractText.
func ExtractDefaults() ExtractOptions {
	return ExtractOptions{
		IncludeComments:   true,
		IncludeNoteBodies: true,
		IncludeHeaders:    true,
	}
}

// ExtractText returns all text content from the document as a single string.
// It includes header, footer, footnote, endnote, and comment text.
// Use ExtractTextWith to omit comments or note bodies, or to convert equations.
func (d *Document) ExtractText() string {
	return d.ExtractTextWith(ExtractDefaults())
}

// ExtractTextWith extracts text using opt.
func (d *Document) ExtractTextWith(opt ExtractOptions) string {
	var b strings.Builder
	ctr := map[string][]int{}
	for _, sec := range d.Sections {
		for _, elem := range sec.Elements {
			extractElementText(&b, elem, d, ctr, opt)
		}
		if opt.IncludeHeaders {
			for _, h := range sec.Headers {
				for _, elem := range h.Elements {
					extractElementText(&b, elem, d, ctr, opt)
				}
			}
			for _, f := range sec.Footers {
				for _, elem := range f.Elements {
					extractElementText(&b, elem, d, ctr, opt)
				}
			}
		}
	}
	if opt.IncludeNoteBodies {
		for _, fn := range d.footnotes {
			for _, elem := range fn.Elements {
				extractElementText(&b, elem, d, ctr, opt)
			}
		}
		for _, en := range d.endnotes {
			for _, elem := range en.Elements {
				extractElementText(&b, elem, d, ctr, opt)
			}
		}
	}
	if opt.IncludeComments {
		for _, c := range d.comments {
			if c.Text != "" {
				b.WriteString(c.Text)
				b.WriteString("\n")
			}
			for _, elem := range c.Elements {
				extractElementText(&b, elem, d, ctr, opt)
			}
		}
	}
	return b.String()
}

// Rubies returns phonetic guides in document order.
func (d *Document) Rubies() []*Ruby {
	if d == nil {
		return nil
	}
	var out []*Ruby
	var walk func([]Element)
	walk = func(elems []Element) {
		for _, el := range elems {
			switch e := el.(type) {
			case *Ruby:
				out = append(out, e)
			case *Paragraph:
				walk(e.Inlines)
			case *Table:
				for _, row := range e.Rows {
					for _, cell := range row.Cells {
						walk(cell.Elements)
					}
				}
			case *Sdt:
				walk(e.Elements)
			case *TextRun:
				walk(e.Elements)
			}
		}
	}
	for _, sec := range d.Sections {
		walk(sec.Elements)
	}
	return out
}

func writeRunText(b *strings.Builder, run *Run) {
	if run.Tab {
		b.WriteByte('\t')
	}
	b.WriteString(run.Text)
	if run.Break && run.BreakType == "textWrapping" {
		b.WriteByte('\n')
	}
}

func extractElementText(b *strings.Builder, elem Element, doc *Document, ctr map[string][]int, opt ExtractOptions) {
	switch e := elem.(type) {
	case *Paragraph:
		numID, numLevel := doc.numberingFor(e)
		if label := doc.listLabel(numID, numLevel, ctr); label != "" {
			b.WriteString(label)
		}
		if paragraphUsesInlines(e, opt) {
			for _, in := range e.Inlines {
				extractElementText(b, in, doc, ctr, opt)
			}
		} else {
			for _, run := range e.Runs {
				writeRunText(b, run)
			}
		}
		b.WriteString("\n")
	case *Run:
		writeRunText(b, e)
	case *Tab:
		b.WriteByte('\t')
	case *Ruby:
		switch opt.Ruby {
		case "reading":
			b.WriteString(e.Reading)
		case "both":
			b.WriteString(e.Base)
			if e.Reading != "" {
				if e.Base != "" {
					b.WriteByte(' ')
				}
				b.WriteString(e.Reading)
			}
		default:
			b.WriteString(e.Base)
		}
	case *RawXML:
		if opt.MathFormat == "latex" && strings.Contains(e.XML, "oMath") {
			b.WriteString(OMMLToLatex(e.XML))
		} else {
			b.WriteString(e.Text)
		}
	case *Field:
		if e.Result != "" {
			b.WriteString(e.Result)
		}
	case *Sdt:
		for _, child := range e.Elements {
			extractElementText(b, child, doc, ctr, opt)
		}
	case *TextRun:
		for _, child := range e.Elements {
			extractElementText(b, child, doc, ctr, opt)
		}
	case *TextBreak:
		for i := 0; i < e.Count; i++ {
			b.WriteString("\n")
		}
	case *Hyperlink:
		b.WriteString(e.Text)
	case *ListItem:
		if label := doc.listLabel(e.ListStyle, e.Depth, ctr); label != "" {
			b.WriteString(label)
		}
		b.WriteString(e.Text)
		b.WriteString("\n")
	case *Table:
		for _, row := range e.Rows {
			for ci, cell := range row.Cells {
				for _, ce := range cell.Elements {
					extractElementText(b, ce, doc, ctr, opt)
				}
				if ci < len(row.Cells)-1 {
					b.WriteString("\t")
				}
			}
		}
	case *CheckBox:
		b.WriteString(e.Text)
	case *Footnote:
		if opt.NoteAnchor == "separate" && e.ID > 0 {
			b.WriteByte(' ')
			b.WriteString(strconv.Itoa(e.ID))
		}
	case *Endnote:
		if opt.NoteAnchor == "separate" && e.ID > 0 {
			b.WriteByte(' ')
			b.WriteString(strconv.Itoa(e.ID))
		}
	}
}

func paragraphUsesInlines(p *Paragraph, opt ExtractOptions) bool {
	latex := opt.MathFormat == "latex"
	ruby := opt.Ruby == "reading" || opt.Ruby == "both"
	separateNotes := opt.NoteAnchor == "separate"
	if p == nil || (!latex && !ruby && !separateNotes) {
		return false
	}
	for _, in := range p.Inlines {
		switch e := in.(type) {
		case *RawXML:
			if latex && strings.Contains(e.XML, "oMath") {
				return true
			}
		case *Ruby:
			if ruby {
				return true
			}
		case *Footnote, *Endnote:
			if separateNotes {
				return true
			}
		}
	}
	return false
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
