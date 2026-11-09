#!/usr/bin/env bash
# Download the Word .docx fixtures described in testdata/word/SOURCES.md.
# Upstream licenses allow redistribution; this script is the documented source list.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BASE="$ROOT/testdata/word"
mkdir -p "$BASE/python-docx" "$BASE/poi" "$BASE/mammoth" "$BASE/docx4j"

fetch() {
  local url="$1" dest="$2"
  echo "GET $dest"
  curl -fsSL "$url" -o "$dest"
}

PD=https://raw.githubusercontent.com/python-openxml/python-docx/master
POI=https://raw.githubusercontent.com/apache/poi/trunk/test-data/document
MM=https://raw.githubusercontent.com/mwilliamson/mammoth.js/master/test/test-data
DJ=https://raw.githubusercontent.com/plutext/docx4j/master

for f in \
  features/steps/test_files/doc-word-default-blank.docx \
  features/steps/test_files/doc-access-sections.docx \
  features/steps/test_files/doc-odd-even-hdrs.docx \
  features/steps/test_files/hdr-header-footer.docx \
  features/steps/test_files/num-having-numbering-part.docx \
  features/steps/test_files/par-hyperlinks.docx \
  features/steps/test_files/par-known-styles.docx \
  features/steps/test_files/run-enumerated-props.docx \
  features/steps/test_files/sct-first-page-hdrftr.docx \
  features/steps/test_files/sct-section-props.docx \
  features/steps/test_files/tbl-cell-props.docx \
  features/steps/test_files/tbl-having-tables.docx \
  features/steps/test_files/txt-font-props.docx \
  features/steps/test_files/comments-rich-para.docx \
  features/steps/test_files/tab-stops.docx \
  features/steps/test_files/par-rendered-page-breaks.docx \
  features/steps/test_files/shp-inline-shape-access.docx \
  features/steps/test_files/doc-coreprops.docx \
  tests/test_files/having-images.docx
do
  fetch "$PD/$f" "$BASE/python-docx/$(basename "$f")"
done

for f in \
  ComplexNumberedLists.docx FieldCodes.docx FldSimple.docx Headers.docx Numbering.docx \
  PageSpecificHeadFoot.docx bookmarks.docx Styles.docx TestTableCellAlign.docx \
  bug56075-changeTracking_on.docx footnotes.docx endnotes.docx comment.docx \
  testComment.docx shapes-with-text.docx HeaderFooterUnicode.docx WithTabs.docx \
  VariousPictures.docx rtl.docx form_footnotes.docx table_footnotes.docx \
  IllustrativeCases.docx SampleDoc.docx FancyFoot.docx capitalized.docx \
  DiffFirstPageHeadFoot.docx EmptyDocumentWithHeaderFooter.docx NumberingWOverrides.docx
do
  fetch "$POI/$f" "$BASE/poi/$f"
done

for f in \
  comments.docx endnotes.docx footnotes.docx footnote-hyperlink.docx simple-list.docx \
  single-paragraph.docx strict-format.docx strikethrough.docx tables.docx text-box.docx \
  tiny-picture.docx underline.docx
do
  fetch "$MM/$f" "$BASE/mammoth/$f"
done

fetch "$DJ/docx4j-core-tests/src/test/resources/2010/2010-mcAlternateContent.docx" "$BASE/docx4j/2010-mcAlternateContent.docx"
fetch "$DJ/docx4j-core-tests/src/test/resources/2010/2010-sample1.docx" "$BASE/docx4j/2010-sample1.docx"
fetch "$DJ/docx4j-core-tests/src/test/resources/2010/DrawingML_GraphicData_wps.docx" "$BASE/docx4j/DrawingML_GraphicData_wps.docx"
fetch "$DJ/docx4j-core-tests/src/test/resources/MERGEFIELD.docx" "$BASE/docx4j/MERGEFIELD.docx"
fetch "$DJ/docx4j-core-tests/src/test/resources/HelloWordOnline.docx" "$BASE/docx4j/HelloWordOnline.docx"
fetch "$DJ/docx4j-core-tests/src/test/resources/AlteredParts/header-simple-plus-image.docx" "$BASE/docx4j/header-simple-plus-image.docx"
fetch "$DJ/docx4j-core-tests/src/test/resources/AlteredParts/hyperlink.docx" "$BASE/docx4j/hyperlink.docx"
fetch "$DJ/docx4j-core-tests/src/test/resources/LegacyForms.docx" "$BASE/docx4j/LegacyForms.docx"
fetch "$DJ/docs/developer/Eastern%20Arabic%20numbering.docx" "$BASE/docx4j/eastern-arabic-numbering.docx"
fetch "$DJ/docx4j-core-tests/src/test/resources/vml/textbox.docx" "$BASE/docx4j/vml-textbox.docx"
fetch "$DJ/docx4j-samples-docx4j/sample-docs/hyperlinks-internal.docx" "$BASE/docx4j/hyperlinks-internal.docx"

echo "Downloaded $(find "$BASE" -name '*.docx' | wc -l) documents"
