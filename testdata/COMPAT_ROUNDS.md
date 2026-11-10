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

## Gap fixes after round 10

No new sample files. The harness was run again on the same 131 documents (`go test ./...` green). Compared with round 10:

| Tool | Read ok | Round-trip ok |
| --- | --- | --- |
| LibreOffice | 125/131 | 127/131 |
| pandoc | 119/131 | 128/131 |
| python-docx | 122/131 | 130/131 |
| Apache POI XWPF | 116/131 | 130/131 |

Nothing that passed in round 10 regressed.

### Fixed

- `docx4j/vml-textbox.docx`: a VML `w:pict` was written as a direct child of `w:p` (`<w:pict` shares its first characters with `<w:p`). It is now inside `w:r`. LibreOffice and POI both export `v:textbox/w:txbxContent` after a round-trip.
- `docx4j/2010-mcAlternateContent.docx`: the original package does not open in LibreOffice because `word/comments.xml` contains `mc:AlternateContent`. The rewrite now drops those elements (the comment sentence stays). LibreOffice converts the rewrite to `Hello.` The original file is unchanged and still does not open.
- `poi/61470.docx`: `w:ruby` is a `Ruby` value with `Base` and `Reading`. `ExtractText` returns the base (`東京`), which is what LibreOffice exports. `Document.Rubies()` and `ExtractOptions.Ruby` (`reading` or `both`) expose the pronunciation. The original `w:ruby` markup is written back, so POI's `東京 (とうきょう)` survives a round-trip.
- `poi/NumberingWOverrides.docx`: `lvlOverride` and `startOverride` replace the abstract level for that `numId`. Extracted markers are `1`, `A`, `B`, `A`, `B`, `2`, which is what pandoc emits. A paragraph style that only carries `w:numPr` (python-docx `num-having-numbering-part.docx`) now contributes its list marker too.
- `poi/Tika-792.docx`: text inside `moveFrom` is included, including an insertion nested in the move. LibreOffice's `s.` / `b` and POI's `s` / `b` are both covered. The `moveFrom` / `moveTo` elements are still written back.
- `poi/shapes-with-text.docx`: text from adjacent drawings is separated so LibreOffice's `Floating text box` stays three words.
- Extraction: `ExtractText` is unchanged (`ExtractDefaults`: headers, footers, note bodies, and comment bodies). `ExtractTextWith` can omit each of those. `NoteAnchor: "separate"` puts a space before a note number (`prostoy 1`) so the result matches pandoc; the default stays glued (`prostoy1`), which is what LibreOffice exports. `MathFormat: "latex"` runs `OMMLToLatex` (`\frac`, `\sqrt`) without changing the default linear extract.
- `Document.SetStrict(true)` writes strict WordprocessingML and relationship URIs (`http://purl.oclc.org/ooxml/...`). The default save stays transitional.

### Not goword defects

These stay mismatches on purpose. Changing goword to imitate them would fight the file or another reference reader.

- LibreOffice re-evaluates a `FILENAME` field, so a round-trip saved as `out.docx` exports `out.docx` while the cached result is still `FldSimple.docx`.
- pandoc invents an image description for pictures that have no alt text (`image-png1.docx`, `images-png2.docx`, `VariousPictures.docx`).
- pandoc, python-docx, and POI refuse the original strict package (`mammoth/strict-format.docx`). goword reads it. A strict save is available only when requested; tools that reject strict input will still reject it.
- LibreOffice's list labels on `NumberingWOverrides.docx` repeat the paragraph text (`B B`, `C C`, `4 4`). The numbering part says upper-letter starting at 1 for numId 22 and a continuing decimal for numId 19 (`A`, `B`, then `2`). pandoc agrees with the numbering part.
- LibreOffice deletes spaces at content-control boundaries (`SDT.docx`, `SDT_Run.docx`, `Bug66263-paragraph.docx`), so `SDT` and `Run` become one token `SDTRun`. POI keeps the spaces, and so does goword.
- POI's comment extractor adds `Comment by <author>`, which is not document text.
- python-docx repeats merged cells (`tbl-cell-props.docx`) and repeats a header that the package stores once (`hdr-header-footer.docx`).
- Note tools disagree with each other: LibreOffice often omits the note body and glues the anchor (`Ouch1.2`); pandoc writes `[1]` and repeats the number under the body; python-docx often skips the notes. `ExtractTextWith` is the way to line up with one of those. One default string cannot match all three.
- POI's plain text for ruby appends the reading in parentheses. The default extract is the base text; the reading is on `Ruby.Reading`.
- pandoc's equation text is TeX. The default extract stays the linear Office form; `MathFormat: "latex"` is the TeX form.

## Round 11 — encrypted packages, macros, templates, charts, CJK

Samples, licenses in `testdata/word/SOURCES.md`:

- `encrypted/`: POI `bug53475-password-is-solrcell.docx` (standard, password `solrcell`) and `bug53475-password-is-pass.docx` (agile, password `pass`); LibreOffice `Encrypted_LO_Standard_abc.docx`, `Encrypted_MSO2007_abc.docx`, `Encrypted_MSO2010_abc.docx`, `Encrypted_MSO2013_abc.docx` (password `abc`).
- `poi/`: `45690.docm`, `60158.docm`, `60316.dotx`, `60316b.dotx`, document-protection files, `ComplexNumberedLists.docx`, `HeaderFooterUnicode.docx`.
- `libreoffice/`: charts, an embedded spreadsheet, SmartArt, vertical text (`dkvert.docx`, `tbrl-textbox.docx`, `btlr-textbox.docx`), CJK numbering, grouped text boxes, and OLE objects.

Encrypted files are covered by `go test` (`TestDecryptStandardAndAgileFixtures`, `TestEncryptRoundTrip`). The differential harness skips `testdata/word/encrypted/` because LibreOffice, pandoc, python-docx, and POI are not given the password. `.docm` and `.dotx` are included.

Harness on the whole corpus (158 packages that are not encrypted):

| Tool | Read ok | Round-trip ok |
| --- | --- | --- |
| LibreOffice | 152/158 | 154/158 |
| pandoc | 137/158 | 155/158 |
| python-docx | 144/158 | 156/158 |
| Apache POI XWPF | 143/158 | 157/158 |

No file that passed in the previous full run regressed. Every new LibreOffice fixture matches LibreOffice on read and round-trip, including charts, SmartArt, vertical text, and CJK lists.

### Fixed

- ECMA-376 encrypted packages. `OpenWithPassword` / `ReadFromBytesWithPassword` decrypt standard (Office 2007, AES-ECB, SHA-1, 50,000 spins) and agile (Office 2010+, AES-CBC) files. A missing password returns `ErrEncrypted`; a wrong password returns `ErrWrongPassword`. `SetPassword` writes an agile package (AES-256, SHA-512, HMAC) inside a compound file, including the DataSpaces tree Word looks for. Opening with a password keeps it for the next save.
- CJK list markers: `ideographTraditional` (甲乙丙…), `ideographZodiac` (子丑寅…), `ideographLegalTraditional` (壹貳參…), `taiwaneseCountingThousand` / Chinese counting (一十十一), and `koreanDigital2` (一零, 一一).
- Text in adjacent paragraphs inside a drawing is separated by a newline, so a grouped text box does not glue `first` and `second` into one word.

### Not defects

- pandoc still prints CJK lists as western digits. LibreOffice and goword use the numbering format.
- python-docx rejects `.docm` and `.dotx` because the content type is not a Word document. pandoc did not finish those four files within the harness timeout. LibreOffice and POI read them, and goword does too.
- Chart and SmartArt words live in the drawing or the embedded workbook, not in `w:t`. Round-trip keeps those parts; plain-text export of the chart series is empty for LibreOffice as well.
- docx.js and officegen publish generators, not redistributable `.docx` binaries. Aspose and Syncfusion sample licenses do not allow committing their outputs. Those sources were not added as fixtures.

## Round 12 — revisions, fields, RTL, charts, and diagram text

New sources, both permissive: Open XML PowerTools (MIT) under `testdata/word/powertools/`, and Apache Tika (Apache-2.0) under `testdata/word/tika/`. The set covers formatting revisions (`rPrChange`, `pPrChange`, `tblPrChange`), tracked moves, Arabic and Hebrew, a table of contents and `MERGEFORMAT` fields, sections, content controls, Chinese counting lists, charts, SmartArt, mail merge, altChunk, a macro-enabled document, and a template.

Harness on the whole non-encrypted corpus (198 packages):

| Tool | Read ok | Round-trip ok |
| --- | --- | --- |
| LibreOffice | 191/198 | 194/198 |
| pandoc | 174/198 | 195/198 |
| python-docx | 180/198 | 195/198 |
| Apache POI XWPF | 179/198 | 196/198 |

No file that passed in round 11 regressed. The Arabic PowerTools document round-trips: a prefix used only on preserved markup (`xmlns:pt14`) is copied onto the document root, so POI and LibreOffice can open the rewrite.

### Fixed

- `ExtractOptions.IncludeGraphicText` appends chart category and series names and SmartArt labels from `word/charts` and `word/diagrams`. `Document.GraphicText` returns the same text. `ExtractText` stays unchanged, so numeric chart caches and diagram text do not appear unless asked. Duplicate labels that exist in both the diagram data and the drawing are emitted once.
- Text stored in `w:t` before an unclosed field separator is kept. `testWORD_unclosedFields.docx` now includes `junk`.
- Chinese counting markers go through 9999 (`十`, `二十一`, `一百零一`, `一千零一十`), which is what the PowerTools Chinese list fixtures use.
- A 2,000-paragraph document is written and read back in the test suite.

Formatting revisions are preserved: `rPrChange`, `pPrChange`, and `tblPrChange` are still in the rewritten `document.xml`.

### Not defects

- LibreOffice re-evaluates the date field in `CA012-fldSimple.docx` to the day it opens the file (`11/10/26`). The cached result is `4/24/2016`.
- python-docx and POI delete a soft hyphen, so `optional hyphen` becomes the single token `optionalhyphen`. LibreOffice keeps the soft hyphen and agrees with goword.
- pandoc prints Windows cache paths as image descriptions in `testWORD_embedded_pics.docx`.
- POI prints the `NOTEREF` field instruction inside `RP043-MERGEFORMAT-Field-Code.docx` and appends the ruby reading on `testWORD_phonetic.docx`. The default extract is the result and the ruby base.
- pandoc times out on `.docm` and `.dotx`, and python-docx rejects those content types. LibreOffice and POI read them.

## Round 13 — editing APIs, glossary, embedded fonts, large files

### Generated documents

Permissive real samples were not available for glossary parts, embedded fonts, or multi-megabyte files, so the tests build them.

- Glossary: `AddBuildingBlock` writes `word/glossary/document.xml` (`w:glossaryDocument` / `w:docPart`). A round trip returns the name and text. LibreOffice, POI, and python-docx open the package.
- Embedded fonts: `EmbedFont` obfuscates the first 32 bytes with the fontKey GUID the way ECMA-376 specifies (mixed-endian GUID, reversed, XOR). Reading the package deobfuscates back to the original bytes.
- Large file: two uncompressed PNGs produce an 11.8 MB `.docx` (11,774,538 bytes) plus a 40-by-4 table. Write took 161 ms, read 29 ms. Peak resident memory (`VmHWM`) was 87 MB, about seven times the file because the images, the zip, and the stored original parts are in memory together. That is not large enough to justify a second streaming writer. LibreOffice, POI, and python-docx all opened it and still saw `LARGE-DOC-MARKER` and the table cells.

The large PNG is generated by the measurement program and is not committed.

### Editing APIs

- `AcceptRevisions` / `RejectRevisions` apply to insertions, deletions, moves, and property changes (`rPrChange`, `pPrChange`, `tblPrChange`). Accept keeps `New` and drops `Gone`. Reject restores `Gone` and the previous property snapshot.
- `UpdateFields` sets `w:updateFields` and fills a TOC result from `Heading1`–`Heading9` and `Title` text. Page numbers stay for Word to calculate.
- Comments and footnotes were already on `Section`. A package written with both opens in the reference tools; POI shows the footnote body.
- `ImageStyle.WrappingStyle` of `square`, `tight`, `topAndBottom`, `behind`, or `inFrontOf` writes `wp:anchor` and the matching wrap element. `TableStyle.FloatXSpec` / `FloatYSpec` write `w:tblpPr`.
- Text inside a drawing is separated from the surrounding paragraph, so `AnchorId.docx` keeps `AnchorId` as its own word (recall 1.0).

### New fixtures

Additional permissive or explicitly requested files: more LibreOffice `ooxmlexport` documents (columns, fields, comments, linked text boxes, TOC fields), five mammoth.js documents (BSD-2-Clause), and docxjs render tests (Apache-2.0). Pandoc’s own docx tests are GPL and were not added.

Harness on 223 non-encrypted packages. Nothing that passed in round 12 regressed.

| Tool | Read ok | Round-trip ok |
| --- | --- | --- |
| LibreOffice | 215/223 | 219/223 |
| pandoc | 196/223 | 220/223 |
| python-docx | 203/223 | 220/223 |
| Apache POI XWPF | 202/223 | 221/223 |

### Still not goword defects

- LibreOffice refreshes date fields (`CA012-fldSimple.docx` becomes the day the file is opened).
- python-docx and POI delete soft hyphens (`optionalhyphen` versus `optional` + `hyphen`).
- pandoc’s equation text is TeX, and it invents image paths.
- python-docx repeats merged cells (`table-spans.docx`).
- Footnote ids in `docxjs/footnote.docx` are 2 and 3 because 0 and 1 are the separator notes. LibreOffice displays them as 1 and 2.
- POI adds `Comment by <author>` and ruby readings. The default extract does not.

## Round 14 — field results and page estimates

`UpdateFields` still sets `w:updateFields` so Word refreshes on open. It now also fills cached results:

- TOC entries are `heading<tab>page`, using `EstimateLayout`.
- `PAGE`, `NUMPAGES`, and `PAGEREF` use that same estimate. Header and footer `PAGE` fields are left unset, because one cached number cannot be right on every page.
- `REF` returns the bookmarked text. `SEQ` numbers captions (`SEQ Figure`, `\c` repeats the current value) in Arabic, roman, or alphabetic form.
- `DATE`, `TIME`, `CREATEDATE`, `PRINTDATE`, and `SAVEDATE` honor a subset of `\@` pictures (`yyyy`, `MMMM`, `dd`, `HH`, `am/pm`, …). `UpdateFieldsAt` pins the clock for tests.
- `MERGEFIELD` reads `SetMergeData`. `IF` evaluates a comparison, including `IF MERGEFIELD Name = "value" "yes" "no"`.

`EstimateLayout` uses the section page size and margins, a 276-twip line, and about 120 twips per character. Explicit page breaks are exact. Checked against LibreOffice PDF export (`pdftotext` form feeds):

| Document | Estimate | LibreOffice |
| --- | --- | --- |
| Two page breaks, heading on the last page | 3 | 3, and the heading is on page 3. The TOC result is `Chapter<tab>3`. |
| 80 paragraphs of two sentences each | 4 | 3 |

Flowing text was one page high. That is the accuracy of this estimate. It is not Word's line breaking, and `w:updateFields` stays on so Word replaces the cached numbers when it opens the file.

### New fixtures

Eight more files: docxjs text, underlines, page break, and text break (Apache-2.0); LibreOffice TOC field, multi-column index, anchor position, and unsupported text fields (MPL-2.0). Pandoc's docx tests remain GPL and were not added.

Harness on 231 non-encrypted packages. Nothing that passed in round 13 regressed.

| Tool | Read ok | Round-trip ok |
| --- | --- | --- |
| LibreOffice | 222/231 | 226/231 |
| pandoc | 202/231 | 227/231 |
| python-docx | 210/231 | 227/231 |
| Apache POI XWPF | 209/231 | 228/231 |

### Still open

- `TOC_field_b.docx` stores the page number in the next run with no space, so the extract is `text1` while the reference tools show `text` and `1` (recall 0.78). Inserting a separator into the stored run changed the round trip, so the cached result is left as it was saved.
- `anchor-position.docx`: pandoc inserts an image placeholder between `A` and `B`. goword keeps `AB`.
- Nested fields that were saved only as raw `w:fldChar` markup are not re-executed by `UpdateFields`; it evaluates `Field` values and TOC elements. Word still recalculates them on open.

