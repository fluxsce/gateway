CREATE TABLE HUB_LOGIN_LOG (
                               logId           VARCHAR2(32)   NOT NULL,
                               userId          VARCHAR2(32)   NOT NULL,
                               tenantId        VARCHAR2(32)   NOT NULL,
                               userName        VARCHAR2(50)   NOT NULL,
                               loginTime       DATE           DEFAULT SYSDATE NOT NULL,
                               loginIp         VARCHAR2(128)  DEFAULT '0.0.0.0' NOT NULL,
                               loginLocation   VARCHAR2(255),
                               loginType       NUMBER(10)     DEFAULT 1 NOT NULL,
                               deviceType      VARCHAR2(50),
                               deviceInfo      CLOB,
                               browserInfo     CLOB,
                               osInfo          VARCHAR2(255),
                               loginStatus     CHAR(1)        DEFAULT 'N' NOT NULL,
                               logoutTime      DATE,
                               sessionDuration NUMBER(10),
                               failReason      CLOB,
                               addTime         DATE           DEFAULT SYSDATE NOT NULL,
                               addWho          VARCHAR2(32)   DEFAULT 'system' NOT NULL,
                               editTime        DATE           DEFAULT SYSDATE NOT NULL,
                               editWho         VARCHAR2(32)   DEFAULT 'system' NOT NULL,
                               oprSeqFlag      VARCHAR2(32)   NOT NULL,
                               currentVersion  NUMBER(10)     DEFAULT 1 NOT NULL,
                               activeFlag      CHAR(1)        DEFAULT 'Y' NOT NULL,
                               CONSTRAINT PK_LOGIN_LOG PRIMARY KEY (logId)
);

CREATE INDEX IDX_LOGIN_USER     ON HUB_LOGIN_LOG(userId);
CREATE INDEX IDX_LOGIN_TIME     ON HUB_LOGIN_LOG(loginTime);
CREATE INDEX IDX_LOGIN_TENANT   ON HUB_LOGIN_LOG(tenantId);
CREATE INDEX IDX_LOGIN_STATUS   ON HUB_LOGIN_LOG(loginStatus);
CREATE INDEX IDX_LOGIN_TYPE     ON HUB_LOGIN_LOG(loginType);
COMMENT ON TABLE HUB_LOGIN_LOG IS '用户登录日志表';