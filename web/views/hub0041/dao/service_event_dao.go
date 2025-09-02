package dao

import (
	"context"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/ctime"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0041/models"
	"strings"
)

// ServiceEventDAO 服务事件数据访问对象
// 负责服务事件日志信息的数据库操作（查询）
type ServiceEventDAO struct {
	db database.Database
}

// NewServiceEventDAO 创建服务事件DAO实例
func NewServiceEventDAO(db database.Database) *ServiceEventDAO {
	return &ServiceEventDAO{
		db: db,
	}
}

// QueryServiceEvents 分页查询服务事件列表
// 性能优化：列表查询不返回大字段（eventDataJson, extProperty, reserved字段）
// 这些字段可能包含大量数据，在列表展示时不需要，只在详情查询时获取
// 这样可以显著提高查询性能，减少网络传输量和内存使用
func (dao *ServiceEventDAO) QueryServiceEvents(ctx context.Context, req *models.ServiceEventQueryRequest) ([]models.ServiceEventSummary, int, error) {
	// 构建查询条件
	var whereConditions []string
	var params []interface{}

	// 租户ID条件
	if req.TenantId != "" {
		whereConditions = append(whereConditions, "tenantId = ?")
		params = append(params, req.TenantId)
	}

	// 活动状态条件
	if req.ActiveFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		params = append(params, req.ActiveFlag)
	} else {
		// 默认只查询活动的记录
		whereConditions = append(whereConditions, "activeFlag = 'Y'")
	}

	// 事件类型过滤
	if req.EventType != "" {
		whereConditions = append(whereConditions, "eventType = ?")
		params = append(params, req.EventType)
	}

	// 服务名称过滤（模糊查询）
	if req.ServiceName != "" {
		whereConditions = append(whereConditions, "serviceName LIKE ?")
		params = append(params, "%"+req.ServiceName+"%")
	}

	// 分组名称过滤
	if req.GroupName != "" {
		whereConditions = append(whereConditions, "groupName = ?")
		params = append(params, req.GroupName)
	}

	// 主机地址过滤（模糊查询）
	if req.HostAddress != "" {
		whereConditions = append(whereConditions, "hostAddress LIKE ?")
		params = append(params, "%"+req.HostAddress+"%")
	}

	// 事件来源过滤
	if req.EventSource != "" {
		whereConditions = append(whereConditions, "eventSource = ?")
		params = append(params, req.EventSource)
	}

	// 时间条件处理 - 使用ctime包正确解析时间字符串
	// Oracle数据库需要将字符串转换为time.Time类型才能正确进行时间比较
	if req.StartTime != "" {
		startTime, err := ctime.ParseTimeString(req.StartTime)
		if err != nil {
			logger.ErrorWithTrace(ctx, "开始时间格式不正确", "startTime", req.StartTime, "error", err)
			return nil, 0, huberrors.WrapError(err, "开始时间格式不正确: %s", req.StartTime)
		}
		whereConditions = append(whereConditions, "eventTime >= ?")
		params = append(params, startTime)
	}

	if req.EndTime != "" {
		endTime, err := ctime.ParseTimeString(req.EndTime)
		if err != nil {
			logger.ErrorWithTrace(ctx, "结束时间格式不正确", "endTime", req.EndTime, "error", err)
			return nil, 0, huberrors.WrapError(err, "结束时间格式不正确: %s", req.EndTime)
		}
		whereConditions = append(whereConditions, "eventTime <= ?")
		params = append(params, endTime)
	}

	// 关键字搜索
	if req.Keyword != "" {
		whereConditions = append(whereConditions, "(serviceName LIKE ? OR eventMessage LIKE ? OR groupName LIKE ? OR hostAddress LIKE ?)")
		keyword := "%" + req.Keyword + "%"
		params = append(params, keyword, keyword, keyword, keyword)
	}

	// 构建WHERE子句
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 构建基础查询语句 - 列表查询不返回大字段
	baseQuery := fmt.Sprintf(`
		SELECT serviceEventId, tenantId, serviceGroupId, serviceInstanceId,
			   groupName, serviceName, hostAddress, portNumber, nodeIpAddress,
			   eventType, eventSource, eventMessage, eventTime,
			   addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText
		FROM HUB_REGISTRY_SERVICE_EVENT %s
		ORDER BY eventTime DESC
	`, whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务事件总数失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "查询服务事件总数失败")
	}

	// 如果没有记录，直接返回空列表
	if countResult.Count == 0 {
		return []models.ServiceEventSummary{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	// 执行分页查询
	var events []models.ServiceEventSummary
	err = dao.db.Query(ctx, &events, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务事件数据失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "查询服务事件数据失败")
	}

	return events, countResult.Count, nil
}

// GetServiceEvent 根据事件ID获取服务事件详情
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//   - serviceEventId: 服务事件ID
//   - activeFlag: 活动状态标记(为空则不过滤)
//
// 返回值：
//   - *models.ServiceEvent: 服务事件详情
//   - error: 错误信息
func (dao *ServiceEventDAO) GetServiceEvent(ctx context.Context, tenantId, serviceEventId, activeFlag string) (*models.ServiceEvent, error) {
	whereConditions := []string{"tenantId = ?", "serviceEventId = ?"}
	args := []interface{}{tenantId, serviceEventId}

	if activeFlag != "" {
		whereConditions = append(whereConditions, "activeFlag = ?")
		args = append(args, activeFlag)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := `SELECT serviceEventId, tenantId, serviceGroupId, serviceInstanceId,
		groupName, serviceName, hostAddress, portNumber, nodeIpAddress,
		eventType, eventSource, eventDataJson, eventMessage, eventTime,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
		reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
	FROM HUB_REGISTRY_SERVICE_EVENT WHERE ` + whereClause

	event := &models.ServiceEvent{}
	err := dao.db.QueryOne(ctx, event, query, args, true)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "not found") {
			return nil, huberrors.WrapError(err, "服务事件不存在")
		}
		return nil, huberrors.WrapError(err, "获取服务事件信息失败")
	}

	return event, nil
}

// GetEventTypes 获取事件类型列表
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//
// 返回值：
//   - []string: 事件类型列表
//   - error: 错误信息
func (dao *ServiceEventDAO) GetEventTypes(ctx context.Context, tenantId string) ([]string, error) {
	query := `SELECT DISTINCT eventType FROM HUB_REGISTRY_SERVICE_EVENT 
		WHERE tenantId = ? AND activeFlag = 'Y' 
		ORDER BY eventType`

	var eventTypes []string
	err := dao.db.Query(ctx, &eventTypes, query, []interface{}{tenantId}, false)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询事件类型列表失败")
	}

	return eventTypes, nil
}

// GetEventSources 获取事件来源列表
//
// 参数：
//   - ctx: 上下文对象
//   - tenantId: 租户ID
//
// 返回值：
//   - []string: 事件来源列表
//   - error: 错误信息
func (dao *ServiceEventDAO) GetEventSources(ctx context.Context, tenantId string) ([]string, error) {
	query := `SELECT DISTINCT eventSource FROM HUB_REGISTRY_SERVICE_EVENT 
		WHERE tenantId = ? AND activeFlag = 'Y' 
		ORDER BY eventSource`

	var eventSources []string
	err := dao.db.Query(ctx, &eventSources, query, []interface{}{tenantId}, false)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询事件来源列表失败")
	}

	return eventSources, nil
}
