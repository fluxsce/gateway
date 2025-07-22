package dao

import (
	"context"
	"math"
	"sort"
	"strconv"
	"time"

	"gohub/pkg/logger"
	"gohub/pkg/mongo/client"
	"gohub/pkg/mongo/types"
	"gohub/pkg/mongo/utils"
	"gohub/pkg/utils/ctime"
	"gohub/pkg/utils/huberrors"
	"gohub/web/views/hub0023/models"
)

// MongoMonitoringDAO MongoDB监控数据访问对象
// 专门用于从 HUB_GW_ACCESS_LOG 表中抽取各种监控统计数据
//
// 重要提示：
// 1. 为了优化查询性能，建议在MongoDB中创建以下复合索引：
//    - {gatewayStartProcessingTime: 1, tenantId: 1, gatewayInstanceId: 1}
//    - {requestPath: 1, gatewayStartProcessingTime: 1}
//    - {serviceName: 1, gatewayStartProcessingTime: 1}
//    - {gatewayStatusCode: 1, gatewayStartProcessingTime: 1}
//
// 2. 所有聚合查询都使用投影(projection)来限制返回字段，避免传输不必要的大字段
// 3. 查询时间范围已在控制器层限制为24小时内，防止大数据量查询
// 4. 聚合查询使用了合理的超时时间和错误处理
// 5. 使用mongo/utils包进行结果集转换，确保类型安全和性能优化
// 6. 直接使用models中定义的结构体，避免重复定义
type MongoMonitoringDAO struct {
	mongoClient *client.Client
}

// NewMongoMonitoringDAO 创建MongoDB监控数据DAO
func NewMongoMonitoringDAO(mongoClient *client.Client) *MongoMonitoringDAO {
	return &MongoMonitoringDAO{
		mongoClient: mongoClient,
	}
}

// GetGatewayMonitoringOverview 获取网关监控概览数据
// 基于查询条件统计总体监控指标
//
// 查询优化说明：
// - 使用MongoDB聚合管道进行数据统计，避免在应用层处理大量数据
// - 只聚合必要的统计字段，不返回原始文档内容
// - 通过$match阶段预先过滤数据，减少后续处理的数据量
// - 使用条件表达式($cond)进行分类统计，一次查询完成多个指标计算
// - 直接使用models.GatewayMonitoringOverview结构体进行结果转换
func (dao *MongoMonitoringDAO) GetGatewayMonitoringOverview(ctx context.Context, req *models.GatewayMonitoringQueryRequest) (*models.GatewayMonitoringOverview, error) {
	// 构建查询条件
	filter, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 获取网关日志集合
	collection := dao.getGatewayLogCollection()
	
	// 构建聚合管道
	// 注意：这里只进行统计聚合，不返回具体的文档内容，避免大数据传输
	pipeline := []types.Document{
		// 第一阶段：匹配查询条件，充分利用索引
		{
			"$match": filter,
		},
		// 第二阶段：分组统计，一次性计算所有概览指标
		// 使用条件表达式避免多次查询数据库
		{
			"$group": types.Document{
				"_id": nil, // 全局统计，不分组
				// 总请求数：直接计数
				"totalRequests": types.Document{"$sum": 1},
				// 成功请求数：状态码在200-299范围内
				"successRequests": types.Document{
					"$sum": types.Document{
						"$cond": types.Document{
							"if": types.Document{
								"$and": []interface{}{
									types.Document{"$gte": []interface{}{"$gatewayStatusCode", 200}},
									types.Document{"$lt": []interface{}{"$gatewayStatusCode", 300}},
								},
							},
							"then": 1,
							"else": 0,
						},
					},
				},
				// 失败请求数：状态码小于200或大于等于400
				"failedRequests": types.Document{
					"$sum": types.Document{
						"$cond": types.Document{
							"if": types.Document{
								"$or": []interface{}{
									types.Document{"$gte": []interface{}{"$gatewayStatusCode", 400}},
									types.Document{"$lt": []interface{}{"$gatewayStatusCode", 200}},
								},
							},
							"then": 1,
							"else": 0,
						},
					},
				},
				// 平均响应时间：只计算非空值
				"avgResponseTime": types.Document{
					"$avg": types.Document{
						"$cond": types.Document{
							"if": types.Document{"$ne": []interface{}{"$totalProcessingTimeMs", nil}},
							"then": "$totalProcessingTimeMs",
							"else": "$$REMOVE", // 排除空值，不参与平均值计算
						},
					},
				},
				// 最小响应时间：只考虑非空值
				"minResponseTime": types.Document{
					"$min": types.Document{
						"$cond": types.Document{
							"if": types.Document{"$ne": []interface{}{"$totalProcessingTimeMs", nil}},
							"then": "$totalProcessingTimeMs",
							"else": "$$REMOVE",
						},
					},
				},
				// 最大响应时间：只考虑非空值
				"maxResponseTime": types.Document{
					"$max": types.Document{
						"$cond": types.Document{
							"if": types.Document{"$ne": []interface{}{"$totalProcessingTimeMs", nil}},
							"then": "$totalProcessingTimeMs",
							"else": "$$REMOVE",
						},
					},
				},
			},
		},
	}

	// 执行聚合查询，直接使用传入的context
	cursor, err := collection.Aggregate(ctx, pipeline, nil)
	if err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB监控概览数据聚合查询失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB监控概览数据聚合查询失败")
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []types.Document
	if err := cursor.All(ctx, &results); err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB监控概览数据聚合结果解析失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB监控概览数据聚合结果解析失败")
	}

	// 处理结果，直接使用models.GatewayMonitoringOverview结构体
	overview := &models.GatewayMonitoringOverview{}
	if len(results) > 0 {
		// 使用mongo/utils转换聚合结果
		if err := utils.ConvertDocument(results[0], overview); err != nil {
			logger.ErrorWithTrace(ctx, "聚合结果转换失败", "error", err)
			return nil, huberrors.WrapError(err, "聚合结果转换失败")
		}
		
		// 平均响应时间最多保留两位小数
		overview.AvgResponseTimeMs = roundToTwoDecimalPlaces(overview.AvgResponseTimeMs)
	}

	return overview, nil
}

// GetRequestMetricsTrend 获取请求指标趋势数据
// 按时间粒度分组统计请求量数据
//
// 查询优化说明：
// - 使用$dateToString进行时间分组，避免在应用层处理时间聚合
// - 时间分组格式根据粒度动态调整，减少不必要的精度
// - 只返回趋势统计数据，不包含原始日志内容
// - 按时间排序确保前端可以直接使用数据绘制趋势图
// - 直接使用models.RequestMetrics结构体进行批量结果转换
func (dao *MongoMonitoringDAO) GetRequestMetricsTrend(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.RequestMetrics, error) {
	// 构建查询条件
	filter, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 获取时间分组格式，根据粒度优化分组策略
	timeGroupFormat := dao.getTimeGroupFormat(req.TimeGranularity)
	
	// 获取网关日志集合
	collection := dao.getGatewayLogCollection()
	
	// 构建聚合管道
	// 注意：使用时间分组减少数据点数量，避免返回过多细粒度数据
	pipeline := []types.Document{
		// 第一阶段：匹配查询条件
		{
			"$match": filter,
		},
		// 第二阶段：按时间分组统计
		// 使用$dateToString将时间标准化到指定粒度
		{
			"$group": types.Document{
				"_id": types.Document{
					"$dateToString": types.Document{
						"format": timeGroupFormat, // 时间格式根据粒度决定
						"date":   "$gatewayStartProcessingTime",
					},
				},
				// 每个时间段的总请求数
				"totalRequests": types.Document{"$sum": 1},
				// 每个时间段的成功请求数
				"successRequests": types.Document{
					"$sum": types.Document{
						"$cond": types.Document{
							"if": types.Document{
								"$and": []interface{}{
									types.Document{"$gte": []interface{}{"$gatewayStatusCode", 200}},
									types.Document{"$lt": []interface{}{"$gatewayStatusCode", 300}},
								},
							},
							"then": 1,
							"else": 0,
						},
					},
				},
				// 每个时间段的失败请求数
				"failedRequests": types.Document{
					"$sum": types.Document{
						"$cond": types.Document{
							"if": types.Document{
								"$or": []interface{}{
									types.Document{"$gte": []interface{}{"$gatewayStatusCode", 400}},
									types.Document{"$lt": []interface{}{"$gatewayStatusCode", 200}},
								},
							},
							"then": 1,
							"else": 0,
						},
					},
				},
				// 保留原始时间戳用于前端展示
				"timestamp": types.Document{"$first": "$gatewayStartProcessingTime"},
			},
		},
		// 第三阶段：按时间排序，确保趋势图的时间顺序
		{
			"$sort": types.Document{
				"_id": 1,
			},
		},
	}

	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline, nil)
	if err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB请求指标趋势数据聚合查询失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB请求指标趋势数据聚合查询失败")
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []types.Document
	if err := cursor.All(ctx, &results); err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB请求指标趋势数据聚合结果解析失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB请求指标趋势数据聚合结果解析失败")
	}

	// 检查结果是否为空，如果为空直接返回空切片
	if len(results) == 0 {
		return []models.RequestMetrics{}, nil
	}

	// 使用mongo/utils批量转换结果，直接转换为models.RequestMetrics
	var resultItems []models.RequestMetrics
	if err := utils.ConvertDocuments(results, &resultItems); err != nil {
		logger.ErrorWithTrace(ctx, "请求指标趋势数据转换失败", "error", err)
		return nil, huberrors.WrapError(err, "请求指标趋势数据转换失败")
	}

	// 处理转换结果，计算衍生字段
	timeGranularitySeconds := dao.getTimeGranularitySeconds(req.TimeGranularity)
	
	for i := range resultItems {
		// 将SourceTimestamp转换为毫秒时间戳
		if !resultItems[i].SourceTimestamp.IsZero() {
			resultItems[i].Timestamp = resultItems[i].SourceTimestamp.UnixMilli()
		}
		
		// 计算QPS：考虑时间粒度，提供准确的每秒请求数
		if resultItems[i].TotalRequests > 0 && timeGranularitySeconds > 0 {
			resultItems[i].RequestsPerSecond = float64(resultItems[i].TotalRequests) / float64(timeGranularitySeconds)
		}
	}

	return resultItems, nil
}

// GetResponseTimeMetricsTrend 获取响应时间指标趋势数据
// 按时间粒度分组统计响应时间数据
//
// 查询优化说明：
// - 预先过滤响应时间为空的记录，减少无效数据处理
// - 使用$push收集响应时间值用于百分位数计算，但限制在内存可处理范围内
// - 百分位数计算在应用层进行，避免复杂的MongoDB聚合操作
// - 时间分组减少数据点数量，适合前端图表展示
// - 直接使用models.ResponseTimeMetrics结构体进行批量结果转换
func (dao *MongoMonitoringDAO) GetResponseTimeMetricsTrend(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.ResponseTimeMetrics, error) {
	// 构建查询条件
	filter, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 添加响应时间非空条件，避免处理无效数据
	filter["totalProcessingTimeMs"] = types.Document{
		"$exists": true,
		"$ne":     nil,
		"$gt":     0, // 响应时间应该大于0
	}

	// 获取时间分组格式
	timeGroupFormat := dao.getTimeGroupFormat(req.TimeGranularity)
	
	// 获取网关日志集合
	collection := dao.getGatewayLogCollection()
	
	// 构建聚合管道
	pipeline := []types.Document{
		// 第一阶段：匹配查询条件，包括响应时间有效性检查
		{
			"$match": filter,
		},
		// 第二阶段：按时间分组统计响应时间指标
		{
			"$group": types.Document{
				"_id": types.Document{
					"$dateToString": types.Document{
						"format": timeGroupFormat,
						"date":   "$gatewayStartProcessingTime",
					},
				},
				// 平均响应时间
				"avgResponseTime": types.Document{"$avg": "$totalProcessingTimeMs"},
				// 最小响应时间
				"minResponseTime": types.Document{"$min": "$totalProcessingTimeMs"},
				// 最大响应时间
				"maxResponseTime": types.Document{"$max": "$totalProcessingTimeMs"},
				// 收集响应时间值用于百分位数计算
				// 注意：限制收集的数据量，避免内存溢出
				"responseTimeValues": types.Document{
					"$push": types.Document{
						"$cond": types.Document{
							"if": types.Document{"$lte": []interface{}{"$$ROOT.totalProcessingTimeMs", 60000}}, // 限制最大值60秒
							"then": "$totalProcessingTimeMs",
							"else": "$$REMOVE",
						},
					},
				},
				// 该时间段的请求数量
				"requestCount": types.Document{"$sum": 1},
				// 保留时间戳
				"timestamp": types.Document{"$first": "$gatewayStartProcessingTime"},
			},
		},
		// 第三阶段：按时间排序
		{
			"$sort": types.Document{
				"_id": 1,
			},
		},
	}

	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline, nil)
	if err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB响应时间指标趋势数据聚合查询失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB响应时间指标趋势数据聚合查询失败")
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []types.Document
	if err := cursor.All(ctx, &results); err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB响应时间指标趋势数据聚合结果解析失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB响应时间指标趋势数据聚合结果解析失败")
	}

	// 检查结果是否为空，如果为空直接返回空切片
	if len(results) == 0 {
		return []models.ResponseTimeMetrics{}, nil
	}

	// 使用mongo/utils批量转换结果，直接转换为models.ResponseTimeMetrics
	var resultItems []models.ResponseTimeMetrics
	if err := utils.ConvertDocuments(results, &resultItems); err != nil {
		logger.ErrorWithTrace(ctx, "响应时间指标趋势数据转换失败", "error", err)
		return nil, huberrors.WrapError(err, "响应时间指标趋势数据转换失败")
	}

	// 处理转换结果，计算衍生字段
	for i := range resultItems {
		// 将SourceTimestamp转换为毫秒时间戳
		if !resultItems[i].SourceTimestamp.IsZero() {
			resultItems[i].Timestamp = resultItems[i].SourceTimestamp.UnixMilli()
		}
		
		// 平均响应时间最多保留两位小数
		resultItems[i].AvgResponseTimeMs = roundToTwoDecimalPlaces(resultItems[i].AvgResponseTimeMs)
		
		// 计算百分位数：在应用层处理，避免复杂的MongoDB聚合
		if len(resultItems[i].ResponseTimeValues) > 0 {
			var values []int
			// 限制处理的数据量，避免内存问题
			maxValues := 10000 // 最多处理10000个值
			for j, v := range resultItems[i].ResponseTimeValues {
				if j >= maxValues {
					logger.WarnWithTrace(ctx, "响应时间值数量过多，已限制处理数量", "maxValues", maxValues, "totalValues", len(resultItems[i].ResponseTimeValues))
					break
				}
				// 现在 v 已经是 int 类型，直接使用
				values = append(values, v)
			}
			
			// 计算百分位数
			if len(values) > 0 {
				sort.Ints(values)
				resultItems[i].P50ResponseTimeMs = calculatePercentile(values, 0.5)
				resultItems[i].P90ResponseTimeMs = calculatePercentile(values, 0.9)
				resultItems[i].P99ResponseTimeMs = calculatePercentile(values, 0.99)
			}
		}
	}

	return resultItems, nil
}

// GetStatusCodeDistribution 获取状态码分布数据
// 统计各状态码的数量和百分比
//
// 查询优化说明：
// - 简单的分组聚合，性能较好
// - 只统计状态码和数量，不返回其他字段
// - 按状态码排序便于前端展示
// - 百分比计算在应用层进行，避免复杂的聚合操作
// - 直接使用models.GatewayMonitoringStatusCodeData结构体进行批量结果转换
func (dao *MongoMonitoringDAO) GetStatusCodeDistribution(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.GatewayMonitoringStatusCodeData, error) {
	// 构建查询条件
	filter, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 获取网关日志集合
	collection := dao.getGatewayLogCollection()
	
	// 构建聚合管道
	// 这是一个简单的分组统计，性能良好
	pipeline := []types.Document{
		// 第一阶段：匹配查询条件
		{
			"$match": filter,
		},
		// 第二阶段：按状态码分组统计
		{
			"$group": types.Document{
				"_id":   "$gatewayStatusCode", // 按状态码分组
				"count": types.Document{"$sum": 1}, // 计算每个状态码的出现次数
			},
		},
		// 第三阶段：按状态码排序，便于前端展示
		{
			"$sort": types.Document{
				"_id": 1,
			},
		},
	}

	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline, nil)
	if err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB状态码分布数据聚合查询失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB状态码分布数据聚合查询失败")
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []types.Document
	if err := cursor.All(ctx, &results); err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB状态码分布数据聚合结果解析失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB状态码分布数据聚合结果解析失败")
	}

	// 检查结果是否为空，如果为空直接返回空切片
	if len(results) == 0 {
		return []models.GatewayMonitoringStatusCodeData{}, nil
	}

	// 使用mongo/utils批量转换结果，直接转换为models.GatewayMonitoringStatusCodeData
	var resultItems []models.GatewayMonitoringStatusCodeData
	if err := utils.ConvertDocuments(results, &resultItems); err != nil {
		logger.ErrorWithTrace(ctx, "状态码分布数据转换失败", "error", err)
		return nil, huberrors.WrapError(err, "状态码分布数据转换失败")
	}

	// 先计算总数，用于百分比计算
	var totalCount int64
	for _, resultItem := range resultItems {
		totalCount += resultItem.Count
	}

	// 处理转换结果，计算衍生字段
	for i := range resultItems {
		// 将StatusCodeValue转换为字符串
		resultItems[i].StatusCode = strconv.FormatInt(resultItems[i].StatusCodeValue, 10)
		
		// 计算百分比
		if totalCount > 0 {
			resultItems[i].Percentage = float64(resultItems[i].Count) / float64(totalCount) * 100
		}
		
		// 设置状态码分类和描述
		resultItems[i].Category = dao.getStatusCodeCategory(resultItems[i].StatusCode)
		resultItems[i].Description = dao.getStatusCodeDescription(resultItems[i].StatusCode)
	}

	return resultItems, nil
}

// GetHotRoutes 获取热点路由数据
// 统计访问量最高的路由
//
// 查询优化说明：
// - 按请求路径分组，减少分组数量
// - 使用$first获取冗余字段，避免重复计算
// - 限制返回数量，防止返回过多数据
// - 按请求数量降序排序，直接获取热点路由
// - 直接使用models.GatewayMonitoringHotRouteData结构体进行批量结果转换
func (dao *MongoMonitoringDAO) GetHotRoutes(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.GatewayMonitoringHotRouteData, error) {
	// 构建查询条件
	filter, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 获取网关日志集合
	collection := dao.getGatewayLogCollection()
	
	// 设置合理的默认限制，防止返回过多数据
	limit := req.HotRouteLimit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 { // 限制最大值，防止大数据查询
		limit = 50
		logger.WarnWithTrace(ctx, "热点路由查询数量被限制", "requestedLimit", req.HotRouteLimit, "actualLimit", limit)
	}
	
	// 构建聚合管道
	pipeline := []types.Document{
		// 第一阶段：匹配查询条件
		{
			"$match": filter,
		},
		// 第二阶段：按路由分组统计
		{
			"$group": types.Document{
				"_id": "$requestPath", // 按请求路径分组
				// 请求数量
				"requestCount": types.Document{"$sum": 1},
				// 最大响应时间（仅统计有效值）
				"maxResponseTime": types.Document{
					"$max": types.Document{
						"$cond": types.Document{
							"if": types.Document{
								"$and": []interface{}{
									types.Document{"$ne": []interface{}{"$totalProcessingTimeMs", nil}},
									types.Document{"$gt": []interface{}{"$totalProcessingTimeMs", 0}},
								},
							},
							"then": "$totalProcessingTimeMs",
							"else": "$$REMOVE",
						},
					},
				},
				// 最小响应时间（仅统计有效值）
				"minResponseTime": types.Document{
					"$min": types.Document{
						"$cond": types.Document{
							"if": types.Document{
								"$and": []interface{}{
									types.Document{"$ne": []interface{}{"$totalProcessingTimeMs", nil}},
									types.Document{"$gt": []interface{}{"$totalProcessingTimeMs", 0}},
								},
							},
							"then": "$totalProcessingTimeMs",
							"else": "$$REMOVE",
						},
					},
				},
				// 错误数量统计
				"errorCount": types.Document{
					"$sum": types.Document{
						"$cond": types.Document{
							"if": types.Document{
								"$or": []interface{}{
									types.Document{"$gte": []interface{}{"$gatewayStatusCode", 400}},
									types.Document{"$lt": []interface{}{"$gatewayStatusCode", 200}},
								},
							},
							"then": 1,
							"else": 0,
						},
					},
				},
				// 使用$first获取冗余字段，避免重复统计
				"routeConfigId": types.Document{"$first": "$routeConfigId"},
				"routeName":     types.Document{"$first": "$routeName"},
				"serviceName":   types.Document{"$first": "$serviceName"},
			},
		},
		// 第三阶段：按请求数量降序排序，获取热点路由
		{
			"$sort": types.Document{
				"requestCount": -1,
			},
		},
		// 第四阶段：限制结果数量，防止返回过多数据
		{
			"$limit": int64(limit),
		},
	}

	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline, nil)
	if err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB热点路由数据聚合查询失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB热点路由数据聚合查询失败")
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []types.Document
	if err := cursor.All(ctx, &results); err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB热点路由数据聚合结果解析失败", "error", err)
		return nil, huberrors.WrapError(err, "MongoDB热点路由数据聚合结果解析失败")
	}

	// 检查结果是否为空，如果为空直接返回空切片
	if len(results) == 0 {
		return []models.GatewayMonitoringHotRouteData{}, nil
	}

	// 使用mongo/utils批量转换结果，直接转换为models.GatewayMonitoringHotRouteData
	var resultItems []models.GatewayMonitoringHotRouteData
	if err := utils.ConvertDocuments(results, &resultItems); err != nil {
		logger.ErrorWithTrace(ctx, "热点路由数据转换失败", "error", err)
		return nil, huberrors.WrapError(err, "热点路由数据转换失败")
	}

	// 计算时间范围秒数，用于QPS计算
	var timeRangeSeconds float64
	startTime, _ := dao.parseTimeString(req.StartTime)
	endTime, _ := dao.parseTimeString(req.EndTime)
	if !startTime.IsZero() && !endTime.IsZero() {
		timeRangeSeconds = endTime.Sub(startTime).Seconds()
	}

	// 处理转换结果，计算衍生字段
	for i := range resultItems {
		// 将RoutePathValue转换为RoutePath
		resultItems[i].RoutePath = resultItems[i].RoutePathValue
		
		// 计算错误率
		if resultItems[i].RequestCount > 0 {
			resultItems[i].ErrorRate = float64(resultItems[i].ErrorCount) / float64(resultItems[i].RequestCount) * 100
		}
		
		// 计算QPS
		if resultItems[i].RequestCount > 0 && timeRangeSeconds > 0 {
			resultItems[i].QPS = float64(resultItems[i].RequestCount) / timeRangeSeconds
		}
	}

	return resultItems, nil
}

// buildMonitoringFilter 构建监控数据查询条件
// 
// 查询优化说明：
// - 优先使用精确匹配的字段，便于索引利用
// - 时间范围查询放在最前面，充分利用时间索引
// - 对于模糊查询字段（如requestPath），使用正则表达式
// - 所有字段都进行非空检查，避免无效查询
func (dao *MongoMonitoringDAO) buildMonitoringFilter(req *models.GatewayMonitoringQueryRequest) (types.Filter, error) {
	filter := types.Filter{}

	// 时间范围查询（必须字段，优先设置以利用时间索引）
	if req.StartTime != "" || req.EndTime != "" {
		timeFilter := map[string]interface{}{}
		
		if req.StartTime != "" {
			startTime, err := dao.parseTimeString(req.StartTime)
			if err != nil {
				return nil, huberrors.WrapError(err, "开始时间格式错误")
			}
			// 应用时区转换，确保与MongoDB存储的时间格式一致
			timeFilter["$gte"] = utils.ConvertGoTimeToMongo(startTime)
		}
		
		if req.EndTime != "" {
			endTime, err := dao.parseTimeString(req.EndTime)
			if err != nil {
				return nil, huberrors.WrapError(err, "结束时间格式错误")
			}
			// 应用时区转换，确保与MongoDB存储的时间格式一致
			timeFilter["$lte"] = utils.ConvertGoTimeToMongo(endTime)
		}
		
		filter["gatewayStartProcessingTime"] = timeFilter
	}

	// 基础查询条件（精确匹配，利于索引）
	if req.TenantId != "" {
		filter["tenantId"] = req.TenantId
	}
	if req.GatewayInstanceId != "" {
		filter["gatewayInstanceId"] = req.GatewayInstanceId
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
	
	// 模糊查询字段（使用正则表达式，性能相对较低，放在最后）
	if req.RequestPath != "" {
		filter["requestPath"] = types.Document{
			"$regex":   req.RequestPath,
			"$options": "i", // 忽略大小写
		}
	}

	return filter, nil
}

// getGatewayLogCollection 获取网关日志集合
func (dao *MongoMonitoringDAO) getGatewayLogCollection() types.MongoCollection {
	// 使用默认数据库和表名
	db, _ := dao.mongoClient.DefaultDatabase()
	return db.Collection(models.GatewayAccessLog{}.TableName())
}

// parseTimeString 解析时间字符串为time.Time
func (dao *MongoMonitoringDAO) parseTimeString(timeStr string) (time.Time, error) {
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

// getTimeGroupFormat 获取时间分组格式
// 根据时间粒度返回相应的MongoDB日期格式字符串
func (dao *MongoMonitoringDAO) getTimeGroupFormat(granularity models.TimeGranularity) string {
	switch granularity {
	case models.TimeGranularityMinute:
		return "%Y-%m-%d %H:%M" // 精确到分钟
	case models.TimeGranularityHour:
		return "%Y-%m-%d %H" // 精确到小时
	case models.TimeGranularityDay:
		return "%Y-%m-%d" // 精确到天
	default:
		return "%Y-%m-%d %H:%M" // 默认按分钟分组
	}
}

// getTimeGranularitySeconds 获取时间粒度对应的秒数
// 用于QPS计算
func (dao *MongoMonitoringDAO) getTimeGranularitySeconds(granularity models.TimeGranularity) int {
	switch granularity {
	case models.TimeGranularityMinute:
		return 60 // 1分钟 = 60秒
	case models.TimeGranularityHour:
		return 3600 // 1小时 = 3600秒
	case models.TimeGranularityDay:
		return 86400 // 1天 = 86400秒
	default:
		return 60 // 默认1分钟
	}
}

// getStatusCodeCategory 获取状态码分类
func (dao *MongoMonitoringDAO) getStatusCodeCategory(statusCode string) string {
	if statusCode == "" {
		return "未知"
	}
	
	switch statusCode[0] {
	case '1':
		return "信息响应"
	case '2':
		return "成功响应"
	case '3':
		return "重定向"
	case '4':
		return "客户端错误"
	case '5':
		return "服务端错误"
	default:
		return "未知"
	}
}

// getStatusCodeDescription 获取状态码描述
func (dao *MongoMonitoringDAO) getStatusCodeDescription(statusCode string) string {
	descriptions := map[string]string{
		"200": "成功",
		"201": "已创建",
		"400": "请求错误",
		"401": "未授权",
		"403": "禁止访问",
		"404": "未找到",
		"405": "方法不允许",
		"500": "服务器内部错误",
		"502": "网关错误",
		"503": "服务不可用",
		"504": "网关超时",
	}
	
	if desc, exists := descriptions[statusCode]; exists {
		return desc
	}
	return "未知状态码"
}

// calculatePercentile 计算百分位数
// 使用线性插值方法计算精确的百分位数值
func calculatePercentile(values []int, percentile float64) int {
	if len(values) == 0 {
		return 0
	}
	
	// 计算索引位置
	index := percentile * float64(len(values)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	// 如果索引相同，直接返回该值
	if lower == upper {
		return values[lower]
	}
	
	// 线性插值计算
	weight := index - float64(lower)
	return int(float64(values[lower])*(1-weight) + float64(values[upper])*weight)
}

// roundToTwoDecimalPlaces 将浮点数最多保留两位小数
// 自动去除不必要的尾随零
//
// 示例：
//   120.00 -> 120
//   120.50 -> 120.5
//   120.56 -> 120.56
//   120.567 -> 120.57 (四舍五入)
func roundToTwoDecimalPlaces(value float64) float64 {
	if value == 0 {
		return 0
	}
	
	// 先四舍五入到两位小数
	rounded := math.Round(value*100) / 100
	
	// 使用 strconv.FormatFloat 去除尾随零，然后转换回 float64
	formatted := strconv.FormatFloat(rounded, 'f', -1, 64)
	result, err := strconv.ParseFloat(formatted, 64)
	if err != nil {
		// 如果转换失败，返回原始的四舍五入结果
		return rounded
	}
	
	return result
} 