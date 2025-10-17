package dao

import (
	"context"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/ctime"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0042/models"
	"strings"
	"time"
)

// JvmQueryDao JVM监控查询DAO
type JvmQueryDao struct {
	db database.Database
}

// NewJvmQueryDao 创建JVM查询DAO实例
func NewJvmQueryDao(db database.Database) *JvmQueryDao {
	return &JvmQueryDao{
		db: db,
	}
}

// parseTimeString 解析时间字符串为time.Time，如果字符串为空则返回零值
// 使用ctime工具包进行解析，支持多种时间格式
func (dao *JvmQueryDao) parseTimeString(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}
	return ctime.ParseTimeString(timeStr)
}

// addTimeCondition 添加时间范围条件到WHERE子句
// 简化时间字符串解析和条件添加的通用逻辑
func (dao *JvmQueryDao) addTimeCondition(ctx context.Context, whereClauses *[]string, params *[]interface{},
	startTimeStr, endTimeStr, fieldName string) error {

	// 解析开始时间
	if startTimeStr != "" {
		startTime, err := dao.parseTimeString(startTimeStr)
		if err != nil {
			logger.ErrorWithTrace(ctx, "开始时间解析失败", err)
			return huberrors.WrapError(err, "开始时间格式错误")
		}
		if !startTime.IsZero() {
			*whereClauses = append(*whereClauses, fieldName+" >= ?")
			*params = append(*params, startTime)
		}
	}

	// 解析结束时间
	if endTimeStr != "" {
		endTime, err := dao.parseTimeString(endTimeStr)
		if err != nil {
			logger.ErrorWithTrace(ctx, "结束时间解析失败", err)
			return huberrors.WrapError(err, "结束时间格式错误")
		}
		if !endTime.IsZero() {
			*whereClauses = append(*whereClauses, fieldName+" <= ?")
			*params = append(*params, endTime)
		}
	}

	return nil
}

// ===============================
// JVM资源监控查询
// ===============================

// QueryJvmResources 查询JVM资源列表
func (dao *JvmQueryDao) QueryJvmResources(ctx context.Context, req *models.JvmResourceQueryRequest) (*models.JvmResourceListResponse, error) {
	// 构建WHERE条件
	whereClauses := []string{"r.tenantId = ?"}
	params := []interface{}{req.TenantId}

	if req.ServiceGroupId != "" {
		whereClauses = append(whereClauses, "r.serviceGroupId = ?")
		params = append(params, req.ServiceGroupId)
	}

	if req.GroupName != "" {
		whereClauses = append(whereClauses, "r.groupName LIKE ?")
		params = append(params, "%"+req.GroupName+"%")
	}

	if req.JvmResourceId != "" {
		whereClauses = append(whereClauses, "r.jvmResourceId = ?")
		params = append(params, req.JvmResourceId)
	}

	if req.ApplicationName != "" {
		whereClauses = append(whereClauses, "r.applicationName LIKE ?")
		params = append(params, "%"+req.ApplicationName+"%")
	}

	if req.HostIpAddress != "" {
		whereClauses = append(whereClauses, "r.hostIpAddress LIKE ?")
		params = append(params, "%"+req.HostIpAddress+"%")
	}

	if req.HealthyFlag != "" {
		whereClauses = append(whereClauses, "r.healthyFlag = ?")
		params = append(params, req.HealthyFlag)
	}

	if req.RequiresAttention != "" {
		whereClauses = append(whereClauses, "r.requiresAttentionFlag = ?")
		params = append(params, req.RequiresAttention)
	}

	// 添加时间范围条件
	if err := dao.addTimeCondition(ctx, &whereClauses, &params, req.StartTime, req.EndTime, "r.collectionTime"); err != nil {
		return nil, err
	}

	whereClause := "WHERE " + strings.Join(whereClauses, " AND ") + " AND r.activeFlag = 'Y'"

	// 构建排序
	orderBy := "r.collectionTime"
	if req.OrderBy != "" {
		orderBy = "r." + req.OrderBy
	}
	orderDirection := "DESC"
	if req.OrderDirection != "" {
		orderDirection = req.OrderDirection
	}

	// 构建基础查询语句
	baseQuery := fmt.Sprintf(`
		SELECT 
			r.jvmResourceId, r.tenantId, r.serviceGroupId, r.applicationName, r.groupName,
			r.hostName, r.hostIpAddress, r.collectionTime, r.jvmStartTime, r.jvmUptimeMs,
			r.healthyFlag, r.healthGrade, r.requiresAttentionFlag, r.summaryText,
			r.addTime, r.editTime, r.oprSeqFlag, r.currentVersion, r.activeFlag
		FROM HUB_MONITOR_JVM_RESOURCE r
		%s
		ORDER BY %s %s
	`, whereClause, orderBy, orderDirection)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建统计查询失败", err)
		return nil, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var countResult struct {
		Count int64 `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询JVM资源总数失败", err)
		return nil, huberrors.WrapError(err, "查询JVM资源总数失败")
	}

	// 如果没有记录，直接返回空列表
	if countResult.Count == 0 {
		return &models.JvmResourceListResponse{
			PageInfo: models.PageInfo{
				PageNum:    req.PageNum,
				PageSize:   req.PageSize,
				TotalCount: 0,
				TotalPages: 0,
			},
			List: []models.JvmResourceResponse{},
		}, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(req.PageNum, req.PageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建分页查询失败", err)
		return nil, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	// 执行分页查询
	var list []models.JvmResourceResponse
	err = dao.db.Query(ctx, &list, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询JVM资源列表失败", err)
		return nil, huberrors.WrapError(err, "查询JVM资源列表失败")
	}

	// 计算总页数
	totalPages := int((countResult.Count + int64(pagination.PageSize) - 1) / int64(pagination.PageSize))

	return &models.JvmResourceListResponse{
		PageInfo: models.PageInfo{
			PageNum:    req.PageNum,
			PageSize:   req.PageSize,
			TotalCount: countResult.Count,
			TotalPages: totalPages,
		},
		List: list,
	}, nil
}

// GetJvmResourceDetail 获取JVM资源详情
func (dao *JvmQueryDao) GetJvmResourceDetail(ctx context.Context, tenantId, jvmResourceId string) (*models.JvmResourceResponse, error) {
	sql := `
		SELECT 
			r.jvmResourceId, r.tenantId, r.serviceGroupId, r.applicationName, r.groupName,
			r.hostName, r.hostIpAddress, r.collectionTime, r.jvmStartTime, r.jvmUptimeMs,
			r.healthyFlag, r.healthGrade, r.requiresAttentionFlag, r.summaryText,
			r.systemPropertiesJson,
			r.addTime, r.addWho, r.editTime, r.editWho, r.oprSeqFlag, r.currentVersion, r.activeFlag, r.noteText
		FROM HUB_MONITOR_JVM_RESOURCE r
		WHERE r.tenantId = ? AND r.jvmResourceId = ? AND r.activeFlag = 'Y'
		ORDER BY r.collectionTime DESC
		LIMIT 1
	`

	var result models.JvmResourceResponse
	err := dao.db.QueryOne(ctx, &result, sql, []interface{}{tenantId, jvmResourceId}, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询JVM资源详情失败", err)
		return nil, huberrors.WrapError(err, "查询JVM资源详情失败")
	}

	return &result, nil
}

// ===============================
// GC快照查询
// ===============================

// QueryGCSnapshots 查询GC快照列表
func (dao *JvmQueryDao) QueryGCSnapshots(ctx context.Context, req *models.GCSnapshotQueryRequest) ([]models.GCSnapshotResponse, error) {
	whereClauses := []string{"g.tenantId = ?", "g.jvmResourceId = ?", "g.activeFlag = 'Y'"}
	params := []interface{}{req.TenantId, req.JvmResourceId}

	// 添加时间范围条件
	if err := dao.addTimeCondition(ctx, &whereClauses, &params, req.StartTime, req.EndTime, "g.collectionTime"); err != nil {
		return nil, err
	}

	whereClause := "WHERE " + strings.Join(whereClauses, " AND ")

	limit := req.Limit
	if limit <= 0 {
		limit = 200 // 默认返回最近200条
	}

	// 构建基础查询语句 - 使用DESC倒序获取最新的N条数据
	baseQuery := fmt.Sprintf(`
		SELECT 
			g.gcSnapshotId, g.tenantId, g.jvmResourceId, g.collectionTime,
			g.collectionCount, g.collectionTimeMs,
			g.s0c, g.s1c, g.s0u, g.s1u,
			g.ec, g.eu,
			g.oc, g.ou,
			g.mc, g.mu,
			g.ccsc, g.ccsu,
			g.ygc, g.ygct, g.fgc, g.fgct, g.gct,
			g.addTime, g.editTime, g.oprSeqFlag, g.currentVersion, g.activeFlag
		FROM HUB_MONITOR_JVM_GC g
		%s
		ORDER BY g.collectionTime DESC
	`, whereClause)

	// 使用统一的分页构建工具，支持多数据库
	pagination := sqlutils.NewPaginationInfo(1, limit)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建分页查询失败", err)
		return nil, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	var list []models.GCSnapshotResponse
	err = dao.db.Query(ctx, &list, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询GC快照列表失败", err)
		return nil, huberrors.WrapError(err, "查询GC快照列表失败")
	}

	// 反转结果顺序，使前端按时间顺序展示（从旧到新）
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return list, nil
}

// GetLatestGCSnapshot 获取最新的GC快照
func (dao *JvmQueryDao) GetLatestGCSnapshot(ctx context.Context, tenantId, jvmResourceId string) (*models.GCSnapshotResponse, error) {
	sql := `
		SELECT 
			g.gcSnapshotId, g.tenantId, g.jvmResourceId, g.collectionTime,
			g.collectionCount, g.collectionTimeMs,
			g.s0c, g.s1c, g.s0u, g.s1u,
			g.ec, g.eu,
			g.oc, g.ou,
			g.mc, g.mu,
			g.ccsc, g.ccsu,
			g.ygc, g.ygct, g.fgc, g.fgct, g.gct,
			g.addTime, g.editTime, g.oprSeqFlag, g.currentVersion, g.activeFlag
		FROM HUB_MONITOR_JVM_GC g
		WHERE g.tenantId = ? AND g.jvmResourceId = ? AND g.activeFlag = 'Y'
		ORDER BY g.collectionTime DESC
		LIMIT 1
	`

	var result models.GCSnapshotResponse
	err := dao.db.QueryOne(ctx, &result, sql, []interface{}{tenantId, jvmResourceId}, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询最新GC快照失败", err)
		return nil, huberrors.WrapError(err, "查询最新GC快照失败")
	}

	return &result, nil
}

// ===============================
// 内存监控查询
// ===============================

// QueryMemory 查询内存记录列表
func (dao *JvmQueryDao) QueryMemory(ctx context.Context, req *models.MemoryQueryRequest) ([]models.MemoryResponse, error) {
	whereClauses := []string{"m.tenantId = ?", "m.jvmResourceId = ?"}
	params := []interface{}{req.TenantId, req.JvmResourceId}

	if req.MemoryType != "" {
		whereClauses = append(whereClauses, "m.memoryType = ?")
		params = append(params, req.MemoryType)
	}

	// 添加时间范围条件
	if err := dao.addTimeCondition(ctx, &whereClauses, &params, req.StartTime, req.EndTime, "m.collectionTime"); err != nil {
		return nil, err
	}

	whereClause := strings.Join(whereClauses, " AND ")

	limit := req.Limit
	if limit <= 0 {
		limit = 200 // 默认返回最近200条
	}

	// 构建基础查询语句 - 使用DESC倒序获取最新的N条数据
	baseQuery := fmt.Sprintf(`
		SELECT 
			m.jvmMemoryId, m.jvmResourceId, m.memoryType,
			m.initMemoryBytes, m.usedMemoryBytes, m.committedMemoryBytes, m.maxMemoryBytes,
			m.usagePercent, m.healthyFlag, m.collectionTime
		FROM HUB_MONITOR_JVM_MEMORY m
		WHERE %s AND m.activeFlag = 'Y'
		ORDER BY m.collectionTime DESC
	`, whereClause)

	// 使用统一的分页构建工具，支持多数据库
	pagination := sqlutils.NewPaginationInfo(1, limit)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建分页查询失败", err)
		return nil, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	var list []models.MemoryResponse
	err = dao.db.Query(ctx, &list, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询内存记录失败", err)
		return nil, huberrors.WrapError(err, "查询内存记录失败")
	}

	// 反转结果顺序，使前端按时间顺序展示（从旧到新）
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return list, nil
}

// ===============================
// 内存池查询
// ===============================

// QueryMemoryPools 查询内存池列表
func (dao *JvmQueryDao) QueryMemoryPools(ctx context.Context, req *models.MemoryPoolQueryRequest) ([]models.MemoryPoolResponse, error) {
	whereClauses := []string{"mp.tenantId = ?", "mp.jvmResourceId = ?"}
	params := []interface{}{req.TenantId, req.JvmResourceId}

	if req.PoolType != "" {
		whereClauses = append(whereClauses, "mp.poolType = ?")
		params = append(params, req.PoolType)
	}

	if req.PoolCategory != "" {
		whereClauses = append(whereClauses, "mp.poolCategory = ?")
		params = append(params, req.PoolCategory)
	}

	// 添加时间范围条件
	if err := dao.addTimeCondition(ctx, &whereClauses, &params, req.StartTime, req.EndTime, "mp.collectionTime"); err != nil {
		return nil, err
	}

	whereClause := strings.Join(whereClauses, " AND ")

	limit := req.Limit
	if limit <= 0 {
		limit = 200 // 默认返回最近200条
	}

	// 构建基础查询语句 - 使用DESC倒序获取最新的N条数据
	baseQuery := fmt.Sprintf(`
		SELECT 
			mp.memoryPoolId, mp.jvmResourceId, mp.poolName, mp.poolType, mp.poolCategory,
			mp.currentUsedBytes, mp.currentCommittedBytes, mp.currentMaxBytes, mp.currentUsagePercent,
			mp.peakUsedBytes, mp.peakUsagePercent, mp.healthyFlag, mp.collectionTime
		FROM HUB_MONITOR_JVM_MEM_POOL mp
		WHERE %s AND mp.activeFlag = 'Y'
		ORDER BY mp.collectionTime DESC
	`, whereClause)

	// 使用统一的分页构建工具，支持多数据库
	pagination := sqlutils.NewPaginationInfo(1, limit)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建分页查询失败", err)
		return nil, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	var list []models.MemoryPoolResponse
	err = dao.db.Query(ctx, &list, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询内存池列表失败", err)
		return nil, huberrors.WrapError(err, "查询内存池列表失败")
	}

	// 反转结果顺序，使前端按时间顺序展示（从旧到新）
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return list, nil
}

// ===============================
// 线程监控查询
// ===============================

// QueryThreads 查询线程记录列表
func (dao *JvmQueryDao) QueryThreads(ctx context.Context, req *models.ThreadQueryRequest) ([]models.ThreadResponse, error) {
	whereClauses := []string{"t.tenantId = ?", "t.jvmResourceId = ?"}
	params := []interface{}{req.TenantId, req.JvmResourceId}

	if req.HealthyFlag != "" {
		whereClauses = append(whereClauses, "t.healthyFlag = ?")
		params = append(params, req.HealthyFlag)
	}

	// 添加时间范围条件
	if err := dao.addTimeCondition(ctx, &whereClauses, &params, req.StartTime, req.EndTime, "t.collectionTime"); err != nil {
		return nil, err
	}

	whereClause := strings.Join(whereClauses, " AND ")

	limit := req.Limit
	if limit <= 0 {
		limit = 200 // 默认返回最近200条
	}

	// 构建基础查询语句 - 使用DESC倒序获取最新的N条数据
	baseQuery := fmt.Sprintf(`
		SELECT 
			t.jvmThreadId, t.jvmResourceId,
			t.currentThreadCount, t.daemonThreadCount, t.userThreadCount,
			t.peakThreadCount, t.totalStartedThreadCount,
			t.threadGrowthRatePercent, t.daemonThreadRatioPercent,
			t.healthyFlag, t.healthGrade, t.requiresAttentionFlag, t.collectionTime
		FROM HUB_MONITOR_JVM_THREAD t
		WHERE %s AND t.activeFlag = 'Y'
		ORDER BY t.collectionTime DESC
	`, whereClause)

	// 使用统一的分页构建工具，支持多数据库
	pagination := sqlutils.NewPaginationInfo(1, limit)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建分页查询失败", err)
		return nil, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	var list []models.ThreadResponse
	err = dao.db.Query(ctx, &list, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询线程记录失败", err)
		return nil, huberrors.WrapError(err, "查询线程记录失败")
	}

	// 反转结果顺序，使前端按时间顺序展示（从旧到新）
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return list, nil
}

// QueryThreadStates 查询线程状态记录列表
func (dao *JvmQueryDao) QueryThreadStates(ctx context.Context, req *models.ThreadStateQueryRequest) ([]models.ThreadStateResponse, error) {
	whereClauses := []string{"ts.tenantId = ?", "ts.jvmResourceId = ?"}
	params := []interface{}{req.TenantId, req.JvmResourceId}

	if req.JvmThreadId != "" {
		whereClauses = append(whereClauses, "ts.jvmThreadId = ?")
		params = append(params, req.JvmThreadId)
	}

	// 添加时间范围条件
	if err := dao.addTimeCondition(ctx, &whereClauses, &params, req.StartTime, req.EndTime, "ts.collectionTime"); err != nil {
		return nil, err
	}

	whereClause := strings.Join(whereClauses, " AND ")

	limit := req.Limit
	if limit <= 0 {
		limit = 200 // 默认返回最近200条
	}

	// 构建基础查询语句 - 使用DESC倒序获取最新的N条数据
	baseQuery := fmt.Sprintf(`
		SELECT 
			ts.threadStateId, ts.tenantId, ts.jvmThreadId, ts.jvmResourceId,
			ts.newThreadCount, ts.runnableThreadCount, ts.blockedThreadCount,
			ts.waitingThreadCount, ts.timedWaitingThreadCount, ts.terminatedThreadCount,
			ts.totalThreadCount,
			ts.activeThreadRatioPercent, ts.blockedThreadRatioPercent, ts.waitingThreadRatioPercent,
			ts.healthyFlag, ts.healthGrade, ts.collectionTime
		FROM HUB_MONITOR_JVM_THR_STATE ts
		WHERE %s AND ts.activeFlag = 'Y'
		ORDER BY ts.collectionTime DESC
	`, whereClause)

	// 使用统一的分页构建工具，支持多数据库
	pagination := sqlutils.NewPaginationInfo(1, limit)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建分页查询失败", err)
		return nil, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	var list []models.ThreadStateResponse
	err = dao.db.Query(ctx, &list, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询线程状态记录失败", err)
		return nil, huberrors.WrapError(err, "查询线程状态记录失败")
	}

	// 反转结果顺序，使前端按时间顺序展示（从旧到新）
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return list, nil
}

// ===============================
// 死锁查询
// ===============================

// QueryDeadlocks 查询死锁记录列表
func (dao *JvmQueryDao) QueryDeadlocks(ctx context.Context, req *models.DeadlockQueryRequest) ([]models.DeadlockResponse, error) {
	whereClauses := []string{"d.tenantId = ?"}
	params := []interface{}{req.TenantId}

	if req.JvmResourceId != "" {
		whereClauses = append(whereClauses, "d.jvmResourceId = ?")
		params = append(params, req.JvmResourceId)
	}

	if req.HasDeadlockFlag != "" {
		whereClauses = append(whereClauses, "d.hasDeadlockFlag = ?")
		params = append(params, req.HasDeadlockFlag)
	}

	if req.SeverityLevel != "" {
		whereClauses = append(whereClauses, "d.severityLevel = ?")
		params = append(params, req.SeverityLevel)
	}

	// 添加时间范围条件
	if err := dao.addTimeCondition(ctx, &whereClauses, &params, req.StartTime, req.EndTime, "d.collectionTime"); err != nil {
		return nil, err
	}

	whereClause := strings.Join(whereClauses, " AND ")

	limit := req.Limit
	if limit <= 0 {
		limit = 200 // 默认返回最近200条
	}

	// 构建基础查询语句 - 使用DESC倒序获取最新的N条数据
	baseQuery := fmt.Sprintf(`
		SELECT 
			d.deadlockId, d.jvmResourceId,
			d.hasDeadlockFlag, d.deadlockThreadCount, d.deadlockThreadIds, d.deadlockThreadNames,
			d.severityLevel, d.severityDescription, d.affectedThreadGroups,
			d.detectionTime, d.deadlockDurationMs, d.collectionTime,
			d.descriptionText, d.recommendedAction, d.alertLevel, d.requiresActionFlag
		FROM HUB_MONITOR_JVM_DEADLOCK d
		WHERE %s AND d.activeFlag = 'Y'
		ORDER BY d.collectionTime DESC
	`, whereClause)

	// 使用统一的分页构建工具，支持多数据库
	pagination := sqlutils.NewPaginationInfo(1, limit)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建分页查询失败", err)
		return nil, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	var list []models.DeadlockResponse
	err = dao.db.Query(ctx, &list, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询死锁记录失败", err)
		return nil, huberrors.WrapError(err, "查询死锁记录失败")
	}

	// 反转结果顺序，使前端按时间顺序展示（从旧到新）
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return list, nil
}

// ===============================
// 类加载监控查询
// ===============================

// QueryClassLoading 查询类加载记录列表
func (dao *JvmQueryDao) QueryClassLoading(ctx context.Context, req *models.ClassLoadingQueryRequest) ([]models.ClassLoadingResponse, error) {
	whereClauses := []string{"c.tenantId = ?", "c.jvmResourceId = ?"}
	params := []interface{}{req.TenantId, req.JvmResourceId}

	if req.HealthyFlag != "" {
		whereClauses = append(whereClauses, "c.healthyFlag = ?")
		params = append(params, req.HealthyFlag)
	}

	// 添加时间范围条件
	if err := dao.addTimeCondition(ctx, &whereClauses, &params, req.StartTime, req.EndTime, "c.collectionTime"); err != nil {
		return nil, err
	}

	whereClause := strings.Join(whereClauses, " AND ")

	limit := req.Limit
	if limit <= 0 {
		limit = 200 // 默认返回最近200条
	}

	// 构建基础查询语句 - 使用DESC倒序获取最新的N条数据
	baseQuery := fmt.Sprintf(`
		SELECT 
			c.classLoadingId, c.jvmResourceId,
			c.loadedClassCount, c.totalLoadedClassCount, c.unloadedClassCount,
			c.classUnloadRatePercent, c.classRetentionRatePercent,
			c.healthyFlag, c.healthGrade, c.requiresAttentionFlag, c.collectionTime
		FROM HUB_MONITOR_JVM_CLASS c
		WHERE %s AND c.activeFlag = 'Y'
		ORDER BY c.collectionTime DESC
	`, whereClause)

	// 使用统一的分页构建工具，支持多数据库
	pagination := sqlutils.NewPaginationInfo(1, limit)
	dbType := sqlutils.GetDatabaseType(dao.db)
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建分页查询失败", err)
		return nil, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	var list []models.ClassLoadingResponse
	err = dao.db.Query(ctx, &list, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询类加载记录失败", err)
		return nil, huberrors.WrapError(err, "查询类加载记录失败")
	}

	// 反转结果顺序，使前端按时间顺序展示（从旧到新）
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return list, nil
}

// ===============================
// 统计和概览
// ===============================

// GetJvmOverview 获取JVM监控概览
func (dao *JvmQueryDao) GetJvmOverview(ctx context.Context, tenantId, applicationName string) (*models.JvmOverviewResponse, error) {
	whereClauses := []string{"r.tenantId = ?"}
	params := []interface{}{tenantId}

	if applicationName != "" {
		whereClauses = append(whereClauses, "r.applicationName = ?")
		params = append(params, applicationName)
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// 统计SQL
	sql := fmt.Sprintf(`
		SELECT 
			COUNT(*) as totalInstances,
			SUM(CASE WHEN r.healthyFlag = 'Y' THEN 1 ELSE 0 END) as healthyInstances,
			SUM(CASE WHEN r.healthyFlag = 'N' THEN 1 ELSE 0 END) as unhealthyInstances,
			SUM(CASE WHEN r.requiresAttentionFlag = 'Y' THEN 1 ELSE 0 END) as attentionRequired,
			COUNT(DISTINCT r.applicationName) as applicationCount
		FROM (
			SELECT r.*, ROW_NUMBER() OVER (PARTITION BY r.jvmResourceId ORDER BY r.collectionTime DESC) as rn
			FROM HUB_MONITOR_JVM_RESOURCE r
			WHERE %s AND r.activeFlag = 'Y'
		) r
		WHERE r.rn = 1
	`, whereClause)

	var result models.JvmOverviewResponse
	err := dao.db.QueryOne(ctx, &result, sql, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询JVM概览统计失败", err)
		return nil, huberrors.WrapError(err, "查询JVM概览统计失败")
	}

	return &result, nil
}

// ===============================
// 应用监控数据查询
// ===============================

// QueryAppMonitorData 查询应用监控数据列表
func (dao *JvmQueryDao) QueryAppMonitorData(ctx context.Context, req *models.QueryAppMonitorDataRequest) ([]models.AppMonitorDataResponse, int, error) {
	// 构建WHERE条件
	whereClauses := []string{"d.tenantId = ?"}
	params := []interface{}{req.TenantId}

	if req.JvmResourceId != "" {
		whereClauses = append(whereClauses, "d.jvmResourceId = ?")
		params = append(params, req.JvmResourceId)
	}

	if req.ApplicationName != "" {
		whereClauses = append(whereClauses, "r.applicationName LIKE ?")
		params = append(params, "%"+req.ApplicationName+"%")
	}

	if req.DataType != "" {
		whereClauses = append(whereClauses, "d.dataType = ?")
		params = append(params, req.DataType)
	}

	if req.DataName != "" {
		whereClauses = append(whereClauses, "d.dataName = ?")
		params = append(params, req.DataName)
	}

	if req.DataCategory != "" {
		whereClauses = append(whereClauses, "d.dataCategory = ?")
		params = append(params, req.DataCategory)
	}

	if req.HealthyFlag != "" {
		whereClauses = append(whereClauses, "d.healthyFlag = ?")
		params = append(params, req.HealthyFlag)
	}

	if req.HealthGrade != "" {
		whereClauses = append(whereClauses, "d.healthGrade = ?")
		params = append(params, req.HealthGrade)
	}

	if req.RequiresAttentionFlag != "" {
		whereClauses = append(whereClauses, "d.requiresAttentionFlag = ?")
		params = append(params, req.RequiresAttentionFlag)
	}

	// 添加时间范围条件
	if err := dao.addTimeCondition(ctx, &whereClauses, &params, req.StartTime, req.EndTime, "d.collectionTime"); err != nil {
		return nil, 0, err
	}

	whereClause := "WHERE " + strings.Join(whereClauses, " AND ") + " AND d.activeFlag = 'Y'"

	// 构建排序
	orderBy := "d.collectionTime"
	if req.OrderBy != "" {
		orderBy = "d." + req.OrderBy
	}
	orderDirection := "DESC"
	if req.OrderDirection != "" {
		orderDirection = req.OrderDirection
	}

	// 构建基础查询语句（关联JVM资源表获取应用信息）
	baseQuery := fmt.Sprintf(`
		SELECT 
			d.appDataId, d.tenantId, d.jvmResourceId, d.dataType, d.dataName, d.dataCategory,
			d.dataJson, d.primaryValue, d.secondaryValue, d.statusValue,
			d.healthyFlag, d.healthGrade, d.requiresAttentionFlag, d.tagsJson,
			d.collectionTime, d.addTime, d.editTime, d.oprSeqFlag, d.currentVersion, d.activeFlag, d.noteText
		FROM HUB_MONITOR_APP_DATA d
		LEFT JOIN HUB_MONITOR_JVM_RESOURCE r ON d.jvmResourceId = r.jvmResourceId AND d.tenantId = r.tenantId
		%s
		ORDER BY %s %s
	`, whereClause, orderBy, orderDirection)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建统计查询失败", err)
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var countResult struct {
		Count int64 `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询应用监控数据总数失败", err)
		return nil, 0, huberrors.WrapError(err, "查询应用监控数据总数失败")
	}

	// 如果没有记录，直接返回空列表
	if countResult.Count == 0 {
		return []models.AppMonitorDataResponse{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)

	// 构建分页查询
	dbType := sqlutils.GetDatabaseType(dao.db)
	pagedQuery, pageArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		logger.ErrorWithTrace(ctx, "构建分页查询失败", err)
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并参数
	allParams := append(params, pageArgs...)

	// 执行分页查询
	var list []models.AppMonitorDataResponse
	err = dao.db.Query(ctx, &list, pagedQuery, allParams, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询应用监控数据列表失败", err)
		return nil, 0, huberrors.WrapError(err, "查询应用监控数据列表失败")
	}

	return list, int(countResult.Count), nil
}

// GetAppMonitorDataDetail 获取应用监控数据详情
func (dao *JvmQueryDao) GetAppMonitorDataDetail(ctx context.Context, tenantId, appDataId string) (*models.AppMonitorDataResponse, error) {
	sql := `
		SELECT 
			d.appDataId, d.tenantId, d.jvmResourceId, d.dataType, d.dataName, d.dataCategory,
			d.dataJson, d.primaryValue, d.secondaryValue, d.statusValue,
			d.healthyFlag, d.healthGrade, d.requiresAttentionFlag, d.tagsJson,
			d.collectionTime, d.addTime, d.editTime, d.oprSeqFlag, d.currentVersion, d.activeFlag, d.noteText
		FROM HUB_MONITOR_APP_DATA d
		WHERE d.tenantId = ? AND d.appDataId = ? AND d.activeFlag = 'Y'
	`

	var result models.AppMonitorDataResponse
	err := dao.db.QueryOne(ctx, &result, sql, []interface{}{tenantId, appDataId}, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询应用监控数据详情失败", err)
		return nil, huberrors.WrapError(err, "查询应用监控数据详情失败")
	}

	return &result, nil
}
