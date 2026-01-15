package dao

import (
	"context"
	"fmt"
	"gateway/internal/metric_collect/types"
	"gateway/pkg/database"
	"gateway/pkg/utils/ctime"
	"gateway/pkg/utils/huberrors"
	"strings"
)

// MetricDAO 监控数据访问对象
type MetricDAO struct {
	db database.Database
}

// NewMetricDAO 创建监控数据DAO
func NewMetricDAO(db database.Database) *MetricDAO {
	return &MetricDAO{
		db: db,
	}
}

// buildTimeCondition 构建时间条件
func (dao *MetricDAO) buildTimeCondition(startTimeStr, endTimeStr string) (string, []interface{}) {
	var conditions []string
	var params []interface{}

	if startTimeStr != "" {
		if startTime, err := ctime.ParseTimeString(startTimeStr); err == nil {
			conditions = append(conditions, "collectTime >= ?")
			params = append(params, startTime)
		}
	}

	if endTimeStr != "" {
		if endTime, err := ctime.ParseTimeString(endTimeStr); err == nil {
			conditions = append(conditions, "collectTime <= ?")
			params = append(params, endTime)
		}
	}

	if len(conditions) == 0 {
		return "", params
	}

	return strings.Join(conditions, " AND "), params
}

// QueryCPUMetrics 查询CPU监控数据
func (dao *MetricDAO) QueryCPUMetrics(ctx context.Context, tenantId, metricServerId, startTime, endTime string) ([]*types.CpuLog, error) {
	whereClause := "WHERE tenantId = ? AND metricServerId = ? AND activeFlag = 'Y'"
	params := []interface{}{tenantId, metricServerId}

	// 添加时间条件
	timeCondition, timeParams := dao.buildTimeCondition(startTime, endTime)
	if timeCondition != "" {
		whereClause += " AND " + timeCondition
		params = append(params, timeParams...)
	}

	query := fmt.Sprintf(`
		SELECT * FROM %s
		%s
		ORDER BY collectTime DESC
		LIMIT 1000
	`, (&types.CpuLog{}).TableName(), whereClause)

	var results []*types.CpuLog
	err := dao.db.Query(ctx, &results, query, params, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询CPU监控数据失败")
	}

	return results, nil
}

// QueryMemoryMetrics 查询内存监控数据
func (dao *MetricDAO) QueryMemoryMetrics(ctx context.Context, tenantId, metricServerId, startTime, endTime string) ([]*types.MemoryLog, error) {
	whereClause := "WHERE tenantId = ? AND metricServerId = ? AND activeFlag = 'Y'"
	params := []interface{}{tenantId, metricServerId}

	// 添加时间条件
	timeCondition, timeParams := dao.buildTimeCondition(startTime, endTime)
	if timeCondition != "" {
		whereClause += " AND " + timeCondition
		params = append(params, timeParams...)
	}

	query := fmt.Sprintf(`
		SELECT * FROM %s
		%s
		ORDER BY collectTime DESC
		LIMIT 1000
	`, (&types.MemoryLog{}).TableName(), whereClause)

	var results []*types.MemoryLog
	err := dao.db.Query(ctx, &results, query, params, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询内存监控数据失败")
	}

	return results, nil
}

// QueryDiskMetrics 查询磁盘监控数据
func (dao *MetricDAO) QueryDiskMetrics(ctx context.Context, tenantId, metricServerId, startTime, endTime string) ([]*types.DiskPartitionLog, error) {
	whereClause := "WHERE tenantId = ? AND metricServerId = ? AND activeFlag = 'Y'"
	params := []interface{}{tenantId, metricServerId}

	// 添加时间条件
	timeCondition, timeParams := dao.buildTimeCondition(startTime, endTime)
	if timeCondition != "" {
		whereClause += " AND " + timeCondition
		params = append(params, timeParams...)
	}

	query := fmt.Sprintf(`
		SELECT * FROM %s
		%s
		ORDER BY collectTime DESC, deviceName ASC
		LIMIT 1000
	`, (&types.DiskPartitionLog{}).TableName(), whereClause)

	var results []*types.DiskPartitionLog
	err := dao.db.Query(ctx, &results, query, params, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询磁盘监控数据失败")
	}

	return results, nil
}

// QueryDiskIOMetrics 查询磁盘IO监控数据
func (dao *MetricDAO) QueryDiskIOMetrics(ctx context.Context, tenantId, metricServerId, startTime, endTime string) ([]*types.DiskIoLog, error) {
	whereClause := "WHERE tenantId = ? AND metricServerId = ? AND activeFlag = 'Y'"
	params := []interface{}{tenantId, metricServerId}

	// 添加时间条件
	timeCondition, timeParams := dao.buildTimeCondition(startTime, endTime)
	if timeCondition != "" {
		whereClause += " AND " + timeCondition
		params = append(params, timeParams...)
	}

	query := fmt.Sprintf(`
		SELECT * FROM %s
		%s
		ORDER BY collectTime DESC, deviceName ASC
		LIMIT 1000
	`, (&types.DiskIoLog{}).TableName(), whereClause)

	var results []*types.DiskIoLog
	err := dao.db.Query(ctx, &results, query, params, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询磁盘IO监控数据失败")
	}

	return results, nil
}

// QueryNetworkMetrics 查询网络监控数据
func (dao *MetricDAO) QueryNetworkMetrics(ctx context.Context, tenantId, metricServerId, startTime, endTime string) ([]*types.NetworkLog, error) {
	whereClause := "WHERE tenantId = ? AND metricServerId = ? AND activeFlag = 'Y'"
	params := []interface{}{tenantId, metricServerId}

	// 添加时间条件
	timeCondition, timeParams := dao.buildTimeCondition(startTime, endTime)
	if timeCondition != "" {
		whereClause += " AND " + timeCondition
		params = append(params, timeParams...)
	}

	query := fmt.Sprintf(`
		SELECT * FROM %s
		%s
		ORDER BY collectTime DESC, interfaceName ASC
		LIMIT 1000
	`, (&types.NetworkLog{}).TableName(), whereClause)

	var results []*types.NetworkLog
	err := dao.db.Query(ctx, &results, query, params, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询网络监控数据失败")
	}

	return results, nil
}

// QueryProcessMetrics 查询进程监控数据
func (dao *MetricDAO) QueryProcessMetrics(ctx context.Context, tenantId, metricServerId, startTime, endTime string) ([]*types.ProcessStatsLog, error) {
	whereClause := "WHERE tenantId = ? AND metricServerId = ? AND activeFlag = 'Y'"
	params := []interface{}{tenantId, metricServerId}

	// 添加时间条件
	timeCondition, timeParams := dao.buildTimeCondition(startTime, endTime)
	if timeCondition != "" {
		whereClause += " AND " + timeCondition
		params = append(params, timeParams...)
	}

	query := fmt.Sprintf(`
		SELECT * FROM %s
		%s
		ORDER BY collectTime DESC
		LIMIT 1000
	`, (&types.ProcessStatsLog{}).TableName(), whereClause)

	var results []*types.ProcessStatsLog
	err := dao.db.Query(ctx, &results, query, params, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询进程监控数据失败")
	}

	return results, nil
}
