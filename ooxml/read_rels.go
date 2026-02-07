package ooxml

import (
	"encoding/xml"
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
	r.rels = make(map[string]string)

	data, err := r.readZipFile("word/_rels/document.xml.rels")
	if err != nil {
		return nil // No rels file is OK
	}

	var rels xmlRelationships
	if err := xml.Unmarshal(data, &rels); err != nil {
		return err
	}

	for _, rel := range rels.Relationship {
		r.rels[rel.ID] = rel.Target
	}

	return nil
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
