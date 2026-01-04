CREATE TABLE HUB_GW_INSTANCE (
                                      tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                      gatewayInstanceId VARCHAR2(32) NOT NULL, -- 网关实例ID
                                      instanceName VARCHAR2(100) NOT NULL, -- 实例名称
                                      instanceDesc VARCHAR2(200), -- 实例描述
                                      bindAddress VARCHAR2(100) DEFAULT '0.0.0.0', -- 绑定地址

    -- HTTP/HTTPS 端口配置
                                      httpPort NUMBER(10), -- HTTP监听端口
                                      httpsPort NUMBER(10), -- HTTPS监听端口
                                      tlsEnabled VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否启用TLS(N否,Y是)

    -- 证书配置 - 支持文件路径和数据库存储
                                      certStorageType VARCHAR2(20) DEFAULT 'FILE' NOT NULL, -- 证书存储类型(FILE文件,DATABASE数据库)
                                      certFilePath VARCHAR2(255), -- 证书文件路径
                                      keyFilePath VARCHAR2(255), -- 私钥文件路径
                                      certContent CLOB, -- 证书内容(PEM格式)
                                      keyContent CLOB, -- 私钥内容(PEM格式)
                                      certChainContent CLOB, -- 证书链内容(PEM格式)
                                      certPassword VARCHAR2(255), -- 证书密码(加密存储)

    -- Go HTTP Server 核心配置
                                      maxConnections NUMBER(10) DEFAULT 10000 NOT NULL, -- 最大连接数
                                      readTimeoutMs NUMBER(10) DEFAULT 30000 NOT NULL, -- 读取超时时间(毫秒)
                                      writeTimeoutMs NUMBER(10) DEFAULT 30000 NOT NULL, -- 写入超时时间(毫秒)
                                      idleTimeoutMs NUMBER(10) DEFAULT 60000 NOT NULL, -- 空闲连接超时时间(毫秒)
                                      maxHeaderBytes NUMBER(10) DEFAULT 1048576 NOT NULL, -- 最大请求头字节数(默认1MB)

    -- 性能和并发配置
                                      maxWorkers NUMBER(10) DEFAULT 1000 NOT NULL, -- 最大工作协程数
                                      keepAliveEnabled VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用Keep-Alive(N否,Y是)
                                      tcpKeepAliveEnabled VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用TCP Keep-Alive(N否,Y是)
                                      gracefulShutdownTimeoutMs NUMBER(10) DEFAULT 30000 NOT NULL, -- 优雅关闭超时时间(毫秒)

    -- TLS安全配置
                                      enableHttp2 VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用HTTP/2(N否,Y是)
                                      tlsVersion VARCHAR2(10) DEFAULT '1.2', -- TLS协议版本(1.0,1.1,1.2,1.3)
                                      tlsCipherSuites VARCHAR2(1000), -- TLS密码套件列表,逗号分隔
                                      disableGeneralOptionsHandler VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否禁用默认OPTIONS处理器(N否,Y是)

    -- 日志配置关联字段
                                      logConfigId VARCHAR2(32), -- 关联的日志配置ID
                                      healthStatus VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 健康状态(N不健康,Y健康)
                                      lastHeartbeatTime DATE, -- 最后心跳时间
                                      instanceMetadata CLOB, -- 实例元数据,JSON格式
                                      reserved1 VARCHAR2(100), -- 预留字段1
                                      reserved2 VARCHAR2(100), -- 预留字段2
                                      reserved3 NUMBER(10), -- 预留字段3
                                      reserved4 NUMBER(10), -- 预留字段4
                                      reserved5 DATE, -- 预留字段5
                                      extProperty CLOB, -- 扩展属性,JSON格式
                                      addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                      addWho VARCHAR2(32) NOT NULL, -- 创建人ID
                                      editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                      editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                      oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
                                      currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                      activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                      noteText VARCHAR2(500), -- 备注信息

                                      CONSTRAINT PK_GW_INSTANCE PRIMARY KEY (tenantId, gatewayInstanceId)
);
CREATE INDEX IDX_GW_INST_BIND_HTTP ON HUB_GW_INSTANCE(bindAddress, httpPort);
CREATE INDEX IDX_GW_INST_BIND_HTTPS ON HUB_GW_INSTANCE(bindAddress, httpsPort);
CREATE INDEX IDX_GW_INST_LOG ON HUB_GW_INSTANCE(logConfigId);
CREATE INDEX IDX_GW_INST_HEALTH ON HUB_GW_INSTANCE(healthStatus);
CREATE INDEX IDX_GW_INST_TLS ON HUB_GW_INSTANCE(tlsEnabled);
COMMENT ON TABLE HUB_GW_INSTANCE IS '网关实例表 - 记录网关实例基础配置，完整支持Go HTTP Server配置';