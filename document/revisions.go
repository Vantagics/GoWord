package document

import "strings"

// AcceptRevisions keeps inserted and moved-in text and drops deletions and
// the previous-property snapshots stored in w:rPrChange, w:pPrChange, and
// w:tblPrChange.
func (d *Document) AcceptRevisions() {
	if d != nil {
		d.applyRevisions(true)
	}
}

// RejectRevisions drops insertions and restores deleted text. Property
// changes revert to the snapshot stored inside the change element.
func (d *Document) RejectRevisions() {
	if d != nil {
		d.applyRevisions(false)
	}
}

func (d *Document) applyRevisions(accept bool) {
	for _, sec := range d.Sections {
		reviseElements(sec.Elements, accept)
		sec.RawSectPr = rewriteMarkup(sec.RawSectPr, accept)
		for _, h := range sec.Headers {
			h.RawPart = []byte(rewriteMarkup(string(h.RawPart), accept))
			reviseElements(h.Elements, accept)
		}
		for _, f := range sec.Footers {
			f.RawPart = []byte(rewriteMarkup(string(f.RawPart), accept))
			reviseElements(f.Elements, accept)
		}
	}
	for _, fn := range d.footnotes {
		reviseElements(fn.Elements, accept)
	}
	for _, en := range d.endnotes {
		reviseElements(en.Elements, accept)
	}
	for _, c := range d.comments {
		reviseElements(c.Elements, accept)
	}
}

func reviseElements(elems []Element, accept bool) {
	for _, el := range elems {
		switch e := el.(type) {
		case *Paragraph:
			e.RawPPr = rewriteMarkup(e.RawPPr, accept)
			if len(e.Inlines) > 0 {
				var next []Element
				for _, in := range e.Inlines {
					next = append(next, reviseInline(in, accept)...)
				}
				e.Inlines = next
				e.Runs = runsFromInlines(next)
			} else {
				for _, r := range e.Runs {
					r.RawRPr = rewriteMarkup(r.RawRPr, accept)
				}
			}
		case *Table:
			e.RawTblPr = rewriteMarkup(e.RawTblPr, accept)
			for _, row := range e.Rows {
				row.RawTrPr = rewriteMarkup(row.RawTrPr, accept)
				row.PrefixXML = rewriteMarkup(row.PrefixXML, accept)
				for _, cell := range row.Cells {
					cell.RawTcPr = rewriteMarkup(cell.RawTcPr, accept)
					reviseElements(cell.Elements, accept)
				}
			}
		case *Sdt:
			reviseElements(e.Elements, accept)
		case *TextRun:
			reviseElements(e.Elements, accept)
		}
	}
}

func reviseInline(el Element, accept bool) []Element {
	switch e := el.(type) {
	case *Run:
		e.RawRPr = rewriteMarkup(e.RawRPr, accept)
		return []Element{e}
	case *RawXML:
		xml := rewriteMarkup(e.XML, accept)
		if strings.TrimSpace(xml) == "" {
			return nil
		}
		e.XML = xml
		e.Text = markupText(xml)
		return []Element{e}
	default:
		return []Element{el}
	}
}

func runsFromInlines(inlines []Element) []*Run {
	var runs []*Run
	for _, in := range inlines {
		switch e := in.(type) {
		case *Run:
			runs = append(runs, e)
		case *RawXML:
			if e.Text != "" {
				runs = append(runs, &Run{Text: e.Text})
			}
		}
	}
	return runs
}

func revisionMarkup(s string) bool {
	return strings.Contains(s, "<w:del") || strings.Contains(s, "<w:ins") ||
		strings.Contains(s, "moveFrom") || strings.Contains(s, "moveTo") ||
		strings.Contains(s, "Change")
}

// rewriteMarkup rewrites one XML fragment. accept keeps insertions and drops
// deletions. reject does the opposite and restores property snapshots.
func rewriteMarkup(s string, accept bool) string {
	if s == "" || !revisionMarkup(s) {
		return s
	}
	if !accept {
		s = restorePropertyChanges(s)
	}
	return rewriteContent(s, accept, false)
}

// restorePropertyChanges replaces each property element that has a direct
// *Change child with the snapshot stored inside that child.
func restorePropertyChanges(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		lt := strings.IndexByte(s[i:], '<')
		if lt < 0 {
			b.WriteString(s[i:])
			break
		}
		b.WriteString(s[i : i+lt])
		i += lt
		if strings.HasPrefix(s[i:], "</") || strings.HasPrefix(s[i:], "<?") || strings.HasPrefix(s[i:], "<!") {
			end := indexTagEnd(s, i)
			if end < 0 {
				b.WriteString(s[i:])
				break
			}
			b.WriteString(s[i:end])
			i = end
			continue
		}
		local, end, self := tagAt(s, i)
		if end < 0 {
			b.WriteString(s[i:])
			break
		}
		if self {
			b.WriteString(s[i:end])
			i = end
			continue
		}
		closeAt, closeEnd := findEnd(s, end, local)
		if closeAt < 0 {
			b.WriteString(s[i:])
			break
		}
		if snap, ok := directChangeSnapshot(s[end:closeAt]); ok {
			b.WriteString(s[i:end])
			b.WriteString(snap)
			b.WriteString(s[closeAt:closeEnd])
		} else {
			b.WriteString(s[i:closeEnd])
		}
		i = closeEnd
	}
	return b.String()
}

func directChangeSnapshot(inner string) (string, bool) {
	i := 0
	for i < len(inner) {
		lt := strings.IndexByte(inner[i:], '<')
		if lt < 0 {
			return "", false
		}
		i += lt
		if strings.HasPrefix(inner[i:], "</") || strings.HasPrefix(inner[i:], "<?") || strings.HasPrefix(inner[i:], "<!") {
			end := indexTagEnd(inner, i)
			if end < 0 {
				return "", false
			}
			i = end
			continue
		}
		local, end, self := tagAt(inner, i)
		if end < 0 {
			return "", false
		}
		if strings.HasSuffix(local, "Change") && !self {
			closeAt, _ := findEnd(inner, end, local)
			if closeAt < 0 {
				return "", false
			}
			body := inner[end:closeAt]
			j := strings.IndexByte(body, '<')
			if j < 0 {
				return "", false
			}
			childLocal, childEnd, childSelf := tagAt(body, j)
			if childEnd < 0 || childSelf {
				return "", false
			}
			bodyAt, _ := findEnd(body, childEnd, childLocal)
			if bodyAt < 0 {
				return "", false
			}
			return body[childEnd:bodyAt], true
		}
		if self {
			i = end
			continue
		}
		_, closeEnd := findEnd(inner, end, local)
		if closeEnd < 0 {
			return "", false
		}
		i = closeEnd
	}
	return "", false
}

func rewriteContent(s string, accept, inDeleted bool) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		lt := strings.IndexByte(s[i:], '<')
		if lt < 0 {
			b.WriteString(s[i:])
			break
		}
		b.WriteString(s[i : i+lt])
		i += lt
		if strings.HasPrefix(s[i:], "</") || strings.HasPrefix(s[i:], "<?") || strings.HasPrefix(s[i:], "<!") {
			end := indexTagEnd(s, i)
			if end < 0 {
				b.WriteString(s[i:])
				break
			}
			b.WriteString(renameTag(s[i:end], inDeleted))
			i = end
			continue
		}
		local, end, self := tagAt(s, i)
		if end < 0 {
			b.WriteString(s[i:])
			break
		}
		switch revisionAction(local, accept) {
		case revDrop:
			if self {
				i = end
				continue
			}
			_, closeEnd := findEnd(s, end, local)
			if closeEnd < 0 {
				return b.String()
			}
			i = closeEnd
		case revUnwrap:
			if self {
				i = end
				continue
			}
			closeAt, closeEnd := findEnd(s, end, local)
			if closeAt < 0 {
				return b.String()
			}
			deleted := inDeleted || local == "del" || local == "moveFrom"
			b.WriteString(rewriteContent(s[end:closeAt], accept, deleted))
			i = closeEnd
		default:
			start := renameTag(s[i:end], inDeleted)
			if self {
				b.WriteString(start)
				i = end
				continue
			}
			closeAt, closeEnd := findEnd(s, end, local)
			if closeAt < 0 {
				b.WriteString(s[i:])
				return b.String()
			}
			inner := s[end:closeAt]
			if !accept {
				if snap, ok := directChangeSnapshot(inner); ok {
					inner = snap
				}
			}
			b.WriteString(start)
			b.WriteString(rewriteContent(inner, accept, inDeleted))
			b.WriteString(renameTag(s[closeAt:closeEnd], inDeleted))
			i = closeEnd
		}
	}
	return b.String()
}

const (
	revKeep = iota
	revDrop
	revUnwrap
)

func revisionAction(local string, accept bool) int {
	switch local {
	case "moveFromRangeStart", "moveFromRangeEnd", "moveToRangeStart", "moveToRangeEnd",
		"customXmlDelRangeStart", "customXmlDelRangeEnd", "customXmlInsRangeStart", "customXmlInsRangeEnd",
		"customXmlMoveFromRangeStart", "customXmlMoveFromRangeEnd", "customXmlMoveToRangeStart", "customXmlMoveToRangeEnd":
		return revDrop
	case "del", "moveFrom":
		if accept {
			return revDrop
		}
		return revUnwrap
	case "ins", "moveTo":
		if accept {
			return revUnwrap
		}
		return revDrop
	}
	if strings.HasSuffix(local, "Change") {
		if accept {
			return revDrop
		}
		// Reject of a property change is handled by rejectedPropertySnapshot
		// on the enclosing fragment. A leftover change element is dropped.
		return revDrop
	}
	return revKeep
}

func renameTag(tag string, inDeleted bool) string {
	if !inDeleted || !strings.Contains(tag, "delText") {
		return tag
	}
	return strings.Replace(tag, "delText", "t", 1)
}

func markupText(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		lt := strings.IndexByte(s[i:], '<')
		if lt < 0 {
			break
		}
		i += lt
		local, end, self := tagAt(s, i)
		if end < 0 {
			break
		}
		if local == "t" && !self && !strings.HasPrefix(s[i:], "</") {
			closeAt, _ := findEnd(s, end, local)
			if closeAt < 0 {
				break
			}
			b.WriteString(s[end:closeAt])
			i = closeAt
			continue
		}
		i = end
	}
	return b.String()
}

func tagAt(s string, i int) (local string, end int, self bool) {
	end = indexTagEnd(s, i)
	if end < 0 || i >= len(s) || s[i] != '<' {
		return "", -1, false
	}
	self = strings.HasSuffix(strings.TrimSpace(s[i:end]), "/>")
	j := i + 1
	if j < len(s) && s[j] == '/' {
		j++
	}
	start := j
	for j < len(s) && isNameChar(s[j]) {
		j++
	}
	local = s[start:j]
	if k := strings.IndexByte(local, ':'); k >= 0 {
		local = local[k+1:]
	}
	return local, end, self
}

func findEnd(s string, from int, local string) (contentEnd, tagEnd int) {
	depth := 1
	i := from
	for i < len(s) && depth > 0 {
		lt := strings.IndexByte(s[i:], '<')
		if lt < 0 {
			return -1, -1
		}
		i += lt
		name, end, self := tagAt(s, i)
		if end < 0 {
			return -1, -1
		}
		closing := strings.HasPrefix(s[i:], "</")
		if name == local {
			if closing {
				depth--
				if depth == 0 {
					return i, end
				}
			} else if !self {
				depth++
			}
		}
		i = end
	}
	return -1, -1
}

func indexTagEnd(s string, i int) int {
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

func isNameChar(c byte) bool {
	return c == ':' || c == '_' || c == '-' || c == '.' ||
		(c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}
