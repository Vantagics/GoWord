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

## Round 2 — PHPWord reader documents

Samples added under `testdata/word/phpword/` (LGPL-3.0): `reader.docx`, `reader-2011.docx`, `reader-comments.docx`, `reader-formula.docx`, `reader-styles.docx`, `Sample_07_TemplateCloneRow.docx`, `Sample_11_ReadWord2007.docx`, `Sample_23_TemplateBlock.docx`, `Sample_42_TemplateSetCheckbox.docx`, `without_table_macros.docx`.

Harness on this batch: 10 files. Round-trip opened cleanly in all four reference tools (10/10). Read mismatches: 1 file each for LibreOffice, python-docx, and POI (`reader-comments.docx`, tracked-change display — goword keeps the accepted text `[test]`, LibreOffice keeps the deleted `T` as `[Ttest]`). pandoc mismatch on `reader-formula.docx` is TeX (`\frac{\pi}{2}`) versus Office math characters.

### Fixed

- OMML fractions, superscripts, and subscripts are linearized (`π/2` instead of `π2`) when text is extracted from equations.

### Remaining

- Same gaps as round 1, plus tracked-change rendering does not try to match LibreOffice's combined deleted+inserted string.

## Round 3 — content controls (Open XML SDK)

Samples added under `testdata/word/openxmlsdk/` (MIT): block and run content controls, a content control inside a table cell, data binding, a comment, a header/footer, and a paragraph wrapped in an SDT.

### Found

`sdtCell.docx` stores each cell as `w:tr/w:sdt/w:sdtContent/w:tc`. The row reader skipped that control, so LibreOffice, pandoc, and POI all saw the sentence "This is an sdtCell" and goword saw only whitespace (read and round-trip recall 0). A few ISO-strict SDK files were not kept: pandoc, python-docx, and POI refuse the original package.

### Fixed

Cell-level structured document tags are unwrapped into table cells, so the text is read and survives a rewrite.

### Remaining

LibreOffice is still about one word short on `SDT.docx` and `SDT_Run.docx` (recall 0.875). POI still under-reads `comment.docx` relative to goword, which also emits the comment body.

## Round 4 — more Apache POI documents

Samples added in `testdata/word/poi/`: `WithGIF.docx`, `ThreeColHeadFoot.docx`, `EmbeddedDocument.docx`, `documentProtection_trackedChanges_no_password.docx`, `SimpleHeadThreeColFoot.docx`, `Tika-792.docx`, `bug66312.docx`.

Six of the seven round-tripped through every reference tool. `Tika-792.docx` is almost entirely tracked deletions (`s` and `.`). LibreOffice and POI showed that text and goword showed blank lines (recall 0).

### Fixed

Deleted and moved-from text (`w:delText`) is included in extracted text, in document order, while the original `w:del` markup is still what gets written back. `Tika-792.docx` now yields `s.`. The PHPWord comment sample's accepted-plus-deleted string is `[Ttest]`, which matches LibreOffice.

### Remaining

A trailing `b` in `Tika-792.docx` is still inside a nested move that this pass does not surface. Comment bodies are still emitted in addition to the anchor text, so tools that ignore comments report extra words.

## Round 5 — more python-docx documents

Samples added in `testdata/word/python-docx/`: `blk-containing-table.docx`, `blk-paras-and-tables.docx`, `fnt-color.docx`, `par-alignment.docx`, `par-hlink-frags.docx`, `run-char-style.docx`, `sty-known-styles.docx`, `tbl-2x2-table.docx`, `tbl-having-applied-style.docx`, `txt-font-highlight-color.docx` (MIT).

Harness: 10/10 read and round-trip matches for LibreOffice, pandoc, python-docx, and POI at the 0.90 recall bar. No code change; the round-1 reader already covered these tables, styles, colors, and hyperlink fragments.

### Remaining

Unchanged from round 4.

## Round 6 — docx4j parts and footnote numbers

Samples added in `testdata/word/docx4j/` (Apache-2.0): `comments-one.docx`, `header-simple.docx`, `image-png1.docx`, `paragraph-two.docx`, `NumberingIndents.docx`.

`header-simple.docx`, `paragraph-two.docx`, and `NumberingIndents.docx` match all four tools. `image-png1.docx` has no words; pandoc emits a stand-in description, so recall against pandoc is 0. `comments-one.docx` is 0.71 versus POI because goword also prints the comment body.

### Fixed

Footnote and endnote reference numbers are written into the extracted text next to the anchor. `mammoth/footnotes.docx` is now `Ouch1.2` plus the note body, which covers the LibreOffice text that previously scored 0.

### Remaining

pandoc and POI still tokenize note bodies differently from the anchor line. VML text boxes stored inside an escaped drawing attribute (`docx4j/vml-textbox.docx`) still do not round-trip into LibreOffice text.

`FldSimple.docx` round-trip recall against LibreOffice is low because the field is `FILENAME`: LibreOffice refreshes it to the new file name (`out.docx` instead of `FldSimple.docx`). The cached result is preserved; the reference tool re-evaluates the field.

## Round 7 — text effects, fonts, duplicate hyperlinks

Samples added in `testdata/word/docx4j/`: `w14-texteffects.docx`, `hyperlink-dupe.docx`, `Mac_OSX_Fonts.docx`, `comments-two.docx`, `images-png2.docx`.

`Mac_OSX_Fonts.docx` and `w14-texteffects.docx` match all four tools. `images-png2.docx` is a picture with no words (pandoc invents a description). `comments-two.docx` is 0.58 versus POI for the same comment-body reason as round 6. `hyperlink-dupe.docx` was 0.50 for every tool, read and round-trip: a TOC hyperlink is `H1` plus a tab plus a PAGEREF result `1`, and goword concatenated that into `H11`.

### Fixed

Tabs inside a hyperlink are kept, so the extracted text is `H1`, tab, `1` instead of `H11`.

### Remaining

Picture-only files still disagree with pandoc's image description. Comment bodies are still included.

## Round 8 — PHPWord templates

Samples added in `testdata/word/phpword/`: `Sample_40_TemplateSetComplexValue.docx`, `Sample_41_TemplateSetChart.docx`, `Sample_44_ExtractVariables.docx`, `blank.docx`, `clone-delete-block.docx` (LGPL-3.0).

Harness: 5/5 read and round-trip matches on LibreOffice, pandoc, python-docx, and POI. No code change.

### Remaining

Unchanged from round 7.

## Round 9 — POI bug documents and SDT rows

Samples added in `testdata/word/poi/`: `61470.docx` (East Asian ruby text), `65099.docx`, `Bug66263-paragraph.docx`, `Bug66263-table.docx`, `Bug66988.docx`.

`65099.docx` and `Bug66988.docx` match. `61470.docx` is hiragana plus kanji (`とうきょう` / `東京`); goword keeps both runs, and LibreOffice/POI segment them so word-token recall is 0. `Bug66263-paragraph.docx` is 0.55 versus LibreOffice. `Bug66263-table.docx` was empty in goword and `SDT Cell 1` / `SDT Cell 2` in LibreOffice.

### Fixed

A content control may wrap a whole row (`w:tbl/w:sdt/w:sdtContent/w:tr`). Those rows are now read. `Bug66263-table.docx` extracts `SDT Cell 1` and `SDT Cell 2`.

### Remaining

Ruby text is still one concatenated string. `FILENAME` fields change when LibreOffice opens the round-trip under a new name.

## Round 10 — final POI batch and full remeasure

Samples added in `testdata/word/poi/`: `57312.docx`, `58618.docx`, `59030.docx`, `59378.docx`, `bug65738.docx`. All five match every reference tool on read and round-trip.

The harness was then run on the whole corpus (131 documents).

| Tool | Read ok | Round-trip ok |
| --- | --- | --- |
| LibreOffice | 121/131 | 124/131 |
| pandoc | 117/131 | 128/131 |
| python-docx | 122/131 | 130/131 |
| Apache POI XWPF | 115/131 | 128/131 |

No further code change in this round. Compared with round 1's before column (70 files on the old main: LibreOffice read 56 ok, POI 47 ok), the library now agrees with the reference readers on a larger and harder set.

### Remaining gaps

- Strict OOXML originals (`mammoth/strict-format.docx`) are read by goword but not by pandoc, python-docx, or POI. The goword rewrite of that file is readable.
- `docx4j/2010-mcAlternateContent.docx` does not open in LibreOffice.
- VML text boxed inside an escaped shape attribute (`docx4j/vml-textbox.docx`) still loses its words after a round-trip when LibreOffice or POI reads it.
- `FILENAME` fields are re-evaluated by LibreOffice to the new file name.
- Comment bodies are included in goword's text and omitted by several reference extractors, so comment files look like mismatches.
- Footnote bodies and anchors are not lined up the same way by pandoc and POI.
- East Asian ruby (`poi/61470.docx`) is kept as both reading and base text in one string; other tools segment it differently.
- Equation plain text is not TeX (`docx4j/2010-sample1.docx`, `phpword/reader-formula.docx`).
- Numbering overrides (`poi/NumberingWOverrides.docx`) and a nested move in `poi/Tika-792.docx` are only partly reflected.
- Picture-only files disagree with pandoc's generated image description.

