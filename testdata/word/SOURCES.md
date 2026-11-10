# Word fixture documents

These `.docx` files are real documents used by the compatibility tests in
`ooxml/word_compat_test.go`. They are redistributed under their upstream
licenses, which are separate from GoWord's license.

| Directory | Upstream | License |
| --- | --- | --- |
| `python-docx/` | [python-openxml/python-docx](https://github.com/python-openxml/python-docx) `features/steps/test_files` and `tests/test_files` | MIT |
| `poi/` | [apache/poi](https://github.com/apache/poi) `test-data/document` | Apache-2.0 |
| `mammoth/` | [mwilliamson/mammoth.js](https://github.com/mwilliamson/mammoth.js) `test/test-data` | BSD-2-Clause |
| `docx4j/` | [plutext/docx4j](https://github.com/plutext/docx4j) test resources and sample docs | Apache-2.0 |
| `phpword/` | [PHPOffice/PHPWord](https://github.com/PHPOffice/PHPWord) `tests/PhpWordTests/_files` and `samples/resources` | LGPL-3.0 |
| `openxmlsdk/` | [dotnet/Open-XML-SDK](https://github.com/dotnet/Open-XML-SDK) `test/DocumentFormat.OpenXml.Tests.Assets` | MIT |
| `libreoffice/` | [LibreOffice](https://github.com/LibreOffice/core) `sw/qa/extras/ooxmlexport/data` and `ooxmlimport/data` | MPL-2.0 |
| `encrypted/` | Apache POI `test-data/document` and LibreOffice `ooxmlexport/data` encrypted packages | Apache-2.0 and MPL-2.0 |

`scripts/fetch_word_samples.sh` downloads the same set. The files were produced
by Microsoft Word or by libraries whose test corpus is Word-generated OOXML,
and they cover multiple sections, headers and footers, styles, numbering,
tables (including nested tables and merged cells), pictures, text boxes,
fields, footnotes, endnotes, comments, hyperlinks, strict-namespace units,
and an alternate main-document part name.
