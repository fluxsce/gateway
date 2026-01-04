CREATE TABLE HUB_GW_LOG_CONFIG (
                                        tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                        logConfigId VARCHAR2(32) NOT NULL, -- 日志配置ID
                                        configName VARCHAR2(100) NOT NULL, -- 配置名称
                                        configDesc VARCHAR2(200) DEFAULT NULL, -- 配置描述
                                        
                                        -- 日志内容控制
                                        logFormat VARCHAR2(50) DEFAULT 'JSON' NOT NULL, -- 日志格式(JSON,TEXT,CSV)
                                        recordRequestBody VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否记录请求体(N否,Y是)
                                        recordResponseBody VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否记录响应体(N否,Y是)
                                        recordHeaders VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否记录请求/响应头(N否,Y是)
                                        maxBodySizeBytes NUMBER(10) DEFAULT 4096 NOT NULL, -- 最大记录报文大小(字节)
                                        
                                        -- 日志输出目标配置
                                        outputTargets VARCHAR2(200) DEFAULT 'CONSOLE' NOT NULL, -- 输出目标,逗号分隔(CONSOLE,FILE,DATABASE,MONGODB,ELASTICSEARCH)
                                        fileConfig CLOB DEFAULT NULL, -- 文件输出配置,JSON格式
                                        databaseConfig CLOB DEFAULT NULL, -- 数据库输出配置,JSON格式
                                        mongoConfig CLOB DEFAULT NULL, -- MongoDB输出配置,JSON格式
                                        elasticsearchConfig CLOB DEFAULT NULL, -- Elasticsearch输出配置,JSON格式
                                        clickhouseConfig CLOB DEFAULT NULL, -- Clickhouse输出配置,JSON格式
                                        
                                        -- 异步和批量处理配置
                                        enableAsyncLogging VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用异步日志(N否,Y是)
                                        asyncQueueSize NUMBER(10) DEFAULT 10000 NOT NULL, -- 异步队列大小
                                        asyncFlushIntervalMs NUMBER(10) DEFAULT 1000 NOT NULL, -- 异步刷新间隔(毫秒)
                                        enableBatchProcessing VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用批量处理(N否,Y是)
                                        batchSize NUMBER(10) DEFAULT 100 NOT NULL, -- 批处理大小
                                        batchTimeoutMs NUMBER(10) DEFAULT 5000 NOT NULL, -- 批处理超时时间(毫秒)
                                        
                                        -- 日志保留和轮转配置
                                        logRetentionDays NUMBER(10) DEFAULT 30 NOT NULL, -- 日志保留天数
                                        enableFileRotation VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用文件轮转(N否,Y是)
                                        maxFileSizeMB NUMBER(10) DEFAULT 100, -- 最大文件大小(MB)
                                        maxFileCount NUMBER(10) DEFAULT 10, -- 最大文件数量
                                        rotationPattern VARCHAR2(100) DEFAULT 'DAILY', -- 轮转模式(HOURLY,DAILY,WEEKLY,SIZE_BASED)
                                        
                                        -- 敏感数据处理
                                        enableSensitiveDataMasking VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用敏感数据脱敏(N否,Y是)
                                        sensitiveFields CLOB DEFAULT NULL, -- 敏感字段列表,JSON数组格式
                                        maskingPattern VARCHAR2(100) DEFAULT '***', -- 脱敏替换模式
                                        
                                        -- 性能优化配置
                                        bufferSize NUMBER(10) DEFAULT 8192 NOT NULL, -- 缓冲区大小(字节)
                                        flushThreshold NUMBER(10) DEFAULT 100 NOT NULL, -- 刷新阈值(条目数)
                                        
                                        configPriority NUMBER(10) DEFAULT 0 NOT NULL, -- 配置优先级,数值越小优先级越高
                                        reserved1 VARCHAR2(100) DEFAULT NULL, -- 预留字段1
                                        reserved2 VARCHAR2(100) DEFAULT NULL, -- 预留字段2
                                        reserved3 NUMBER(10) DEFAULT NULL, -- 预留字段3
                                        reserved4 NUMBER(10) DEFAULT NULL, -- 预留字段4
                                        reserved5 DATE DEFAULT NULL, -- 预留字段5
                                        extProperty CLOB DEFAULT NULL, -- 扩展属性,JSON格式
                                        addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                        addWho VARCHAR2(32) NOT NULL, -- 创建人ID
                                        editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                        editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                        oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
                                        currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                        activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                        noteText VARCHAR2(500) DEFAULT NULL, -- 备注信息
                                        CONSTRAINT PK_GW_LOG_CONFIG PRIMARY KEY (tenantId, logConfigId)
);

COMMENT ON TABLE HUB_GW_LOG_CONFIG IS '日志配置表 - 存储网关日志相关配置';
CREATE INDEX IDX_GW_LOG_NAME ON HUB_GW_LOG_CONFIG (configName);
CREATE INDEX IDX_GW_LOG_PRIORITY ON HUB_GW_LOG_CONFIG (configPriority);
