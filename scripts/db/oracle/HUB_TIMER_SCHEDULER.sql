CREATE TABLE HUB_TIMER_SCHEDULER (
                                     schedulerId           VARCHAR2(32) NOT NULL, -- 调度器ID，主键
                                     tenantId              VARCHAR2(32) NOT NULL, -- 租户ID
                                     schedulerName         VARCHAR2(100) NOT NULL, -- 调度器名称
                                     schedulerInstanceId   VARCHAR2(100) NOT NULL, -- 调度器实例ID，用于集群环境区分

                                     maxWorkers            NUMBER(10) DEFAULT 5 NOT NULL, -- 最大工作线程数
                                     queueSize             NUMBER(10) DEFAULT 100 NOT NULL, -- 任务队列大小
                                     defaultTimeoutSeconds NUMBER(20) DEFAULT 1800 NOT NULL, -- 默认超时时间秒数
                                     defaultRetries        NUMBER(10) DEFAULT 3 NOT NULL, -- 默认重试次数

                                     schedulerStatus       NUMBER(10) DEFAULT 1 NOT NULL, -- 调度器状态(1停止,2运行中,3暂停)
                                     lastStartTime         DATE, -- 最后启动时间
                                     lastStopTime          DATE, -- 最后停止时间

                                     serverName            VARCHAR2(100), -- 服务器名称
                                     serverIp              VARCHAR2(50), -- 服务器IP地址
                                     serverPort            NUMBER(10), -- 服务器端口

                                     totalTaskCount        NUMBER(10) DEFAULT 0 NOT NULL, -- 总任务数
                                     runningTaskCount      NUMBER(10) DEFAULT 0 NOT NULL, -- 运行中任务数
                                     lastHeartbeatTime     DATE, -- 最后心跳时间

                                     addTime               DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                     addWho                VARCHAR2(32) NOT NULL, -- 创建人ID
                                     editTime              DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                     editWho               VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                     oprSeqFlag            VARCHAR2(32) NOT NULL, -- 操作序列标识
                                     currentVersion        NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                     activeFlag            VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                     noteText              VARCHAR2(500), -- 备注信息

                                     reserved1             VARCHAR2(500), -- 预留字段1
                                     reserved2             VARCHAR2(500), -- 预留字段2
                                     reserved3             VARCHAR2(500), -- 预留字段3

                                     CONSTRAINT PK_TIMER_SCHEDULER PRIMARY KEY (tenantId, schedulerId)
);
CREATE INDEX IDX_TIMER_SCHED_NAME ON HUB_TIMER_SCHEDULER(schedulerName);
CREATE INDEX IDX_TIMER_SCHED_INST ON HUB_TIMER_SCHEDULER(schedulerInstanceId);
CREATE INDEX IDX_TIMER_SCHED_STATUS ON HUB_TIMER_SCHEDULER(schedulerStatus);
CREATE INDEX IDX_TIMER_SCHED_HEART ON HUB_TIMER_SCHEDULER(lastHeartbeatTime);
COMMENT ON TABLE HUB_TIMER_SCHEDULER IS '定时任务调度器表 - 存储调度器配置和状态信息';