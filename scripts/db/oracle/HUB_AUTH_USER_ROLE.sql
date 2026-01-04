-- =====================================================
-- 用户角色关联表 - 存储用户与角色的关联关系
-- =====================================================
CREATE TABLE HUB_AUTH_USER_ROLE (
  -- 主键和租户信息
  userRoleId VARCHAR2(100) NOT NULL, -- 用户角色关联ID，主键
  tenantId VARCHAR2(32) NOT NULL, -- 租户ID，用于多租户数据隔离
  
  -- 关联信息
  userId VARCHAR2(32) NOT NULL, -- 用户ID
  roleId VARCHAR2(32) NOT NULL, -- 角色ID
  
  -- 授权控制
  grantedBy VARCHAR2(32) NOT NULL, -- 授权人ID
  grantedTime DATE DEFAULT SYSDATE NOT NULL, -- 授权时间
  expireTime DATE, -- 过期时间，NULL表示永不过期
  primaryRoleFlag CHAR(1) DEFAULT 'N' NOT NULL, -- 主要角色标记(Y:主要角色,N:次要角色)
  
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
  
  CONSTRAINT PK_AUTH_USER_ROLE PRIMARY KEY (tenantId, userRoleId),
  CONSTRAINT UK_AUTH_USER_ROLE UNIQUE (tenantId, userId, roleId)
);

CREATE INDEX IDX_AUTH_USER_ROLE_USER ON HUB_AUTH_USER_ROLE(tenantId, userId);
CREATE INDEX IDX_AUTH_USER_ROLE_ROLE ON HUB_AUTH_USER_ROLE(tenantId, roleId);
CREATE INDEX IDX_AUTH_USER_ROLE_PRIMARY ON HUB_AUTH_USER_ROLE(primaryRoleFlag);
CREATE INDEX IDX_AUTH_USER_ROLE_EXPIRE ON HUB_AUTH_USER_ROLE(expireTime);
COMMENT ON TABLE HUB_AUTH_USER_ROLE IS '用户角色关联表 - 存储用户与角色的关联关系';

