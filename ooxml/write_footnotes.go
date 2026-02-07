package ooxml

import (
	"fmt"
	"strings"
)

func (w *Writer) writeFootnotes() error {
	footnotes := w.doc.Footnotes()
	if len(footnotes) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:footnotes xmlns:w="%s" xmlns:r="%s">`, nsW, nsR))

	// Separator footnotes (required by spec)
	b.WriteString(`<w:footnote w:type="separator" w:id="-1">`)
	b.WriteString(`<w:p><w:r><w:separator/></w:r></w:p>`)
	b.WriteString(`</w:footnote>`)
	b.WriteString(`<w:footnote w:type="continuationSeparator" w:id="0">`)
	b.WriteString(`<w:p><w:r><w:continuationSeparator/></w:r></w:p>`)
	b.WriteString(`</w:footnote>`)

	for _, fn := range footnotes {
		b.WriteString(fmt.Sprintf(`<w:footnote w:id="%d">`, fn.ID))
		for _, elem := range fn.Elements {
			w.writeElementXML(&b, elem)
		}
		b.WriteString(`</w:footnote>`)
	}

	b.WriteString(`</w:footnotes>`)
	return w.addZipFile("word/footnotes.xml", []byte(b.String()))
}

func (w *Writer) writeEndnotes() error {
	endnotes := w.doc.Endnotes()
	if len(endnotes) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:endnotes xmlns:w="%s" xmlns:r="%s">`, nsW, nsR))

	// Separator endnotes
	b.WriteString(`<w:endnote w:type="separator" w:id="-1">`)
	b.WriteString(`<w:p><w:r><w:separator/></w:r></w:p>`)
	b.WriteString(`</w:endnote>`)
	b.WriteString(`<w:endnote w:type="continuationSeparator" w:id="0">`)
	b.WriteString(`<w:p><w:r><w:continuationSeparator/></w:r></w:p>`)
	b.WriteString(`</w:endnote>`)

	for _, en := range endnotes {
		b.WriteString(fmt.Sprintf(`<w:endnote w:id="%d">`, en.ID))
		for _, elem := range en.Elements {
			w.writeElementXML(&b, elem)
		}
		b.WriteString(`</w:endnote>`)
	}

	b.WriteString(`</w:endnotes>`)
	return w.addZipFile("word/endnotes.xml", []byte(b.String()))
}

func (w *Writer) writeComments() error {
	comments := w.doc.Comments()
	if len(comments) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:comments xmlns:w="%s" xmlns:r="%s">`, nsW, nsR))

	for _, c := range comments {
		b.WriteString(fmt.Sprintf(`<w:comment w:id="%d" w:author="%s" w:date="%s">`,
			c.ID, escapeXML(c.Author), c.Date.UTC().Format("2006-01-02T15:04:05Z")))
		if c.Text != "" {
			b.WriteString(`<w:p><w:r><w:t>`)
			b.WriteString(escapeXML(c.Text))
			b.WriteString(`</w:t></w:r></w:p>`)
		}
		for _, elem := range c.Elements {
			w.writeElementXML(&b, elem)
		}
		b.WriteString(`</w:comment>`)
	}

	b.WriteString(`</w:comments>`)
	return w.addZipFile("word/comments.xml", []byte(b.String()))
}
