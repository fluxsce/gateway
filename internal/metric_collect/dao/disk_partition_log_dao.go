package dao

import (
	"context"
	"fmt"
	"gohub/internal/metric_collect/types"
	"gohub/pkg/database"
	"time"
)

// DiskPartitionLogDAO 磁盘分区日志数据访问对象
type DiskPartitionLogDAO struct {
	db database.Database
}

// NewDiskPartitionLogDAO 创建磁盘分区日志DAO实例
func NewDiskPartitionLogDAO(db database.Database) *DiskPartitionLogDAO {
	return &DiskPartitionLogDAO{db: db}
}

// InsertDiskPartitionLog 插入磁盘分区日志
func (dao *DiskPartitionLogDAO) InsertDiskPartitionLog(ctx context.Context, diskPartitionLog *types.DiskPartitionLog) error {
	// 设置通用字段默认值
	now := time.Now()
	diskPartitionLog.AddTime = now
	diskPartitionLog.EditTime = now
	diskPartitionLog.CurrentVersion = 1
	diskPartitionLog.ActiveFlag = types.ActiveFlagYes
	
	// 验证必填字段
	if diskPartitionLog.MetricDiskPartitionLogId == "" {
		return fmt.Errorf("磁盘分区日志ID不能为空")
	}
	if diskPartitionLog.TenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if diskPartitionLog.MetricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	if diskPartitionLog.AddWho == "" {
		return fmt.Errorf("创建人ID不能为空")
	}
	if diskPartitionLog.EditWho == "" {
		return fmt.Errorf("修改人ID不能为空")
	}
	if diskPartitionLog.OprSeqFlag == "" {
		return fmt.Errorf("操作序列标识不能为空")
	}
	
	tableName := diskPartitionLog.TableName()
	_, err := dao.db.Insert(ctx, tableName, diskPartitionLog, true)
	if err != nil {
		return fmt.Errorf("插入磁盘分区日志失败: %w", err)
	}
	return nil
}

// BatchInsertDiskPartitionLog 批量插入磁盘分区日志
func (dao *DiskPartitionLogDAO) BatchInsertDiskPartitionLog(ctx context.Context, diskPartitionLogs []*types.DiskPartitionLog) error {
	if len(diskPartitionLogs) == 0 {
		return nil
	}
	
	// 设置通用字段默认值
	now := time.Now()
	for _, diskPartitionLog := range diskPartitionLogs {
		diskPartitionLog.AddTime = now
		diskPartitionLog.EditTime = now
		diskPartitionLog.CurrentVersion = 1
		diskPartitionLog.ActiveFlag = types.ActiveFlagYes
		
		// 验证必填字段
		if diskPartitionLog.MetricDiskPartitionLogId == "" {
			return fmt.Errorf("磁盘分区日志ID不能为空")
		}
		if diskPartitionLog.TenantId == "" {
			return fmt.Errorf("租户ID不能为空")
		}
		if diskPartitionLog.MetricServerId == "" {
			return fmt.Errorf("服务器ID不能为空")
		}
		if diskPartitionLog.AddWho == "" {
			return fmt.Errorf("创建人ID不能为空")
		}
		if diskPartitionLog.EditWho == "" {
			return fmt.Errorf("修改人ID不能为空")
		}
		if diskPartitionLog.OprSeqFlag == "" {
			return fmt.Errorf("操作序列标识不能为空")
		}
	}
	
	tableName := diskPartitionLogs[0].TableName()
	
	// 转换为interface{}切片
	items := make([]interface{}, len(diskPartitionLogs))
	for i, diskPartitionLog := range diskPartitionLogs {
		items[i] = diskPartitionLog
	}
	
	_, err := dao.db.BatchInsert(ctx, tableName, items, true)
	if err != nil {
		return fmt.Errorf("批量插入磁盘分区日志失败: %w", err)
	}
	return nil
}

// DeleteDiskPartitionLog 删除磁盘分区日志（软删除）
func (dao *DiskPartitionLogDAO) DeleteDiskPartitionLog(ctx context.Context, tenantId, metricDiskPartitionLogId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricDiskPartitionLogId == "" {
		return fmt.Errorf("磁盘分区日志ID不能为空")
	}
	
	tableName := (&types.DiskPartitionLog{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = ? WHERE tenantId = ? AND metricDiskPartitionLogId = ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{types.ActiveFlagNo, tenantId, metricDiskPartitionLogId}, true)
	if err != nil {
		return fmt.Errorf("删除磁盘分区日志失败: %w", err)
	}
	return nil
}

// DeleteDiskPartitionLogByTime 根据时间范围删除磁盘分区日志（物理删除）
func (dao *DiskPartitionLogDAO) DeleteDiskPartitionLogByTime(ctx context.Context, tenantId string, beforeTime time.Time) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	
	tableName := (&types.DiskPartitionLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND collectTime < ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, beforeTime}, true)
	if err != nil {
		return fmt.Errorf("根据时间删除磁盘分区日志失败: %w", err)
	}
	return nil
}

// DeleteDiskPartitionLogByServer 根据服务器ID删除磁盘分区日志（物理删除）
func (dao *DiskPartitionLogDAO) DeleteDiskPartitionLogByServer(ctx context.Context, tenantId, metricServerId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	
	tableName := (&types.DiskPartitionLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND metricServerId = ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, metricServerId}, true)
	if err != nil {
		return fmt.Errorf("根据服务器ID删除磁盘分区日志失败: %w", err)
	}
	return nil
} 