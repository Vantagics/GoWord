package ooxml

import (
	"fmt"
	"strings"
)

func (w *Writer) writeDocument() error {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:document xmlns:w="%s" xmlns:r="%s" xmlns:wp="%s" xmlns:a="%s" xmlns:pic="%s" xmlns:mc="%s">`,
		nsW, nsR, nsWP, nsA, nsPIC, nsMC))
	b.WriteString(`<w:body>`)

	for i, sec := range w.doc.Sections {
		for _, elem := range sec.Elements {
			w.writeElementXML(&b, elem)
		}
		// Section properties
		w.writeSectionProperties(&b, sec, i == len(w.doc.Sections)-1)
	}

	b.WriteString(`</w:body>`)
	b.WriteString(`</w:document>`)
	return w.addZipFile("word/document.xml", []byte(b.String()))
}
