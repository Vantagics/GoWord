package document

import (
	"encoding/xml"
	"strings"
)

// OMMLToLatex converts an Office Math (OMML) fragment to a LaTeX expression.
// It covers fractions, radicals, superscripts, and subscripts. Unknown
// elements contribute the text of their children.
func OMMLToLatex(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			return ""
		}
		if se, ok := tok.(xml.StartElement); ok {
			return readMath(dec, se)
		}
	}
}

func readMath(dec *xml.Decoder, start xml.StartElement) string {
	switch start.Name.Local {
	case "t":
		return escapeLatex(readMathText(dec))
	case "f":
		ch := readMathChildren(dec)
		return `\frac{` + ch["num"] + `}{` + ch["den"] + `}`
	case "sSup":
		ch := readMathChildren(dec)
		return `{` + ch["e"] + `}^{` + ch["sup"] + `}`
	case "sSub":
		ch := readMathChildren(dec)
		return `{` + ch["e"] + `}_{` + ch["sub"] + `}`
	case "sSubSup":
		ch := readMathChildren(dec)
		return `{` + ch["e"] + `}_{` + ch["sub"] + `}^{` + ch["sup"] + `}`
	case "rad":
		ch := readMathChildren(dec)
		if strings.TrimSpace(ch["deg"]) == "" {
			return `\sqrt{` + ch["e"] + `}`
		}
		return `\sqrt[` + ch["deg"] + `]{` + ch["e"] + `}`
	default:
		var b strings.Builder
		for {
			tok, err := dec.Token()
			if err != nil {
				return b.String()
			}
			switch t := tok.(type) {
			case xml.StartElement:
				b.WriteString(readMath(dec, t))
			case xml.EndElement:
				return b.String()
			}
		}
	}
}

func readMathChildren(dec *xml.Decoder) map[string]string {
	out := map[string]string{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return out
		}
		switch t := tok.(type) {
		case xml.StartElement:
			out[t.Name.Local] += readMath(dec, t)
		case xml.EndElement:
			return out
		}
	}
}

func readMathText(dec *xml.Decoder) string {
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
			// Unexpected markup inside m:t. Skip it.
			skipMath(dec)
		}
	}
}

func skipMath(dec *xml.Decoder) {
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
}

func escapeLatex(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '{', '}', '$', '&', '%', '_', '#':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '^':
			b.WriteString(`\^{}`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
