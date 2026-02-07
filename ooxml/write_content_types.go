package ooxml

import (
	"fmt"
	"strings"
)

func (w *Writer) writeContentTypes() error {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<Types xmlns="%s">`, nsCT))

	// Default types
	b.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	b.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)

	// Image defaults
	imageExts := w.collectImageExtensions()
	for _, ext := range imageExts {
		b.WriteString(fmt.Sprintf(`<Default Extension="%s" ContentType="%s"/>`, ext, imageContentType(ext)))
	}

	// Override parts
	b.WriteString(`<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>`)
	b.WriteString(`<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>`)
	b.WriteString(`<Override PartName="/word/settings.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"/>`)
	b.WriteString(`<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`)
	b.WriteString(`<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`)

	if len(w.doc.NumberingStyles()) > 0 {
		b.WriteString(`<Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/>`)
	}
	if len(w.doc.Footnotes()) > 0 {
		b.WriteString(`<Override PartName="/word/footnotes.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footnotes+xml"/>`)
	}
	if len(w.doc.Endnotes()) > 0 {
		b.WriteString(`<Override PartName="/word/endnotes.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.endnotes+xml"/>`)
	}
	if len(w.doc.Comments()) > 0 {
		b.WriteString(`<Override PartName="/word/comments.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.comments+xml"/>`)
	}

	// Header/footer overrides
	for i := range w.headerParts {
		b.WriteString(fmt.Sprintf(`<Override PartName="/word/%s" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>`, w.headerParts[i].filename))
	}
	for i := range w.footerParts {
		b.WriteString(fmt.Sprintf(`<Override PartName="/word/%s" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>`, w.footerParts[i].filename))
	}

	b.WriteString(`</Types>`)
	return w.addZipFile("[Content_Types].xml", []byte(b.String()))
}

func (w *Writer) collectImageExtensions() []string {
	seen := map[string]bool{}
	var exts []string
	for _, img := range w.doc.Images() {
		ext := ""
		if img.MimeType != "" {
			ext = imageExtension(img.MimeType)
		} else if img.Source != "" {
			ext = strings.TrimPrefix(strings.ToLower(getFileExt(img.Source)), ".")
			if ext == "jpg" {
				ext = "jpeg"
			}
		}
		if ext != "" && !seen[ext] {
			seen[ext] = true
			exts = append(exts, ext)
		}
	}
	return exts
}

func getFileExt(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}
