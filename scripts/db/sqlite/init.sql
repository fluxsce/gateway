-- =====================================================
-- SQLite数据库初始化脚本
-- 基于MySQL脚本直接翻译，保持原有表结构
-- 创建时间: 2024-12-19
-- 说明: 
-- 1. 保持与MySQL相同的表结构和字段名
-- 2. 将MySQL数据类型映射为SQLite对应类型
-- 3. 不添加额外的CHECK约束
-- 4. 保持原有的索引和约束逻辑
-- =====================================================

-- 启用外键约束
PRAGMA foreign_keys = ON;

-- 启用WAL模式以支持并发
PRAGMA journal_mode = WAL;

-- =====================================================
-- Source all table files
-- =====================================================

.read HUB_USER.sql
.read HUB_LOGIN_LOG.sql
.read HUB_GW_INSTANCE.sql
.read HUB_GW_ROUTER_CONFIG.sql
.read HUB_GW_ROUTE_CONFIG.sql
.read HUB_GW_ROUTE_ASSERTION.sql
.read HUB_GW_FILTER_CONFIG.sql
.read HUB_GW_CORS_CONFIG.sql
.read HUB_GW_RATE_LIMIT_CONFIG.sql
.read HUB_GW_CIRCUIT_BREAKER_CONFIG.sql
.read HUB_GW_AUTH_CONFIG.sql
.read HUB_GW_SERVICE_DEFINITION.sql
.read HUB_GW_SERVICE_NODE.sql
.read HUB_GW_PROXY_CONFIG.sql
.read HUB_TIMER_SCHEDULER.sql
.read HUB_TIMER_TASK.sql
.read HUB_TIMER_EXECUTION_LOG.sql
.read HUB_TOOL_CONFIG.sql
.read HUB_TOOL_CONFIG_GROUP.sql
.read HUB_GW_LOG_CONFIG.sql
.read HUB_GW_ACCESS_LOG.sql
.read HUB_GW_BACKEND_TRACE_LOG.sql
.read HUB_GW_SECURITY_CONFIG.sql
.read HUB_GW_IP_ACCESS_CONFIG.sql
.read HUB_GW_UA_ACCESS_CONFIG.sql
.read HUB_GW_API_ACCESS_CONFIG.sql
.read HUB_GW_DOMAIN_ACCESS_CONFIG.sql
.read HUB_METRIC_SERVER_INFO.sql
.read HUB_METRIC_CPU_LOG.sql
.read HUB_METRIC_MEMORY_LOG.sql
.read HUB_METRIC_DISK_PART_LOG.sql
.read HUB_METRIC_DISK_IO_LOG.sql
.read HUB_METRIC_NETWORK_LOG.sql
.read HUB_METRIC_PROCESS_LOG.sql
.read HUB_METRIC_PROCSTAT_LOG.sql
.read HUB_METRIC_TEMP_LOG.sql
.read HUB_REGISTRY_SERVICE_GROUP.sql
.read HUB_REGISTRY_SERVICE.sql
.read HUB_REGISTRY_SERVICE_INSTANCE.sql
.read HUB_REGISTRY_SERVICE_EVENT.sql
.read HUB_MONITOR_JVM_RESOURCE.sql
.read HUB_MONITOR_JVM_MEMORY.sql
.read HUB_MONITOR_JVM_MEM_POOL.sql
.read HUB_MONITOR_JVM_GC.sql
.read HUB_MONITOR_JVM_THREAD.sql
.read HUB_MONITOR_JVM_THR_STATE.sql
.read HUB_MONITOR_JVM_DEADLOCK.sql
.read HUB_MONITOR_JVM_CLASS.sql
.read HUB_MONITOR_APP_DATA.sql
.read HUB_TUNNEL_SERVER.sql
.read HUB_TUNNEL_SERVER_NODE.sql
.read HUB_TUNNEL_CLIENT.sql
.read HUB_TUNNEL_SERVICE.sql
.read HUB_AUTH_ROLE.sql
.read HUB_AUTH_RESOURCE.sql
.read HUB_AUTH_ROLE_RESOURCE.sql
.read HUB_AUTH_USER_ROLE.sql
.read HUB_AUTH_DATA_PERMISSION.sql

-- 索引说明
-- ==========================================
-- 1. 所有表都建立了tenantId相关的复合主键，支持多租户数据隔离
-- 2. 为关联字段（jvmResourceId等）创建了索引，提高关联查询性能
-- 3. 为时间字段（collectionTime）创建了索引，支持时间范围查询
-- 4. 为健康状态字段创建了索引，便于快速筛选异常数据
-- 5. 为常用查询条件字段创建了索引，提高查询效率

-- ==========================================
-- 表关系说明
-- ==========================================
-- HUB_MONITOR_JVM_RESOURCE (主表)
--   ├── HUB_MONITOR_JVM_MEMORY (1:N，一个JVM资源对应多个内存记录：堆内存+非堆内存)
--   ├── HUB_MONITOR_JVM_MEM_POOL (1:N，一个JVM资源对应多个内存池)
--   ├── HUB_MONITOR_JVM_GC (1:N，一个JVM资源对应多个GC收集器)
--   ├── HUB_MONITOR_JVM_THREAD (1:1，一个JVM资源对应一个线程信息记录)
--   │   ├── HUB_MONITOR_JVM_THR_STATE (1:1，一个线程信息对应一个线程状态统计)
--   │   └── HUB_MONITOR_JVM_DEADLOCK (1:1，一个线程信息对应一个死锁检测记录)
--   ├── HUB_MONITOR_JVM_CLASS (1:1，一个JVM资源对应一个类加载信息记录)
--   └── HUB_MONITOR_APP_DATA (1:N，一个JVM资源对应多个应用监控数据)
-- ==========================================

-- =====================================================
-- SQLite特殊配置和优化
-- =====================================================

-- 设置同步模式为NORMAL以平衡性能和安全性
PRAGMA synchronous = NORMAL;

-- 设置页面缓存大小（2MB）
PRAGMA cache_size = -2000;

-- 设置临时存储为内存模式
PRAGMA temp_store = MEMORY;

-- 设置锁定超时时间（30秒）
PRAGMA busy_timeout = 30000;

-- 分析数据库以优化查询计划
ANALYZE;

-- =====================================================
-- 脚本执行完成提示
-- =====================================================
SELECT 'SQLite数据库初始化完成！' as message,
       'Created ' || COUNT(*) || ' tables' as table_count
FROM sqlite_master 
WHERE type = 'table' AND name LIKE 'HUB_%';
