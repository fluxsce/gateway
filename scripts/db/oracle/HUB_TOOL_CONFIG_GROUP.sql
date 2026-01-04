CREATE TABLE HUB_TOOL_CONFIG_GROUP (
                                       configGroupId     VARCHAR2(32)   NOT NULL,
                                       tenantId          VARCHAR2(32)   NOT NULL,

    -- 分组信息
                                       groupName         VARCHAR2(100)  NOT NULL,
                                       groupDescription  VARCHAR2(500),
                                       parentGroupId     VARCHAR2(32),
                                       groupLevel        NUMBER(10)     DEFAULT 1,
                                       groupPath         VARCHAR2(500),

    -- 分组属性
                                       groupType         VARCHAR2(50),
                                       sortOrder         NUMBER(10)     DEFAULT 100,
                                       groupIcon         VARCHAR2(100),
                                       groupColor        VARCHAR2(20),

    -- 权限控制
                                       accessLevel       VARCHAR2(20)   DEFAULT 'private',
                                       allowedUsers      CLOB,
                                       allowedRoles      CLOB,

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
                                       CONSTRAINT PK_TOOL_CONFIG_GROUP PRIMARY KEY (tenantId, configGroupId)
);

CREATE INDEX IDX_TOOL_GROUP_NAME       ON HUB_TOOL_CONFIG_GROUP(groupName);
CREATE INDEX IDX_TOOL_GROUP_PARENT     ON HUB_TOOL_CONFIG_GROUP(parentGroupId);
CREATE INDEX IDX_TOOL_GROUP_TYPE       ON HUB_TOOL_CONFIG_GROUP(groupType);
CREATE INDEX IDX_TOOL_GROUP_SORT       ON HUB_TOOL_CONFIG_GROUP(sortOrder);
CREATE INDEX IDX_TOOL_GROUP_ACTIVE     ON HUB_TOOL_CONFIG_GROUP(activeFlag);
COMMENT ON TABLE HUB_TOOL_CONFIG_GROUP IS '工具配置分组表 - 用于对工具配置进行分组管理';