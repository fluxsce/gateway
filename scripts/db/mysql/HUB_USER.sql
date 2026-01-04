CREATE TABLE HUB_USER (
    userId          VARCHAR(32)   NOT NULL COMMENT '用户ID，联合主键',
    tenantId        VARCHAR(32)   NOT NULL COMMENT '租户ID，联合主键',
    userName        VARCHAR(50)   NOT NULL COMMENT '用户名，登录账号',
    password        VARCHAR(128)  NOT NULL COMMENT '密码，加密存储',
    realName        VARCHAR(50)   NOT NULL COMMENT '真实姓名',
    deptId          VARCHAR(32)   NOT NULL COMMENT '所属部门ID',
    email           VARCHAR(255)  NULL     COMMENT '电子邮箱',
    mobile          VARCHAR(20)   NULL     COMMENT '手机号码',
    avatar          LONGTEXT      NULL     COMMENT '头像URL或Base64数据',
    gender          INT           NULL     DEFAULT 0 COMMENT '性别：1-男，2-女，0-未知',
    statusFlag      VARCHAR(1)    NOT NULL DEFAULT 'Y' COMMENT '状态：Y-启用，N-禁用',
    deptAdminFlag   VARCHAR(1)    NOT NULL DEFAULT 'N' COMMENT '是否部门管理员：Y-是，N-否',
    tenantAdminFlag VARCHAR(1)    NOT NULL DEFAULT 'N' COMMENT '是否租户管理员：Y-是，N-否',
    userExpireDate  DATETIME      NOT NULL COMMENT '用户过期时间',
    lastLoginTime   DATETIME      NULL     COMMENT '最后登录时间',
    lastLoginIp     VARCHAR(128)  NULL     COMMENT '最后登录IP',
    pwdUpdateTime   DATETIME      NULL     COMMENT '密码最后更新时间',
    pwdErrorCount   INT           NOT NULL DEFAULT 0 COMMENT '密码错误次数',
    lockTime        DATETIME      NULL     COMMENT '账号锁定时间',
    addTime         DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho          VARCHAR(32)   NOT NULL COMMENT '创建人',
    editTime        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    editWho         VARCHAR(32)   NOT NULL COMMENT '修改人',
    oprSeqFlag      VARCHAR(32)   NOT NULL COMMENT '操作序列标识',
    currentVersion  INT           NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag      VARCHAR(1)    NOT NULL DEFAULT 'Y' COMMENT '活动状态标记：Y-活动，N-非活动',
    noteText        TEXT          NULL     COMMENT '备注信息',
    extProperty     TEXT          DEFAULT NULL COMMENT '扩展属性，JSON格式',
    reserved1       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段1',
    reserved2       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段2',
    reserved3       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段3',
    reserved4       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段4',
    reserved5       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段5',
    reserved6       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段6',
    reserved7       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段7',
    reserved8       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段8',
    reserved9       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段9',
    reserved10      VARCHAR(500)  DEFAULT NULL COMMENT '预留字段10',
    PRIMARY KEY (userId, tenantId),
    INDEX UK_USER_NAME_TENANT (userName, tenantId), -- 普通索引代替 UNIQUE KEY
    INDEX IDX_USER_TENANT (tenantId),
    INDEX IDX_USER_DEPT (deptId),
    INDEX IDX_USER_STATUS (statusFlag),
    INDEX IDX_USER_EMAIL (email),
    INDEX IDX_USER_MOBILE (mobile)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户信息表';

INSERT INTO HUB_USER (
    userId,
    tenantId,
    userName,
    password,
    realName,
    deptId,
    email,
    mobile,
    avatar,
    gender,
    statusFlag,
    deptAdminFlag,
    tenantAdminFlag,
    userExpireDate,
    oprSeqFlag,
    currentVersion,
    activeFlag,
		addWho,
		editWho,
    noteText
) VALUES (
    'admin',                            -- userId
    'default',                          -- tenantId
    'admin',                            -- userName
    '123456',                      -- password（使用 MySQL 内置 MD5 加密）
    '系统管理员',                         -- realName
    'D00000001',                        -- deptId
    'admin@example.com',                -- email
    '13800000000',                      -- mobile
    'https://example.com/avatar.png',   -- avatar
    1,                                  -- gender (1:男)
    'Y',                                -- statusFlag
    'N',                                -- deptAdminFlag
    'Y',                                -- tenantAdminFlag
    NOW() + INTERVAL 5 YEAR,            -- userExpireDate（5年后过期）
    'SEQFLAG_001',                      -- oprSeqFlag
    1,                                  -- currentVersion
    'Y',                                -- activeFlag
		'system',
		'system',
    '系统初始化管理员账号'              -- noteText
);

-- =====================================================
-- ALTER 变更语句：用户头像字段类型调整
-- 变更日期：2025-10-10
-- 变更原因：支持存储Base64编码的图片数据
-- 兼容性：向后兼容，现有URL数据不受影响
-- =====================================================
ALTER TABLE HUB_USER MODIFY COLUMN avatar LONGTEXT NULL COMMENT '头像URL或Base64数据';
