// ==========================================
// MongoDB 索引创建脚本 - 精简版
// 用于优化 HUB_GW_ACCESS_LOG 集合的查询性能
// 基于 mongo_query_dao.go 和 mongo_monitoring_dao.go 中的查询条件设计
// ==========================================

// 使用目标数据库
// use your_database_name;

// 获取 HUB_GW_ACCESS_LOG 集合
var collection = db.HUB_GW_ACCESS_LOG;

print("开始创建 HUB_GW_ACCESS_LOG 集合核心索引...");

// ==========================================
// 1. 主键索引（必须）
// ==========================================

// 主键唯一索引：tenantId + traceId
// 用于 GetGatewayLogByKey 方法的精确查询
collection.createIndex(
    { "tenantId": 1, "traceId": 1, "gatewayStartProcessingTime": 1 }, 
    { 
        "name": "idx_tenantId_traceId_unique", 
        "unique": true,
        "background": true
    }
);
print("✓ 创建主键索引：idx_tenantId_traceId_unique");

// ==========================================
// 2. 核心监控索引（覆盖大部分查询）
// ==========================================

// 监控查询主索引：时间 + 租户 + 网关实例
// 可以覆盖：时间查询、时间+租户查询、时间+租户+网关实例查询
// 支持所有监控方法：GetGatewayMonitoringOverview、GetRequestMetricsTrend 等
collection.createIndex(
    { "gatewayStartProcessingTime": 1, "tenantId": 1, "gatewayInstanceId": 1 }, 
    { 
        "name": "idx_monitoring_main", 
        "background": true
    }
);
print("✓ 创建监控主索引：idx_monitoring_main");

// 路由热点索引：时间 + 请求路径
// 用于 GetHotRoutes 方法和路径相关查询
collection.createIndex(
    { "gatewayStartProcessingTime": 1, "requestPath": 1 }, 
    { 
        "name": "idx_hot_routes", 
        "background": true
    }
);
print("✓ 创建路由热点索引：idx_hot_routes");

// ==========================================
// 3. 日志查询索引（用于列表查询）
// ==========================================

// 日志查询主索引：租户 + 时间 + 状态码
// 可以覆盖大部分日志查询场景
collection.createIndex(
    { "tenantId": 1, "gatewayStartProcessingTime": 1, "gatewayStatusCode": 1 }, 
    { 
        "name": "idx_log_query_main", 
        "background": true
    }
);
print("✓ 创建日志查询主索引：idx_log_query_main");

// ==========================================
// 4. 文本搜索索引（关键词搜索）
// ==========================================

// 文本搜索索引：支持关键词搜索
// 用于 buildGatewayLogFilter 中的 $or 关键词搜索
// collection.createIndex(
//     { 
//         "traceId": "text", 
//         "requestPath": "text", 
//         "serviceName": "text", 
//         "errorMessage": "text" 
//     }, 
//     { 
//         "name": "idx_text_search", 
//         "background": true,
//         "weights": {
//             "traceId": 10,        // 链路追踪ID权重最高
//             "serviceName": 5,     // 服务名称权重较高
//             "requestPath": 3,     // 请求路径权重中等
//             "errorMessage": 1     // 错误信息权重最低
//         }
//     }
// );
// print("✓ 创建文本搜索索引：idx_text_search");
collection.createIndex(
    { "gatewayStartProcessingTime": 1, "routeName": 1 }, 
    { 
        "name": "idx_route_name", 
        "background": true
    }
);
print("✓ 创建路由名称索引：idx_route_name");

// ==========================================
// 5. 性能分析索引（可选）
// ==========================================

// 响应时间分析索引：响应时间 + 时间
// 用于性能分析和响应时间范围查询，稀疏索引
collection.createIndex(
    { "totalProcessingTimeMs": 1, "gatewayStartProcessingTime": 1 }, 
    { 
        "name": "idx_response_time", 
        "background": true,
        "sparse": true  // 稀疏索引，因为响应时间可能为null
    }
);
print("✓ 创建响应时间索引：idx_response_time");

// ==========================================
// 6. TTL索引（数据生命周期管理）
// ==========================================

// TTL索引：自动删除30天前的数据
// 可根据实际需求调整过期时间
collection.createIndex(
    { "gatewayStartProcessingTime": 1 }, 
    { 
        "name": "idx_ttl_cleanup", 
        "background": true,
        "expireAfterSeconds": 2592000  // 30天 = 30 * 24 * 60 * 60 秒
    }
);
print("✓ 创建TTL索引：idx_ttl_cleanup (30天过期)");

// ==========================================
// 索引创建完成
// ==========================================

print("\n==========================================");
print("HUB_GW_ACCESS_LOG 集合核心索引创建完成！");
print("==========================================");

// 显示所有索引
print("\n当前集合的所有索引：");
db.HUB_GW_ACCESS_LOG.getIndexes().forEach(function(index) {
    print("- " + index.name + ": " + JSON.stringify(index.key));
});
