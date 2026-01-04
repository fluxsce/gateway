
-- 18. 工具配置主表
CREATE TABLE IF NOT EXISTS HUB_TOOL_CONFIG (
    toolConfigId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    toolName TEXT NOT NULL,
    toolType TEXT NOT NULL,
    toolVersion TEXT,
    configName TEXT NOT NULL,
    configDescription TEXT,
    configGroupId TEXT,
    configGroupName TEXT,
    hostAddress TEXT,
    portNumber INTEGER,
    protocolType TEXT,
    authType TEXT,
    userName TEXT,
    passwordEncrypted TEXT,
    keyFilePath TEXT,
    keyFileContent TEXT,
    configParameters TEXT,
    environmentVariables TEXT,
    customSettings TEXT,
    configStatus TEXT NOT NULL DEFAULT 'Y',
    defaultFlag TEXT NOT NULL DEFAULT 'N',
    priorityLevel INTEGER DEFAULT 100,
    encryptionType TEXT,
    encryptionKey TEXT,
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
    PRIMARY KEY (tenantId, toolConfigId)
);
CREATE INDEX IDX_TOOL_CONFIG_NAME ON HUB_TOOL_CONFIG(toolName);
CREATE INDEX IDX_TOOL_CONFIG_TYPE ON HUB_TOOL_CONFIG(toolType);
CREATE INDEX IDX_TOOL_CONFIG_CFGNAME ON HUB_TOOL_CONFIG(configName);
CREATE INDEX IDX_TOOL_CONFIG_GROUP ON HUB_TOOL_CONFIG(configGroupId);
CREATE INDEX IDX_TOOL_CONFIG_STATUS ON HUB_TOOL_CONFIG(configStatus);
CREATE INDEX IDX_TOOL_CONFIG_DEFAULT ON HUB_TOOL_CONFIG(defaultFlag);
CREATE INDEX IDX_TOOL_CONFIG_ACTIVE ON HUB_TOOL_CONFIG(activeFlag);