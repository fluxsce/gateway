-- Vastbase（PostgreSQL 兼容模式）。由 scripts/db/mysql 转换。
CREATE TABLE IF NOT EXISTS HUB_AUTH_ROLE (
  roleId VARCHAR(32) NOT NULL,
  tenantId VARCHAR(32) NOT NULL,
  roleName VARCHAR(100) NOT NULL,
  roleDescription VARCHAR(500) DEFAULT NULL,
  roleStatus VARCHAR(1) NOT NULL DEFAULT 'Y',
  builtInFlag VARCHAR(1) NOT NULL DEFAULT 'N',
  dataScope TEXT DEFAULT NULL,
  addTime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho VARCHAR(32) NOT NULL,
  editTime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho VARCHAR(32) NOT NULL,
  oprSeqFlag VARCHAR(32) NOT NULL,
  currentVersion INT NOT NULL DEFAULT 1,
  activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y',
  noteText VARCHAR(500) DEFAULT NULL,
  extProperty TEXT DEFAULT NULL,
  reserved1 VARCHAR(500) DEFAULT NULL,
  reserved2 VARCHAR(500) DEFAULT NULL,
  reserved3 VARCHAR(500) DEFAULT NULL,
  reserved4 VARCHAR(500) DEFAULT NULL,
  reserved5 VARCHAR(500) DEFAULT NULL,
  reserved6 VARCHAR(500) DEFAULT NULL,
  reserved7 VARCHAR(500) DEFAULT NULL,
  reserved8 VARCHAR(500) DEFAULT NULL,
  reserved9 VARCHAR(500) DEFAULT NULL,
  reserved10 VARCHAR(500) DEFAULT NULL,
  PRIMARY KEY (tenantId, roleId)
);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_NAME ON HUB_AUTH_ROLE (tenantId, roleName);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_STATUS ON HUB_AUTH_ROLE (roleStatus);

INSERT INTO HUB_AUTH_ROLE ( roleId, tenantId, roleName, roleDescription, roleStatus, builtInFlag, dataScope, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag ) VALUES ( 'ROLE_SUPER_ADMIN', 'default', '超级管理员', '拥有系统所有权限的超级管理员', 'Y', 'Y', '{"type":"ALL"}', NOW(), 'system', NOW(), 'system', 'INIT_001', 1, 'Y' );

INSERT INTO HUB_AUTH_ROLE ( roleId, tenantId, roleName, roleDescription, roleStatus, builtInFlag, dataScope, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag ) VALUES ( 'ROLE_VIEWER', 'default', '只读用户', '仅拥有查询、查看和重置筛选的只读权限', 'Y', 'Y', '{"type":"ALL"}', NOW(), 'system', NOW(), 'system', 'INIT_002', 1, 'Y' );
