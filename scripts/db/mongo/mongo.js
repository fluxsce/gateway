// ==========================================
// MongoDB 索引：最小集（按写入成本约束）
// 覆盖详情、按实例列表/监控、TTL 清理。
// 路由名、状态码、IP、路径等筛选走实例+时间后再过滤，不再各建一条。
// 对齐 GetMongoInitCommands（internal/script/mongo/mongo_commands.go）
// ==========================================

// use your_database_name;

print("开始创建 MongoDB 最小索引集...\n");

var collection = db.HUB_GW_ACCESS_LOG;

print("[HUB_GW_ACCESS_LOG] 创建索引...");

// 详情 GetGatewayLogByKey：{tenantId, traceId}。不加 unique，避免历史重复键导致整条建失败。
collection.createIndex(
    { "tenantId": 1, "traceId": 1 },
    { "name": "idx_tenant_trace", "background": true }
);
print("  1. idx_tenant_trace (详情)");

// 列表 / 监控 / 应用层清理：实例 + 时间倒序（前端按实例查，cleaner 同条件）
collection.createIndex(
    { "gatewayInstanceId": 1, "gatewayStartProcessingTime": -1 },
    { "name": "idx_instance_time", "background": true }
);
print("  2. idx_instance_time (列表/监控/清理)");

// TTL 必须单字段 Date。与复合索引键不同，可并存。
collection.createIndex(
    { "gatewayStartProcessingTime": 1 },
    { "name": "idx_ttl_cleanup", "background": true, "expireAfterSeconds": 2592000 }
);
print("  3. idx_ttl_cleanup (TTL 30天)");

print("[HUB_GW_ACCESS_LOG] 完成（3 条 + 默认 _id）\n");

print("[HUB_GW_BACKEND_TRACE_LOG] 创建索引...");

var backendTraceCollection = db.HUB_GW_BACKEND_TRACE_LOG;

// 按 trace 拉列表走左前缀；按主键取单条带上 backendTraceId
backendTraceCollection.createIndex(
    { "tenantId": 1, "traceId": 1, "backendTraceId": 1 },
    { "name": "idx_tenant_trace", "background": true }
);
print("  1. idx_tenant_trace (从表列表/单条)");

backendTraceCollection.createIndex(
    { "requestStartTime": 1 },
    { "name": "idx_ttl_cleanup", "background": true, "expireAfterSeconds": 2592000 }
);
print("  2. idx_ttl_cleanup (TTL 30天，兼时间扫描)");

print("[HUB_GW_BACKEND_TRACE_LOG] 完成（2 条 + 默认 _id）\n");

print("MongoDB 最小索引集创建完成");
print("- 主表 3 条，从表 2 条（另有默认 _id）");
print("- 程序启动会自动建这 5 条并删除集合上其余旧索引，无需手工 drop");
