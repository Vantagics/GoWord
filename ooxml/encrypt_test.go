package ooxml

import (
	"errors"
	"strings"
	"testing"

	"github.com/VantageDataChat/GoWord/document"
)

func TestDecryptStandardAndAgileFixtures(t *testing.T) {
	sol, err := ReadWithPassword("../testdata/word/encrypted/bug53475-password-is-solrcell.docx", "solrcell")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sol.ExtractText(), "password protected") {
		t.Fatalf("solrcell text = %q", sol.ExtractText())
	}
	if _, err := Read("../testdata/word/encrypted/bug53475-password-is-solrcell.docx"); !errors.Is(err, document.ErrEncrypted) {
		t.Fatalf("missing password: %v", err)
	}
	if _, err := ReadWithPassword("../testdata/word/encrypted/bug53475-password-is-solrcell.docx", "nope"); !errors.Is(err, document.ErrWrongPassword) {
		t.Fatalf("wrong password: %v", err)
	}

	pass, err := ReadWithPassword("../testdata/word/encrypted/bug53475-password-is-pass.docx", "pass")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(pass.ExtractText()) == "" {
		t.Fatal("agile document produced no text")
	}
	for _, name := range []string{
		"Encrypted_LO_Standard_abc.docx",
		"Encrypted_MSO2007_abc.docx",
		"Encrypted_MSO2010_abc.docx",
		"Encrypted_MSO2013_abc.docx",
	} {
		doc, err := ReadWithPassword("../testdata/word/encrypted/"+name, "abc")
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if strings.TrimSpace(doc.ExtractText()) == "" {
			t.Errorf("%s: empty text", name)
		}
	}
}

func TestCJKNumberingMarkers(t *testing.T) {
	doc, err := Read("../testdata/word/libreoffice/cjklist30.docx")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.ExtractText(), "甲.") || !strings.Contains(doc.ExtractText(), "癸.") {
		t.Fatalf("traditional stems missing: %q", doc.ExtractText()[:80])
	}
	doc, err = Read("../testdata/word/libreoffice/cjklist44.docx")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.ExtractText(), "一零.") || !strings.Contains(doc.ExtractText(), "一一.") {
		t.Fatalf("korean digital missing: %q", doc.ExtractText()[:80])
	}
}

func TestEncryptRoundTrip(t *testing.T) {
	doc := document.New()
	doc.AddSection().AddText("Secret meeting notes", nil, nil)
	doc.SetPassword("correct horse")
	out, err := WriteToBytes(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !isCompoundFile(out) {
		t.Fatal("encrypted save was not a compound file")
	}
	if _, err := ReadFromBytes(out); !errors.Is(err, document.ErrEncrypted) {
		t.Fatalf("unprotected read: %v", err)
	}
	again, err := ReadFromBytesWithPassword(out, "correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(again.ExtractText(), "Secret meeting notes") {
		t.Fatalf("text = %q", again.ExtractText())
	}
	if again.Password() != "correct horse" {
		t.Fatalf("password not retained: %q", again.Password())
	}
	// Saving again keeps the password and the text.
	out2, err := WriteToBytes(again)
	if err != nil {
		t.Fatal(err)
	}
	third, err := ReadFromBytesWithPassword(out2, "correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(third.ExtractText(), "Secret meeting notes") {
		t.Fatalf("second save = %q", third.ExtractText())
	}
}
