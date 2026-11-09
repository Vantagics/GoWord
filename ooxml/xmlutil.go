package ooxml

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
)

func attr(start xml.StartElement, local string) (string, bool) {
	for _, a := range start.Attr {
		if a.Name.Local == local {
			return a.Value, true
		}
	}
	return "", false
}

func onOff(start xml.StartElement) bool {
	v, ok := attr(start, "val")
	if !ok {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

// parseTwips parses an OOXML measurement. Transitional files use twips as a
// bare integer. Strict files often use a unit suffix such as "595.30pt".
func parseTwips(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	mult := 1.0
	switch {
	case strings.HasSuffix(s, "pt"):
		s, mult = strings.TrimSuffix(s, "pt"), 20
	case strings.HasSuffix(s, "in"):
		s, mult = strings.TrimSuffix(s, "in"), 1440
	case strings.HasSuffix(s, "cm"):
		s, mult = strings.TrimSuffix(s, "cm"), 567
	case strings.HasSuffix(s, "mm"):
		s, mult = strings.TrimSuffix(s, "mm"), 56.7
	case strings.HasSuffix(s, "pc"), strings.HasSuffix(s, "pi"):
		s, mult = s[:len(s)-2], 240
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, false
	}
	if f >= 0 {
		return int(f*mult + 0.5), true
	}
	return int(f*mult - 0.5), true
}

func hasUnitSuffix(s string) bool {
	s = strings.TrimSpace(s)
	for _, suf := range []string{"pt", "in", "cm", "mm", "pc", "pi"} {
		if strings.HasSuffix(s, suf) {
			return true
		}
	}
	return false
}

func resolvePart(basePart, target string) string {
	if target == "" {
		return ""
	}
	low := strings.ToLower(target)
	if strings.HasPrefix(low, "http://") || strings.HasPrefix(low, "https://") || strings.HasPrefix(low, "mailto:") {
		return target
	}
	if unesc, err := url.PathUnescape(target); err == nil {
		target = unesc
	}
	if strings.HasPrefix(target, "/") {
		return strings.TrimPrefix(path.Clean(target), "/")
	}
	dir := path.Dir(basePart)
	if dir == "." {
		dir = ""
	}
	return path.Clean(path.Join(dir, target))
}

func (r *docxReader) resolveMainDocument() string {
	rels := r.relsByPart[""]
	for _, rel := range rels {
		if rel.Type == relDoc || strings.HasSuffix(rel.Type, "/officeDocument") {
			return resolvePart("", rel.Target)
		}
	}
	if _, err := r.readZipFile("word/document.xml"); err == nil {
		return "word/document.xml"
	}
	return "word/document.xml"
}

func skipRest(dec *xml.Decoder) error {
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}

func qName(n xml.Name) string {
	if n.Space == "" {
		return n.Local
	}
	if n.Space == "xmlns" || n.Space == "http://www.w3.org/2000/xmlns/" {
		if n.Local == "xmlns" || n.Local == "" {
			return "xmlns"
		}
		return "xmlns:" + n.Local
	}
	if p, ok := nsPrefixes[n.Space]; ok {
		return p + ":" + n.Local
	}
	return n.Local
}

var nsPrefixes = map[string]string{
	nsW:                                    "w",
	nsR:                                    "r",
	nsWP:                                   "wp",
	nsA:                                    "a",
	nsPIC:                                  "pic",
	nsMC:                                   "mc",
	nsM:                                    "m",
	nsO:                                    "o",
	nsV:                                    "v",
	nsW10:                                  "w10",
	nsW14:                                  "w14",
	nsW15:                                  "w15",
	nsWP14:                                 "wp14",
	nsWPC:                                  "wpc",
	nsWPG:                                  "wpg",
	nsWPI:                                  "wpi",
	nsWNE:                                  "wne",
	nsWPS:                                  "wps",
	"http://www.w3.org/XML/1998/namespace": "xml",
	"http://purl.oclc.org/ooxml/wordprocessingml/main":        "w",
	"http://purl.oclc.org/ooxml/officeDocument/relationships": "r",
}

func writeStart(b *strings.Builder, start xml.StartElement) {
	b.WriteByte('<')
	b.WriteString(qName(start.Name))
	for _, a := range start.Attr {
		b.WriteByte(' ')
		b.WriteString(qName(a.Name))
		b.WriteString(`="`)
		xml.EscapeText(b, []byte(a.Value))
		b.WriteByte('"')
	}
	b.WriteByte('>')
}

func (r *docxReader) openXML(src string) *xml.Decoder {
	dec := xml.NewDecoder(strings.NewReader(src))
	r.caps = append(r.caps, xmlSrc{dec: dec, src: src})
	return dec
}

func (r *docxReader) closeXML(dec *xml.Decoder) {
	for i := len(r.caps) - 1; i >= 0; i-- {
		if r.caps[i].dec == dec {
			r.caps = append(r.caps[:i], r.caps[i+1:]...)
			return
		}
	}
}

func (r *docxReader) srcOf(dec *xml.Decoder) (string, bool) {
	for i := len(r.caps) - 1; i >= 0; i-- {
		if r.caps[i].dec == dec {
			return r.caps[i].src, true
		}
	}
	return "", false
}

// captureElement returns the original markup for start, which the caller has
// already pulled from dec. Prefer a byte slice of the source XML so prefixes
// and mc:AlternateContent survive; fall back to a re-serialization otherwise.
func (r *docxReader) captureElement(dec *xml.Decoder, start xml.StartElement) (string, error) {
	if src, ok := r.srcOf(dec); ok {
		endOfStart := int(dec.InputOffset())
		if endOfStart > len(src) {
			endOfStart = len(src)
		}
		startAt := strings.LastIndex(src[:endOfStart], "<")
		if startAt < 0 {
			startAt = 0
		}
		depth := 1
		for depth > 0 {
			tok, err := dec.Token()
			if err != nil {
				return "", err
			}
			switch tok.(type) {
			case xml.StartElement:
				depth++
			case xml.EndElement:
				depth--
			}
		}
		end := int(dec.InputOffset())
		if end > len(src) {
			end = len(src)
		}
		if startAt > end {
			startAt = 0
		}
		return src[startAt:end], nil
	}
	return captureElementSerialized(dec, start)
}

func captureElementSerialized(dec *xml.Decoder, start xml.StartElement) (string, error) {
	var b strings.Builder
	if err := writeCaptured(&b, dec, start); err != nil {
		return "", err
	}
	return b.String(), nil
}

func writeCaptured(b *strings.Builder, dec *xml.Decoder, start xml.StartElement) error {
	writeStart(b, start)
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if err := writeCaptured(b, dec, t); err != nil {
				return err
			}
		case xml.EndElement:
			b.WriteString("</")
			b.WriteString(qName(t.Name))
			b.WriteByte('>')
			return nil
		case xml.CharData:
			xml.EscapeText(b, t)
		case xml.Comment:
			b.WriteString("<!--")
			b.Write(t)
			b.WriteString("-->")
		}
	}
}

// visibleText returns text from w:t elements, preferring mc:Choice over
// mc:Fallback and skipping deleted text.
func visibleText(raw string) string {
	dec := xml.NewDecoder(strings.NewReader(raw))
	var b strings.Builder
	skip := 0
	choicePending := false
	choiceText := false
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := t.Name.Local
			switch local {
			case "AlternateContent":
				choicePending = true
				choiceText = false
			case "del", "moveFrom":
				skip++
			case "Fallback":
				if choicePending && choiceText {
					skip++
					if err := skipRest(dec); err != nil {
						return b.String()
					}
					if skip > 0 {
						skip--
					}
					continue
				}
			case "den":
				if skip == 0 {
					b.WriteByte('/')
				}
			case "sup":
				if skip == 0 {
					b.WriteByte('^')
				}
			case "sub":
				if skip == 0 {
					b.WriteByte('_')
				}
			case "t":
				s, err := readCharData(dec)
				if err != nil {
					return b.String()
				}
				if skip == 0 {
					b.WriteString(s)
					if choicePending {
						choiceText = choiceText || s != ""
					}
				}
			case "delText":
				_, _ = readCharData(dec)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "del", "moveFrom":
				if skip > 0 {
					skip--
				}
			case "AlternateContent":
				choicePending = false
			}
		}
	}
	return b.String()
}

// deletedText collects w:delText, which is the text Word stores inside w:del.
func deletedText(raw string) string {
	dec := xml.NewDecoder(strings.NewReader(raw))
	var b strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "delText" {
			continue
		}
		s, err := readCharData(dec)
		if err != nil {
			break
		}
		b.WriteString(s)
	}
	return b.String()
}

func atoiAttr(start xml.StartElement, local string) int {
	v, ok := attr(start, local)
	if !ok {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

func findQuoted(raw, key string) string {
	// key is a local attribute name such as embed or id.
	patterns := []string{":" + key + `="`, key + `="`}
	for _, p := range patterns {
		i := strings.Index(raw, p)
		if i < 0 {
			continue
		}
		rest := raw[i+len(p):]
		j := strings.IndexByte(rest, '"')
		if j >= 0 {
			return rest[:j]
		}
	}
	return ""
}

func xmlError(part string, err error) error {
	return fmt.Errorf("%s: %w", part, err)
}
