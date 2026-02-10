package document

import (
	"github.com/VantageDataChat/GoWord/style"
)

// AddText adds a simple text paragraph to the section.
func (s *Section) AddText(text string, fontStyle *style.FontStyle, paraStyle *style.ParagraphStyle) *Paragraph {
	p := &Paragraph{doc: s.doc}
	if paraStyle != nil {
		p.Style = *paraStyle
	}
	run := &Run{Text: text}
	if fontStyle != nil {
		run.Style = *fontStyle
	}
	p.Runs = append(p.Runs, run)
	s.Elements = append(s.Elements, p)
	return p
}

// AddTextWithStyle adds text using named font and paragraph styles.
func (s *Section) AddTextWithStyle(text string, fontStyleName string, paraStyleName string) *Paragraph {
	p := &Paragraph{doc: s.doc, StyleName: paraStyleName}
	run := &Run{Text: text, StyleName: fontStyleName}
	p.Runs = append(p.Runs, run)
	s.Elements = append(s.Elements, p)
	return p
}

// AddTextRun adds a complex paragraph (text run) to the section.
func (s *Section) AddTextRun(paraStyle *style.ParagraphStyle) *TextRun {
	tr := &TextRun{doc: s.doc}
	if paraStyle != nil {
		tr.Style = *paraStyle
	}
	s.Elements = append(s.Elements, tr)
	return tr
}

// AddTextBreak adds empty line(s) to the section.
func (s *Section) AddTextBreak(count int) {
	if count < 1 {
		count = 1
	}
	s.Elements = append(s.Elements, &TextBreak{Count: count})
}

// AddPageBreak adds a page break to the section.
func (s *Section) AddPageBreak() {
	s.Elements = append(s.Elements, &PageBreak{})
}

// AddLink adds a hyperlink to the section.
func (s *Section) AddLink(url, text string, fontStyle *style.FontStyle, paraStyle *style.ParagraphStyle) *Hyperlink {
	h := &Hyperlink{
		URL:   url,
		Text:  text,
		RelID: s.doc.allocRelID(),
	}
	if fontStyle != nil {
		h.Font = *fontStyle
	}
	if paraStyle != nil {
		h.Para = *paraStyle
	}
	s.Elements = append(s.Elements, h)
	return h
}

// AddTitle adds a title/heading to the section.
func (s *Section) AddTitle(text string, depth int) *Paragraph {
	p := &Paragraph{doc: s.doc}
	p.StyleName = headingStyleName(depth)
	run := &Run{Text: text}
	p.Runs = append(p.Runs, run)
	s.Elements = append(s.Elements, p)
	return p
}

// AddTable adds a table to the section.
func (s *Section) AddTable(tableStyle *style.TableStyle) *Table {
	t := &Table{doc: s.doc}
	if tableStyle != nil {
		t.Style = *tableStyle
	}
	s.Elements = append(s.Elements, t)
	return t
}

// AddTableWithStyle adds a table using a named table style.
func (s *Section) AddTableWithStyle(styleName string) *Table {
	t := &Table{doc: s.doc, StyleName: styleName}
	s.Elements = append(s.Elements, t)
	return t
}

// AddImage adds an image from a file path.
func (s *Section) AddImage(src string, imgStyle *style.ImageStyle) *Image {
	img := &Image{
		Source: src,
		RelID:  s.doc.allocRelID(),
		ID:     s.doc.allocImageID(),
	}
	if imgStyle != nil {
		img.Style = *imgStyle
	}
	s.doc.images = append(s.doc.images, img)
	s.Elements = append(s.Elements, img)
	return img
}

// AddImageFromBytes adds an image from raw bytes.
func (s *Section) AddImageFromBytes(data []byte, mimeType string, imgStyle *style.ImageStyle) *Image {
	img := &Image{
		Data:     data,
		MimeType: mimeType,
		RelID:    s.doc.allocRelID(),
		ID:       s.doc.allocImageID(),
	}
	if imgStyle != nil {
		img.Style = *imgStyle
	}
	s.doc.images = append(s.doc.images, img)
	s.Elements = append(s.Elements, img)
	return img
}

// AddListItem adds a list item to the section.
func (s *Section) AddListItem(text string, depth int, fontStyle *style.FontStyle, listStyleName string, paraStyle *style.ParagraphStyle) *ListItem {
	li := &ListItem{
		Text:      text,
		Depth:     depth,
		ListStyle: listStyleName,
	}
	if fontStyle != nil {
		li.Font = *fontStyle
	}
	if paraStyle != nil {
		li.Para = *paraStyle
	}
	s.Elements = append(s.Elements, li)
	return li
}

// AddFootnote adds a footnote and returns it for adding content.
func (s *Section) AddFootnote() *Footnote {
	fn := &Footnote{
		doc: s.doc,
		ID:  s.doc.allocFootnoteID(),
	}
	s.doc.footnotes = append(s.doc.footnotes, fn)
	s.Elements = append(s.Elements, fn)
	return fn
}

// AddEndnote adds an endnote and returns it for adding content.
func (s *Section) AddEndnote() *Endnote {
	en := &Endnote{
		doc: s.doc,
		ID:  s.doc.allocEndnoteID(),
	}
	s.doc.endnotes = append(s.doc.endnotes, en)
	s.Elements = append(s.Elements, en)
	return en
}

// AddTOC adds a table of contents.
func (s *Section) AddTOC(fontStyle *style.FontStyle, minDepth, maxDepth int) *TOC {
	toc := &TOC{
		MinDepth: minDepth,
		MaxDepth: maxDepth,
	}
	if fontStyle != nil {
		toc.Font = *fontStyle
	}
	s.Elements = append(s.Elements, toc)
	return toc
}

// AddCheckBox adds a checkbox form field.
func (s *Section) AddCheckBox(name, text string, fontStyle *style.FontStyle, paraStyle *style.ParagraphStyle) *CheckBox {
	cb := &CheckBox{Name: name, Text: text}
	if fontStyle != nil {
		cb.Font = *fontStyle
	}
	if paraStyle != nil {
		cb.Para = *paraStyle
	}
	s.Elements = append(s.Elements, cb)
	return cb
}

// AddLine adds a line shape.
func (s *Section) AddLine(lineStyle Line) *Line {
	s.Elements = append(s.Elements, &lineStyle)
	return &lineStyle
}

// AddBookmark adds a bookmark anchor.
func (s *Section) AddBookmark(name string) *Bookmark {
	bm := &Bookmark{
		ID:   s.doc.allocBookmarkID(),
		Name: name,
	}
	s.Elements = append(s.Elements, bm)
	return bm
}

// AddComment adds a comment to the section.
func (s *Section) AddComment(author, text string) *Comment {
	c := &Comment{
		doc:    s.doc,
		ID:     s.doc.allocCommentID(),
		Author: author,
		Text:   text,
	}
	s.doc.comments = append(s.doc.comments, c)
	s.Elements = append(s.Elements, c)
	return c
}

// AddFormField adds a form field (text input, dropdown, or checkbox) to the section.
func (s *Section) AddFormField(fieldType FormFieldType, name string) *FormField {
	ff := &FormField{
		Type:    fieldType,
		Name:    name,
		Enabled: true,
	}
	s.Elements = append(s.Elements, ff)
	return ff
}

// AddTextInput adds a text input form field to the section.
func (s *Section) AddTextInput(name string) *FormField {
	return s.AddFormField(FormFieldTypeText, name)
}

// AddDropdownList adds a dropdown list form field to the section.
func (s *Section) AddDropdownList(name string, values []string) *FormField {
	ff := s.AddFormField(FormFieldTypeDropDown, name)
	ff.PossibleValues = values
	return ff
}

// AddHeader adds a header to the section.
func (s *Section) AddHeader(headerType string) *Header {
	if headerType == "" {
		headerType = "default"
	}
	h := &Header{doc: s.doc, Type: headerType}
	s.Headers = append(s.Headers, h)
	return h
}

// AddFooter adds a footer to the section.
func (s *Section) AddFooter(footerType string) *Footer {
	if footerType == "" {
		footerType = "default"
	}
	f := &Footer{doc: s.doc, Type: footerType}
	s.Footers = append(s.Footers, f)
	return f
}

func headingStyleName(depth int) string {
	if depth == 0 {
		return "Title"
	}
	return "Heading" + itoa(depth)
}
