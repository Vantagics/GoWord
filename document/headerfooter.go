package document

import (
	"github.com/VantageDataChat/GoWord/style"
)

// AddText adds text to a header.
func (h *Header) AddText(text string, fontStyle *style.FontStyle, paraStyle *style.ParagraphStyle) *Paragraph {
	h.RawPart = nil
	p := &Paragraph{doc: h.doc}
	if paraStyle != nil {
		p.Style = *paraStyle
	}
	run := &Run{Text: text}
	if fontStyle != nil {
		run.Style = *fontStyle
	}
	p.Runs = append(p.Runs, run)
	h.Elements = append(h.Elements, p)
	return p
}

// AddPreserveText adds preserve text (page numbers etc.) to a header.
func (h *Header) AddPreserveText(text string, fontStyle *style.FontStyle, paraStyle *style.ParagraphStyle) *PreserveText {
	h.RawPart = nil
	pt := &PreserveText{Text: text}
	if fontStyle != nil {
		pt.Font = *fontStyle
	}
	if paraStyle != nil {
		pt.Para = *paraStyle
	}
	h.Elements = append(h.Elements, pt)
	return pt
}

// AddImage adds an image to a header.
func (h *Header) AddImage(src string, imgStyle *style.ImageStyle) *Image {
	h.RawPart = nil
	img := &Image{
		Source: src,
		RelID:  h.doc.allocRelID(),
		ID:     h.doc.allocImageID(),
	}
	if imgStyle != nil {
		img.Style = *imgStyle
	}
	h.doc.images = append(h.doc.images, img)
	h.Elements = append(h.Elements, img)
	return img
}

// AddTable adds a table to a header.
func (h *Header) AddTable(tableStyle *style.TableStyle) *Table {
	h.RawPart = nil
	t := &Table{doc: h.doc}
	if tableStyle != nil {
		t.Style = *tableStyle
	}
	h.Elements = append(h.Elements, t)
	return t
}

// AddText adds text to a footer.
func (f *Footer) AddText(text string, fontStyle *style.FontStyle, paraStyle *style.ParagraphStyle) *Paragraph {
	f.RawPart = nil
	p := &Paragraph{doc: f.doc}
	if paraStyle != nil {
		p.Style = *paraStyle
	}
	run := &Run{Text: text}
	if fontStyle != nil {
		run.Style = *fontStyle
	}
	p.Runs = append(p.Runs, run)
	f.Elements = append(f.Elements, p)
	return p
}

// AddPreserveText adds preserve text (page numbers etc.) to a footer.
func (f *Footer) AddPreserveText(text string, fontStyle *style.FontStyle, paraStyle *style.ParagraphStyle) *PreserveText {
	f.RawPart = nil
	pt := &PreserveText{Text: text}
	if fontStyle != nil {
		pt.Font = *fontStyle
	}
	if paraStyle != nil {
		pt.Para = *paraStyle
	}
	f.Elements = append(f.Elements, pt)
	return pt
}

// AddImage adds an image to a footer.
func (f *Footer) AddImage(src string, imgStyle *style.ImageStyle) *Image {
	f.RawPart = nil
	img := &Image{
		Source: src,
		RelID:  f.doc.allocRelID(),
		ID:     f.doc.allocImageID(),
	}
	if imgStyle != nil {
		img.Style = *imgStyle
	}
	f.doc.images = append(f.doc.images, img)
	f.Elements = append(f.Elements, img)
	return img
}

// AddTable adds a table to a footer.
func (f *Footer) AddTable(tableStyle *style.TableStyle) *Table {
	f.RawPart = nil
	t := &Table{doc: f.doc}
	if tableStyle != nil {
		t.Style = *tableStyle
	}
	f.Elements = append(f.Elements, t)
	return t
}

// AddText adds text to a footnote.
func (fn *Footnote) AddText(text string, fontStyle *style.FontStyle) {
	if fn.doc != nil {
		fn.doc.notesDirty = true
	}
	run := &Run{Text: text}
	if fontStyle != nil {
		run.Style = *fontStyle
	}
	p := &Paragraph{doc: fn.doc, Runs: []*Run{run}}
	fn.Elements = append(fn.Elements, p)
}

// AddLink adds a hyperlink to a footnote.
func (fn *Footnote) AddLink(url, text string) *Hyperlink {
	if fn.doc != nil {
		fn.doc.notesDirty = true
	}
	h := &Hyperlink{
		URL:   url,
		Text:  text,
		RelID: fn.doc.allocRelID(),
	}
	fn.Elements = append(fn.Elements, h)
	return h
}

// AddTextBreak adds a line break to a footnote.
func (fn *Footnote) AddTextBreak() {
	if fn.doc != nil {
		fn.doc.notesDirty = true
	}
	fn.Elements = append(fn.Elements, &TextBreak{Count: 1})
}

// AddText adds text to an endnote.
func (en *Endnote) AddText(text string, fontStyle *style.FontStyle) {
	if en.doc != nil {
		en.doc.notesDirty = true
	}
	run := &Run{Text: text}
	if fontStyle != nil {
		run.Style = *fontStyle
	}
	p := &Paragraph{doc: en.doc, Runs: []*Run{run}}
	en.Elements = append(en.Elements, p)
}

// AddLink adds a hyperlink to an endnote.
func (en *Endnote) AddLink(url, text string) *Hyperlink {
	if en.doc != nil {
		en.doc.notesDirty = true
	}
	h := &Hyperlink{
		URL:   url,
		Text:  text,
		RelID: en.doc.allocRelID(),
	}
	en.Elements = append(en.Elements, h)
	return h
}

// AddTextBreak adds a line break to an endnote.
func (en *Endnote) AddTextBreak() {
	if en.doc != nil {
		en.doc.notesDirty = true
	}
	en.Elements = append(en.Elements, &TextBreak{Count: 1})
}
