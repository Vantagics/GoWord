// Package ooxml handles reading and writing Office Open XML (.docx) files.
package ooxml

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

// Namespaces used in OOXML documents.
const (
	nsW    = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	nsR    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	nsWP   = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
	nsA    = "http://schemas.openxmlformats.org/drawingml/2006/main"
	nsPIC  = "http://schemas.openxmlformats.org/drawingml/2006/picture"
	nsMC   = "http://schemas.openxmlformats.org/markup-compatibility/2006"
	nsCP   = "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
	nsDC   = "http://purl.org/dc/elements/1.1/"
	nsDCT  = "http://purl.org/dc/terms/"
	nsXSI  = "http://www.w3.org/2001/XMLSchema-instance"
	nsCT   = "http://schemas.openxmlformats.org/package/2006/content-types"
	nsEP   = "http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"
	nsVT   = "http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"

	relDoc       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	relCoreProps = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties"
	relExtProps  = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties"
	relStyles    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"
	relNumbering = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering"
	relFontTable = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable"
	relSettings  = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings"
	relImage     = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
	relHyperlink = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink"
	relHeader    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/header"
	relFooter    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer"
	relFootnotes = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footnotes"
	relEndnotes  = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/endnotes"
	relComments  = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/comments"
)

// Writer writes a Document to .docx format.
type Writer struct {
	doc *document.Document
	buf *bytes.Buffer
	zw  *zip.Writer

	// Track relationships for document.xml.rels
	docRels []relationship
	// Track content types
	contentTypes []contentTypeOverride
	// Track header/footer parts
	headerParts []headerFooterPart
	footerParts []headerFooterPart

	nextDocRelID int

	// Watermark tracking
	watermarkWritten bool
}

type relationship struct {
	ID         string
	Type       string
	Target     string
	TargetMode string // "External" for hyperlinks
}

type contentTypeOverride struct {
	PartName    string
	ContentType string
}

type headerFooterPart struct {
	filename string
	content  []byte
	header   *document.Header // non-nil for header parts
	footer   *document.Footer // non-nil for footer parts
}

// Save writes the document to a file.
func Save(doc *document.Document, path string) error {
	data, err := WriteToBytes(doc)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}

// WriteToBytes writes the document to a byte slice.
func WriteToBytes(doc *document.Document) ([]byte, error) {
	w := &Writer{
		doc:          doc,
		buf:          &bytes.Buffer{},
		nextDocRelID: 1,
	}
	w.zw = zip.NewWriter(w.buf)

	if err := w.write(); err != nil {
		return nil, err
	}

	if err := w.zw.Close(); err != nil {
		return nil, err
	}

	return w.buf.Bytes(), nil
}

// WriteToWriter writes the document to an io.Writer.
func WriteToWriter(doc *document.Document, out io.Writer) error {
	data, err := WriteToBytes(doc)
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}

func (w *Writer) write() error {
	// Pre-generate header/footer parts
	w.generateHeaderFooterParts()

	// 1. [Content_Types].xml
	if err := w.writeContentTypes(); err != nil {
		return err
	}
	// 2. _rels/.rels
	if err := w.writeRootRels(); err != nil {
		return err
	}
	// 3. docProps/core.xml
	if err := w.writeCoreProperties(); err != nil {
		return err
	}
	// 4. docProps/app.xml
	if err := w.writeAppProperties(); err != nil {
		return err
	}
	// 5. word/styles.xml
	if err := w.writeStyles(); err != nil {
		return err
	}
	// 6. word/settings.xml
	if err := w.writeSettings(); err != nil {
		return err
	}
	// 7. word/numbering.xml (if needed)
	if err := w.writeNumbering(); err != nil {
		return err
	}
	// 8. word/footnotes.xml (if needed)
	if err := w.writeFootnotes(); err != nil {
		return err
	}
	// 9. word/endnotes.xml (if needed)
	if err := w.writeEndnotes(); err != nil {
		return err
	}
	// 10. word/comments.xml (if needed)
	if err := w.writeComments(); err != nil {
		return err
	}
	// 11. Images
	if err := w.writeImages(); err != nil {
		return err
	}
	// 12. Header/footer parts
	if err := w.writeHeaderFooterFiles(); err != nil {
		return err
	}
	// 13. word/document.xml (main content)
	if err := w.writeDocument(); err != nil {
		return err
	}
	// 14. word/_rels/document.xml.rels
	if err := w.writeDocRels(); err != nil {
		return err
	}

	return nil
}

func (w *Writer) addDocRel(relType, target string) string {
	id := fmt.Sprintf("rId%d", w.nextDocRelID)
	w.nextDocRelID++
	w.docRels = append(w.docRels, relationship{ID: id, Type: relType, Target: target})
	return id
}

func (w *Writer) addDocRelExternal(relType, target string) string {
	id := fmt.Sprintf("rId%d", w.nextDocRelID)
	w.nextDocRelID++
	w.docRels = append(w.docRels, relationship{ID: id, Type: relType, Target: target, TargetMode: "External"})
	return id
}

func (w *Writer) addZipFile(name string, data []byte) error {
	fw, err := w.zw.Create(name)
	if err != nil {
		return err
	}
	_, err = fw.Write(data)
	return err
}

func xmlHeader() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"
}

// escapeXML escapes special XML characters.
func escapeXML(s string) string {
	var buf strings.Builder
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// fontStyleToXML writes font style (rPr) XML.
func fontStyleToXML(fs *style.FontStyle, defaultFont *style.FontStyle) string {
	if fs == nil && defaultFont == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("<w:rPr>")

	f := fs
	if f == nil {
		f = defaultFont
	}
	if f == nil {
		b.WriteString("</w:rPr>")
		return b.String()
	}

	if f.Name != "" {
		b.WriteString(fmt.Sprintf(`<w:rFonts w:ascii="%s" w:hAnsi="%s" w:cs="%s"`, escapeXML(f.Name), escapeXML(f.Name), escapeXML(f.Name)))
		if f.NameEastAsia != "" {
			b.WriteString(fmt.Sprintf(` w:eastAsia="%s"`, escapeXML(f.NameEastAsia)))
		}
		b.WriteString(`/>`)
	} else if f.NameEastAsia != "" {
		b.WriteString(fmt.Sprintf(`<w:rFonts w:eastAsia="%s"/>`, escapeXML(f.NameEastAsia)))
	}
	if f.Bold {
		b.WriteString(`<w:b/>`)
	}
	if f.Italic {
		b.WriteString(`<w:i/>`)
	}
	if f.Strikethrough {
		b.WriteString(`<w:strike/>`)
	}
	if f.DoubleStrikethrough {
		b.WriteString(`<w:dstrike/>`)
	}
	if f.Outline {
		b.WriteString(`<w:outline/>`)
	}
	if f.Shadow {
		b.WriteString(`<w:shadow/>`)
	}
	if f.Emboss {
		b.WriteString(`<w:emboss/>`)
	}
	if f.Imprint {
		b.WriteString(`<w:imprint/>`)
	}
	if f.Underline != "" && f.Underline != "none" {
		if f.UnderlineColor != "" {
			b.WriteString(fmt.Sprintf(`<w:u w:val="%s" w:color="%s"/>`, f.Underline, f.UnderlineColor))
		} else {
			b.WriteString(fmt.Sprintf(`<w:u w:val="%s"/>`, f.Underline))
		}
	}
	if f.TextEffect != "" {
		b.WriteString(fmt.Sprintf(`<w:effect w:val="%s"/>`, f.TextEffect))
	}
	if f.Color != "" {
		b.WriteString(fmt.Sprintf(`<w:color w:val="%s"/>`, f.Color))
	}
	if f.HighlightColor != "" {
		b.WriteString(fmt.Sprintf(`<w:highlight w:val="%s"/>`, f.HighlightColor))
	}
	if f.Size > 0 {
		hp := int(f.Size * 2)
		b.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/>`, hp))
		b.WriteString(fmt.Sprintf(`<w:szCs w:val="%d"/>`, hp))
	}
	if f.SuperScript {
		b.WriteString(`<w:vertAlign w:val="superscript"/>`)
	}
	if f.SubScript {
		b.WriteString(`<w:vertAlign w:val="subscript"/>`)
	}
	if f.AllCaps {
		b.WriteString(`<w:caps/>`)
	}
	if f.SmallCaps {
		b.WriteString(`<w:smallCaps/>`)
	}
	if f.Hidden {
		b.WriteString(`<w:vanish/>`)
	}
	if f.NoProof {
		b.WriteString(`<w:noProof/>`)
	}
	if f.Spacing != 0 {
		sp := int(f.Spacing * 20) // points to twips
		b.WriteString(fmt.Sprintf(`<w:spacing w:val="%d"/>`, sp))
	}
	if f.Kerning > 0 {
		hp := int(f.Kerning * 2) // points to half-points
		b.WriteString(fmt.Sprintf(`<w:kern w:val="%d"/>`, hp))
	}
	if f.RightToLeft {
		b.WriteString(`<w:rtl/>`)
	}
	if f.Lang != "" {
		b.WriteString(fmt.Sprintf(`<w:lang w:val="%s"/>`, f.Lang))
	}

	b.WriteString("</w:rPr>")
	return b.String()
}

// paraStyleToXML writes paragraph style (pPr) XML.
func paraStyleToXML(ps *style.ParagraphStyle, styleName string) string {
	if ps == nil && styleName == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("<w:pPr>")

	if styleName != "" {
		b.WriteString(fmt.Sprintf(`<w:pStyle w:val="%s"/>`, escapeXML(styleName)))
	}

	if ps != nil {
		if ps.Alignment != "" && ps.Alignment != style.AlignLeft {
			b.WriteString(fmt.Sprintf(`<w:jc w:val="%s"/>`, ps.Alignment))
		}
		if ps.SpaceBefore > 0 || ps.SpaceAfter > 0 || ps.LineSpacing > 0 {
			b.WriteString("<w:spacing")
			if ps.SpaceBefore > 0 {
				b.WriteString(fmt.Sprintf(` w:before="%d"`, ps.SpaceBefore))
			}
			if ps.SpaceAfter > 0 {
				b.WriteString(fmt.Sprintf(` w:after="%d"`, ps.SpaceAfter))
			}
			if ps.LineSpacing > 0 {
				rule := ps.LineRule
				if rule == "" {
					rule = "auto"
				}
				b.WriteString(fmt.Sprintf(` w:line="%d" w:lineRule="%s"`, ps.LineSpacing, rule))
			}
			b.WriteString("/>")
		}
		if ps.Indent > 0 || ps.IndentRight > 0 || ps.FirstLine > 0 || ps.Hanging > 0 {
			b.WriteString("<w:ind")
			if ps.Indent > 0 {
				b.WriteString(fmt.Sprintf(` w:left="%d"`, ps.Indent))
			}
			if ps.IndentRight > 0 {
				b.WriteString(fmt.Sprintf(` w:right="%d"`, ps.IndentRight))
			}
			if ps.FirstLine > 0 {
				b.WriteString(fmt.Sprintf(` w:firstLine="%d"`, ps.FirstLine))
			}
			if ps.Hanging > 0 {
				b.WriteString(fmt.Sprintf(` w:hanging="%d"`, ps.Hanging))
			}
			b.WriteString("/>")
		}
		if ps.KeepNext {
			b.WriteString(`<w:keepNext/>`)
		}
		if ps.KeepLines {
			b.WriteString(`<w:keepLines/>`)
		}
		if ps.PageBreakBefore {
			b.WriteString(`<w:pageBreakBefore/>`)
		}
		if ps.WidowControl {
			b.WriteString(`<w:widowControl/>`)
		}
		if ps.NumStyleName != "" {
			b.WriteString("<w:numPr>")
			b.WriteString(fmt.Sprintf(`<w:ilvl w:val="%d"/>`, ps.NumLevel))
			b.WriteString(fmt.Sprintf(`<w:numId w:val="%s"/>`, ps.NumStyleName))
			b.WriteString("</w:numPr>")
		}
		if ps.Shading != nil {
			b.WriteString(fmt.Sprintf(`<w:shd w:val="%s" w:color="%s" w:fill="%s"/>`,
				ps.Shading.Pattern, ps.Shading.Color, ps.Shading.Fill))
		}
		if len(ps.TabStops) > 0 {
			b.WriteString("<w:tabs>")
			for _, ts := range ps.TabStops {
				tabType := ts.Type
				if tabType == "" {
					tabType = "left"
				}
				leader := ts.Leader
				if leader == "" {
					leader = "none"
				}
				b.WriteString(fmt.Sprintf(`<w:tab w:val="%s" w:pos="%d" w:leader="%s"/>`, tabType, ts.Position, leader))
			}
			b.WriteString("</w:tabs>")
		}
		writeBorders(&b, ps.Borders)
	}

	b.WriteString("</w:pPr>")
	return b.String()
}

func writeBorders(b *strings.Builder, borders *style.ParagraphBorders) {
	if borders == nil {
		return
	}
	b.WriteString("<w:pBdr>")
	writeBorder(b, "top", borders.Top)
	writeBorder(b, "bottom", borders.Bottom)
	writeBorder(b, "left", borders.Left)
	writeBorder(b, "right", borders.Right)
	b.WriteString("</w:pBdr>")
}

func writeBorder(b *strings.Builder, name string, border *style.Border) {
	if border == nil {
		return
	}
	b.WriteString(fmt.Sprintf(`<w:%s w:val="%s" w:sz="%d" w:space="%d" w:color="%s"/>`,
		name, border.Style, border.Size, border.Space, border.Color))
}

// imageExtension returns the file extension for an image MIME type.
func imageExtension(mimeType string) string {
	switch mimeType {
	case "image/png":
		return "png"
	case "image/jpeg", "image/jpg":
		return "jpeg"
	case "image/gif":
		return "gif"
	case "image/bmp":
		return "bmp"
	case "image/tiff":
		return "tiff"
	default:
		return "png"
	}
}

// imageContentType returns the content type for an image extension.
func imageContentType(ext string) string {
	switch ext {
	case "png":
		return "image/png"
	case "jpeg", "jpg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "bmp":
		return "image/bmp"
	case "tiff", "tif":
		return "image/tiff"
	default:
		return "image/png"
	}
}
