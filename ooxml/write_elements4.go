package ooxml

import (
	"fmt"
	"strings"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

func (w *Writer) writeTableBorders(b *strings.Builder, ts *style.TableStyle) {
	if ts.BorderTop == nil && ts.BorderBottom == nil && ts.BorderLeft == nil &&
		ts.BorderRight == nil && ts.BorderInsideH == nil && ts.BorderInsideV == nil {
		return
	}
	b.WriteString(`<w:tblBorders>`)
	writeTableBorder(b, "top", ts.BorderTop)
	writeTableBorder(b, "left", ts.BorderLeft)
	writeTableBorder(b, "bottom", ts.BorderBottom)
	writeTableBorder(b, "right", ts.BorderRight)
	writeTableBorder(b, "insideH", ts.BorderInsideH)
	writeTableBorder(b, "insideV", ts.BorderInsideV)
	b.WriteString(`</w:tblBorders>`)
}

func writeTableBorder(b *strings.Builder, name string, border *style.Border) {
	if border == nil {
		return
	}
	b.WriteString(fmt.Sprintf(`<w:%s w:val="%s" w:sz="%d" w:space="%d" w:color="%s"/>`,
		name, border.Style, border.Size, border.Space, border.Color))
}

func (w *Writer) writeTableCellMargins(b *strings.Builder, ts *style.TableStyle) {
	if ts.CellMarginTop == 0 && ts.CellMarginBottom == 0 && ts.CellMarginLeft == 0 && ts.CellMarginRight == 0 {
		return
	}
	b.WriteString(`<w:tblCellMar>`)
	if ts.CellMarginTop > 0 {
		b.WriteString(fmt.Sprintf(`<w:top w:w="%d" w:type="dxa"/>`, ts.CellMarginTop))
	}
	if ts.CellMarginLeft > 0 {
		b.WriteString(fmt.Sprintf(`<w:left w:w="%d" w:type="dxa"/>`, ts.CellMarginLeft))
	}
	if ts.CellMarginBottom > 0 {
		b.WriteString(fmt.Sprintf(`<w:bottom w:w="%d" w:type="dxa"/>`, ts.CellMarginBottom))
	}
	if ts.CellMarginRight > 0 {
		b.WriteString(fmt.Sprintf(`<w:right w:w="%d" w:type="dxa"/>`, ts.CellMarginRight))
	}
	b.WriteString(`</w:tblCellMar>`)
}

func (w *Writer) writeTableRow(b *strings.Builder, row *document.Row) {
	b.WriteString(`<w:tr>`)

	// Row properties
	rs := &row.Style
	if rs.Height > 0 || rs.IsHeader || rs.CantSplit {
		b.WriteString(`<w:trPr>`)
		if rs.Height > 0 {
			rule := rs.HeightRule
			if rule == "" {
				rule = "atLeast"
			}
			b.WriteString(fmt.Sprintf(`<w:trHeight w:val="%d" w:hRule="%s"/>`, rs.Height, rule))
		}
		if rs.IsHeader {
			b.WriteString(`<w:tblHeader/>`)
		}
		if rs.CantSplit {
			b.WriteString(`<w:cantSplit/>`)
		}
		b.WriteString(`</w:trPr>`)
	}

	for _, cell := range row.Cells {
		w.writeTableCell(b, cell)
	}

	b.WriteString(`</w:tr>`)
}
