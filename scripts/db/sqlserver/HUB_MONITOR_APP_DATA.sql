-- SQL Server 方言，由 scripts/db/mysql/HUB_MONITOR_APP_DATA.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_MONITOR_APP_DATA', N'U') IS NULL
CREATE TABLE HUB_MONITOR_APP_DATA (
                                      appDataId NVARCHAR(32) NOT NULL,
                                      tenantId NVARCHAR(32) NOT NULL,
                                      jvmResourceId NVARCHAR(100) NOT NULL,

    -- 数据分类标识
                                      dataType NVARCHAR(50) NOT NULL,
                                      dataName NVARCHAR(100) NOT NULL,
                                      dataCategory NVARCHAR(50) DEFAULT NULL,

    -- 监控数据（JSON格式存储，支持不同类型的数据结构）
                                      dataJson NVARCHAR(MAX) NOT NULL,

    -- 核心指标（从JSON中提取的关键指标，便于查询和索引）
                                      primaryValue DECIMAL(20,4) DEFAULT NULL,
                                      secondaryValue DECIMAL(20,4) DEFAULT NULL,
                                      statusValue NVARCHAR(50) DEFAULT NULL,

    -- 健康状态
                                      healthyFlag NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                      healthGrade NVARCHAR(20) DEFAULT NULL,
                                      requiresAttentionFlag NVARCHAR(1) NOT NULL DEFAULT N'N',

    -- 标签和维度（便于分组查询）
                                      tagsJson NVARCHAR(MAX) DEFAULT NULL,

    -- 时间字段
                                      collectionTime DATETIME2 NOT NULL,

    -- 通用字段
                                      addTime DATETIME2 NOT NULL DEFAULT GETDATE(),
                                      addWho NVARCHAR(32) DEFAULT NULL,
                                      editTime DATETIME2 NOT NULL DEFAULT GETDATE(),
                                      editWho NVARCHAR(32) DEFAULT NULL,
                                      oprSeqFlag NVARCHAR(32) DEFAULT NULL,
                                      currentVersion INT NOT NULL DEFAULT 1,
                                      activeFlag NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                      noteText NVARCHAR(500) DEFAULT NULL,

                                      PRIMARY KEY (tenantId, appDataId),
                                      INDEX IDX_MONITOR_APP_DATA_RES (jvmResourceId),
                                      INDEX IDX_MONITOR_APP_DATA_TYPE (dataType),
                                      INDEX IDX_MONITOR_APP_DATA_NAME (dataName),
                                      INDEX IDX_MONITOR_APP_DATA_TIME (collectionTime),
                                      INDEX IDX_MONITOR_APP_DATA_HEALTH (healthyFlag, requiresAttentionFlag),
                                      INDEX IDX_MONITOR_APP_DATA_PRIMARY (primaryValue),
                                      INDEX IDX_MONITOR_APP_DATA_STATUS (statusValue),
                                      INDEX IDX_MONITOR_APP_DATA_RES_TYPE_NAME (jvmResourceId, dataType, dataName, collectionTime)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN dataType NVARCHAR(50) NOT NULL;
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN dataName NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN dataCategory NVARCHAR(50) NULL;
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN statusValue NVARCHAR(50) NULL;
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN healthyFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN healthGrade NVARCHAR(20) NULL;
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN requiresAttentionFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN addWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN editWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN oprSeqFlag NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_APP_DATA ALTER COLUMN noteText NVARCHAR(500) NULL;
