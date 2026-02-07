package ooxml

import (
	"fmt"
	"strings"
)

func (w *Writer) writeCoreProperties() error {
	p := w.doc.Properties
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<cp:coreProperties xmlns:cp="%s" xmlns:dc="%s" xmlns:dcterms="%s" xmlns:xsi="%s">`,
		nsCP, nsDC, nsDCT, nsXSI))

	if p.Title != "" {
		b.WriteString(fmt.Sprintf(`<dc:title>%s</dc:title>`, escapeXML(p.Title)))
	}
	if p.Subject != "" {
		b.WriteString(fmt.Sprintf(`<dc:subject>%s</dc:subject>`, escapeXML(p.Subject)))
	}
	if p.Creator != "" {
		b.WriteString(fmt.Sprintf(`<dc:creator>%s</dc:creator>`, escapeXML(p.Creator)))
	}
	if p.Keywords != "" {
		b.WriteString(fmt.Sprintf(`<cp:keywords>%s</cp:keywords>`, escapeXML(p.Keywords)))
	}
	if p.Description != "" {
		b.WriteString(fmt.Sprintf(`<dc:description>%s</dc:description>`, escapeXML(p.Description)))
	}
	if p.Category != "" {
		b.WriteString(fmt.Sprintf(`<cp:category>%s</cp:category>`, escapeXML(p.Category)))
	}
	if p.LastModifiedBy != "" {
		b.WriteString(fmt.Sprintf(`<cp:lastModifiedBy>%s</cp:lastModifiedBy>`, escapeXML(p.LastModifiedBy)))
	}
	if p.Revision > 0 {
		b.WriteString(fmt.Sprintf(`<cp:revision>%d</cp:revision>`, p.Revision))
	}

	b.WriteString(fmt.Sprintf(`<dcterms:created xsi:type="dcterms:W3CDTF">%s</dcterms:created>`,
		p.Created.UTC().Format("2006-01-02T15:04:05Z")))
	b.WriteString(fmt.Sprintf(`<dcterms:modified xsi:type="dcterms:W3CDTF">%s</dcterms:modified>`,
		p.Modified.UTC().Format("2006-01-02T15:04:05Z")))

	b.WriteString(`</cp:coreProperties>`)
	return w.addZipFile("docProps/core.xml", []byte(b.String()))
}

func (w *Writer) writeAppProperties() error {
	p := w.doc.Properties
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<Properties xmlns="%s" xmlns:vt="%s">`, nsEP, nsVT))
	if p.Company != "" {
		b.WriteString(fmt.Sprintf(`<Company>%s</Company>`, escapeXML(p.Company)))
	}
	if p.Manager != "" {
		b.WriteString(fmt.Sprintf(`<Manager>%s</Manager>`, escapeXML(p.Manager)))
	}
	b.WriteString(`<Application>GoWord</Application>`)
	b.WriteString(`</Properties>`)
	return w.addZipFile("docProps/app.xml", []byte(b.String()))
}
