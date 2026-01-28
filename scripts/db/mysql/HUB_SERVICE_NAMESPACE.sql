-- 命名空间表 - 存储服务和配置的命名空间信息，用于多租户和多环境隔离
CREATE TABLE `HUB_SERVICE_NAMESPACE` (
  -- 主键和租户信息
  `namespaceId` VARCHAR(32) NOT NULL COMMENT '命名空间ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联服务中心实例
  `instanceName` VARCHAR(100) NOT NULL COMMENT '服务中心实例名称，关联 HUB_SERVICE_CENTER_CONFIG',
  `environment` VARCHAR(32) NOT NULL COMMENT '部署环境（DEVELOPMENT开发,STAGING预发布,PRODUCTION生产）',
  
  -- 命名空间基本信息
  `namespaceName` VARCHAR(100) NOT NULL COMMENT '命名空间名称',
  `namespaceDescription` VARCHAR(500) DEFAULT NULL COMMENT '命名空间描述',
  
  -- 命名空间配置
  `serviceQuotaLimit` INT DEFAULT 200 COMMENT '服务数量配额限制，0表示无限制',
  `configQuotaLimit` INT DEFAULT 200 COMMENT '配置数量配额限制，0表示无限制',
  
  -- 通用字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  
  -- 主键和索引
  PRIMARY KEY (`tenantId`, `namespaceId`),
  KEY `IDX_SVC_NS_NAME` (`tenantId`, `namespaceName`),
  KEY `IDX_SVC_NS_INSTANCE` (`tenantId`, `instanceName`, `environment`),
  KEY `IDX_SVC_NS_ENV` (`environment`),
  KEY `IDX_SVC_NS_ACTIVE` (`activeFlag`),
  CONSTRAINT `FK_NS_INSTANCE_CONFIG` FOREIGN KEY (`tenantId`, `instanceName`, `environment`) 
    REFERENCES `HUB_SERVICE_CENTER_CONFIG` (`tenantId`, `instanceName`, `environment`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='命名空间表 - 存储服务和配置的命名空间信息，用于多租户和多环境隔离';

