-- Vastbase（PostgreSQL 兼容模式）。由 scripts/db/mysql 转换。
CREATE TABLE IF NOT EXISTS HUB_USER (
  userId          VARCHAR(32)   NOT NULL,
  tenantId        VARCHAR(32)   NOT NULL,
  userName        VARCHAR(50)   NOT NULL,
  password        VARCHAR(128)  NOT NULL,
  realName        VARCHAR(50)   NOT NULL,
  deptId          VARCHAR(32)   NOT NULL,
  email           VARCHAR(255)  NULL,
  mobile          VARCHAR(20)   NULL,
  avatar          TEXT      NULL,
  gender          INT           NULL     DEFAULT 0,
  statusFlag      VARCHAR(1)    NOT NULL DEFAULT 'Y',
  deptAdminFlag   VARCHAR(1)    NOT NULL DEFAULT 'N',
  tenantAdminFlag VARCHAR(1)    NOT NULL DEFAULT 'N',
  userExpireDate  TIMESTAMP      NOT NULL,
  lastLoginTime   TIMESTAMP      NULL,
  lastLoginIp     VARCHAR(128)  NULL,
  pwdUpdateTime   TIMESTAMP      NULL,
  pwdErrorCount   INT           NOT NULL DEFAULT 0,
  lockTime        TIMESTAMP      NULL,
  addTime         TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho          VARCHAR(32)   NOT NULL,
  editTime        TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho         VARCHAR(32)   NOT NULL,
  oprSeqFlag      VARCHAR(32)   NOT NULL,
  currentVersion  INT           NOT NULL DEFAULT 1,
  activeFlag      VARCHAR(1)    NOT NULL DEFAULT 'Y',
  noteText        TEXT          NULL,
  extProperty     TEXT          DEFAULT NULL,
  reserved1       VARCHAR(500)  DEFAULT NULL,
  reserved2       VARCHAR(500)  DEFAULT NULL,
  reserved3       VARCHAR(500)  DEFAULT NULL,
  reserved4       VARCHAR(500)  DEFAULT NULL,
  reserved5       VARCHAR(500)  DEFAULT NULL,
  reserved6       VARCHAR(500)  DEFAULT NULL,
  reserved7       VARCHAR(500)  DEFAULT NULL,
  reserved8       VARCHAR(500)  DEFAULT NULL,
  reserved9       VARCHAR(500)  DEFAULT NULL,
  reserved10      VARCHAR(500)  DEFAULT NULL,
  PRIMARY KEY (userId, tenantId)
);
CREATE INDEX IF NOT EXISTS UK_USER_NAME_TENANT ON HUB_USER (userName, tenantId);
CREATE INDEX IF NOT EXISTS IDX_USER_TENANT ON HUB_USER (tenantId);
CREATE INDEX IF NOT EXISTS IDX_USER_DEPT ON HUB_USER (deptId);
CREATE INDEX IF NOT EXISTS IDX_USER_STATUS ON HUB_USER (statusFlag);
CREATE INDEX IF NOT EXISTS IDX_USER_EMAIL ON HUB_USER (email);
CREATE INDEX IF NOT EXISTS IDX_USER_MOBILE ON HUB_USER (mobile);

INSERT INTO HUB_USER ( userId, tenantId, userName, password, realName, deptId, email, mobile, avatar, gender, statusFlag, deptAdminFlag, tenantAdminFlag, userExpireDate, oprSeqFlag, currentVersion, activeFlag, addWho, editWho, noteText ) VALUES ( 'admin', 'default', 'admin', '$2a$10$S9Yqyb9LI5PqAutYj.kR0OI/Zm7EcJSKbxKaLCThw8djqwqsPiDQi', '系统管理员', 'D00000001', 'admin@example.com', '13800000000', 'https://example.com/avatar.png', 1, 'Y', 'N', 'Y', NOW() + INTERVAL '5 years', 'SEQFLAG_001', 1, 'Y', 'system', 'system', '系统初始化管理员账号' );

ALTER TABLE HUB_USER ALTER COLUMN avatar TYPE TEXT;

ALTER TABLE HUB_USER ADD COLUMN mustChangePwd VARCHAR(1) NOT NULL DEFAULT 'N';

UPDATE HUB_USER SET mustChangePwd = 'Y' WHERE userId = 'admin' AND tenantId = 'default' AND password = '$2a$10$S9Yqyb9LI5PqAutYj.kR0OI/Zm7EcJSKbxKaLCThw8djqwqsPiDQi';
