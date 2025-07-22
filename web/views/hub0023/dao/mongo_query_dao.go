package dao

import (
	"context"
	"time"

	"gohub/pkg/logger"
	"gohub/pkg/mongo/client"
	"gohub/pkg/mongo/types"
	"gohub/pkg/mongo/utils"
	"gohub/pkg/utils/ctime"
	"gohub/pkg/utils/huberrors"
	"gohub/web/views/hub0023/models"
)

// MongoQueryDAO MongoDB查询数据访问对象
type MongoQueryDAO struct {
	mongoClient *client.Client
}

// NewMongoQueryDAO 创建MongoDB查询DAO
func NewMongoQueryDAO(mongoClient *client.Client) *MongoQueryDAO {
	return &MongoQueryDAO{
		mongoClient: mongoClient,
	}
}

// QueryGatewayLogs 查询网关日志列表（MongoDB版本）
func (dao *MongoQueryDAO) QueryGatewayLogs(ctx context.Context, req *models.GatewayAccessLogQueryRequest) ([]models.GatewayAccessLogSummary, int, error) {
	// 构建查询条件
	filter, err := dao.buildGatewayLogFilter(req)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建查询条件失败")
	}
	
	// 检查查询条件，避免无条件查询返回过多数据
	if dao.isEmptyFilter(filter) && req.PageSize > 100 {
		logger.WarnWithTrace(ctx, "查询条件为空且页面大小超过100，可能返回大量数据", "pageSize", req.PageSize)
	}
	
	// 构建查询选项
	findOptions := &types.FindOptions{}
	
	// 设置分页
	if req.PageSize > 0 {
		pageLimit := int64(req.PageSize)
		findOptions.Limit = &pageLimit
		if req.PageIndex > 1 {
			pageSkip := int64((req.PageIndex - 1) * req.PageSize)
			findOptions.Skip = &pageSkip
		}
	}
	
	// 设置排序（默认按网关开始处理时间倒序，与SQL版本保持一致）
	findOptions.Sort = map[string]interface{}{
		"gatewayStartProcessingTime": -1,
	}

	// 设置投影，只返回摘要所需的字段，减少网络传输和内存占用
	findOptions.Projection = dao.getSummaryProjection()

	// 获取网关日志集合
	collection := dao.getGatewayLogCollection()
	
	// 执行查询，直接使用传入的context
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB网关日志查询失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "MongoDB网关日志查询失败")
	}
	defer cursor.Close(ctx)

	// 使用流式处理代替All()方法，避免大量数据时的内存问题
	var logs []models.GatewayAccessLogSummary
	for cursor.Next(ctx) {
		var document types.Document
		if err := cursor.Decode(&document); err != nil {
			logger.ErrorWithTrace(ctx, "MongoDB游标解码失败", "error", err)
			return nil, 0, huberrors.WrapError(err, "MongoDB游标解码失败")
		}
		
		// 使用转换工具转换单个文档
		var log models.GatewayAccessLogSummary
		if err := utils.ConvertDocument(document, &log); err != nil {
			logger.ErrorWithTrace(ctx, "MongoDB文档转换失败", "error", err)
			return nil, 0, huberrors.WrapError(err, "MongoDB文档转换失败")
		}
		
		logs = append(logs, log)
	}

	// 检查游标错误
	if err := cursor.Err(); err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB游标遍历错误", "error", err)
		return nil, 0, huberrors.WrapError(err, "MongoDB游标遍历错误")
	}

	// 获取总数
	var total int64
	if findOptions.Limit != nil || findOptions.Skip != nil {
		// 使用相同的context执行count操作
		total, err = collection.Count(ctx, filter, nil)
		if err != nil {
			logger.ErrorWithTrace(ctx, "MongoDB统计失败", "error", err)
			total = int64(len(logs))
		}
	} else {
		total = int64(len(logs))
	}

	return logs, int(total), nil
}

// GetGatewayLogByKey 根据主键获取网关日志详情（MongoDB版本）
func (dao *MongoQueryDAO) GetGatewayLogByKey(ctx context.Context, tenantId, traceId string) (*models.GatewayAccessLog, error) {
	// 验证参数
	if tenantId == "" {
		return nil, huberrors.NewError("租户ID不能为空")
	}
	if traceId == "" {
		return nil, huberrors.NewError("链路追踪ID不能为空")
	}

	// 构建查询条件
	filter := types.Filter{
		"tenantId": tenantId,
		"traceId":  traceId,
	}

	// 获取网关日志集合
	collection := dao.getGatewayLogCollection()
	
	// 执行查询，直接使用传入的context
	result := collection.FindOne(ctx, filter, nil)
	if result.Err() != nil {
		logger.ErrorWithTrace(ctx, "MongoDB网关日志获取失败", "tenantId", tenantId, "traceId", traceId, "error", result.Err())
		return nil, huberrors.WrapError(result.Err(), "MongoDB网关日志获取失败")
	}

	// 解析结果
	var document types.Document
	if err := result.Decode(&document); err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB文档解析失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB文档解析失败")
	}

	// 使用转换工具转换为网关日志格式
	var log models.GatewayAccessLog
	if err := utils.ConvertDocument(document, &log); err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB文档转换失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB文档转换失败")
	}

	return &log, nil
}

// CountGatewayLogs 统计网关日志数量（MongoDB版本）
func (dao *MongoQueryDAO) CountGatewayLogs(ctx context.Context, req *models.GatewayAccessLogQueryRequest) (int64, error) {
	// 构建查询条件
	filter, err := dao.buildGatewayLogFilter(req)
	if err != nil {
		return 0, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 获取网关日志集合
	collection := dao.getGatewayLogCollection()
	
	// 执行统计，直接使用传入的context
	count, err := collection.Count(ctx, filter, nil)
	if err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB网关日志统计失败", "error", err)
		return 0, huberrors.WrapError(err, "MongoDB网关日志统计失败")
	}

	return count, nil
}

// getGatewayLogCollection 获取网关日志集合
func (dao *MongoQueryDAO) getGatewayLogCollection() types.MongoCollection {
	// 使用默认数据库和表名
	db, _ := dao.mongoClient.DefaultDatabase()
	return db.Collection(models.GatewayAccessLog{}.TableName())
}

// getSummaryProjection 获取摘要字段投影，减少网络传输和内存占用
func (dao *MongoQueryDAO) getSummaryProjection() types.Document {
	// 排除大字段，只返回摘要所需的字段
	return types.Document{
		"requestHeaders":  0, // 排除请求头
		"requestBody":     0, // 排除请求体
		"responseHeaders": 0, // 排除响应头
		"responseBody":    0, // 排除响应体
		"forwardParams":   0, // 排除转发参数
		"forwardHeaders":  0, // 排除转发头
		"forwardBody":     0, // 排除转发体
		"extProperty":     0, // 排除扩展属性
	}
}

// isEmptyFilter 检查过滤器是否为空
func (dao *MongoQueryDAO) isEmptyFilter(filter types.Filter) bool {
	return len(filter) == 0
}

// parseTimeString 解析时间字符串为time.Time（使用ctime包）
func (dao *MongoQueryDAO) parseTimeString(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}
	
	// 使用ctime包解析时间字符串，支持多种格式
	parsedTime, err := ctime.ParseTimeString(timeStr)
	if err != nil {
		return time.Time{}, huberrors.WrapError(err, "时间格式解析失败")
	}
	
	return parsedTime, nil
}

// buildGatewayLogFilter 构建网关日志查询条件
func (dao *MongoQueryDAO) buildGatewayLogFilter(req *models.GatewayAccessLogQueryRequest) (types.Filter, error) {
	filter := types.Filter{}

	// 使用ctime包处理时间范围查询，并使用convert方法处理时区
	if req.StartTime != "" || req.EndTime != "" {
		timeFilter := map[string]interface{}{}
		
		if req.StartTime != "" {
			startTime, err := dao.parseTimeString(req.StartTime)
			if err != nil {
				return nil, huberrors.WrapError(err, "开始时间格式错误")
			}
			// 使用convert方法处理时区，确保MongoDB查询条件的时间正确
			mongoStartTime := utils.ConvertGoTimeToMongo(startTime)
			timeFilter["$gte"] = mongoStartTime
		}
		
		if req.EndTime != "" {
			endTime, err := dao.parseTimeString(req.EndTime)
			if err != nil {
				return nil, huberrors.WrapError(err, "结束时间格式错误")
			}
			// 使用convert方法处理时区，确保MongoDB查询条件的时间正确
			mongoEndTime := utils.ConvertGoTimeToMongo(endTime)
			timeFilter["$lte"] = mongoEndTime
		}
		
		filter["gatewayStartProcessingTime"] = timeFilter
	}

	// 基础查询条件
	if req.TenantId != "" {
		filter["tenantId"] = req.TenantId
	}
	
	if req.TraceId != "" {
		filter["traceId"] = req.TraceId
	}
	if req.GatewayInstanceId != "" {
		filter["gatewayInstanceId"] = req.GatewayInstanceId
	}
	if req.GatewayInstanceName != "" {
		filter["gatewayInstanceName"] = req.GatewayInstanceName
	}
	if req.RouteConfigId != "" {
		filter["routeConfigId"] = req.RouteConfigId
	}
	if req.RouteName != "" {
		filter["routeName"] = req.RouteName
	}
	if req.ServiceDefinitionId != "" {
		filter["serviceDefinitionId"] = req.ServiceDefinitionId
	}
	if req.ServiceName != "" {
		filter["serviceName"] = req.ServiceName
	}
	if req.ProxyType != "" {
		filter["proxyType"] = req.ProxyType
	}

	// 请求信息查询条件
	if req.RequestMethod != "" {
		filter["requestMethod"] = req.RequestMethod
	}
	if req.RequestPath != "" {
		filter["requestPath"] = map[string]interface{}{
			"$regex": req.RequestPath,
		}
	}
	if req.ClientIpAddress != "" {
		filter["clientIpAddress"] = req.ClientIpAddress
	}
	if req.UserAgent != "" {
		filter["userAgent"] = map[string]interface{}{
			"$regex": req.UserAgent,
		}
	}
	if req.UserIdentifier != "" {
		filter["userIdentifier"] = req.UserIdentifier
	}

	// 响应信息查询条件
	if req.GatewayStatusCode > 0 {
		filter["gatewayStatusCode"] = req.GatewayStatusCode
	}
	if req.BackendStatusCode > 0 {
		filter["backendStatusCode"] = req.BackendStatusCode
	}

	// 错误信息查询条件
	if req.ErrorCode != "" {
		filter["errorCode"] = req.ErrorCode
	}
	if req.ErrorMessage != "" {
		filter["errorMessage"] = map[string]interface{}{
			"$regex": req.ErrorMessage,
		}
	}

	// 性能查询
	if req.MinProcessingTime > 0 || req.MaxProcessingTime > 0 {
		timeFilter := map[string]interface{}{}
		if req.MinProcessingTime > 0 {
			timeFilter["$gte"] = req.MinProcessingTime
		}
		if req.MaxProcessingTime > 0 {
			timeFilter["$lte"] = req.MaxProcessingTime
		}
		filter["totalProcessingTimeMs"] = timeFilter
	}

	// 日志级别和类型
	if req.LogLevel != "" {
		filter["logLevel"] = req.LogLevel
	}
	if req.LogType != "" {
		filter["logType"] = req.LogType
	}

	// 重置标记查询
	if req.ResetFlag != "" {
		filter["resetFlag"] = req.ResetFlag
	}

	// 关键词搜索 - 删除OR条件，因为其他字段已经有like查询

	return filter, nil
} 