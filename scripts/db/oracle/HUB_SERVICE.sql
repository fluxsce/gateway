-- 服务表 - 存储服务的基本信息和元数据
CREATE TABLE HUB_SERVICE (
  -- 主键和租户信息
  tenantId VARCHAR2(32) NOT NULL,
  namespaceId VARCHAR2(32) NOT NULL,
  groupName VARCHAR2(64) DEFAULT 'DEFAULT_GROUP' NOT NULL,
  serviceName VARCHAR2(100) NOT NULL,
  
  -- 服务类型
  serviceType VARCHAR2(50) DEFAULT 'INTERNAL' NOT NULL,
  
  -- 服务基本信息
  serviceVersion VARCHAR2(50),
  serviceDescription VARCHAR2(500),
  
  -- 外部服务配置（仅当serviceType非INTERNAL时使用）
  externalServiceConfig CLOB,
  
  -- 服务元数据
  metadataJson CLOB,
  tagsJson CLOB,
  
  -- 服务保护阈值（0-1之间的小数，表示健康实例比例低于该值时触发保护）
  protectThreshold NUMBER(3,2) DEFAULT 0.00,
  
  -- 服务选择器（用于服务路由）
  selectorJson CLOB,
  
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
  
  -- 预留字段
  reserved1 VARCHAR2(500),
  reserved2 VARCHAR2(500),
  reserved3 VARCHAR2(500),
  reserved4 VARCHAR2(500),
  reserved5 VARCHAR2(500),
  reserved6 VARCHAR2(500),
  reserved7 VARCHAR2(500),
  reserved8 VARCHAR2(500),
  reserved9 VARCHAR2(500),
  reserved10 VARCHAR2(500),
  
  CONSTRAINT PK_SVC PRIMARY KEY (tenantId, namespaceId, groupName, serviceName)
);

CREATE INDEX IDX_SVC_NS_ID ON HUB_SERVICE(tenantId, namespaceId);
CREATE INDEX IDX_SVC_NAME ON HUB_SERVICE(serviceName);
CREATE INDEX IDX_SVC_TYPE ON HUB_SERVICE(serviceType);
CREATE INDEX IDX_SVC_ACTIVE ON HUB_SERVICE(activeFlag);

COMMENT ON TABLE HUB_SERVICE IS '服务表 - 存储服务的基本信息和元数据，支持多命名空间和多分组管理';

