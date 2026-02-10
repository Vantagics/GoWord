# GoWord

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Zero Dependencies](https://img.shields.io/badge/依赖-零外部依赖-brightgreen.svg)]()

纯 Go 语言实现的 Microsoft Word 文档（.docx / Office Open XML）读写库。灵感来自 [PHPWord](https://github.com/PHPOffice/PHPWord)，零外部依赖。

## 功能特性

- 创建和读取 `.docx` 文件
- 文档属性（标题、作者、主题、关键词等）
- 节（Section）：页面大小、方向、页边距、分栏、页码起始
- 段落：丰富的字体样式（加粗、斜体、下划线、颜色、字号等）
- 段落格式：对齐、间距、缩进、边框、底纹
- 标题（Title、Heading1–9）
- 超链接
- 表格：单元格合并（gridSpan、vMerge）、边框、底纹、嵌套表格
- 列表：自定义编号样式（项目符号、数字、罗马数字等）
- 图片：支持文件路径或字节数据（PNG、JPEG、GIF、BMP、TIFF）
- 页眉页脚：支持页码等域代码
- 脚注和尾注
- 批注、书签、目录
- 复选框、线条图形
- 水印（文字水印和图片水印）
- 表单字段（文本输入、下拉列表、复选框）
- 制表位和制表符
- 命名样式（字体、段落、表格、编号）
- TextRun 混合格式内联内容
- ExtractText 全文提取
- 文档级操作（Paragraphs、Tables、InsertParagraph、RemoveParagraph）
- 域代码常量（PAGE、NUMPAGES、DATE、AUTHOR 等）
- UpdateFieldsOnOpen 设置
- 单位转换工具（twip、厘米、英寸、磅、EMU、像素）

## 安装

```bash
go get github.com/VantageDataChat/GoWord
```

## 快速开始

```go
package main

import (
    "log"

    goword "github.com/VantageDataChat/GoWord"
    "github.com/VantageDataChat/GoWord/style"
)

func main() {
    doc := goword.New()
    doc.Properties.Title = "我的文档"
    doc.Properties.Creator = "GoWord"

    sec := doc.AddSection()
    sec.AddTitle("你好 GoWord", 1)
    sec.AddText("这是一段加粗文本。",
        &style.FontStyle{Bold: true, Size: 12, Color: "333333"}, nil)

    if err := doc.Save("hello.docx"); err != nil {
        log.Fatal(err)
    }
}
```

## 读取文档

```go
doc, err := goword.Open("existing.docx")
if err != nil {
    log.Fatal(err)
}
for _, sec := range doc.Sections {
    for _, elem := range sec.Elements {
        // 处理元素...
    }
}
```


## 使用示例

### 表格

```go
sec := doc.AddSection()
ts := &style.TableStyle{Width: 9000, Alignment: "center"}
ts.SetAllBorders("single", 4, "000000")
tbl := sec.AddTable(ts)
tbl.Grid = []int{3000, 3000, 3000}

row := tbl.AddRow(0, &style.RowStyle{IsHeader: true})
row.AddCell(3000, nil).AddText("姓名", &style.FontStyle{Bold: true}, nil)
row.AddCell(3000, nil).AddText("年龄", &style.FontStyle{Bold: true}, nil)
row.AddCell(3000, nil).AddText("城市", &style.FontStyle{Bold: true}, nil)

row2 := tbl.AddRow(0, nil)
row2.AddCell(3000, nil).AddText("张三", nil, nil)
row2.AddCell(3000, nil).AddText("30", nil, nil)
row2.AddCell(3000, nil).AddText("北京", nil, nil)
```

### 图片

```go
// 从文件路径
sec.AddImage("photo.png", &style.ImageStyle{Width: 200, Height: 150})

// 从字节数据
sec.AddImageFromBytes(pngData, "image/png", &style.ImageStyle{Width: 100, Height: 100})
```

### 页眉、页脚与页码

```go
header := sec.AddHeader("")
header.AddText("公司名称", &style.FontStyle{Bold: true}, nil)

footer := sec.AddFooter("")
footer.AddPreserveText("第 {PAGE} 页 / 共 {NUMPAGES} 页", nil,
    &style.ParagraphStyle{Alignment: style.AlignCenter})
```

### 列表

```go
doc.AddNumberingStyle("bullets", goword.NumberingStyle{
    Type: "singleLevel",
    Levels: []goword.NumberingLevel{
        {Format: "bullet", Text: "\u2022", Left: 360, Hanging: 360, Font: "Symbol"},
    },
})
sec.AddListItem("第一项", 0, nil, "bullets", nil)
sec.AddListItem("第二项", 0, nil, "bullets", nil)
```

### TextRun（混合内联内容）

```go
tr := sec.AddTextRun(nil)
tr.AddText("普通文本 ", nil)
tr.AddText("加粗文本 ", &style.FontStyle{Bold: true})
tr.AddLink("https://example.com", "一个链接", nil)
fn := tr.AddFootnote()
fn.AddText("脚注内容", nil)
```

### 脚注与尾注

```go
fn := sec.AddFootnote()
fn.AddText("参见参考文献。", nil)
fn.AddLink("https://example.com", "来源")

en := sec.AddEndnote()
en.AddText("尾注文本。", nil)
en.AddLink("https://example.com", "参考")
en.AddTextBreak()
en.AddText("第二行。", nil)
```

### 节与页面布局

```go
// 横向 A4
ss := style.DefaultSectionStyle()
ss.Orientation = style.OrientLandscape
ss.PageWidth = 16838
ss.PageHeight = 11906
sec := doc.AddSectionWithStyle(ss)

// 双栏
ss2 := style.DefaultSectionStyle()
ss2.ColumnCount = 2
ss2.ColumnSpacing = 720
```

### 命名样式

```go
doc.AddFontStyle("emphasis", style.FontStyle{Italic: true, Color: "0000FF"})
doc.AddParagraphStyle("centered", style.ParagraphStyle{Alignment: style.AlignCenter})

sec.AddTextWithStyle("带样式的文本", "emphasis", "centered")
```

### 水印

```go
// 文字水印
wm := doc.AddWatermarkText("草稿")
wm.Font = "SimSun"
wm.Color = "C0C0C0"
wm.Bold = true

// 图片水印
wm2 := doc.AddWatermarkPictureFromBytes(pngData, "image/png")
wm2.Washout = true
```

### 表单字段

```go
// 文本输入
ff := sec.AddTextInput("username")
ff.DefaultValue = "请输入姓名"
ff.MaxLength = 50

// 下拉列表
dd := sec.AddDropdownList("color", []string{"红色", "绿色", "蓝色"})
dd.DefaultValue = "绿色"

// 复选框表单字段
cb := sec.AddFormField(goword.FormFieldTypeCheckBox, "agree")
cb.Value = "true"
```

### 制表位

```go
ps := &style.ParagraphStyle{
    TabStops: []style.TabStop{
        {Position: 2880, Type: "center", Leader: "dot"},
        {Position: 5760, Type: "right"},
    },
}
tr := sec.AddTextRun(ps)
tr.AddText("左对齐", nil)
tr.AddTab()
tr.AddText("居中", nil)
tr.AddTab()
tr.AddText("右对齐", nil)
```

### 提取文本

```go
doc, _ := goword.Open("document.docx")
text := doc.ExtractText()
fmt.Println(text)
```

### 文档级操作

```go
// 获取所有段落/表格（跨节）
paragraphs := doc.Paragraphs()
tables := doc.Tables()

// 插入/删除段落
newP := doc.InsertParagraphAfter(paragraphs[0])
doc.RemoveParagraph(paragraphs[1])

// 打开时自动更新域
doc.SetUpdateFieldsOnOpen(true)
```

## 项目结构

```
github.com/VantageDataChat/GoWord
├── goword.go              # 顶层 API：New()、Open()、类型别名
├── common/
│   ├── properties.go      # 文档属性 DocProperties
│   └── units.go           # 单位转换（twip、厘米、英寸、磅、EMU、像素）
├── style/
│   ├── font.go            # 字体样式 FontStyle
│   ├── paragraph.go       # 段落样式 ParagraphStyle、边框、底纹
│   ├── table.go           # 表格样式 TableStyle、行样式、单元格样式
│   ├── section.go         # 节样式 SectionStyle、纸张大小
│   └── image.go           # 图片样式 ImageStyle
├── document/
│   ├── document.go        # Document 结构体、NumberingStyle
│   ├── document_api.go    # Document 方法（AddSection、AddStyle 等）
│   ├── elements.go        # 所有元素类型（Paragraph、Table、Image 等）
│   ├── section.go         # Section.Add* 方法
│   ├── table.go           # Table/Row/Cell 方法
│   ├── textrun.go         # TextRun.Add* 方法
│   ├── headerfooter.go    # 页眉/页脚/脚注/尾注方法
│   └── io.go              # Save/Open 可插拔后端
└── ooxml/
    ├── reader.go          # Read/ReadFromBytes
    ├── writer.go          # Save/WriteToBytes/WriteToWriter
    └── ...                # OOXML XML 生成与解析
```

## 测试

```bash
go test ./... -race
```

共 374 个测试，全部通过（含竞态检测）。覆盖率：

| 包         | 覆盖率   |
|------------|----------|
| `common`   | 100%     |
| `style`    | 100%     |
| `document` | 100%     |
| `ooxml`    | 96.1%    |

## 许可证

MIT
