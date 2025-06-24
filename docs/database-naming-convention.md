# 数据库设计规范

## 1. 命名规范

### 1.1 数据库命名

- 使用小写字母
- 多个单词使用下划线分隔
- 例如：`web_hub_here`

### 1.2 表命名规范

- 所有表名必须以 `HUB` 开头
- 表名使用大写字母
- 多个单词使用下划线分隔
- 模块前缀\_功能，例如：`HUB_USER_ACCOUNT`、`HUB_ORDER_INFO`
- 避免使用数据库关键字
- 表名应当使用单数形式

### 1.3 字段命名规范

- 使用驼峰命名法(camelCase)
- 禁止直接使用 `id` 作为字段名，主键命名应当体现业务含义，如 `userId`、`orderId` 等
- 严格禁止使用自增字段
- 避免使用数据库关键字作为字段名
- 字段名称必须清晰表达其用途和业务含义，避免使用缩写
- 外键字段名应当表明与其相关联的表和字段，如 `parentUserId`、`relatedOrderId` 等
- 禁止字段使用 `is` 前缀，例如：`isDeleted`、`isActive`、`isApproved`
- 日期时间类型统一使用DATETIME，时间字段应设置合适的默认值：
  - `addTime`：使用 `DEFAULT CURRENT_TIMESTAMP`
  - `editTime`：使用 `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`
- 金额类型字段应以 `amount` 或 `price` 结尾，如 `totalAmount`、`unitPrice`
- 数量类型字段应以 `count`、`num` 或 `quantity` 结尾，如 `visitCount`、`stockQuantity`
- 状态标记字段使用 `VARCHAR(1)` 类型，值为 'Y'/'N'，如 `activeFlag VARCHAR(1) DEFAULT 'Y'`

## 2. 必须包含的通用字段

所有表必须包含以下字段：

| 字段名          | 类型         | 默认值                                  | 说明                                                                                       |
| --------------- | ------------ | --------------------------------------- | ------------------------------------------------------------------------------------------ |
| tablePrimaryKey | VARCHAR(32)  | -                                       | 主键，使用UUID或其他生成策略，不使用自增。根据表名具体命名，如userAccountId, orderInfoId等 |
| tenantId        | VARCHAR(32)  | NOT NULL                                | 租户ID，用于多租户数据隔离                                                                 |
| addTime         | DATETIME     | DEFAULT CURRENT_TIMESTAMP               | 创建时间，自动设置为当前时间                                                               |
| addWho          | VARCHAR(32)  | -                                       | 创建人ID，关联用户表                                                                       |
| editTime        | DATETIME     | DEFAULT CURRENT_TIMESTAMP ON UPDATE... | 最后修改时间，创建时设置为当前时间，更新时自动更新                                         |
| editWho         | VARCHAR(32)  | -                                       | 最后修改人ID，关联用户表                                                                   |
| oprSeqFlag      | VARCHAR(32)  | -                                       | 操作序列标识，用于乐观锁控制                                                               |
| currentVersion  | INT          | DEFAULT 1                               | 当前版本号，初始值为1，每次更新+1                                                          |
| activeFlag      | VARCHAR(1)   | DEFAULT 'Y'                             | 活动状态标记，'N'表示非活动，'Y'表示活动                                                   |
| noteText        | VARCHAR(500) | DEFAULT NULL                            | 备注信息                                                                                   |
| reserved1       | VARCHAR(500) | DEFAULT NULL                            | 预留字段1，用于业务扩展                                                                   |
| reserved2       | VARCHAR(500) | DEFAULT NULL                            | 预留字段2，用于业务扩展                                                                   |
| reserved3       | VARCHAR(500) | DEFAULT NULL                            | 预留字段3，用于业务扩展                                                                   |
| reserved4       | VARCHAR(500) | DEFAULT NULL                            | 预留字段4，用于业务扩展                                                                   |
| reserved5       | VARCHAR(500) | DEFAULT NULL                            | 预留字段5，用于业务扩展                                                                   |
| reserved6       | VARCHAR(500) | DEFAULT NULL                            | 预留字段6，用于业务扩展                                                                   |
| reserved7       | VARCHAR(500) | DEFAULT NULL                            | 预留字段7，用于业务扩展                                                                   |
| reserved8       | VARCHAR(500) | DEFAULT NULL                            | 预留字段8，用于业务扩展                                                                   |
| reserved9       | VARCHAR(500) | DEFAULT NULL                            | 预留字段9，用于业务扩展                                                                   |
| reserved10      | VARCHAR(500) | DEFAULT NULL                            | 预留字段10，用于业务扩展                                                                  |

## 3. 字段命名语义明确性要求

为确保数据库设计的可读性和可维护性，字段命名必须做到语义明确：

1. **业务含义明确**：字段名称应直接反映其业务含义，例如用`customerPhone`而不是`phone`或`mobile`
2. **避免使用通用词汇**：如`name`、`code`、`type`、`status`等，应使用更具体的词如`productName`、`orderCode`、`accountType`、`paymentStatus`
3. **使用完整词汇**：避免使用不直观的缩写，如使用`description`而非`desc`，使用`address`而非`addr`
4. **复合字段名**：由多个词组成的字段名，每个部分都应体现其含义，如`deliveryAddressCity`而非`city`
5. **字段组一致性**：相关字段应使用一致的命名方式，如地址相关的字段可以是`shippingAddressCity`、`shippingAddressState`等

## 4. 其他规范

### 4.1 数据类型选择

- 字符串类型：
  - VARCHAR：所有字符串字段统一使用VARCHAR类型
  - TEXT：大文本内容
- 数值类型：统一使用 INT 类型，避免使用 TINYINT、SMALLINT、BIGINT 等
- 时间类型：统一使用 DATETIME
- 精确数值：使用 DECIMAL，避免使用 FLOAT/DOUBLE
- 布尔值/状态标记：使用 VARCHAR(1)，值为 'Y'/'N'

### 4.2 索引设计规范

- 主键索引：每张表必须有主键
- 外键索引：所有外键字段必须创建索引
- 常用查询字段：经常作为查询条件的字段应创建索引
- 联合索引：多字段组合查询应考虑创建联合索引
- 索引命名：idx*表名*字段名

### 4.3 外键约束

- 根据实际业务需求决定是否使用外键约束
- 如使用外键，命名为：fk*表名*关联表名
- 一般推荐在应用层实现外键关系，数据库层不强制使用外键约束

### 4.4 表注释和字段注释

- 每个表必须有表注释，描述表的用途
- 每个字段必须有字段注释，描述字段的含义
- 注释应详细表明业务含义，便于后期维护

### 4.5 多租户支持

- 所有表必须包含 `tenantId` 字段，用于多租户数据隔离
- 查询时必须加上租户ID条件，确保数据安全隔离
- 建议为 `tenantId` 字段创建索引，提高查询性能
- 租户ID应在应用层统一管理和注入

### 4.6 预留字段使用规范

- 所有表必须包含 `reserved1` 到 `reserved10` 共10个预留字段
- 预留字段类型统一为 `VARCHAR(500) DEFAULT NULL`
- 预留字段用于业务快速扩展，避免频繁修改表结构
- 使用预留字段时应在注释中说明具体用途
- 预留字段的使用应有统一的命名和使用约定

## 5. 示例

### 用户表 (HUB_USER_ACCOUNT)

```sql
CREATE TABLE `HUB_USER_ACCOUNT` (
  `userAccountId` VARCHAR(32) NOT NULL COMMENT '用户账号ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `userName` VARCHAR(50) NOT NULL COMMENT '用户名，登录账号',
  `password` VARCHAR(100) NOT NULL COMMENT '密码，加密存储',
  `passwordSalt` VARCHAR(32) NOT NULL COMMENT '密码盐值',
  `nickName` VARCHAR(100) DEFAULT NULL COMMENT '用户昵称',
  `emailAddress` VARCHAR(100) DEFAULT NULL COMMENT '电子邮箱地址',
  `mobilePhone` VARCHAR(20) DEFAULT NULL COMMENT '手机号码',
  `avatarUrl` VARCHAR(255) DEFAULT NULL COMMENT '头像URL',
  `genderType` INT DEFAULT 0 COMMENT '性别类型(0未知，1男，2女)',
  `accountStatus` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '账号状态(N禁用,Y正常)',
  `lastLoginTime` DATETIME DEFAULT NULL COMMENT '最后登录时间',
  `lastLoginIpAddress` VARCHAR(50) DEFAULT NULL COMMENT '最后登录IP地址',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  PRIMARY KEY (`userAccountId`),
  UNIQUE KEY `idx_HUB_USER_ACCOUNT_userName` (`userName`),
  UNIQUE KEY `idx_HUB_USER_ACCOUNT_emailAddress` (`emailAddress`),
  UNIQUE KEY `idx_HUB_USER_ACCOUNT_mobilePhone` (`mobilePhone`),
  KEY `idx_HUB_USER_ACCOUNT_tenantId` (`tenantId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户账号表';
```
