package ooxml

import (
	"encoding/xml"
	"strings"

	"github.com/VantageDataChat/GoWord/document"
)

func (r *docxReader) readGlossary() error {
	data, err := r.readZipFile("word/glossary/document.xml")
	if err != nil || len(data) == 0 {
		return nil
	}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	var name, text string
	var inPart bool
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "docPart":
				inPart = true
				name, text = "", ""
			case "name":
				if inPart && name == "" {
					name, _ = attr(t, "val")
				}
			case "t":
				s, err := readCharData(dec)
				if err != nil {
					return err
				}
				if inPart {
					text += s
				}
			}
		case xml.EndElement:
			if t.Name.Local == "docPart" && inPart {
				r.doc.AddBuildingBlock(name, text)
				inPart = false
			}
		}
	}
	return nil
}

func (r *docxReader) readEmbeddedFonts() error {
	data, err := r.readZipFile("word/fontTable.xml")
	if err != nil || len(data) == 0 {
		return nil
	}
	rels := r.relsByPart["word/fontTable.xml"]
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	var fontName string
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "font":
			fontName, _ = attr(se, "name")
		case "embedRegular", "embedBold", "embedItalic", "embedBoldItalic":
			id, _ := attr(se, "id")
			key, _ := attr(se, "fontKey")
			rel, ok := rels[id]
			if !ok || key == "" {
				continue
			}
			target := rel.Target
			if !strings.HasPrefix(target, "word/") && !strings.HasPrefix(target, "/") {
				target = "word/" + strings.TrimPrefix(target, "/")
			}
			target = strings.TrimPrefix(target, "/")
			raw, err := r.readZipFile(target)
			if err != nil {
				continue
			}
			clear, err := document.ObfuscateFont(key, raw)
			if err != nil {
				continue
			}
			name := fontName
			if name == "" {
				name = target
			}
			r.doc.AddEmbeddedFont(document.EmbeddedFont{Name: name, Data: clear, Key: key})
		}
	}
	return nil
}
