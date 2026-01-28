-- 命名空间表 - 存储服务和配置的命名空间信息，用于多租户和多环境隔离
CREATE TABLE HUB_SERVICE_NAMESPACE (
  -- 主键和租户信息
  namespaceId VARCHAR2(32) NOT NULL,
  tenantId VARCHAR2(32) NOT NULL,
  
  -- 关联服务中心实例
  instanceName VARCHAR2(100) NOT NULL,
  environment VARCHAR2(32) NOT NULL,
  
  -- 命名空间基本信息
  namespaceName VARCHAR2(100) NOT NULL,
  namespaceDescription VARCHAR2(500),
  
  -- 命名空间配置
  serviceQuotaLimit NUMBER(10) DEFAULT 200,
  configQuotaLimit NUMBER(10) DEFAULT 200,
  
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
  
  CONSTRAINT PK_SVC_NS PRIMARY KEY (tenantId, namespaceId)
);

CREATE INDEX IDX_SVC_NS_NAME ON HUB_SERVICE_NAMESPACE(tenantId, namespaceName);
CREATE INDEX IDX_SVC_NS_INSTANCE ON HUB_SERVICE_NAMESPACE(tenantId, instanceName, environment);
CREATE INDEX IDX_SVC_NS_ENV ON HUB_SERVICE_NAMESPACE(environment);
CREATE INDEX IDX_SVC_NS_ACTIVE ON HUB_SERVICE_NAMESPACE(activeFlag);

-- 外键约束
ALTER TABLE HUB_SERVICE_NAMESPACE
ADD CONSTRAINT FK_NS_INSTANCE_CONFIG
FOREIGN KEY (tenantId, instanceName, environment)
REFERENCES HUB_SERVICE_CENTER_CONFIG(tenantId, instanceName, environment)
ON DELETE RESTRICT;

COMMENT ON TABLE HUB_SERVICE_NAMESPACE IS '命名空间表 - 存储服务和配置的命名空间信息，用于多租户和多环境隔离';

