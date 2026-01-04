CREATE TABLE HUB_TOOL_CONFIG (
                                 toolConfigId      VARCHAR2(32)   NOT NULL,
                                 tenantId          VARCHAR2(32)   NOT NULL,

    -- 工具基础信息
                                 toolName          VARCHAR2(100)  NOT NULL,
                                 toolType          VARCHAR2(50)   NOT NULL,
                                 toolVersion       VARCHAR2(20),
                                 configName        VARCHAR2(100)  NOT NULL,
                                 configDescription VARCHAR2(500),

    -- 分组信息
                                 configGroupId     VARCHAR2(32),
                                 configGroupName   VARCHAR2(100),

    -- 连接配置
                                 hostAddress       VARCHAR2(255),
                                 portNumber        NUMBER(10),
                                 protocolType      VARCHAR2(20),

    -- 认证配置
                                 authType          VARCHAR2(50),
                                 userName          VARCHAR2(100),
                                 passwordEncrypted VARCHAR2(500),
                                 keyFilePath       VARCHAR2(500),
                                 keyFileContent    CLOB,

    -- 配置参数
                                 configParameters  CLOB,
                                 environmentVariables CLOB,
                                 customSettings    CLOB,

    -- 状态和控制
                                 configStatus      CHAR(1)        DEFAULT 'Y' NOT NULL,
                                 defaultFlag       CHAR(1)        DEFAULT 'N' NOT NULL,
                                 priorityLevel     NUMBER(10)     DEFAULT 100,

    -- 安全和加密
                                 encryptionType    VARCHAR2(50),
                                 encryptionKey     VARCHAR2(100),

    -- 标准字段
                                 addTime           DATE           DEFAULT SYSDATE NOT NULL,
                                 addWho            VARCHAR2(32)   NOT NULL,
                                 editTime          DATE           DEFAULT SYSDATE NOT NULL,
                                 editWho           VARCHAR2(32)   NOT NULL,
                                 oprSeqFlag        VARCHAR2(32)   NOT NULL,
                                 currentVersion    NUMBER(10)     DEFAULT 1 NOT NULL,
                                 activeFlag        CHAR(1)        DEFAULT 'Y' NOT NULL,
                                 noteText          VARCHAR2(500),
                                 extProperty       CLOB,
                                 reserved1         VARCHAR2(500),
                                 reserved2         VARCHAR2(500),
                                 reserved3         VARCHAR2(500),
                                 reserved4         VARCHAR2(500),
                                 reserved5         VARCHAR2(500),
                                 reserved6         VARCHAR2(500),
                                 reserved7         VARCHAR2(500),
                                 reserved8         VARCHAR2(500),
                                 reserved9         VARCHAR2(500),
                                 reserved10        VARCHAR2(500),

    -- 主键定义
                                 CONSTRAINT PK_TOOL_CONFIG PRIMARY KEY (tenantId, toolConfigId)
);

CREATE INDEX IDX_TOOL_CONFIG_NAME      ON HUB_TOOL_CONFIG(toolName);
CREATE INDEX IDX_TOOL_CONFIG_TYPE      ON HUB_TOOL_CONFIG(toolType);
CREATE INDEX IDX_TOOL_CONFIG_CFGNAME   ON HUB_TOOL_CONFIG(configName);
CREATE INDEX IDX_TOOL_CONFIG_GROUP     ON HUB_TOOL_CONFIG(configGroupId);
CREATE INDEX IDX_TOOL_CONFIG_STATUS    ON HUB_TOOL_CONFIG(configStatus);
CREATE INDEX IDX_TOOL_CONFIG_DEFAULT   ON HUB_TOOL_CONFIG(defaultFlag);
CREATE INDEX IDX_TOOL_CONFIG_ACTIVE    ON HUB_TOOL_CONFIG(activeFlag);
COMMENT ON TABLE HUB_TOOL_CONFIG IS '工具配置主表 - 存储各种工具的基础配置信息';
