package ooxml

import (
	"encoding/xml"
	"time"
)

type xmlCoreProperties struct {
	XMLName        xml.Name `xml:"coreProperties"`
	Title          string   `xml:"title"`
	Subject        string   `xml:"subject"`
	Creator        string   `xml:"creator"`
	Keywords       string   `xml:"keywords"`
	Description    string   `xml:"description"`
	Category       string   `xml:"category"`
	LastModifiedBy string   `xml:"lastModifiedBy"`
	Revision       string   `xml:"revision"`
	Created        string   `xml:"created"`
	Modified       string   `xml:"modified"`
}

func (r *docxReader) readCoreProperties() error {
	data, err := r.readZipFile("docProps/core.xml")
	if err != nil {
		return err
	}

	var props xmlCoreProperties
	if err := xml.Unmarshal(data, &props); err != nil {
		return err
	}

	p := &r.doc.Properties
	p.Title = props.Title
	p.Subject = props.Subject
	p.Creator = props.Creator
	p.Keywords = props.Keywords
	p.Description = props.Description
	p.Category = props.Category
	p.LastModifiedBy = props.LastModifiedBy

	if props.Created != "" {
		if t, err := time.Parse(time.RFC3339, props.Created); err == nil {
			p.Created = t
		}
	}
	if props.Modified != "" {
		if t, err := time.Parse(time.RFC3339, props.Modified); err == nil {
			p.Modified = t
		}
	}

	return nil
}
