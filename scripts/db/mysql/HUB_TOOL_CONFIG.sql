CREATE TABLE `HUB_TOOL_CONFIG` (
  `toolConfigId` VARCHAR(32) NOT NULL COMMENT '工具配置ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  
  -- 工具基础信息
  `toolName` VARCHAR(100) NOT NULL COMMENT '工具名称，如SFTP、SSH、FTP等',
  `toolType` VARCHAR(50) NOT NULL COMMENT '工具类型，如transfer、database、monitor等',
  `toolVersion` VARCHAR(20) DEFAULT NULL COMMENT '工具版本号',
  `configName` VARCHAR(100) NOT NULL COMMENT '配置名称，用于区分同一工具的不同配置',
  `configDescription` VARCHAR(500) DEFAULT NULL COMMENT '配置描述信息',
  
  -- 分组信息
  `configGroupId` VARCHAR(32) DEFAULT NULL COMMENT '配置分组ID',
  `configGroupName` VARCHAR(100) DEFAULT NULL COMMENT '配置分组名称',
  
  -- 连接配置
  `hostAddress` VARCHAR(255) DEFAULT NULL COMMENT '主机地址或域名',
  `portNumber` INT DEFAULT NULL COMMENT '端口号',
  `protocolType` VARCHAR(20) DEFAULT NULL COMMENT '协议类型，如TCP、UDP、HTTP等',
  
  -- 认证配置
  `authType` VARCHAR(50) DEFAULT NULL COMMENT '认证类型，如password、publickey、oauth等',
  `userName` VARCHAR(100) DEFAULT NULL COMMENT '用户名',
  `passwordEncrypted` VARCHAR(500) DEFAULT NULL COMMENT '加密后的密码',
  `keyFilePath` VARCHAR(500) DEFAULT NULL COMMENT '密钥文件路径',
  `keyFileContent` TEXT DEFAULT NULL COMMENT '密钥文件内容，加密存储',
  
  -- 配置参数
  `configParameters` TEXT DEFAULT NULL COMMENT '配置参数，JSON格式存储',
  `environmentVariables` TEXT DEFAULT NULL COMMENT '环境变量配置，JSON格式存储',
  `customSettings` TEXT DEFAULT NULL COMMENT '自定义设置，JSON格式存储',
  
  -- 状态和控制
  `configStatus` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '配置状态(N禁用,Y启用)',
  `defaultFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否为默认配置(N否,Y是)',
  `priorityLevel` INT DEFAULT 100 COMMENT '优先级，数值越小优先级越高',
  
  -- 安全和加密
  `encryptionType` VARCHAR(50) DEFAULT NULL COMMENT '加密类型，如AES256、RSA等',
  `encryptionKey` VARCHAR(100) DEFAULT NULL COMMENT '加密密钥标识',
  
  -- 标准字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  
  PRIMARY KEY (`tenantId`, `toolConfigId`),
  KEY `IDX_TOOL_CONFIG_NAME` (`toolName`),
  KEY `IDX_TOOL_CONFIG_TYPE` (`toolType`),
  KEY `IDX_TOOL_CONFIG_CFGNAME` (`configName`),
  KEY `IDX_TOOL_CONFIG_GROUP` (`configGroupId`),
  KEY `IDX_TOOL_CONFIG_STATUS` (`configStatus`),
  KEY `IDX_TOOL_CONFIG_DEFAULT` (`defaultFlag`),
  KEY `IDX_TOOL_CONFIG_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工具配置主表 - 存储各种工具的基础配置信息';
