package dao

import (
	"context"
	"fmt"
	"gateway/internal/metric_collect/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/ctime"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0000/models"
	"strings"
)

// MetricQueryDAO 指标查询数据访问对象
type MetricQueryDAO struct {
	db database.Database
}

// NewMetricQueryDAO 创建指标查询DAO
func NewMetricQueryDAO(db database.Database) *MetricQueryDAO {
	return &MetricQueryDAO{
		db: db,
	}
}

// buildTimeCondition 构建时间条件
// 接受时间字符串参数，解析后使用time.Time类型进行数据库查询
func (dao *MetricQueryDAO) buildTimeCondition(startTimeStr, endTimeStr string) (string, []interface{}) {
	var conditions []string
	var params []interface{}

	if startTimeStr != "" {
		// 使用ctime解析时间字符串，但参数使用time.Time类型
		if startTime, err := ctime.ParseTimeString(startTimeStr); err == nil {
			conditions = append(conditions, "collectTime >= ?")
			params = append(params, startTime)
		}
	}

	if endTimeStr != "" {
		// 使用ctime解析时间字符串，但参数使用time.Time类型
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

// buildTimeConditionForServerInfo 构建服务器信息时间条件
// 服务器信息表使用lastUpdateTime字段
func (dao *MetricQueryDAO) buildTimeConditionForServerInfo(startTimeStr, endTimeStr string) (string, []interface{}) {
	var conditions []string
	var params []interface{}

	if startTimeStr != "" {
		if startTime, err := ctime.ParseTimeString(startTimeStr); err == nil {
			conditions = append(conditions, "lastUpdateTime >= ?")
			params = append(params, startTime)
		}
	}

	if endTimeStr != "" {
		if endTime, err := ctime.ParseTimeString(endTimeStr); err == nil {
			conditions = append(conditions, "lastUpdateTime <= ?")
			params = append(params, endTime)
		}
	}

	if len(conditions) == 0 {
		return "", params
	}

	return strings.Join(conditions, " AND "), params
}

// buildOrderByClause 构建排序条件
func (dao *MetricQueryDAO) buildOrderByClause(orderBy, orderType string) string {
	// 默认排序
	if orderBy == "" {
		orderBy = "collectTime"
	}
	if orderType == "" {
		orderType = "DESC"
	}

	// 验证排序类型
	if orderType != "ASC" && orderType != "DESC" {
		orderType = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", orderBy, orderType)
}

// buildOrderByClauseForServerInfo 构建服务器信息排序条件
func (dao *MetricQueryDAO) buildOrderByClauseForServerInfo(orderBy, orderType string) string {
	// 服务器信息表默认排序字段为 lastUpdateTime
	if orderBy == "" {
		orderBy = "lastUpdateTime"
	}
	if orderType == "" {
		orderType = "DESC"
	}

	// 验证排序类型
	if orderType != "ASC" && orderType != "DESC" {
		orderType = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", orderBy, orderType)
}

// buildPaginatedQuery 构建分页查询语句
func (dao *MetricQueryDAO) buildPaginatedQuery(baseQuery string, page, pageSize int) (string, []interface{}, error) {
	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 构建分页查询
	return sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
}

// QueryServerInfoList 查询服务器信息列表
func (dao *MetricQueryDAO) QueryServerInfoList(ctx context.Context, req *models.ServerInfoQueryRequest) ([]*types.ServerInfo, int, error) {
	var conditions []string
	var params []interface{}

	// 基础条件：只查询指定租户的活跃服务器
	conditions = append(conditions, "tenantId = ?")
	params = append(params, req.TenantId)

	conditions = append(conditions, "activeFlag = ?")
	params = append(params, "Y")

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) as total FROM %s %s", (&types.ServerInfo{}).TableName(), whereClause)
	var totalResult struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务器信息总数失败")
	}

	// 构建基础查询语句
	orderByClause := dao.buildOrderByClauseForServerInfo(req.OrderBy, req.OrderType)
	baseQuery := fmt.Sprintf("SELECT * FROM %s %s %s",
		(&types.ServerInfo{}).TableName(), whereClause, orderByClause)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := dao.buildPaginatedQuery(baseQuery, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并参数
	allParams := append(params, paginationArgs...)

	var servers []*types.ServerInfo
	err = dao.db.Query(ctx, &servers, paginatedQuery, allParams, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询服务器信息列表失败")
	}

	return servers, totalResult.Total, nil
}

// GetServerInfoDetail 获取服务器信息详情
func (dao *MetricQueryDAO) GetServerInfoDetail(ctx context.Context, tenantId, serverId string) (*types.ServerInfo, error) {
	if tenantId == "" || serverId == "" {
		return nil, fmt.Errorf("tenantId和serverId不能为空")
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND metricServerId = ?", (&types.ServerInfo{}).TableName())

	var server types.ServerInfo
	err := dao.db.QueryOne(ctx, &server, query, []interface{}{tenantId, serverId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询服务器信息详情失败")
	}

	return &server, nil
}

// QueryCpuLogList 查询CPU性能日志列表
func (dao *MetricQueryDAO) QueryCpuLogList(ctx context.Context, req *models.CpuLogQueryRequest) ([]*types.CpuLog, int, error) {
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
	if req.CpuCore != nil && *req.CpuCore != "" {
		conditions = append(conditions, "cpuCore = ?")
		params = append(params, *req.CpuCore)
	}

	if req.MinUsagePercent != nil {
		conditions = append(conditions, "usagePercent >= ?")
		params = append(params, *req.MinUsagePercent)
	}

	if req.MaxUsagePercent != nil {
		conditions = append(conditions, "usagePercent <= ?")
		params = append(params, *req.MaxUsagePercent)
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) as total FROM %s %s", (&types.CpuLog{}).TableName(), whereClause)
	var totalResult struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询CPU日志总数失败")
	}

	// 构建基础查询语句
	orderByClause := dao.buildOrderByClause(req.OrderBy, req.OrderType)
	baseQuery := fmt.Sprintf("SELECT * FROM %s %s %s",
		(&types.CpuLog{}).TableName(), whereClause, orderByClause)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := dao.buildPaginatedQuery(baseQuery, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并参数
	allParams := append(params, paginationArgs...)

	var cpuLogs []*types.CpuLog
	err = dao.db.Query(ctx, &cpuLogs, paginatedQuery, allParams, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询CPU日志列表失败")
	}

	return cpuLogs, totalResult.Total, nil
}

// QueryMemoryLogList 查询内存性能日志列表
func (dao *MetricQueryDAO) QueryMemoryLogList(ctx context.Context, req *models.MemoryLogQueryRequest) ([]*types.MemoryLog, int, error) {
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
	if req.MinUsagePercent != nil {
		conditions = append(conditions, "usagePercent >= ?")
		params = append(params, *req.MinUsagePercent)
	}

	if req.MaxUsagePercent != nil {
		conditions = append(conditions, "usagePercent <= ?")
		params = append(params, *req.MaxUsagePercent)
	}

	if req.MinAvailableGB != nil {
		conditions = append(conditions, "availableGB >= ?")
		params = append(params, *req.MinAvailableGB)
	}

	if req.MaxAvailableGB != nil {
		conditions = append(conditions, "availableGB <= ?")
		params = append(params, *req.MaxAvailableGB)
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) as total FROM %s %s", (&types.MemoryLog{}).TableName(), whereClause)
	var totalResult struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询内存日志总数失败")
	}

	// 构建基础查询语句
	orderByClause := dao.buildOrderByClause(req.OrderBy, req.OrderType)
	baseQuery := fmt.Sprintf("SELECT * FROM %s %s %s",
		(&types.MemoryLog{}).TableName(), whereClause, orderByClause)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := dao.buildPaginatedQuery(baseQuery, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并参数
	allParams := append(params, paginationArgs...)

	var memoryLogs []*types.MemoryLog
	err = dao.db.Query(ctx, &memoryLogs, paginatedQuery, allParams, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询内存日志列表失败")
	}

	return memoryLogs, totalResult.Total, nil
}

// QueryDiskPartitionLogList 查询磁盘分区日志列表
func (dao *MetricQueryDAO) QueryDiskPartitionLogList(ctx context.Context, req *models.DiskPartitionLogQueryRequest) ([]*types.DiskPartitionLog, int, error) {
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

	if req.MountPoint != nil && *req.MountPoint != "" {
		conditions = append(conditions, "mountPoint = ?")
		params = append(params, *req.MountPoint)
	}

	if req.FsType != nil && *req.FsType != "" {
		conditions = append(conditions, "fileSystem = ?")
		params = append(params, *req.FsType)
	}

	if req.MinUsagePercent != nil {
		conditions = append(conditions, "usagePercent >= ?")
		params = append(params, *req.MinUsagePercent)
	}

	if req.MaxUsagePercent != nil {
		conditions = append(conditions, "usagePercent <= ?")
		params = append(params, *req.MaxUsagePercent)
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) as total FROM %s %s", (&types.DiskPartitionLog{}).TableName(), whereClause)
	var totalResult struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalResult, countQuery, params, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询磁盘分区日志总数失败")
	}

	// 构建基础查询语句
	orderByClause := dao.buildOrderByClause(req.OrderBy, req.OrderType)
	baseQuery := fmt.Sprintf("SELECT * FROM %s %s %s",
		(&types.DiskPartitionLog{}).TableName(), whereClause, orderByClause)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := dao.buildPaginatedQuery(baseQuery, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并参数
	allParams := append(params, paginationArgs...)

	var diskPartitionLogs []*types.DiskPartitionLog
	err = dao.db.Query(ctx, &diskPartitionLogs, paginatedQuery, allParams, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询磁盘分区日志列表失败")
	}

	return diskPartitionLogs, totalResult.Total, nil
}
