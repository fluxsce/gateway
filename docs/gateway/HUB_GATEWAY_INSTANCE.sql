CREATE TABLE `HUB_GATEWAY_INSTANCE` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '网关实例ID',
    `instanceName` VARCHAR(100) NOT NULL COMMENT '实例名称',
  `instanceDesc` VARCHAR(200) DEFAULT NULL COMMENT '实例描述',
  `bindAddress` VARCHAR(100) DEFAULT '0.0.0.0' COMMENT '绑定地址',

  -- HTTP/HTTPS 端口配置
  `httpPort` INT DEFAULT NULL COMMENT 'HTTP监听端口',
  `httpsPort` INT DEFAULT NULL COMMENT 'HTTPS监听端口',
  `tlsEnabled` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用TLS(N否,Y是)',

  -- 证书配置 - 支持文件路径和数据库存储
  `certStorageType` VARCHAR(20) NOT NULL DEFAULT 'FILE' COMMENT '证书存储类型(FILE文件,DATABASE数据库)',
  `certFilePath` VARCHAR(255) DEFAULT NULL COMMENT '证书文件路径',
  `keyFilePath` VARCHAR(255) DEFAULT NULL COMMENT '私钥文件路径',
  `certContent` TEXT DEFAULT NULL COMMENT '证书内容(PEM格式)',
  `keyContent` TEXT DEFAULT NULL COMMENT '私钥内容(PEM格式)',
  `certChainContent` TEXT DEFAULT NULL COMMENT '证书链内容(PEM格式)',
  `certPassword` VARCHAR(255) DEFAULT NULL COMMENT '证书密码(加密存储)',

  -- Go HTTP Server 核心配置
  `maxConnections` INT NOT NULL DEFAULT 10000 COMMENT '最大连接数',
  `readTimeoutMs` INT NOT NULL DEFAULT 30000 COMMENT '读取超时时间(毫秒)',
  `writeTimeoutMs` INT NOT NULL DEFAULT 30000 COMMENT '写入超时时间(毫秒)',
  `idleTimeoutMs` INT NOT NULL DEFAULT 60000 COMMENT '空闲连接超时时间(毫秒)',
  `maxHeaderBytes` INT NOT NULL DEFAULT 1048576 COMMENT '最大请求头字节数(默认1MB)',

  -- 性能和并发配置
  `maxWorkers` INT NOT NULL DEFAULT 1000 COMMENT '最大工作协程数',
  `keepAliveEnabled` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用Keep-Alive(N否,Y是)',
  `tcpKeepAliveEnabled` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用TCP Keep-Alive(N否,Y是)',
  `gracefulShutdownTimeoutMs` INT NOT NULL DEFAULT 30000 COMMENT '优雅关闭超时时间(毫秒)',

  -- TLS安全配置
  `enableHttp2` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用HTTP/2(N否,Y是)',
  `tlsVersion` VARCHAR(10) DEFAULT '1.2' COMMENT 'TLS协议版本(1.0,1.1,1.2,1.3)',
  `tlsCipherSuites` VARCHAR(1000) DEFAULT NULL COMMENT 'TLS密码套件列表,逗号分隔',
  `disableGeneralOptionsHandler` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否禁用默认OPTIONS处理器(N否,Y是)',
  -- 日志配置关联字段
  `logConfigId` VARCHAR(32) DEFAULT NULL COMMENT '关联的日志配置ID',
  `healthStatus` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '健康状态(N不健康,Y健康)',
  `lastHeartbeatTime` DATETIME DEFAULT NULL COMMENT '最后心跳时间',
  `instanceMetadata` TEXT DEFAULT NULL COMMENT '实例元数据,JSON格式',
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
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_bind_http` (`bindAddress`, `httpPort`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_bind_https` (`bindAddress`, `httpsPort`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_log` (`logConfigId`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_health` (`healthStatus`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_tls` (`tlsEnabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网关实例表 - 记录网关实例基础配置，完整支持Go HTTP Server配置';