package ooxml

import (
	"fmt"
	"sort"
	"strings"
)

func (w *Writer) writeStyles() error {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:styles xmlns:w="%s" xmlns:r="%s">`, nsW, nsR))

	// Default styles
	df := w.doc.DefaultFont
	b.WriteString(`<w:docDefaults>`)
	b.WriteString(`<w:rPrDefault><w:rPr>`)
	b.WriteString(fmt.Sprintf(`<w:rFonts w:ascii="%s" w:hAnsi="%s" w:cs="%s"/>`, escapeXML(df.Name), escapeXML(df.Name), escapeXML(df.Name)))
	hp := int(df.Size * 2)
	b.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/>`, hp))
	b.WriteString(fmt.Sprintf(`<w:szCs w:val="%d"/>`, hp))
	b.WriteString(`</w:rPr></w:rPrDefault>`)
	b.WriteString(`<w:pPrDefault><w:pPr>`)
	b.WriteString(`<w:spacing w:after="0" w:line="240" w:lineRule="auto"/>`)
	b.WriteString(`</w:pPr></w:pPrDefault>`)
	b.WriteString(`</w:docDefaults>`)

	// Normal style
	b.WriteString(`<w:style w:type="paragraph" w:default="1" w:styleId="Normal">`)
	b.WriteString(`<w:name w:val="Normal"/>`)
	b.WriteString(`<w:qFormat/>`)
	b.WriteString(`</w:style>`)

	// Heading styles (1-9)
	for i := 1; i <= 9; i++ {
		b.WriteString(fmt.Sprintf(`<w:style w:type="paragraph" w:styleId="Heading%d">`, i))
		b.WriteString(fmt.Sprintf(`<w:name w:val="heading %d"/>`, i))
		b.WriteString(`<w:basedOn w:val="Normal"/>`)
		b.WriteString(`<w:next w:val="Normal"/>`)
		b.WriteString(`<w:qFormat/>`)
		b.WriteString(`<w:pPr><w:keepNext/><w:keepLines/>`)
		b.WriteString(fmt.Sprintf(`<w:outlineLvl w:val="%d"/>`, i-1))
		b.WriteString(`</w:pPr>`)
		b.WriteString(`<w:rPr>`)
		fontSize := 28 - (i-1)*2
		if fontSize < 16 {
			fontSize = 16
		}
		b.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/>`, fontSize))
		b.WriteString(fmt.Sprintf(`<w:szCs w:val="%d"/>`, fontSize))
		b.WriteString(`<w:b/>`)
		b.WriteString(`</w:rPr>`)
		b.WriteString(`</w:style>`)
	}

	// Title style
	b.WriteString(`<w:style w:type="paragraph" w:styleId="Title">`)
	b.WriteString(`<w:name w:val="Title"/>`)
	b.WriteString(`<w:basedOn w:val="Normal"/>`)
	b.WriteString(`<w:qFormat/>`)
	b.WriteString(`<w:pPr><w:jc w:val="center"/></w:pPr>`)
	b.WriteString(`<w:rPr><w:sz w:val="56"/><w:szCs w:val="56"/><w:b/></w:rPr>`)
	b.WriteString(`</w:style>`)

	// Hyperlink character style
	b.WriteString(`<w:style w:type="character" w:styleId="Hyperlink">`)
	b.WriteString(`<w:name w:val="Hyperlink"/>`)
	b.WriteString(`<w:rPr><w:color w:val="0563C1"/><w:u w:val="single"/></w:rPr>`)
	b.WriteString(`</w:style>`)

	// FootnoteReference character style
	b.WriteString(`<w:style w:type="character" w:styleId="FootnoteReference">`)
	b.WriteString(`<w:name w:val="footnote reference"/>`)
	b.WriteString(`<w:rPr><w:vertAlign w:val="superscript"/></w:rPr>`)
	b.WriteString(`</w:style>`)

	// FootnoteText paragraph style
	b.WriteString(`<w:style w:type="paragraph" w:styleId="FootnoteText">`)
	b.WriteString(`<w:name w:val="footnote text"/>`)
	b.WriteString(`<w:basedOn w:val="Normal"/>`)
	b.WriteString(`<w:rPr><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr>`)
	b.WriteString(`</w:style>`)

	// EndnoteReference character style
	b.WriteString(`<w:style w:type="character" w:styleId="EndnoteReference">`)
	b.WriteString(`<w:name w:val="endnote reference"/>`)
	b.WriteString(`<w:rPr><w:vertAlign w:val="superscript"/></w:rPr>`)
	b.WriteString(`</w:style>`)

	// EndnoteText paragraph style
	b.WriteString(`<w:style w:type="paragraph" w:styleId="EndnoteText">`)
	b.WriteString(`<w:name w:val="endnote text"/>`)
	b.WriteString(`<w:basedOn w:val="Normal"/>`)
	b.WriteString(`<w:rPr><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr>`)
	b.WriteString(`</w:style>`)

	// User-defined font styles as character styles (sorted for deterministic output)
	fontStyles := w.doc.FontStyles()
	fontNames := make([]string, 0, len(fontStyles))
	for name := range fontStyles {
		fontNames = append(fontNames, name)
	}
	sort.Strings(fontNames)
	for _, name := range fontNames {
		fs := fontStyles[name]
		b.WriteString(fmt.Sprintf(`<w:style w:type="character" w:styleId="%s">`, escapeXML(name)))
		b.WriteString(fmt.Sprintf(`<w:name w:val="%s"/>`, escapeXML(name)))
		b.WriteString(fontStyleToXML(&fs, nil))
		b.WriteString(`</w:style>`)
	}

	// User-defined paragraph styles (sorted for deterministic output)
	paraStyles := w.doc.ParagraphStyles()
	paraNames := make([]string, 0, len(paraStyles))
	for name := range paraStyles {
		paraNames = append(paraNames, name)
	}
	sort.Strings(paraNames)
	for _, name := range paraNames {
		ps := paraStyles[name]
		b.WriteString(fmt.Sprintf(`<w:style w:type="paragraph" w:styleId="%s">`, escapeXML(name)))
		b.WriteString(fmt.Sprintf(`<w:name w:val="%s"/>`, escapeXML(name)))
		b.WriteString(`<w:basedOn w:val="Normal"/>`)
		b.WriteString(paraStyleToXML(&ps, ""))
		b.WriteString(`</w:style>`)
	}

	b.WriteString(`</w:styles>`)
	return w.addZipFile("word/styles.xml", []byte(b.String()))
}
