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
		b.WriteString(`<w:p>`)
		w.writeFootnoteRef(b, e)
		b.WriteString(`</w:p>`)
	case *document.Endnote:
		b.WriteString(`<w:p>`)
		w.writeEndnoteRef(b, e)
		b.WriteString(`</w:p>`)
	case *document.Bookmark:
		w.writeBookmark(b, e)
	case *document.BookmarkEnd:
		w.writeBookmarkEnd(b, e)
	case *document.Comment:
		b.WriteString(`<w:p>`)
		w.writeCommentRef(b, e)
		b.WriteString(`</w:p>`)
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
	case *document.RawXML:
		b.WriteString(e.XML)
	case *document.Field:
		b.WriteString(`<w:p>`)
		w.writeField(b, e)
		b.WriteString(`</w:p>`)
	case *document.Sdt:
		w.writeSdt(b, e)
	case *document.CommentRangeStart:
		b.WriteString(fmt.Sprintf(`<w:commentRangeStart w:id="%d"/>`, e.ID))
	case *document.CommentRangeEnd:
		b.WriteString(fmt.Sprintf(`<w:commentRangeEnd w:id="%d"/>`, e.ID))
	case *document.Run:
		w.writeRun(b, e)
	}
}

func (w *Writer) writeParagraph(b *strings.Builder, p *document.Paragraph) {
	w.writeParagraphSect(b, p, "")
}

// writeParagraphSect writes a paragraph. sectXML, when non-empty, is a full
// w:sectPr element inserted as the last child of pPr (the section-break paragraph).
func (w *Writer) writeParagraphSect(b *strings.Builder, p *document.Paragraph, sectXML string) {
	b.WriteString(`<w:p>`)
	switch {
	case p.RawPPr != "" && sectXML == "":
		b.WriteString(p.RawPPr)
	case p.RawPPr != "" && strings.HasSuffix(strings.TrimSpace(p.RawPPr), "</w:pPr>"):
		trimmed := strings.TrimSpace(p.RawPPr)
		trimmed = strings.TrimSuffix(trimmed, "</w:pPr>")
		b.WriteString(trimmed)
		b.WriteString(sectXML)
		b.WriteString("</w:pPr>")
	default:
		ppr := paraStyleToXML(&p.Style, p.StyleName, w.numID(p.Style.NumStyleName))
		if sectXML != "" {
			if ppr == "" {
				b.WriteString(`<w:pPr>`)
				b.WriteString(sectXML)
				b.WriteString(`</w:pPr>`)
			} else {
				b.WriteString(strings.TrimSuffix(ppr, "</w:pPr>"))
				b.WriteString(sectXML)
				b.WriteString("</w:pPr>")
			}
		} else if ppr != "" {
			b.WriteString(ppr)
		}
	}
	if len(p.Inlines) > 0 {
		for _, el := range p.Inlines {
			w.writeInline(b, el)
		}
	} else {
		for _, run := range p.Runs {
			w.writeRun(b, run)
		}
	}
	b.WriteString(`</w:p>`)
}

func (w *Writer) writeInline(b *strings.Builder, elem document.Element) {
	switch e := elem.(type) {
	case *document.Run:
		w.writeRun(b, e)
	case *document.Hyperlink:
		w.writeHyperlinkInline(b, e)
	case *document.Image:
		w.writeInlineImage(b, e)
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
	case *document.Field:
		w.writeField(b, e)
	case *document.Tab:
		b.WriteString(`<w:r><w:tab/></w:r>`)
	case *document.Ruby:
		if e.XML != "" {
			w.writeRawInline(b, e.XML)
		}
	case *document.RawXML:
		w.writeRawInline(b, e.XML)
	case *document.CommentRangeStart:
		b.WriteString(fmt.Sprintf(`<w:commentRangeStart w:id="%d"/>`, e.ID))
	case *document.CommentRangeEnd:
		b.WriteString(fmt.Sprintf(`<w:commentRangeEnd w:id="%d"/>`, e.ID))
	case *document.TextBreak:
		for i := 0; i < e.Count; i++ {
			b.WriteString(`<w:r><w:br/></w:r>`)
		}
	default:
		w.writeElementXML(b, elem)
	}
}

func (w *Writer) writeRun(b *strings.Builder, r *document.Run) {
	b.WriteString(`<w:r>`)
	if r.RawRPr != "" {
		b.WriteString(r.RawRPr)
	} else {
		rpr := fontStyleToXML(&r.Style, nil)
		if r.StyleName != "" {
			styleEl := fmt.Sprintf(`<w:rStyle w:val="%s"/>`, escapeXML(r.StyleName))
			if rpr == "" {
				rpr = `<w:rPr>` + styleEl + `</w:rPr>`
			} else {
				rpr = strings.Replace(rpr, "<w:rPr>", "<w:rPr>"+styleEl, 1)
			}
		}
		if rpr != "" {
			b.WriteString(rpr)
		}
	}
	if r.Tab {
		b.WriteString(`<w:tab/>`)
	}
	if r.Break || r.BreakType != "" {
		bt := r.BreakType
		if bt == "" || bt == "textWrapping" {
			b.WriteString(`<w:br/>`)
		} else {
			b.WriteString(fmt.Sprintf(`<w:br w:type="%s"/>`, bt))
		}
	}
	w.writeTextElement(b, r.Text)
	b.WriteString(`</w:r>`)
}

func startsWithTag(s, tag string) bool {
	prefix := "<" + tag
	if !strings.HasPrefix(s, prefix) {
		return false
	}
	if len(s) == len(prefix) {
		return true
	}
	switch s[len(prefix)] {
	case '>', ' ', '/', '\n', '\r', '\t':
		return true
	default:
		return false
	}
}

func (w *Writer) writeRawInline(b *strings.Builder, raw string) {
	trim := strings.TrimSpace(raw)
	// Drawings and VML pictures are run-level. <w:pict> begins with the
	// characters <w:p, so the check has to be a real tag boundary. A bare
	// w:pict under w:p is ignored by LibreOffice and POI.
	needsRun := (strings.Contains(trim, "<w:pict") || strings.Contains(trim, "<w:drawing") || strings.Contains(trim, "<w:ruby") || strings.Contains(trim, "AlternateContent")) &&
		!startsWithTag(trim, "w:r") && !startsWithTag(trim, "w:p") && !startsWithTag(trim, "w:ins") && !startsWithTag(trim, "w:del") && !startsWithTag(trim, "w:moveTo") && !startsWithTag(trim, "w:moveFrom")
	if needsRun {
		b.WriteString(`<w:r>`)
		b.WriteString(raw)
		b.WriteString(`</w:r>`)
		return
	}
	b.WriteString(raw)
}

func (w *Writer) writeField(b *strings.Builder, f *document.Field) {
	b.WriteString(`<w:r><w:fldChar w:fldCharType="begin"/></w:r>`)
	b.WriteString(`<w:r><w:instrText xml:space="preserve"> `)
	b.WriteString(escapeXML(strings.TrimSpace(f.Instr)))
	b.WriteString(` </w:instrText></w:r>`)
	b.WriteString(`<w:r><w:fldChar w:fldCharType="separate"/></w:r>`)
	if f.Result != "" {
		b.WriteString(`<w:r>`)
		if rpr := fontStyleToXML(&f.Font, nil); rpr != "" {
			b.WriteString(rpr)
		}
		w.writeTextElement(b, f.Result)
		b.WriteString(`</w:r>`)
	}
	b.WriteString(`<w:r><w:fldChar w:fldCharType="end"/></w:r>`)
}

func (w *Writer) writeSdt(b *strings.Builder, s *document.Sdt) {
	b.WriteString(`<w:sdt>`)
	if s.PrXML != "" {
		b.WriteString(s.PrXML)
	} else {
		b.WriteString(`<w:sdtPr><w:id w:val="1"/></w:sdtPr>`)
	}
	if s.EndPrXML != "" {
		b.WriteString(s.EndPrXML)
	}
	b.WriteString(`<w:sdtContent>`)
	if len(s.Elements) == 0 {
		b.WriteString(`<w:p/>`)
	} else {
		for _, el := range s.Elements {
			w.writeElementXML(b, el)
		}
	}
	b.WriteString(`</w:sdtContent></w:sdt>`)
}
