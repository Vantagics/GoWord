package document

// SetLoading marks the document as being filled from an existing package.
// Style and numbering registration during load does not mark those parts dirty,
// so a later save can keep the original styles, numbering, and settings.
func (d *Document) SetLoading(loading bool) {
	d.loading = loading
}

// Loading reports whether SetLoading(true) is in effect.
func (d *Document) Loading() bool {
	return d.loading
}

// StylesDirty reports whether named styles were changed through the API.
func (d *Document) StylesDirty() bool { return d.stylesDirty }

// NumberingDirty reports whether numbering definitions were changed through the API.
func (d *Document) NumberingDirty() bool { return d.numberingDirty }

// SettingsDirty reports whether document settings were changed through the API.
func (d *Document) SettingsDirty() bool { return d.settingsDirty }

// NotesDirty reports whether notes or comments were changed through the API.
func (d *Document) NotesDirty() bool { return d.notesDirty }

// SetStrict requests a strict OOXML package on the next save.
// The default remains transitional WordprocessingML, which is what Word,
// LibreOffice, and the other reference readers open.
func (d *Document) SetStrict(strict bool) {
	if d != nil {
		d.strict = strict
	}
}

// Strict reports whether the next save should use strict OOXML namespaces.
func (d *Document) Strict() bool {
	return d != nil && d.strict
}

// SetPassword makes the next save an encrypted package (ECMA-376 agile).
// An empty password writes a normal .docx. Opening an encrypted file with
// OpenWithPassword stores the password so a later Save keeps it encrypted.
func (d *Document) SetPassword(password string) {
	if d != nil {
		d.password = password
	}
}

// Password returns the password used to open or to encrypt this document.
func (d *Document) Password() string {
	if d == nil {
		return ""
	}
	return d.password
}

// StoreOriginalPart records a package part loaded from an existing .docx.
func (d *Document) StoreOriginalPart(name string, data []byte) {
	if d.originalParts == nil {
		d.originalParts = make(map[string][]byte)
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	d.originalParts[name] = cp
}

// OriginalPart returns a part previously stored by the reader.
func (d *Document) OriginalPart(name string) ([]byte, bool) {
	if d.originalParts == nil {
		return nil, false
	}
	b, ok := d.originalParts[name]
	return b, ok
}

// OriginalPartNames returns the stored part names in no particular order.
func (d *Document) OriginalPartNames() []string {
	if len(d.originalParts) == 0 {
		return nil
	}
	names := make([]string, 0, len(d.originalParts))
	for name := range d.originalParts {
		names = append(names, name)
	}
	return names
}

// SetMainPart records the package path of the main document part
// (normally "word/document.xml").
func (d *Document) SetMainPart(name string) {
	d.mainPart = name
}

// MainPart returns the main document part path. Empty means word/document.xml.
func (d *Document) MainPart() string {
	if d.mainPart == "" {
		return "word/document.xml"
	}
	return d.mainPart
}

// RegisterFootnote adds a footnote loaded from a package without marking notes dirty
// and without inserting a reference into the body.
func (d *Document) RegisterFootnote(fn *Footnote) {
	if fn == nil {
		return
	}
	fn.doc = d
	d.footnotes = append(d.footnotes, fn)
	if fn.ID >= d.nextFootnoteID {
		d.nextFootnoteID = fn.ID + 1
	}
}

// RegisterEndnote adds an endnote loaded from a package.
func (d *Document) RegisterEndnote(en *Endnote) {
	if en == nil {
		return
	}
	en.doc = d
	d.endnotes = append(d.endnotes, en)
	if en.ID >= d.nextEndnoteID {
		d.nextEndnoteID = en.ID + 1
	}
}

// RegisterComment adds a comment loaded from a package.
func (d *Document) RegisterComment(c *Comment) {
	if c == nil {
		return
	}
	c.doc = d
	d.comments = append(d.comments, c)
	if c.ID >= d.nextCommentID {
		d.nextCommentID = c.ID + 1
	}
}
