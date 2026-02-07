package ooxml

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"

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

// ReadFromBytes reads a .docx from byte data.
func ReadFromBytes(data []byte) (*document.Document, error) {
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

type docxReader struct {
	zr   *zip.Reader
	doc  *document.Document
	rels map[string]string // relID -> target
}

func (r *docxReader) read() (*document.Document, error) {
	// Read relationships
	if err := r.readRels(); err != nil {
		return nil, err
	}

	// Read core properties
	if err := r.readCoreProperties(); err != nil {
		// Non-fatal: properties are optional
		_ = err
	}

	// Read main document
	if err := r.readDocument(); err != nil {
		return nil, err
	}

	return r.doc, nil
}
