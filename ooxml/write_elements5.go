package ooxml

import (
	"fmt"
	"sort"
	"strings"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

func (w *Writer) writeTableCell(b *strings.Builder, cell *document.Cell) {
	b.WriteString(`<w:tc>`)

	// Cell properties
	cs := &cell.Style
	b.WriteString(`<w:tcPr>`)
	if cs.Width > 0 {
		wType := cs.WidthType
		if wType == "" {
			wType = "dxa"
		}
		b.WriteString(fmt.Sprintf(`<w:tcW w:w="%d" w:type="%s"/>`, cs.Width, wType))
	}
	if cs.GridSpan > 1 {
		b.WriteString(fmt.Sprintf(`<w:gridSpan w:val="%d"/>`, cs.GridSpan))
	}
	if cs.VMerge != "" {
		if cs.VMerge == "restart" {
			b.WriteString(`<w:vMerge w:val="restart"/>`)
		} else {
			b.WriteString(`<w:vMerge/>`)
		}
	}
	if cs.VAlign != "" {
		b.WriteString(fmt.Sprintf(`<w:vAlign w:val="%s"/>`, cs.VAlign))
	}
	if cs.Shading != nil {
		b.WriteString(fmt.Sprintf(`<w:shd w:val="%s" w:color="%s" w:fill="%s"/>`,
			cs.Shading.Pattern, cs.Shading.Color, cs.Shading.Fill))
	}
	if cs.TextDirection != "" {
		b.WriteString(fmt.Sprintf(`<w:textDirection w:val="%s"/>`, cs.TextDirection))
	}
	if cs.NoWrap {
		b.WriteString(`<w:noWrap/>`)
	}
	writeCellBorders(b, cs)
	b.WriteString(`</w:tcPr>`)

	// Cell content (must have at least one paragraph)
	if len(cell.Elements) == 0 {
		b.WriteString(`<w:p/>`)
	} else {
		for _, elem := range cell.Elements {
			w.writeElementXML(b, elem)
		}
	}

	b.WriteString(`</w:tc>`)
}

func writeCellBorders(b *strings.Builder, cs *style.CellStyle) {
	if cs.BorderTop == nil && cs.BorderBottom == nil && cs.BorderLeft == nil && cs.BorderRight == nil {
		return
	}
	b.WriteString(`<w:tcBorders>`)
	writeTableBorder(b, "top", cs.BorderTop)
	writeTableBorder(b, "left", cs.BorderLeft)
	writeTableBorder(b, "bottom", cs.BorderBottom)
	writeTableBorder(b, "right", cs.BorderRight)
	b.WriteString(`</w:tcBorders>`)
}

func (w *Writer) writeListItem(b *strings.Builder, li *document.ListItem) {
	b.WriteString(`<w:p>`)
	b.WriteString(`<w:pPr>`)
	b.WriteString(`<w:numPr>`)
	b.WriteString(fmt.Sprintf(`<w:ilvl w:val="%d"/>`, li.Depth))

	// Determine numId
	numID := "1" // default bullet
	if li.ListStyle != "" {
		// Sort names for deterministic numbering ID lookup (must match writeNumbering order)
		ns := w.doc.NumberingStyles()
		names := make([]string, 0, len(ns))
		for name := range ns {
			names = append(names, name)
		}
		sort.Strings(names)
		for idx, name := range names {
			if name == li.ListStyle {
				numID = fmt.Sprintf("%d", idx+1)
				break
			}
		}
	}
	b.WriteString(fmt.Sprintf(`<w:numId w:val="%s"/>`, numID))
	b.WriteString(`</w:numPr>`)
	b.WriteString(`</w:pPr>`)

	b.WriteString(`<w:r>`)
	rpr := fontStyleToXML(&li.Font, nil)
	if rpr != "" {
		b.WriteString(rpr)
	}
	w.writeTextElement(b, li.Text)
	b.WriteString(`</w:r>`)
	b.WriteString(`</w:p>`)
}

func (w *Writer) writeFootnoteRef(b *strings.Builder, fn *document.Footnote) {
	b.WriteString(`<w:r>`)
	b.WriteString(`<w:rPr><w:rStyle w:val="FootnoteReference"/></w:rPr>`)
	b.WriteString(fmt.Sprintf(`<w:footnoteReference w:id="%d"/>`, fn.ID))
	b.WriteString(`</w:r>`)
}

func (w *Writer) writeEndnoteRef(b *strings.Builder, en *document.Endnote) {
	b.WriteString(`<w:r>`)
	b.WriteString(`<w:rPr><w:rStyle w:val="EndnoteReference"/></w:rPr>`)
	b.WriteString(fmt.Sprintf(`<w:endnoteReference w:id="%d"/>`, en.ID))
	b.WriteString(`</w:r>`)
}

func (w *Writer) writeBookmark(b *strings.Builder, bm *document.Bookmark) {
	b.WriteString(fmt.Sprintf(`<w:bookmarkStart w:id="%d" w:name="%s"/>`, bm.ID, escapeXML(bm.Name)))
	b.WriteString(fmt.Sprintf(`<w:bookmarkEnd w:id="%d"/>`, bm.ID))
}

func (w *Writer) writeBookmarkEnd(b *strings.Builder, bm *document.BookmarkEnd) {
	b.WriteString(fmt.Sprintf(`<w:bookmarkEnd w:id="%d"/>`, bm.ID))
}

func (w *Writer) writeCommentRef(b *strings.Builder, c *document.Comment) {
	b.WriteString(`<w:r>`)
	b.WriteString(fmt.Sprintf(`<w:commentReference w:id="%d"/>`, c.ID))
	b.WriteString(`</w:r>`)
}

func (w *Writer) writePreserveText(b *strings.Builder, pt *document.PreserveText) {
	b.WriteString(`<w:p>`)
	ppr := paraStyleToXML(&pt.Para, "")
	if ppr != "" {
		b.WriteString(ppr)
	}

	// Parse preserve text for field codes like {PAGE}, {NUMPAGES}
	parts := parsePreserveText(pt.Text)
	for _, part := range parts {
		if part.isField {
			b.WriteString(`<w:r>`)
			rpr := fontStyleToXML(&pt.Font, nil)
			if rpr != "" {
				b.WriteString(rpr)
			}
			b.WriteString(`<w:fldChar w:fldCharType="begin"/></w:r>`)
			b.WriteString(`<w:r>`)
			b.WriteString(fmt.Sprintf(`<w:instrText xml:space="preserve"> %s </w:instrText>`, part.text))
			b.WriteString(`</w:r>`)
			b.WriteString(`<w:r><w:fldChar w:fldCharType="separate"/></w:r>`)
			b.WriteString(`<w:r><w:t>0</w:t></w:r>`)
			b.WriteString(`<w:r><w:fldChar w:fldCharType="end"/></w:r>`)
		} else {
			b.WriteString(`<w:r>`)
			rpr := fontStyleToXML(&pt.Font, nil)
			if rpr != "" {
				b.WriteString(rpr)
			}
			w.writeTextElement(b, part.text)
			b.WriteString(`</w:r>`)
		}
	}
	b.WriteString(`</w:p>`)
}

type preservePart struct {
	text    string
	isField bool
}

func parsePreserveText(text string) []preservePart {
	var parts []preservePart
	for {
		start := strings.Index(text, "{")
		if start == -1 {
			if text != "" {
				parts = append(parts, preservePart{text: text})
			}
			break
		}
		end := strings.Index(text[start:], "}")
		if end == -1 {
			parts = append(parts, preservePart{text: text})
			break
		}
		end += start

		if start > 0 {
			parts = append(parts, preservePart{text: text[:start]})
		}
		fieldName := text[start+1 : end]
		parts = append(parts, preservePart{text: fieldName, isField: true})
		text = text[end+1:]
	}
	return parts
}

func (w *Writer) writeTOC(b *strings.Builder, toc *document.TOC) {
	minD := toc.MinDepth
	if minD == 0 {
		minD = 1
	}
	maxD := toc.MaxDepth
	if maxD == 0 {
		maxD = 9
	}

	b.WriteString(`<w:p>`)
	b.WriteString(`<w:r><w:fldChar w:fldCharType="begin"/></w:r>`)
	b.WriteString(fmt.Sprintf(`<w:r><w:instrText xml:space="preserve"> TOC \o "%d-%d" \h \z \u </w:instrText></w:r>`, minD, maxD))
	b.WriteString(`<w:r><w:fldChar w:fldCharType="separate"/></w:r>`)
	b.WriteString(`<w:r><w:t>Table of Contents (right-click to update)</w:t></w:r>`)
	b.WriteString(`<w:r><w:fldChar w:fldCharType="end"/></w:r>`)
	b.WriteString(`</w:p>`)
}

func (w *Writer) writeCheckBox(b *strings.Builder, cb *document.CheckBox) {
	b.WriteString(`<w:p>`)
	ppr := paraStyleToXML(&cb.Para, "")
	if ppr != "" {
		b.WriteString(ppr)
	}
	b.WriteString(`<w:r>`)
	b.WriteString(`<w:fldChar w:fldCharType="begin">`)
	b.WriteString(`<w:ffData>`)
	b.WriteString(fmt.Sprintf(`<w:name w:val="%s"/>`, escapeXML(cb.Name)))
	b.WriteString(`<w:enabled/>`)
	checked := "0"
	if cb.Checked {
		checked = "1"
	}
	b.WriteString(fmt.Sprintf(`<w:checkBox><w:sizeAuto/><w:default w:val="%s"/></w:checkBox>`, checked))
	b.WriteString(`</w:ffData>`)
	b.WriteString(`</w:fldChar>`)
	b.WriteString(`</w:r>`)
	b.WriteString(`<w:r><w:instrText xml:space="preserve"> FORMCHECKBOX </w:instrText></w:r>`)
	b.WriteString(`<w:r><w:fldChar w:fldCharType="end"/></w:r>`)
	if cb.Text != "" {
		b.WriteString(`<w:r>`)
		rpr := fontStyleToXML(&cb.Font, nil)
		if rpr != "" {
			b.WriteString(rpr)
		}
		w.writeTextElement(b, cb.Text)
		b.WriteString(`</w:r>`)
	}
	b.WriteString(`</w:p>`)
}

func (w *Writer) writeLine(b *strings.Builder, l *document.Line) {
	// Lines are rendered as VML shapes for compatibility
	b.WriteString(`<w:p>`)
	b.WriteString(`<w:r>`)
	b.WriteString(`<w:pict>`)
	color := l.Color
	if color == "" {
		color = "000000"
	}
	weight := l.Weight
	if weight == 0 {
		weight = 1
	}
	b.WriteString(fmt.Sprintf(`<v:line xmlns:v="urn:schemas-microsoft-com:vml" style="width:%dpt;height:%dpt" strokecolor="#%s" strokeweight="%dpt"/>`,
		l.Width, l.Height, color, weight))
	b.WriteString(`</w:pict>`)
	b.WriteString(`</w:r>`)
	b.WriteString(`</w:p>`)
}

func (w *Writer) writeFormField(b *strings.Builder, ff *document.FormField) {
	b.WriteString(`<w:p>`)
	ppr := paraStyleToXML(&ff.Para, "")
	if ppr != "" {
		b.WriteString(ppr)
	}

	switch ff.Type {
	case document.FormFieldTypeText:
		w.writeTextInputField(b, ff)
	case document.FormFieldTypeDropDown:
		w.writeDropDownField(b, ff)
	case document.FormFieldTypeCheckBox:
		w.writeCheckBoxFormField(b, ff)
	}

	b.WriteString(`</w:p>`)
}

func (w *Writer) writeTextInputField(b *strings.Builder, ff *document.FormField) {
	b.WriteString(`<w:r>`)
	b.WriteString(`<w:fldChar w:fldCharType="begin">`)
	b.WriteString(`<w:ffData>`)
	b.WriteString(fmt.Sprintf(`<w:name w:val="%s"/>`, escapeXML(ff.Name)))
	if ff.Enabled {
		b.WriteString(`<w:enabled/>`)
	}
	if ff.CalcOnExit {
		b.WriteString(`<w:calcOnExit w:val="true"/>`)
	}
	b.WriteString(`<w:textInput>`)
	if ff.DefaultValue != "" {
		b.WriteString(fmt.Sprintf(`<w:default w:val="%s"/>`, escapeXML(ff.DefaultValue)))
	}
	if ff.MaxLength > 0 {
		b.WriteString(fmt.Sprintf(`<w:maxLength w:val="%d"/>`, ff.MaxLength))
	}
	b.WriteString(`</w:textInput>`)
	b.WriteString(`</w:ffData>`)
	b.WriteString(`</w:fldChar>`)
	b.WriteString(`</w:r>`)
	b.WriteString(`<w:r><w:instrText xml:space="preserve"> FORMTEXT </w:instrText></w:r>`)
	b.WriteString(`<w:r><w:fldChar w:fldCharType="separate"/></w:r>`)
	val := ff.Value
	if val == "" {
		val = ff.DefaultValue
	}
	b.WriteString(`<w:r>`)
	rpr := fontStyleToXML(&ff.Font, nil)
	if rpr != "" {
		b.WriteString(rpr)
	}
	w.writeTextElement(b, val)
	b.WriteString(`</w:r>`)
	b.WriteString(`<w:r><w:fldChar w:fldCharType="end"/></w:r>`)
}

func (w *Writer) writeDropDownField(b *strings.Builder, ff *document.FormField) {
	b.WriteString(`<w:r>`)
	b.WriteString(`<w:fldChar w:fldCharType="begin">`)
	b.WriteString(`<w:ffData>`)
	b.WriteString(fmt.Sprintf(`<w:name w:val="%s"/>`, escapeXML(ff.Name)))
	if ff.Enabled {
		b.WriteString(`<w:enabled/>`)
	}
	if ff.CalcOnExit {
		b.WriteString(`<w:calcOnExit w:val="true"/>`)
	}
	b.WriteString(`<w:ddList>`)
	if ff.DefaultValue != "" {
		// Find index of default value
		for i, v := range ff.PossibleValues {
			if v == ff.DefaultValue {
				b.WriteString(fmt.Sprintf(`<w:default w:val="%d"/>`, i))
				break
			}
		}
	}
	for _, v := range ff.PossibleValues {
		b.WriteString(fmt.Sprintf(`<w:listEntry w:val="%s"/>`, escapeXML(v)))
	}
	b.WriteString(`</w:ddList>`)
	b.WriteString(`</w:ffData>`)
	b.WriteString(`</w:fldChar>`)
	b.WriteString(`</w:r>`)
	b.WriteString(`<w:r><w:instrText xml:space="preserve"> FORMDROPDOWN </w:instrText></w:r>`)
	b.WriteString(`<w:r><w:fldChar w:fldCharType="separate"/></w:r>`)
	b.WriteString(`<w:r><w:fldChar w:fldCharType="end"/></w:r>`)
}

func (w *Writer) writeCheckBoxFormField(b *strings.Builder, ff *document.FormField) {
	b.WriteString(`<w:r>`)
	b.WriteString(`<w:fldChar w:fldCharType="begin">`)
	b.WriteString(`<w:ffData>`)
	b.WriteString(fmt.Sprintf(`<w:name w:val="%s"/>`, escapeXML(ff.Name)))
	if ff.Enabled {
		b.WriteString(`<w:enabled/>`)
	}
	checked := "0"
	if ff.Value == "true" || ff.Value == "1" {
		checked = "1"
	}
	b.WriteString(fmt.Sprintf(`<w:checkBox><w:sizeAuto/><w:default w:val="%s"/></w:checkBox>`, checked))
	b.WriteString(`</w:ffData>`)
	b.WriteString(`</w:fldChar>`)
	b.WriteString(`</w:r>`)
	b.WriteString(`<w:r><w:instrText xml:space="preserve"> FORMCHECKBOX </w:instrText></w:r>`)
	b.WriteString(`<w:r><w:fldChar w:fldCharType="end"/></w:r>`)
}
