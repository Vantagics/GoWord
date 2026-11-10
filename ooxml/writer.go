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
	"sort"
	"strconv"
	"strings"

	"github.com/VantageDataChat/GoWord/document"
	"github.com/VantageDataChat/GoWord/style"
)

// Namespaces used in OOXML documents.
const (
	nsW   = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	nsR   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	nsWP  = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
	nsA   = "http://schemas.openxmlformats.org/drawingml/2006/main"
	nsPIC = "http://schemas.openxmlformats.org/drawingml/2006/picture"
	nsMC  = "http://schemas.openxmlformats.org/markup-compatibility/2006"
	nsCP  = "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
	nsDC  = "http://purl.org/dc/elements/1.1/"
	nsDCT = "http://purl.org/dc/terms/"
	nsXSI = "http://www.w3.org/2001/XMLSchema-instance"
	nsCT  = "http://schemas.openxmlformats.org/package/2006/content-types"
	nsEP  = "http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"
	nsVT  = "http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"

	relDoc         = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	relCoreProps   = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties"
	relExtProps    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties"
	relStyles      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"
	relNumbering   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering"
	relFontTable   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable"
	relSettings    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings"
	relImage       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
	relHyperlink   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink"
	relHeader      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/header"
	relFooter      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer"
	relFootnotes   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footnotes"
	relEndnotes    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/endnotes"
	relComments    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/comments"
	relTheme       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme"
	relWebSettings = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/webSettings"

	nsM    = "http://schemas.openxmlformats.org/officeDocument/2006/math"
	nsO    = "urn:schemas-microsoft-com:office:office"
	nsV    = "urn:schemas-microsoft-com:vml"
	nsW10  = "urn:schemas-microsoft-com:office:word"
	nsW14  = "http://schemas.microsoft.com/office/word/2010/wordml"
	nsW15  = "http://schemas.microsoft.com/office/word/2012/wordml"
	nsWP14 = "http://schemas.microsoft.com/office/word/2010/wordprocessingDrawing"
	nsWPC  = "http://schemas.microsoft.com/office/word/2010/wordprocessingCanvas"
	nsWPG  = "http://schemas.microsoft.com/office/word/2010/wordprocessingGroup"
	nsWPI  = "http://schemas.microsoft.com/office/word/2010/wordprocessingInk"
	nsWNE  = "http://schemas.microsoft.com/office/word/2006/wordml"
	nsWPS  = "http://schemas.microsoft.com/office/word/2010/wordprocessingShape"
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

	// Buffered package parts. addZipFile overwrites an existing name so a
	// regenerated part replaces the copy taken from the original package.
	files     []memFile
	fileIndex map[string]int

	// cur, when non-nil, receives relationships for the part being written
	// (a header or footer) instead of document.xml.rels.
	cur *relCollector
}

type memFile struct {
	name string
	data []byte
}

type relCollector struct {
	rels []relationship
	next int
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
	filename  string
	content   []byte
	header    *document.Header // non-nil for header parts
	footer    *document.Footer // non-nil for footer parts
	rels      []relationship
	generated bool
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
		fileIndex:    map[string]int{},
	}

	if err := w.write(); err != nil {
		return nil, err
	}

	zw := zip.NewWriter(w.buf)
	for pass := 0; pass < 2; pass++ {
		for _, f := range w.files {
			isCT := f.name == "[Content_Types].xml"
			if pass == 0 && !isCT || pass == 1 && isCT {
				continue
			}
			fw, err := zw.Create(f.name)
			if err != nil {
				return nil, err
			}
			if _, err := fw.Write(f.data); err != nil {
				return nil, err
			}
		}
	}
	if err := zw.Close(); err != nil {
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
	w.copyOriginalParts()
	w.seedDocRels()

	// Pre-generate header/footer parts
	w.generateHeaderFooterParts()

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
	// 12. word/fontTable.xml when the package does not already have one
	if err := w.writeFontTable(); err != nil {
		return err
	}
	// 13. Header/footer parts
	if err := w.writeHeaderFooterFiles(); err != nil {
		return err
	}
	if err := w.writeHeaderFooterRels(); err != nil {
		return err
	}
	// 14. word/document.xml (main content)
	if err := w.writeDocument(); err != nil {
		return err
	}
	// 15. word/_rels/document.xml.rels
	if err := w.writeDocRels(); err != nil {
		return err
	}
	// Content types last, once every part is known.
	if err := w.writeContentTypes(); err != nil {
		return err
	}

	return nil
}

func (w *Writer) addDocRel(relType, target string) string {
	return w.addRel(relType, target, "")
}

func (w *Writer) addDocRelExternal(relType, target string) string {
	return w.addRel(relType, target, "External")
}

func (w *Writer) addRel(relType, target, mode string) string {
	relType = w.relType(relType)
	if w.cur != nil {
		w.cur.next++
		id := fmt.Sprintf("rId%d", w.cur.next)
		w.cur.rels = append(w.cur.rels, relationship{ID: id, Type: relType, Target: target, TargetMode: mode})
		return id
	}
	id := fmt.Sprintf("rId%d", w.nextDocRelID)
	w.nextDocRelID++
	w.docRels = append(w.docRels, relationship{ID: id, Type: relType, Target: target, TargetMode: mode})
	return id
}

// ensureRel reuses a document relationship with the same type and target so
// identifiers loaded from an existing package stay valid.
func (w *Writer) ensureRel(relType, target string) string {
	relType = w.relType(relType)
	for _, rel := range w.docRels {
		if rel.Type == relType && rel.Target == target && rel.TargetMode == "" {
			return rel.ID
		}
	}
	return w.addDocRel(relType, target)
}

func (w *Writer) addZipFile(name string, data []byte) error {
	if w.fileIndex == nil {
		w.fileIndex = map[string]int{}
	}
	cp := append([]byte(nil), data...)
	if i, ok := w.fileIndex[name]; ok {
		w.files[i].data = cp
		return nil
	}
	w.fileIndex[name] = len(w.files)
	w.files = append(w.files, memFile{name: name, data: cp})
	return nil
}

func (w *Writer) hasFile(name string) bool {
	if w.fileIndex == nil {
		return false
	}
	_, ok := w.fileIndex[name]
	return ok
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
// Child order follows ECMA-376 CT_RPr / EG_RPrBase. Word repairs documents
// whose rPr children are out of that sequence.
func fontStyleToXML(fs *style.FontStyle, defaultFont *style.FontStyle) string {
	if fs == nil && defaultFont == nil {
		return ""
	}
	f := fs
	if f == nil {
		f = defaultFont
	}
	if f == nil {
		return ""
	}

	var inner strings.Builder
	if f.Name != "" || f.NameEastAsia != "" {
		inner.WriteString(`<w:rFonts`)
		if f.Name != "" {
			inner.WriteString(fmt.Sprintf(` w:ascii="%s" w:hAnsi="%s"`, escapeXML(f.Name), escapeXML(f.Name)))
		}
		if f.NameEastAsia != "" {
			inner.WriteString(fmt.Sprintf(` w:eastAsia="%s"`, escapeXML(f.NameEastAsia)))
		}
		if f.Name != "" {
			inner.WriteString(fmt.Sprintf(` w:cs="%s"`, escapeXML(f.Name)))
		}
		inner.WriteString(`/>`)
	}
	writeToggle(&inner, f.Bold, f.Off, "b")
	if f.Bold {
		inner.WriteString(`<w:bCs/>`)
	} else {
		writeToggle(&inner, false, f.Off, "bCs")
	}
	writeToggle(&inner, f.Italic, f.Off, "i")
	if f.Italic {
		inner.WriteString(`<w:iCs/>`)
	} else {
		writeToggle(&inner, false, f.Off, "iCs")
	}
	writeToggle(&inner, f.AllCaps, f.Off, "caps")
	writeToggle(&inner, f.SmallCaps, f.Off, "smallCaps")
	writeToggle(&inner, f.Strikethrough, f.Off, "strike")
	writeToggle(&inner, f.DoubleStrikethrough, f.Off, "dstrike")
	writeToggle(&inner, f.Outline, f.Off, "outline")
	writeToggle(&inner, f.Shadow, f.Off, "shadow")
	writeToggle(&inner, f.Emboss, f.Off, "emboss")
	writeToggle(&inner, f.Imprint, f.Off, "imprint")
	writeToggle(&inner, f.NoProof, f.Off, "noProof")
	writeToggle(&inner, f.Hidden, f.Off, "vanish")
	if f.Color != "" {
		inner.WriteString(fmt.Sprintf(`<w:color w:val="%s"/>`, f.Color))
	}
	if f.Spacing != 0 {
		sp := int(f.Spacing * 20) // points to twips
		inner.WriteString(fmt.Sprintf(`<w:spacing w:val="%d"/>`, sp))
	}
	if f.Kerning > 0 {
		hp := int(f.Kerning * 2) // points to half-points
		inner.WriteString(fmt.Sprintf(`<w:kern w:val="%d"/>`, hp))
	}
	if f.Size > 0 {
		hp := int(f.Size * 2)
		inner.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/>`, hp))
		inner.WriteString(fmt.Sprintf(`<w:szCs w:val="%d"/>`, hp))
	}
	if f.HighlightColor != "" {
		inner.WriteString(fmt.Sprintf(`<w:highlight w:val="%s"/>`, f.HighlightColor))
	}
	if f.Underline != "" && f.Underline != "none" {
		if f.UnderlineColor != "" {
			inner.WriteString(fmt.Sprintf(`<w:u w:val="%s" w:color="%s"/>`, f.Underline, f.UnderlineColor))
		} else {
			inner.WriteString(fmt.Sprintf(`<w:u w:val="%s"/>`, f.Underline))
		}
	}
	if f.TextEffect != "" {
		inner.WriteString(fmt.Sprintf(`<w:effect w:val="%s"/>`, f.TextEffect))
	}
	if f.SuperScript {
		inner.WriteString(`<w:vertAlign w:val="superscript"/>`)
	} else if f.SubScript {
		inner.WriteString(`<w:vertAlign w:val="subscript"/>`)
	}
	writeToggle(&inner, f.RightToLeft, f.Off, "rtl")
	if f.Lang != "" {
		inner.WriteString(fmt.Sprintf(`<w:lang w:val="%s"/>`, f.Lang))
	}
	if inner.Len() == 0 {
		return ""
	}
	return "<w:rPr>" + inner.String() + "</w:rPr>"
}

func writeToggle(b *strings.Builder, on bool, off map[string]bool, tag string) {
	if on {
		b.WriteString("<w:" + tag + "/>")
		return
	}
	if off != nil && off[tag] {
		b.WriteString(`<w:` + tag + ` w:val="0"/>`)
	}
}

// paraStyleToXML writes paragraph style (pPr) XML.
// Child order follows ECMA-376 CT_PPrBase. numID, when non-empty, is the
// numeric numbering id; otherwise a numeric NumStyleName is used as-is.
func paraStyleToXML(ps *style.ParagraphStyle, styleName string, numID ...string) string {
	if ps == nil && styleName == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("<w:pPr>")

	if styleName != "" {
		b.WriteString(fmt.Sprintf(`<w:pStyle w:val="%s"/>`, escapeXML(styleName)))
	}

	if ps != nil {
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
		id := ""
		if len(numID) > 0 {
			id = numID[0]
		} else if isDigits(ps.NumStyleName) {
			id = ps.NumStyleName
		}
		if id != "" {
			b.WriteString("<w:numPr>")
			b.WriteString(fmt.Sprintf(`<w:ilvl w:val="%d"/>`, ps.NumLevel))
			b.WriteString(fmt.Sprintf(`<w:numId w:val="%s"/>`, id))
			b.WriteString("</w:numPr>")
		}
		writeBorders(&b, ps.Borders)
		if ps.Shading != nil {
			pattern := ps.Shading.Pattern
			if pattern == "" {
				pattern = "clear"
			}
			color := ps.Shading.Color
			if color == "" {
				color = "auto"
			}
			fill := ps.Shading.Fill
			if fill == "" {
				fill = "auto"
			}
			b.WriteString(fmt.Sprintf(`<w:shd w:val="%s" w:color="%s" w:fill="%s"/>`, pattern, color, fill))
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
		if ps.Alignment != "" && ps.Alignment != style.AlignLeft {
			b.WriteString(fmt.Sprintf(`<w:jc w:val="%s"/>`, ps.Alignment))
		}
	}

	if b.Len() == len("<w:pPr>") {
		return ""
	}
	b.WriteString("</w:pPr>")
	return b.String()
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func writeBorders(b *strings.Builder, borders *style.ParagraphBorders) {
	if borders == nil {
		return
	}
	b.WriteString("<w:pBdr>")
	// CT_PBdr sequence: top, left, bottom, right.
	writeBorder(b, "top", borders.Top)
	writeBorder(b, "left", borders.Left)
	writeBorder(b, "bottom", borders.Bottom)
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

// numID resolves a numbering style name to the numeric w:numId used in numbering.xml.
// Numeric names are returned unchanged so identifiers loaded from Word stay stable.
func (w *Writer) numID(name string) string {
	if name == "" {
		return ""
	}
	if isDigits(name) {
		return name
	}
	ns := w.doc.NumberingStyles()
	names := make([]string, 0, len(ns))
	for n := range ns {
		names = append(names, n)
	}
	sort.Strings(names)
	for i, n := range names {
		if n == name {
			return strconv.Itoa(i + 1)
		}
	}
	return "1"
}

func (w *Writer) strict() bool {
	return w != nil && w.doc != nil && w.doc.Strict()
}

func (w *Writer) wNS() string {
	if w.strict() {
		return "http://purl.oclc.org/ooxml/wordprocessingml/main"
	}
	return nsW
}

func (w *Writer) rNS() string {
	if w.strict() {
		return "http://purl.oclc.org/ooxml/officeDocument/relationships"
	}
	return nsR
}

func (w *Writer) mNS() string {
	if w.strict() {
		return "http://purl.oclc.org/ooxml/officeDocument/math"
	}
	return nsM
}

func (w *Writer) wpNS() string {
	if w.strict() {
		return "http://purl.oclc.org/ooxml/drawingml/wordprocessingDrawing"
	}
	return nsWP
}

// relType rewrites a transitional officeDocument relationship URI when the
// document asked for a strict package. Package relationships are unchanged.
func (w *Writer) relType(relType string) string {
	if !w.strict() {
		return relType
	}
	const from = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	const to = "http://purl.oclc.org/ooxml/officeDocument/relationships"
	return strings.Replace(relType, from, to, 1)
}

// wordMLRoot opens a WordprocessingML root element with the namespace set Word
// writes, including mc:Ignorable for the extension prefixes.
func (w *Writer) wordMLRoot(tag string) string {
	return fmt.Sprintf(`<%s xmlns:wpc="%s" xmlns:mc="%s" xmlns:o="%s" xmlns:r="%s" xmlns:m="%s" xmlns:v="%s" xmlns:wp14="%s" xmlns:wp="%s" xmlns:w10="%s" xmlns:w="%s" xmlns:w14="%s" xmlns:w15="%s" xmlns:wpg="%s" xmlns:wpi="%s" xmlns:wne="%s" xmlns:wps="%s" xmlns:a="%s" xmlns:pic="%s" mc:Ignorable="w14 w15 wp14">`,
		tag, nsWPC, nsMC, nsO, w.rNS(), w.mNS(), nsV, nsWP14, w.wpNS(), nsW10, w.wNS(), nsW14, nsW15, nsWPG, nsWPI, nsWNE, nsWPS, nsA, nsPIC)
}

func (w *Writer) imageFileName(img *document.Image) string {
	if img.Name != "" {
		base := filepath.Base(img.Name)
		if base != "" && base != "." && strings.Contains(base, ".") {
			img.Name = base
			return base
		}
	}
	ext := "png"
	if img.MimeType != "" {
		ext = imageExtension(img.MimeType)
	} else if img.Source != "" {
		ext = strings.TrimPrefix(strings.ToLower(filepath.Ext(img.Source)), ".")
		if ext == "jpg" {
			ext = "jpeg"
		}
		if ext == "" {
			ext = "png"
		}
	}
	if img.ID <= 0 {
		img.ID = w.doc.AllocImageIDPublic()
	}
	img.Name = fmt.Sprintf("image%d.%s", img.ID, ext)
	return img.Name
}

func (w *Writer) imageRel(img *document.Image) string {
	name := w.imageFileName(img)
	target := "media/" + name
	if w.cur != nil {
		return w.addDocRel(relImage, target)
	}
	id := w.ensureRel(relImage, target)
	img.RelID = id
	return id
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
