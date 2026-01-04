
-- 19. 工具配置分组表
CREATE TABLE IF NOT EXISTS HUB_TOOL_CONFIG_GROUP (
    configGroupId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    groupName TEXT NOT NULL,
    groupDescription TEXT,
    parentGroupId TEXT,
    groupLevel INTEGER DEFAULT 1,
    groupPath TEXT,
    groupType TEXT,
    sortOrder INTEGER DEFAULT 100,
    groupIcon TEXT,
    groupColor TEXT,
    accessLevel TEXT DEFAULT 'private',
    allowedUsers TEXT,
    allowedRoles TEXT,
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
    PRIMARY KEY (tenantId, configGroupId)
);
CREATE INDEX IDX_TOOL_GROUP_NAME ON HUB_TOOL_CONFIG_GROUP(groupName);
CREATE INDEX IDX_TOOL_GROUP_PARENT ON HUB_TOOL_CONFIG_GROUP(parentGroupId);
CREATE INDEX IDX_TOOL_GROUP_TYPE ON HUB_TOOL_CONFIG_GROUP(groupType);
CREATE INDEX IDX_TOOL_GROUP_SORT ON HUB_TOOL_CONFIG_GROUP(sortOrder);
CREATE INDEX IDX_TOOL_GROUP_ACTIVE ON HUB_TOOL_CONFIG_GROUP(activeFlag);