-- =====================================================
-- 权限审计表 - 记录谁在哪个模块对哪条数据做了何种操作
-- =====================================================
-- 已建库补列：
-- ALTER TABLE `HUB_AUTH_AUDIT_LOG` ADD COLUMN `moduleCode` VARCHAR(64) DEFAULT NULL COMMENT '模块编码' AFTER `action`;
-- ALTER TABLE `HUB_AUTH_AUDIT_LOG` ADD COLUMN `targetName` VARCHAR(255) DEFAULT NULL COMMENT '目标显示名' AFTER `targetId`;
CREATE TABLE IF NOT EXISTS `HUB_AUTH_AUDIT_LOG` (
  `auditId` VARCHAR(32) NOT NULL COMMENT '审计ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `userId` VARCHAR(32) NOT NULL COMMENT '操作人用户ID',
  `userName` VARCHAR(100) DEFAULT NULL COMMENT '操作人用户名',

  `action` VARCHAR(32) NOT NULL COMMENT '动作细码：CREATE/UPDATE/DELETE/ROLLBACK/GRANT',
  `moduleCode` VARCHAR(64) DEFAULT NULL COMMENT '模块编码，如 hub0021',
  `targetType` VARCHAR(32) DEFAULT NULL COMMENT '目标类型(ROLE/USER/ROUTE/INSTANCE/API)',
  `targetId` VARCHAR(128) DEFAULT NULL COMMENT '目标业务主键',
  `targetName` VARCHAR(255) DEFAULT NULL COMMENT '目标显示名',
  `resourceCode` VARCHAR(128) DEFAULT NULL COMMENT '权限资源编码',
  `requestPath` VARCHAR(255) DEFAULT NULL COMMENT '请求路径',
  `requestMethod` VARCHAR(16) DEFAULT NULL COMMENT 'HTTP方法',
  `clientIP` VARCHAR(128) DEFAULT NULL COMMENT '客户端IP',
  `result` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '结果：Y成功 N失败',
  `detail` TEXT DEFAULT NULL COMMENT '补充说明，JSON或摘要',

  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记',

  PRIMARY KEY (`auditId`),
  INDEX `IDX_AUTH_AUDIT_TENANT` (`tenantId`, `addTime`),
  INDEX `IDX_AUTH_AUDIT_USER` (`userId`, `addTime`),
  INDEX `IDX_AUTH_AUDIT_ACTION` (`action`),
  INDEX `IDX_AUTH_AUDIT_MODULE` (`tenantId`, `moduleCode`, `addTime`)
) ENGINE=InnoDB COMMENT='权限审计表';
