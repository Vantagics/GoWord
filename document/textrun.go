package document

import (
	"github.com/VantageDataChat/GoWord/style"
)

// AddText adds a text run to the TextRun.
func (tr *TextRun) AddText(text string, fontStyle *style.FontStyle) *Run {
	run := &Run{Text: text}
	if fontStyle != nil {
		run.Style = *fontStyle
	}
	p := &Paragraph{doc: tr.doc, Runs: []*Run{run}}
	tr.Elements = append(tr.Elements, p)
	return run
}

// AddTextWithStyle adds text using a named font style.
func (tr *TextRun) AddTextWithStyle(text string, styleName string) *Run {
	run := &Run{Text: text, StyleName: styleName}
	p := &Paragraph{doc: tr.doc, Runs: []*Run{run}}
	tr.Elements = append(tr.Elements, p)
	return run
}

// AddLink adds a hyperlink to the TextRun.
func (tr *TextRun) AddLink(url, text string, fontStyle *style.FontStyle) *Hyperlink {
	h := &Hyperlink{
		URL:   url,
		Text:  text,
		RelID: tr.doc.allocRelID(),
	}
	if fontStyle != nil {
		h.Font = *fontStyle
	}
	tr.Elements = append(tr.Elements, h)
	return h
}

// AddTextBreak adds a line break within the TextRun.
func (tr *TextRun) AddTextBreak(count int) {
	if count < 1 {
		count = 1
	}
	tr.Elements = append(tr.Elements, &TextBreak{Count: count})
}

// AddImage adds an inline image to the TextRun.
func (tr *TextRun) AddImage(src string, imgStyle *style.ImageStyle) *Image {
	img := &Image{
		Source: src,
		RelID:  tr.doc.allocRelID(),
		ID:     tr.doc.allocImageID(),
	}
	if imgStyle != nil {
		img.Style = *imgStyle
	}
	tr.doc.images = append(tr.doc.images, img)
	tr.Elements = append(tr.Elements, img)
	return img
}

// AddFootnote adds a footnote reference in the TextRun.
func (tr *TextRun) AddFootnote() *Footnote {
	fn := &Footnote{
		doc: tr.doc,
		ID:  tr.doc.allocFootnoteID(),
	}
	tr.doc.footnotes = append(tr.doc.footnotes, fn)
	tr.Elements = append(tr.Elements, fn)
	return fn
}

// AddEndnote adds an endnote reference in the TextRun.
func (tr *TextRun) AddEndnote() *Endnote {
	en := &Endnote{
		doc: tr.doc,
		ID:  tr.doc.allocEndnoteID(),
	}
	tr.doc.endnotes = append(tr.doc.endnotes, en)
	tr.Elements = append(tr.Elements, en)
	return en
}

// AddBookmark adds a bookmark in the TextRun.
func (tr *TextRun) AddBookmark(name string) *Bookmark {
	bm := &Bookmark{
		ID:   tr.doc.allocBookmarkID(),
		Name: name,
	}
	tr.Elements = append(tr.Elements, bm)
	return bm
}

// AddTab adds a tab character to the TextRun.
func (tr *TextRun) AddTab() {
	tr.Elements = append(tr.Elements, &Tab{})
}
