# Word compatibility rounds

Differential harness: `make diff` (`scripts/diffharness.py`).

Reference tools, all optional: LibreOffice Writer headless (`soffice`), pandoc, python-docx, Apache POI XWPF (`scripts/diffharness/PoiExtract.java`, jars not committed). A missing tool is skipped. Recall is the fraction of the reference reader's words that goword (or a goword round-trip re-read by that tool) still contains.

The "before" column is goword at `main` before this work, measured on the round-1 corpus. Later rounds add files and compare against the previous round.

## Round 1 — OOXML read/write and first corpus

Samples (70), licenses in `testdata/word/SOURCES.md`:

- python-docx test documents (MIT)
- Apache POI `test-data/document` (Apache-2.0)
- mammoth.js test documents (BSD-2-Clause)
- docx4j samples (Apache-2.0), including strict-namespace units, math, VML text boxes, and a package whose main part is `word/document22.xml`

### Read mismatches against reference tools (ok / mismatch / error)

| Tool | Before (main) | After round 1 |
| --- | --- | --- |
| LibreOffice | 56 / 12 / 1 | 61 / 8 / 1 |
| pandoc | 55 / 13 / 1 | 60 / 9 / 1 |
| python-docx | 54 / 14 / 1 | 66 / 3 / 1 |
| Apache POI XWPF | 47 / 21 / 1 | 60 / 9 / 1 |

Round-trip (reference opens the goword rewrite and keeps ≥90% of the words it saw in the original): LibreOffice 67/70, pandoc 69/70, POI 68/70, python-docx 70/70.

### Fixed

- Schema order for `rPr`, `pPr`, `tblPr`, `tcPr`, `sectPr`, numbering (`abstractNum` before `num`), settings, and extended properties. `mc:Ignorable` on the document root.
- Multiple sections, headers and footers, pictures, nested tables, merged cells, hyperlinks, fields, bookmarks, comments, footnotes, endnotes, styles, and numbering.
- Original package parts kept on an unmodified save; unknown markup sliced from the original XML and drawings wrapped in `w:r` so LibreOffice still exports text boxes.
- Tabs and list markers included in extracted text.
- Strict relationship types are not copied onto transitional parts, so POI can open a rewrite of a strict file.

### Remaining

- Footnote and endnote reference numbers versus note body text (LibreOffice, pandoc, and POI still disagree on several note files).
- Equation text is only partly visible to pandoc (`docx4j/2010-sample1.docx`).
- VML text box round-trip: LibreOffice and POI lose the words (`docx4j/vml-textbox.docx`).
- `FldSimple.docx` round-trip drops about half the words LibreOffice saw.
- `2010-mcAlternateContent.docx` does not open in LibreOffice at all (original or rewrite).
- python-docx, pandoc, and POI cannot read the original strict-namespace file; the goword rewrite of that file is readable.
- Numbering overrides and some form fields are only partly reflected in extracted text.
