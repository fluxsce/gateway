-- =====================================================
-- MySQL数据库初始化脚本
-- =====================================================
-- 此脚本将按顺序执行所有表的创建语句
-- 使用方法：在MySQL客户端中执行：source init.sql 或 mysql < init.sql
-- =====================================================

source HUB_USER.sql;
source HUB_LOGIN_LOG.sql;
source HUB_GW_INSTANCE.sql;
source HUB_GW_ROUTER_CONFIG.sql;
source HUB_GW_ROUTE_CONFIG.sql;
source HUB_GW_ROUTE_ASSERTION.sql;
source HUB_GW_FILTER_CONFIG.sql;
source HUB_GW_CORS_CONFIG.sql;
source HUB_GW_RATE_LIMIT_CONFIG.sql;
source HUB_GW_CIRCUIT_BREAKER_CONFIG.sql;
source HUB_GW_AUTH_CONFIG.sql;
source HUB_GW_SERVICE_DEFINITION.sql;
source HUB_GW_SERVICE_NODE.sql;
source HUB_GW_PROXY_CONFIG.sql;
source HUB_TIMER_SCHEDULER.sql;
source HUB_TIMER_TASK.sql;
source HUB_TIMER_EXECUTION_LOG.sql;
source HUB_TOOL_CONFIG.sql;
source HUB_TOOL_CONFIG_GROUP.sql;
source HUB_GW_LOG_CONFIG.sql;
source HUB_GW_ACCESS_LOG.sql;
source HUB_GW_BACKEND_TRACE_LOG.sql;
source HUB_GW_SECURITY_CONFIG.sql;
source HUB_GW_IP_ACCESS_CONFIG.sql;
source HUB_GW_UA_ACCESS_CONFIG.sql;
source HUB_GW_API_ACCESS_CONFIG.sql;
source HUB_GW_DOMAIN_ACCESS_CONFIG.sql;
source HUB_METRIC_SERVER_INFO.sql;
source HUB_METRIC_CPU_LOG.sql;
source HUB_METRIC_MEMORY_LOG.sql;
source HUB_METRIC_DISK_PART_LOG.sql;
source HUB_METRIC_DISK_IO_LOG.sql;
source HUB_METRIC_NETWORK_LOG.sql;
source HUB_METRIC_PROCESS_LOG.sql;
source HUB_METRIC_PROCSTAT_LOG.sql;
source HUB_METRIC_TEMP_LOG.sql;
source HUB_REGISTRY_SERVICE_GROUP.sql;
source HUB_REGISTRY_SERVICE.sql;
source HUB_REGISTRY_SERVICE_INSTANCE.sql;
source HUB_REGISTRY_SERVICE_EVENT.sql;
source HUB_MONITOR_JVM_RESOURCE.sql;
source HUB_MONITOR_JVM_MEMORY.sql;
source HUB_MONITOR_JVM_MEM_POOL.sql;
source HUB_MONITOR_JVM_GC.sql;
source HUB_MONITOR_JVM_THREAD.sql;
source HUB_MONITOR_JVM_THR_STATE.sql;
source HUB_MONITOR_JVM_DEADLOCK.sql;
source HUB_MONITOR_JVM_CLASS.sql;
source HUB_MONITOR_APP_DATA.sql;
source HUB_TUNNEL_SERVER.sql;
source HUB_TUNNEL_SERVER_NODE.sql;
source HUB_TUNNEL_CLIENT.sql;
source HUB_TUNNEL_SERVICE.sql;

-- =====================================================
-- 字段长度调整：支持多服务定义ID和服务名称（多服务场景）
-- 注意：此处使用独立ALTER语句，避免直接修改历史建表语句，保证向后兼容
-- 变更内容：
--  1) HUB_GW_ACCESS_LOG.serviceDefinitionId 扩展为 1000 字符
--  2) HUB_GW_ACCESS_LOG.serviceName 扩展为 1000 字符
--  3) HUB_GW_ROUTE_CONFIG.serviceDefinitionId 扩展为 1000 字符
-- =====================================================
ALTER TABLE `HUB_GW_ACCESS_LOG`
  MODIFY COLUMN `serviceDefinitionId` VARCHAR(1000) DEFAULT NULL COMMENT '服务定义ID（支持多服务，逗号分隔）';

ALTER TABLE `HUB_GW_ACCESS_LOG`
  MODIFY COLUMN `serviceName` VARCHAR(1000) DEFAULT NULL COMMENT '服务名称(冗余字段,便于查询显示,支持多服务)';

ALTER TABLE `HUB_GW_ROUTE_CONFIG`
  MODIFY COLUMN `serviceDefinitionId` VARCHAR(1000) DEFAULT NULL COMMENT '关联的服务定义ID（支持多服务，逗号分隔）';
