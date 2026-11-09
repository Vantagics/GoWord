package ooxml

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type ctDoc struct {
	XMLName   xml.Name `xml:"Types"`
	Defaults  []ctItem `xml:"Default"`
	Overrides []ctItem `xml:"Override"`
}

type ctItem struct {
	Ext         string `xml:"Extension,attr"`
	PartName    string `xml:"PartName,attr"`
	ContentType string `xml:"ContentType,attr"`
}

func (w *Writer) writeContentTypes() error {
	defaults := map[string]string{
		"rels": "application/vnd.openxmlformats-package.relationships+xml",
		"xml":  "application/xml",
	}
	overrides := map[string]string{}

	if data, ok := w.doc.OriginalPart("[Content_Types].xml"); ok {
		var ct ctDoc
		if err := xml.Unmarshal(data, &ct); err == nil {
			for _, d := range ct.Defaults {
				if d.Ext != "" && d.ContentType != "" {
					defaults[d.Ext] = d.ContentType
				}
			}
			for _, o := range ct.Overrides {
				name := strings.TrimPrefix(o.PartName, "/")
				if name != "" && w.hasFile(name) && o.ContentType != "" {
					overrides["/"+name] = o.ContentType
				}
			}
		}
	}

	for _, ext := range w.collectImageExtensions() {
		if _, ok := defaults[ext]; !ok {
			defaults[ext] = imageContentType(ext)
		}
	}
	for _, f := range w.files {
		if ct, ok := standardContentType(f.name); ok {
			overrides["/"+f.name] = ct
		}
		if strings.HasPrefix(f.name, "word/media/") {
			ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(f.name)), ".")
			if ext != "" {
				if _, ok := defaults[ext]; !ok {
					defaults[ext] = imageContentType(ext)
				}
			}
		}
	}

	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<Types xmlns="%s">`, nsCT))
	exts := make([]string, 0, len(defaults))
	for ext := range defaults {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	for _, ext := range exts {
		b.WriteString(fmt.Sprintf(`<Default Extension="%s" ContentType="%s"/>`, ext, defaults[ext]))
	}
	parts := make([]string, 0, len(overrides))
	for part := range overrides {
		parts = append(parts, part)
	}
	sort.Strings(parts)
	for _, part := range parts {
		b.WriteString(fmt.Sprintf(`<Override PartName="%s" ContentType="%s"/>`, part, overrides[part]))
	}
	b.WriteString(`</Types>`)
	return w.addZipFile("[Content_Types].xml", []byte(b.String()))
}

func standardContentType(name string) (string, bool) {
	switch name {
	case "word/document.xml":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml", true
	case "word/styles.xml":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml", true
	case "word/settings.xml":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml", true
	case "word/fontTable.xml":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml", true
	case "word/numbering.xml":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml", true
	case "word/footnotes.xml":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.footnotes+xml", true
	case "word/endnotes.xml":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.endnotes+xml", true
	case "word/comments.xml":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.comments+xml", true
	case "word/webSettings.xml":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.webSettings+xml", true
	case "word/theme/theme1.xml":
		return "application/vnd.openxmlformats-officedocument.theme+xml", true
	case "docProps/core.xml":
		return "application/vnd.openxmlformats-package.core-properties+xml", true
	case "docProps/app.xml":
		return "application/vnd.openxmlformats-officedocument.extended-properties+xml", true
	case "docProps/custom.xml":
		return "application/vnd.openxmlformats-officedocument.custom-properties+xml", true
	}
	base := filepath.Base(name)
	if strings.HasPrefix(name, "word/") && strings.HasSuffix(name, ".xml") && strings.HasPrefix(base, "header") {
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml", true
	}
	if strings.HasPrefix(name, "word/") && strings.HasSuffix(name, ".xml") && strings.HasPrefix(base, "footer") {
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml", true
	}
	return "", false
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
