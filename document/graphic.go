package document

import (
	"encoding/xml"
	"sort"
	"strings"
	"unicode"
)

// GraphicText returns chart and SmartArt labels stored in chart and diagram
// parts. Paragraph text does not include them. The same label is returned once
// even when both the diagram data and the drawing copy it.
func (d *Document) GraphicText() string {
	if d == nil {
		return ""
	}
	names := d.OriginalPartNames()
	sort.Strings(names)
	var parts []string
	seen := map[string]bool{}
	for _, name := range names {
		low := strings.ToLower(name)
		if !strings.HasSuffix(low, ".xml") {
			continue
		}
		if !strings.Contains(low, "/charts/") && !strings.Contains(low, "/diagrams/") {
			continue
		}
		data, ok := d.OriginalPart(name)
		if !ok {
			continue
		}
		for _, s := range graphicLabels(data) {
			if s == "" || seen[s] {
				continue
			}
			seen[s] = true
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n") + "\n"
}

func graphicLabels(data []byte) []string {
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	var out []string
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
		case "t", "v":
			s := readGraphicText(dec)
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if se.Name.Local == "v" && !hasLetter(s) {
				continue
			}
			out = append(out, s)
		}
	}
	return out
}

func readGraphicText(dec *xml.Decoder) string {
	var b strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return b.String()
		}
		switch t := tok.(type) {
		case xml.CharData:
			b.Write(t)
		case xml.EndElement:
			return b.String()
		case xml.StartElement:
			b.WriteString(readGraphicText(dec))
		}
	}
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
