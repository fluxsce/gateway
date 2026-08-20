package controllers

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"gateway/web/views/hub0004/models"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// auditLogCSVHeaders 导出列名，与列表字段对齐；动作/模块/类型保留编码便于核对。
var auditLogCSVHeaders = []string{
	"操作时间", "操作人", "用户ID", "动作", "模块",
	"目标类型", "目标名称", "目标ID", "资源编码", "结果",
	"客户端IP", "请求方法", "请求路径", "补充说明",
}

// writeAuditLogCSV 把审计行写成带 UTF-8 BOM 的 CSV，便于 Excel 直接打开。
func writeAuditLogCSV(w io.Writer, rows []*models.AuthAuditLog) error {
	cw, err := startAuditLogCSV(w)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		if err := cw.Write(auditLogCSVRecord(row)); err != nil {
			return err
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("写入CSV失败: %w", err)
	}
	return nil
}

// startAuditLogCSV 写出 BOM 与表头，后续由调用方逐行 Write。
func startAuditLogCSV(w io.Writer) (*csv.Writer, error) {
	if _, err := w.Write(utf8BOM); err != nil {
		return nil, err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write(auditLogCSVHeaders); err != nil {
		return nil, err
	}
	return cw, nil
}

func auditLogCSVRecord(row *models.AuthAuditLog) []string {
	addTime := ""
	if !row.AddTime.IsZero() {
		addTime = row.AddTime.Format("2006-01-02 15:04:05")
	}
	return []string{
		addTime,
		row.UserName,
		row.UserId,
		row.Action,
		row.ModuleCode,
		row.TargetType,
		row.TargetName,
		row.TargetId,
		row.ResourceCode,
		row.Result,
		row.ClientIP,
		row.RequestMethod,
		row.RequestPath,
		row.Detail,
	}
}

// auditExportFilename 生成导出文件名。
func auditExportFilename(now time.Time) string {
	return "audit-log-" + now.Format("20060102150405") + ".csv"
}
