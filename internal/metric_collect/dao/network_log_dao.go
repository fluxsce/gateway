package dao

import (
	"context"
	"fmt"
	"gohub/internal/metric_collect/types"
	"gohub/pkg/database"
	"time"
)

// NetworkLogDAO 网络接口日志数据访问对象
type NetworkLogDAO struct {
	db database.Database
}

// NewNetworkLogDAO 创建网络接口日志DAO实例
func NewNetworkLogDAO(db database.Database) *NetworkLogDAO {
	return &NetworkLogDAO{db: db}
}

// InsertNetworkLog 插入网络接口日志
func (dao *NetworkLogDAO) InsertNetworkLog(ctx context.Context, networkLog *types.NetworkLog) error {
	// 设置通用字段默认值
	now := time.Now()
	networkLog.AddTime = now
	networkLog.EditTime = now
	networkLog.CurrentVersion = 1
	networkLog.ActiveFlag = types.ActiveFlagYes
	
	// 验证必填字段
	if networkLog.MetricNetworkLogId == "" {
		return fmt.Errorf("网络接口日志ID不能为空")
	}
	if networkLog.TenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if networkLog.MetricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	if networkLog.AddWho == "" {
		return fmt.Errorf("创建人ID不能为空")
	}
	if networkLog.EditWho == "" {
		return fmt.Errorf("修改人ID不能为空")
	}
	if networkLog.OprSeqFlag == "" {
		return fmt.Errorf("操作序列标识不能为空")
	}
	
	tableName := networkLog.TableName()
	_, err := dao.db.Insert(ctx, tableName, networkLog, true)
	if err != nil {
		return fmt.Errorf("插入网络接口日志失败: %w", err)
	}
	return nil
}

// BatchInsertNetworkLog 批量插入网络接口日志
func (dao *NetworkLogDAO) BatchInsertNetworkLog(ctx context.Context, networkLogs []*types.NetworkLog) error {
	if len(networkLogs) == 0 {
		return nil
	}
	
	// 设置通用字段默认值
	now := time.Now()
	for _, networkLog := range networkLogs {
		networkLog.AddTime = now
		networkLog.EditTime = now
		networkLog.CurrentVersion = 1
		networkLog.ActiveFlag = types.ActiveFlagYes
		
		// 验证必填字段
		if networkLog.MetricNetworkLogId == "" {
			return fmt.Errorf("网络接口日志ID不能为空")
		}
		if networkLog.TenantId == "" {
			return fmt.Errorf("租户ID不能为空")
		}
		if networkLog.MetricServerId == "" {
			return fmt.Errorf("服务器ID不能为空")
		}
		if networkLog.AddWho == "" {
			return fmt.Errorf("创建人ID不能为空")
		}
		if networkLog.EditWho == "" {
			return fmt.Errorf("修改人ID不能为空")
		}
		if networkLog.OprSeqFlag == "" {
			return fmt.Errorf("操作序列标识不能为空")
		}
	}
	
	tableName := networkLogs[0].TableName()
	
	// 转换为interface{}切片
	items := make([]interface{}, len(networkLogs))
	for i, networkLog := range networkLogs {
		items[i] = networkLog
	}
	
	_, err := dao.db.BatchInsert(ctx, tableName, items, true)
	if err != nil {
		return fmt.Errorf("批量插入网络接口日志失败: %w", err)
	}
	return nil
}

// DeleteNetworkLog 删除网络接口日志（软删除）
func (dao *NetworkLogDAO) DeleteNetworkLog(ctx context.Context, tenantId, metricNetworkLogId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricNetworkLogId == "" {
		return fmt.Errorf("网络接口日志ID不能为空")
	}
	
	tableName := (&types.NetworkLog{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = ? WHERE tenantId = ? AND metricNetworkLogId = ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{types.ActiveFlagNo, tenantId, metricNetworkLogId}, true)
	if err != nil {
		return fmt.Errorf("删除网络接口日志失败: %w", err)
	}
	return nil
}

// DeleteNetworkLogByTime 根据时间范围删除网络接口日志（物理删除）
func (dao *NetworkLogDAO) DeleteNetworkLogByTime(ctx context.Context, tenantId string, beforeTime time.Time) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	
	tableName := (&types.NetworkLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND collectTime < ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, beforeTime}, true)
	if err != nil {
		return fmt.Errorf("根据时间删除网络接口日志失败: %w", err)
	}
	return nil
}

// DeleteNetworkLogByServer 根据服务器ID删除网络接口日志（物理删除）
func (dao *NetworkLogDAO) DeleteNetworkLogByServer(ctx context.Context, tenantId, metricServerId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	
	tableName := (&types.NetworkLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND metricServerId = ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, metricServerId}, true)
	if err != nil {
		return fmt.Errorf("根据服务器ID删除网络接口日志失败: %w", err)
	}
	return nil
} 