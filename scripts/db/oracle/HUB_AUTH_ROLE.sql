-- =====================================================
-- 角色表 - 存储系统角色信息和数据权限范围
-- =====================================================
CREATE TABLE HUB_AUTH_ROLE (
  -- 主键和租户信息
  roleId VARCHAR2(32) NOT NULL, -- 角色ID，主键
  tenantId VARCHAR2(32) NOT NULL, -- 租户ID，用于多租户数据隔离
  
  -- 角色基本信息
  roleName VARCHAR2(100) NOT NULL, -- 角色名称
  roleDescription VARCHAR2(500), -- 角色描述
  
  -- 角色状态
  roleStatus CHAR(1) DEFAULT 'Y' NOT NULL, -- 角色状态(Y:启用,N:禁用)
  builtInFlag CHAR(1) DEFAULT 'N' NOT NULL, -- 内置角色标记(Y:内置,N:自定义)
  
  -- 数据权限范围
  dataScope CLOB, -- 数据权限范围，支持存储复杂的权限配置(JSON格式)
  
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
  
  CONSTRAINT PK_AUTH_ROLE PRIMARY KEY (tenantId, roleId)
);

CREATE INDEX IDX_AUTH_ROLE_NAME ON HUB_AUTH_ROLE(tenantId, roleName);
CREATE INDEX IDX_AUTH_ROLE_STATUS ON HUB_AUTH_ROLE(roleStatus);
COMMENT ON TABLE HUB_AUTH_ROLE IS '角色表 - 存储系统角色信息和数据权限范围';

-- =====================================================
-- 初始化角色数据
-- =====================================================

-- 超级管理员角色
INSERT INTO HUB_AUTH_ROLE (
  roleId, tenantId, roleName, roleDescription, 
  roleStatus, builtInFlag, dataScope,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_SUPER_ADMIN', 'default', '超级管理员', '拥有系统所有权限的超级管理员',
  'Y', 'Y', '{"type":"ALL"}',
  SYSDATE, 'system', SYSDATE, 'system', 'INIT_001', 1, 'Y'
);

COMMIT;

