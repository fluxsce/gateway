-- 已有库补丁：归档清理复合索引。新库执行建表脚本即可，不必再跑本文件。
-- 条件均为 tenantId = ? AND 时间列 < ?。已具备同类索引的表不再补。

CREATE INDEX IDX_ALERT_LOG_CLEANUP ON HUB_ALERT_LOG (tenantId, alertTimestamp);
CREATE INDEX IDX_TIMER_LOG_CLEANUP ON HUB_TIMER_EXECUTION_LOG (tenantId, addTime);
CREATE INDEX IDX_CLS_EVT_CLEANUP ON HUB_CLUSTER_EVENT (tenantId, eventTime);
