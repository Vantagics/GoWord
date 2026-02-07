package document

import (
	"github.com/VantageDataChat/GoWord/style"
)

// AddRow adds a row to the table.
func (t *Table) AddRow(height int, rowStyle *style.RowStyle) *Row {
	r := &Row{doc: t.doc}
	if rowStyle != nil {
		r.Style = *rowStyle
	} else {
		r.Style.Height = height
	}
	t.Rows = append(t.Rows, r)
	return r
}

// AddCell adds a cell to the current (last) row.
func (r *Row) AddCell(width int, cellStyle *style.CellStyle) *Cell {
	c := &Cell{doc: r.doc}
	if cellStyle != nil {
		c.Style = *cellStyle
	}
	c.Style.Width = width
	r.Cells = append(r.Cells, c)
	return c
}

// AddText adds a text paragraph to a cell.
func (c *Cell) AddText(text string, fontStyle *style.FontStyle, paraStyle *style.ParagraphStyle) *Paragraph {
	p := &Paragraph{doc: c.doc}
	if paraStyle != nil {
		p.Style = *paraStyle
	}
	run := &Run{Text: text}
	if fontStyle != nil {
		run.Style = *fontStyle
	}
	p.Runs = append(p.Runs, run)
	c.Elements = append(c.Elements, p)
	return p
}

// AddTextRun adds a complex paragraph to a cell.
func (c *Cell) AddTextRun(paraStyle *style.ParagraphStyle) *TextRun {
	tr := &TextRun{doc: c.doc}
	if paraStyle != nil {
		tr.Style = *paraStyle
	}
	c.Elements = append(c.Elements, tr)
	return tr
}

// AddImage adds an image to a cell.
func (c *Cell) AddImage(src string, imgStyle *style.ImageStyle) *Image {
	img := &Image{
		Source: src,
		RelID:  c.doc.allocRelID(),
		ID:     c.doc.allocImageID(),
	}
	if imgStyle != nil {
		img.Style = *imgStyle
	}
	c.doc.images = append(c.doc.images, img)
	c.Elements = append(c.Elements, img)
	return img
}

// AddTable adds a nested table to a cell.
func (c *Cell) AddTable(tableStyle *style.TableStyle) *Table {
	t := &Table{doc: c.doc}
	if tableStyle != nil {
		t.Style = *tableStyle
	}
	c.Elements = append(c.Elements, t)
	return t
}

// AddListItem adds a list item to a cell.
func (c *Cell) AddListItem(text string, depth int, fontStyle *style.FontStyle, listStyleName string) *ListItem {
	li := &ListItem{
		Text:      text,
		Depth:     depth,
		ListStyle: listStyleName,
	}
	if fontStyle != nil {
		li.Font = *fontStyle
	}
	c.Elements = append(c.Elements, li)
	return li
}
