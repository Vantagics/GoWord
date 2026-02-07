package ooxml

import (
	"fmt"
	"strings"
)

const nsRels = "http://schemas.openxmlformats.org/package/2006/relationships"

func (w *Writer) writeRootRels() error {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<Relationships xmlns="%s">`, nsRels))
	b.WriteString(fmt.Sprintf(`<Relationship Id="rId1" Type="%s" Target="word/document.xml"/>`, relDoc))
	b.WriteString(fmt.Sprintf(`<Relationship Id="rId2" Type="%s" Target="docProps/core.xml"/>`, relCoreProps))
	b.WriteString(fmt.Sprintf(`<Relationship Id="rId3" Type="%s" Target="docProps/app.xml"/>`, relExtProps))
	b.WriteString(`</Relationships>`)
	return w.addZipFile("_rels/.rels", []byte(b.String()))
}

func (w *Writer) writeDocRels() error {
	// Add standard relationships
	w.addDocRel(relStyles, "styles.xml")
	w.addDocRel(relSettings, "settings.xml")

	if len(w.doc.NumberingStyles()) > 0 {
		w.addDocRel(relNumbering, "numbering.xml")
	}
	if len(w.doc.Footnotes()) > 0 {
		w.addDocRel(relFootnotes, "footnotes.xml")
	}
	if len(w.doc.Endnotes()) > 0 {
		w.addDocRel(relEndnotes, "endnotes.xml")
	}
	if len(w.doc.Comments()) > 0 {
		w.addDocRel(relComments, "comments.xml")
	}

	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<Relationships xmlns="%s">`, nsRels))
	for _, rel := range w.docRels {
		if rel.TargetMode != "" {
			b.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="%s" Target="%s" TargetMode="%s"/>`,
				rel.ID, rel.Type, escapeXML(rel.Target), rel.TargetMode))
		} else {
			b.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="%s" Target="%s"/>`,
				rel.ID, rel.Type, rel.Target))
		}
	}
	b.WriteString(`</Relationships>`)
	return w.addZipFile("word/_rels/document.xml.rels", []byte(b.String()))
}
