package ooxml

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"

	"github.com/VantageDataChat/GoWord/common"
	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

func (r *docxReader) readDocument() error {
	data, err := r.readZipFile(r.docPart)
	if err != nil {
		data, err = r.readZipFile("word/document.xml")
		if err != nil {
			return err
		}
		r.docPart = "word/document.xml"
		r.doc.SetMainPart(r.docPart)
		r.usePart(r.docPart)
	}
	dec := r.openXML(string(data))
	defer r.closeXML(dec)
	var sec *document.Section
	inBody := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if !inBody {
				if t.Name.Local == "body" {
					inBody = true
					sec = r.doc.AddSection()
				}
				continue
			}
			el, sect, err := r.readOneBlock(dec, t)
			if err != nil {
				return err
			}
			if el != nil && sec != nil {
				sec.Elements = append(sec.Elements, el)
			}
			if sect != "" && sec != nil {
				sec.RawSectPr = sect
				r.applySectPr(sec, sect)
				if t.Name.Local != "sectPr" {
					next := r.doc.AddSectionWithStyle(sec.Style)
					sec = next
				}
			}
		case xml.EndElement:
			if t.Name.Local == "body" {
				inBody = false
			}
		}
	}
	if len(r.doc.Sections) == 0 {
		r.doc.AddSection()
	}
	return nil
}

// readBlockChildren reads sibling blocks until the parent element ends.
func (r *docxReader) readBlockChildren(dec *xml.Decoder) ([]document.Element, error) {
	var elems []document.Element
	for {
		tok, err := dec.Token()
		if err != nil {
			return elems, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			el, _, err := r.readOneBlock(dec, t)
			if err != nil {
				return elems, err
			}
			if el != nil {
				elems = append(elems, el)
			}
		case xml.EndElement:
			return elems, nil
		}
	}
}

func (r *docxReader) readOneBlock(dec *xml.Decoder, start xml.StartElement) (document.Element, string, error) {
	switch start.Name.Local {
	case "p":
		p, sect, err := r.readParagraph(dec)
		if err != nil {
			return nil, sect, err
		}
		if !paragraphHasContent(p) {
			return nil, sect, nil
		}
		return p, sect, nil
	case "tbl":
		tbl, err := r.readTable(dec)
		return tbl, "", err
	case "sdt":
		sdt, err := r.readSdt(dec)
		return sdt, "", err
	case "sectPr":
		raw, err := r.captureElement(dec, start)
		return nil, raw, err
	case "bookmarkStart":
		bm := &document.Bookmark{ID: atoiAttr(start, "id"), Span: true}
		bm.Name, _ = attr(start, "name")
		return bm, "", skipRest(dec)
	case "bookmarkEnd":
		bm := &document.BookmarkEnd{ID: atoiAttr(start, "id")}
		return bm, "", skipRest(dec)
	case "customXml", "smartTag":
		raw, err := r.captureElement(dec, start)
		if err != nil {
			return nil, "", err
		}
		return &document.RawXML{XML: raw, Text: visibleText(raw)}, "", nil
	default:
		raw, err := r.captureElement(dec, start)
		if err != nil {
			return nil, "", err
		}
		if strings.TrimSpace(raw) == "" {
			return nil, "", nil
		}
		return &document.RawXML{XML: raw, Text: visibleText(raw)}, "", nil
	}
}

func paragraphHasContent(p *document.Paragraph) bool {
	if p == nil {
		return false
	}
	if len(p.Runs) > 0 || len(p.Inlines) > 0 || p.StyleName != "" || p.RawPPr != "" {
		return true
	}
	st := p.Style
	return st.Alignment != "" || st.NumStyleName != "" || st.SpaceBefore != 0 || st.SpaceAfter != 0 ||
		st.LineSpacing != 0 || st.Indent != 0 || st.IndentRight != 0 || st.FirstLine != 0 || st.Hanging != 0 ||
		st.KeepNext || st.KeepLines || st.PageBreakBefore || st.Shading != nil || st.Borders != nil || len(st.TabStops) > 0
}

type fieldState struct {
	active     bool
	inResult   bool
	depth      int
	pendingEnd bool
	fontSet    bool
	instr      strings.Builder
	result     strings.Builder
	font       style.FontStyle
}

func (r *docxReader) readParagraph(dec *xml.Decoder) (*document.Paragraph, string, error) {
	p := &document.Paragraph{}
	var sect string
	var fld fieldState
	for {
		tok, err := dec.Token()
		if err != nil {
			return p, sect, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "pPr":
				s, err := r.readPPr(dec, p)
				if err != nil {
					return p, sect, err
				}
				if s != "" {
					sect = s
				}
			case "r":
				if err := r.readRun(dec, p, &fld); err != nil {
					return p, sect, err
				}
			case "hyperlink":
				if err := r.readHyperlink(dec, t, p); err != nil {
					return p, sect, err
				}
			case "fldSimple":
				if err := r.readFldSimple(dec, t, p); err != nil {
					return p, sect, err
				}
			case "bookmarkStart":
				bm := &document.Bookmark{ID: atoiAttr(t, "id"), Span: true}
				bm.Name, _ = attr(t, "name")
				p.Inlines = append(p.Inlines, bm)
				if err := skipRest(dec); err != nil {
					return p, sect, err
				}
			case "bookmarkEnd":
				p.Inlines = append(p.Inlines, &document.BookmarkEnd{ID: atoiAttr(t, "id")})
				if err := skipRest(dec); err != nil {
					return p, sect, err
				}
			case "commentRangeStart":
				p.Inlines = append(p.Inlines, &document.CommentRangeStart{ID: atoiAttr(t, "id")})
				if err := skipRest(dec); err != nil {
					return p, sect, err
				}
			case "commentRangeEnd":
				p.Inlines = append(p.Inlines, &document.CommentRangeEnd{ID: atoiAttr(t, "id")})
				if err := skipRest(dec); err != nil {
					return p, sect, err
				}
			case "ins", "moveTo":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return p, sect, err
				}
				text := visibleText(raw) + deletedText(raw)
				p.Inlines = append(p.Inlines, &document.RawXML{XML: raw, Text: text})
				if text != "" {
					p.Runs = append(p.Runs, &document.Run{Text: text})
				}
			case "del":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return p, sect, err
				}
				text := deletedText(raw)
				p.Inlines = append(p.Inlines, &document.RawXML{XML: raw, Text: text})
				if text != "" {
					p.Runs = append(p.Runs, &document.Run{Text: text})
				}
			case "moveFrom":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return p, sect, err
				}
				// The moved words live here when the destination does not repeat
				// them (insertions nested inside the move). Keep the raw element
				// so a rewrite still writes moveFrom markup.
				text := collectVisible(raw, false) + deletedText(raw)
				p.Inlines = append(p.Inlines, &document.RawXML{XML: raw, Text: text})
				if text != "" {
					p.Runs = append(p.Runs, &document.Run{Text: text})
				}
			case "sdt", "smartTag", "customXml":
				if err := r.unwrapInline(dec, p, &fld); err != nil {
					return p, sect, err
				}
			default:
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return p, sect, err
				}
				text := visibleText(raw)
				p.Inlines = append(p.Inlines, &document.RawXML{XML: raw, Text: text})
				if text != "" {
					p.Runs = append(p.Runs, &document.Run{Text: text})
				}
			}
		case xml.EndElement:
			if fld.active {
				r.emitField(p, &fld)
			}
			return p, sect, nil
		}
	}
}

func (r *docxReader) unwrapInline(dec *xml.Decoder, p *document.Paragraph, fld *fieldState) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "sdtContent", "sdtPr", "sdtEndPr":
				if t.Name.Local != "sdtContent" {
					if _, err := r.captureElement(dec, t); err != nil {
						return err
					}
					continue
				}
				if err := r.unwrapInline(dec, p, fld); err != nil {
					return err
				}
			case "r":
				if err := r.readRun(dec, p, fld); err != nil {
					return err
				}
			case "hyperlink":
				if err := r.readHyperlink(dec, t, p); err != nil {
					return err
				}
			case "t":
				text, err := readCharData(dec)
				if err != nil {
					return err
				}
				if text != "" {
					run := &document.Run{Text: text}
					p.Inlines = append(p.Inlines, run)
					p.Runs = append(p.Runs, run)
				}
			default:
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return err
				}
				text := visibleText(raw)
				p.Inlines = append(p.Inlines, &document.RawXML{XML: raw, Text: text})
				if text != "" {
					p.Runs = append(p.Runs, &document.Run{Text: text})
				}
			}
		case xml.EndElement:
			return nil
		}
	}
}

func (r *docxReader) readPPr(dec *xml.Decoder, p *document.Paragraph) (string, error) {
	var children []string
	var sect string
	for {
		tok, err := dec.Token()
		if err != nil {
			return sect, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			raw, err := r.captureElement(dec, t)
			if err != nil {
				return sect, err
			}
			if t.Name.Local == "sectPr" {
				sect = raw
				continue
			}
			children = append(children, raw)
			r.interpretPPrChild(p, t.Name.Local, raw)
		case xml.EndElement:
			joined := strings.Join(children, "")
			if joined != "" && !strings.Contains(joined, "pt\"") && !strings.Contains(joined, "in\"") {
				p.RawPPr = "<w:pPr>" + joined + "</w:pPr>"
			}
			return sect, nil
		}
	}
}

func (r *docxReader) interpretPPr(p *document.Paragraph, raw string) {
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local == "pPr" {
			continue
		}
		child, err := r.captureElement(dec, se)
		if err != nil {
			return
		}
		r.interpretPPrChild(p, se.Name.Local, child)
	}
}

func (r *docxReader) interpretPPrChild(p *document.Paragraph, local, raw string) {
	dec := xml.NewDecoder(strings.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return
	}
	se, ok := tok.(xml.StartElement)
	if !ok {
		return
	}
	switch local {
	case "pStyle":
		p.StyleName, _ = attr(se, "val")
	case "jc":
		p.Style.Alignment, _ = attr(se, "val")
	case "keepNext":
		p.Style.KeepNext = onOff(se)
	case "keepLines":
		p.Style.KeepLines = onOff(se)
	case "pageBreakBefore":
		p.Style.PageBreakBefore = onOff(se)
	case "widowControl":
		p.Style.WidowControl = onOff(se)
	case "spacing":
		if n, ok := parseTwips(attrMust(se, "before")); ok {
			p.Style.SpaceBefore = n
		}
		if n, ok := parseTwips(attrMust(se, "after")); ok {
			p.Style.SpaceAfter = n
		}
		if n, ok := parseTwips(attrMust(se, "line")); ok {
			p.Style.LineSpacing = n
		}
		if v, ok := attr(se, "lineRule"); ok {
			p.Style.LineRule = v
		}
	case "ind":
		if n, ok := parseTwips(attrMust(se, "left")); ok {
			p.Style.Indent = n
		} else if n, ok := parseTwips(attrMust(se, "start")); ok {
			p.Style.Indent = n
		}
		if n, ok := parseTwips(attrMust(se, "right")); ok {
			p.Style.IndentRight = n
		} else if n, ok := parseTwips(attrMust(se, "end")); ok {
			p.Style.IndentRight = n
		}
		if n, ok := parseTwips(attrMust(se, "firstLine")); ok {
			p.Style.FirstLine = n
		}
		if n, ok := parseTwips(attrMust(se, "hanging")); ok {
			p.Style.Hanging = n
		}
	case "numPr":
		r.readNumPr(raw, p)
	case "shd":
		p.Style.Shading = shadingFrom(se)
	}
}

func (r *docxReader) readNumPr(raw string, p *document.Paragraph) {
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "ilvl":
			if n, err := strconv.Atoi(attrMust(se, "val")); err == nil {
				p.Style.NumLevel = n
			}
		case "numId":
			if v, ok := attr(se, "val"); ok {
				// numId 0 turns numbering off, including numbering inherited
				// from the paragraph style.
				p.Style.NumStyleName = v
			}
		}
	}
}

func shadingFrom(se xml.StartElement) *style.Shading {
	sh := &style.Shading{}
	sh.Fill, _ = attr(se, "fill")
	sh.Color, _ = attr(se, "color")
	sh.Pattern, _ = attr(se, "val")
	return sh
}

func (r *docxReader) readRun(dec *xml.Decoder, p *document.Paragraph, fld *fieldState) error {
	run := &document.Run{}
	var extras []document.Element
	sawText := false
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "rPr":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return err
				}
				fs, styleName, simple := interpretRPr(raw)
				run.Style = fs
				run.StyleName = styleName
				if !simple {
					run.RawRPr = raw
				}
			case "t":
				text, err := readCharData(dec)
				if err != nil {
					return err
				}
				run.Text += text
				sawText = true
			case "delText":
				if _, err := readCharData(dec); err != nil {
					return err
				}
			case "instrText":
				text, err := readCharData(dec)
				if err != nil {
					return err
				}
				if fld != nil && fld.active && !fld.inResult {
					fld.instr.WriteString(text)
				}
			case "tab":
				run.Tab = true
				if err := skipRest(dec); err != nil {
					return err
				}
			case "br", "cr":
				run.Break = true
				run.BreakType = "textWrapping"
				if t.Name.Local == "br" {
					if v, ok := attr(t, "type"); ok && v != "" {
						run.BreakType = v
					}
				}
				if err := skipRest(dec); err != nil {
					return err
				}
			case "fldChar":
				kind, _ := attr(t, "fldCharType")
				noteField(fld, kind)
				if err := skipRest(dec); err != nil {
					return err
				}
			case "footnoteReference":
				id := atoiAttr(t, "id")
				fn := r.footnotes[id]
				if fn == nil {
					fn = &document.Footnote{ID: id}
					r.doc.RegisterFootnote(fn)
					r.footnotes[id] = fn
				}
				extras = append(extras, fn)
				if err := skipRest(dec); err != nil {
					return err
				}
			case "endnoteReference":
				id := atoiAttr(t, "id")
				en := r.endnotes[id]
				if en == nil {
					en = &document.Endnote{ID: id}
					r.doc.RegisterEndnote(en)
					r.endnotes[id] = en
				}
				extras = append(extras, en)
				if err := skipRest(dec); err != nil {
					return err
				}
			case "commentReference":
				id := atoiAttr(t, "id")
				c := r.comments[id]
				if c == nil {
					c = &document.Comment{ID: id}
					r.doc.RegisterComment(c)
					r.comments[id] = c
				}
				extras = append(extras, c)
				if err := skipRest(dec); err != nil {
					return err
				}
			case "ruby":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return err
				}
				extras = append(extras, parseRuby(raw))
			case "drawing", "pict", "object", "AlternateContent":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return err
				}
				extras = append(extras, r.interpretGraphic(raw))
			case "sym":
				if ch, ok := attr(t, "char"); ok {
					if n, err := strconv.ParseInt(ch, 16, 32); err == nil {
						run.Text += string(rune(n))
						sawText = true
					}
				}
				if err := skipRest(dec); err != nil {
					return err
				}
			case "softHyphen", "noBreakHyphen":
				run.Text += "-"
				sawText = true
				if err := skipRest(dec); err != nil {
					return err
				}
			case "lastRenderedPageBreak", "footnoteRef", "endnoteRef", "separator", "continuationSeparator":
				if err := skipRest(dec); err != nil {
					return err
				}
			default:
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return err
				}
				if strings.TrimSpace(raw) != "" {
					extras = append(extras, &document.RawXML{XML: raw, Text: visibleText(raw)})
				}
			}
		case xml.EndElement:
			r.finishRun(p, fld, run, sawText, extras)
			return nil
		}
	}
}

func noteField(fld *fieldState, kind string) {
	if fld == nil {
		return
	}
	switch kind {
	case "begin":
		if !fld.active {
			fld.active = true
			fld.inResult = false
			fld.depth = 1
			fld.instr.Reset()
			fld.result.Reset()
		} else {
			fld.depth++
		}
	case "separate":
		if fld.active && fld.depth == 1 {
			fld.inResult = true
		}
	case "end":
		if !fld.active {
			return
		}
		if fld.depth > 1 {
			fld.depth--
			return
		}
		fld.pendingEnd = true
	}
}

func (r *docxReader) finishRun(p *document.Paragraph, fld *fieldState, run *document.Run, sawText bool, extras []document.Element) {
	inField := fld != nil && fld.active
	if inField && !fld.inResult {
		// w:t before a field separator is still document text. Unclosed
		// fields store ordinary words there; instrText stays in the field.
		if sawText && run.Text != "" {
			p.Inlines = append(p.Inlines, run)
			p.Runs = append(p.Runs, run)
		}
		for _, ex := range extras {
			p.Inlines = append(p.Inlines, ex)
		}
		if fld.pendingEnd {
			r.emitField(p, fld)
		}
		return
	}
	if inField && fld.inResult {
		if sawText || run.Text != "" {
			fld.result.WriteString(run.Text)
			if !fld.fontSet {
				fld.font = run.Style
				fld.fontSet = true
			}
			p.Runs = append(p.Runs, &document.Run{Text: run.Text, Style: run.Style, StyleName: run.StyleName})
		}
		for _, ex := range extras {
			p.Inlines = append(p.Inlines, ex)
		}
		if fld.pendingEnd {
			r.emitField(p, fld)
		}
		return
	}
	meaningful := sawText || run.Text != "" || run.Tab || run.Break || run.RawRPr != "" || run.StyleName != "" || runHasProps(&run.Style)
	if meaningful {
		p.Inlines = append(p.Inlines, run)
		p.Runs = append(p.Runs, run)
	}
	for _, ex := range extras {
		p.Inlines = append(p.Inlines, ex)
		switch e := ex.(type) {
		case *document.RawXML:
			if e.Text != "" {
				text := e.Text
				if drawingMarkup(e.XML) && !strings.HasSuffix(text, "\n") {
					// Separate text boxes so "box" and the next shape's "A"
					// stay distinct words.
					text += "\n"
				}
				p.Runs = append(p.Runs, &document.Run{Text: text})
			}
		case *document.Ruby:
			if e.Base != "" {
				p.Runs = append(p.Runs, &document.Run{Text: e.Base})
			}
		case *document.Footnote:
			// The reference number is not a w:t. Surface it so extracted text
			// matches what Word, LibreOffice, and POI show at the anchor.
			if e.ID > 0 {
				p.Runs = append(p.Runs, &document.Run{Text: strconv.Itoa(e.ID)})
			}
		case *document.Endnote:
			if e.ID > 0 {
				p.Runs = append(p.Runs, &document.Run{Text: strconv.Itoa(e.ID)})
			}
		}
	}
}

func drawingMarkup(xml string) bool {
	return strings.Contains(xml, "<w:pict") || strings.Contains(xml, "<w:drawing") ||
		strings.Contains(xml, "<w:txbx") || strings.Contains(xml, "AlternateContent") ||
		strings.Contains(xml, "<v:textbox") || strings.Contains(xml, "<w:txbxContent")
}

func runHasProps(fs *style.FontStyle) bool {
	return fs.Bold || fs.Italic || fs.Size > 0 || fs.Color != "" || fs.Name != "" || fs.Underline != "" ||
		fs.Strikethrough || fs.HighlightColor != "" || fs.AllCaps || fs.SmallCaps || fs.Hidden
}

func (r *docxReader) emitField(p *document.Paragraph, fld *fieldState) {
	f := &document.Field{
		Instr:  strings.TrimSpace(fld.instr.String()),
		Result: fld.result.String(),
		Font:   fld.font,
	}
	p.Inlines = append(p.Inlines, f)
	if f.Result == "" {
		// Keep the instruction visible when Word stored no cached result.
		p.Runs = append(p.Runs, &document.Run{Text: f.Instr})
	}
	fld.active = false
	fld.pendingEnd = false
	fld.inResult = false
	fld.depth = 0
	fld.instr.Reset()
	fld.result.Reset()
}

func (r *docxReader) readHyperlink(dec *xml.Decoder, start xml.StartElement, p *document.Paragraph) error {
	h := &document.Hyperlink{}
	if id, ok := attr(start, "id"); ok {
		h.RelID = id
		if tgt, ok := r.rels[id]; ok {
			h.URL = tgt
		}
	}
	if anchor, ok := attr(start, "anchor"); ok && anchor != "" {
		if h.URL == "" {
			h.URL = anchor
			h.IsInternal = true
		}
	}
	h.Tooltip, _ = attr(start, "tooltip")
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "r" {
				tmp := &document.Paragraph{}
				if err := r.readRun(dec, tmp, nil); err != nil {
					return err
				}
				for _, run := range tmp.Runs {
					if run.Tab {
						h.Text += "\t"
					}
					h.Text += run.Text
					if h.Font.Name == "" && h.Font.Size == 0 {
						h.Font = run.Style
					}
				}
				continue
			}
			raw, err := r.captureElement(dec, t)
			if err != nil {
				return err
			}
			h.Text += visibleText(raw)
		case xml.EndElement:
			p.Inlines = append(p.Inlines, h)
			if h.Text != "" {
				p.Runs = append(p.Runs, &document.Run{Text: h.Text, Style: h.Font})
			}
			return nil
		}
	}
}

func (r *docxReader) readFldSimple(dec *xml.Decoder, start xml.StartElement, p *document.Paragraph) error {
	instr, _ := attr(start, "instr")
	tmp := &document.Paragraph{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "r" {
				if err := r.readRun(dec, tmp, nil); err != nil {
					return err
				}
				continue
			}
			if _, err := r.captureElement(dec, t); err != nil {
				return err
			}
		case xml.EndElement:
			var result strings.Builder
			for _, run := range tmp.Runs {
				result.WriteString(run.Text)
			}
			f := &document.Field{Instr: strings.TrimSpace(instr), Result: result.String()}
			p.Inlines = append(p.Inlines, f)
			if f.Result != "" {
				p.Runs = append(p.Runs, &document.Run{Text: f.Result})
			}
			return nil
		}
	}
}

func interpretRPr(raw string) (style.FontStyle, string, bool) {
	var fs style.FontStyle
	var styleName string
	simple := true
	modeled := map[string]bool{
		"rPr": true, "rStyle": true, "rFonts": true, "b": true, "bCs": true, "i": true, "iCs": true,
		"caps": true, "smallCaps": true, "strike": true, "dstrike": true, "outline": true, "shadow": true,
		"emboss": true, "imprint": true, "noProof": true, "vanish": true, "color": true, "spacing": true,
		"kern": true, "sz": true, "szCs": true, "highlight": true, "u": true, "effect": true, "vertAlign": true,
		"rtl": true, "lang": true,
	}
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
		if !modeled[se.Name.Local] {
			simple = false
		}
		switch se.Name.Local {
		case "rStyle":
			styleName, _ = attr(se, "val")
		case "rFonts":
			if v, ok := attr(se, "ascii"); ok && v != "" {
				fs.Name = v
			} else if v, ok := attr(se, "hAnsi"); ok {
				fs.Name = v
			}
			if v, ok := attr(se, "eastAsia"); ok {
				fs.NameEastAsia = v
			}
			if _, ok := attr(se, "asciiTheme"); ok {
				simple = false
			}
			if _, ok := attr(se, "hAnsiTheme"); ok {
				simple = false
			}
			if _, ok := attr(se, "hint"); ok {
				simple = false
			}
			if v, ok := attr(se, "cs"); ok && v != "" && v != fs.Name {
				simple = false
			}
		case "b":
			if onOff(se) {
				fs.Bold = true
			} else if fs.Off == nil {
				fs.Off = map[string]bool{}
				fs.Off["b"] = true
			} else {
				fs.Off["b"] = true
			}
		case "i":
			if onOff(se) {
				fs.Italic = true
			} else {
				if fs.Off == nil {
					fs.Off = map[string]bool{}
				}
				fs.Off["i"] = true
			}
		case "strike":
			fs.Strikethrough = onOff(se)
		case "dstrike":
			fs.DoubleStrikethrough = onOff(se)
		case "outline":
			fs.Outline = onOff(se)
		case "shadow":
			fs.Shadow = onOff(se)
		case "emboss":
			fs.Emboss = onOff(se)
		case "imprint":
			fs.Imprint = onOff(se)
		case "noProof":
			fs.NoProof = onOff(se)
		case "vanish":
			fs.Hidden = onOff(se)
		case "caps":
			fs.AllCaps = onOff(se)
		case "smallCaps":
			fs.SmallCaps = onOff(se)
		case "rtl":
			fs.RightToLeft = onOff(se)
		case "color":
			fs.Color, _ = attr(se, "val")
			if _, ok := attr(se, "themeColor"); ok {
				simple = false
			}
		case "highlight":
			fs.HighlightColor, _ = attr(se, "val")
		case "u":
			fs.Underline, _ = attr(se, "val")
			if fs.Underline == "" {
				fs.Underline = "single"
			}
			fs.UnderlineColor, _ = attr(se, "color")
		case "sz":
			if n, err := strconv.Atoi(attrMust(se, "val")); err == nil {
				fs.Size = float64(n) / 2
			}
		case "spacing":
			if n, err := strconv.Atoi(attrMust(se, "val")); err == nil {
				fs.Spacing = float64(n) / 20
			}
		case "kern":
			if n, err := strconv.Atoi(attrMust(se, "val")); err == nil {
				fs.Kerning = float64(n) / 2
			}
		case "vertAlign":
			switch attrMust(se, "val") {
			case "superscript":
				fs.SuperScript = true
			case "subscript":
				fs.SubScript = true
			}
		case "effect":
			fs.TextEffect, _ = attr(se, "val")
		case "lang":
			fs.Lang, _ = attr(se, "val")
		case "bCs":
			if fs.Bold && !onOff(se) || !fs.Bold && onOff(se) {
				simple = false
			}
		case "iCs":
			if fs.Italic && !onOff(se) || !fs.Italic && onOff(se) {
				simple = false
			}
		case "szCs":
			if n, err := strconv.Atoi(attrMust(se, "val")); err == nil && fs.Size > 0 && int(fs.Size*2) != n {
				simple = false
			}
		}
	}
	return fs, styleName, simple
}

func (r *docxReader) interpretGraphic(raw string) document.Element {
	complex := strings.Contains(raw, "txbx") || strings.Contains(raw, "textbox") ||
		strings.Contains(raw, ":anchor") ||
		(strings.Contains(raw, "AlternateContent") && (strings.Contains(raw, "txbx") || strings.Contains(raw, "wps:") || strings.Contains(raw, ":anchor")))
	id := findQuoted(raw, "embed")
	if id == "" && strings.Contains(raw, "imagedata") {
		id = findQuoted(raw, "id")
	}
	var img *document.Image
	if id != "" && !strings.EqualFold(r.relMode[id], "External") {
		if target := r.rels[id]; target != "" {
			img = r.imageFor(target, raw)
		}
	}
	// Floating drawings, text boxes, and alternate-content shapes are kept
	// verbatim. A plain inline picture becomes an Image element.
	if complex || img == nil {
		return &document.RawXML{XML: raw, Text: visibleText(raw)}
	}
	return img
}

func (r *docxReader) imageFor(target, raw string) *document.Image {
	part := resolvePart(r.curPart, target)
	base := pathBase(part)
	if img, ok := r.imagesByName[base]; ok {
		return img
	}
	data, err := r.readZipFile(part)
	if err != nil {
		return nil
	}
	ext := strings.TrimPrefix(strings.ToLower(pathBaseExt(base)), ".")
	if ext == "jpg" {
		ext = "jpeg"
	}
	img := &document.Image{
		Name:     base,
		Data:     data,
		MimeType: imageContentType(ext),
		ID:       r.doc.AllocImageIDPublic(),
	}
	if cx := findQuoted(raw, "cx"); cx != "" {
		if n, err := strconv.ParseInt(cx, 10, 64); err == nil && n > 0 {
			img.Style.Width = common.EmuToPixel(n)
		}
	}
	if cy := findQuoted(raw, "cy"); cy != "" {
		if n, err := strconv.ParseInt(cy, 10, 64); err == nil && n > 0 {
			img.Style.Height = common.EmuToPixel(n)
		}
	}
	r.doc.AppendImage(img)
	r.imagesByName[base] = img
	return img
}

func pathBaseExt(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[i:]
	}
	return ""
}

func (r *docxReader) readSdt(dec *xml.Decoder) (*document.Sdt, error) {
	sdt := &document.Sdt{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return sdt, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "sdtPr":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return sdt, err
				}
				sdt.PrXML = raw
			case "sdtEndPr":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return sdt, err
				}
				sdt.EndPrXML = raw
			case "sdtContent":
				elems, err := r.readBlockChildren(dec)
				if err != nil {
					return sdt, err
				}
				sdt.Elements = elems
			default:
				if _, err := r.captureElement(dec, t); err != nil {
					return sdt, err
				}
			}
		case xml.EndElement:
			return sdt, nil
		}
	}
}

func (r *docxReader) readTable(dec *xml.Decoder) (*document.Table, error) {
	tbl := &document.Table{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return tbl, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "tblPr":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return tbl, err
				}
				tbl.RawTblPr = raw
				r.interpretTblPr(tbl, raw)
			case "tblGrid":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return tbl, err
				}
				r.interpretGrid(tbl, raw)
			case "tr":
				row, err := r.readRow(dec)
				if err != nil {
					return tbl, err
				}
				tbl.Rows = append(tbl.Rows, row)
			case "sdt":
				rows, err := r.readSdtRows(dec)
				if err != nil {
					return tbl, err
				}
				tbl.Rows = append(tbl.Rows, rows...)
			default:
				if _, err := r.captureElement(dec, t); err != nil {
					return tbl, err
				}
			}
		case xml.EndElement:
			return tbl, nil
		}
	}
}

func (r *docxReader) interpretTblPr(tbl *document.Table, raw string) {
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "tblStyle":
			tbl.StyleName, _ = attr(se, "val")
		case "tblW":
			if n, ok := parseTwips(attrMust(se, "w")); ok {
				tbl.Style.Width = n
			}
			tbl.Style.WidthType, _ = attr(se, "type")
		case "jc":
			tbl.Style.Alignment, _ = attr(se, "val")
		case "tblLayout":
			tbl.Style.Layout, _ = attr(se, "type")
		case "tblInd":
			if n, ok := parseTwips(attrMust(se, "w")); ok {
				tbl.Style.Indent = n
			}
		case "tblCellSpacing":
			if n, ok := parseTwips(attrMust(se, "w")); ok {
				tbl.Style.CellSpacing = n
			}
		}
	}
}

func (r *docxReader) interpretGrid(tbl *document.Table, raw string) {
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "gridCol" {
			continue
		}
		if n, ok := parseTwips(attrMust(se, "w")); ok {
			tbl.Grid = append(tbl.Grid, n)
		}
	}
}

func (r *docxReader) readRow(dec *xml.Decoder) (*document.Row, error) {
	row := &document.Row{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return row, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "trPr":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return row, err
				}
				row.RawTrPr = raw
				r.interpretTrPr(row, raw)
			case "tblPrEx":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return row, err
				}
				row.PrefixXML = raw
			case "tc":
				cell, err := r.readCell(dec)
				if err != nil {
					return row, err
				}
				row.Cells = append(row.Cells, cell)
			case "sdt":
				cells, err := r.readSdtCells(dec)
				if err != nil {
					return row, err
				}
				row.Cells = append(row.Cells, cells...)
			default:
				if _, err := r.captureElement(dec, t); err != nil {
					return row, err
				}
			}
		case xml.EndElement:
			return row, nil
		}
	}
}

func (r *docxReader) interpretTrPr(row *document.Row, raw string) {
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "trHeight":
			if n, ok := parseTwips(attrMust(se, "val")); ok {
				row.Style.Height = n
			}
			if v, ok := attr(se, "hRule"); ok {
				row.Style.HeightRule = v
			}
		case "tblHeader":
			row.Style.IsHeader = onOff(se)
		case "cantSplit":
			row.Style.CantSplit = onOff(se)
		}
	}
}

// readSdtRows reads a content control that wraps table rows
// (w:tbl/w:sdt/w:sdtContent/w:tr).
func (r *docxReader) readSdtRows(dec *xml.Decoder) ([]*document.Row, error) {
	var rows []*document.Row
	for {
		tok, err := dec.Token()
		if err != nil {
			return rows, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "sdtContent", "sdt":
				inner, err := r.readSdtRows(dec)
				if err != nil {
					return rows, err
				}
				rows = append(rows, inner...)
			case "tr":
				row, err := r.readRow(dec)
				if err != nil {
					return rows, err
				}
				rows = append(rows, row)
			default:
				if _, err := r.captureElement(dec, t); err != nil {
					return rows, err
				}
			}
		case xml.EndElement:
			return rows, nil
		}
	}
}

// readSdtCells reads a content control that wraps one or more w:tc elements
// (w:tr/w:sdt/w:sdtContent/w:tc), which Word uses for cell-level controls.
func (r *docxReader) readSdtCells(dec *xml.Decoder) ([]*document.Cell, error) {
	var cells []*document.Cell
	for {
		tok, err := dec.Token()
		if err != nil {
			return cells, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "sdtContent":
				inner, err := r.readSdtCells(dec)
				if err != nil {
					return cells, err
				}
				cells = append(cells, inner...)
			case "tc":
				cell, err := r.readCell(dec)
				if err != nil {
					return cells, err
				}
				cells = append(cells, cell)
			case "sdt":
				inner, err := r.readSdtCells(dec)
				if err != nil {
					return cells, err
				}
				cells = append(cells, inner...)
			default:
				if _, err := r.captureElement(dec, t); err != nil {
					return cells, err
				}
			}
		case xml.EndElement:
			return cells, nil
		}
	}
}

func (r *docxReader) readCell(dec *xml.Decoder) (*document.Cell, error) {
	cell := &document.Cell{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return cell, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "tcPr":
				raw, err := r.captureElement(dec, t)
				if err != nil {
					return cell, err
				}
				cell.RawTcPr = raw
				r.interpretTcPr(cell, raw)
			default:
				el, _, err := r.readOneBlock(dec, t)
				if err != nil {
					return cell, err
				}
				if el != nil {
					cell.Elements = append(cell.Elements, el)
				}
			}
		case xml.EndElement:
			return cell, nil
		}
	}
}

func (r *docxReader) interpretTcPr(cell *document.Cell, raw string) {
	dec := xml.NewDecoder(strings.NewReader(raw))
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "tcW":
			if n, ok := parseTwips(attrMust(se, "w")); ok {
				cell.Style.Width = n
			}
			cell.Style.WidthType, _ = attr(se, "type")
		case "gridSpan":
			if n, err := strconv.Atoi(attrMust(se, "val")); err == nil {
				cell.Style.GridSpan = n
			}
		case "vMerge":
			cell.Style.VMerge = "continue"
			if v, ok := attr(se, "val"); ok && v != "" {
				cell.Style.VMerge = v
			}
		case "vAlign":
			cell.Style.VAlign, _ = attr(se, "val")
		case "shd":
			cell.Style.Shading = shadingFrom(se)
		case "noWrap":
			cell.Style.NoWrap = onOff(se)
		case "textDirection":
			cell.Style.TextDirection, _ = attr(se, "val")
		}
	}
}

// parseRuby splits a phonetic guide into the base text and the reading.
// The original element is kept so a rewrite does not rebuild ruby markup.
func parseRuby(raw string) *document.Ruby {
	rb := &document.Ruby{XML: raw}
	dec := xml.NewDecoder(strings.NewReader(raw))
	mode := ""
	depth := 0
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "rt":
				if mode == "" {
					mode = "rt"
					depth = 1
				} else {
					depth++
				}
			case "rubyBase":
				if mode == "" {
					mode = "base"
					depth = 1
				} else {
					depth++
				}
			case "t":
				s, err := readCharData(dec)
				if err != nil {
					return rb
				}
				switch mode {
				case "rt":
					rb.Reading += s
				case "base":
					rb.Base += s
				}
			}
		case xml.EndElement:
			if depth == 0 {
				continue
			}
			if (mode == "rt" && t.Name.Local == "rt") || (mode == "base" && t.Name.Local == "rubyBase") {
				depth--
				if depth == 0 {
					mode = ""
				}
			}
		}
	}
	return rb
}

func readCharData(decoder *xml.Decoder) (string, error) {
	var text strings.Builder
	for {
		tok, err := decoder.Token()
		if err != nil {
			return text.String(), err
		}
		switch t := tok.(type) {
		case xml.CharData:
			text.Write(t)
		case xml.EndElement:
			return text.String(), nil
		case xml.StartElement:
			if err := skipRest(decoder); err != nil {
				return text.String(), err
			}
		}
	}
}

func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
