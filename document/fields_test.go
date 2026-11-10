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
