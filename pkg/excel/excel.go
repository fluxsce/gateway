// Package excel 提供 Excel 文件生成（导出）和解析（导入）工具，基于 excelize/v2。
//
// # 导出
//
// Build 将若干 Sheet 写入指定路径的 xlsx 文件，返回 BuildResult（不持有文件内容字节）：
//
//	result, err := excel.Build("/tmp/export.xlsx",
//	    excel.Sheet{Name: "RouteConfig", Headers: []string{"id", "name"}, Rows: rows},
//	)
//	// result.Path  — 文件绝对路径
//	// result.Size  — 文件字节数
//
// 指定 Index 可将 Sheet 写入文件的固定位置（1-based，0 表示追加到末尾）：
//
//	excel.Sheet{Index: 2, Name: "Detail", Headers: headers, Rows: rows}
//
// # 导入
//
//	sheets, err := excel.Parse(reader)
//	rows := sheets["RouteConfig"] // [][]string，第 0 行为标题行
package excel

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Sheet 描述单个 Sheet 的结构与数据。
type Sheet struct {
	// Name Sheet 名称；超过 31 字符会自动截断
	Name string
	// Index 目标 Sheet 位置（1-based）。
	// 0 或未设置时按传入顺序追加；已存在同名 Sheet 时原地覆写。
	Index int
	// Headers 列头，顺序与 Rows 每行元素对应
	Headers []string
	// ColWidths 可选列宽（按 Headers 顺序），0 或缺省时使用默认宽度 16
	ColWidths []float64
	// Rows 数据行，每行长度应与 Headers 一致
	Rows [][]any
}

// BuildResult 描述已生成的 Excel 文件，不持有文件内容字节。
type BuildResult struct {
	// Path 生成文件的绝对路径（与传入 path 一致）
	Path string
	// Size 文件字节数
	Size int64
	// SheetCount 实际写入的 Sheet 数量
	SheetCount int
}

// Build 将 sheets 写入 path 指定的 xlsx 文件。
// 若文件已存在则追加/覆写对应 Sheet；不存在则新建。
// 成功后返回 BuildResult，调用方可用 Path 读取文件内容并发送给客户端。
func Build(path string, sheets ...Sheet) (*BuildResult, error) {
	if len(sheets) == 0 {
		return nil, fmt.Errorf("至少需要一个 Sheet")
	}

	var f *excelize.File
	if _, statErr := os.Stat(path); statErr == nil {
		// 文件已存在，追加模式打开
		var openErr error
		f, openErr = excelize.OpenFile(path)
		if openErr != nil {
			return nil, fmt.Errorf("打开已有文件 [%s] 失败: %w", path, openErr)
		}
	} else {
		f = excelize.NewFile()
	}
	defer f.Close()

	hStyle := headerStyle(f)
	// strStyle 强制文本格式，防止 "1.2"、"01" 等字符串被 Excel 自动转为数字
	strStyle := textStyle(f)
	defaultSheetRenamed := false

	for _, s := range sheets {
		name := truncateSheetName(s.Name)

		// 确定 Sheet 是否已存在
		existingIdx, _ := f.GetSheetIndex(name)
		if existingIdx == -1 {
			// Sheet 不存在，需要新建
			if !defaultSheetRenamed {
				// 将 excelize 默认的 Sheet1 重命名为第一个目标 Sheet
				f.SetSheetName("Sheet1", name)
				defaultSheetRenamed = true
			} else {
				if _, err := f.NewSheet(name); err != nil {
					return nil, fmt.Errorf("创建 Sheet [%s] 失败: %w", name, err)
				}
			}
		}

		// 按 Index 调整 Sheet 顺序（1-based，0 表示不调整）
		if s.Index > 0 {
			f.SetSheetVisible(name, true)
			sheetList := f.GetSheetList()
			currentPos := -1
			for i, n := range sheetList {
				if n == name {
					currentPos = i + 1 // 转为 1-based
					break
				}
			}
			if currentPos != s.Index {
				f.MoveSheet(name, sheetList[s.Index-1])
			}
		}

		sw, err := f.NewStreamWriter(name)
		if err != nil {
			return nil, fmt.Errorf("创建 StreamWriter [%s] 失败: %w", name, err)
		}

		for j := range s.Headers {
			colW := 16.0
			if j < len(s.ColWidths) && s.ColWidths[j] > 0 {
				colW = s.ColWidths[j]
			}
			_ = sw.SetColWidth(j+1, j+1, colW)
		}

		headerRow := make([]any, len(s.Headers))
		for j, h := range s.Headers {
			headerRow[j] = excelize.Cell{StyleID: hStyle, Value: h}
		}
		if err := sw.SetRow("A1", headerRow); err != nil {
			return nil, fmt.Errorf("写入标题行 [%s] 失败: %w", name, err)
		}

		for rowIdx, row := range s.Rows {
			cells := make([]any, len(row))
			for j, c := range row {
				cells[j] = formatCellWithStyle(c, strStyle)
			}
			addr, _ := excelize.JoinCellName("A", rowIdx+2)
			if err := sw.SetRow(addr, cells); err != nil {
				return nil, fmt.Errorf("写入数据行 [%s] 第 %d 行失败: %w", name, rowIdx+2, err)
			}
		}

		if err := sw.Flush(); err != nil {
			return nil, fmt.Errorf("Flush [%s] 失败: %w", name, err)
		}
	}

	if err := f.SaveAs(path); err != nil {
		return nil, fmt.Errorf("保存文件 [%s] 失败: %w", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件信息 [%s] 失败: %w", path, err)
	}

	return &BuildResult{
		Path:       path,
		Size:       info.Size(),
		SheetCount: len(sheets),
	}, nil
}

// ─── 导入 ─────────────────────────────────────────────────────────────────

// ParseResult 解析结果，key 为 Sheet 名，value 为二维字符串（含标题行）。
type ParseResult map[string][][]string

// Parse 从 io.Reader 读取 xlsx 文件，返回各 Sheet 的原始字符串数据。
// 每个 [][]string 第 0 行为标题行，后续行为数据行。
func Parse(r io.Reader) (ParseResult, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 文件失败: %w", err)
	}
	defer f.Close()

	result := make(ParseResult)
	for _, sheetName := range f.GetSheetList() {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			return nil, fmt.Errorf("读取 Sheet [%s] 失败: %w", sheetName, err)
		}
		result[sheetName] = rows
	}
	return result, nil
}

// HeaderIndex 根据标题行构建列名 → 列索引的映射，便于按名称取值。
func HeaderIndex(headerRow []string) map[string]int {
	idx := make(map[string]int, len(headerRow))
	for i, h := range headerRow {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		idx[h] = i
	}
	return idx
}

// GetCell 安全取值：根据 HeaderIndex 和列名取单元格值，超出范围返回空字符串。
func GetCell(row []string, idx map[string]int, col string) string {
	col = strings.TrimSpace(col)
	i, ok := idx[col]
	if !ok || i >= len(row) {
		return ""
	}
	return row[i]
}

// ─── 内部工具 ──────────────────────────────────────────────────────────────

func truncateSheetName(name string) string {
	runes := []rune(name)
	if len(runes) > 31 {
		return string(runes[:31])
	}
	return name
}

func headerStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	return style
}

// textStyle 返回强制文本格式的样式 ID（NumFmt 代码 "@"），
// 防止 Excel 将 "1.2"、"01"、"true" 等字符串自动推断为数值/布尔类型。
func textStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		CustomNumFmt: func() *string { s := "@"; return &s }(),
	})
	return style
}

// formatCell 将值规范化为 excelize 可接受的类型：
//   - string / *string  → excelize.Cell{StyleID: strStyle, Value: s}，强制文本格式，
//     防止 "1.2"、"01"、"true" 等内容被 Excel 自动转为数值/布尔类型。
//   - time.Time / *time.Time → 格式化为 "2006-01-02 15:04:05" 字符串（同样走文本单元格）。
//   - *int / *int64 / *float64 / *bool → 解引用；nil → ""（空字符串文本单元格）。
//   - 其他数值类型直接返回，由 excelize 按数值写入。
func formatCell(v any) any {
	return formatCellWithStyle(v, 0)
}

// formatCellWithStyle 是 formatCell 的带样式版本，供 Build 内部传入文本样式 ID。
func formatCellWithStyle(v any, strStyleID int) any {
	// wrapStr 将字符串包装为强制文本单元格，避免 Excel 自动推断类型。
	wrapStr := func(s string) any {
		if strStyleID == 0 {
			return s
		}
		return excelize.Cell{StyleID: strStyleID, Value: s}
	}

	if v == nil {
		return wrapStr("")
	}
	switch val := v.(type) {
	case string:
		return wrapStr(val)
	case *string:
		if val == nil {
			return wrapStr("")
		}
		return wrapStr(*val)
	case time.Time:
		if val.IsZero() {
			return wrapStr("")
		}
		return wrapStr(val.Format("2006-01-02 15:04:05"))
	case *time.Time:
		if val == nil || val.IsZero() {
			return wrapStr("")
		}
		return wrapStr(val.Format("2006-01-02 15:04:05"))
	case *int:
		if val == nil {
			return wrapStr("")
		}
		return *val
	case *int64:
		if val == nil {
			return wrapStr("")
		}
		return *val
	case *float64:
		if val == nil {
			return wrapStr("")
		}
		return *val
	case *bool:
		if val == nil {
			return wrapStr("")
		}
		return *val
	default:
		return v
	}
}
