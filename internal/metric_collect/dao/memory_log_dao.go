package dao

import (
	"context"
	"fmt"
	"gohub/internal/metric_collect/types"
	"gohub/pkg/database"
	"time"
)

// MemoryLogDAO 内存采集日志数据访问对象
type MemoryLogDAO struct {
	db database.Database
}

// NewMemoryLogDAO 创建内存采集日志DAO实例
func NewMemoryLogDAO(db database.Database) *MemoryLogDAO {
	return &MemoryLogDAO{db: db}
}

// InsertMemoryLog 插入内存采集日志
func (dao *MemoryLogDAO) InsertMemoryLog(ctx context.Context, memoryLog *types.MemoryLog) error {
	// 设置通用字段默认值
	now := time.Now()
	memoryLog.AddTime = now
	memoryLog.EditTime = now
	memoryLog.CurrentVersion = 1
	memoryLog.ActiveFlag = types.ActiveFlagYes
	
	// 验证必填字段
	if memoryLog.MetricMemoryLogId == "" {
		return fmt.Errorf("内存采集日志ID不能为空")
	}
	if memoryLog.TenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if memoryLog.MetricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	if memoryLog.AddWho == "" {
		return fmt.Errorf("创建人ID不能为空")
	}
	if memoryLog.EditWho == "" {
		return fmt.Errorf("修改人ID不能为空")
	}
	if memoryLog.OprSeqFlag == "" {
		return fmt.Errorf("操作序列标识不能为空")
	}
	
	tableName := memoryLog.TableName()
	_, err := dao.db.Insert(ctx, tableName, memoryLog, true)
	if err != nil {
		return fmt.Errorf("插入内存采集日志失败: %w", err)
	}
	return nil
}

// BatchInsertMemoryLog 批量插入内存采集日志
func (dao *MemoryLogDAO) BatchInsertMemoryLog(ctx context.Context, memoryLogs []*types.MemoryLog) error {
	if len(memoryLogs) == 0 {
		return nil
	}
	
	// 设置通用字段默认值
	now := time.Now()
	for _, memoryLog := range memoryLogs {
		memoryLog.AddTime = now
		memoryLog.EditTime = now
		memoryLog.CurrentVersion = 1
		memoryLog.ActiveFlag = types.ActiveFlagYes
		
		// 验证必填字段
		if memoryLog.MetricMemoryLogId == "" {
			return fmt.Errorf("内存采集日志ID不能为空")
		}
		if memoryLog.TenantId == "" {
			return fmt.Errorf("租户ID不能为空")
		}
		if memoryLog.MetricServerId == "" {
			return fmt.Errorf("服务器ID不能为空")
		}
		if memoryLog.AddWho == "" {
			return fmt.Errorf("创建人ID不能为空")
		}
		if memoryLog.EditWho == "" {
			return fmt.Errorf("修改人ID不能为空")
		}
		if memoryLog.OprSeqFlag == "" {
			return fmt.Errorf("操作序列标识不能为空")
		}
	}
	
	tableName := memoryLogs[0].TableName()
	
	// 转换为interface{}切片
	items := make([]interface{}, len(memoryLogs))
	for i, memoryLog := range memoryLogs {
		items[i] = memoryLog
	}
	
	_, err := dao.db.BatchInsert(ctx, tableName, items, true)
	if err != nil {
		return fmt.Errorf("批量插入内存采集日志失败: %w", err)
	}
	return nil
}

// DeleteMemoryLog 删除内存采集日志（软删除）
func (dao *MemoryLogDAO) DeleteMemoryLog(ctx context.Context, tenantId, metricMemoryLogId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricMemoryLogId == "" {
		return fmt.Errorf("内存采集日志ID不能为空")
	}
	
	tableName := (&types.MemoryLog{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = ? WHERE tenantId = ? AND metricMemoryLogId = ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{types.ActiveFlagNo, tenantId, metricMemoryLogId}, true)
	if err != nil {
		return fmt.Errorf("删除内存采集日志失败: %w", err)
	}
	return nil
}

// DeleteMemoryLogByTime 根据时间范围删除内存采集日志（物理删除）
func (dao *MemoryLogDAO) DeleteMemoryLogByTime(ctx context.Context, tenantId string, beforeTime time.Time) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	
	tableName := (&types.MemoryLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND collectTime < ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, beforeTime}, true)
	if err != nil {
		return fmt.Errorf("根据时间删除内存采集日志失败: %w", err)
	}
	return nil
}

// DeleteMemoryLogByServer 根据服务器ID删除内存采集日志（物理删除）
func (dao *MemoryLogDAO) DeleteMemoryLogByServer(ctx context.Context, tenantId, metricServerId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	
	tableName := (&types.MemoryLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND metricServerId = ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, metricServerId}, true)
	if err != nil {
		return fmt.Errorf("根据服务器ID删除内存采集日志失败: %w", err)
	}
	return nil
} 