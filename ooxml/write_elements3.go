package ooxml

import (
	"fmt"
	"strings"

	"github.com/VantageDataChat/GoWord/common"
	"github.com/VantageDataChat/GoWord/document"
)

func (w *Writer) writeImage(b *strings.Builder, img *document.Image) {
	b.WriteString(`<w:p>`)
	w.writeInlineImage(b, img)
	b.WriteString(`</w:p>`)
}

func (w *Writer) writeInlineImage(b *strings.Builder, img *document.Image) {
	width := img.Style.Width
	if width == 0 {
		width = 100
	}
	height := img.Style.Height
	if height == 0 {
		height = 100
	}

	cx := common.PixelToEmu(width)
	cy := common.PixelToEmu(height)

	b.WriteString(`<w:r>`)
	b.WriteString(`<w:drawing>`)
	b.WriteString(fmt.Sprintf(`<wp:inline distT="0" distB="0" distL="0" distR="0">`))
	b.WriteString(fmt.Sprintf(`<wp:extent cx="%d" cy="%d"/>`, cx, cy))
	b.WriteString(`<wp:docPr id="` + fmt.Sprintf("%d", img.ID) + `" name="` + escapeXML(img.Name) + `"/>`)
	b.WriteString(`<a:graphic>`)
	b.WriteString(fmt.Sprintf(`<a:graphicData uri="%s">`, nsPIC))
	b.WriteString(fmt.Sprintf(`<pic:pic xmlns:pic="%s">`, nsPIC))
	b.WriteString(fmt.Sprintf(`<pic:nvPicPr><pic:cNvPr id="%d" name="%s"/><pic:cNvPicPr/></pic:nvPicPr>`, img.ID, escapeXML(img.Name)))
	b.WriteString(fmt.Sprintf(`<pic:blipFill><a:blip r:embed="%s"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill>`, img.RelID))
	b.WriteString(fmt.Sprintf(`<pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr>`, cx, cy))
	b.WriteString(`</pic:pic>`)
	b.WriteString(`</a:graphicData>`)
	b.WriteString(`</a:graphic>`)
	b.WriteString(`</wp:inline>`)
	b.WriteString(`</w:drawing>`)
	b.WriteString(`</w:r>`)
}

func (w *Writer) writeTable(b *strings.Builder, t *document.Table) {
	b.WriteString(`<w:tbl>`)

	// Table properties
	b.WriteString(`<w:tblPr>`)
	if t.StyleName != "" {
		b.WriteString(fmt.Sprintf(`<w:tblStyle w:val="%s"/>`, escapeXML(t.StyleName)))
	}
	ts := &t.Style
	if ts.Width > 0 {
		wType := ts.WidthType
		if wType == "" {
			wType = "dxa"
		}
		b.WriteString(fmt.Sprintf(`<w:tblW w:w="%d" w:type="%s"/>`, ts.Width, wType))
	} else {
		b.WriteString(`<w:tblW w:w="0" w:type="auto"/>`)
	}
	if ts.Alignment != "" {
		b.WriteString(fmt.Sprintf(`<w:jc w:val="%s"/>`, ts.Alignment))
	}
	w.writeTableBorders(b, ts)
	w.writeTableCellMargins(b, ts)
	if ts.Layout != "" {
		b.WriteString(fmt.Sprintf(`<w:tblLayout w:type="%s"/>`, ts.Layout))
	}
	b.WriteString(`</w:tblPr>`)

	// Table grid
	if len(t.Grid) > 0 {
		b.WriteString(`<w:tblGrid>`)
		for _, colW := range t.Grid {
			b.WriteString(fmt.Sprintf(`<w:gridCol w:w="%d"/>`, colW))
		}
		b.WriteString(`</w:tblGrid>`)
	}

	// Rows
	for _, row := range t.Rows {
		w.writeTableRow(b, row)
	}

	b.WriteString(`</w:tbl>`)
}
