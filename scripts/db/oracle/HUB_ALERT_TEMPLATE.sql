CREATE TABLE HUB_ALERT_TEMPLATE (
  -- 主键和租户
  tenantId VARCHAR2(32) NOT NULL, -- 租户ID，主键
  templateName VARCHAR2(100) NOT NULL, -- 模板名称，主键
  
  -- 模板基本信息
  templateDesc VARCHAR2(500) DEFAULT NULL, -- 模板描述
  channelType VARCHAR2(50) DEFAULT NULL, -- 适用的渠道类型：email/qq/wechat_work/dingtalk/webhook/sms，为空表示通用模板
  
  -- 模板内容
  titleTemplate VARCHAR2(500) DEFAULT NULL, -- 标题模板，支持变量占位符如{{.Title}}
  contentTemplate CLOB DEFAULT NULL, -- 内容模板，支持变量占位符
  displayFormat VARCHAR2(20) DEFAULT 'table' NOT NULL, -- 显示格式：table表格格式/text文本格式
  templateVariables CLOB DEFAULT NULL, -- 模板变量定义，JSON格式，描述可用的变量和说明
  
  -- 附件配置
  attachmentConfig CLOB DEFAULT NULL, -- 附件配置，JSON格式，用于邮件附件等
  
  -- 通用字段
  addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
  addWho VARCHAR2(32) NOT NULL, -- 创建人ID
  editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
  editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
  oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
  currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
  activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记：N非活动，Y活动
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
  
  CONSTRAINT PK_ALERT_TEMPLATE PRIMARY KEY (tenantId, templateName)
);

COMMENT ON TABLE HUB_ALERT_TEMPLATE IS '告警模板表 - 存储告警消息模板，支持变量占位符和多种格式';
COMMENT ON COLUMN HUB_ALERT_TEMPLATE.tenantId IS '租户ID，主键';
COMMENT ON COLUMN HUB_ALERT_TEMPLATE.templateName IS '模板名称，主键';
COMMENT ON COLUMN HUB_ALERT_TEMPLATE.templateDesc IS '模板描述';
COMMENT ON COLUMN HUB_ALERT_TEMPLATE.channelType IS '适用的渠道类型：email/qq/wechat_work/dingtalk/webhook/sms，为空表示通用模板';
COMMENT ON COLUMN HUB_ALERT_TEMPLATE.titleTemplate IS '标题模板，支持变量占位符如{{.Title}}';
COMMENT ON COLUMN HUB_ALERT_TEMPLATE.contentTemplate IS '内容模板，支持变量占位符';
COMMENT ON COLUMN HUB_ALERT_TEMPLATE.displayFormat IS '显示格式：table表格格式/text文本格式';
COMMENT ON COLUMN HUB_ALERT_TEMPLATE.templateVariables IS '模板变量定义，JSON格式，描述可用的变量和说明';
COMMENT ON COLUMN HUB_ALERT_TEMPLATE.attachmentConfig IS '附件配置，JSON格式，用于邮件附件等';

CREATE INDEX IDX_ALERT_TEMPLATE_TENANT ON HUB_ALERT_TEMPLATE (tenantId);
CREATE INDEX IDX_ALERT_TEMPLATE_CHANNEL ON HUB_ALERT_TEMPLATE (channelType);
CREATE INDEX IDX_ALERT_TEMPLATE_ACTIVE ON HUB_ALERT_TEMPLATE (activeFlag);

-- 创建触发器自动更新 editTime
CREATE OR REPLACE TRIGGER TRG_ALERT_TEMPLATE_EDIT_TIME
BEFORE UPDATE ON HUB_ALERT_TEMPLATE
FOR EACH ROW
BEGIN
  :NEW.editTime := SYSDATE;
END;
/

