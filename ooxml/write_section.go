package ooxml

import (
	"fmt"
	"strings"

	"github.com/VantageDataChat/GoWord/document"
)

func (w *Writer) writeSectionProperties(b *strings.Builder, sec *document.Section, isLast bool) {
	if isLast {
		// Last section properties go inside w:body > w:sectPr
		b.WriteString(`<w:sectPr>`)
	} else {
		// Non-last sections: properties go inside the last paragraph's pPr
		b.WriteString(`<w:p><w:pPr><w:sectPr>`)
	}

	// Header/footer references
	for _, hp := range w.headerParts {
		for _, h := range sec.Headers {
			if w.headerPartMatches(hp.filename, h) {
				relID := w.addDocRel(relHeader, hp.filename)
				b.WriteString(fmt.Sprintf(`<w:headerReference w:type="%s" r:id="%s"/>`, h.Type, relID))
			}
		}
	}
	for _, fp := range w.footerParts {
		for _, f := range sec.Footers {
			if w.footerPartMatches(fp.filename, f) {
				relID := w.addDocRel(relFooter, fp.filename)
				b.WriteString(fmt.Sprintf(`<w:footerReference w:type="%s" r:id="%s"/>`, f.Type, relID))
			}
		}
	}

	ss := &sec.Style
	b.WriteString(fmt.Sprintf(`<w:pgSz w:w="%d" w:h="%d"`, ss.PageWidth, ss.PageHeight))
	if ss.Orientation == "landscape" {
		b.WriteString(` w:orient="landscape"`)
	}
	b.WriteString(`/>`)

	b.WriteString(fmt.Sprintf(`<w:pgMar w:top="%d" w:right="%d" w:bottom="%d" w:left="%d" w:header="%d" w:footer="%d"/>`,
		ss.MarginTop, ss.MarginRight, ss.MarginBottom, ss.MarginLeft, ss.HeaderHeight, ss.FooterHeight))

	if ss.ColumnCount > 1 {
		b.WriteString(fmt.Sprintf(`<w:cols w:num="%d"`, ss.ColumnCount))
		if ss.ColumnSpacing > 0 {
			b.WriteString(fmt.Sprintf(` w:space="%d"`, ss.ColumnSpacing))
		}
		b.WriteString(`/>`)
	}

	if ss.PageNumberStart != nil {
		b.WriteString(fmt.Sprintf(`<w:pgNumType w:start="%d"/>`, *ss.PageNumberStart))
	}

	if ss.BreakType != "" && ss.BreakType != "nextPage" {
		b.WriteString(fmt.Sprintf(`<w:type w:val="%s"/>`, ss.BreakType))
	}

	if isLast {
		b.WriteString(`</w:sectPr>`)
	} else {
		b.WriteString(`</w:sectPr></w:pPr></w:p>`)
	}
}

func (w *Writer) generateHeaderFooterParts() {
	headerIdx := 1
	footerIdx := 1

	// Track if we've written watermark to at least one header
	w.watermarkWritten = false

	for _, sec := range w.doc.Sections {
		for _, h := range sec.Headers {
			filename := fmt.Sprintf("header%d.xml", headerIdx)
			content := w.generateHeaderXML(h)
			w.headerParts = append(w.headerParts, headerFooterPart{
				filename: filename,
				content:  content,
				header:   h,
			})
			headerIdx++
		}

		// If document has watermark but section has no header, create one
		if len(sec.Headers) == 0 && !w.watermarkWritten && (w.doc.WatermarkTextValue() != nil || w.doc.WatermarkPictureValue() != nil) {
			h := &document.Header{Type: "default"}
			sec.Headers = append(sec.Headers, h)
			filename := fmt.Sprintf("header%d.xml", headerIdx)
			content := w.generateHeaderXML(h)
			w.headerParts = append(w.headerParts, headerFooterPart{
				filename: filename,
				content:  content,
				header:   h,
			})
			headerIdx++
		}

		for _, f := range sec.Footers {
			filename := fmt.Sprintf("footer%d.xml", footerIdx)
			content := w.generateFooterXML(f)
			w.footerParts = append(w.footerParts, headerFooterPart{
				filename: filename,
				content:  content,
				footer:   f,
			})
			footerIdx++
		}
	}
}

func (w *Writer) generateHeaderXML(h *document.Header) []byte {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:hdr xmlns:w="%s" xmlns:r="%s" xmlns:wp="%s" xmlns:a="%s" xmlns:pic="%s">`,
		nsW, nsR, nsWP, nsA, nsPIC))

	// Watermark (added to first header)
	w.writeWatermarkInHeader(&b)

	for _, elem := range h.Elements {
		w.writeElementXML(&b, elem)
	}
	if len(h.Elements) == 0 {
		b.WriteString(`<w:p/>`)
	}
	b.WriteString(`</w:hdr>`)
	return []byte(b.String())
}

func (w *Writer) generateFooterXML(f *document.Footer) []byte {
	var b strings.Builder
	b.WriteString(xmlHeader())
	b.WriteString(fmt.Sprintf(`<w:ftr xmlns:w="%s" xmlns:r="%s" xmlns:wp="%s" xmlns:a="%s" xmlns:pic="%s">`,
		nsW, nsR, nsWP, nsA, nsPIC))
	for _, elem := range f.Elements {
		w.writeElementXML(&b, elem)
	}
	if len(f.Elements) == 0 {
		b.WriteString(`<w:p/>`)
	}
	b.WriteString(`</w:ftr>`)
	return []byte(b.String())
}

func (w *Writer) writeHeaderFooterFiles() error {
	for _, hp := range w.headerParts {
		if err := w.addZipFile("word/"+hp.filename, hp.content); err != nil {
			return err
		}
	}
	for _, fp := range w.footerParts {
		if err := w.addZipFile("word/"+fp.filename, fp.content); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) headerPartMatches(filename string, h *document.Header) bool {
	for _, hp := range w.headerParts {
		if hp.filename == filename {
			return hp.header == h
		}
	}
	return false
}

func (w *Writer) footerPartMatches(filename string, f *document.Footer) bool {
	for _, fp := range w.footerParts {
		if fp.filename == filename {
			return fp.footer == f
		}
	}
	return false
}

func (w *Writer) writeWatermarkInHeader(b *strings.Builder) {
	if w.watermarkWritten {
		return
	}

	if wt := w.doc.WatermarkTextValue(); wt != nil {
		w.watermarkWritten = true
		w.writeTextWatermark(b, wt)
	} else if wp := w.doc.WatermarkPictureValue(); wp != nil {
		w.watermarkWritten = true
		w.writePictureWatermark(b, wp)
	}
}

func (w *Writer) writeTextWatermark(b *strings.Builder, wt *document.WatermarkText) {
	color := wt.Color
	if color == "" {
		color = "C0C0C0"
	}
	font := wt.Font
	if font == "" {
		font = "Calibri"
	}
	size := wt.Size
	if size == 0 {
		size = 1 // auto
	}

	fontStyle := ""
	if wt.Bold {
		fontStyle += "font-weight:bold;"
	}
	if wt.Italic {
		fontStyle += "font-style:italic;"
	}

	b.WriteString(`<w:p><w:pPr><w:pStyle w:val="Header"/></w:pPr>`)
	b.WriteString(`<w:r><w:pict>`)
	b.WriteString(`<v:shapetype xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office" id="_x0000_t136" coordsize="21600,21600" o:spt="136" adj="10800" path="m@7,l@8,m@5,21600l@6,21600e">`)
	b.WriteString(`<v:formulas><v:f eqn="sum #0 0 10800"/><v:f eqn="prod #0 2 1"/><v:f eqn="sum 21600 0 @1"/><v:f eqn="sum 0 0 @2"/><v:f eqn="sum 21600 0 @3"/><v:f eqn="if @0 @3 0"/><v:f eqn="if @0 21600 @1"/><v:f eqn="if @0 0 @2"/><v:f eqn="if @0 @4 21600"/></v:formulas>`)
	b.WriteString(`<v:path textpathok="t" o:connecttype="custom" o:connectlocs="@9,0;@10,10800;@11,21600;@12,10800" o:connectangles="270,180,90,0"/>`)
	b.WriteString(`<v:textpath on="t" fitshape="t"/>`)
	b.WriteString(`<v:handles><v:h position="#0,bottomRight" xrange="6629,14971"/></v:handles>`)
	b.WriteString(`<o:lock v:ext="edit" text="t" shapetype="t"/>`)
	b.WriteString(`</v:shapetype>`)
	b.WriteString(fmt.Sprintf(`<v:shape xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office" id="PowerPlusWaterMarkObject" o:spid="_x0000_s2049" type="#_x0000_t136" style="position:absolute;margin-left:0;margin-top:0;width:468pt;height:234pt;z-index:-251658752;mso-position-horizontal:center;mso-position-horizontal-relative:margin;mso-position-vertical:center;mso-position-vertical-relative:margin" o:allowincell="f" fillcolor="#%s" stroked="f">`, color))
	b.WriteString(fmt.Sprintf(`<v:textpath style="font-family:&quot;%s&quot;;font-size:%dpt;%s" string="%s"/>`, escapeXML(font), int(size), fontStyle, escapeXML(wt.Text)))
	b.WriteString(`</v:shape>`)
	b.WriteString(`</w:pict></w:r>`)
	b.WriteString(`</w:p>`)
}

func (w *Writer) writePictureWatermark(b *strings.Builder, wp *document.WatermarkPicture) {
	// Add the watermark image to the document
	img := &document.Image{
		Source:   wp.Source,
		Data:     wp.Data,
		MimeType: wp.MimeType,
		RelID:    w.doc.AllocRelIDPublic(),
		ID:       w.doc.AllocImageIDPublic(),
	}
	w.doc.AppendImage(img)

	width := wp.Width
	if width == 0 {
		width = 468
	}
	height := wp.Height
	if height == 0 {
		height = 468
	}

	gain := "19661f" // ~30% for washout
	blacklevel := "22938f"
	if !wp.Washout {
		gain = "65536f"
		blacklevel = "0"
	}

	b.WriteString(`<w:p><w:pPr><w:pStyle w:val="Header"/></w:pPr>`)
	b.WriteString(`<w:r><w:pict>`)
	b.WriteString(fmt.Sprintf(`<v:shapetype xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office" id="_x0000_t75" coordsize="21600,21600" o:spt="75" o:preferrelative="t" path="m@4@5l@4@11@9@11@9@5xe" filled="f" stroked="f">`))
	b.WriteString(`<v:stroke joinstyle="miter"/>`)
	b.WriteString(`<v:formulas><v:f eqn="if lineDrawn pixelLineWidth 0"/><v:f eqn="sum @0 1 0"/><v:f eqn="sum 0 0 @1"/><v:f eqn="prod @2 1 2"/><v:f eqn="prod @3 21600 pixelWidth"/><v:f eqn="prod @3 21600 pixelHeight"/><v:f eqn="sum @0 0 1"/><v:f eqn="prod @6 1 2"/><v:f eqn="prod @7 21600 pixelWidth"/><v:f eqn="sum @8 21600 0"/><v:f eqn="prod @7 21600 pixelHeight"/><v:f eqn="sum @10 21600 0"/></v:formulas>`)
	b.WriteString(`<v:path o:extrusionok="f" gradientshapeok="t" o:connecttype="rect"/>`)
	b.WriteString(`<o:lock v:ext="edit" aspectratio="t"/>`)
	b.WriteString(`</v:shapetype>`)
	b.WriteString(fmt.Sprintf(`<v:shape xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:r="%s" id="WordPictureWatermark" o:spid="_x0000_s2050" type="#_x0000_t75" style="position:absolute;margin-left:0;margin-top:0;width:%dpt;height:%dpt;z-index:-251658751;mso-position-horizontal:center;mso-position-horizontal-relative:margin;mso-position-vertical:center;mso-position-vertical-relative:margin" o:allowincell="f">`, nsR, width, height))
	b.WriteString(fmt.Sprintf(`<v:imagedata r:id="%s" o:title="" gain="%s" blacklevel="%s"/>`, img.RelID, gain, blacklevel))
	b.WriteString(`</v:shape>`)
	b.WriteString(`</w:pict></w:r>`)
	b.WriteString(`</w:p>`)
}
