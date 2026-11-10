package ooxml

import (
	"fmt"
	"sort"
	"strings"
)

func (w *Writer) writeNumbering() error {
	if !w.doc.NumberingDirty() {
		if _, ok := w.doc.OriginalPart("word/numbering.xml"); ok && w.hasFile("word/numbering.xml") {
			return nil
		}
	}
	numStyles := w.doc.NumberingStyles()
	if len(numStyles) == 0 {
		return nil
	}

	// Sort style names for deterministic output
	names := make([]string, 0, len(numStyles))
	for name := range numStyles {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:numbering xmlns:w="%s">`, w.wNS()))

	// CT_Numbering requires every abstractNum before every num. Interleaving
	// them is a common reason Word offers to repair the file.
	type numDef struct {
		id   int
		name string
	}
	defs := make([]numDef, 0, len(names))
	for i, name := range names {
		defs = append(defs, numDef{id: i + 1, name: name})
	}
	for _, def := range defs {
		ns := numStyles[def.name]
		abstractID := def.id

		// Abstract numbering definition
		b.WriteString(fmt.Sprintf(`<w:abstractNum w:abstractNumId="%d">`, abstractID))
		b.WriteString(fmt.Sprintf(`<w:nsid w:val="%08X"/>`, abstractID))
		multiLevel := "singleLevel"
		if ns.Type == "multilevel" {
			multiLevel = "multilevel"
		}
		b.WriteString(fmt.Sprintf(`<w:multiLevelType w:val="%s"/>`, multiLevel))

		for lvl, level := range ns.Levels {
			b.WriteString(fmt.Sprintf(`<w:lvl w:ilvl="%d">`, lvl))
			start := level.Start
			if start <= 0 {
				start = 1
			}
			b.WriteString(fmt.Sprintf(`<w:start w:val="%d"/>`, start))

			numFmt := level.Format
			if numFmt == "" {
				numFmt = "decimal"
			}
			b.WriteString(fmt.Sprintf(`<w:numFmt w:val="%s"/>`, numFmt))

			lvlText := level.Text
			if lvlText == "" {
				lvlText = fmt.Sprintf("%%%d.", lvl+1)
			}
			b.WriteString(fmt.Sprintf(`<w:lvlText w:val="%s"/>`, escapeXML(lvlText)))
			b.WriteString(`<w:lvlJc w:val="left"/>`)

			if level.Left > 0 || level.Hanging > 0 {
				b.WriteString("<w:pPr><w:ind")
				if level.Left > 0 {
					b.WriteString(fmt.Sprintf(` w:left="%d"`, level.Left))
				}
				if level.Hanging > 0 {
					b.WriteString(fmt.Sprintf(` w:hanging="%d"`, level.Hanging))
				}
				b.WriteString("/></w:pPr>")
			}

			if level.Font != "" {
				b.WriteString(fmt.Sprintf(`<w:rPr><w:rFonts w:ascii="%s" w:hAnsi="%s"/></w:rPr>`, escapeXML(level.Font), escapeXML(level.Font)))
			}

			b.WriteString(`</w:lvl>`)
		}
		b.WriteString(`</w:abstractNum>`)
	}
	for _, def := range defs {
		b.WriteString(fmt.Sprintf(`<w:num w:numId="%d">`, def.id))
		b.WriteString(fmt.Sprintf(`<w:abstractNumId w:val="%d"/>`, def.id))
		b.WriteString(`</w:num>`)
	}

	b.WriteString(`</w:numbering>`)
	return w.addZipFile("word/numbering.xml", []byte(b.String()))
}
