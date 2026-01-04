// ==========================================
// MongoDB 索引创建脚本 - 完整版
// 用于优化 HUB_GW_ACCESS_LOG 和 HUB_GW_BACKEND_TRACE_LOG 集合的查询性能
// 基于 mongo_query_dao.go 和 mongo_monitoring_dao.go 中的查询条件设计
// 对齐 MySQL 版本（scripts/db/mysql.sql 第1229-1335行）
// ==========================================

// 使用目标数据库
// use your_database_name;

print("开始创建 MongoDB 索引...\n");

// ==========================================
// 第一部分：HUB_GW_ACCESS_LOG 主表索引
// ==========================================

// 获取 HUB_GW_ACCESS_LOG 集合
var collection = db.HUB_GW_ACCESS_LOG;

print("[HUB_GW_ACCESS_LOG] 创建索引...");

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
print("  1. idx_tenantId_traceId_unique (主键唯一索引)");

collection.createIndex(
    { "gatewayStartProcessingTime": 1, "tenantId": 1, "gatewayInstanceId": 1 }, 
    { "name": "idx_monitoring_main", "background": true }
);
print("  2. idx_monitoring_main (监控查询主索引)");

collection.createIndex(
    { "gatewayStartProcessingTime": 1, "requestPath": 1 }, 
    { "name": "idx_hot_routes", "background": true }
);
print("  3. idx_hot_routes (路由热点索引)");

collection.createIndex(
    { "tenantId": 1, "gatewayStartProcessingTime": 1, "gatewayStatusCode": 1 }, 
    { "name": "idx_log_query_main", "background": true }
);
print("  4. idx_log_query_main (日志查询主索引)");

collection.createIndex(
    { "gatewayStartProcessingTime": 1, "routeName": 1 }, 
    { "name": "idx_route_name", "background": true }
);
print("  5. idx_route_name (路由名称索引)");

collection.createIndex(
    { "totalProcessingTimeMs": 1, "gatewayStartProcessingTime": 1 }, 
    { "name": "idx_response_time", "background": true, "sparse": true }
);
print("  6. idx_response_time (响应时间索引)");

collection.createIndex(
    { "gatewayStartProcessingTime": 1, "routeConfigId": 1 }, 
    { "name": "idx_route_config_id", "background": true }
);
print("  7. idx_route_config_id (路由配置ID索引)");

collection.createIndex(
    { "gatewayStartProcessingTime": 1, "serviceDefinitionId": 1 }, 
    { "name": "idx_service_definition_id", "background": true }
);
print("  8. idx_service_definition_id (服务定义ID索引)");

collection.createIndex(
    { "gatewayInstanceId": 1, "gatewayStartProcessingTime": 1 }, 
    { "name": "idx_gateway_instance_id", "background": true }
);
print("  9. idx_gateway_instance_id (网关实例ID索引)");

collection.createIndex(
    { "gatewayInstanceName": 1, "gatewayStartProcessingTime": 1 }, 
    { "name": "idx_gateway_instance_name", "background": true }
);
print(" 10. idx_gateway_instance_name (网关实例名称索引)");

collection.createIndex(
    { "serviceName": 1, "gatewayStartProcessingTime": 1 }, 
    { "name": "idx_service_name", "background": true }
);
print(" 11. idx_service_name (服务名称索引)");

collection.createIndex(
    { "clientIpAddress": 1, "gatewayStartProcessingTime": 1 }, 
    { "name": "idx_client_ip", "background": true }
);
print(" 12. idx_client_ip (客户端IP索引)");

collection.createIndex(
    { "gatewayStatusCode": 1, "gatewayStartProcessingTime": 1 }, 
    { "name": "idx_status_code", "background": true }
);
print(" 13. idx_status_code (状态码索引)");

collection.createIndex(
    { "proxyType": 1, "gatewayStartProcessingTime": 1 }, 
    { "name": "idx_proxy_type", "background": true }
);
print(" 14. idx_proxy_type (代理类型索引)");

collection.createIndex(
    { "gatewayStartProcessingTime": 1 }, 
    { "name": "idx_ttl_cleanup", "background": true, "expireAfterSeconds": 2592000 }
);
print(" 15. idx_ttl_cleanup (TTL索引，30天过期)");

print("[HUB_GW_ACCESS_LOG] 索引创建完成\n");

// ==========================================
// 第二部分：HUB_GW_BACKEND_TRACE_LOG 从表索引
// ==========================================

print("[HUB_GW_BACKEND_TRACE_LOG] 创建索引...");

// 获取 HUB_GW_BACKEND_TRACE_LOG 集合
var backendTraceCollection = db.HUB_GW_BACKEND_TRACE_LOG;

backendTraceCollection.createIndex(
    { "traceId": 1, "backendTraceId": 1 }, 
    { "name": "idx_traceId_backendTraceId_unique", "unique": true, "background": true }
);
print("  1. idx_traceId_backendTraceId_unique (主键唯一索引)");

backendTraceCollection.createIndex(
    { "tenantId": 1, "traceId": 1 }, 
    { "name": "idx_tenant_trace", "background": true }
);
print("  2. idx_tenant_trace (租户追踪索引)");

backendTraceCollection.createIndex(
    { "tenantId": 1, "serviceDefinitionId": 1, "requestStartTime": 1 }, 
    { "name": "idx_tenant_service_time", "background": true }
);
print("  3. idx_tenant_service_time (服务维度索引)");

backendTraceCollection.createIndex(
    { "requestStartTime": 1 }, 
    { "name": "idx_request_start_time", "background": true }
);
print("  4. idx_request_start_time (时间索引)");

backendTraceCollection.createIndex(
    { "tenantId": 1, "traceStatus": 1, "requestStartTime": 1 }, 
    { "name": "idx_tenant_status_time", "background": true }
);
print("  5. idx_tenant_status_time (追踪状态索引)");

backendTraceCollection.createIndex(
    { "tenantId": 1, "addTime": 1 }, 
    { "name": "idx_tenant_add_time", "background": true }
);
print("  6. idx_tenant_add_time (审计时间索引)");

backendTraceCollection.createIndex(
    { "requestStartTime": 1 }, 
    { "name": "idx_ttl_cleanup", "background": true, "expireAfterSeconds": 2592000 }
);
print("  7. idx_ttl_cleanup (TTL索引，30天过期)");

print("[HUB_GW_BACKEND_TRACE_LOG] 索引创建完成\n");

// ==========================================
// 索引创建完成汇总
// ==========================================

print("MongoDB 索引创建完成");
print("- 主表索引：16个（含默认_id索引）");
print("- 从表索引：8个（含默认_id索引）");
print("- TTL设置：30天自动清理过期数据");
