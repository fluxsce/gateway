-- 权限审计表 - 记录谁在哪个模块对哪条数据做了何种操作
-- =====================================================
-- 已建库补列：
-- ALTER TABLE HUB_AUTH_AUDIT_LOG ADD COLUMN moduleCode TEXT;
-- ALTER TABLE HUB_AUTH_AUDIT_LOG ADD COLUMN targetName TEXT;
CREATE TABLE IF NOT EXISTS HUB_AUTH_AUDIT_LOG (
  auditId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  userId TEXT NOT NULL,
  userName TEXT,

  -- 动作细码：CREATE / UPDATE / DELETE / ROLLBACK / GRANT
  action TEXT NOT NULL,
  moduleCode TEXT,
  targetType TEXT,
  targetId TEXT,
  targetName TEXT,
  resourceCode TEXT,
  requestPath TEXT,
  requestMethod TEXT,
  clientIP TEXT,
  result TEXT NOT NULL DEFAULT 'Y',
  detail TEXT,

  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',

  PRIMARY KEY (auditId)
);

CREATE INDEX IF NOT EXISTS IDX_AUTH_AUDIT_TENANT ON HUB_AUTH_AUDIT_LOG(tenantId, addTime);
CREATE INDEX IF NOT EXISTS IDX_AUTH_AUDIT_USER ON HUB_AUTH_AUDIT_LOG(userId, addTime);
CREATE INDEX IF NOT EXISTS IDX_AUTH_AUDIT_ACTION ON HUB_AUTH_AUDIT_LOG(action);
CREATE INDEX IF NOT EXISTS IDX_AUTH_AUDIT_MODULE ON HUB_AUTH_AUDIT_LOG(tenantId, moduleCode, addTime);
