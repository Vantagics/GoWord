package document

import (
	"strconv"
	"strings"
)

// numberingFor returns the numId and ilvl for a paragraph. A direct numPr
// wins. Otherwise the paragraph style's numPr is used, following w:basedOn.
func (d *Document) numberingFor(p *Paragraph) (string, int) {
	if p == nil {
		return "", 0
	}
	if p.Style.NumStyleName != "" || d == nil {
		return p.Style.NumStyleName, p.Style.NumLevel
	}
	name := p.StyleName
	seen := map[string]bool{}
	for name != "" && !seen[name] {
		seen[name] = true
		ps, ok := d.paragraphStyles[name]
		if !ok {
			break
		}
		if ps.NumStyleName != "" {
			level := ps.NumLevel
			return ps.NumStyleName, level
		}
		name = ps.BasedOn
	}
	return "", 0
}

// listLabel renders the visible marker for a numbering instance, advancing
// counters in ctr. An empty string means the paragraph is not numbered or the
// definition is not available.
func (d *Document) listLabel(numID string, level int, ctr map[string][]int) string {
	if d == nil || numID == "" || numID == "0" || ctr == nil {
		return ""
	}
	ns, ok := d.numberingStyles["num"+numID]
	if !ok {
		ns, ok = d.numberingStyles[numID]
	}
	if !ok || len(ns.Levels) == 0 {
		return ""
	}
	if level < 0 {
		level = 0
	}
	if level >= len(ns.Levels) {
		level = len(ns.Levels) - 1
	}
	arr := ctr[numID]
	for len(arr) <= level {
		arr = append(arr, 0)
	}
	arr = arr[:level+1]
	start := lvStart(ns.Levels[level])
	if arr[level] == 0 {
		arr[level] = start - 1
	}
	arr[level]++
	ctr[numID] = arr

	lv := ns.Levels[level]
	if lv.Format == "bullet" {
		text := lv.Text
		if text == "" {
			text = "•"
		}
		return text + " "
	}
	pattern := lv.Text
	if pattern == "" {
		pattern = "%" + strconv.Itoa(level+1) + "."
	}
	out := pattern
	for i := 1; i <= level+1 && i <= len(arr); i++ {
		fmtName := "decimal"
		if i-1 < len(ns.Levels) && ns.Levels[i-1].Format != "" && ns.Levels[i-1].Format != "bullet" {
			fmtName = ns.Levels[i-1].Format
		}
		out = strings.ReplaceAll(out, "%"+strconv.Itoa(i), formatListNumber(arr[i-1], fmtName))
	}
	if out == "" {
		return ""
	}
	return out + " "
}

func lvStart(lv NumberingLevel) int {
	if lv.Start <= 0 {
		return 1
	}
	return lv.Start
}

func formatListNumber(n int, format string) string {
	if n < 1 {
		n = 1
	}
	switch format {
	case "lowerLetter", "lowerLatin":
		return alpha(n, 'a')
	case "upperLetter", "upperLatin":
		return alpha(n, 'A')
	case "lowerRoman":
		return roman(n, false)
	case "upperRoman":
		return roman(n, true)
	case "decimalZero":
		if n < 10 {
			return "0" + strconv.Itoa(n)
		}
		return strconv.Itoa(n)
	default:
		return strconv.Itoa(n)
	}
}

func alpha(n int, base byte) string {
	n--
	if n < 0 {
		n = 0
	}
	if n < 26 {
		return string(base + byte(n))
	}
	return string(base+byte(n/26-1)) + string(base+byte(n%26))
}

func roman(n int, upper bool) string {
	if n <= 0 {
		return "0"
	}
	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syms := []string{"m", "cm", "d", "cd", "c", "xc", "l", "xl", "x", "ix", "v", "iv", "i"}
	var b strings.Builder
	for i, v := range vals {
		for n >= v {
			b.WriteString(syms[i])
			n -= v
		}
	}
	if upper {
		return strings.ToUpper(b.String())
	}
	return b.String()
}
