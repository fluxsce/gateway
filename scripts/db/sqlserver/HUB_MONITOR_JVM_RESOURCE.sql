-- SQL Server 方言，由 scripts/db/mysql/HUB_MONITOR_JVM_RESOURCE.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_MONITOR_JVM_RESOURCE', N'U') IS NULL
CREATE TABLE HUB_MONITOR_JVM_RESOURCE (
                                          jvmResourceId NVARCHAR(100) NOT NULL,
                                          tenantId NVARCHAR(32) NOT NULL,
                                          serviceGroupId NVARCHAR(32) NOT NULL,

    -- 应用标识信息
                                          applicationName NVARCHAR(100) NOT NULL,
                                          groupName NVARCHAR(100) NOT NULL,
                                          hostName NVARCHAR(100) DEFAULT NULL,
                                          hostIpAddress NVARCHAR(50) DEFAULT NULL,

    -- 时间相关字段
                                          collectionTime DATETIME2 NOT NULL,
                                          jvmStartTime DATETIME2 NOT NULL,
                                          jvmUptimeMs BIGINT NOT NULL DEFAULT 0,

    -- 健康状态字段
                                          healthyFlag NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                          healthGrade NVARCHAR(20) DEFAULT NULL,
                                          requiresAttentionFlag NVARCHAR(1) NOT NULL DEFAULT N'N',
                                          summaryText NVARCHAR(500) DEFAULT NULL,

    -- 系统属性（JSON格式）
                                          systemPropertiesJson NVARCHAR(MAX) DEFAULT NULL,

    -- 通用字段
                                          addTime DATETIME2 NOT NULL DEFAULT GETDATE(),
                                          addWho NVARCHAR(32) DEFAULT NULL,
                                          editTime DATETIME2 NOT NULL DEFAULT GETDATE(),
                                          editWho NVARCHAR(32) DEFAULT NULL,
                                          oprSeqFlag NVARCHAR(32) DEFAULT NULL,
                                          currentVersion INT NOT NULL DEFAULT 1,
                                          activeFlag NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                          noteText NVARCHAR(500) DEFAULT NULL,

                                          PRIMARY KEY (tenantId, serviceGroupId, jvmResourceId),
                                          INDEX IDX_MONITOR_JVM_APP (applicationName),
                                          INDEX IDX_MONITOR_JVM_TIME (collectionTime),
                                          INDEX IDX_MONITOR_JVM_HEALTH (healthyFlag, requiresAttentionFlag),
                                          INDEX IDX_MONITOR_JVM_HOST (hostIpAddress),
                                          INDEX IDX_MONITOR_JVM_GROUP (serviceGroupId, groupName)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN applicationName NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN groupName NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN hostName NVARCHAR(100) NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN hostIpAddress NVARCHAR(50) NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN healthyFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN healthGrade NVARCHAR(20) NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN requiresAttentionFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN summaryText NVARCHAR(500) NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN addWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN editWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN oprSeqFlag NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_RESOURCE ALTER COLUMN noteText NVARCHAR(500) NULL;
