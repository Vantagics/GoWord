package document

import (
	"strings"
	"testing"
)

func TestAcceptAndRejectRevisions(t *testing.T) {
	const raw = `<w:r><w:t>Keep</w:t></w:r><w:del w:id="1"><w:r><w:delText>Gone</w:delText></w:r></w:del><w:ins w:id="2"><w:r><w:t>New</w:t></w:r></w:ins>`
	acc := rewriteMarkup(raw, true)
	if !strings.Contains(acc, "Keep") || !strings.Contains(acc, "New") || strings.Contains(acc, "Gone") || strings.Contains(acc, "<w:del") || strings.Contains(acc, "<w:ins") {
		t.Fatalf("accept = %s", acc)
	}
	rej := rewriteMarkup(raw, false)
	if !strings.Contains(rej, "Keep") || !strings.Contains(rej, "Gone") || strings.Contains(rej, "New") || strings.Contains(rej, "delText") {
		t.Fatalf("reject = %s", rej)
	}

	const props = `<w:rPr><w:b/><w:rPrChange w:id="1" w:author="A"><w:rPr><w:i/></w:rPr></w:rPrChange></w:rPr>`
	accP := rewriteMarkup(props, true)
	if !strings.Contains(accP, "<w:b/>") || strings.Contains(accP, "rPrChange") || strings.Contains(accP, "<w:i/>") {
		t.Fatalf("accept props = %s", accP)
	}
	rejP := rewriteMarkup(props, false)
	if !strings.Contains(rejP, "<w:i/>") || strings.Contains(rejP, "rPrChange") || strings.Contains(rejP, "<w:b/>") {
		t.Fatalf("reject props = %s", rejP)
	}

	doc := New()
	sec := doc.AddSection()
	p := sec.AddText("x", nil, nil)
	p.Inlines = []Element{&RawXML{XML: raw, Text: "KeepGoneNew"}}
	p.Runs = []*Run{{Text: "KeepGoneNew"}}
	doc.AcceptRevisions()
	if !strings.Contains(doc.ExtractText(), "Keep") || !strings.Contains(doc.ExtractText(), "New") || strings.Contains(doc.ExtractText(), "Gone") {
		t.Fatalf("document accept = %q", doc.ExtractText())
	}
}
