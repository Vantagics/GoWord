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
	ppr := paraStyleToXML(&tr.Style, tr.StyleName, w.numID(tr.Style.NumStyleName))
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
		case *document.Tab:
			b.WriteString(`<w:r><w:tab/></w:r>`)
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
	b.WriteString(`<w:p>`)
	if ppr := paraStyleToXML(&h.Para, ""); ppr != "" {
		b.WriteString(ppr)
	}
	w.writeHyperlinkInline(b, h)
	b.WriteString(`</w:p>`)
}

// writeHyperlinkFontProps writes font style properties without the outer <w:rPr> wrapper.
func writeHyperlinkFontProps(b *strings.Builder, f *style.FontStyle) {
	if f == nil {
		return
	}
	// Same CT_RPr order as fontStyleToXML, without the rPr wrapper.
	if f.Name != "" {
		b.WriteString(fmt.Sprintf(`<w:rFonts w:ascii="%s" w:hAnsi="%s" w:cs="%s"/>`, escapeXML(f.Name), escapeXML(f.Name), escapeXML(f.Name)))
	}
	if f.Bold {
		b.WriteString(`<w:b/><w:bCs/>`)
	}
	if f.Italic {
		b.WriteString(`<w:i/><w:iCs/>`)
	}
	if f.Strikethrough {
		b.WriteString(`<w:strike/>`)
	}
	if f.DoubleStrikethrough {
		b.WriteString(`<w:dstrike/>`)
	}
	if f.Color != "" {
		b.WriteString(fmt.Sprintf(`<w:color w:val="%s"/>`, f.Color))
	}
	if f.Size > 0 {
		hp := int(f.Size * 2)
		b.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/>`, hp))
		b.WriteString(fmt.Sprintf(`<w:szCs w:val="%d"/>`, hp))
	}
	if f.Underline != "" && f.Underline != "none" {
		b.WriteString(fmt.Sprintf(`<w:u w:val="%s"/>`, f.Underline))
	}
}

func (w *Writer) writeHyperlinkInline(b *strings.Builder, h *document.Hyperlink) {
	b.WriteString(`<w:hyperlink`)
	if h.IsInternal {
		b.WriteString(fmt.Sprintf(` w:anchor="%s"`, escapeXML(h.URL)))
	} else if h.URL != "" {
		relID := w.addDocRelExternal(relHyperlink, h.URL)
		b.WriteString(fmt.Sprintf(` r:id="%s"`, relID))
	}
	if h.Tooltip != "" {
		b.WriteString(fmt.Sprintf(` w:tooltip="%s"`, escapeXML(h.Tooltip)))
	}
	b.WriteString(`>`)
	b.WriteString(`<w:r>`)
	b.WriteString(`<w:rPr><w:rStyle w:val="Hyperlink"/>`)
	writeHyperlinkFontProps(b, &h.Font)
	b.WriteString(`</w:rPr>`)
	w.writeTextElement(b, h.Text)
	b.WriteString(`</w:r></w:hyperlink>`)
}

// Ensure imports are used.
var _ = common.PixelToEmu
