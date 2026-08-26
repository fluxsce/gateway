-- SQL Server 方言，由 scripts/db/mysql/init.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
-- =====================================================
-- SQL Server数据库初始化脚本
-- =====================================================
-- 此脚本将按顺序执行所有表的创建语句
-- 使用方法：在SQL Server客户端中执行：:r init.sql 或 sqlcmd -i init.sql
-- =====================================================

:r HUB_USER.sql;
:r HUB_LOGIN_LOG.sql;
:r HUB_GW_INSTANCE.sql;
:r HUB_GW_ROUTER_CONFIG.sql;
:r HUB_GW_ROUTE_CONFIG.sql;
:r HUB_GW_ROUTE_ASSERTION.sql;
:r HUB_GW_FILTER_CONFIG.sql;
:r HUB_GW_CORS_CONFIG.sql;
:r HUB_GW_RATE_LIMIT_CONFIG.sql;
:r HUB_GW_CIRCUIT_BREAKER_CONFIG.sql;
:r HUB_GW_STATIC_HOST_CONFIG.sql;
:r HUB_GW_AUTH_CONFIG.sql;
:r HUB_GW_SERVICE_DEFINITION.sql;
:r HUB_GW_SERVICE_NODE.sql;
:r HUB_GW_PROXY_CONFIG.sql;
:r HUB_TIMER_SCHEDULER.sql;
:r HUB_TIMER_TASK.sql;
:r HUB_TIMER_EXECUTION_LOG.sql;
:r HUB_TOOL_CONFIG.sql;
:r HUB_TOOL_CONFIG_GROUP.sql;
:r HUB_GW_LOG_CONFIG.sql;
:r HUB_GW_ACCESS_LOG.sql;
:r HUB_GW_BACKEND_TRACE_LOG.sql;
:r HUB_GW_SECURITY_CONFIG.sql;
:r HUB_GW_IP_ACCESS_CONFIG.sql;
:r HUB_GW_UA_ACCESS_CONFIG.sql;
:r HUB_GW_API_ACCESS_CONFIG.sql;
:r HUB_GW_DOMAIN_ACCESS_CONFIG.sql;
:r HUB_METRIC_SERVER_INFO.sql;
:r HUB_METRIC_CPU_LOG.sql;
:r HUB_METRIC_MEMORY_LOG.sql;
:r HUB_METRIC_DISK_PART_LOG.sql;
:r HUB_METRIC_DISK_IO_LOG.sql;
:r HUB_METRIC_NETWORK_LOG.sql;
:r HUB_METRIC_PROCESS_LOG.sql;
:r HUB_METRIC_PROCSTAT_LOG.sql;
:r HUB_METRIC_TEMP_LOG.sql;
-- 源目录无此文件，已跳过: HUB_REGISTRY_SERVICE_GROUP.sql
-- 源目录无此文件，已跳过: HUB_REGISTRY_SERVICE.sql
-- 源目录无此文件，已跳过: HUB_REGISTRY_SERVICE_INSTANCE.sql
-- 源目录无此文件，已跳过: HUB_REGISTRY_SERVICE_EVENT.sql
:r HUB_MONITOR_JVM_RESOURCE.sql;
:r HUB_MONITOR_JVM_MEMORY.sql;
:r HUB_MONITOR_JVM_MEM_POOL.sql;
:r HUB_MONITOR_JVM_GC.sql;
:r HUB_MONITOR_JVM_THREAD.sql;
:r HUB_MONITOR_JVM_THR_STATE.sql;
:r HUB_MONITOR_JVM_DEADLOCK.sql;
:r HUB_MONITOR_JVM_CLASS.sql;
:r HUB_MONITOR_APP_DATA.sql;
:r HUB_TUNNEL_SERVER.sql;
:r HUB_TUNNEL_SERVER_NODE.sql;
:r HUB_TUNNEL_CLIENT.sql;
:r HUB_TUNNEL_SERVICE.sql;
:r HUB_AUTH_AUDIT_LOG.sql;

-- =====================================================
-- 字段长度调整：支持多服务定义ID和服务名称（多服务场景）
-- 注意：此处使用独立ALTER语句，避免直接修改历史建表语句，保证向后兼容
-- 变更内容：
--  1) HUB_GW_ACCESS_LOG.serviceDefinitionId 扩展为 1000 字符
--  2) HUB_GW_ACCESS_LOG.serviceName 扩展为 1000 字符
--  3) HUB_GW_ROUTE_CONFIG.serviceDefinitionId 扩展为 1000 字符
-- =====================================================
ALTER TABLE HUB_GW_ACCESS_LOG ALTER COLUMN serviceDefinitionId VARCHAR(1000) NULL;

ALTER TABLE HUB_GW_ACCESS_LOG ALTER COLUMN serviceName NVARCHAR(1000) NULL;

ALTER TABLE HUB_GW_ROUTE_CONFIG ALTER COLUMN serviceDefinitionId VARCHAR(1000) NULL;
