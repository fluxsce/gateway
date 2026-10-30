-- Vastbase（PostgreSQL 兼容模式）。由 scripts/db/mysql 转换。
CREATE TABLE IF NOT EXISTS HUB_AUTH_ROLE_RESOURCE (
  roleResourceId VARCHAR(100) NOT NULL,
  tenantId VARCHAR(32) NOT NULL,
  roleId VARCHAR(32) NOT NULL,
  resourceId VARCHAR(100) NOT NULL,
  permissionType VARCHAR(20) NOT NULL DEFAULT 'ALLOW',
  grantedBy VARCHAR(32) NOT NULL,
  grantedTime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expireTime TIMESTAMP DEFAULT NULL,
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
  PRIMARY KEY (tenantId, roleResourceId),
  CONSTRAINT IDX_AUTH_ROLE_RES_UNIQUE UNIQUE (tenantId, roleId, resourceId)
);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_ROLE ON HUB_AUTH_ROLE_RESOURCE (tenantId, roleId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_RESOURCE ON HUB_AUTH_ROLE_RESOURCE (tenantId, resourceId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_TYPE ON HUB_AUTH_ROLE_RESOURCE (permissionType);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_EXPIRE ON HUB_AUTH_ROLE_RESOURCE (expireTime);

INSERT INTO HUB_AUTH_ROLE_RESOURCE ( roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag ) SELECT CONCAT('ROLE_RES_SUPER_ADMIN_', REPLACE(resourceId, ':', '_')), tenantId, 'ROLE_SUPER_ADMIN', resourceId, 'ALLOW', 'system', NOW(), NOW(), 'system', NOW(), 'system', 'INIT_SA', 1, 'Y' FROM HUB_AUTH_RESOURCE WHERE tenantId = 'default' AND activeFlag = 'Y' AND resourceStatus = 'Y';

INSERT INTO HUB_AUTH_ROLE_RESOURCE ( roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag ) SELECT CONCAT('ROLE_RES_VIEWER_', REPLACE(resourceId, ':', '_')), tenantId, 'ROLE_VIEWER', resourceId, 'ALLOW', 'system', NOW(), NOW(), 'system', NOW(), 'system', 'INIT_VIEWER', 1, 'Y' FROM HUB_AUTH_RESOURCE WHERE tenantId = 'default' AND activeFlag = 'Y' AND resourceStatus = 'Y' AND ( resourceType IN ('GROUP', 'MODULE', 'MENU') OR resourceCode LIKE '%:view' OR resourceCode LIKE '%:search' OR resourceCode LIKE '%:resetQuery' OR resourceCode LIKE '%:back' OR resourceCode LIKE '%:refresh' OR ( resourceCode LIKE '%:reset' AND resourceCode <> 'hub0023:reset' ) );

INSERT INTO HUB_AUTH_ROLE_RESOURCE ( roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag ) SELECT CONCAT('ROLE_RES_VIEWER_', REPLACE(resourceId, ':', '_')), tenantId, 'ROLE_VIEWER', resourceId, 'ALLOW', 'system', NOW(), NOW(), 'system', NOW(), 'system', 'INIT_VIEWER', 1, 'Y' FROM HUB_AUTH_RESOURCE WHERE tenantId = 'default' AND resourceId = 'hub0004:export';

INSERT INTO HUB_AUTH_ROLE_RESOURCE ( roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag ) SELECT CONCAT('ROLE_RES_SUPER_ADMIN_', REPLACE(resourceId, ':', '_')), tenantId, 'ROLE_SUPER_ADMIN', resourceId, 'ALLOW', 'system', NOW(), NOW(), 'system', NOW(), 'system', 'INIT_SA', 1, 'Y' FROM HUB_AUTH_RESOURCE WHERE tenantId = 'default' AND resourceId = 'hub0020:clusterTopology';

INSERT INTO HUB_AUTH_ROLE_RESOURCE ( roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag ) SELECT CONCAT('ROLE_RES_VIEWER_', REPLACE(resourceId, ':', '_')), tenantId, 'ROLE_VIEWER', resourceId, 'ALLOW', 'system', NOW(), NOW(), 'system', NOW(), 'system', 'INIT_VIEWER', 1, 'Y' FROM HUB_AUTH_RESOURCE WHERE tenantId = 'default' AND resourceId = 'hub0020:clusterTopology';
