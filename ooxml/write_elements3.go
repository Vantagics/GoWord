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

	name := w.imageFileName(img)
	relID := w.imageRel(img)
	docID := img.ID
	if docID <= 0 {
		docID = 1
	}
	if name == "" {
		name = "Picture"
	}

	b.WriteString(`<w:r>`)
	b.WriteString(`<w:drawing>`)
	anchorOpen, wrap, floating := drawingAnchor(img.Style.WrappingStyle, img.Style.Alignment)
	if floating {
		b.WriteString(anchorOpen)
	} else {
		b.WriteString(`<wp:inline distT="0" distB="0" distL="0" distR="0">`)
	}
	b.WriteString(fmt.Sprintf(`<wp:extent cx="%d" cy="%d"/>`, cx, cy))
	b.WriteString(`<wp:effectExtent l="0" t="0" r="0" b="0"/>`)
	if floating {
		b.WriteString(wrap)
	}
	b.WriteString(fmt.Sprintf(`<wp:docPr id="%d" name="%s"/>`, docID, escapeXML(name)))
	b.WriteString(`<wp:cNvGraphicFramePr><a:graphicFrameLocks noChangeAspect="1"/></wp:cNvGraphicFramePr>`)
	b.WriteString(`<a:graphic>`)
	b.WriteString(fmt.Sprintf(`<a:graphicData uri="%s">`, nsPIC))
	b.WriteString(fmt.Sprintf(`<pic:pic xmlns:pic="%s">`, nsPIC))
	b.WriteString(fmt.Sprintf(`<pic:nvPicPr><pic:cNvPr id="%d" name="%s"/><pic:cNvPicPr><a:picLocks noChangeAspect="1"/></pic:cNvPicPr></pic:nvPicPr>`, docID, escapeXML(name)))
	b.WriteString(fmt.Sprintf(`<pic:blipFill><a:blip r:embed="%s"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill>`, relID))
	b.WriteString(fmt.Sprintf(`<pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr>`, cx, cy))
	b.WriteString(`</pic:pic>`)
	b.WriteString(`</a:graphicData>`)
	b.WriteString(`</a:graphic>`)
	if floating {
		b.WriteString(`</wp:anchor>`)
	} else {
		b.WriteString(`</wp:inline>`)
	}
	b.WriteString(`</w:drawing>`)
	b.WriteString(`</w:r>`)
}

// drawingAnchor returns the wp:anchor opening tag (through positionV) and the
// wrap element, which belongs after effectExtent. inline stays wp:inline.
func drawingAnchor(wrapping, align string) (open, wrap string, ok bool) {
	behind := "0"
	switch strings.ToLower(wrapping) {
	case "square":
		wrap = `<wp:wrapSquare wrapText="bothSides"/>`
	case "tight":
		wrap = `<wp:wrapTight wrapText="bothSides"><wp:wrapPolygon><wp:start x="0" y="0"/><wp:line x="0" y="21600"/><wp:line x="21600" y="21600"/><wp:line x="21600" y="0"/></wp:wrapPolygon></wp:wrapTight>`
	case "topandbottom", "topAndBottom":
		wrap = `<wp:wrapTopAndBottom/>`
	case "behind":
		behind = "1"
		wrap = `<wp:wrapNone/>`
	case "infrontof", "infront", "inFrontOf":
		wrap = `<wp:wrapNone/>`
	default:
		return "", "", false
	}
	if align == "" {
		align = "left"
	}
	open = fmt.Sprintf(`<wp:anchor distT="0" distB="0" distL="114300" distR="114300" simplePos="0" relativeHeight="251658240" behindDoc="%s" locked="0" layoutInCell="1" allowOverlap="1"><wp:simplePos x="0" y="0"/><wp:positionH relativeFrom="column"><wp:align>%s</wp:align></wp:positionH><wp:positionV relativeFrom="paragraph"><wp:posOffset>0</wp:posOffset></wp:positionV>`, behind, escapeXML(align))
	return open, wrap, true
}

func (w *Writer) writeTable(b *strings.Builder, t *document.Table) {
	b.WriteString(`<w:tbl>`)

	// Table properties. CT_TblPr order:
	// tblStyle, tblW, jc, tblCellSpacing, tblInd, tblBorders, tblLayout, tblCellMar.
	if t.RawTblPr != "" {
		b.WriteString(t.RawTblPr)
	} else {
		b.WriteString(`<w:tblPr>`)
		ts := &t.Style
		if t.StyleName != "" {
			b.WriteString(fmt.Sprintf(`<w:tblStyle w:val="%s"/>`, escapeXML(t.StyleName)))
		}
		if ts.FloatXSpec != "" || ts.FloatYSpec != "" {
			x, y := ts.FloatXSpec, ts.FloatYSpec
			if x == "" {
				x = "left"
			}
			if y == "" {
				y = "top"
			}
			b.WriteString(fmt.Sprintf(`<w:tblpPr w:horzAnchor="margin" w:vertAnchor="text" w:tblpXSpec="%s" w:tblpYSpec="%s"/>`, escapeXML(x), escapeXML(y)))
		}
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
		if ts.CellSpacing > 0 {
			b.WriteString(fmt.Sprintf(`<w:tblCellSpacing w:w="%d" w:type="dxa"/>`, ts.CellSpacing))
		}
		if ts.Indent > 0 {
			b.WriteString(fmt.Sprintf(`<w:tblInd w:w="%d" w:type="dxa"/>`, ts.Indent))
		}
		w.writeTableBorders(b, ts)
		if ts.Layout != "" {
			b.WriteString(fmt.Sprintf(`<w:tblLayout w:type="%s"/>`, ts.Layout))
		}
		w.writeTableCellMargins(b, ts)
		b.WriteString(`</w:tblPr>`)
	}

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
