CREATE TABLE `HUB_GW_PROXY_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `proxyConfigId` VARCHAR(32) NOT NULL COMMENT '代理配置ID',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '网关实例ID(代理配置仅支持实例级)',
  `proxyName` VARCHAR(100) NOT NULL COMMENT '代理名称',
  
  -- 根据ProxyType枚举值设计
  `proxyType` VARCHAR(50) NOT NULL DEFAULT 'http' COMMENT '代理类型(http,websocket,tcp,udp)',
  
  -- 基础配置
  `proxyId` VARCHAR(100) DEFAULT NULL COMMENT '代理ID(来自ProxyConfig.ID)',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  
  -- 通用配置，JSON格式存储不同类型的具体配置
  `proxyConfig` TEXT NOT NULL COMMENT '代理具体配置,JSON格式,根据proxyType存储对应配置',
  `customConfig` TEXT DEFAULT NULL COMMENT '自定义配置,JSON格式',
  
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动/禁用,Y活动/启用)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `proxyConfigId`),
  INDEX `IDX_GW_PROXY_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_PROXY_TYPE` (`proxyType`),
  INDEX `IDX_GW_PROXY_PRIORITY` (`configPriority`),
  INDEX `IDX_GW_PROXY_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='代理配置表 - 根据proxy.go逻辑设计,仅支持实例级代理配置';

