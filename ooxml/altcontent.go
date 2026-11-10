package ooxml

import (
	"bytes"
	"strings"
)

// toStrictNamespaces rewrites the transitional OOXML URIs that have a strict
// counterpart. Content types are left alone; strict packages use the same
// part content types as transitional ones.
func toStrictNamespaces(data []byte) []byte {
	reps := []struct{ from, to string }{
		{"http://schemas.openxmlformats.org/wordprocessingml/2006/main", "http://purl.oclc.org/ooxml/wordprocessingml/main"},
		{"http://schemas.openxmlformats.org/officeDocument/2006/relationships", "http://purl.oclc.org/ooxml/officeDocument/relationships"},
		{"http://schemas.openxmlformats.org/officeDocument/2006/math", "http://purl.oclc.org/ooxml/officeDocument/math"},
		{"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing", "http://purl.oclc.org/ooxml/drawingml/wordprocessingDrawing"},
	}
	for _, r := range reps {
		data = bytes.ReplaceAll(data, []byte(r.from), []byte(r.to))
	}
	return data
}

// stripAlternateContent removes mc:AlternateContent elements, including their
// Choice and Fallback children, and leaves the surrounding markup in place.
func stripAlternateContent(data []byte) []byte {
	if !bytes.Contains(data, []byte("AlternateContent")) {
		return data
	}
	s := string(data)
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		lt := strings.IndexByte(s[i:], '<')
		if lt < 0 {
			b.WriteString(s[i:])
			break
		}
		b.WriteString(s[i : i+lt])
		i += lt
		if altOpen(s, i) {
			end := skipAltElement(s, i)
			if end < 0 {
				b.WriteString(s[i:])
				break
			}
			// Leave an empty run of text so a paragraph whose only content was
			// the drawing still round-trips as a blank line.
			b.WriteString(`<w:t></w:t>`)
			i = end
			continue
		}
		end := scanTagEnd(s, i)
		if end < 0 {
			b.WriteString(s[i:])
			break
		}
		b.WriteString(s[i:end])
		i = end
	}
	return []byte(b.String())
}

func altOpen(s string, i int) bool {
	if i >= len(s) || s[i] != '<' {
		return false
	}
	if strings.HasPrefix(s[i:], "</") || strings.HasPrefix(s[i:], "<?") || strings.HasPrefix(s[i:], "<!") {
		return false
	}
	name, _, self := tagInfoAt(s, i)
	return name == "AlternateContent" && !self
}

func skipAltElement(s string, i int) int {
	depth := 0
	for i < len(s) {
		if strings.HasPrefix(s[i:], "<!--") {
			end := strings.Index(s[i:], "-->")
			if end < 0 {
				return -1
			}
			i += end + 3
			continue
		}
		end := scanTagEnd(s, i)
		if end < 0 {
			return -1
		}
		name, closing, self := tagInfoAt(s, i)
		if name == "AlternateContent" {
			if closing {
				depth--
				if depth == 0 {
					return end
				}
			} else if !self {
				depth++
			}
		}
		i = end
		lt := strings.IndexByte(s[i:], '<')
		if lt < 0 {
			return -1
		}
		i += lt
	}
	return -1
}

func tagInfoAt(s string, i int) (local string, closing, self bool) {
	end := scanTagEnd(s, i)
	if end < 0 {
		end = len(s)
	}
	body := s[i:end]
	if strings.HasPrefix(body, "</") {
		closing = true
		body = body[2:]
	} else if strings.HasPrefix(body, "<") {
		body = body[1:]
	}
	body = strings.TrimSpace(body)
	self = strings.HasSuffix(strings.TrimSpace(s[i:end]), "/>")
	j := 0
	for j < len(body) && isXMLName(body[j]) {
		j++
	}
	local = body[:j]
	if k := strings.IndexByte(local, ':'); k >= 0 {
		local = local[k+1:]
	}
	return local, closing, self
}

func scanTagEnd(s string, i int) int {
	quote := byte(0)
	for j := i; j < len(s); j++ {
		c := s[j]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			quote = c
		case '>':
			return j + 1
		}
	}
	return -1
}

func isXMLName(c byte) bool {
	return c == ':' || c == '_' || c == '-' || c == '.' ||
		(c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}
