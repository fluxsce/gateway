-- 服务节点表 - 存储服务节点的详细信息，包括网络地址、健康状态等
CREATE TABLE HUB_SERVICE_NODE (
  -- 主键和租户信息
  nodeId VARCHAR2(32) NOT NULL,
  tenantId VARCHAR2(32) NOT NULL,
  
  -- 关联服务（通过联合主键关联HUB_SERVICE表）
  namespaceId VARCHAR2(32) NOT NULL,
  groupName VARCHAR2(64) NOT NULL,
  serviceName VARCHAR2(100) NOT NULL,
  
  -- 网络连接信息
  ipAddress VARCHAR2(50) NOT NULL,
  portNumber NUMBER(10) NOT NULL,
  
  -- 节点状态信息
  instanceStatus VARCHAR2(20) DEFAULT 'UP' NOT NULL,
  healthyStatus VARCHAR2(20) DEFAULT 'UNKNOWN' NOT NULL,
  ephemeral VARCHAR2(1) DEFAULT 'Y' NOT NULL,
  
  -- 负载均衡配置
  weight NUMBER(6,2) DEFAULT 1.00 NOT NULL,
  
  -- 节点元数据
  metadataJson CLOB,
  
  -- 时间戳信息
  registerTime DATE DEFAULT SYSDATE NOT NULL,
  lastBeatTime DATE,
  lastCheckTime DATE,
  
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
  
  CONSTRAINT PK_SVC_NODE PRIMARY KEY (tenantId, nodeId)
);

CREATE UNIQUE INDEX IDX_SVC_NODE_UNIQUE ON HUB_SERVICE_NODE(tenantId, namespaceId, groupName, serviceName, ipAddress, portNumber);
CREATE INDEX IDX_SVC_NODE_SERVICE ON HUB_SERVICE_NODE(tenantId, namespaceId, groupName, serviceName);
CREATE INDEX IDX_SVC_NODE_SERVICE_NAME ON HUB_SERVICE_NODE(tenantId, serviceName);
CREATE INDEX IDX_SVC_NODE_NS_ID ON HUB_SERVICE_NODE(tenantId, namespaceId);
CREATE INDEX IDX_SVC_NODE_GROUP_NAME ON HUB_SERVICE_NODE(tenantId, namespaceId, groupName);
CREATE INDEX IDX_SVC_NODE_IP_PORT ON HUB_SERVICE_NODE(ipAddress, portNumber);
CREATE INDEX IDX_SVC_NODE_STATUS ON HUB_SERVICE_NODE(instanceStatus);
CREATE INDEX IDX_SVC_NODE_HEALTHY ON HUB_SERVICE_NODE(healthyStatus);
CREATE INDEX IDX_SVC_NODE_EPHEMERAL ON HUB_SERVICE_NODE(ephemeral);
CREATE INDEX IDX_SVC_NODE_ACTIVE ON HUB_SERVICE_NODE(activeFlag);

COMMENT ON TABLE HUB_SERVICE_NODE IS '服务节点表 - 存储服务节点的详细信息，包括网络地址、健康状态、权重等';

