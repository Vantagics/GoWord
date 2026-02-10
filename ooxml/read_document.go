package ooxml

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

func (r *docxReader) readDocument() error {
	data, err := r.readZipFile("word/document.xml")
	if err != nil {
		return err
	}

	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	// Find the body element
	var inBody bool
	sec := r.doc.AddSection()

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			localName := t.Name.Local
			switch {
			case localName == "body":
				inBody = true
			case inBody && localName == "p":
				p, err := r.readParagraph(decoder, t)
				if err != nil {
					return err
				}
				if p != nil {
					sec.Elements = append(sec.Elements, p)
				}
			case inBody && localName == "tbl":
				tbl, err := r.readTable(decoder)
				if err != nil {
					return err
				}
				if tbl != nil {
					sec.Elements = append(sec.Elements, tbl)
				}
			case inBody && localName == "sectPr":
				r.readSectionProperties(decoder, sec)
			}
		case xml.EndElement:
			if t.Name.Local == "body" {
				inBody = false
			}
		}
	}

	return nil
}

func (r *docxReader) readParagraph(decoder *xml.Decoder, start xml.StartElement) (*document.Paragraph, error) {
	p := &document.Paragraph{}
	depth := 1

	for {
		tok, err := decoder.Token()
		if err != nil {
			return p, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			switch t.Name.Local {
			case "pStyle":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						p.StyleName = attr.Value
					}
				}
			case "jc":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						p.Style.Alignment = attr.Value
					}
				}
			case "r":
				run, err := r.readRun(decoder)
				if err != nil {
					return p, err
				}
				if run != nil {
					p.Runs = append(p.Runs, run)
					depth-- // readRun consumed the end element
				}
			case "hyperlink":
				hl := r.readHyperlinkStart(t)
				run, err := r.readHyperlinkContent(decoder)
				if err != nil {
					return p, err
				}
				if run != nil && hl != nil {
					// Add hyperlink text as a run with the URL info
					run.Style.Color = "0563C1"
					run.Style.Underline = "single"
					p.Runs = append(p.Runs, run)
				}
				depth-- // consumed end element
			}
		case xml.EndElement:
			depth--
			if depth == 0 {
				return p, nil
			}
		}
	}
}

func (r *docxReader) readRun(decoder *xml.Decoder) (*document.Run, error) {
	run := &document.Run{}
	depth := 1

	for {
		tok, err := decoder.Token()
		if err != nil {
			return run, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			switch t.Name.Local {
			case "t":
				text, err := readCharData(decoder)
				if err != nil {
					return run, err
				}
				run.Text += text
				depth-- // consumed end element
			case "br":
				run.Break = true
			case "b":
				run.Style.Bold = true
			case "i":
				run.Style.Italic = true
			case "u":
				run.Style.Underline = "single"
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "val":
						run.Style.Underline = attr.Value
					case "color":
						run.Style.UnderlineColor = attr.Value
					}
				}
			case "strike":
				run.Style.Strikethrough = true
			case "dstrike":
				run.Style.DoubleStrikethrough = true
			case "outline":
				run.Style.Outline = true
			case "shadow":
				run.Style.Shadow = true
			case "emboss":
				run.Style.Emboss = true
			case "imprint":
				run.Style.Imprint = true
			case "effect":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						run.Style.TextEffect = attr.Value
					}
				}
			case "rtl":
				run.Style.RightToLeft = true
			case "color":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						run.Style.Color = attr.Value
					}
				}
			case "highlight":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						run.Style.HighlightColor = attr.Value
					}
				}
			case "sz":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						if hp, err := strconv.Atoi(attr.Value); err == nil {
							run.Style.Size = float64(hp) / 2.0
						}
					}
				}
			case "kern":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						if hp, err := strconv.Atoi(attr.Value); err == nil {
							run.Style.Kerning = float64(hp) / 2.0
						}
					}
				}
			case "spacing":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						if tw, err := strconv.Atoi(attr.Value); err == nil {
							run.Style.Spacing = float64(tw) / 20.0
						}
					}
				}
			case "rFonts":
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "ascii":
						run.Style.Name = attr.Value
					case "eastAsia":
						run.Style.NameEastAsia = attr.Value
					}
				}
			case "rStyle":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						run.StyleName = attr.Value
					}
				}
			case "vertAlign":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						switch attr.Value {
						case "superscript":
							run.Style.SuperScript = true
						case "subscript":
							run.Style.SubScript = true
						}
					}
				}
			case "caps":
				run.Style.AllCaps = true
			case "smallCaps":
				run.Style.SmallCaps = true
			case "vanish":
				run.Style.Hidden = true
			case "noProof":
				run.Style.NoProof = true
			case "lang":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						run.Style.Lang = attr.Value
					}
				}
			}
		case xml.EndElement:
			depth--
			if depth == 0 {
				return run, nil
			}
		}
	}
}

func (r *docxReader) readHyperlinkStart(start xml.StartElement) *document.Hyperlink {
	h := &document.Hyperlink{}
	for _, attr := range start.Attr {
		if attr.Name.Local == "id" {
			if target, ok := r.rels[attr.Value]; ok {
				h.URL = target
			}
			h.RelID = attr.Value
		}
	}
	return h
}

func (r *docxReader) readHyperlinkContent(decoder *xml.Decoder) (*document.Run, error) {
	run := &document.Run{}
	depth := 1

	for {
		tok, err := decoder.Token()
		if err != nil {
			return run, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "t" {
				text, err := readCharData(decoder)
				if err != nil {
					return run, err
				}
				run.Text += text
				depth--
			}
		case xml.EndElement:
			depth--
			if depth == 0 {
				return run, nil
			}
		}
	}
}

func (r *docxReader) readTable(decoder *xml.Decoder) (*document.Table, error) {
	tbl := &document.Table{}
	depth := 1

	for {
		tok, err := decoder.Token()
		if err != nil {
			return tbl, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "tr" {
				row, err := r.readTableRow(decoder)
				if err != nil {
					return tbl, err
				}
				tbl.Rows = append(tbl.Rows, row)
				depth--
			}
		case xml.EndElement:
			depth--
			if depth == 0 {
				return tbl, nil
			}
		}
	}
}

func (r *docxReader) readTableRow(decoder *xml.Decoder) (*document.Row, error) {
	row := &document.Row{}
	depth := 1

	for {
		tok, err := decoder.Token()
		if err != nil {
			return row, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "tc" {
				cell, err := r.readTableCell(decoder)
				if err != nil {
					return row, err
				}
				row.Cells = append(row.Cells, cell)
				depth--
			}
		case xml.EndElement:
			depth--
			if depth == 0 {
				return row, nil
			}
		}
	}
}

func (r *docxReader) readTableCell(decoder *xml.Decoder) (*document.Cell, error) {
	cell := &document.Cell{}
	depth := 1

	for {
		tok, err := decoder.Token()
		if err != nil {
			return cell, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			switch t.Name.Local {
			case "p":
				p, err := r.readParagraph(decoder, t)
				if err != nil {
					return cell, err
				}
				if p != nil {
					cell.Elements = append(cell.Elements, p)
				}
				depth--
			case "gridSpan":
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						if v, err := strconv.Atoi(attr.Value); err == nil {
							cell.Style.GridSpan = v
						}
					}
				}
			case "vMerge":
				cell.Style.VMerge = "continue"
				for _, attr := range t.Attr {
					if attr.Name.Local == "val" {
						cell.Style.VMerge = attr.Value
					}
				}
			case "tcW":
				for _, attr := range t.Attr {
					if attr.Name.Local == "w" {
						if v, err := strconv.Atoi(attr.Value); err == nil {
							cell.Style.Width = v
						}
					}
				}
			case "shd":
				shading := &style.Shading{}
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "fill":
						shading.Fill = attr.Value
					case "color":
						shading.Color = attr.Value
					case "val":
						shading.Pattern = attr.Value
					}
				}
				cell.Style.Shading = shading
			}
		case xml.EndElement:
			depth--
			if depth == 0 {
				return cell, nil
			}
		}
	}
}

func (r *docxReader) readSectionProperties(decoder *xml.Decoder, sec *document.Section) {
	depth := 1
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			switch t.Name.Local {
			case "pgSz":
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "w":
						if v, err := strconv.Atoi(attr.Value); err == nil {
							sec.Style.PageWidth = v
						}
					case "h":
						if v, err := strconv.Atoi(attr.Value); err == nil {
							sec.Style.PageHeight = v
						}
					case "orient":
						sec.Style.Orientation = attr.Value
					}
				}
			case "pgMar":
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "top":
						if v, err := strconv.Atoi(attr.Value); err == nil {
							sec.Style.MarginTop = v
						}
					case "bottom":
						if v, err := strconv.Atoi(attr.Value); err == nil {
							sec.Style.MarginBottom = v
						}
					case "left":
						if v, err := strconv.Atoi(attr.Value); err == nil {
							sec.Style.MarginLeft = v
						}
					case "right":
						if v, err := strconv.Atoi(attr.Value); err == nil {
							sec.Style.MarginRight = v
						}
					}
				}
			case "cols":
				for _, attr := range t.Attr {
					if attr.Name.Local == "num" {
						if v, err := strconv.Atoi(attr.Value); err == nil {
							sec.Style.ColumnCount = v
						}
					}
				}
			}
		case xml.EndElement:
			depth--
			if depth == 0 {
				return
			}
		}
	}
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
		}
	}
}

func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
