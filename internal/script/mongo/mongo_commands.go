package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoIndexCommand MongoDB 索引创建命令
type MongoIndexCommand struct {
	// CollectionName 集合名称
	CollectionName string
	// IndexModel 索引模型
	IndexModel IndexModel
	// Description 索引描述
	Description string
}

// IndexModel 索引模型
type IndexModel struct {
	// Keys 索引键
	Keys bson.D
	// Options 索引选项
	Options *IndexOptions
}

// IndexOptions 索引选项
type IndexOptions struct {
	// Name 索引名称
	Name string
	// Unique 是否唯一索引
	Unique bool
	// Background 是否后台创建
	Background bool
	// Sparse 是否稀疏索引
	Sparse bool
	// ExpireAfterSeconds TTL过期时间（秒）
	ExpireAfterSeconds *int32
}

// ToMongoIndexModel 转换为 MongoDB 驱动的 IndexModel
func (im IndexModel) ToMongoIndexModel() mongo.IndexModel {
	opts := options.Index()

	if im.Options != nil {
		if im.Options.Name != "" {
			opts.SetName(im.Options.Name)
		}
		if im.Options.Unique {
			opts.SetUnique(true)
		}
		if im.Options.Background {
			opts.SetBackground(true)
		}
		if im.Options.Sparse {
			opts.SetSparse(true)
		}
		if im.Options.ExpireAfterSeconds != nil {
			opts.SetExpireAfterSeconds(*im.Options.ExpireAfterSeconds)
		}
	}

	return mongo.IndexModel{
		Keys:    im.Keys,
		Options: opts,
	}
}

// GetMongoInitCommands 获取 MongoDB 初始化命令列表
// 这些命令是从 scripts/db/mongo.js 转换而来的
func GetMongoInitCommands() []MongoIndexCommand {
	ttl30Days := int32(2592000) // 30天 = 2592000秒

	return []MongoIndexCommand{
		// ==========================================
		// HUB_GW_ACCESS_LOG 集合索引
		// ==========================================

		// 1. 主键唯一索引：tenantId + traceId + gatewayStartProcessingTime
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "tenantId", Value: 1},
					{Key: "traceId", Value: 1},
					{Key: "gatewayStartProcessingTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_tenantId_traceId_unique",
					Unique:     true,
					Background: true,
				},
			},
			Description: "主键唯一索引 - 用于 GetGatewayLogByKey 方法的精确查询",
		},

		// 2. 监控查询主索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "gatewayStartProcessingTime", Value: 1},
					{Key: "tenantId", Value: 1},
					{Key: "gatewayInstanceId", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_monitoring_main",
					Background: true,
				},
			},
			Description: "监控查询主索引",
		},

		// 3. 路由热点索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "gatewayStartProcessingTime", Value: 1},
					{Key: "requestPath", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_hot_routes",
					Background: true,
				},
			},
			Description: "路由热点索引",
		},

		// 4. 日志查询主索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "tenantId", Value: 1},
					{Key: "gatewayStartProcessingTime", Value: 1},
					{Key: "gatewayStatusCode", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_log_query_main",
					Background: true,
				},
			},
			Description: "日志查询主索引",
		},

		// 5. 路由名称索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "gatewayStartProcessingTime", Value: 1},
					{Key: "routeName", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_route_name",
					Background: true,
				},
			},
			Description: "路由名称索引",
		},

		// 6. 响应时间索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "totalProcessingTimeMs", Value: 1},
					{Key: "gatewayStartProcessingTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_response_time",
					Background: true,
					Sparse:     true,
				},
			},
			Description: "响应时间索引（稀疏索引）",
		},

		// 7. 路由配置ID索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "gatewayStartProcessingTime", Value: 1},
					{Key: "routeConfigId", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_route_config_id",
					Background: true,
				},
			},
			Description: "路由配置ID索引",
		},

		// 8. 服务定义ID索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "gatewayStartProcessingTime", Value: 1},
					{Key: "serviceDefinitionId", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_service_definition_id",
					Background: true,
				},
			},
			Description: "服务定义ID索引",
		},

		// 9. 网关实例ID索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "gatewayInstanceId", Value: 1},
					{Key: "gatewayStartProcessingTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_gateway_instance_id",
					Background: true,
				},
			},
			Description: "网关实例ID索引",
		},

		// 10. 网关实例名称索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "gatewayInstanceName", Value: 1},
					{Key: "gatewayStartProcessingTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_gateway_instance_name",
					Background: true,
				},
			},
			Description: "网关实例名称索引",
		},

		// 11. 服务名称索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "serviceName", Value: 1},
					{Key: "gatewayStartProcessingTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_service_name",
					Background: true,
				},
			},
			Description: "服务名称索引",
		},

		// 12. 客户端IP索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "clientIpAddress", Value: 1},
					{Key: "gatewayStartProcessingTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_client_ip",
					Background: true,
				},
			},
			Description: "客户端IP索引",
		},

		// 13. 状态码索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "gatewayStatusCode", Value: 1},
					{Key: "gatewayStartProcessingTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_status_code",
					Background: true,
				},
			},
			Description: "状态码索引",
		},

		// 14. 代理类型索引
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "proxyType", Value: 1},
					{Key: "gatewayStartProcessingTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_proxy_type",
					Background: true,
				},
			},
			Description: "代理类型索引",
		},

		// 15. TTL索引（30天自动过期）
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "gatewayStartProcessingTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:               "idx_ttl_cleanup",
					Background:         true,
					ExpireAfterSeconds: &ttl30Days,
				},
			},
			Description: "TTL索引 - 30天自动清理过期数据",
		},

		// ==========================================
		// HUB_GW_BACKEND_TRACE_LOG 集合索引
		// ==========================================

		// 1. 主键唯一索引：traceId + backendTraceId
		{
			CollectionName: "HUB_GW_BACKEND_TRACE_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "traceId", Value: 1},
					{Key: "backendTraceId", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_traceId_backendTraceId_unique",
					Unique:     true,
					Background: true,
				},
			},
			Description: "主键唯一索引",
		},

		// 2. 租户追踪索引
		{
			CollectionName: "HUB_GW_BACKEND_TRACE_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "tenantId", Value: 1},
					{Key: "traceId", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_tenant_trace",
					Background: true,
				},
			},
			Description: "租户追踪索引",
		},

		// 3. 服务维度索引
		{
			CollectionName: "HUB_GW_BACKEND_TRACE_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "tenantId", Value: 1},
					{Key: "serviceDefinitionId", Value: 1},
					{Key: "requestStartTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_tenant_service_time",
					Background: true,
				},
			},
			Description: "服务维度索引",
		},

		// 4. 时间索引
		{
			CollectionName: "HUB_GW_BACKEND_TRACE_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "requestStartTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_request_start_time",
					Background: true,
				},
			},
			Description: "请求开始时间索引",
		},

		// 5. 追踪状态索引
		{
			CollectionName: "HUB_GW_BACKEND_TRACE_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "tenantId", Value: 1},
					{Key: "traceStatus", Value: 1},
					{Key: "requestStartTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_tenant_status_time",
					Background: true,
				},
			},
			Description: "追踪状态索引",
		},

		// 6. 审计时间索引
		{
			CollectionName: "HUB_GW_BACKEND_TRACE_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "tenantId", Value: 1},
					{Key: "addTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_tenant_add_time",
					Background: true,
				},
			},
			Description: "审计时间索引",
		},

		// 7. TTL索引（30天自动过期）
		{
			CollectionName: "HUB_GW_BACKEND_TRACE_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "requestStartTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:               "idx_ttl_cleanup",
					Background:         true,
					ExpireAfterSeconds: &ttl30Days,
				},
			},
			Description: "TTL索引 - 30天自动清理过期数据",
		},
	}
}

// GetIndexCommandsByCollection 按集合分组获取索引命令
func GetIndexCommandsByCollection() map[string][]MongoIndexCommand {
	commands := GetMongoInitCommands()
	result := make(map[string][]MongoIndexCommand)

	for _, cmd := range commands {
		result[cmd.CollectionName] = append(result[cmd.CollectionName], cmd)
	}

	return result
}

// GetCollectionNames 获取所有需要创建索引的集合名称
func GetCollectionNames() []string {
	commandsByCollection := GetIndexCommandsByCollection()
	collections := make([]string, 0, len(commandsByCollection))

	for collName := range commandsByCollection {
		collections = append(collections, collName)
	}

	return collections
}
