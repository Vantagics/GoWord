package ooxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
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
		// Strict OOXML (purl.oclc.org) parts are not copied. The writer emits
		// transitional markup instead, which Word, LibreOffice, POI, and pandoc
		// all read.
		if bytes.Contains(data, []byte("purl.oclc.org/ooxml")) {
			continue
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
		// add the transitional relationship.
		if strings.Contains(rel.Type, "purl.oclc.org/ooxml") {
			continue
		}
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
	if w.hasFile("word/fontTable.xml") {
		return nil
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
	b.WriteString(fmt.Sprintf(`<w:fonts xmlns:w="%s" xmlns:r="%s">`, nsW, nsR))
	for _, name := range names {
		b.WriteString(fmt.Sprintf(`<w:font w:name="%s"><w:charset w:val="00"/><w:family w:val="auto"/><w:pitch w:val="variable"/></w:font>`, escapeXML(name)))
	}
	b.WriteString(`</w:fonts>`)
	return w.addZipFile("word/fontTable.xml", []byte(b.String()))
}
