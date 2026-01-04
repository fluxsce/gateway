CREATE TABLE HUB_USER (
                          userId          VARCHAR2(32)   NOT NULL, -- 用户ID，联合主键
                          tenantId        VARCHAR2(32)   NOT NULL,              -- 租户ID，联合主键
                          userName        VARCHAR2(50)   NOT NULL,              -- 用户名，登录账号
                          password        VARCHAR2(128)  NOT NULL,              -- 密码，加密存储
                          realName        VARCHAR2(50)   NOT NULL,              -- 真实姓名
                          deptId          VARCHAR2(32)   NOT NULL,              -- 所属部门ID
                          email           VARCHAR2(255),                         -- 电子邮箱
                          mobile          VARCHAR2(20),                          -- 手机号码
                          avatar          CLOB,                                  -- 头像URL或Base64数据
                          gender          NUMBER(10),                            -- 性别：1-男，2-女，0-未知
                          statusFlag      CHAR(1)        DEFAULT 'Y' NOT NULL,  -- 状态：Y-启用，N-禁用
                          deptAdminFlag   CHAR(1)        DEFAULT 'N' NOT NULL,  -- 是否部门管理员：Y-是，N-否
                          tenantAdminFlag CHAR(1)        DEFAULT 'N' NOT NULL,  -- 是否租户管理员：Y-是，N-否
                          userExpireDate  DATE           NOT NULL,              -- 用户过期时间
                          lastLoginTime   DATE,                                  -- 最后登录时间
                          lastLoginIp     VARCHAR2(128),                         -- 最后登录IP
                          pwdUpdateTime   DATE,                                  -- 密码最后更新时间
                          pwdErrorCount   NUMBER(10)     DEFAULT 0 NOT NULL,    -- 密码错误次数
                          lockTime        DATE,                                  -- 账号锁定时间
                          addTime         DATE           DEFAULT SYSDATE NOT NULL, -- 创建时间
                          addWho          VARCHAR2(32)   DEFAULT 'system' NOT NULL, -- 创建人
                          editTime        DATE           DEFAULT SYSDATE NOT NULL, -- 修改时间
                          editWho         VARCHAR2(32)   DEFAULT 'system' NOT NULL, -- 修改人
                          oprSeqFlag      VARCHAR2(32)   NOT NULL,              -- 操作序列标识
                          currentVersion  NUMBER(10)     DEFAULT 1 NOT NULL,    -- 当前版本号
                          activeFlag      CHAR(1)        DEFAULT 'Y' NOT NULL,  -- 活动状态标记：Y-活动，N-非活动
                          noteText        CLOB,                                  -- 备注信息
                          extProperty     CLOB,                                  -- 扩展属性，JSON格式
                          reserved1       VARCHAR2(500),                         -- 预留字段1
                          reserved2       VARCHAR2(500),                         -- 预留字段2
                          reserved3       VARCHAR2(500),                         -- 预留字段3
                          reserved4       VARCHAR2(500),                         -- 预留字段4
                          reserved5       VARCHAR2(500),                         -- 预留字段5
                          reserved6       VARCHAR2(500),                         -- 预留字段6
                          reserved7       VARCHAR2(500),                         -- 预留字段7
                          reserved8       VARCHAR2(500),                         -- 预留字段8
                          reserved9       VARCHAR2(500),                         -- 预留字段9
                          reserved10      VARCHAR2(500),                         -- 预留字段10
                          CONSTRAINT PK_USER PRIMARY KEY (tenantId,userId)
);

CREATE INDEX IDX_USER_TENANT ON HUB_USER(tenantId);
CREATE INDEX IDX_USER_DEPT ON HUB_USER(deptId);
CREATE INDEX IDX_USER_STATUS ON HUB_USER(statusFlag);
CREATE INDEX IDX_USER_EMAIL ON HUB_USER(email);
CREATE INDEX IDX_USER_MOBILE ON HUB_USER(mobile);
COMMENT ON TABLE HUB_USER IS '用户信息表';

