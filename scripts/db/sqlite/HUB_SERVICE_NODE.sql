-- 服务节点表 - 存储服务节点的详细信息，包括网络地址、健康状态等
CREATE TABLE IF NOT EXISTS HUB_SERVICE_NODE (
  -- 主键和租户信息
  nodeId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联服务（通过联合主键关联HUB_SERVICE表）
  namespaceId TEXT NOT NULL,
  groupName TEXT NOT NULL,
  serviceName TEXT NOT NULL,
  
  -- 网络连接信息
  ipAddress TEXT NOT NULL,
  portNumber INTEGER NOT NULL,
  
  -- 节点状态信息
  instanceStatus TEXT NOT NULL DEFAULT 'UP',
  healthyStatus TEXT NOT NULL DEFAULT 'UNKNOWN',
  ephemeral TEXT NOT NULL DEFAULT 'Y',
  
  -- 负载均衡配置
  weight REAL NOT NULL DEFAULT 1.00,
  
  -- 节点元数据
  metadataJson TEXT,
  
  -- 时间戳信息
  registerTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  lastBeatTime DATETIME,
  lastCheckTime DATETIME,
  
  -- 通用字段
  addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho TEXT NOT NULL,
  editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  
  PRIMARY KEY (tenantId, nodeId),
  
  UNIQUE (tenantId, namespaceId, groupName, serviceName, ipAddress, portNumber)
);

CREATE INDEX IDX_SVC_NODE_SERVICE ON HUB_SERVICE_NODE(tenantId, namespaceId, groupName, serviceName);
CREATE INDEX IDX_SVC_NODE_SERVICE_NAME ON HUB_SERVICE_NODE(tenantId, serviceName);
CREATE INDEX IDX_SVC_NODE_NS_ID ON HUB_SERVICE_NODE(tenantId, namespaceId);
CREATE INDEX IDX_SVC_NODE_GROUP_NAME ON HUB_SERVICE_NODE(tenantId, namespaceId, groupName);
CREATE INDEX IDX_SVC_NODE_IP_PORT ON HUB_SERVICE_NODE(ipAddress, portNumber);
CREATE INDEX IDX_SVC_NODE_STATUS ON HUB_SERVICE_NODE(instanceStatus);
CREATE INDEX IDX_SVC_NODE_HEALTHY ON HUB_SERVICE_NODE(healthyStatus);
CREATE INDEX IDX_SVC_NODE_EPHEMERAL ON HUB_SERVICE_NODE(ephemeral);
CREATE INDEX IDX_SVC_NODE_ACTIVE ON HUB_SERVICE_NODE(activeFlag);

