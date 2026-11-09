package ooxml

import (
	"encoding/xml"
	"strconv"
	"strings"
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

type xmlAppProperties struct {
	XMLName xml.Name `xml:"Properties"`
	Company string   `xml:"Company"`
	Manager string   `xml:"Manager"`
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
	if props.Revision != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(props.Revision)); err == nil {
			p.Revision = n
		}
	}

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

func (r *docxReader) readAppProperties() error {
	data, err := r.readZipFile("docProps/app.xml")
	if err != nil {
		return err
	}
	var props xmlAppProperties
	if err := xml.Unmarshal(data, &props); err != nil {
		return err
	}
	if props.Company != "" {
		r.doc.Properties.Company = props.Company
	}
	if props.Manager != "" {
		r.doc.Properties.Manager = props.Manager
	}
	return nil
}
