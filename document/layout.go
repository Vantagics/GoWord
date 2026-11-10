package document

import "unicode/utf8"

// Layout is an estimate of where body content falls on pages. It is not
// Word's layout: it uses section page size, margins, a fixed line height,
// and a fixed average character width. Explicit page breaks are exact.
// Flowing text is usually within a page of LibreOffice.
type Layout struct {
	PageCount    int
	paraPage     map[*Paragraph]int
	bookmarkPage map[string]int
	bookmarkText map[string]string
}

// PageOf returns the 1-based page estimated for a paragraph. Unknown
// paragraphs are reported as page 1.
func (l Layout) PageOf(p *Paragraph) int {
	if l.paraPage == nil || p == nil {
		if l.PageCount < 1 {
			return 1
		}
		return 1
	}
	if n, ok := l.paraPage[p]; ok && n > 0 {
		return n
	}
	return 1
}

// BookmarkPage returns the estimated page of a bookmark.
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
		for _, el := range sec.Elements {
			cur.place(el, &lay)
		}
	}
	if cur.page < 1 {
		cur.page = 1
	}
	lay.PageCount = cur.page
	return lay
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
	out.breakType = st.BreakType
	out.pageStart = st.PageNumberStart
	return out
}

type styleSection struct {
	width, height           int
	marginTop, marginBottom int
	marginLeft, marginRight int
	columns                 int
	breakType               string
	pageStart               *int
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
	switch kind {
	case "continuous":
		return false
	default:
		return true
	}
}

type pageCursor struct {
	page, y          int
	usableH, usableW int
	columns          int
	line             int
	charW            int
}

func newPageCursor(st styleSection) *pageCursor {
	c := &pageCursor{page: 1, line: 276, charW: 120, columns: 1}
	c.applySection(st)
	return c
}

func (c *pageCursor) applySection(st styleSection) {
	h := st.height - st.marginTop - st.marginBottom
	w := st.width - st.marginLeft - st.marginRight
	if h < 2000 {
		h = 2000
	}
	if w < 2000 {
		w = 2000
	}
	cols := st.columns
	if cols < 1 {
		cols = 1
	}
	c.usableH = h
	c.usableW = w
	c.columns = cols
	c.line = 276
	c.charW = 120
}

func (c *pageCursor) breakPage(kind string) {
	if c.y > 0 {
		c.page++
		c.y = 0
	} else if c.page == 0 {
		c.page = 1
	}
	switch kind {
	case "oddPage":
		if c.page%2 == 0 {
			c.page++
		}
	case "evenPage":
		if c.page%2 == 1 {
			c.page++
			c.y = 0
		}
	}
}

func (c *pageCursor) addHeight(h int) int {
	if h < c.line {
		h = c.line
	}
	if c.y > 0 && c.y+h > c.usableH {
		c.page++
		c.y = 0
	}
	start := c.page
	c.y += h
	for c.y > c.usableH {
		c.y -= c.usableH
		c.page++
	}
	return start
}

func (c *pageCursor) place(el Element, lay *Layout) {
	switch e := el.(type) {
	case *PageBreak:
		c.page++
		c.y = 0
	case *Paragraph:
		start := c.addHeight(c.paragraphHeight(e))
		lay.paraPage[e] = start
		saved := c.page
		c.page = start
		c.noteBookmarks(e, lay)
		c.page = saved
	case *Table:
		h := 0
		for _, row := range e.Rows {
			rh := row.Style.Height
			if rh <= 0 {
				rh = c.line
				for _, cell := range row.Cells {
					ch := c.textHeight(cellText(cell))
					if ch > rh {
						rh = ch
					}
				}
			}
			h += rh
		}
		if h == 0 {
			h = c.line
		}
		c.addHeight(h)
		for _, row := range e.Rows {
			for _, cell := range row.Cells {
				for _, ce := range cell.Elements {
					if p, ok := ce.(*Paragraph); ok {
						c.noteBookmarks(p, lay)
						lay.paraPage[p] = c.page
					}
				}
			}
		}
	case *TextBreak:
		n := e.Count
		if n < 1 {
			n = 1
		}
		c.addHeight(n * c.line)
	case *Image:
		h := e.Style.Height * 20
		if h <= 0 {
			h = 40 * 20
		}
		c.addHeight(h)
	case *Sdt:
		for _, child := range e.Elements {
			c.place(child, lay)
		}
	case *ListItem:
		c.addHeight(c.textHeight(e.Text))
	default:
		c.addHeight(c.line)
	}
}

func (c *pageCursor) paragraphHeight(p *Paragraph) int {
	h := c.textHeight(paragraphPlain(p))
	if p.Style.SpaceBefore > 0 {
		h += p.Style.SpaceBefore
	}
	if p.Style.SpaceAfter > 0 {
		h += p.Style.SpaceAfter
	}
	for _, r := range p.Runs {
		if r.Break && (r.BreakType == "page" || r.BreakType == "column") {
			h += c.usableH
		}
	}
	return h
}

func (c *pageCursor) textHeight(text string) int {
	n := utf8.RuneCountInString(text)
	if n == 0 {
		n = 1
	}
	per := c.usableW / c.charW
	if c.columns > 1 {
		per /= c.columns
	}
	if per < 8 {
		per = 8
	}
	lines := (n + per - 1) / per
	if lines < 1 {
		lines = 1
	}
	if c.columns > 1 {
		lines = (lines + c.columns - 1) / c.columns
	}
	return lines * c.line
}

func (c *pageCursor) noteBookmarks(p *Paragraph, lay *Layout) {
	var open []string
	var bufs = map[string]*stringBuf{}
	finish := func() {
		for _, name := range open {
			if b, ok := bufs[name]; ok {
				lay.bookmarkText[name] = b.s
			}
		}
	}
	consider := func(el Element) {
		switch e := el.(type) {
		case *Bookmark:
			if e.Name == "" {
				return
			}
			lay.bookmarkPage[e.Name] = c.page
			bufs[e.Name] = &stringBuf{}
			open = append(open, e.Name)
		case *BookmarkEnd:
			// Point bookmarks have no span. Keep whatever text followed.
		case *Run:
			for _, name := range open {
				bufs[name].s += e.Text
			}
		case *RawXML:
			for _, name := range open {
				bufs[name].s += e.Text
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
	finish()
	if len(open) == 0 {
		return
	}
}

type stringBuf struct{ s string }

func cellText(cell *Cell) string {
	var b string
	for _, el := range cell.Elements {
		if p, ok := el.(*Paragraph); ok {
			b += paragraphPlain(p)
		}
	}
	return b
}
