CREATE TABLE HUB_ALERT_CONFIG (
  -- 主键和租户
  tenantId VARCHAR2(32) NOT NULL, -- 租户ID，主键
  channelName VARCHAR2(100) NOT NULL, -- 渠道名称，主键
  
  -- 渠道基本信息
  channelType VARCHAR2(50) NOT NULL, -- 渠道类型：email/qq/wechat_work/dingtalk/webhook/sms
  channelDesc VARCHAR2(500) DEFAULT NULL, -- 渠道描述
  activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 启用状态：Y-启用，N-禁用
  defaultFlag VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否默认渠道：Y-是，N-否
  priorityLevel NUMBER(10) DEFAULT 10 NOT NULL, -- 优先级：1-10，数字越小优先级越高
  defaultTemplateName VARCHAR2(100) DEFAULT NULL, -- 默认关联的模板名称
  
  -- 服务器配置（JSON格式）
  serverConfig CLOB DEFAULT NULL, -- 服务器配置，JSON格式，如SMTP配置、Webhook URL等
  sendConfig CLOB DEFAULT NULL, -- 发送配置，JSON格式，如默认收件人、超时设置等
  
  -- 消息格式配置
  messageContentFormat VARCHAR2(20) DEFAULT NULL, -- 消息内容格式：text/html/markdown
  
  -- 重试和超时配置
  timeoutSeconds NUMBER(10) DEFAULT 30 NOT NULL, -- 超时时间（秒）
  retryCount NUMBER(10) DEFAULT 3 NOT NULL, -- 重试次数
  retryIntervalSecs NUMBER(10) DEFAULT 5 NOT NULL, -- 重试间隔（秒）
  asyncSendFlag VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否异步发送：Y-是，N-否
  
  -- 统计信息
  totalSentCount NUMBER(19) DEFAULT 0 NOT NULL, -- 总发送次数
  successCount NUMBER(19) DEFAULT 0 NOT NULL, -- 成功次数
  failureCount NUMBER(19) DEFAULT 0 NOT NULL, -- 失败次数
  lastSendTime DATE DEFAULT NULL, -- 最后发送时间
  lastSuccessTime DATE DEFAULT NULL, -- 最后成功时间
  lastFailureTime DATE DEFAULT NULL, -- 最后失败时间
  lastErrorMessage VARCHAR2(1000) DEFAULT NULL, -- 最后错误信息
  
  -- 通用字段
  addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
  addWho VARCHAR2(32) NOT NULL, -- 创建人ID
  editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
  editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
  oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
  currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
  noteText VARCHAR2(500) DEFAULT NULL, -- 备注信息
  extProperty CLOB DEFAULT NULL, -- 扩展属性，JSON格式
  
  -- 预留字段
  reserved1 VARCHAR2(500) DEFAULT NULL, -- 预留字段1
  reserved2 VARCHAR2(500) DEFAULT NULL, -- 预留字段2
  reserved3 VARCHAR2(500) DEFAULT NULL, -- 预留字段3
  reserved4 VARCHAR2(500) DEFAULT NULL, -- 预留字段4
  reserved5 VARCHAR2(500) DEFAULT NULL, -- 预留字段5
  reserved6 VARCHAR2(500) DEFAULT NULL, -- 预留字段6
  reserved7 VARCHAR2(500) DEFAULT NULL, -- 预留字段7
  reserved8 VARCHAR2(500) DEFAULT NULL, -- 预留字段8
  reserved9 VARCHAR2(500) DEFAULT NULL, -- 预留字段9
  reserved10 VARCHAR2(500) DEFAULT NULL, -- 预留字段10
  
  CONSTRAINT PK_ALERT_CONFIG PRIMARY KEY (tenantId, channelName)
);

COMMENT ON TABLE HUB_ALERT_CONFIG IS '告警渠道配置表 - 存储多渠道告警配置信息';
COMMENT ON COLUMN HUB_ALERT_CONFIG.tenantId IS '租户ID，主键';
COMMENT ON COLUMN HUB_ALERT_CONFIG.channelName IS '渠道名称，主键';
COMMENT ON COLUMN HUB_ALERT_CONFIG.channelType IS '渠道类型：email/qq/wechat_work/dingtalk/webhook/sms';
COMMENT ON COLUMN HUB_ALERT_CONFIG.channelDesc IS '渠道描述';
COMMENT ON COLUMN HUB_ALERT_CONFIG.activeFlag IS '启用状态：Y-启用，N-禁用';
COMMENT ON COLUMN HUB_ALERT_CONFIG.defaultFlag IS '是否默认渠道：Y-是，N-否';
COMMENT ON COLUMN HUB_ALERT_CONFIG.priorityLevel IS '优先级：1-10，数字越小优先级越高';
COMMENT ON COLUMN HUB_ALERT_CONFIG.defaultTemplateName IS '默认关联的模板名称';
COMMENT ON COLUMN HUB_ALERT_CONFIG.serverConfig IS '服务器配置，JSON格式，如SMTP配置、Webhook URL等';
COMMENT ON COLUMN HUB_ALERT_CONFIG.sendConfig IS '发送配置，JSON格式，如默认收件人、超时设置等';
COMMENT ON COLUMN HUB_ALERT_CONFIG.messageContentFormat IS '消息内容格式：text/html/markdown';
COMMENT ON COLUMN HUB_ALERT_CONFIG.timeoutSeconds IS '超时时间（秒）';
COMMENT ON COLUMN HUB_ALERT_CONFIG.retryCount IS '重试次数';
COMMENT ON COLUMN HUB_ALERT_CONFIG.retryIntervalSecs IS '重试间隔（秒）';
COMMENT ON COLUMN HUB_ALERT_CONFIG.asyncSendFlag IS '是否异步发送：Y-是，N-否';
COMMENT ON COLUMN HUB_ALERT_CONFIG.totalSentCount IS '总发送次数';
COMMENT ON COLUMN HUB_ALERT_CONFIG.successCount IS '成功次数';
COMMENT ON COLUMN HUB_ALERT_CONFIG.failureCount IS '失败次数';
COMMENT ON COLUMN HUB_ALERT_CONFIG.lastSendTime IS '最后发送时间';
COMMENT ON COLUMN HUB_ALERT_CONFIG.lastSuccessTime IS '最后成功时间';
COMMENT ON COLUMN HUB_ALERT_CONFIG.lastFailureTime IS '最后失败时间';
COMMENT ON COLUMN HUB_ALERT_CONFIG.lastErrorMessage IS '最后错误信息';

CREATE INDEX IDX_ALERT_CONFIG_TENANT ON HUB_ALERT_CONFIG (tenantId);
CREATE INDEX IDX_ALERT_CONFIG_TYPE ON HUB_ALERT_CONFIG (channelType);
CREATE INDEX IDX_ALERT_CONFIG_ACTIVE ON HUB_ALERT_CONFIG (activeFlag);
CREATE INDEX IDX_ALERT_CONFIG_DEFAULT ON HUB_ALERT_CONFIG (defaultFlag);
CREATE INDEX IDX_ALERT_CONFIG_PRIORITY ON HUB_ALERT_CONFIG (priorityLevel);
CREATE INDEX IDX_ALERT_CONFIG_TEMPLATE ON HUB_ALERT_CONFIG (defaultTemplateName);

-- 创建触发器自动更新 editTime
CREATE OR REPLACE TRIGGER TRG_ALERT_CONFIG_EDIT_TIME
BEFORE UPDATE ON HUB_ALERT_CONFIG
FOR EACH ROW
BEGIN
  :NEW.editTime := SYSDATE;
END;
/

