package ooxml

import (
	"bytes"
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

func (r *docxReader) readSettingsPart() error {
	data, err := r.readZipFile("word/settings.xml")
	if err != nil {
		return err
	}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "updateFields" {
			continue
		}
		if onOff(se) {
			r.doc.SetUpdateFieldsOnOpen(true)
		}
		return nil
	}
}

func (r *docxReader) readStyles() error {
	data, err := r.readZipFile("word/styles.xml")
	if err != nil {
		return err
	}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "docDefaults":
			raw, err := r.captureElement(dec, se)
			if err != nil {
				return err
			}
			r.applyDocDefaults(raw)
		case "style":
			raw, err := r.captureElement(dec, se)
			if err != nil {
				return err
			}
			r.applyStyle(se, raw)
		}
	}
}

func (r *docxReader) applyDocDefaults(raw string) {
	dec := xml.NewDecoder(strings.NewReader(raw))
	var fs style.FontStyle
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
		case "rFonts":
			if v, ok := attr(se, "ascii"); ok && v != "" {
				fs.Name = v
			} else if v, ok := attr(se, "hAnsi"); ok {
				fs.Name = v
			}
			if v, ok := attr(se, "eastAsia"); ok {
				fs.NameEastAsia = v
			}
		case "sz":
			if n, ok := parseTwips(attrMust(se, "val")); ok {
				// sz is half-points, not twips. parseTwips without a unit
				// returns the integer unchanged, which is what we want to divide.
				fs.Size = float64(n) / 2
			}
		case "spacing":
			var ps style.ParagraphStyle
			set := false
			if n, ok := parseTwips(attrMust(se, "before")); ok {
				ps.SpaceBefore = n
				set = true
			}
			if n, ok := parseTwips(attrMust(se, "after")); ok {
				ps.SpaceAfter = n
				set = true
			}
			if n, ok := parseTwips(attrMust(se, "line")); ok {
				ps.LineSpacing = n
				set = true
			}
			if v, ok := attr(se, "lineRule"); ok {
				ps.LineRule = v
				set = true
			}
			if set {
				r.doc.SetDefaultParagraph(ps)
			}
		}
	}
	if fs.Name != "" {
		r.doc.SetDefaultFontName(fs.Name)
	}
	if v, ok := eastAsiaDefault(raw); ok && v != "" {
		r.doc.DefaultFont.NameEastAsia = v
	}
	if fs.Size > 0 {
		r.doc.SetDefaultFontSize(fs.Size)
	}
}

func eastAsiaDefault(raw string) (string, bool) {
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", false
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "rFonts" {
			continue
		}
		return attr(se, "eastAsia")
	}
}

func attrMust(se xml.StartElement, local string) string {
	v, _ := attr(se, local)
	return v
}

func (r *docxReader) applyStyle(start xml.StartElement, raw string) {
	id, _ := attr(start, "styleId")
	if id == "" {
		return
	}
	typ, _ := attr(start, "type")
	dec := xml.NewDecoder(strings.NewReader(raw))
	var fs style.FontStyle
	var ps style.ParagraphStyle
	var basedOn string
	seenP, seenR := false, false
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
		case "basedOn":
			if v, ok := attr(se, "val"); ok {
				basedOn = v
			}
		case "pPr":
			pRaw, err := r.captureElement(dec, se)
			if err != nil {
				return
			}
			tmp := &document.Paragraph{}
			r.interpretPPr(tmp, pRaw)
			ps = tmp.Style
			seenP = true
		case "rPr":
			rRaw, err := r.captureElement(dec, se)
			if err != nil {
				return
			}
			fs, _, _ = interpretRPr(rRaw)
			seenR = true
		}
	}
	if basedOn != "" {
		ps.BasedOn = basedOn
	}
	if def, ok := attr(start, "default"); ok && (def == "1" || def == "true") && (typ == "paragraph" || typ == "") {
		r.doc.SetDefaultStyle(id)
	}
	switch typ {
	case "character":
		if seenR {
			r.doc.AddFontStyle(id, fs)
		}
	case "table":
		// Table styles are preserved via the original styles part.
	default:
		if seenP || typ == "paragraph" || typ == "" {
			r.doc.AddParagraphStyle(id, ps)
		}
		if seenR {
			r.doc.SetStyleFont(id, fs)
		}
	}
}

func (r *docxReader) readTheme() {
	data, err := r.readZipFile("word/theme/theme1.xml")
	if err != nil || len(data) == 0 {
		return
	}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	var minorLatin, majorLatin, minorEast, majorEast string
	slot := ""
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
		case "majorFont":
			slot = "major"
		case "minorFont":
			slot = "minor"
		case "latin":
			face, _ := attr(se, "typeface")
			if slot == "major" && majorLatin == "" {
				majorLatin = face
			}
			if slot == "minor" && minorLatin == "" {
				minorLatin = face
			}
		case "ea":
			face, _ := attr(se, "typeface")
			if slot == "major" && majorEast == "" {
				majorEast = face
			}
			if slot == "minor" && minorEast == "" {
				minorEast = face
			}
		}
	}
	r.doc.SetThemeFonts(minorLatin, majorLatin, minorEast, majorEast)
}

func (r *docxReader) readNumberingPart() error {
	data, err := r.readZipFile("word/numbering.xml")
	if err != nil {
		return err
	}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	abstract := map[string]document.NumberingStyle{}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "abstractNum":
			id, _ := attr(se, "abstractNumId")
			raw, err := r.captureElement(dec, se)
			if err != nil {
				return err
			}
			abstract[id] = parseAbstractNum(raw)
		case "num":
			numID, _ := attr(se, "numId")
			raw, err := r.captureElement(dec, se)
			if err != nil {
				return err
			}
			abs := findQuoted(raw, "val")
			// The first w:val inside num is abstractNumId.
			if i := strings.Index(raw, "abstractNumId"); i >= 0 {
				abs = findQuoted(raw[i:], "val")
			}
			ns := applyNumOverrides(abstract[abs], raw)
			if numID != "" {
				r.doc.AddNumberingStyle("num"+numID, ns)
			}
		}
	}
}

func parseAbstractNum(raw string) document.NumberingStyle {
	ns := document.NumberingStyle{Type: "singleLevel"}
	dec := xml.NewDecoder(strings.NewReader(raw))
	var levels []document.NumberingLevel
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
		case "multiLevelType":
			if v, ok := attr(se, "val"); ok && v != "" {
				ns.Type = v
			}
		case "lvl":
			lvlRaw, err := captureElementSerialized(dec, se)
			if err != nil {
				return ns
			}
			levels = append(levels, parseLevel(lvlRaw))
		}
	}
	ns.Levels = levels
	if len(levels) > 1 && ns.Type == "singleLevel" {
		ns.Type = "multilevel"
	}
	return ns
}

// applyNumOverrides copies an abstract numbering definition and applies this
// num's lvlOverride / startOverride entries. Each num keeps its own level slice.
func applyNumOverrides(ns document.NumberingStyle, raw string) document.NumberingStyle {
	levels := reindexLevels(ns.Levels)
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "lvlOverride" {
			continue
		}
		ilvl := atoiAttr(se, "ilvl")
		body, err := captureElementSerialized(dec, se)
		if err != nil {
			break
		}
		repl, startOver, hasLvl := parseLvlOverride(body)
		for len(levels) <= ilvl {
			levels = append(levels, document.NumberingLevel{ILvl: len(levels), Start: 1})
		}
		if hasLvl {
			if repl.Start <= 0 {
				repl.Start = 1
			}
			repl.ILvl = ilvl
			levels[ilvl] = repl
		}
		if startOver > 0 {
			levels[ilvl].Start = startOver
		}
	}
	ns.Levels = levels
	return ns
}

func reindexLevels(levels []document.NumberingLevel) []document.NumberingLevel {
	if len(levels) == 0 {
		return nil
	}
	max := len(levels) - 1
	for _, lv := range levels {
		if lv.ILvl > max {
			max = lv.ILvl
		}
	}
	out := make([]document.NumberingLevel, max+1)
	for i, lv := range levels {
		idx := lv.ILvl
		if idx < 0 || idx > max {
			idx = i
		}
		out[idx] = lv
		if out[idx].Start <= 0 {
			out[idx].Start = 1
		}
	}
	for i := range out {
		if out[i].Start <= 0 {
			out[i].Start = 1
			out[i].ILvl = i
		}
	}
	return out
}

func parseLvlOverride(raw string) (document.NumberingLevel, int, bool) {
	dec := xml.NewDecoder(strings.NewReader(raw))
	var lvl document.NumberingLevel
	var has bool
	var startOver int
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
		case "startOverride":
			startOver = atoiAttr(se, "val")
			if err := skipRest(dec); err != nil {
				return lvl, startOver, has
			}
		case "lvl":
			rawLvl, err := captureElementSerialized(dec, se)
			if err != nil {
				return lvl, startOver, has
			}
			lvl = parseLevel(rawLvl)
			has = true
		}
	}
	return lvl, startOver, has
}

func parseLevel(raw string) document.NumberingLevel {
	var lvl document.NumberingLevel
	dec := xml.NewDecoder(strings.NewReader(raw))
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
		case "lvl":
			if v, ok := attr(se, "ilvl"); ok {
				lvl.ILvl, _ = strconv.Atoi(v)
			}
		case "start":
			if v, ok := attr(se, "val"); ok {
				lvl.Start, _ = strconv.Atoi(v)
			}
		case "numFmt":
			lvl.Format, _ = attr(se, "val")
		case "lvlText":
			lvl.Text, _ = attr(se, "val")
		case "ind":
			if n, ok := parseTwips(attrMust(se, "left")); ok {
				lvl.Left = n
			}
			if n, ok := parseTwips(attrMust(se, "hanging")); ok {
				lvl.Hanging = n
			}
		case "rFonts":
			if v, ok := attr(se, "ascii"); ok {
				lvl.Font = v
			}
		}
	}
	return lvl
}

func (r *docxReader) readNotes(part string, footnotes bool) error {
	data, err := r.readZipFile(part)
	if err != nil {
		return err
	}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local != "footnote" && se.Name.Local != "endnote" {
			continue
		}
		typ, _ := attr(se, "type")
		if typ == "separator" || typ == "continuationSeparator" || typ == "continuationNotice" {
			if err := skipRest(dec); err != nil {
				return err
			}
			continue
		}
		id := atoiAttr(se, "id")
		elems, err := r.readBlockChildren(dec)
		if err != nil {
			return err
		}
		if footnotes {
			fn := &document.Footnote{ID: id, Elements: elems}
			r.doc.RegisterFootnote(fn)
			r.footnotes[id] = fn
			r.fnDisplay[id] = len(r.fnDisplay) + 1
		} else {
			en := &document.Endnote{ID: id, Elements: elems}
			r.doc.RegisterEndnote(en)
			r.endnotes[id] = en
			r.enDisplay[id] = len(r.enDisplay) + 1
		}
	}
}

func (r *docxReader) readCommentsPart() error {
	data, err := r.readZipFile("word/comments.xml")
	if err != nil {
		return err
	}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "comment" {
			continue
		}
		c := &document.Comment{ID: atoiAttr(se, "id")}
		c.Author, _ = attr(se, "author")
		if ds, ok := attr(se, "date"); ok {
			if t, err := time.Parse(time.RFC3339, ds); err == nil {
				c.Date = t
			}
		}
		elems, err := r.readBlockChildren(dec)
		if err != nil {
			return err
		}
		c.Elements = elems
		var text strings.Builder
		for _, el := range elems {
			extractPlain(&text, el)
		}
		c.Text = strings.TrimSpace(text.String())
		r.doc.RegisterComment(c)
		r.comments[c.ID] = c
	}
}

func extractPlain(b *strings.Builder, el document.Element) {
	switch e := el.(type) {
	case *document.Paragraph:
		for _, run := range e.Runs {
			b.WriteString(run.Text)
		}
	case *document.RawXML:
		b.WriteString(e.Text)
	case *document.Sdt:
		for _, child := range e.Elements {
			extractPlain(b, child)
		}
	}
}

func (r *docxReader) loadHeader(target, typ string) *document.Header {
	part := resolvePart(r.docPart, target)
	data, err := r.readZipFile(part)
	if err != nil {
		return &document.Header{Type: typ, PartName: pathBase(target)}
	}
	raw := append([]byte(nil), data...)
	if bytes.Contains(raw, []byte("purl.oclc.org/ooxml")) {
		raw = nil
	}
	h := &document.Header{Type: typ, PartName: pathBase(part), RawPart: raw}
	r.withPart(part, func() {
		h.Elements = r.parseBlocks(data)
	})
	return h
}

func (r *docxReader) loadFooter(target, typ string) *document.Footer {
	part := resolvePart(r.docPart, target)
	data, err := r.readZipFile(part)
	if err != nil {
		return &document.Footer{Type: typ, PartName: pathBase(target)}
	}
	raw := append([]byte(nil), data...)
	if bytes.Contains(raw, []byte("purl.oclc.org/ooxml")) {
		raw = nil
	}
	f := &document.Footer{Type: typ, PartName: pathBase(part), RawPart: raw}
	r.withPart(part, func() {
		f.Elements = r.parseBlocks(data)
	})
	return f
}

func (r *docxReader) withPart(part string, fn func()) {
	prevPart, prevRels, prevMode := r.curPart, r.rels, r.relMode
	r.usePart(part)
	fn()
	r.curPart, r.rels, r.relMode = prevPart, prevRels, prevMode
}

func pathBase(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func (r *docxReader) parseBlocks(data []byte) []document.Element {
	dec := r.openXML(string(data))
	defer r.closeXML(dec)
	// Advance to the root element's children.
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil
		}
		if _, ok := tok.(xml.StartElement); ok {
			elems, err := r.readBlockChildren(dec)
			if err != nil {
				return elems
			}
			return elems
		}
	}
}

func (r *docxReader) applySectPr(sec *document.Section, raw string) {
	if strings.Contains(raw, "pt\"") || strings.Contains(raw, "in\"") || strings.Contains(raw, "cm\"") {
		sec.RawSectPr = ""
	}
	// Each sectPr stands on its own. A missing w:cols is one column, not a
	// copy of the previous section's count.
	sec.Style.ColumnCount = 1
	parent := ""
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.EndElement:
			if t.Name.Local == parent {
				parent = ""
			}
		case xml.StartElement:
			se := t
			switch se.Name.Local {
			case "pgSz":
				if n, ok := parseTwips(attrMust(se, "w")); ok {
					sec.Style.PageWidth = n
				}
				if n, ok := parseTwips(attrMust(se, "h")); ok {
					sec.Style.PageHeight = n
				}
				if v, ok := attr(se, "orient"); ok && v != "" {
					sec.Style.Orientation = v
				}
			case "pgMar":
				setMar := func(local string, dst *int) {
					if n, ok := parseTwips(attrMust(se, local)); ok {
						*dst = n
					}
				}
				setMar("top", &sec.Style.MarginTop)
				setMar("bottom", &sec.Style.MarginBottom)
				setMar("left", &sec.Style.MarginLeft)
				setMar("right", &sec.Style.MarginRight)
				setMar("header", &sec.Style.HeaderHeight)
				setMar("footer", &sec.Style.FooterHeight)
				setMar("gutter", &sec.Style.Gutter)
			case "cols":
				// A cols element with no num is a single column. Copying the
				// previous section's count would keep a 2-column section open.
				sec.Style.ColumnCount = 1
				if n, err := strconv.Atoi(attrMust(se, "num")); err == nil && n > 0 {
					sec.Style.ColumnCount = n
				}
				if n, ok := parseTwips(attrMust(se, "space")); ok {
					sec.Style.ColumnSpacing = n
				}
			case "type":
				if v, ok := attr(se, "val"); ok {
					sec.Style.BreakType = v
				}
			case "pgNumType":
				if v, ok := attr(se, "fmt"); ok {
					sec.Style.PageNumberFormat = v
				}
				if v, ok := attr(se, "start"); ok {
					if n, err := strconv.Atoi(v); err == nil {
						sec.Style.PageNumberStart = &n
					}
				}
			case "docGrid":
				if n, ok := parseTwips(attrMust(se, "linePitch")); ok {
					sec.Style.LinePitch = n
				}
			case "footnotePr", "endnotePr":
				parent = se.Name.Local
			case "numFmt":
				if v, ok := attr(se, "val"); ok && v != "" {
					switch parent {
					case "footnotePr":
						r.fnNumFmt = v
					case "endnotePr":
						r.enNumFmt = v
					}
				}
			case "titlePg":
				sec.Style.TitlePg = onOff(se)
			case "headerReference":
				id, _ := attr(se, "id")
				typ, _ := attr(se, "type")
				if typ == "" {
					typ = "default"
				}
				target := r.rels[id]
				if target != "" {
					sec.Headers = append(sec.Headers, r.loadHeader(target, typ))
				}
			case "footerReference":
				id, _ := attr(se, "id")
				typ, _ := attr(se, "type")
				if typ == "" {
					typ = "default"
				}
				target := r.rels[id]
				if target != "" {
					sec.Footers = append(sec.Footers, r.loadFooter(target, typ))
				}
			}
		}
	}
}
