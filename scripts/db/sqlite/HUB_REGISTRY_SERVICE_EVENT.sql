
-- 服务事件日志表 - 记录服务变更事件
CREATE TABLE IF NOT EXISTS HUB_REGISTRY_SERVICE_EVENT (
  -- 主键和租户信息
  serviceEventId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联主键字段（用于精确关联到对应表记录）
  serviceGroupId TEXT,
  serviceInstanceId TEXT,
  
  -- 事件基本信息（冗余字段，便于查询和展示）
  groupName TEXT,
  serviceName TEXT,
  hostAddress TEXT,
  portNumber INTEGER,
  nodeIpAddress TEXT,
  eventType TEXT NOT NULL,
  eventSource TEXT,
  
  -- 事件数据
  eventDataJson TEXT,
  eventMessage TEXT,
  
  -- 时间信息
  eventTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
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
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  
  PRIMARY KEY (tenantId, serviceEventId)
);

-- =====================================================
-- 数据库表结构设计说明
-- =====================================================
-- 
-- 注册类型说明：
-- 1. INTERNAL: 内部管理（默认）- 服务实例直接注册到本系统数据库
-- 2. NACOS: Nacos注册中心 - 服务实例注册到Nacos，本系统作为代理
-- 3. CONSUL: Consul注册中心 - 服务实例注册到Consul，本系统作为代理
-- 4. EUREKA: Eureka注册中心 - 服务实例注册到Eureka，本系统作为代理
-- 5. ETCD: ETCD注册中心 - 服务实例注册到ETCD，本系统作为代理
-- 6. ZOOKEEPER: ZooKeeper注册中心 - 服务实例注册到ZooKeeper，本系统作为代理
--
-- 外部注册中心配置格式（externalRegistryConfig字段JSON示例）：
-- {
--   "serverAddress": "192.168.0.120:8848",
--   "namespace": "ea63c755-3d65-4203-87d7-5ee6837f5bc9",
--   "groupName": "datahub-test-group",
--   "username": "nacos",
--   "password": "nacos",
--   "timeout": 10000,
--   "enableAuth": true,
--   "connectionPool": {
--     "maxConnections": 10,
--     "connectionTimeout": 5000
--   }
-- }
--
-- 使用场景：
-- - registryType = 'INTERNAL': 传统的服务注册，实例信息存储在本地数据库
-- - registryType = 'NACOS': 服务作为Nacos和第三方应用的代理，提供统一的服务发现接口
-- - 其他类型: 类似Nacos，作为对应注册中心的代理
-- =====================================================

-- =====================================================
-- HUB_REGISTRY_SERVICE_EVENT 索引
-- =====================================================
