package ooxml

import (
	"fmt"
	"strings"

	"github.com/VantageDataChat/GoWord/common"
	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

func (w *Writer) writeTextElement(b *strings.Builder, text string) {
	if text == "" {
		return
	}
	// Use xml:space="preserve" if text has leading/trailing whitespace
	if strings.HasPrefix(text, " ") || strings.HasSuffix(text, " ") || strings.Contains(text, "  ") {
		b.WriteString(`<w:t xml:space="preserve">`)
	} else {
		b.WriteString(`<w:t>`)
	}
	b.WriteString(escapeXML(text))
	b.WriteString(`</w:t>`)
}

func (w *Writer) writeTextRun(b *strings.Builder, tr *document.TextRun) {
	b.WriteString(`<w:p>`)
	ppr := paraStyleToXML(&tr.Style, tr.StyleName)
	if ppr != "" {
		b.WriteString(ppr)
	}
	for _, elem := range tr.Elements {
		switch e := elem.(type) {
		case *document.Paragraph:
			// Inline text within textrun - write runs directly
			for _, run := range e.Runs {
				w.writeRun(b, run)
			}
		case *document.Hyperlink:
			w.writeHyperlinkInline(b, e)
		case *document.TextBreak:
			for i := 0; i < e.Count; i++ {
				b.WriteString(`<w:r><w:br/></w:r>`)
			}
		case *document.Image:
			w.writeInlineImage(b, e)
		case *document.Footnote:
			w.writeFootnoteRef(b, e)
		case *document.Endnote:
			w.writeEndnoteRef(b, e)
		case *document.Bookmark:
			w.writeBookmark(b, e)
		}
	}
	b.WriteString(`</w:p>`)
}

func (w *Writer) writeTextBreak(b *strings.Builder, tb *document.TextBreak) {
	for i := 0; i < tb.Count; i++ {
		b.WriteString(`<w:p>`)
		if tb.Font != nil || tb.Para != nil {
			if tb.Para != nil {
				b.WriteString(paraStyleToXML(tb.Para, ""))
			}
		}
		b.WriteString(`</w:p>`)
	}
}

func (w *Writer) writePageBreak(b *strings.Builder) {
	b.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
}

func (w *Writer) writeHyperlink(b *strings.Builder, h *document.Hyperlink) {
	relID := w.addDocRelExternal(relHyperlink, h.URL)
	b.WriteString(`<w:p>`)
	b.WriteString(fmt.Sprintf(`<w:hyperlink r:id="%s">`, relID))
	b.WriteString(`<w:r>`)
	b.WriteString(`<w:rPr><w:rStyle w:val="Hyperlink"/>`)
	rpr := fontStyleToXML(&h.Font, nil)
	// Strip outer rPr tags since we already opened one
	rpr = strings.TrimPrefix(rpr, "<w:rPr>")
	rpr = strings.TrimSuffix(rpr, "</w:rPr>")
	b.WriteString(rpr)
	b.WriteString(`</w:rPr>`)
	w.writeTextElement(b, h.Text)
	b.WriteString(`</w:r>`)
	b.WriteString(`</w:hyperlink>`)
	b.WriteString(`</w:p>`)
}

func (w *Writer) writeHyperlinkInline(b *strings.Builder, h *document.Hyperlink) {
	relID := w.addDocRelExternal(relHyperlink, h.URL)
	b.WriteString(fmt.Sprintf(`<w:hyperlink r:id="%s">`, relID))
	b.WriteString(`<w:r>`)
	b.WriteString(`<w:rPr><w:rStyle w:val="Hyperlink"/></w:rPr>`)
	w.writeTextElement(b, h.Text)
	b.WriteString(`</w:r>`)
	b.WriteString(`</w:hyperlink>`)
}

// Ensure imports are used.
var _ = common.PixelToEmu
var _ style.FontStyle
