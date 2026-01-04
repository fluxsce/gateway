CREATE TABLE HUB_GW_ROUTE_CONFIG (
  tenantId VARCHAR2(32) NOT NULL, -- 租户ID
  routeConfigId VARCHAR2(32) NOT NULL, -- 路由配置ID
  gatewayInstanceId VARCHAR2(32) NOT NULL, -- 关联的网关实例ID
  routeName VARCHAR2(100) NOT NULL, -- 路由名称
  routePath VARCHAR2(200) NOT NULL, -- 路由路径
  allowedMethods VARCHAR2(200), -- 允许的HTTP方法,JSON数组格式["GET","POST"]
  allowedHosts VARCHAR2(500), -- 允许的域名,逗号分隔
  matchType NUMBER(10) DEFAULT 1 NOT NULL, -- 匹配类型(0精确匹配,1前缀匹配,2正则匹配)
  routePriority NUMBER(10) DEFAULT 100 NOT NULL, -- 路由优先级,数值越小优先级越高
  stripPathPrefix VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否剥离路径前缀(N否,Y是)
  rewritePath VARCHAR2(200), -- 重写路径
  enableWebsocket VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否支持WebSocket(N否,Y是)
  timeoutMs NUMBER(10) DEFAULT 30000 NOT NULL, -- 超时时间(毫秒)
  retryCount NUMBER(10) DEFAULT 0 NOT NULL, -- 重试次数
  retryIntervalMs NUMBER(10) DEFAULT 1000 NOT NULL, -- 重试间隔(毫秒)

  -- 服务关联字段，直接关联服务定义表
  serviceDefinitionId VARCHAR2(32), -- 关联的服务定义ID

  -- 日志配置关联字段
  logConfigId VARCHAR2(32), -- 关联的日志配置ID(路由级日志配置)

  -- 路由元数据，用于存储额外配置信息
  routeMetadata CLOB, -- 路由元数据,JSON格式,存储Methods等配置

  -- 预留字段
  reserved1 VARCHAR2(100), -- 预留字段1
  reserved2 VARCHAR2(100), -- 预留字段2
  reserved3 NUMBER(10), -- 预留字段3
  reserved4 NUMBER(10), -- 预留字段4
  reserved5 DATE, -- 预留字段5
  extProperty CLOB, -- 扩展属性,JSON格式

  -- 标准字段
  addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
  addWho VARCHAR2(32) NOT NULL, -- 创建人ID
  editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
  editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
  oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
  currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
  activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动/禁用,Y活动/启用)
  noteText VARCHAR2(500), -- 备注信息

  CONSTRAINT PK_GW_ROUTE_CONFIG PRIMARY KEY (tenantId, routeConfigId)
);
CREATE INDEX IDX_GW_ROUTE_INST ON HUB_GW_ROUTE_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_ROUTE_SERVICE ON HUB_GW_ROUTE_CONFIG(serviceDefinitionId);
CREATE INDEX IDX_GW_ROUTE_LOG ON HUB_GW_ROUTE_CONFIG(logConfigId);
CREATE INDEX IDX_GW_ROUTE_PRIORITY ON HUB_GW_ROUTE_CONFIG(routePriority);
CREATE INDEX IDX_GW_ROUTE_PATH ON HUB_GW_ROUTE_CONFIG(routePath);
CREATE INDEX IDX_GW_ROUTE_ACTIVE ON HUB_GW_ROUTE_CONFIG(activeFlag);
COMMENT ON TABLE HUB_GW_ROUTE_CONFIG IS '路由定义表 - 存储API路由配置,使用activeFlag统一管理启用状态';

-- 兼容已存在的数据库，修改字段长度
ALTER TABLE HUB_GW_ROUTE_CONFIG MODIFY serviceDefinitionId VARCHAR2(1000);
