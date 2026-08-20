-- =====================================================
-- 角色权限关联表 - 存储角色与权限资源的关联关系
-- =====================================================
CREATE TABLE HUB_AUTH_ROLE_RESOURCE (
  -- 主键和租户信息
  roleResourceId VARCHAR2(100) NOT NULL, -- 角色资源关联ID，主键
  tenantId VARCHAR2(32) NOT NULL, -- 租户ID，用于多租户数据隔离
  
  -- 关联信息
  roleId VARCHAR2(32) NOT NULL, -- 角色ID
  resourceId VARCHAR2(100) NOT NULL, -- 资源ID
  
  -- 权限控制
  permissionType VARCHAR2(20) DEFAULT 'ALLOW' NOT NULL, -- 权限类型(ALLOW:允许,DENY:拒绝)
  grantedBy VARCHAR2(32) NOT NULL, -- 授权人ID
  grantedTime DATE DEFAULT SYSDATE NOT NULL, -- 授权时间
  expireTime DATE, -- 过期时间，NULL表示永不过期
  
  -- 通用字段
  addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
  addWho VARCHAR2(32) NOT NULL, -- 创建人ID
  editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
  editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
  oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
  currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
  activeFlag CHAR(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
  noteText VARCHAR2(500), -- 备注信息
  extProperty CLOB, -- 扩展属性，JSON格式
  reserved1 VARCHAR2(500), -- 预留字段1
  reserved2 VARCHAR2(500), -- 预留字段2
  reserved3 VARCHAR2(500), -- 预留字段3
  reserved4 VARCHAR2(500), -- 预留字段4
  reserved5 VARCHAR2(500), -- 预留字段5
  reserved6 VARCHAR2(500), -- 预留字段6
  reserved7 VARCHAR2(500), -- 预留字段7
  reserved8 VARCHAR2(500), -- 预留字段8
  reserved9 VARCHAR2(500), -- 预留字段9
  reserved10 VARCHAR2(500), -- 预留字段10
  
  CONSTRAINT PK_AUTH_ROLE_RESOURCE PRIMARY KEY (tenantId, roleResourceId),
  CONSTRAINT UK_AUTH_ROLE_RES_UNIQUE UNIQUE (tenantId, roleId, resourceId)
);

CREATE INDEX IDX_AUTH_ROLE_RES_ROLE ON HUB_AUTH_ROLE_RESOURCE(tenantId, roleId);
CREATE INDEX IDX_AUTH_ROLE_RES_RESOURCE ON HUB_AUTH_ROLE_RESOURCE(tenantId, resourceId);
CREATE INDEX IDX_AUTH_ROLE_RES_TYPE ON HUB_AUTH_ROLE_RESOURCE(permissionType);
CREATE INDEX IDX_AUTH_ROLE_RES_EXPIRE ON HUB_AUTH_ROLE_RESOURCE(expireTime);
COMMENT ON TABLE HUB_AUTH_ROLE_RESOURCE IS '角色权限关联表 - 存储角色与权限资源的关联关系';

-- =====================================================
-- 初始化角色权限关联数据
-- 登录权限按角色资源表精确匹配，不会把 MODULE 自动展开为按钮。
-- 超级管理员：授予资源目录中全部有效资源（含 GROUP/MODULE/MENU/BUTTON）。
-- 只读用户：仅授予导航资源以及查看/查询/重置筛选/刷新/返回，排除写操作。
-- hub0023:reset 为日志重发，不属于只读范围。
-- =====================================================

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  'ROLE_RES_SUPER_ADMIN_' || REPLACE(resourceId, ':', '_'),
  tenantId,
  'ROLE_SUPER_ADMIN',
  resourceId,
  'ALLOW',
  'system',
  SYSDATE,
  SYSDATE,
  'system',
  SYSDATE,
  'system',
  'INIT_SA',
  1,
  'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = 'default'
  AND activeFlag = 'Y'
  AND resourceStatus = 'Y';

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  'ROLE_RES_VIEWER_' || REPLACE(resourceId, ':', '_'),
  tenantId,
  'ROLE_VIEWER',
  resourceId,
  'ALLOW',
  'system',
  SYSDATE,
  SYSDATE,
  'system',
  SYSDATE,
  'system',
  'INIT_VIEWER',
  1,
  'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = 'default'
  AND activeFlag = 'Y'
  AND resourceStatus = 'Y'
  AND (
    resourceType IN ('GROUP', 'MODULE', 'MENU')
    OR resourceCode LIKE '%:view'
    OR resourceCode LIKE '%:search'
    OR resourceCode LIKE '%:resetQuery'
    OR resourceCode LIKE '%:back'
    OR resourceCode LIKE '%:refresh'
    OR (
      resourceCode LIKE '%:reset'
      AND resourceCode <> 'hub0023:reset'
    )
  );

-- 审计日志导出对只读角色开放（仍记 EXPORT 审计）
INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  'ROLE_RES_VIEWER_' || REPLACE(resourceId, ':', '_'),
  tenantId,
  'ROLE_VIEWER',
  resourceId,
  'ALLOW',
  'system',
  SYSDATE,
  SYSDATE,
  'system',
  SYSDATE,
  'system',
  'INIT_VIEWER',
  1,
  'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = 'default'
  AND resourceId = 'hub0004:export';

COMMIT;
