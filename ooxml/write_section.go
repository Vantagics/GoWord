package ooxml

import (
	"fmt"
	"strings"

	"github.com/VantageDataChat/GoWord/document"
)

func (w *Writer) writeSectionProperties(b *strings.Builder, sec *document.Section, isLast bool) {
	if isLast {
		// Last section properties go inside w:body > w:sectPr
		b.WriteString(`<w:sectPr>`)
	} else {
		// Non-last sections: properties go inside the last paragraph's pPr
		b.WriteString(`<w:p><w:pPr><w:sectPr>`)
	}

	// Header/footer references
	for _, hp := range w.headerParts {
		for _, h := range sec.Headers {
			if w.headerPartMatches(hp.filename, h) {
				relID := w.addDocRel(relHeader, hp.filename)
				b.WriteString(fmt.Sprintf(`<w:headerReference w:type="%s" r:id="%s"/>`, h.Type, relID))
			}
		}
	}
	for _, fp := range w.footerParts {
		for _, f := range sec.Footers {
			if w.footerPartMatches(fp.filename, f) {
				relID := w.addDocRel(relFooter, fp.filename)
				b.WriteString(fmt.Sprintf(`<w:footerReference w:type="%s" r:id="%s"/>`, f.Type, relID))
			}
		}
	}

	ss := &sec.Style
	b.WriteString(fmt.Sprintf(`<w:pgSz w:w="%d" w:h="%d"`, ss.PageWidth, ss.PageHeight))
	if ss.Orientation == "landscape" {
		b.WriteString(` w:orient="landscape"`)
	}
	b.WriteString(`/>`)

	b.WriteString(fmt.Sprintf(`<w:pgMar w:top="%d" w:right="%d" w:bottom="%d" w:left="%d" w:header="%d" w:footer="%d"/>`,
		ss.MarginTop, ss.MarginRight, ss.MarginBottom, ss.MarginLeft, ss.HeaderHeight, ss.FooterHeight))

	if ss.ColumnCount > 1 {
		b.WriteString(fmt.Sprintf(`<w:cols w:num="%d"`, ss.ColumnCount))
		if ss.ColumnSpacing > 0 {
			b.WriteString(fmt.Sprintf(` w:space="%d"`, ss.ColumnSpacing))
		}
		b.WriteString(`/>`)
	}

	if ss.PageNumberStart != nil {
		b.WriteString(fmt.Sprintf(`<w:pgNumType w:start="%d"/>`, *ss.PageNumberStart))
	}

	if ss.BreakType != "" && ss.BreakType != "nextPage" {
		b.WriteString(fmt.Sprintf(`<w:type w:val="%s"/>`, ss.BreakType))
	}

	if isLast {
		b.WriteString(`</w:sectPr>`)
	} else {
		b.WriteString(`</w:sectPr></w:pPr></w:p>`)
	}
}

func (w *Writer) generateHeaderFooterParts() {
	headerIdx := 1
	footerIdx := 1

	for _, sec := range w.doc.Sections {
		for _, h := range sec.Headers {
			filename := fmt.Sprintf("header%d.xml", headerIdx)
			content := w.generateHeaderXML(h)
			w.headerParts = append(w.headerParts, headerFooterPart{
				filename: filename,
				content:  content,
			})
			headerIdx++
		}
		for _, f := range sec.Footers {
			filename := fmt.Sprintf("footer%d.xml", footerIdx)
			content := w.generateFooterXML(f)
			w.footerParts = append(w.footerParts, headerFooterPart{
				filename: filename,
				content:  content,
			})
			footerIdx++
		}
	}
}

func (w *Writer) generateHeaderXML(h *document.Header) []byte {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:hdr xmlns:w="%s" xmlns:r="%s" xmlns:wp="%s" xmlns:a="%s" xmlns:pic="%s">`,
		nsW, nsR, nsWP, nsA, nsPIC))
	for _, elem := range h.Elements {
		w.writeElementXML(&b, elem)
	}
	if len(h.Elements) == 0 {
		b.WriteString(`<w:p/>`)
	}
	b.WriteString(`</w:hdr>`)
	return []byte(b.String())
}

func (w *Writer) generateFooterXML(f *document.Footer) []byte {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:ftr xmlns:w="%s" xmlns:r="%s" xmlns:wp="%s" xmlns:a="%s" xmlns:pic="%s">`,
		nsW, nsR, nsWP, nsA, nsPIC))
	for _, elem := range f.Elements {
		w.writeElementXML(&b, elem)
	}
	if len(f.Elements) == 0 {
		b.WriteString(`<w:p/>`)
	}
	b.WriteString(`</w:ftr>`)
	return []byte(b.String())
}

func (w *Writer) writeHeaderFooterFiles() error {
	for _, hp := range w.headerParts {
		if err := w.addZipFile("word/"+hp.filename, hp.content); err != nil {
			return err
		}
	}
	for _, fp := range w.footerParts {
		if err := w.addZipFile("word/"+fp.filename, fp.content); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) headerPartMatches(filename string, h *document.Header) bool {
	for _, hp := range w.headerParts {
		if hp.filename == filename {
			// Check if this part was generated from this header
			expected := w.generateHeaderXML(h)
			return string(hp.content) == string(expected)
		}
	}
	return false
}

func (w *Writer) footerPartMatches(filename string, f *document.Footer) bool {
	for _, fp := range w.footerParts {
		if fp.filename == filename {
			expected := w.generateFooterXML(f)
			return string(fp.content) == string(expected)
		}
	}
	return false
}
