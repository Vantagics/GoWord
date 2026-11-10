package ooxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/VantageDataChat/GoWord/document"
)

func (w *Writer) copyOriginalParts() {
	main := w.doc.MainPart()
	relPath := filepath.ToSlash(filepath.Join(filepath.Dir(main), "_rels", filepath.Base(main)+".rels"))
	skip := map[string]bool{
		"[Content_Types].xml": true,
		"_rels/.rels":         true,
		main:                  true,
		relPath:               true,
		"docProps/core.xml":   true,
		"docProps/app.xml":    true,
	}
	names := w.doc.OriginalPartNames()
	sort.Strings(names)
	for _, name := range names {
		if skip[name] || strings.HasSuffix(name, "/") {
			continue
		}
		data, ok := w.doc.OriginalPart(name)
		if !ok {
			continue
		}
		// Strict OOXML (purl.oclc.org) parts are not copied unless this save
		// was asked to stay strict. The transitional writer emits markup
		// Word, LibreOffice, POI, and pandoc all read.
		if w.doc.Strict() {
			data = toStrictNamespaces(data)
		} else if bytes.Contains(data, []byte("purl.oclc.org/ooxml")) {
			continue
		}
		if name == "word/comments.xml" {
			// LibreOffice refuses a comments part that contains mc:AlternateContent
			// (a drawing inside a comment). Drop those elements and keep the text.
			data = stripAlternateContent(data)
		}
		_ = w.addZipFile(name, data)
	}
}

func (w *Writer) seedDocRels() {
	part := w.doc.MainPart()
	relPath := filepath.ToSlash(filepath.Join(filepath.Dir(part), "_rels", filepath.Base(part)+".rels"))
	data, ok := w.doc.OriginalPart(relPath)
	if !ok {
		return
	}
	var rels xmlRelationships
	if err := xml.Unmarshal(data, &rels); err != nil {
		return
	}
	max := 0
	for _, rel := range rels.Relationship {
		// Strict relationship types point at a different part class than the
		// transitional parts this writer emits. Drop them and let ensureRel
		// add the transitional relationship. A strict save keeps them.
		if !w.doc.Strict() && strings.Contains(rel.Type, "purl.oclc.org/ooxml") {
			continue
		}
		rel.Type = w.relType(rel.Type)
		w.docRels = append(w.docRels, relationship{
			ID:         rel.ID,
			Type:       rel.Type,
			Target:     rel.Target,
			TargetMode: rel.TargetMode,
		})
		if n := relNumber(rel.ID); n > max {
			max = n
		}
	}
	if max+1 > w.nextDocRelID {
		w.nextDocRelID = max + 1
	}
}

func relNumber(id string) int {
	id = strings.TrimPrefix(id, "rId")
	n := 0
	for _, r := range id {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func (w *Writer) writeFontTable() error {
	var fresh []document.EmbeddedFont
	for _, ef := range w.doc.EmbeddedFonts() {
		if !ef.Stored {
			fresh = append(fresh, ef)
		}
	}
	if len(fresh) == 0 && w.hasFile("word/fontTable.xml") {
		return nil
	}
	if len(fresh) > 0 {
		if err := w.writeEmbeddedFonts(fresh); err != nil {
			return err
		}
		if w.hasFile("word/fontTable.xml") {
			return nil
		}
	}
	fonts := map[string]struct{}{}
	if w.doc.DefaultFont.Name != "" {
		fonts[w.doc.DefaultFont.Name] = struct{}{}
	}
	for _, fs := range w.doc.FontStyles() {
		if fs.Name != "" {
			fonts[fs.Name] = struct{}{}
		}
		if fs.NameEastAsia != "" {
			fonts[fs.NameEastAsia] = struct{}{}
		}
	}
	if len(fonts) == 0 {
		fonts["Arial"] = struct{}{}
	}
	names := make([]string, 0, len(fonts))
	for name := range fonts {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:fonts xmlns:w="%s" xmlns:r="%s">`, w.wNS(), w.rNS()))
	for _, name := range names {
		b.WriteString(fmt.Sprintf(`<w:font w:name="%s"><w:charset w:val="00"/><w:family w:val="auto"/><w:pitch w:val="variable"/></w:font>`, escapeXML(name)))
	}
	b.WriteString(`</w:fonts>`)
	return w.addZipFile("word/fontTable.xml", []byte(b.String()))
}

func (w *Writer) writeEmbeddedFonts(embeds []document.EmbeddedFont) error {
	var rels []relationship
	var fonts strings.Builder
	for i, ef := range embeds {
		id := fmt.Sprintf("rId%d", i+1)
		target := fmt.Sprintf("fonts/font%d.odttf", i+1)
		fonts.WriteString(fmt.Sprintf(`<w:font w:name="%s"><w:embedRegular r:id="%s" w:fontKey="%s"/></w:font>`, escapeXML(ef.Name), id, escapeXML(ef.Key)))
		obf, err := document.ObfuscateFont(ef.Key, ef.Data)
		if err != nil {
			return err
		}
		if err := w.addZipFile("word/"+target, obf); err != nil {
			return err
		}
		rels = append(rels, relationship{ID: id, Type: relFont, Target: target})
	}
	var table string
	if data, ok := w.doc.OriginalPart("word/fontTable.xml"); ok && bytes.Contains(data, []byte("</w:fonts>")) {
		s := string(data)
		cut := strings.LastIndex(s, "</w:fonts>")
		table = s[:cut] + fonts.String() + "</w:fonts>"
	} else {
		var b strings.Builder
		b.WriteString(xmlHeader())
		b.WriteString(fmt.Sprintf(`<w:fonts xmlns:w="%s" xmlns:r="%s">`, w.wNS(), w.rNS()))
		b.WriteString(fonts.String())
		b.WriteString(`</w:fonts>`)
		table = b.String()
	}
	if err := w.addZipFile("word/fontTable.xml", []byte(table)); err != nil {
		return err
	}
	return w.writeRelsFile("word/_rels/fontTable.xml.rels", rels)
}

func (w *Writer) writeGlossary() error {
	blocks := w.doc.BuildingBlocks()
	if len(blocks) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(w.wordMLRoot("w:glossaryDocument"))
	b.WriteString(`<w:docParts>`)
	for _, bb := range blocks {
		b.WriteString(`<w:docPart><w:docPartPr>`)
		b.WriteString(fmt.Sprintf(`<w:name w:val="%s"/>`, escapeXML(bb.Name)))
		if bb.Style != "" {
			b.WriteString(fmt.Sprintf(`<w:style w:val="%s"/>`, escapeXML(bb.Style)))
		}
		b.WriteString(`<w:category><w:name w:val="General"/><w:gallery w:val="auto"/></w:category>`)
		b.WriteString(`</w:docPartPr><w:docPartBody><w:p><w:r><w:t>`)
		b.WriteString(escapeXML(bb.Text))
		b.WriteString(`</w:t></w:r></w:p></w:docPartBody></w:docPart>`)
	}
	b.WriteString(`</w:docParts></w:glossaryDocument>`)
	return w.addZipFile("word/glossary/document.xml", []byte(b.String()))
}
