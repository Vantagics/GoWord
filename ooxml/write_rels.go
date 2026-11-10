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
	b.WriteString(fmt.Sprintf(`<Relationship Id="rId1" Type="%s" Target="word/document.xml"/>`, w.relType(relDoc)))
	b.WriteString(fmt.Sprintf(`<Relationship Id="rId2" Type="%s" Target="docProps/core.xml"/>`, relCoreProps))
	b.WriteString(fmt.Sprintf(`<Relationship Id="rId3" Type="%s" Target="docProps/app.xml"/>`, w.relType(relExtProps)))
	b.WriteString(`</Relationships>`)
	return w.addZipFile("_rels/.rels", []byte(b.String()))
}

func (w *Writer) writeDocRels() error {
	w.ensureRel(relStyles, "styles.xml")
	w.ensureRel(relSettings, "settings.xml")
	if w.hasFile("word/numbering.xml") || len(w.doc.NumberingStyles()) > 0 {
		w.ensureRel(relNumbering, "numbering.xml")
	}
	if w.hasFile("word/fontTable.xml") {
		w.ensureRel(relFontTable, "fontTable.xml")
	}
	if w.hasFile("word/webSettings.xml") {
		w.ensureRel(relWebSettings, "webSettings.xml")
	}
	if w.hasFile("word/theme/theme1.xml") {
		w.ensureRel(relTheme, "theme/theme1.xml")
	}
	if w.hasFile("word/footnotes.xml") || len(w.doc.Footnotes()) > 0 {
		w.ensureRel(relFootnotes, "footnotes.xml")
	}
	if w.hasFile("word/endnotes.xml") || len(w.doc.Endnotes()) > 0 {
		w.ensureRel(relEndnotes, "endnotes.xml")
	}
	if w.hasFile("word/comments.xml") || len(w.doc.Comments()) > 0 {
		w.ensureRel(relComments, "comments.xml")
	}
	return w.writeRelsFile("word/_rels/document.xml.rels", w.docRels)
}

func (w *Writer) writeRelsFile(name string, rels []relationship) error {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<Relationships xmlns="%s">`, nsRels))
	for _, rel := range rels {
		if rel.TargetMode != "" {
			b.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="%s" Target="%s" TargetMode="%s"/>`,
				rel.ID, rel.Type, escapeXML(rel.Target), rel.TargetMode))
		} else {
			b.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="%s" Target="%s"/>`,
				rel.ID, rel.Type, escapeXML(rel.Target)))
		}
	}
	b.WriteString(`</Relationships>`)
	return w.addZipFile(name, []byte(b.String()))
}
