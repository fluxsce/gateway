CREATE TABLE HUB_MONITOR_APP_DATA (
                                      appDataId VARCHAR(32) NOT NULL COMMENT '应用监控数据ID，主键',
                                      tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                      jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 数据分类标识
                                      dataType VARCHAR(50) NOT NULL COMMENT '数据类型(THREAD_POOL:线程池/CONNECTION_POOL:连接池/CUSTOM_METRIC:自定义指标/CACHE_POOL:缓存池/MESSAGE_QUEUE:消息队列)',
                                      dataName VARCHAR(100) NOT NULL COMMENT '数据名称（如：线程池名称、指标名称等）',
                                      dataCategory VARCHAR(50) DEFAULT NULL COMMENT '数据分类（如：业务线程池/IO线程池/业务指标/技术指标）',

    -- 监控数据（JSON格式存储，支持不同类型的数据结构）
                                      dataJson LONGTEXT NOT NULL COMMENT '监控数据，JSON格式，包含具体的监控指标和值',

    -- 核心指标（从JSON中提取的关键指标，便于查询和索引）
                                      primaryValue DECIMAL(20,4) DEFAULT NULL COMMENT '主要指标值（如：使用率、数量等）',
                                      secondaryValue DECIMAL(20,4) DEFAULT NULL COMMENT '次要指标值（如：最大值、平均值等）',
                                      statusValue VARCHAR(50) DEFAULT NULL COMMENT '状态值（如：健康状态、连接状态等）',

    -- 健康状态
                                      healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '健康标记(Y健康,N异常)',
                                      healthGrade VARCHAR(20) DEFAULT NULL COMMENT '健康等级(EXCELLENT/GOOD/FAIR/POOR/CRITICAL)',
                                      requiresAttentionFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否需要立即关注(Y是,N否)',

    -- 标签和维度（便于分组查询）
                                      tagsJson TEXT DEFAULT NULL COMMENT '标签信息，JSON格式（如：{"poolType":"business","environment":"prod"}）',

    -- 时间字段
                                      collectionTime DATETIME NOT NULL COMMENT '数据采集时间',

    -- 通用字段
                                      addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                      addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                      editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                      editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                      oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                      currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                      activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                      noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                      PRIMARY KEY (tenantId, appDataId),
                                      KEY IDX_MONITOR_APP_DATA_RES (jvmResourceId),
                                      KEY IDX_MONITOR_APP_DATA_TYPE (dataType),
                                      KEY IDX_MONITOR_APP_DATA_NAME (dataName),
                                      KEY IDX_MONITOR_APP_DATA_TIME (collectionTime),
                                      KEY IDX_MONITOR_APP_DATA_HEALTH (healthyFlag, requiresAttentionFlag),
                                      KEY IDX_MONITOR_APP_DATA_PRIMARY (primaryValue),
                                      KEY IDX_MONITOR_APP_DATA_STATUS (statusValue),
                                      KEY IDX_MONITOR_APP_DATA_RES_TYPE_NAME (jvmResourceId, dataType, dataName, collectionTime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用监控数据表';

