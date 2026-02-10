package ooxml

import (
	"fmt"
	"strings"

	"github.com/VantageDataChat/GoWord/document"
)

func (w *Writer) writeElementXML(b *strings.Builder, elem document.Element) {
	switch e := elem.(type) {
	case *document.Paragraph:
		w.writeParagraph(b, e)
	case *document.TextRun:
		w.writeTextRun(b, e)
	case *document.TextBreak:
		w.writeTextBreak(b, e)
	case *document.PageBreak:
		w.writePageBreak(b)
	case *document.Hyperlink:
		w.writeHyperlink(b, e)
	case *document.Image:
		w.writeImage(b, e)
	case *document.Table:
		w.writeTable(b, e)
	case *document.ListItem:
		w.writeListItem(b, e)
	case *document.Footnote:
		w.writeFootnoteRef(b, e)
	case *document.Endnote:
		w.writeEndnoteRef(b, e)
	case *document.Bookmark:
		w.writeBookmark(b, e)
	case *document.BookmarkEnd:
		w.writeBookmarkEnd(b, e)
	case *document.Comment:
		w.writeCommentRef(b, e)
	case *document.PreserveText:
		w.writePreserveText(b, e)
	case *document.TOC:
		w.writeTOC(b, e)
	case *document.CheckBox:
		w.writeCheckBox(b, e)
	case *document.Line:
		w.writeLine(b, e)
	case *document.FormField:
		w.writeFormField(b, e)
	}
}

func (w *Writer) writeParagraph(b *strings.Builder, p *document.Paragraph) {
	b.WriteString(`<w:p>`)
	ppr := paraStyleToXML(&p.Style, p.StyleName)
	if ppr != "" {
		b.WriteString(ppr)
	}
	for _, run := range p.Runs {
		w.writeRun(b, run)
	}
	b.WriteString(`</w:p>`)
}

func (w *Writer) writeRun(b *strings.Builder, r *document.Run) {
	if r.Break {
		b.WriteString(`<w:r><w:br/></w:r>`)
		return
	}
	b.WriteString(`<w:r>`)
	if r.StyleName != "" {
		b.WriteString(fmt.Sprintf(`<w:rPr><w:rStyle w:val="%s"/></w:rPr>`, escapeXML(r.StyleName)))
	} else {
		rpr := fontStyleToXML(&r.Style, nil)
		if rpr != "" {
			b.WriteString(rpr)
		}
	}
	w.writeTextElement(b, r.Text)
	b.WriteString(`</w:r>`)
}
