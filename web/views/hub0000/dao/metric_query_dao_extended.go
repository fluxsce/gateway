package dao

import (
	"context"
	"fmt"
	"gateway/internal/metric_collect/types"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0000/models"
	"strings"
)

// QueryDiskIoLogList 查询磁盘IO日志列表
func (dao *MetricQueryDAO) QueryDiskIoLogList(ctx context.Context, req *models.DiskIoLogQueryRequest) ([]*types.DiskIoLog, int, error) {
	var conditions []string
	var params []interface{}

	// 基础条件
	conditions = append(conditions, "tenantId = ?")
	params = append(params, req.TenantId)

	if req.MetricServerId != nil && *req.MetricServerId != "" {
		conditions = append(conditions, "metricServerId = ?")
		params = append(params, *req.MetricServerId)
	}

	// 时间条件
	if timeCondition, timeParams := dao.buildTimeCondition(req.StartTime, req.EndTime); timeCondition != "" {
		conditions = append(conditions, timeCondition)
		params = append(params, timeParams...)
	}

	// 可选条件
	if req.Device != nil && *req.Device != "" {
		conditions = append(conditions, "deviceName LIKE ?")
		params = append(params, "%"+*req.Device+"%")
	}

	if req.MinReadRate != nil {
		conditions = append(conditions, "readRate >= ?")
		params = append(params, *req.MinReadRate)
	}

	if req.MaxReadRate != nil {
		conditions = append(conditions, "readRate <= ?")
		params = append(params, *req.MaxReadRate)
	}

	if req.MinWriteRate != nil {
		conditions = append(conditions, "writeRate >= ?")
		params = append(params, *req.MinWriteRate)
	}

	if req.MaxWriteRate != nil {
		conditions = append(conditions, "writeRate <= ?")
		params = append(params, *req.MaxWriteRate)
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) as total FROM %s %s", (&types.DiskIoLog{}).TableName(), whereClause)
	var totalResult struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询磁盘IO日志总数失败")
	}

	// 构建基础查询语句
	orderByClause := dao.buildOrderByClause(req.OrderBy, req.OrderType)
	baseQuery := fmt.Sprintf("SELECT * FROM %s %s %s",
		(&types.DiskIoLog{}).TableName(), whereClause, orderByClause)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := dao.buildPaginatedQuery(baseQuery, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并参数
	allParams := append(params, paginationArgs...)

	var diskIoLogs []*types.DiskIoLog
	err = dao.db.Query(ctx, &diskIoLogs, paginatedQuery, allParams, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询磁盘IO日志列表失败")
	}

	return diskIoLogs, totalResult.Total, nil
}

// QueryNetworkLogList 查询网络日志列表
func (dao *MetricQueryDAO) QueryNetworkLogList(ctx context.Context, req *models.NetworkLogQueryRequest) ([]*types.NetworkLog, int, error) {
	var conditions []string
	var params []interface{}

	// 基础条件
	conditions = append(conditions, "tenantId = ?")
	params = append(params, req.TenantId)

	if req.MetricServerId != nil && *req.MetricServerId != "" {
		conditions = append(conditions, "metricServerId = ?")
		params = append(params, *req.MetricServerId)
	}

	// 时间条件
	if timeCondition, timeParams := dao.buildTimeCondition(req.StartTime, req.EndTime); timeCondition != "" {
		conditions = append(conditions, timeCondition)
		params = append(params, timeParams...)
	}

	// 可选条件
	if req.InterfaceName != nil && *req.InterfaceName != "" {
		conditions = append(conditions, "interfaceName = ?")
		params = append(params, *req.InterfaceName)
	}

	if req.MinBytesRecv != nil {
		conditions = append(conditions, "bytesReceived >= ?")
		params = append(params, *req.MinBytesRecv)
	}

	if req.MaxBytesRecv != nil {
		conditions = append(conditions, "bytesReceived <= ?")
		params = append(params, *req.MaxBytesRecv)
	}

	if req.MinBytesSent != nil {
		conditions = append(conditions, "bytesSent >= ?")
		params = append(params, *req.MinBytesSent)
	}

	if req.MaxBytesSent != nil {
		conditions = append(conditions, "bytesSent <= ?")
		params = append(params, *req.MaxBytesSent)
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) as total FROM %s %s", (&types.NetworkLog{}).TableName(), whereClause)
	var totalResult struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网络日志总数失败")
	}

	// 构建基础查询语句
	orderByClause := dao.buildOrderByClause(req.OrderBy, req.OrderType)
	baseQuery := fmt.Sprintf("SELECT * FROM %s %s %s",
		(&types.NetworkLog{}).TableName(), whereClause, orderByClause)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := dao.buildPaginatedQuery(baseQuery, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并参数
	allParams := append(params, paginationArgs...)

	var networkLogs []*types.NetworkLog
	err = dao.db.Query(ctx, &networkLogs, paginatedQuery, allParams, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询网络日志列表失败")
	}

	return networkLogs, totalResult.Total, nil
}

// QueryProcessLogList 查询进程日志列表
func (dao *MetricQueryDAO) QueryProcessLogList(ctx context.Context, req *models.ProcessLogQueryRequest) ([]*types.ProcessLog, int, error) {
	var conditions []string
	var params []interface{}

	// 基础条件
	conditions = append(conditions, "tenantId = ?")
	params = append(params, req.TenantId)

	if req.MetricServerId != nil && *req.MetricServerId != "" {
		conditions = append(conditions, "metricServerId = ?")
		params = append(params, *req.MetricServerId)
	}

	// 时间条件
	if timeCondition, timeParams := dao.buildTimeCondition(req.StartTime, req.EndTime); timeCondition != "" {
		conditions = append(conditions, timeCondition)
		params = append(params, timeParams...)
	}

	// 可选条件
	if req.ProcessName != nil && *req.ProcessName != "" {
		conditions = append(conditions, "processName LIKE ?")
		params = append(params, "%"+*req.ProcessName+"%")
	}

	if req.ProcessOwner != nil && *req.ProcessOwner != "" {
		conditions = append(conditions, "processOwner = ?")
		params = append(params, *req.ProcessOwner)
	}

	if req.MinPid != nil {
		conditions = append(conditions, "processId >= ?")
		params = append(params, *req.MinPid)
	}

	if req.MaxPid != nil {
		conditions = append(conditions, "processId <= ?")
		params = append(params, *req.MaxPid)
	}

	if req.MinCpuPercent != nil {
		conditions = append(conditions, "cpuPercent >= ?")
		params = append(params, *req.MinCpuPercent)
	}

	if req.MaxCpuPercent != nil {
		conditions = append(conditions, "cpuPercent <= ?")
		params = append(params, *req.MaxCpuPercent)
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) as total FROM %s %s", (&types.ProcessLog{}).TableName(), whereClause)
	var totalResult struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询进程日志总数失败")
	}

	// 构建基础查询语句
	orderByClause := dao.buildOrderByClause(req.OrderBy, req.OrderType)
	baseQuery := fmt.Sprintf("SELECT * FROM %s %s %s",
		(&types.ProcessLog{}).TableName(), whereClause, orderByClause)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := dao.buildPaginatedQuery(baseQuery, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并参数
	allParams := append(params, paginationArgs...)

	var processLogs []*types.ProcessLog
	err = dao.db.Query(ctx, &processLogs, paginatedQuery, allParams, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询进程日志列表失败")
	}

	return processLogs, totalResult.Total, nil
}

// QueryProcessStatsLogList 查询进程统计日志列表
func (dao *MetricQueryDAO) QueryProcessStatsLogList(ctx context.Context, req *models.ProcessStatsLogQueryRequest) ([]*types.ProcessStatsLog, int, error) {
	var conditions []string
	var params []interface{}

	// 基础条件
	conditions = append(conditions, "tenantId = ?")
	params = append(params, req.TenantId)

	if req.MetricServerId != nil && *req.MetricServerId != "" {
		conditions = append(conditions, "metricServerId = ?")
		params = append(params, *req.MetricServerId)
	}

	// 时间条件
	if timeCondition, timeParams := dao.buildTimeCondition(req.StartTime, req.EndTime); timeCondition != "" {
		conditions = append(conditions, timeCondition)
		params = append(params, timeParams...)
	}

	// 可选条件
	if req.MinProcessCount != nil {
		conditions = append(conditions, "totalCount >= ?")
		params = append(params, *req.MinProcessCount)
	}

	if req.MaxProcessCount != nil {
		conditions = append(conditions, "totalCount <= ?")
		params = append(params, *req.MaxProcessCount)
	}

	if req.MinThreadCount != nil {
		conditions = append(conditions, "threadCount >= ?")
		params = append(params, *req.MinThreadCount)
	}

	if req.MaxThreadCount != nil {
		conditions = append(conditions, "threadCount <= ?")
		params = append(params, *req.MaxThreadCount)
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) as total FROM %s %s", (&types.ProcessStatsLog{}).TableName(), whereClause)
	var totalResult struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询进程统计日志总数失败")
	}

	// 构建基础查询语句
	orderByClause := dao.buildOrderByClause(req.OrderBy, req.OrderType)
	baseQuery := fmt.Sprintf("SELECT * FROM %s %s %s",
		(&types.ProcessStatsLog{}).TableName(), whereClause, orderByClause)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := dao.buildPaginatedQuery(baseQuery, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并参数
	allParams := append(params, paginationArgs...)

	var processStatsLogs []*types.ProcessStatsLog
	err = dao.db.Query(ctx, &processStatsLogs, paginatedQuery, allParams, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询进程统计日志列表失败")
	}

	return processStatsLogs, totalResult.Total, nil
}

// QueryTemperatureLogList 查询温度日志列表
func (dao *MetricQueryDAO) QueryTemperatureLogList(ctx context.Context, req *models.TemperatureLogQueryRequest) ([]*types.TemperatureLog, int, error) {
	var conditions []string
	var params []interface{}

	// 基础条件
	conditions = append(conditions, "tenantId = ?")
	params = append(params, req.TenantId)

	if req.MetricServerId != nil && *req.MetricServerId != "" {
		conditions = append(conditions, "metricServerId = ?")
		params = append(params, *req.MetricServerId)
	}

	// 时间条件
	if timeCondition, timeParams := dao.buildTimeCondition(req.StartTime, req.EndTime); timeCondition != "" {
		conditions = append(conditions, timeCondition)
		params = append(params, timeParams...)
	}

	// 可选条件
	if req.SensorName != nil && *req.SensorName != "" {
		conditions = append(conditions, "sensorName LIKE ?")
		params = append(params, "%"+*req.SensorName+"%")
	}

	if req.MinTemperature != nil {
		conditions = append(conditions, "temperatureValue >= ?")
		params = append(params, *req.MinTemperature)
	}

	if req.MaxTemperature != nil {
		conditions = append(conditions, "temperatureValue <= ?")
		params = append(params, *req.MaxTemperature)
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) as total FROM %s %s", (&types.TemperatureLog{}).TableName(), whereClause)
	var totalResult struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询温度日志总数失败")
	}

	// 构建基础查询语句
	orderByClause := dao.buildOrderByClause(req.OrderBy, req.OrderType)
	baseQuery := fmt.Sprintf("SELECT * FROM %s %s %s",
		(&types.TemperatureLog{}).TableName(), whereClause, orderByClause)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := dao.buildPaginatedQuery(baseQuery, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并参数
	allParams := append(params, paginationArgs...)

	var temperatureLogs []*types.TemperatureLog
	err = dao.db.Query(ctx, &temperatureLogs, paginatedQuery, allParams, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询温度日志列表失败")
	}

	return temperatureLogs, totalResult.Total, nil
}
