-- ===================================================================
-- Gateway 告警系统数据库表结构设计（简化版）
-- 版本: 2.0.0
-- 创建时间: 2025-10-09
-- 说明: 支持 MySQL、Oracle、SQLite 数据库
-- 核心设计: 3张表 - 渠道配置表、模板表、告警记录表
-- ===================================================================

-- ===================================================================
-- 1. 告警渠道配置表 (HUB_ALERT_CHANNEL)
-- 用途: 存储告警渠道的配置信息，如邮件、企业微信、钉钉等
--      渠道可以关联默认模板，也可以有自己的消息格式配置
-- ===================================================================

CREATE TABLE HUB_ALERT_CHANNEL (
    -- 主键和租户ID
    channelId VARCHAR(32) NOT NULL COMMENT '渠道ID，主键',
    tenantId VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
    
    -- 渠道基本信息
    channelName VARCHAR(100) NOT NULL COMMENT '渠道名称',
    channelType VARCHAR(50) NOT NULL COMMENT '渠道类型：email/qq/wechat_work/dingtalk/webhook/sms',
    channelDesc VARCHAR(500) COMMENT '渠道描述',
    enabledFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '启用状态：Y-启用，N-禁用',
    defaultFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否默认渠道：Y-是，N-否',
    priorityLevel INT DEFAULT 5 COMMENT '优先级：1-10，数字越小优先级越高',
    categoryName VARCHAR(50) COMMENT '分类名称：用于分组管理',
    defaultTemplateId VARCHAR(32) COMMENT '默认关联的模板ID（可选）',
    
    -- 服务器配置（JSON格式）
    serverConfig TEXT COMMENT '服务器配置：SMTP配置、Webhook URL等（JSON格式）',
    sendConfig TEXT COMMENT '发送配置：默认收件人、超时设置等（JSON格式）',
    
    -- 消息格式配置（可覆盖模板）
    messageTitlePrefix VARCHAR(50) COMMENT '消息标题前缀，如：【生产环境】',
    messageTitleSuffix VARCHAR(50) COMMENT '消息标题后缀',
    messageContentFormat VARCHAR(20) COMMENT '消息内容格式：text/html/markdown',
    customStyleConfig TEXT COMMENT '自定义样式配置（JSON格式），用于邮件HTML样式等',
    
    -- 重试和超时配置
    timeoutSeconds INT DEFAULT 30 COMMENT '超时时间（秒）',
    retryCount INT DEFAULT 3 COMMENT '重试次数',
    retryIntervalSecs INT DEFAULT 5 COMMENT '重试间隔（秒）',
    asyncSendFlag VARCHAR(1) DEFAULT 'N' COMMENT '异步发送：Y-是，N-否',
    
    -- 限流配置
    rateLimitCount INT DEFAULT 0 COMMENT '限流次数（0表示不限流）',
    rateLimitInterval INT DEFAULT 60 COMMENT '限流时间窗口（秒）',
    
    -- 统计信息
    totalSentCount BIGINT DEFAULT 0 COMMENT '总发送次数',
    successCount BIGINT DEFAULT 0 COMMENT '成功次数',
    failureCount BIGINT DEFAULT 0 COMMENT '失败次数',
    lastSendTime DATETIME COMMENT '最后发送时间',
    lastSuccessTime DATETIME COMMENT '最后成功时间',
    lastFailureTime DATETIME COMMENT '最后失败时间',
    lastErrorMessage VARCHAR(500) COMMENT '最后错误信息',
    avgDurationMillis INT DEFAULT 0 COMMENT '平均耗时（毫秒）',
    
    -- 健康检查
    healthCheckFlag VARCHAR(1) DEFAULT 'Y' COMMENT '健康检查：Y-健康，N-不健康',
    lastHealthCheckTime DATETIME COMMENT '最后健康检查时间',
    healthCheckIntervalSecs INT DEFAULT 300 COMMENT '健康检查间隔（秒）',
    
    -- 通用字段
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho VARCHAR(32) NOT NULL COMMENT '创建人ID',
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
    editWho VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
    oprSeqFlag VARCHAR(32) NOT NULL COMMENT '操作序列标识',
    currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
    noteText VARCHAR(500) COMMENT '备注信息',
    extProperty TEXT COMMENT '扩展属性，JSON格式',
    reserved1 VARCHAR(500) COMMENT '预留字段1',
    reserved2 VARCHAR(500) COMMENT '预留字段2',
    reserved3 VARCHAR(500) COMMENT '预留字段3',
    reserved4 VARCHAR(500) COMMENT '预留字段4',
    reserved5 VARCHAR(500) COMMENT '预留字段5',
    reserved6 VARCHAR(500) COMMENT '预留字段6',
    reserved7 VARCHAR(500) COMMENT '预留字段7',
    reserved8 VARCHAR(500) COMMENT '预留字段8',
    reserved9 VARCHAR(500) COMMENT '预留字段9',
    reserved10 VARCHAR(500) COMMENT '预留字段10',
    
    -- 主键和索引
    CONSTRAINT PK_ALERT_CHANNEL PRIMARY KEY (tenantId, channelId)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警渠道配置表';

-- 创建索引
CREATE INDEX IDX_ALERT_CH_NAME ON HUB_ALERT_CHANNEL(channelName);
CREATE INDEX IDX_ALERT_CH_TYPE ON HUB_ALERT_CHANNEL(channelType);
CREATE INDEX IDX_ALERT_CH_ENABLED ON HUB_ALERT_CHANNEL(enabledFlag);
CREATE INDEX IDX_ALERT_CH_DEFAULT ON HUB_ALERT_CHANNEL(defaultFlag);
CREATE INDEX IDX_ALERT_CH_CATEGORY ON HUB_ALERT_CHANNEL(categoryName);
CREATE INDEX IDX_ALERT_CH_TEMPLATE ON HUB_ALERT_CHANNEL(defaultTemplateId);

-- ===================================================================
-- 2. 告警模板表 (HUB_ALERT_TEMPLATE)
-- 用途: 存储告警模板，包括标题、内容模板和默认渠道配置
--      模板支持变量替换，可被多个告警记录复用
-- ===================================================================

CREATE TABLE HUB_ALERT_TEMPLATE (
    -- 主键和租户ID
    templateId VARCHAR(32) NOT NULL COMMENT '模板ID，主键',
    tenantId VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
    
    -- 模板基本信息
    templateName VARCHAR(100) NOT NULL COMMENT '模板名称',
    templateDesc VARCHAR(500) COMMENT '模板描述',
    templateType VARCHAR(50) NOT NULL COMMENT '模板类型：threshold/anomaly/status/business/custom',
    severityLevel VARCHAR(20) NOT NULL COMMENT '严重级别：critical/high/medium/low/info',
    enabledFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '启用状态：Y-启用，N-禁用',
    categoryName VARCHAR(50) COMMENT '分类名称：system/service/api/business等',
    ownerUserId VARCHAR(32) COMMENT '负责人用户ID',
    
    -- 告警内容模板
    titleTemplate VARCHAR(200) NOT NULL COMMENT '标题模板，支持变量如：{{service}} CPU使用率告警',
    contentTemplate TEXT NOT NULL COMMENT '内容模板，支持变量如：服务器 {{host}} CPU使用率达到 {{value}}%',
    tagsTemplate VARCHAR(500) COMMENT '标签模板（JSON格式）',
    
    -- 默认通知渠道配置（可在发送时覆盖）
    defaultChannelIds VARCHAR(500) COMMENT '默认告警渠道ID列表（逗号分隔），发送时可覆盖',
    notifyRecipients TEXT COMMENT '默认收件人配置（JSON格式），发送时可覆盖',
    sendConfig TEXT COMMENT '默认发送配置（JSON格式：超时、重试等），发送时可覆盖',
    
    -- 触发条件（可选，用于自动触发）
    triggerCondition TEXT COMMENT '触发条件（JSON格式）',
    triggerDurationSecs INT DEFAULT 0 COMMENT '持续时间（秒），0表示立即触发',
    checkIntervalSecs INT DEFAULT 0 COMMENT '检查间隔（秒），0表示手动触发',
    autoTriggerFlag VARCHAR(1) DEFAULT 'N' COMMENT '自动触发：Y-自动，N-手动',
    monitorTarget VARCHAR(200) COMMENT '监控目标（用于自动触发）',
    metricName VARCHAR(100) COMMENT '指标名称（用于自动触发）',
    
    -- 静默和抑制
    silenceFlag VARCHAR(1) DEFAULT 'N' COMMENT '静默标记：Y-静默，N-正常',
    silenceStartTime DATETIME COMMENT '静默开始时间',
    silenceEndTime DATETIME COMMENT '静默结束时间',
    silenceReason VARCHAR(500) COMMENT '静默原因',
    repeatIntervalSecs INT DEFAULT 3600 COMMENT '重复通知间隔（秒）',
    deduplicationFlag VARCHAR(1) DEFAULT 'Y' COMMENT '去重标记：Y-启用去重，N-不去重',
    deduplicationWindowSecs INT DEFAULT 300 COMMENT '去重时间窗口（秒）',
    
    -- 统计信息
    usageCount BIGINT DEFAULT 0 COMMENT '使用次数',
    lastUsedTime DATETIME COMMENT '最后使用时间',
    totalAlertCount BIGINT DEFAULT 0 COMMENT '总告警次数',
    lastAlertTime DATETIME COMMENT '最后告警时间',
    
    -- 通用字段
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho VARCHAR(32) NOT NULL COMMENT '创建人ID',
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
    editWho VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
    oprSeqFlag VARCHAR(32) NOT NULL COMMENT '操作序列标识',
    currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
    noteText VARCHAR(500) COMMENT '备注信息',
    extProperty TEXT COMMENT '扩展属性，JSON格式',
    reserved1 VARCHAR(500) COMMENT '预留字段1',
    reserved2 VARCHAR(500) COMMENT '预留字段2',
    reserved3 VARCHAR(500) COMMENT '预留字段3',
    reserved4 VARCHAR(500) COMMENT '预留字段4',
    reserved5 VARCHAR(500) COMMENT '预留字段5',
    reserved6 VARCHAR(500) COMMENT '预留字段6',
    reserved7 VARCHAR(500) COMMENT '预留字段7',
    reserved8 VARCHAR(500) COMMENT '预留字段8',
    reserved9 VARCHAR(500) COMMENT '预留字段9',
    reserved10 VARCHAR(500) COMMENT '预留字段10',
    
    -- 主键和索引
    CONSTRAINT PK_ALERT_TEMPLATE PRIMARY KEY (tenantId, templateId)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警模板表';

-- 创建索引
CREATE INDEX IDX_ALERT_TPL_NAME ON HUB_ALERT_TEMPLATE(templateName);
CREATE INDEX IDX_ALERT_TPL_TYPE ON HUB_ALERT_TEMPLATE(templateType);
CREATE INDEX IDX_ALERT_TPL_SEVERITY ON HUB_ALERT_TEMPLATE(severityLevel);
CREATE INDEX IDX_ALERT_TPL_ENABLED ON HUB_ALERT_TEMPLATE(enabledFlag);
CREATE INDEX IDX_ALERT_TPL_CATEGORY ON HUB_ALERT_TEMPLATE(categoryName);
CREATE INDEX IDX_ALERT_TPL_AUTO ON HUB_ALERT_TEMPLATE(autoTriggerFlag);

-- ===================================================================
-- 3. 告警记录表 (HUB_ALERT_RECORD)
-- 用途: 存储每次告警的完整记录，包括触发、通知、处理等全生命周期信息
--      整合了原来的告警记录和通知日志，避免数据冗余
-- ===================================================================

CREATE TABLE HUB_ALERT_RECORD (
    -- 主键和租户ID
    alertRecordId VARCHAR(32) NOT NULL COMMENT '告警记录ID，主键',
    tenantId VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
    
    -- 关联信息
    channelIds VARCHAR(500) NOT NULL COMMENT '使用的渠道ID列表（逗号分隔，支持多渠道）',
    channelNames VARCHAR(500) COMMENT '渠道名称列表（冗余字段，逗号分隔）',
    
    -- 告警基本信息
    alertTitle VARCHAR(200) NOT NULL COMMENT '告警标题',
    alertContent TEXT NOT NULL COMMENT '告警内容',
    alertType VARCHAR(50) NOT NULL COMMENT '告警类型',
    severityLevel VARCHAR(20) NOT NULL COMMENT '严重级别',
    categoryName VARCHAR(50) COMMENT '分类名称',
    alertStatus VARCHAR(20) NOT NULL DEFAULT 'open' COMMENT '告警状态：open/acknowledged/resolved/closed',
    alertTags TEXT COMMENT '告警标签（JSON格式）',
    
    -- 触发信息
    triggerTime DATETIME NOT NULL COMMENT '触发时间',
    triggerSource VARCHAR(50) NOT NULL COMMENT '触发来源：auto/manual/api/schedule/event',
    triggerValue VARCHAR(200) COMMENT '触发值',
    triggerCondition TEXT COMMENT '触发条件（快照）',
    monitorTarget VARCHAR(200) COMMENT '监控目标',
    metricName VARCHAR(100) COMMENT '指标名称',
    metricValue VARCHAR(100) COMMENT '指标值',
    metricUnit VARCHAR(20) COMMENT '指标单位',
    sourceSystem VARCHAR(100) COMMENT '来源系统',
    sourceHost VARCHAR(100) COMMENT '来源主机',
    
    -- 恢复信息
    recoveryTime DATETIME COMMENT '恢复时间',
    recoveryValue VARCHAR(200) COMMENT '恢复时的值',
    durationSecs INT COMMENT '持续时间（秒）',
    
    -- 通知信息
    notifyTarget VARCHAR(500) COMMENT '通知目标：收件人、手机号等',
    notifyStatus VARCHAR(20) DEFAULT 'pending' COMMENT '通知状态：pending/sending/sent/failed/partial',
    notifySendTime DATETIME COMMENT '通知发送时间',
    notifyCompleteTime DATETIME COMMENT '通知完成时间',
    notifyDurationMillis INT COMMENT '通知耗时（毫秒）',
    notifyRetryCount INT DEFAULT 0 COMMENT '通知重试次数',
    notifyErrorMsg VARCHAR(500) COMMENT '通知错误信息',
    notifyErrorCode VARCHAR(50) COMMENT '通知错误码',
    messageId VARCHAR(100) COMMENT '消息ID（渠道返回）',
    responseData TEXT COMMENT '响应数据（JSON格式）',
    
    -- 处理信息
    ackFlag VARCHAR(1) DEFAULT 'N' COMMENT '确认标记：Y-已确认，N-未确认',
    ackTime DATETIME COMMENT '确认时间',
    ackUserId VARCHAR(32) COMMENT '确认人用户ID',
    ackUserName VARCHAR(100) COMMENT '确认人姓名',
    ackComment VARCHAR(500) COMMENT '确认备注',
    resolveFlag VARCHAR(1) DEFAULT 'N' COMMENT '解决标记：Y-已解决，N-未解决',
    resolveTime DATETIME COMMENT '解决时间',
    resolveUserId VARCHAR(32) COMMENT '解决人用户ID',
    resolveUserName VARCHAR(100) COMMENT '解决人姓名',
    resolveComment VARCHAR(500) COMMENT '解决备注',
    resolveDurationSecs INT COMMENT '解决耗时（秒）',
    closeTime DATETIME COMMENT '关闭时间',
    closeUserId VARCHAR(32) COMMENT '关闭人用户ID',
    closeUserName VARCHAR(100) COMMENT '关闭人姓名',
    
    -- 元数据和追踪
    alertMetadata TEXT COMMENT '告警元数据（JSON格式）',
    relatedRecordIds VARCHAR(500) COMMENT '关联的记录ID（逗号分隔）',
    requestId VARCHAR(100) COMMENT '请求ID（用于追踪）',
    traceId VARCHAR(100) COMMENT '追踪ID',
    batchId VARCHAR(32) COMMENT '批次ID（同一批发送）',
    sequenceNum INT COMMENT '序列号',
    
    -- 统计字段
    viewCount INT DEFAULT 0 COMMENT '查看次数',
    lastViewTime DATETIME COMMENT '最后查看时间',
    commentCount INT DEFAULT 0 COMMENT '评论数',
    escalateFlag VARCHAR(1) DEFAULT 'N' COMMENT '升级标记：Y-已升级，N-未升级',
    escalateTime DATETIME COMMENT '升级时间',
    escalateLevel INT DEFAULT 0 COMMENT '升级级别',
    
    -- 通用字段
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho VARCHAR(32) NOT NULL COMMENT '创建人ID',
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
    editWho VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
    oprSeqFlag VARCHAR(32) NOT NULL COMMENT '操作序列标识',
    currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
    noteText VARCHAR(500) COMMENT '备注信息',
    extProperty TEXT COMMENT '扩展属性，JSON格式',
    reserved1 VARCHAR(500) COMMENT '预留字段1',
    reserved2 VARCHAR(500) COMMENT '预留字段2',
    reserved3 VARCHAR(500) COMMENT '预留字段3',
    reserved4 VARCHAR(500) COMMENT '预留字段4',
    reserved5 VARCHAR(500) COMMENT '预留字段5',
    reserved6 VARCHAR(500) COMMENT '预留字段6',
    reserved7 VARCHAR(500) COMMENT '预留字段7',
    reserved8 VARCHAR(500) COMMENT '预留字段8',
    reserved9 VARCHAR(500) COMMENT '预留字段9',
    reserved10 VARCHAR(500) COMMENT '预留字段10',
    
    -- 主键和索引
    CONSTRAINT PK_ALERT_RECORD PRIMARY KEY (tenantId, alertRecordId)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警记录表';

-- 创建索引
CREATE INDEX IDX_ALERT_REC_STATUS ON HUB_ALERT_RECORD(alertStatus);
CREATE INDEX IDX_ALERT_REC_SEVERITY ON HUB_ALERT_RECORD(severityLevel);
CREATE INDEX IDX_ALERT_REC_TIME ON HUB_ALERT_RECORD(triggerTime);
CREATE INDEX IDX_ALERT_REC_NOTIFY ON HUB_ALERT_RECORD(notifyStatus);
CREATE INDEX IDX_ALERT_REC_ACK ON HUB_ALERT_RECORD(ackFlag);
CREATE INDEX IDX_ALERT_REC_RESOLVE ON HUB_ALERT_RECORD(resolveFlag);
CREATE INDEX IDX_ALERT_REC_TARGET ON HUB_ALERT_RECORD(monitorTarget);
CREATE INDEX IDX_ALERT_REC_SOURCE ON HUB_ALERT_REC(triggerSource);
CREATE INDEX IDX_ALERT_REC_CATEGORY ON HUB_ALERT_RECORD(categoryName);
CREATE INDEX IDX_ALERT_REC_TRACE ON HUB_ALERT_RECORD(traceId);
CREATE INDEX IDX_ALERT_REC_BATCH ON HUB_ALERT_RECORD(batchId);

-- ===================================================================
-- 说明：
-- 1. 渠道配置表：独立管理各类告警渠道，可关联默认模板
-- 2. 模板表：定义告警内容模板和默认渠道，支持变量替换
-- 3. 告警记录表：存储所有告警数据，包括通知日志信息
-- 
-- 使用流程：
-- 1. 配置渠道 -> 创建模板(关联渠道) -> 发送告警(选择模板和渠道)
-- 2. 或者：配置渠道 -> 直接发送告警(不使用模板)
-- ===================================================================
