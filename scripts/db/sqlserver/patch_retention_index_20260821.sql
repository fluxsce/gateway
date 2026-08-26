-- SQL Server 方言，由 scripts/db/mysql/patch_retention_index_20260821.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
-- 已有库补丁：归档清理复合索引。新库执行建表脚本即可，不必再跑本文件。
-- 条件均为 tenantId = ? AND 时间列 < ?。已具备同类索引的表不再补。

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'IDX_ALERT_LOG_CLEANUP' AND object_id = OBJECT_ID(N'dbo.HUB_ALERT_LOG'))
CREATE INDEX IDX_ALERT_LOG_CLEANUP ON HUB_ALERT_LOG (tenantId, alertTimestamp);
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'IDX_TIMER_LOG_CLEANUP' AND object_id = OBJECT_ID(N'dbo.HUB_TIMER_EXECUTION_LOG'))
CREATE INDEX IDX_TIMER_LOG_CLEANUP ON HUB_TIMER_EXECUTION_LOG (tenantId, addTime);
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'IDX_CLS_EVT_CLEANUP' AND object_id = OBJECT_ID(N'dbo.HUB_CLUSTER_EVENT'))
CREATE INDEX IDX_CLS_EVT_CLEANUP ON HUB_CLUSTER_EVENT (tenantId, eventTime);
