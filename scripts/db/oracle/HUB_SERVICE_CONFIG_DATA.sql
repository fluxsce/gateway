-- 配置数据表 - 存储具体的配置数据，支持多种配置格式
CREATE TABLE HUB_SERVICE_CONFIG_DATA (
  -- 主键和租户信息
  configDataId VARCHAR2(100) NOT NULL,
  tenantId VARCHAR2(32) NOT NULL,
  
  -- 关联命名空间和分组
  namespaceId VARCHAR2(32) NOT NULL,
  groupName VARCHAR2(64) DEFAULT 'DEFAULT_GROUP' NOT NULL,
  
  -- 配置基本信息
  configContent CLOB NOT NULL,
  contentType VARCHAR2(50) DEFAULT 'text',
  
  -- 配置描述和属性
  configDescription VARCHAR2(500),
  encrypted VARCHAR2(1) DEFAULT 'N',
  
  -- 版本信息
  version NUMBER(19) DEFAULT 1 NOT NULL,
  
  -- MD5校验值（用于配置变更检测）
  md5Value VARCHAR2(32),
  
  -- 通用字段
  addTime DATE DEFAULT SYSDATE NOT NULL,
  addWho VARCHAR2(32) NOT NULL,
  editTime DATE DEFAULT SYSDATE NOT NULL,
  editWho VARCHAR2(32) NOT NULL,
  oprSeqFlag VARCHAR2(32) NOT NULL,
  currentVersion NUMBER(10) DEFAULT 1 NOT NULL,
  activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL,
  noteText VARCHAR2(500),
  extProperty CLOB,
  
  CONSTRAINT PK_SVC_CFG_DATA PRIMARY KEY (tenantId, namespaceId, groupName, configDataId)
);

CREATE INDEX IDX_SVC_CFG_DATA_NS_ID ON HUB_SERVICE_CONFIG_DATA(tenantId, namespaceId);
CREATE INDEX IDX_SVC_CFG_DATA_MD5 ON HUB_SERVICE_CONFIG_DATA(md5Value);
CREATE INDEX IDX_SVC_CFG_DATA_ACTIVE ON HUB_SERVICE_CONFIG_DATA(activeFlag);

COMMENT ON TABLE HUB_SERVICE_CONFIG_DATA IS '配置数据表 - 存储具体的配置数据，支持多种配置格式和版本管理';

