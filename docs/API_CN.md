# GoWord API 参考文档

> 仓库地址：https://github.com/VantageDataChat/GoWord

## 顶层函数

```go
import goword "github.com/VantageDataChat/GoWord"
```

| 函数 | 说明 |
|------|------|
| `goword.New() *Document` | 创建新的空文档 |
| `goword.Open(path string) (*Document, error)` | 读取 .docx 文件 |
| `goword.OpenFromBytes(data []byte) (*Document, error)` | 从字节数据读取 .docx |

## Document 文档

```go
type Document struct {
    Properties  DocProperties  // 文档属性
    Sections    []*Section     // 节列表
    DefaultFont FontStyle      // 默认字体
}
```

### 方法

| 方法 | 说明 |
|------|------|
| `Save(path string) error` | 保存为 .docx 文件 |
| `ToBytes() ([]byte, error)` | 输出为字节切片 |
| `AddSection() *Section` | 添加默认 A4 纵向节 |
| `AddSectionWithStyle(s SectionStyle) *Section` | 添加自定义样式节 |
| `SetDefaultFontName(name string)` | 设置默认字体名称 |
| `SetDefaultFontSize(size float64)` | 设置默认字号（磅） |
| `AddFontStyle(name string, fs FontStyle)` | 注册命名字体样式 |
| `AddParagraphStyle(name string, ps ParagraphStyle)` | 注册命名段落样式 |
| `AddTableStyle(name string, ts TableStyle)` | 注册命名表格样式 |
| `AddNumberingStyle(name string, ns NumberingStyle)` | 注册命名编号样式 |
| `GetFontStyle(name string) (FontStyle, bool)` | 获取命名字体样式 |
| `GetParagraphStyle(name string) (ParagraphStyle, bool)` | 获取命名段落样式 |
| `GetTableStyle(name string) (TableStyle, bool)` | 获取命名表格样式 |
| `GetNumberingStyle(name string) (NumberingStyle, bool)` | 获取命名编号样式 |
| `FontStyles() map[string]FontStyle` | 所有字体样式 |
| `ParagraphStyles() map[string]ParagraphStyle` | 所有段落样式 |
| `TableStyles() map[string]TableStyle` | 所有表格样式 |
| `NumberingStyles() map[string]NumberingStyle` | 所有编号样式 |
| `Footnotes() []*Footnote` | 所有脚注 |
| `Endnotes() []*Endnote` | 所有尾注 |
| `Comments() []*Comment` | 所有批注 |
| `Images() []*Image` | 所有图片 |

## Section 节

### 方法

| 方法 | 说明 |
|------|------|
| `AddText(text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *Paragraph` | 添加文本段落 |
| `AddTextWithStyle(text, fontStyleName, paraStyleName string) *Paragraph` | 使用命名样式添加文本 |
| `AddTextRun(paraStyle *ParagraphStyle) *TextRun` | 添加混合格式段落 |
| `AddTextBreak(count int)` | 添加空行 |
| `AddPageBreak()` | 添加分页符 |
| `AddLink(url, text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *Hyperlink` | 添加超链接 |
| `AddTitle(text string, depth int) *Paragraph` | 添加标题（0=Title，1–9=Heading1–9） |
| `AddTable(tableStyle *TableStyle) *Table` | 添加表格 |
| `AddTableWithStyle(styleName string) *Table` | 使用命名样式添加表格 |
| `AddImage(src string, imgStyle *ImageStyle) *Image` | 从文件添加图片 |
| `AddImageFromBytes(data []byte, mimeType string, imgStyle *ImageStyle) *Image` | 从字节添加图片 |
| `AddListItem(text string, depth int, fontStyle *FontStyle, listStyleName string, paraStyle *ParagraphStyle) *ListItem` | 添加列表项 |
| `AddFootnote() *Footnote` | 添加脚注 |
| `AddEndnote() *Endnote` | 添加尾注 |
| `AddTOC(fontStyle *FontStyle, minDepth, maxDepth int) *TOC` | 添加目录 |
| `AddCheckBox(name, text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *CheckBox` | 添加复选框 |
| `AddLine(lineStyle Line) *Line` | 添加线条 |
| `AddBookmark(name string) *Bookmark` | 添加书签 |
| `AddComment(author, text string) *Comment` | 添加批注 |
| `AddHeader(headerType string) *Header` | 添加页眉（"default"、"first"、"even"） |
| `AddFooter(footerType string) *Footer` | 添加页脚（"default"、"first"、"even"） |


## TextRun 混合内联段落

在单个段落中组合文本、链接、图片、脚注和书签。

| 方法 | 说明 |
|------|------|
| `AddText(text string, fontStyle *FontStyle) *Run` | 添加文本 |
| `AddTextWithStyle(text, styleName string) *Run` | 使用命名样式添加文本 |
| `AddLink(url, text string, fontStyle *FontStyle) *Hyperlink` | 添加内联超链接 |
| `AddTextBreak(count int)` | 添加换行 |
| `AddImage(src string, imgStyle *ImageStyle) *Image` | 添加内联图片 |
| `AddFootnote() *Footnote` | 添加脚注引用 |
| `AddEndnote() *Endnote` | 添加尾注引用 |
| `AddBookmark(name string) *Bookmark` | 添加书签 |

## Table / Row / Cell 表格

### Table 表格

| 方法 | 说明 |
|------|------|
| `AddRow(height int, rowStyle *RowStyle) *Row` | 添加行 |

`Grid []int` 字段用于设置列宽（twip 单位）。

### Row 行

| 方法 | 说明 |
|------|------|
| `AddCell(width int, cellStyle *CellStyle) *Cell` | 添加单元格 |

### Cell 单元格

| 方法 | 说明 |
|------|------|
| `AddText(text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *Paragraph` | 添加文本 |
| `AddTextRun(paraStyle *ParagraphStyle) *TextRun` | 添加混合格式段落 |
| `AddImage(src string, imgStyle *ImageStyle) *Image` | 添加图片 |
| `AddTable(tableStyle *TableStyle) *Table` | 添加嵌套表格 |
| `AddListItem(text string, depth int, fontStyle *FontStyle, listStyleName string) *ListItem` | 添加列表项 |

## Header / Footer 页眉页脚

`Header` 和 `Footer` 均支持：

| 方法 | 说明 |
|------|------|
| `AddText(text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *Paragraph` | 添加文本 |
| `AddPreserveText(text string, fontStyle *FontStyle, paraStyle *ParagraphStyle) *PreserveText` | 添加域代码（如 `{PAGE}`、`{NUMPAGES}`） |
| `AddImage(src string, imgStyle *ImageStyle) *Image` | 添加图片 |
| `AddTable(tableStyle *TableStyle) *Table` | 添加表格 |

## Footnote / Endnote 脚注与尾注

`Footnote` 和 `Endnote` 均支持：

| 方法 | 说明 |
|------|------|
| `AddText(text string, fontStyle *FontStyle)` | 添加文本 |
| `AddLink(url, text string) *Hyperlink` | 添加超链接 |
| `AddTextBreak()` | 添加换行 |

## 样式类型

### FontStyle 字体样式

```go
type FontStyle struct {
    Name                string  // 字体名称，如 "Arial"
    Size                float64 // 字号（磅）
    Bold                bool    // 加粗
    Italic              bool    // 斜体
    Underline           string  // "single"、"double"、"wave"、"dash"、"dotted"、"none"
    Strikethrough       bool    // 删除线
    DoubleStrikethrough bool    // 双删除线
    SuperScript         bool    // 上标
    SubScript           bool    // 下标
    Color               string  // 十六进制颜色（不含 #），如 "FF0000"
    HighlightColor      string  // 高亮颜色，如 "yellow"
    AllCaps             bool    // 全部大写
    SmallCaps           bool    // 小型大写
    Hidden              bool    // 隐藏文字
    Spacing             float64 // 字符间距（磅）
    NoProof             bool    // 不检查拼写
    Lang                string  // 语言标签，如 "zh-CN"
}
```

### ParagraphStyle 段落样式

```go
type ParagraphStyle struct {
    Alignment       string // "left"、"center"、"right"、"both"、"distribute"
    SpaceBefore     int    // 段前间距（twip）
    SpaceAfter      int    // 段后间距（twip）
    LineSpacing     int    // 行距（twip，240 = 单倍行距）
    LineRule        string // "auto"、"exact"、"atLeast"
    Indent          int    // 左缩进（twip）
    IndentRight     int    // 右缩进（twip）
    FirstLine       int    // 首行缩进（twip）
    Hanging         int    // 悬挂缩进（twip）
    KeepNext        bool   // 与下段同页
    KeepLines       bool   // 段中不分页
    PageBreakBefore bool   // 段前分页
    WidowControl    bool   // 孤行控制
    Borders         *ParagraphBorders
    Shading         *Shading
}
```

### TableStyle 表格样式

```go
type TableStyle struct {
    Width           int    // 表格宽度（twip，0 = 自动）
    WidthType       string // "auto"、"dxa"、"pct"
    Alignment       string // "left"、"center"、"right"
    CellMarginTop   int    // 默认单元格上边距
    CellMarginBottom int
    CellMarginLeft  int
    CellMarginRight int
    BorderTop       *Border
    BorderBottom    *Border
    BorderLeft      *Border
    BorderRight     *Border
    BorderInsideH   *Border
    BorderInsideV   *Border
    Layout          string // "fixed"、"autofit"
}

// 便捷方法：一次设置所有边框
func (ts *TableStyle) SetAllBorders(style string, size int, color string)
```

### RowStyle 行样式

```go
type RowStyle struct {
    Height     int    // 行高（twip）
    HeightRule string // "auto"、"exact"、"atLeast"
    IsHeader   bool   // 作为表头行重复
    CantSplit  bool   // 不跨页拆分
}
```

### CellStyle 单元格样式

```go
type CellStyle struct {
    Width         int    // 宽度（twip）
    WidthType     string // "auto"、"dxa"、"pct"
    GridSpan      int    // 跨列数
    VMerge        string // "restart" 开始合并、"continue" 继续合并、"" 不合并
    VAlign        string // "top"、"center"、"bottom"
    Shading       *Shading
    BorderTop     *Border
    BorderBottom  *Border
    BorderLeft    *Border
    BorderRight   *Border
    TextDirection string // "lrTb"、"tbRl"、"btLr"
    NoWrap        bool   // 不自动换行
}
```

### SectionStyle 节样式

```go
type SectionStyle struct {
    Orientation     string // "portrait"（纵向）、"landscape"（横向）
    PageWidth       int    // 页宽（twip）
    PageHeight      int    // 页高（twip）
    MarginTop       int    // 上边距
    MarginBottom    int    // 下边距
    MarginLeft      int    // 左边距
    MarginRight     int    // 右边距
    HeaderHeight    int    // 页眉距顶部距离
    FooterHeight    int    // 页脚距底部距离
    ColumnCount     int    // 分栏数
    ColumnSpacing   int    // 栏间距
    PageNumberStart *int   // 起始页码（nil = 续前节）
    BreakType       string // "nextPage"、"continuous"、"evenPage"、"oddPage"
}

func DefaultSectionStyle() SectionStyle  // A4 纵向，1 英寸边距
func (s *SectionStyle) SetPaperSize(name string) // "A4"、"A3"、"A5"、"Letter"、"Legal"
```

### ImageStyle 图片样式

```go
type ImageStyle struct {
    Width         int    // 宽度（磅）
    Height        int    // 高度（磅）
    Alignment     string // "left"、"center"、"right"
    WrappingStyle string // "inline"、"behind"、"inFrontOf"、"square"、"tight"
    MarginTop     int
    MarginBottom  int
    MarginLeft    int
    MarginRight   int
}
```

### NumberingStyle 编号样式

```go
type NumberingStyle struct {
    Type   string           // "multilevel"（多级）、"singleLevel"（单级）
    Levels []NumberingLevel
}

type NumberingLevel struct {
    Format  string // "decimal"、"upperLetter"、"lowerLetter"、"upperRoman"、"lowerRoman"、"bullet"
    Text    string // 如 "%1." 或 "\u2022"
    Left    int    // 左缩进（twip）
    Hanging int    // 悬挂缩进（twip）
    TabPos  int    // 制表位（twip）
    Font    string // 项目符号字体名称
}
```

### DocProperties 文档属性

```go
type DocProperties struct {
    Title          string    // 标题
    Subject        string    // 主题
    Creator        string    // 作者
    Keywords       string    // 关键词
    Description    string    // 描述
    Category       string    // 类别
    LastModifiedBy string    // 最后修改者
    Company        string    // 公司
    Manager        string    // 经理
    Created        time.Time // 创建时间
    Modified       time.Time // 修改时间
    Revision       int       // 修订号
}
```

## 单位转换工具

```go
import "github.com/VantageDataChat/GoWord/common"
```

| 函数 | 说明 |
|------|------|
| `common.InchToTwip(inches float64) int` | 英寸 → twip |
| `common.CmToTwip(cm float64) int` | 厘米 → twip |
| `common.PointToTwip(pt float64) int` | 磅 → twip |
| `common.TwipToInch(twip int) float64` | twip → 英寸 |
| `common.TwipToCm(twip int) float64` | twip → 厘米 |
| `common.TwipToPoint(twip int) float64` | twip → 磅 |
| `common.EmuToTwip(emu int64) int` | EMU → twip |
| `common.TwipToEmu(twip int) int64` | twip → EMU |
| `common.PointToHalfPoint(pt float64) int` | 磅 → 半磅（字号单位） |
| `common.HalfPointToPoint(hp int) float64` | 半磅 → 磅 |
| `common.EmuToPixel(emu int64) int` | EMU → 像素（96 DPI） |
| `common.PixelToEmu(px int) int64` | 像素 → EMU（96 DPI） |

## 元素类型

所有节元素均实现 `Element` 接口。使用类型断言识别：

```go
for _, elem := range section.Elements {
    switch e := elem.(type) {
    case *goword.Paragraph:
        for _, run := range e.Runs {
            fmt.Println(run.Text)
        }
    case *goword.Table:
        for _, row := range e.Rows {
            for _, cell := range row.Cells {
                // ...
            }
        }
    case *goword.Image:
        fmt.Println(e.Source, e.Style.Width, e.Style.Height)
    case *goword.Hyperlink:
        fmt.Println(e.URL, e.Text)
    // ... TextRun, TextBreak, PageBreak, ListItem, Footnote, Endnote,
    //     Bookmark, BookmarkEnd, Comment, PreserveText, TOC, CheckBox, Line
    }
}
```
