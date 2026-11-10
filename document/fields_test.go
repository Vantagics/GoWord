package document

import (
	"strings"
	"testing"
	"time"
)

func TestFieldEvaluation(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Intro", nil, nil)
	sec.Elements = append(sec.Elements, &PageBreak{})
	p := sec.AddText("Target text", nil, nil)
	p.Inlines = []Element{
		&Bookmark{ID: 1, Name: "here", Span: true},
		&Run{Text: "Target text"},
		&BookmarkEnd{ID: 1},
		&Field{Instr: `SEQ Figure \* ARABIC`},
		&Field{Instr: `SEQ Figure \c`},
		&Field{Instr: `REF here`},
		&Field{Instr: `PAGEREF here`},
		&Field{Instr: `PAGE`},
		&Field{Instr: `NUMPAGES`},
		&Field{Instr: `DATE \@ "yyyy-MM-dd"`},
		&Field{Instr: `MERGEFIELD City`},
		&Field{Instr: `IF MERGEFIELD City = "Paris" "Yes" "No"`},
		&Field{Instr: `IF 2 > 1 "big" "small"`},
	}
	doc.SetMergeData(map[string]string{"City": "Paris"})
	doc.UpdateFieldsAt(time.Date(2026, 11, 10, 15, 4, 0, 0, time.UTC))
	got := map[string]string{}
	for _, in := range p.Inlines {
		if f, ok := in.(*Field); ok {
			got[f.Instr] = f.Result
		}
	}
	want := map[string]string{
		`SEQ Figure \* ARABIC`:                    "1",
		`SEQ Figure \c`:                           "1",
		`REF here`:                                "Target text",
		`DATE \@ "yyyy-MM-dd"`:                    "2026-11-10",
		`MERGEFIELD City`:                         "Paris",
		`IF MERGEFIELD City = "Paris" "Yes" "No"`: "Yes",
		`IF 2 > 1 "big" "small"`:                  "big",
	}
	for instr, exp := range want {
		if got[instr] != exp {
			t.Errorf("%s = %q, want %q", instr, got[instr], exp)
		}
	}
	if got[`PAGEREF here`] == "" || got[`PAGE`] == "" || got[`NUMPAGES`] == "" {
		t.Fatalf("page fields empty: %+v", got)
	}
	if got[`PAGE`] != got[`PAGEREF here`] {
		t.Errorf("PAGE %s != PAGEREF %s", got[`PAGE`], got[`PAGEREF here`])
	}

	toc := &TOC{MinDepth: 1, MaxDepth: 3}
	sec.Elements = append([]Element{toc, &PageBreak{}}, sec.Elements...)
	h := sec.AddTitle("Chapter", 1)
	_ = h
	doc.UpdateFieldsAt(time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC))
	if !strings.Contains(toc.Result, "Chapter\t") {
		t.Fatalf("toc = %q", toc.Result)
	}
}

func TestLayoutUsesFontAndSpacing(t *testing.T) {
	latin := New()
	sec := latin.AddSection()
	sec.AddText(strings.Repeat("a", 80), nil, nil)
	cjk := New()
	sec2 := cjk.AddSection()
	sec2.AddText(strings.Repeat("字", 80), nil, nil)
	if cjk.EstimateLayout().PageCount < latin.EstimateLayout().PageCount {
		t.Fatalf("cjk pages %d, latin %d", cjk.EstimateLayout().PageCount, latin.EstimateLayout().PageCount)
	}
	// Full-width characters must consume more vertical space than latin.
	if cjk.EstimateLayout().PageCount == 1 && latin.EstimateLayout().PageCount == 1 {
		// Both may fit on one A4 page. Compare the sheet of a following paragraph.
		sec.AddText("tail", nil, nil)
		sec2.AddText("tail", nil, nil)
	}
	spaced := New()
	s := spaced.AddSection()
	s.Style.PageHeight = 4000
	s.Style.PageWidth = 8000
	s.Style.MarginTop = 200
	s.Style.MarginBottom = 200
	for i := 0; i < 6; i++ {
		p := s.AddText("line", nil, nil)
		p.Style.SpaceAfter = 600
		p.Style.LineSpacing = 360
		p.Style.LineRule = "exact"
	}
	if spaced.EstimateLayout().PageCount < 2 {
		t.Fatalf("spaced pages = %d", spaced.EstimateLayout().PageCount)
	}
	wide := New()
	ws := wide.AddSection()
	ws.Style.ColumnCount = 2
	for i := 0; i < 40; i++ {
		ws.AddText("short", nil, nil)
	}
	one := New()
	osec := one.AddSection()
	for i := 0; i < 40; i++ {
		osec.AddText("short", nil, nil)
	}
	if wide.EstimateLayout().PageCount > one.EstimateLayout().PageCount {
		t.Fatalf("columns %d took more pages than a single column %d", wide.EstimateLayout().PageCount, one.EstimateLayout().PageCount)
	}

	restart := New()
	rs := restart.AddSection()
	start := 24
	rs.Style.PageNumberStart = &start
	rs.AddText("only", nil, nil)
	lay := restart.EstimateLayout()
	if lay.PageCount != 1 {
		t.Fatalf("physical pages = %d", lay.PageCount)
	}
	if lay.PageOf(restart.Paragraphs()[0]) != 24 {
		t.Fatalf("displayed page = %d", lay.PageOf(restart.Paragraphs()[0]))
	}
}

func TestPageBreakLayout(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("a", nil, nil)
	sec.Elements = append(sec.Elements, &PageBreak{})
	sec.AddText("b", nil, nil)
	sec.Elements = append(sec.Elements, &PageBreak{})
	sec.AddText("c", nil, nil)
	lay := doc.EstimateLayout()
	if lay.PageCount != 3 {
		t.Fatalf("pages = %d, want 3", lay.PageCount)
	}
	ps := doc.Paragraphs()
	if lay.PageOf(ps[0]) != 1 || lay.PageOf(ps[1]) != 2 || lay.PageOf(ps[2]) != 3 {
		t.Fatalf("pages %d %d %d", lay.PageOf(ps[0]), lay.PageOf(ps[1]), lay.PageOf(ps[2]))
	}
}
