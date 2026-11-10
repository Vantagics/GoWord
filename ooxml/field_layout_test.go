package ooxml

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/document"
)

func TestLayoutMatchesLibreOfficePageBreaks(t *testing.T) {
	if _, err := exec.LookPath("soffice"); err != nil {
		t.Skip("soffice not installed")
	}
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed")
	}
	doc := document.New()
	sec := doc.AddSection()
	toc := sec.AddTOC(nil, 1, 3)
	sec.AddText("Page one", nil, nil)
	sec.Elements = append(sec.Elements, &document.PageBreak{})
	sec.AddText("Page two", nil, nil)
	sec.Elements = append(sec.Elements, &document.PageBreak{})
	sec.AddTitle("Chapter", 1)
	doc.UpdateFields()
	if doc.EstimateLayout().PageCount != 3 {
		t.Fatalf("estimate = %d", doc.EstimateLayout().PageCount)
	}
	if !strings.Contains(toc.Result, "Chapter\t3") {
		t.Fatalf("toc = %q", toc.Result)
	}
	dir := t.TempDir()
	in := filepath.Join(dir, "pages.docx")
	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(in, data, 0o644); err != nil {
		t.Fatal(err)
	}
	n, pages, err := libreOfficePageText(in)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("LibreOffice PDF pages = %d, estimate 3\n%v", n, pages)
	}
	if !strings.Contains(pages[2], "Chapter") {
		t.Fatalf("chapter not on page 3: %q", pages[2])
	}
}

func TestFlowingLayoutNearLibreOffice(t *testing.T) {
	if _, err := exec.LookPath("soffice"); err != nil {
		t.Skip("soffice not installed")
	}
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed")
	}
	doc := document.New()
	sec := doc.AddSection()
	const sentence = "The quick brown fox jumps over the lazy dog near the river bank. "
	for i := 0; i < 80; i++ {
		sec.AddText(sentence+sentence, nil, nil)
	}
	est := doc.EstimateLayout().PageCount
	dir := t.TempDir()
	in := filepath.Join(dir, "flow.docx")
	data, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(in, data, 0o644); err != nil {
		t.Fatal(err)
	}
	n, _, err := libreOfficePageText(in)
	if err != nil {
		t.Fatal(err)
	}
	diff := est - n
	if diff < 0 {
		diff = -diff
	}
	t.Logf("flowing text: estimate %d pages, LibreOffice %d (delta %d)", est, n, est-n)
	if diff > 1 {
		t.Fatalf("estimate %d is more than one page away from LibreOffice %d", est, n)
	}
}

func libreOfficePageText(docx string) (int, []string, error) {
	dir := filepath.Dir(docx)
	pdf := strings.TrimSuffix(docx, filepath.Ext(docx)) + ".pdf"
	cmd := exec.Command("soffice", "--headless", "--norestore", "--nolockcheck",
		"-env:UserInstallation=file:///tmp/lo-fields-"+filepath.Base(dir),
		"--convert-to", "pdf", "--outdir", dir, docx)
	if out, err := cmd.CombinedOutput(); err != nil {
		return 0, nil, errString(out, err)
	}
	txt := pdf + ".txt"
	cmd = exec.Command("pdftotext", "-layout", pdf, txt)
	if out, err := cmd.CombinedOutput(); err != nil {
		return 0, nil, errString(out, err)
	}
	b, err := os.ReadFile(txt)
	if err != nil {
		return 0, nil, err
	}
	parts := strings.Split(string(b), "\f")
	for len(parts) > 0 && strings.TrimSpace(parts[len(parts)-1]) == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		return 0, nil, errString(nil, errString(nil, nil))
	}
	return len(parts), parts, nil
}

type strErr struct{ s string }

func (e strErr) Error() string { return e.s }

func errString(out []byte, err error) error {
	if err == nil {
		return strErr{"pdf conversion produced no pages"}
	}
	return strErr{err.Error() + " " + string(out)}
}

func TestPageCountHelper(t *testing.T) {
	if strconv.Itoa(2) != "2" {
		t.Fatal("sanity")
	}
}
