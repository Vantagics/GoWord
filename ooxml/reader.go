package ooxml

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/VantageDataChat/GoWord/document"
)

// Read reads a .docx file and returns a Document.
func Read(path string) (*document.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return ReadFromBytes(data)
}

// ReadWithPassword reads a .docx, decrypting an ECMA-376 package when needed.
func ReadWithPassword(path, password string) (*document.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return ReadFromBytesWithPassword(data, password)
}

// ReadFromBytesWithPassword reads a .docx. Encrypted packages (ECMA-376
// standard or agile) are decrypted with password first.
func ReadFromBytesWithPassword(data []byte, password string) (*document.Document, error) {
	if isCompoundFile(data) {
		plain, err := decryptOfficePackage(data, password)
		if err != nil {
			return nil, err
		}
		doc, err := readZipBytes(plain)
		if err != nil {
			return nil, err
		}
		doc.SetPassword(password)
		return doc, nil
	}
	return readZipBytes(data)
}

// ReadFromBytes reads a .docx from byte data.
func ReadFromBytes(data []byte) (*document.Document, error) {
	return ReadFromBytesWithPassword(data, "")
}

func readZipBytes(data []byte) (*document.Document, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("opening zip: %w", err)
	}

	reader := &docxReader{
		zr:  r,
		doc: document.New(),
	}

	return reader.read()
}

type relInfo struct {
	ID         string
	Type       string
	Target     string
	TargetMode string
}

type docxReader struct {
	zr           *zip.Reader
	doc          *document.Document
	rels         map[string]string // relID -> target for the part currently being read
	relMode      map[string]string // relID -> TargetMode for the current part
	relsByPart   map[string]map[string]relInfo
	docPart      string
	curPart      string
	footnotes    map[int]*document.Footnote
	endnotes     map[int]*document.Endnote
	comments     map[int]*document.Comment
	imagesByName map[string]*document.Image
	// caps tracks the original XML text behind each decoder so captured
	// elements can be sliced out byte-for-byte instead of re-serialized.
	caps []xmlSrc
}

type xmlSrc struct {
	dec *xml.Decoder
	src string
}

func (r *docxReader) read() (*document.Document, error) {
	r.relsByPart = map[string]map[string]relInfo{}
	r.footnotes = map[int]*document.Footnote{}
	r.endnotes = map[int]*document.Endnote{}
	r.comments = map[int]*document.Comment{}
	r.imagesByName = map[string]*document.Image{}
	r.doc.SetLoading(true)
	defer r.doc.SetLoading(false)

	r.captureOriginalParts()
	if err := r.loadAllRels(); err != nil {
		return nil, err
	}
	r.docPart = r.resolveMainDocument()
	r.doc.SetMainPart(r.docPart)
	r.usePart(r.docPart)

	if err := r.readCoreProperties(); err != nil {
		_ = err
	}
	_ = r.readAppProperties()
	_ = r.readSettingsPart()
	_ = r.readStyles()
	_ = r.readNumberingPart()
	_ = r.readNotes("word/footnotes.xml", true)
	_ = r.readNotes("word/endnotes.xml", false)
	_ = r.readCommentsPart()

	if err := r.readDocument(); err != nil {
		return nil, err
	}
	return r.doc, nil
}

func (r *docxReader) captureOriginalParts() {
	for _, f := range r.zr.File {
		if f.FileInfo().IsDir() || strings.HasSuffix(f.Name, "/") {
			continue
		}
		data, err := r.readZipFile(f.Name)
		if err != nil {
			continue
		}
		r.doc.StoreOriginalPart(f.Name, data)
	}
}

func (r *docxReader) usePart(part string) {
	r.curPart = part
	r.rels = map[string]string{}
	r.relMode = map[string]string{}
	for id, rel := range r.relsByPart[part] {
		r.rels[id] = rel.Target
		r.relMode[id] = rel.TargetMode
	}
}
