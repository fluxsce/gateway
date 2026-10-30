-- Vastbase（PostgreSQL 兼容模式）。由 scripts/db/mysql 转换。
CREATE TABLE IF NOT EXISTS HUB_LOGIN_LOG (
  logId           VARCHAR(32)   NOT NULL,
  userId          VARCHAR(32)   NOT NULL,
  tenantId        VARCHAR(32)   NOT NULL,
  userName        VARCHAR(50)   NOT NULL,
  loginTime       TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  loginIp         VARCHAR(128)  NOT NULL,
  loginLocation   VARCHAR(255)  NULL,
  loginType       INT           NOT NULL DEFAULT 1,
  deviceType      VARCHAR(50)   NULL,
  deviceInfo      TEXT          NULL,
  browserInfo     TEXT          NULL,
  osInfo          VARCHAR(255)  NULL,
  loginStatus     VARCHAR(1)    NOT NULL DEFAULT 'N',
  logoutTime      TIMESTAMP      NULL,
  sessionDuration INT           NULL,
  failReason      TEXT          NULL,
  addTime         TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho          VARCHAR(32)   NOT NULL,
  editTime        TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho         VARCHAR(32)   NOT NULL,
  oprSeqFlag      VARCHAR(32)   NOT NULL,
  currentVersion  INT           NOT NULL DEFAULT 1,
  activeFlag      VARCHAR(1)    NOT NULL DEFAULT 'Y',
  PRIMARY KEY (logId)
);
CREATE INDEX IF NOT EXISTS IDX_LOGIN_USER ON HUB_LOGIN_LOG (userId);
CREATE INDEX IF NOT EXISTS IDX_LOGIN_TIME ON HUB_LOGIN_LOG (loginTime);
CREATE INDEX IF NOT EXISTS IDX_LOGIN_TENANT ON HUB_LOGIN_LOG (tenantId);
CREATE INDEX IF NOT EXISTS IDX_LOGIN_STATUS ON HUB_LOGIN_LOG (loginStatus);
CREATE INDEX IF NOT EXISTS IDX_LOGIN_TYPE ON HUB_LOGIN_LOG (loginType);
