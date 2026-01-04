
-- 6. 路由断言表
CREATE TABLE IF NOT EXISTS HUB_GW_ROUTE_ASSERTION (
    tenantId TEXT NOT NULL,
    routeAssertionId TEXT NOT NULL,
    routeConfigId TEXT NOT NULL,
    assertionName TEXT NOT NULL,
    assertionType TEXT NOT NULL,
    assertionOperator TEXT NOT NULL DEFAULT 'EQUAL',
    fieldName TEXT,
    expectedValue TEXT,
    patternValue TEXT,
    caseSensitive TEXT NOT NULL DEFAULT 'Y',
    assertionOrder INTEGER NOT NULL DEFAULT 0,
    isRequired TEXT NOT NULL DEFAULT 'Y',
    assertionDesc TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, routeAssertionId)
);
CREATE INDEX IDX_GW_ASSERT_ROUTE ON HUB_GW_ROUTE_ASSERTION(routeConfigId);
CREATE INDEX IDX_GW_ASSERT_TYPE ON HUB_GW_ROUTE_ASSERTION(assertionType);
CREATE INDEX IDX_GW_ASSERT_ORDER ON HUB_GW_ROUTE_ASSERTION(assertionOrder);