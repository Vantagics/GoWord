package ooxml

import (
	"strings"

	"github.com/VantageDataChat/GoWord/document"
)

func (w *Writer) writeDocument() error {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(w.wordMLRoot("w:document"))
	b.WriteString(`<w:body>`)

	if len(w.doc.Sections) == 0 {
		w.doc.AddSection()
	}
	for i, sec := range w.doc.Sections {
		w.writeSectionContent(&b, sec, i == len(w.doc.Sections)-1)
	}

	b.WriteString(`</w:body>`)
	b.WriteString(`</w:document>`)
	return w.addZipFile("word/document.xml", []byte(b.String()))
}

// writeSectionContent writes one section. A non-final section's sectPr is
// attached to its last paragraph when that paragraph exists, which is how
// Word stores section breaks without inserting an extra empty paragraph.
func (w *Writer) writeSectionContent(b *strings.Builder, sec *document.Section, isLast bool) {
	sectEl := ""
	if !isLast {
		sectEl = w.sectPrElement(sec)
	}
	if sectEl != "" && len(sec.Elements) > 0 {
		if p, ok := sec.Elements[len(sec.Elements)-1].(*document.Paragraph); ok {
			for _, el := range sec.Elements[:len(sec.Elements)-1] {
				w.writeElementXML(b, el)
			}
			w.writeParagraphSect(b, p, sectEl)
			return
		}
	}
	for _, el := range sec.Elements {
		w.writeElementXML(b, el)
	}
	if isLast {
		b.WriteString(w.sectPrElement(sec))
		return
	}
	if sectEl == "" {
		w.writeSectionProperties(b, sec, false)
		return
	}
	b.WriteString(`<w:p><w:pPr>`)
	b.WriteString(sectEl)
	b.WriteString(`</w:pPr></w:p>`)
}
