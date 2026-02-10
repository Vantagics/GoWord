package ooxml

import (
	"fmt"
	"strings"
)

func (w *Writer) writeSettings() error {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:settings xmlns:w="%s" xmlns:r="%s">`, nsW, nsR))
	b.WriteString(`<w:defaultTabStop w:val="720"/>`)
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
