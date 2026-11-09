package ooxml

import (
	"fmt"
	"strings"
)

func (w *Writer) writeSettings() error {
	if !w.doc.SettingsDirty() {
		if _, ok := w.doc.OriginalPart("word/settings.xml"); ok && w.hasFile("word/settings.xml") {
			return nil
		}
	}
	var b strings.Builder
	b.WriteString(xmlHeader())
	// CT_Settings order: zoom, defaultTabStop, evenAndOddHeaders,
	// characterSpacingControl, updateFields, compat.
	b.WriteString(fmt.Sprintf(`<w:settings xmlns:mc="%s" xmlns:o="%s" xmlns:r="%s" xmlns:m="%s" xmlns:v="%s" xmlns:w10="%s" xmlns:w="%s" xmlns:w14="%s" xmlns:sl="http://schemas.openxmlformats.org/schemaLibrary/2006/main" mc:Ignorable="w14">`,
		nsMC, nsO, nsR, nsM, nsV, nsW10, nsW, nsW14))
	b.WriteString(`<w:zoom w:percent="100"/>`)
	b.WriteString(`<w:defaultTabStop w:val="720"/>`)
	if w.hasEvenHeaderFooter() {
		b.WriteString(`<w:evenAndOddHeaders/>`)
	}
	b.WriteString(`<w:characterSpacingControl w:val="doNotCompress"/>`)
	if w.doc.UpdateFieldsOnOpen() {
		b.WriteString(`<w:updateFields w:val="true"/>`)
	}
	b.WriteString(`<w:compat>`)
	b.WriteString(`<w:compatSetting w:name="compatibilityMode" w:uri="http://schemas.microsoft.com/office/word" w:val="15"/>`)
	b.WriteString(`</w:compat>`)
	b.WriteString(`</w:settings>`)
	return w.addZipFile("word/settings.xml", []byte(b.String()))
}

func (w *Writer) hasEvenHeaderFooter() bool {
	for _, sec := range w.doc.Sections {
		for _, h := range sec.Headers {
			if h.Type == "even" {
				return true
			}
		}
		for _, f := range sec.Footers {
			if f.Type == "even" {
				return true
			}
		}
	}
	return false
}
