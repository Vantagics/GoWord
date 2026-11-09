package ooxml

import (
	"encoding/xml"
	"fmt"
	"path"
	"strings"
)

type xmlRelationships struct {
	XMLName      xml.Name          `xml:"Relationships"`
	Relationship []xmlRelationship `xml:"Relationship"`
}

type xmlRelationship struct {
	ID         string `xml:"Id,attr"`
	Type       string `xml:"Type,attr"`
	Target     string `xml:"Target,attr"`
	TargetMode string `xml:"TargetMode,attr"`
}

func (r *docxReader) readRels() error {
	if err := r.loadAllRels(); err != nil {
		return err
	}
	if r.docPart == "" {
		r.docPart = "word/document.xml"
	}
	r.usePart(r.docPart)
	return nil
}

func (r *docxReader) loadAllRels() error {
	if r.relsByPart == nil {
		r.relsByPart = map[string]map[string]relInfo{}
	}
	for _, f := range r.zr.File {
		if !strings.HasSuffix(f.Name, ".rels") {
			continue
		}
		data, err := r.readZipFile(f.Name)
		if err != nil {
			continue
		}
		var rels xmlRelationships
		if err := xml.Unmarshal(data, &rels); err != nil {
			return fmt.Errorf("reading relationships %s: %w", f.Name, err)
		}
		owner := relsOwner(f.Name)
		m := map[string]relInfo{}
		for _, rel := range rels.Relationship {
			m[rel.ID] = relInfo{ID: rel.ID, Type: rel.Type, Target: rel.Target, TargetMode: rel.TargetMode}
		}
		r.relsByPart[owner] = m
	}
	return nil
}

// relsOwner maps "word/_rels/document.xml.rels" to "word/document.xml"
// and "_rels/.rels" to "".
func relsOwner(relsPath string) string {
	if relsPath == "_rels/.rels" {
		return ""
	}
	dir := path.Dir(relsPath) // word/_rels
	parent := path.Dir(dir)   // word
	base := path.Base(relsPath)
	base = strings.TrimSuffix(base, ".rels")
	if parent == "." {
		return base
	}
	return path.Join(parent, base)
}

func (r *docxReader) readZipFile(name string) ([]byte, error) {
	for _, f := range r.zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return readAll(rc)
		}
	}
	return nil, errNotFound
}

var errNotFound = &zipFileNotFoundError{}

type zipFileNotFoundError struct{}

func (e *zipFileNotFoundError) Error() string { return "file not found in zip" }
