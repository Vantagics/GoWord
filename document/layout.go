package document

import (
	"strconv"
	"strings"

	"github.com/VantageDataChat/GoWord/common"
	"github.com/VantageDataChat/GoWord/style"
)

// Layout is an estimate of where body content falls. PageCount is the number
// of sheets. PageOf is the page number Word would print, which follows a
// section's page-number restart. SheetOf is the physical sheet, which is what
// a PDF export counts. The estimate uses section size, margins, columns,
// paragraph and font styles, and average character widths. It is not Word's
// line breaking.
type Layout struct {
	PageCount    int
	paraPage     map[*Paragraph]int
	paraSheet    map[*Paragraph]int
	bookmarkPage map[string]int
	bookmarkText map[string]string
}

// PageOf returns the displayed page number for a paragraph.
func (l Layout) PageOf(p *Paragraph) int {
	if n, ok := l.paraPage[p]; ok && n > 0 {
		return n
	}
	return 1
}

// SheetOf returns the 1-based physical sheet for a paragraph.
func (l Layout) SheetOf(p *Paragraph) int {
	if n, ok := l.paraSheet[p]; ok && n > 0 {
		return n
	}
	return 1
}

// BookmarkPage returns the displayed page of a bookmark.
func (l Layout) BookmarkPage(name string) (int, bool) {
	n, ok := l.bookmarkPage[name]
	return n, ok && n > 0
}

// BookmarkText returns the text collected for a bookmark.
func (l Layout) BookmarkText(name string) (string, bool) {
	s, ok := l.bookmarkText[name]
	return s, ok
}

// EstimateLayout assigns a page number to each body paragraph.
func (d *Document) EstimateLayout() Layout {
	lay := Layout{
		PageCount:    1,
		paraPage:     map[*Paragraph]int{},
		paraSheet:    map[*Paragraph]int{},
		bookmarkPage: map[string]int{},
		bookmarkText: map[string]string{},
	}
	if d == nil {
		return lay
	}
	cur := newPageCursor(styleOrDefault(nil))
	for si, sec := range d.Sections {
		st := styleOrDefault(sec)
		if si > 0 && sectionStartsPage(st.breakType) {
			cur.breakPage(st.breakType)
		}
		if st.pageStart != nil && *st.pageStart > 0 {
			cur.page = *st.pageStart
		}
		cur.applySection(st)
		cur.reserveChrome(sec)
		cur.lastKeep = false
		for i, el := range sec.Elements {
			if cur.skipTrailingSpacer(d, sec.Elements, i, &lay) {
				continue
			}
			cur.honorKeep(d, sec.Elements, i)
			cur.place(d, el, &lay)
			if p, ok := el.(*Paragraph); ok {
				cur.lastKeep = d.metricsFor(p).keepNext
			} else {
				cur.lastKeep = false
			}
		}
	}
	cur.placeNotes(d, &lay)
	if cur.physical < 1 {
		cur.physical = 1
	}
	lay.PageCount = cur.physical
	return lay
}

func (c *pageCursor) placeNotes(doc *Document, lay *Layout) {
	if doc == nil {
		return
	}
	if len(doc.Footnotes()) == 0 && len(doc.Endnotes()) == 0 {
		return
	}
	c.addHeight(c.line)
	for _, fn := range doc.Footnotes() {
		for _, el := range fn.Elements {
			c.place(doc, el, lay)
		}
	}
	for _, en := range doc.Endnotes() {
		for _, el := range en.Elements {
			c.place(doc, el, lay)
		}
	}
}

type styleSection struct {
	width, height           int
	marginTop, marginBottom int
	marginLeft, marginRight int
	columns                 int
	colGap                  int
	gutter                  int
	linePitch               int
	breakType               string
	pageStart               *int
}

func styleOrDefault(sec *Section) styleSection {
	if sec == nil {
		return defaultPage()
	}
	st := sec.Style
	out := defaultPage()
	if st.PageWidth > 0 {
		out.width = st.PageWidth
	}
	if st.PageHeight > 0 {
		out.height = st.PageHeight
	}
	if st.MarginTop > 0 {
		out.marginTop = st.MarginTop
	}
	if st.MarginBottom > 0 {
		out.marginBottom = st.MarginBottom
	}
	if st.MarginLeft > 0 {
		out.marginLeft = st.MarginLeft
	}
	if st.MarginRight > 0 {
		out.marginRight = st.MarginRight
	}
	if st.ColumnCount > 1 {
		out.columns = st.ColumnCount
	}
	out.colGap = st.ColumnSpacing
	out.gutter = st.Gutter
	out.linePitch = st.LinePitch
	out.breakType = st.BreakType
	out.pageStart = st.PageNumberStart
	return out
}

func defaultPage() styleSection {
	return styleSection{
		width: 11906, height: 16838,
		marginTop: 1440, marginBottom: 1440,
		marginLeft: 1440, marginRight: 1440,
		columns: 1,
	}
}

func sectionStartsPage(kind string) bool {
	return kind != "continuous"
}

type pageCursor struct {
	physical, page int
	y, col         int
	usableH        int
	colWidth       int
	columns        int
	line           int
	pitch          int
	pendingAfter   int
	lastKeep       bool
	inTable        bool
	justBroke      bool
	placed         bool
}

func newPageCursor(st styleSection) *pageCursor {
	c := &pageCursor{physical: 1, page: 1, columns: 1, line: 276}
	c.applySection(st)
	return c
}

func (c *pageCursor) applySection(st styleSection) {
	h := st.height - st.marginTop - st.marginBottom
	w := st.width - st.marginLeft - st.marginRight - st.gutter
	// A missing page size can compute a tiny or negative body. A real short
	// page, such as a three-inch landscape section, keeps the height it declared.
	if h < 400 {
		h = 2000
	}
	if w < 400 {
		w = 2000
	}
	cols := st.columns
	if cols < 1 {
		cols = 1
	}
	gap := st.colGap
	if cols > 1 && gap <= 0 {
		gap = 720
	}
	colW := w
	if cols > 1 {
		colW = (w - gap*(cols-1)) / cols
		if colW < 800 {
			colW = 800
		}
	}
	c.usableH = h
	c.colWidth = colW
	c.columns = cols
	c.pitch = st.linePitch
	c.line = 276
}

func (c *pageCursor) advanceSheet() {
	c.physical++
	c.page++
	c.y = 0
	c.col = 0
	c.pendingAfter = 0
}

func (c *pageCursor) advanceColumn() {
	c.col++
	c.y = 0
	c.pendingAfter = 0
	if c.col >= c.columns {
		c.advanceSheet()
	}
}

func (c *pageCursor) breakPage(kind string) {
	if c.y > 0 || c.col > 0 {
		c.advanceSheet()
	}
	switch kind {
	case "oddPage":
		if c.page%2 == 0 {
			c.advanceSheet()
		}
	case "evenPage":
		if c.page%2 == 1 {
			c.advanceSheet()
		}
	}
}

func (c *pageCursor) addHeight(h int) (displayed, sheet int) {
	if h < 120 {
		h = 120
	}
	if c.y > 0 && c.y+h > c.usableH {
		c.advanceColumn()
	}
	displayed, sheet = c.page, c.physical
	c.justBroke = false
	c.placed = true
	c.y += h
	for c.y > c.usableH {
		over := c.y - c.usableH
		c.advanceColumn()
		c.y = over
	}
	return displayed, sheet
}

func (c *pageCursor) place(doc *Document, el Element, lay *Layout) {
	switch e := el.(type) {
	case *PageBreak:
		c.advanceSheet()
	case *Paragraph:
		c.placeParagraph(doc, e, lay)
	case *Table:
		c.placeTable(doc, e, lay)
	case *TextBreak:
		n := e.Count
		if n < 1 {
			n = 1
		}
		c.addHeight(n * c.line)
	case *Image:
		h := imageTwips(e)
		if h <= 0 {
			h = 40 * 15
		}
		c.addHeight(h)
	case *RawXML:
		h := c.textBlockHeight(e.Text, doc.metricsFor(nil))
		if dh := drawingTwips(e.XML); dh > h {
			h = dh
		}
		if n := strings.Count(e.XML, "<w:p"); n > 1 {
			block := n * c.line
			if block > h {
				h = block
			}
		}
		c.addHeight(h)
	case *Sdt:
		for _, child := range e.Elements {
			c.place(doc, child, lay)
		}
	case *ListItem:
		c.addHeight(c.textBlockHeight(e.Text, doc.metricsFor(nil)))
	case *TOC:
		lines := 1
		if e.Result != "" {
			lines = strings.Count(e.Result, "\n") + 1
		}
		c.addHeight(lines * c.line)
	default:
		c.addHeight(c.line)
	}
}

// skipTrailingSpacer drops up to three blank lines that would otherwise
// open a sheet after the real story. A long run of blank lines is content
// (a section of empty paragraphs) and is paginated normally.
func (c *pageCursor) skipTrailingSpacer(doc *Document, elems []Element, i int, lay *Layout) bool {
	p, ok := elems[i].(*Paragraph)
	if !ok || !p.spacer || doc == nil {
		return false
	}
	after := 0
	for _, el := range elems[i+1:] {
		np, ok := el.(*Paragraph)
		if !ok || !np.spacer {
			return false
		}
		after++
		if after > 2 {
			return false
		}
	}
	m := doc.metricsFor(p)
	h := c.flowHeight(doc, p, m)
	if c.y == 0 || c.y+h <= c.usableH {
		return false
	}
	lay.paraPage[p] = c.page
	lay.paraSheet[p] = c.physical
	return true
}

// honorKeep applies Word's keep-with-next, keep-lines, and widow/orphan
// rules before the paragraph is placed. A keep chain taller than a page
// breaks, which is why a list of headings that all set keepNext still
// paginates one heading at a time.
func (c *pageCursor) honorKeep(doc *Document, elems []Element, i int) {
	p, ok := elems[i].(*Paragraph)
	if !ok || doc == nil {
		return
	}
	m := doc.metricsFor(p)
	if m.pageBreak && !c.lastKeep && (c.y > 0 || c.col > 0) {
		c.advanceSheet()
	}
	h := c.flowHeight(doc, p, m)
	if m.keepLines && h > 0 && h <= c.usableH && c.y > 0 && c.y+h > c.usableH {
		// Keep lines together: a paragraph that fits on a fresh sheet moves
		// as a unit instead of splitting.
		c.advanceColumn()
		return
	}
	if m.widow && c.y > 0 {
		lh := c.lineBox(m)
		if lh < 1 {
			lh = 1
		}
		lines := h / lh
		remain := c.usableH - c.y
		if lines >= 2 && remain < 2*lh && remain < h && h <= c.usableH {
			c.advanceColumn()
			return
		}
	}
	if !m.keepNext || h > c.usableH {
		return
	}
	chain := h
	for j := i + 1; j < len(elems); j++ {
		np, ok := elems[j].(*Paragraph)
		if !ok {
			break
		}
		nm := doc.metricsFor(np)
		chain += c.flowHeight(doc, np, nm)
		if chain > c.usableH {
			// Longer than a page: break the chain and flow normally.
			return
		}
		if !nm.keepNext {
			break
		}
	}
	if c.y > 0 && c.y+chain > c.usableH {
		c.advanceColumn()
	}
}

func (c *pageCursor) flowHeight(doc *Document, p *Paragraph, m paraMetrics) int {
	if pageSizedAnchor(p) {
		return c.lineBox(m)
	}
	h := c.paragraphBodyHeight(p, m)
	if ih := inlineExtra(p); ih > h && strings.TrimSpace(paragraphPlain(p)) == "" {
		h = ih
	} else if ih > 0 {
		h += ih
	}
	if m.before > 0 {
		h += m.before
	}
	return h
}

func (c *pageCursor) placeParagraph(doc *Document, p *Paragraph, lay *Layout) {
	if paragraphIsPageBreak(p) {
		// After content that already ended on a sheet boundary, LibreOffice
		// does not add another blank sheet. Two breaks in a row still do.
		if !(c.placed && c.y == 0 && c.col == 0 && !c.justBroke) {
			c.advanceSheet()
			c.justBroke = true
		}
		c.noteBookmarks(p, lay)
		lay.paraPage[p] = c.page
		lay.paraSheet[p] = c.physical
		return
	}
	m := doc.metricsFor(p)
	before := m.before
	// Drop a small space-before at the top of a page. A heading gap of
	// half an inch or more still counts, which is what keeps a long list of
	// spaced headings from packing an extra line onto every sheet.
	if c.y == 0 && before < 200 {
		before = 0
	} else if c.pendingAfter > before {
		before = c.pendingAfter
	}
	c.pendingAfter = 0
	h := c.paragraphBodyHeight(p, m)
	if ih := inlineExtra(p); ih > 0 {
		if strings.TrimSpace(paragraphPlain(p)) == "" {
			if ih > h {
				h = ih
			}
		} else {
			h += ih
		}
	}
	if before > 0 {
		h += before
	}
	if pageSizedAnchor(p) {
		// The story lives in the floating frame, which shares the page.
		h = c.lineBox(m)
		if before > 0 && before < h {
			h += before
		}
	}
	displayed, sheet := c.addHeight(h)
	c.pendingAfter = m.after
	lay.paraPage[p] = displayed
	lay.paraSheet[p] = sheet
	savedPage, savedSheet := c.page, c.physical
	c.page, c.physical = displayed, sheet
	c.noteBookmarks(p, lay)
	c.page, c.physical = savedPage, savedSheet
}

func (c *pageCursor) placeTable(doc *Document, t *Table, lay *Layout) {
	fallback := c.colWidth
	if n := len(t.Rows); n > 0 && len(t.Rows[0].Cells) > 0 {
		fallback = c.colWidth / len(t.Rows[0].Cells)
		if fallback < 400 {
			fallback = 400
		}
	}
	headerH := 0
	for _, row := range t.Rows {
		rh := c.rowHeight(doc, t, row, fallback)
		if row.Style.IsHeader {
			headerH = rh
		}
		if !row.Style.IsHeader && headerH > 0 && c.y > 0 && c.y+rh > c.usableH {
			c.advanceColumn()
			c.addHeight(headerH)
		}
		displayed, sheet := c.addHeight(rh)
		c.stampTableRow(doc, row, displayed, sheet, lay)
	}
}

func (c *pageCursor) rowHeight(doc *Document, t *Table, row *Row, fallback int) int {
	prev := c.inTable
	c.inTable = true
	defer func() { c.inTable = prev }()
	content := 0
	for _, cell := range row.Cells {
		w := cellWidth(t, cell, fallback)
		saved := c.colWidth
		c.colWidth = w
		ch := 0
		for _, el := range cell.Elements {
			ch += c.blockHeight(doc, el)
		}
		c.colWidth = saved
		if ch > content {
			content = ch
		}
	}
	if content < 120 {
		content = 120
	}
	rh := content
	switch row.Style.HeightRule {
	case "exact":
		if row.Style.Height > 0 {
			rh = row.Style.Height
		}
	default:
		if row.Style.Height > rh {
			rh = row.Style.Height
		}
	}
	return rh
}

func (c *pageCursor) blockHeight(doc *Document, el Element) int {
	switch e := el.(type) {
	case *Paragraph:
		m := doc.metricsFor(e)
		h := c.paragraphBodyHeight(e, m)
		if ih := inlineExtra(e); ih > h && strings.TrimSpace(paragraphPlain(e)) == "" {
			h = ih
		} else if ih > 0 && strings.TrimSpace(paragraphPlain(e)) != "" {
			h += ih
		}
		return h
	case *Table:
		sum := 0
		fallback := c.colWidth
		if len(e.Rows) > 0 && len(e.Rows[0].Cells) > 0 {
			fallback = c.colWidth / len(e.Rows[0].Cells)
		}
		for _, row := range e.Rows {
			sum += c.rowHeight(doc, e, row, fallback)
		}
		return sum
	case *Sdt:
		h := 0
		for _, ch := range e.Elements {
			h += c.blockHeight(doc, ch)
		}
		return h
	case *Image:
		if h := imageTwips(e); h > 0 {
			return h
		}
		return c.line
	case *RawXML:
		h := drawingTwips(e.XML)
		if th := c.textBlockHeight(e.Text, doc.metricsFor(nil)); th > h {
			h = th
		}
		return h
	default:
		return 0
	}
}

func (c *pageCursor) stampTableRow(doc *Document, row *Row, displayed, sheet int, lay *Layout) {
	var walk func([]Element)
	walk = func(elems []Element) {
		for _, el := range elems {
			switch e := el.(type) {
			case *Paragraph:
				lay.paraPage[e] = displayed
				lay.paraSheet[e] = sheet
				savedPage, savedSheet := c.page, c.physical
				c.page, c.physical = displayed, sheet
				c.noteBookmarks(e, lay)
				c.page, c.physical = savedPage, savedSheet
			case *Table:
				for _, r := range e.Rows {
					for _, cell := range r.Cells {
						walk(cell.Elements)
					}
				}
			case *Sdt:
				walk(e.Elements)
			}
		}
	}
	for _, cell := range row.Cells {
		walk(cell.Elements)
	}
}

func paragraphIsPageBreak(p *Paragraph) bool {
	if p == nil || len(p.Runs) == 0 {
		return false
	}
	saw := false
	for _, r := range p.Runs {
		if r.Break && (r.BreakType == "page" || r.BreakType == "column") {
			saw = true
			continue
		}
		if r.Text != "" || r.Tab {
			return false
		}
	}
	return saw
}

type paraMetrics struct {
	before, after int
	line          int
	rule          string
	indent        int
	first         int
	hanging       int
	size          float64
	face          string
	east          string
	csSize        float64
	csFace        string
	bold          bool
	pageBreak     bool
	keepNext      bool
	keepLines     bool
	widow         bool
}

func (d *Document) metricsFor(p *Paragraph) paraMetrics {
	m := paraMetrics{size: 11, face: "Calibri", widow: true}
	if d == nil {
		return m
	}
	if d.DefaultFont.Size > 0 {
		m.size = d.DefaultFont.Size
	}
	if d.DefaultFont.Name != "" {
		m.face = d.DefaultFont.Name
	}
	// New documents start as Arial. A loaded file that never named a latin
	// face still has that default, so a theme face replaces it.
	if d.themeMinor != "" && (d.DefaultFont.Name == "" || d.DefaultFont.Name == "Arial") {
		m.face = d.themeMinor
	}
	m.east = d.DefaultFont.NameEastAsia
	if m.east == "" {
		m.east = d.themeMinorEA
	}
	apply := func(ps style.ParagraphStyle, fs style.FontStyle, heading bool) {
		if ps.SpaceBefore > 0 {
			m.before = ps.SpaceBefore
		}
		if ps.SpaceAfter > 0 {
			m.after = ps.SpaceAfter
		}
		if ps.LineSpacing > 0 {
			m.line = ps.LineSpacing
			m.rule = ps.LineRule
		} else if ps.LineRule != "" {
			m.rule = ps.LineRule
		}
		if ps.Indent > 0 {
			m.indent = ps.Indent
		}
		if ps.IndentRight > 0 {
			m.indent += ps.IndentRight
		}
		if ps.FirstLine > 0 {
			m.first = ps.FirstLine
		}
		if ps.Hanging > 0 {
			m.hanging = ps.Hanging
		}
		if ps.PageBreakBefore {
			m.pageBreak = true
		}
		if fs.Size > 0 {
			m.size = fs.Size
		}
		if fs.Name != "" {
			m.face = fs.Name
		} else if heading && d.themeMajor != "" {
			m.face = d.themeMajor
		}
		if fs.NameEastAsia != "" {
			m.east = fs.NameEastAsia
		} else if heading && d.themeMajorEA != "" {
			m.east = d.themeMajorEA
		}
		if fs.Bold {
			m.bold = true
		}
		if fs.SizeCS > 0 {
			m.csSize = fs.SizeCS
		}
		if fs.NameCS != "" {
			m.csFace = fs.NameCS
		}
		if ps.KeepNext {
			m.keepNext = true
		}
		if ps.KeepLines {
			m.keepLines = true
		}
		if ps.WidowControl {
			m.widow = true
		}
	}
	apply(d.defaultPara, style.FontStyle{}, false)
	if p == nil {
		return m
	}
	var chain []string
	seen := map[string]bool{}
	name := p.StyleName
	if name == "" {
		name = d.defaultStyle
	}
	for name != "" && !seen[name] && len(chain) < 12 {
		seen[name] = true
		chain = append(chain, name)
		st, ok := d.paragraphStyles[name]
		if !ok {
			break
		}
		name = st.BasedOn
	}
	for i := len(chain) - 1; i >= 0; i-- {
		id := chain[i]
		apply(d.paragraphStyles[id], d.styleFonts[id], headingish(id))
	}
	apply(p.Style, style.FontStyle{}, headingish(p.StyleName))
	for _, r := range p.Runs {
		if r.Style.Hidden {
			continue
		}
		if r.Style.Size > 0 {
			m.size = r.Style.Size
		}
		if r.Style.Name != "" {
			m.face = r.Style.Name
		}
		if r.Style.Bold {
			m.bold = true
		}
		if r.Style.SizeCS > 0 {
			m.csSize = r.Style.SizeCS
		}
		if r.Style.NameCS != "" {
			m.csFace = r.Style.NameCS
		}
		break
	}
	if complexScript(paragraphPlain(p)) && m.csSize > 0 {
		m.size = m.csSize
		if m.csFace != "" {
			m.face = m.csFace
		}
	}
	return m
}

func complexScript(s string) bool {
	for _, r := range s {
		if r >= 0x0590 && r <= 0x08FF {
			return true
		}
	}
	return false
}

func headingish(name string) bool {
	n := strings.ToLower(name)
	return strings.Contains(n, "heading") || n == "title"
}

func (c *pageCursor) paragraphBodyHeight(p *Paragraph, m paraMetrics) int {
	if p != nil && p.spacer {
		return c.lineBox(m)
	}
	lines, maxSize, pageBreak := c.wrapParagraph(p, m)
	if pageBreak {
		// The text before the break stays on this page. The break itself is
		// applied by the caller through the returned height plus a sheet
		// advance when the paragraph is only a break. A break after text
		// consumes the rest of the sheet.
		return c.lineBox(m)*lines + c.usableH
	}
	if maxSize > 0 {
		m.size = maxSize
	}
	return lines * c.lineBox(m)
}

func (c *pageCursor) wrapParagraph(p *Paragraph, m paraMetrics) (lines int, maxSize float64, pageBreak bool) {
	lines = 1
	maxSize = m.size
	if p == nil {
		return lines, maxSize, false
	}
	width := 0
	first := true
	limit := c.colWidth - m.indent
	if limit < 400 {
		limit = 400
	}
	firstLimit := limit - m.first
	if m.hanging > 0 && m.hanging < limit {
		firstLimit = limit - m.hanging
	}
	if firstLimit < 200 {
		firstLimit = 200
	}
	var runs []*Run
	if len(p.Runs) > 0 {
		runs = p.Runs
	}
	if len(runs) == 0 {
		return 1, maxSize, false
	}
	for _, r := range runs {
		if r.Style.Hidden {
			continue
		}
		font := r.Style
		if font.Size <= 0 {
			font.Size = m.size
		}
		if font.Name == "" {
			font.Name = m.face
		}
		if font.NameEastAsia == "" {
			font.NameEastAsia = m.east
		}
		if font.SizeCS <= 0 {
			font.SizeCS = m.csSize
		}
		if font.NameCS == "" {
			font.NameCS = m.csFace
		}
		if font.Size > maxSize {
			maxSize = font.Size
		}
		if r.Break && (r.BreakType == "page" || r.BreakType == "column") {
			pageBreak = true
		}
		if r.Tab {
			width += 720
		}
		for _, ch := range r.Text {
			if ch == '\n' {
				lines++
				width = 0
				first = false
				continue
			}
			face := font
			if face.SizeCS > 0 && ch >= 0x0590 && ch <= 0x08FF {
				face.Size = face.SizeCS
				if face.NameCS != "" {
					face.Name = face.NameCS
				}
			}
			w := runeTwips(ch, face, m.bold)
			lim := limit
			if first {
				lim = firstLimit
			}
			if width > 0 && width+w > lim {
				lines++
				width = w
				first = false
				continue
			}
			width += w
		}
		if r.Break && (r.BreakType == "" || r.BreakType == "textWrapping") {
			lines++
			width = 0
			first = false
		}
	}
	if lines < 1 {
		lines = 1
	}
	return lines, maxSize, pageBreak
}

func (c *pageCursor) textBlockHeight(text string, m paraMetrics) int {
	if strings.TrimSpace(text) == "" {
		return c.lineBox(m)
	}
	p := &Paragraph{Runs: []*Run{{Text: text}}}
	return c.paragraphBodyHeight(p, m)
}

func (c *pageCursor) lineBox(m paraMetrics) int {
	size := m.size
	if size <= 0 {
		size = 11
	}
	single := int(size * 20 * lineFactor(m.face))
	if single < 200 {
		single = 200
	}
	h := single
	switch m.rule {
	case "exact":
		if m.line > 0 {
			h = m.line
		}
	case "atLeast":
		if m.line > h {
			h = m.line
		}
	default:
		mult := 240
		if m.line > 0 {
			mult = m.line
		}
		h = single * mult / 240
	}
	// Snap plain lines up to the document grid. A heading that already has
	// space-before, or a paragraph with no explicit line value, is left at
	// the font's own leading. Stacking the grid on top of heading spacing
	// overshoots LibreOffice, and a grid on an unspecified line does too.
	if c.inTable {
		// Table rows use the font's own leading. The document grid snap that
		// body lines use makes a 52-row grid spill onto a second sheet.
		h = int(size * 24)
		if h < 200 {
			h = 200
		}
		return h
	}
	if c.pitch >= 240 && m.rule != "exact" && m.line > 0 && m.before < 200 && h < c.pitch {
		h = c.pitch
	}
	if c.pitch >= 240 && m.rule != "exact" && h < 240 {
		h = 240
	}
	if h < 160 {
		h = 160
	}
	return h
}

func lineFactor(name string) float64 {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "georgia"), strings.Contains(n, "cambria"), strings.Contains(n, "caladea"):
		return 1.22
	case strings.Contains(n, "times"), strings.Contains(n, "liberation serif"), strings.Contains(n, "perpetua"), strings.Contains(n, "cg times"):
		return 1.08
	case strings.Contains(n, "arabic"):
		return 1.10
	case strings.Contains(n, "calibri"), strings.Contains(n, "carlito"), strings.Contains(n, "aptos"):
		return 1.15
	default:
		return 1.15
	}
}

func runeTwips(r rune, font style.FontStyle, bold bool) int {
	size := font.Size
	if size <= 0 {
		size = 11
	}
	em := size * 20
	ratio := latinRatio(font.Name)
	if isWideRune(r) {
		ratio = 1
	}
	if bold && ratio < 1 {
		ratio += 0.04
	}
	if r == ' ' || r == '\u00a0' {
		ratio *= 0.45
	}
	w := int(em * ratio)
	if font.Spacing != 0 {
		w += int(font.Spacing * 20)
	}
	if w < 40 && r > ' ' {
		w = 40
	}
	return w
}

func latinRatio(name string) float64 {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "calibri"), strings.Contains(n, "carlito"), strings.Contains(n, "aptos"):
		return 0.46
	case strings.Contains(n, "times"), strings.Contains(n, "liberation serif"), strings.Contains(n, "perpetua"), strings.Contains(n, "nimbus rom"):
		return 0.43
	case strings.Contains(n, "cambria"), strings.Contains(n, "caladea"), strings.Contains(n, "georgia"):
		return 0.48
	case strings.Contains(n, "arial"), strings.Contains(n, "liberation sans"), strings.Contains(n, "helvetica"), strings.Contains(n, "nimbus san"):
		return 0.52
	case strings.Contains(n, "courier"), strings.Contains(n, "liberation mono"), strings.Contains(n, "nimbus mon"):
		return 0.60
	case strings.Contains(n, "consolas"), strings.Contains(n, "cascadia"):
		return 0.55
	case strings.Contains(n, "verdana"):
		return 0.58
	case strings.Contains(n, "garamond"):
		return 0.42
	case strings.Contains(n, "segoe"):
		return 0.50
	case strings.Contains(n, "tahoma"):
		return 0.50
	case strings.Contains(n, "noto sans"), strings.Contains(n, "yahei"), strings.Contains(n, "gothic"):
		return 0.52
	case strings.Contains(n, "noto serif"), strings.Contains(n, "mincho"), strings.Contains(n, "simsun"), strings.Contains(n, "nsimsun"):
		return 0.50
	default:
		return 0.50
	}
}

func isWideRune(r rune) bool {
	return (r >= 0x1100 && r <= 0x115F) ||
		(r >= 0x2E80 && r <= 0xA4CF) ||
		(r >= 0xAC00 && r <= 0xD7A3) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE10 && r <= 0xFE19) ||
		(r >= 0xFE30 && r <= 0xFE6F) ||
		(r >= 0xFF00 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x20000 && r <= 0x3FFFD)
}

func imageTwips(img *Image) int {
	if img == nil || img.Style.Height <= 0 {
		return 0
	}
	// Image heights loaded from drawings are pixels at 96 DPI.
	return img.Style.Height * common.TwipsPerInch / 96
}

func inlineExtra(p *Paragraph) int {
	if p == nil {
		return 0
	}
	h := 0
	for _, in := range p.Inlines {
		switch e := in.(type) {
		case *Image:
			if ih := imageTwips(e); ih > h {
				h = ih
			}
		case *RawXML:
			if ih := drawingTwips(e.XML); ih > h {
				h = ih
			}
		}
	}
	return h
}

func (c *pageCursor) reserveChrome(sec *Section) {
	if sec == nil {
		return
	}
	extra := 0
	hh, fh := 0, 0
	for _, h := range sec.Headers {
		if ch := chromeHeight(h.Elements); ch > hh {
			hh = ch
		}
	}
	for _, f := range sec.Footers {
		if ch := chromeHeight(f.Elements); ch > fh {
			fh = ch
		}
	}
	if hh > sec.Style.HeaderHeight {
		extra += hh - sec.Style.HeaderHeight
	}
	if fh > sec.Style.FooterHeight {
		extra += fh - sec.Style.FooterHeight
	}
	if extra > 0 && extra < c.usableH/2 {
		c.usableH -= extra
	}
}

func chromeHeight(elems []Element) int {
	h := 0
	for _, el := range elems {
		switch e := el.(type) {
		case *Paragraph:
			if strings.TrimSpace(paragraphPlain(e)) != "" {
				h += 240
			}
		case *Image:
			ih := imageTwips(e)
			if ih > 400 {
				ih = 400
			}
			h += ih
		case *RawXML:
			dh := drawingTwips(e.XML)
			if dh > 400 {
				dh = 400
			}
			h += dh
		case *Table:
			h += 240 * len(e.Rows)
		}
	}
	return h
}

func pageSizedAnchor(p *Paragraph) bool {
	if p == nil {
		return false
	}
	for _, in := range p.Inlines {
		raw, ok := in.(*RawXML)
		if !ok || !strings.Contains(raw.XML, ":anchor") {
			continue
		}
		if strings.Contains(raw.XML, `cy="7`) || strings.Contains(raw.XML, `cy="8`) || strings.Contains(raw.XML, `cy="9`) {
			return true
		}
	}
	return false
}

func drawingTwips(xml string) int {
	if xml == "" {
		return 0
	}
	// A floating shape behind the text does not add flow height. A text box
	// still occupies its extent even with no wrap.
	behind := strings.Contains(xml, "wrapNone") || strings.Contains(xml, `behindDoc="1"`)
	boxed := strings.Contains(xml, "txbx") || strings.Contains(xml, "textbox")
	if behind && !boxed && !strings.Contains(xml, ":inline") {
		return 0
	}
	best := 0
	rest := xml
	for {
		i := strings.Index(rest, `cy="`)
		if i < 0 {
			break
		}
		rest = rest[i+4:]
		j := strings.IndexByte(rest, '"')
		if j < 0 {
			break
		}
		n, err := strconv.ParseInt(rest[:j], 10, 64)
		rest = rest[j:]
		if err != nil || n < 20000 {
			continue
		}
		tw := int(n / common.EmuPerTwip)
		// A floating frame taller than about six inches is a page-sized
		// text box. Linked boxes of that size sit on the page; they do not
		// each add a sheet of flow.
		if tw > 9000 && strings.Contains(xml, ":anchor") {
			continue
		}
		if tw > best {
			best = tw
		}
	}
	if h := vmlPoints(xml, "height:"); h > best {
		best = h
	}
	return best
}

func vmlPoints(xml, key string) int {
	i := strings.Index(xml, key)
	if i < 0 {
		return 0
	}
	rest := xml[i+len(key):]
	end := 0
	for end < len(rest) && (rest[end] == '.' || (rest[end] >= '0' && rest[end] <= '9')) {
		end++
	}
	if end == 0 {
		return 0
	}
	f, err := strconv.ParseFloat(rest[:end], 64)
	if err != nil {
		return 0
	}
	return int(f * 20)
}

func cellWidth(t *Table, cell *Cell, fallback int) int {
	span := 1
	if cell != nil && cell.Style.GridSpan > 1 {
		span = cell.Style.GridSpan
	}
	if cell != nil && cell.Style.Width > 0 {
		return cell.Style.Width - 160
	}
	if t != nil && len(t.Grid) > 0 {
		sum := 0
		for _, g := range t.Grid {
			sum += g
		}
		if sum > 0 {
			w := sum * span / len(t.Grid)
			if w > 200 {
				return w - 160
			}
		}
	}
	if fallback > 200 {
		return fallback*span - 80
	}
	return fallback
}

func (c *pageCursor) noteBookmarks(p *Paragraph, lay *Layout) {
	var open []string
	bufs := map[string]*stringBuf{}
	consider := func(el Element) {
		switch e := el.(type) {
		case *Bookmark:
			if e.Name == "" {
				return
			}
			lay.bookmarkPage[e.Name] = c.page
			bufs[e.Name] = &stringBuf{}
			open = append(open, e.Name)
		case *Run:
			for _, name := range open {
				bufs[name].s += e.Text
			}
		case *RawXML:
			for _, name := range open {
				bufs[name].s += e.Text
			}
		case *Field:
			for _, name := range open {
				bufs[name].s += e.Result
			}
		}
	}
	if len(p.Inlines) > 0 {
		for _, in := range p.Inlines {
			consider(in)
		}
	} else {
		for _, r := range p.Runs {
			consider(r)
		}
	}
	for _, name := range open {
		if b, ok := bufs[name]; ok {
			lay.bookmarkText[name] = b.s
		}
	}
}

type stringBuf struct{ s string }
