
-- 1. 用户信息表
CREATE TABLE IF NOT EXISTS HUB_USER (
    userId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    userName TEXT NOT NULL,
    password TEXT NOT NULL,
    realName TEXT NOT NULL,
    deptId TEXT NOT NULL,
    email TEXT,
    mobile TEXT,
    avatar TEXT,
    gender INTEGER DEFAULT 0,
    statusFlag TEXT NOT NULL DEFAULT 'Y',
    deptAdminFlag TEXT NOT NULL DEFAULT 'N',
    tenantAdminFlag TEXT NOT NULL DEFAULT 'N',
    userExpireDate DATETIME NOT NULL,
    lastLoginTime DATETIME,
    lastLoginIp TEXT,
    pwdUpdateTime DATETIME,
    pwdErrorCount INTEGER NOT NULL DEFAULT 0,
    lockTime DATETIME,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (userId, tenantId)
);

CREATE INDEX UK_USER_NAME_TENANT ON HUB_USER(userName, tenantId);
CREATE INDEX IDX_USER_TENANT ON HUB_USER(tenantId);
CREATE INDEX IDX_USER_DEPT ON HUB_USER(deptId);
CREATE INDEX IDX_USER_STATUS ON HUB_USER(statusFlag);
CREATE INDEX IDX_USER_EMAIL ON HUB_USER(email);
CREATE INDEX IDX_USER_MOBILE ON HUB_USER(mobile);

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
    'admin',
    'default',
    'admin',
    '123456',
    '系统管理员',
    'D00000001',
    'admin@example.com',
    '13800000000',
    'https://example.com/avatar.png',
    1,
    'Y',
    'N',
    'Y',
    datetime('now', '+5 years'),
    'SEQFLAG_001',
    1,
    'Y',
    'system',
    'system',
    '系统初始化管理员账号'
);