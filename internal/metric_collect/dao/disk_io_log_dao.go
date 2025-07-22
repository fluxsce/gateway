package dao

import (
	"context"
	"fmt"
	"gateway/internal/metric_collect/types"
	"gateway/pkg/database"
	"time"
)

// DiskIoLogDAO 磁盘IO日志数据访问对象
type DiskIoLogDAO struct {
	db database.Database
}

// NewDiskIoLogDAO 创建磁盘IO日志DAO实例
func NewDiskIoLogDAO(db database.Database) *DiskIoLogDAO {
	return &DiskIoLogDAO{db: db}
}

// InsertDiskIoLog 插入磁盘IO日志
func (dao *DiskIoLogDAO) InsertDiskIoLog(ctx context.Context, diskIoLog *types.DiskIoLog) error {
	// 设置通用字段默认值
	now := time.Now()
	diskIoLog.AddTime = now
	diskIoLog.EditTime = now
	diskIoLog.CurrentVersion = 1
	diskIoLog.ActiveFlag = types.ActiveFlagYes

	// 验证必填字段
	if diskIoLog.MetricDiskIoLogId == "" {
		return fmt.Errorf("磁盘IO日志ID不能为空")
	}
	if diskIoLog.TenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if diskIoLog.MetricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	if diskIoLog.AddWho == "" {
		return fmt.Errorf("创建人ID不能为空")
	}
	if diskIoLog.EditWho == "" {
		return fmt.Errorf("修改人ID不能为空")
	}
	if diskIoLog.OprSeqFlag == "" {
		return fmt.Errorf("操作序列标识不能为空")
	}

	tableName := diskIoLog.TableName()
	_, err := dao.db.Insert(ctx, tableName, diskIoLog, true)
	if err != nil {
		return fmt.Errorf("插入磁盘IO日志失败: %w", err)
	}
	return nil
}

// BatchInsertDiskIoLog 批量插入磁盘IO日志
func (dao *DiskIoLogDAO) BatchInsertDiskIoLog(ctx context.Context, diskIoLogs []*types.DiskIoLog) error {
	if len(diskIoLogs) == 0 {
		return nil
	}

	// 设置通用字段默认值
	now := time.Now()
	for _, diskIoLog := range diskIoLogs {
		diskIoLog.AddTime = now
		diskIoLog.EditTime = now
		diskIoLog.CurrentVersion = 1
		diskIoLog.ActiveFlag = types.ActiveFlagYes

		// 验证必填字段
		if diskIoLog.MetricDiskIoLogId == "" {
			return fmt.Errorf("磁盘IO日志ID不能为空")
		}
		if diskIoLog.TenantId == "" {
			return fmt.Errorf("租户ID不能为空")
		}
		if diskIoLog.MetricServerId == "" {
			return fmt.Errorf("服务器ID不能为空")
		}
		if diskIoLog.AddWho == "" {
			return fmt.Errorf("创建人ID不能为空")
		}
		if diskIoLog.EditWho == "" {
			return fmt.Errorf("修改人ID不能为空")
		}
		if diskIoLog.OprSeqFlag == "" {
			return fmt.Errorf("操作序列标识不能为空")
		}
	}

	tableName := diskIoLogs[0].TableName()

	// 转换为interface{}切片
	items := make([]interface{}, len(diskIoLogs))
	for i, diskIoLog := range diskIoLogs {
		items[i] = diskIoLog
	}

	_, err := dao.db.BatchInsert(ctx, tableName, items, true)
	if err != nil {
		return fmt.Errorf("批量插入磁盘IO日志失败: %w", err)
	}
	return nil
}

// DeleteDiskIoLog 删除磁盘IO日志（软删除）
func (dao *DiskIoLogDAO) DeleteDiskIoLog(ctx context.Context, tenantId, metricDiskIoLogId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricDiskIoLogId == "" {
		return fmt.Errorf("磁盘IO日志ID不能为空")
	}

	tableName := (&types.DiskIoLog{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = ? WHERE tenantId = ? AND metricDiskIoLogId = ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{types.ActiveFlagNo, tenantId, metricDiskIoLogId}, true)
	if err != nil {
		return fmt.Errorf("删除磁盘IO日志失败: %w", err)
	}
	return nil
}

// DeleteDiskIoLogByTime 根据时间范围删除磁盘IO日志（物理删除）
func (dao *DiskIoLogDAO) DeleteDiskIoLogByTime(ctx context.Context, tenantId string, beforeTime time.Time) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}

	tableName := (&types.DiskIoLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND collectTime < ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, beforeTime}, true)
	if err != nil {
		return fmt.Errorf("根据时间删除磁盘IO日志失败: %w", err)
	}
	return nil
}

// DeleteDiskIoLogByServer 根据服务器ID删除磁盘IO日志（物理删除）
func (dao *DiskIoLogDAO) DeleteDiskIoLogByServer(ctx context.Context, tenantId, metricServerId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}

	tableName := (&types.DiskIoLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND metricServerId = ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, metricServerId}, true)
	if err != nil {
		return fmt.Errorf("根据服务器ID删除磁盘IO日志失败: %w", err)
	}
	return nil
}
