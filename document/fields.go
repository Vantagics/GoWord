package document

import (
	"strconv"
	"strings"
	"time"
	"unicode"
)

// SetMergeData supplies values for MERGEFIELD instructions. Keys are field
// names, compared case-insensitively.
func (d *Document) SetMergeData(data map[string]string) {
	if d == nil {
		return
	}
	d.mergeData = data
}

// MergeData returns the mail-merge values set on the document.
func (d *Document) MergeData() map[string]string {
	if d == nil {
		return nil
	}
	return d.mergeData
}

// UpdateFields refreshes cached field results from the document and sets
// updateFields so Word recalculates on open. Page numbers come from
// EstimateLayout, which is an approximation. DATE and TIME use time.Now.
func (d *Document) UpdateFields() {
	if d == nil {
		return
	}
	d.UpdateFieldsAt(time.Now())
}

// UpdateFieldsAt is UpdateFields with an explicit clock, so tests can pin DATE.
func (d *Document) UpdateFieldsAt(now time.Time) {
	if d == nil {
		return
	}
	d.SetUpdateFieldsOnOpen(true)
	lay := d.EstimateLayout()
	d.fillFields(lay, now)
}

func (d *Document) fillFields(lay Layout, now time.Time) {
	seq := map[string]int{}
	var walk func(elems []Element, page int)
	walk = func(elems []Element, page int) {
		for _, el := range elems {
			switch e := el.(type) {
			case *Paragraph:
				pg := lay.PageOf(e)
				for _, in := range e.Inlines {
					if f, ok := in.(*Field); ok {
						d.evalFieldTree(f, pg, lay, seq, now)
					}
				}
			case *TOC:
				e.Result = d.tocResult(e.MinDepth, e.MaxDepth, lay)
			case *Field:
				d.evalFieldTree(e, page, lay, seq, now)
			case *Sdt:
				walk(e.Elements, page)
			case *Table:
				for _, row := range e.Rows {
					for _, cell := range row.Cells {
						walk(cell.Elements, page)
					}
				}
			}
		}
	}
	for _, sec := range d.Sections {
		walk(sec.Elements, 1)
	}
}

func (d *Document) tocResult(minD, maxD int, lay Layout) string {
	if minD <= 0 {
		minD = 1
	}
	if maxD <= 0 {
		maxD = 9
	}
	var lines []string
	var walk func([]Element)
	walk = func(elems []Element) {
		for _, el := range elems {
			switch e := el.(type) {
			case *Paragraph:
				lvl := headingLevel(e.StyleName)
				if lvl >= minD && lvl <= maxD {
					if text := paragraphPlain(e); text != "" {
						lines = append(lines, text+"\t"+strconv.Itoa(lay.PageOf(e)))
					}
				}
			case *Sdt:
				walk(e.Elements)
			case *Table:
				for _, row := range e.Rows {
					for _, cell := range row.Cells {
						walk(cell.Elements)
					}
				}
			}
		}
	}
	for _, sec := range d.Sections {
		walk(sec.Elements)
	}
	return strings.Join(lines, "\n")
}

func (d *Document) evalFieldTree(f *Field, page int, lay Layout, seq map[string]int, now time.Time) {
	if f == nil {
		return
	}
	for _, c := range f.Content {
		if inner, ok := c.(*Field); ok {
			d.evalFieldTree(inner, page, lay, seq, now)
		}
	}
	name, _, _ := parseField(f.Instr)
	switch strings.ToUpper(name) {
	case "TOC":
		if fieldHasNested(f) {
			f.Result = renderFieldContent(f.Content)
			return
		}
		f.Result = d.evalField(f.Instr, page, lay, seq, now)
		f.Content = nil
	case "PAGE", "NUMPAGES", "PAGEREF", "REF", "SEQ", "MERGEFIELD", "DATE", "TIME", "CREATEDATE", "PRINTDATE", "SAVEDATE", "IF":
		f.Result = d.evalField(f.Instr, page, lay, seq, now)
		f.Content = nil
	default:
		if len(f.Content) > 0 {
			f.Result = renderFieldContent(f.Content)
		}
	}
}

func fieldHasNested(f *Field) bool {
	for _, c := range f.Content {
		if _, ok := c.(*Field); ok {
			return true
		}
	}
	return false
}

func renderFieldContent(content []Element) string {
	var b strings.Builder
	for _, c := range content {
		switch e := c.(type) {
		case *Run:
			if e.Tab {
				b.WriteByte('\t')
			}
			b.WriteString(e.Text)
		case *Tab:
			b.WriteByte('\t')
		case *Field:
			b.WriteString(e.Result)
		}
	}
	return b.String()
}

func (d *Document) evalField(instr string, page int, lay Layout, seq map[string]int, now time.Time) string {
	name, args, sw := parseField(instr)
	switch strings.ToUpper(name) {
	case "PAGE":
		if page < 1 {
			page = 1
		}
		return formatNumber(page, sw["*"])
	case "NUMPAGES":
		n := lay.PageCount
		if n < 1 {
			n = 1
		}
		return formatNumber(n, sw["*"])
	case "PAGEREF":
		if len(args) == 0 {
			return ""
		}
		if n, ok := lay.BookmarkPage(args[0]); ok {
			return formatNumber(n, sw["*"])
		}
		return ""
	case "REF":
		if len(args) == 0 {
			return ""
		}
		if s, ok := lay.BookmarkText(args[0]); ok {
			return s
		}
		return ""
	case "SEQ":
		if len(args) == 0 {
			return ""
		}
		id := strings.ToLower(args[0])
		if _, repeat := sw["c"]; repeat {
			n := seq[id]
			if n < 1 {
				n = 1
			}
			return formatNumber(n, sw["*"])
		}
		seq[id]++
		return formatNumber(seq[id], sw["*"])
	case "MERGEFIELD":
		if len(args) == 0 {
			return ""
		}
		return d.lookupMerge(args[0])
	case "DATE", "TIME", "CREATEDATE", "PRINTDATE", "SAVEDATE":
		return formatWordDate(now, sw["@"])
	case "IF":
		return d.evalIF(args)
	case "TOC":
		minD, maxD := 1, 9
		if r := sw["o"]; r != "" {
			parts := strings.Split(r, "-")
			if len(parts) == 2 {
				if n, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
					minD = n
				}
				if n, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					maxD = n
				}
			}
		}
		return d.tocResult(minD, maxD, lay)
	default:
		return ""
	}
}

func (d *Document) lookupMerge(name string) string {
	if d == nil || d.mergeData == nil {
		return ""
	}
	if v, ok := d.mergeData[name]; ok {
		return v
	}
	for k, v := range d.mergeData {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	return ""
}

func (d *Document) evalIF(args []string) string {
	// IF <left> <op> <right> "true" "false"
	// left may be MERGEFIELD Name.
	if len(args) < 5 {
		return ""
	}
	i := 0
	left := args[i]
	i++
	if strings.EqualFold(left, "MERGEFIELD") && i < len(args) {
		left = d.lookupMerge(args[i])
		i++
	}
	if i >= len(args) {
		return ""
	}
	op := args[i]
	i++
	if i >= len(args) {
		return ""
	}
	right := args[i]
	i++
	var yes, no string
	if i < len(args) {
		yes = args[i]
		i++
	}
	if i < len(args) {
		no = args[i]
	}
	if ifMatch(left, op, right) {
		return yes
	}
	return no
}

func ifMatch(left, op, right string) bool {
	ln, lok := strconv.Atoi(strings.TrimSpace(left))
	rn, rok := strconv.Atoi(strings.TrimSpace(right))
	if lok == nil && rok == nil {
		switch op {
		case "=":
			return ln == rn
		case "<>":
			return ln != rn
		case ">":
			return ln > rn
		case "<":
			return ln < rn
		case ">=":
			return ln >= rn
		case "<=":
			return ln <= rn
		}
	}
	switch op {
	case "=":
		return left == right
	case "<>":
		return left != right
	case ">":
		return left > right
	case "<":
		return left < right
	default:
		return false
	}
}

func formatNumber(n int, kind string) string {
	switch strings.ToUpper(strings.TrimSpace(kind)) {
	case "ROMAN":
		return romanUpper(n)
	case "ROMANLOWER", "LOWERROMAN":
		return strings.ToLower(romanUpper(n))
	case "ALPHABETIC":
		return alphaNum(n, 'A')
	case "ALPHABETICLOWER", "LOWERALPHABETIC":
		return alphaNum(n, 'a')
	case "ORDINAL":
		return ordinal(n)
	default:
		return strconv.Itoa(n)
	}
}

func romanUpper(n int) string {
	if n <= 0 {
		return strconv.Itoa(n)
	}
	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syms := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var b strings.Builder
	for i, v := range vals {
		for n >= v {
			b.WriteString(syms[i])
			n -= v
		}
	}
	return b.String()
}

func alphaNum(n int, base byte) string {
	if n < 1 {
		n = 1
	}
	n--
	if n < 26 {
		return string(base + byte(n))
	}
	return string(base+byte(n/26-1)) + string(base+byte(n%26))
}

func ordinal(n int) string {
	suf := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suf = "st"
		case 2:
			suf = "nd"
		case 3:
			suf = "rd"
		}
	}
	return strconv.Itoa(n) + suf
}

func formatWordDate(t time.Time, picture string) string {
	if picture == "" {
		return t.Format("2 January 2006")
	}
	var b strings.Builder
	s := picture
	for len(s) > 0 {
		switch {
		case strings.HasPrefix(s, "yyyy"):
			b.WriteString(t.Format("2006"))
			s = s[4:]
		case strings.HasPrefix(s, "yy"):
			b.WriteString(t.Format("06"))
			s = s[2:]
		case strings.HasPrefix(s, "MMMM"):
			b.WriteString(t.Format("January"))
			s = s[4:]
		case strings.HasPrefix(s, "MMM"):
			b.WriteString(t.Format("Jan"))
			s = s[3:]
		case strings.HasPrefix(s, "MM"):
			b.WriteString(t.Format("01"))
			s = s[2:]
		case strings.HasPrefix(s, "M"):
			b.WriteString(strconv.Itoa(int(t.Month())))
			s = s[1:]
		case strings.HasPrefix(s, "dddd"):
			b.WriteString(t.Format("Monday"))
			s = s[4:]
		case strings.HasPrefix(s, "ddd"):
			b.WriteString(t.Format("Mon"))
			s = s[3:]
		case strings.HasPrefix(s, "dd"):
			b.WriteString(t.Format("02"))
			s = s[2:]
		case strings.HasPrefix(s, "d"):
			b.WriteString(strconv.Itoa(t.Day()))
			s = s[1:]
		case strings.HasPrefix(s, "HH"):
			b.WriteString(t.Format("15"))
			s = s[2:]
		case strings.HasPrefix(s, "H"):
			b.WriteString(strconv.Itoa(t.Hour()))
			s = s[1:]
		case strings.HasPrefix(s, "hh"):
			b.WriteString(t.Format("03"))
			s = s[2:]
		case strings.HasPrefix(s, "h"):
			h := t.Hour() % 12
			if h == 0 {
				h = 12
			}
			b.WriteString(strconv.Itoa(h))
			s = s[1:]
		case strings.HasPrefix(s, "mm"):
			b.WriteString(t.Format("04"))
			s = s[2:]
		case strings.HasPrefix(s, "ss"):
			b.WriteString(t.Format("05"))
			s = s[2:]
		case strings.HasPrefix(s, "AM/PM"):
			b.WriteString(t.Format("PM"))
			s = s[5:]
		case strings.HasPrefix(s, "am/pm"):
			b.WriteString(strings.ToLower(t.Format("PM")))
			s = s[5:]
		default:
			b.WriteByte(s[0])
			s = s[1:]
		}
	}
	return b.String()
}

func parseField(instr string) (name string, args []string, switches map[string]string) {
	switches = map[string]string{}
	toks := fieldTokens(instr)
	if len(toks) == 0 {
		return "", nil, switches
	}
	name = toks[0]
	for i := 1; i < len(toks); i++ {
		t := toks[i]
		if strings.HasPrefix(t, "\\") && len(t) > 1 {
			key := strings.TrimPrefix(t, "\\")
			val := ""
			if i+1 < len(toks) && !strings.HasPrefix(toks[i+1], "\\") {
				i++
				val = toks[i]
			} else {
				val = "true"
			}
			switches[key] = val
			continue
		}
		args = append(args, t)
	}
	return name, args, switches
}

func fieldTokens(s string) []string {
	var out []string
	i := 0
	for i < len(s) {
		for i < len(s) && unicode.IsSpace(rune(s[i])) {
			i++
		}
		if i >= len(s) {
			break
		}
		if s[i] == '"' {
			j := i + 1
			var b strings.Builder
			for j < len(s) && s[j] != '"' {
				b.WriteByte(s[j])
				j++
			}
			out = append(out, b.String())
			if j < len(s) {
				j++
			}
			i = j
			continue
		}
		j := i
		for j < len(s) && !unicode.IsSpace(rune(s[j])) {
			j++
		}
		out = append(out, s[i:j])
		i = j
	}
	return out
}
