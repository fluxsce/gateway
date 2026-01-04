-- =====================================================
-- 数据权限表 - 存储用户和角色的数据访问权限
-- =====================================================
CREATE TABLE HUB_AUTH_DATA_PERMISSION (
  -- 主键和租户信息
  dataPermissionId VARCHAR2(32) NOT NULL, -- 数据权限ID，主键
  tenantId VARCHAR2(32) NOT NULL, -- 租户ID，用于多租户数据隔离
  
  -- 关联信息
  userId VARCHAR2(32), -- 用户ID，为空表示角色级权限
  roleId VARCHAR2(32), -- 角色ID，为空表示用户级权限
  
  -- 数据权限信息
  resourceType VARCHAR2(50) NOT NULL, -- 资源类型(TABLE:数据表,API:接口,MODULE:模块)
  resourceCode VARCHAR2(100) NOT NULL, -- 资源编码
  scopeValue CLOB, -- 权限范围值，JSON格式
  
  -- 权限条件
  filterCondition CLOB, -- 过滤条件，SQL WHERE条件
  columnPermissions CLOB, -- 字段权限，JSON格式
  operationPermissions VARCHAR2(50) DEFAULT 'read', -- 操作权限(read:只读,write:读写,delete:删除)
  
  -- 生效时间
  effectiveTime DATE, -- 生效时间
  expireTime DATE, -- 过期时间
  
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
  
  CONSTRAINT PK_AUTH_DATA_PERMISSION PRIMARY KEY (tenantId, dataPermissionId)
);

CREATE INDEX IDX_AUTH_DATA_PERM_USER ON HUB_AUTH_DATA_PERMISSION(tenantId, userId);
CREATE INDEX IDX_AUTH_DATA_PERM_ROLE ON HUB_AUTH_DATA_PERMISSION(tenantId, roleId);
CREATE INDEX IDX_AUTH_DATA_PERM_RESOURCE ON HUB_AUTH_DATA_PERMISSION(resourceType, resourceCode);
CREATE INDEX IDX_AUTH_DATA_PERM_EXPIRE ON HUB_AUTH_DATA_PERMISSION(expireTime);
COMMENT ON TABLE HUB_AUTH_DATA_PERMISSION IS '数据权限表 - 存储用户和角色的数据访问权限';

