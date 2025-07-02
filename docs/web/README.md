# 数据中心登录模块设计说明

## 1. 模块概述

数据中心登录模块(Hub0001)是Web Hub Here平台的核心用户认证和授权管理模块，旨在提供安全、可靠的身份验证系统，确保系统资源的安全访问。

### 1.1 目的

提供统一的用户身份验证入口，支持多种登录方式和安全策略，保障系统安全。

### 1.2 功能范围

- 用户名密码登录
- 验证码校验
- 记住登录状态
- 密码找回
- 登录日志记录
- 安全策略管理（密码强度、账号锁定等）

### 1.3 用户角色

| 角色     | 权限描述                             |
| -------- | ------------------------------------ |
| 普通用户 | 可以使用登录系统访问被授权的资源     |
| 管理员   | 可以管理用户账号、查看登录日志等操作 |
| 系统管理 | 可以配置系统安全策略、审计登录行为   |

## 2. 设计原则

1. **安全性**：采用多层次安全防护，防止暴力破解和账号盗用
2. **可用性**：提供直观简洁的用户界面，操作流程清晰明了
3. **可扩展性**：支持后续扩展多种登录方式（如第三方登录）
4. **可靠性**：提供稳定的身份认证服务，保障用户登录体验

## 3. 技术架构

- **前端**：Vue 3 + TypeScript + Naive UI
- **状态管理**：Pinia
- **身份验证**：JWT (JSON Web Token)
- **路由**：Vue Router

## 4. 登录流程

1. 用户访问登录页面
2. 系统显示登录表单（用户名、密码、验证码）
3. 用户填写表单并提交
4. 服务端验证用户名、密码和验证码
5. 验证通过，生成JWT令牌返回给前端
6. 前端存储令牌并重定向到首页
7. 记录登录日志

## 5. 安全措施

1. **密码加密**：密码采用不可逆加密存储，传输过程中加密
2. **验证码机制**：防止暴力破解
3. **账号锁定**：连续多次登录失败后临时锁定账号
4. **密码策略**：强制密码复杂度要求，定期更换
5. **登录监控**：记录异常登录行为，监控异地登录

## 6. 数据模型

### 6.1 用户表 (HUB_USER)

用于存储用户基本信息和登录认证数据。

#### 表结构

```sql
CREATE TABLE HUB_USER (
    userId          VARCHAR(32)   NOT NULL COMMENT '用户ID，联合主键',
    tenantId        VARCHAR(32)   NOT NULL COMMENT '租户ID，联合主键',
    userName        VARCHAR(50)   NOT NULL COMMENT '用户名，登录账号',
    password        VARCHAR(128)  NOT NULL COMMENT '密码，加密存储',
    realName        VARCHAR(50)   NOT NULL COMMENT '真实姓名',
    deptId          VARCHAR(32)   NOT NULL COMMENT '所属部门ID',
    email           VARCHAR(255)  NULL     COMMENT '电子邮箱',
    mobile          VARCHAR(20)   NULL     COMMENT '手机号码',
    avatar          VARCHAR(500)  NULL     COMMENT '头像URL',
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
    `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
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
    PRIMARY KEY (userId, tenantId),
    UNIQUE KEY UK_USER_NAME_TENANT (userName, tenantId),
    INDEX IDX_USER_TENANT (tenantId),
    INDEX IDX_USER_DEPT (deptId),
    INDEX IDX_USER_STATUS (statusFlag),
    INDEX IDX_USER_EMAIL (email),
    INDEX IDX_USER_MOBILE (mobile)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户信息表';
```

### 6.2 登录日志表 (HUB_LOGIN_LOG)

记录用户登录行为，用于安全审计和行为分析。

#### 表结构

```sql
CREATE TABLE HUB_LOGIN_LOG (
    logId           VARCHAR(32)   NOT NULL COMMENT '日志ID，主键',
    userId          VARCHAR(32)   NOT NULL COMMENT '用户ID',
    tenantId        VARCHAR(32)   NOT NULL COMMENT '租户ID',
    userName        VARCHAR(50)   NOT NULL COMMENT '用户名',
    loginTime       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    loginIp         VARCHAR(128)  NOT NULL COMMENT '登录IP',
    loginLocation   VARCHAR(255)  NULL     COMMENT '登录地点',
    loginType       INT           NOT NULL DEFAULT 1 COMMENT '登录类型：1-用户名密码，2-验证码，3-第三方',
    deviceType      VARCHAR(50)   NULL     COMMENT '设备类型',
    deviceInfo      TEXT          NULL     COMMENT '设备信息',
    browserInfo     TEXT          NULL     COMMENT '浏览器信息',
    osInfo          VARCHAR(255)  NULL     COMMENT '操作系统信息',
    loginStatus     VARCHAR(1)    NOT NULL DEFAULT 'N' COMMENT '登录状态：Y-成功，N-失败',
    logoutTime      DATETIME      NULL     COMMENT '登出时间',
    sessionDuration INT           NULL     COMMENT '会话持续时长(秒)',
    failReason      TEXT          NULL     COMMENT '失败原因',
    addTime         DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho          VARCHAR(32)   NOT NULL COMMENT '创建人',
    editTime        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    editWho         VARCHAR(32)   NOT NULL COMMENT '修改人',
    oprSeqFlag      VARCHAR(32)   NOT NULL COMMENT '操作序列标识',
    currentVersion  INT           NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag      VARCHAR(1)    NOT NULL DEFAULT 'Y' COMMENT '活动状态标记：Y-活动，N-非活动',
    PRIMARY KEY (logId),
    INDEX IDX_LOGIN_USER (userId),
    INDEX IDX_LOGIN_TIME (loginTime),
    INDEX IDX_LOGIN_TENANT (tenantId),
    INDEX IDX_LOGIN_STATUS (loginStatus),
    INDEX IDX_LOGIN_TYPE (loginType)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户登录日志表';
```

### 6.3 刷新令牌表 (HUB_REFRESH_TOKEN)

存储用户刷新令牌，用于JWT认证系统中的令牌刷新机制。

#### 表结构

```sql
CREATE TABLE HUB_REFRESH_TOKEN (
    tokenId        VARCHAR(32)   NOT NULL COMMENT '令牌ID，主键',
    userId         VARCHAR(32)   NOT NULL COMMENT '用户ID',
    tenantId       VARCHAR(32)   NOT NULL COMMENT '租户ID',
    token          VARCHAR(255)  NOT NULL COMMENT '刷新令牌值',
    createTime     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    expireTime     DATETIME      NOT NULL COMMENT '过期时间',
    tokenStatus    VARCHAR(20)   NOT NULL COMMENT '状态：ACTIVE-活动，REVOKED-已撤销，EXPIRED-已过期',
    updateTime     DATETIME      NULL     COMMENT '更新时间',
    clientIp       VARCHAR(128)  NULL     COMMENT '客户端IP',
    userAgent      TEXT          NULL     COMMENT '用户代理信息',
    deviceId       VARCHAR(128)  NULL     COMMENT '设备标识',
    addTime        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho         VARCHAR(32)   NOT NULL COMMENT '创建人',
    editTime       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    editWho        VARCHAR(32)   NOT NULL COMMENT '修改人',
    oprSeqFlag     VARCHAR(32)   NOT NULL COMMENT '操作序列标识',
    currentVersion INT           NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag     VARCHAR(1)    NOT NULL DEFAULT 'Y' COMMENT '活动状态标记：Y-活动，N-非活动',
    PRIMARY KEY (tokenId),
    UNIQUE KEY UK_TOKEN_VALUE (token),
    INDEX IDX_TOKEN_USER (userId),
    INDEX IDX_TOKEN_TENANT (tenantId),
    INDEX IDX_TOKEN_EXPIRE (expireTime),
    INDEX IDX_TOKEN_STATUS (tokenStatus)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='刷新令牌表';
```

### 6.4 验证码记录表 (HUB_CAPTCHA)

存储系统生成的验证码信息，用于登录验证。

#### 表结构

```sql
CREATE TABLE HUB_CAPTCHA (
    captchaId      VARCHAR(32)   NOT NULL COMMENT '验证码ID，主键',
    captchaCode    VARCHAR(10)   NOT NULL COMMENT '验证码内容',
    captchaType    INT           NOT NULL DEFAULT 1 COMMENT '验证码类型：1-图形，2-短信，3-邮件',
    targetValue    VARCHAR(255)  NULL     COMMENT '目标值（手机号/邮箱）',
    createTime     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    expireTime     DATETIME      NOT NULL COMMENT '过期时间',
    usedFlag       VARCHAR(1)    NOT NULL DEFAULT 'N' COMMENT '是否已使用：Y-是，N-否',
    usedTime       DATETIME      NULL     COMMENT '使用时间',
    ipAddress      VARCHAR(128)  NOT NULL COMMENT '请求IP地址',
    sessionId      VARCHAR(128)  NULL     COMMENT '会话ID',
    businessType   VARCHAR(50)   NULL     COMMENT '业务类型：login-登录，register-注册，reset-重置密码',
    verifyCount    INT           NOT NULL DEFAULT 0 COMMENT '验证尝试次数',
    addTime        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho         VARCHAR(32)   NOT NULL COMMENT '创建人',
    editTime       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    editWho        VARCHAR(32)   NOT NULL COMMENT '修改人',
    oprSeqFlag     VARCHAR(32)   NOT NULL COMMENT '操作序列标识',
    currentVersion INT           NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag     VARCHAR(1)    NOT NULL DEFAULT 'Y' COMMENT '活动状态标记：Y-活动，N-非活动',
    PRIMARY KEY (captchaId),
    INDEX IDX_CAPTCHA_TARGET (targetValue),
    INDEX IDX_CAPTCHA_EXPIRE (expireTime),
    INDEX IDX_CAPTCHA_BUSINESS (businessType),
    INDEX IDX_CAPTCHA_SESSION (sessionId)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='验证码记录表';
```

### 6.5 角色权限关联表 (HUB_ROLE_PERM)

记录角色与权限的关联关系。

#### 表结构

```sql
CREATE TABLE HUB_ROLE_PERM (
    roleId         VARCHAR(32)    NOT NULL COMMENT '角色ID，联合主键',
    permCode       VARCHAR(100)   NOT NULL COMMENT '权限标识，联合主键',
    tenantId       VARCHAR(32)    NOT NULL COMMENT '租户ID，联合主键',
    addTime        DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho         VARCHAR(32)    NOT NULL COMMENT '创建人',
    editTime       DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    editWho        VARCHAR(32)    NOT NULL COMMENT '修改人',
    oprSeqFlag     VARCHAR(32)    NOT NULL COMMENT '操作序列标识',
    currentVersion INT            NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag     VARCHAR(1)     NOT NULL DEFAULT 'Y' COMMENT '活动状态标记：Y-活动，N-非活动',
    noteText       TEXT           NULL     COMMENT '备注信息',
    PRIMARY KEY (roleId, permCode, tenantId),
    INDEX IDX_ROLE_PERM_ROLE (roleId),
    INDEX IDX_ROLE_PERM_PERM (permCode),
    INDEX IDX_ROLE_PERM_TENANT (tenantId)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表';
```

## 7. 模块结构

```
src/views/hub0001/
├── LoginView.vue           # 登录主页面
├── components/             # 登录模块专用组件
│   ├── CaptchaInput.vue    # 验证码输入组件
│   ├── RememberMe.vue      # 记住登录组件
│   └── ForgotPassword.vue  # 忘记密码组件
├── hooks/                  # 登录模块专用钩子函数
│   └── useAuth.ts          # 认证相关钩子
├── api/                    # API接口
│   └── auth.ts             # 认证相关API
└── types/                  # 模块类型定义
    └── index.ts            # 类型定义文件
```

## 8. 数据库脚本

### 8.1 创建菜单表

```sql
CREATE TABLE HUB_MENU (
    menuId         VARCHAR(32)    NOT NULL COMMENT '菜单ID，联合主键',
    menuName       VARCHAR(100)   NOT NULL COMMENT '菜单名称',
    parentId       VARCHAR(32)    NULL     COMMENT '父菜单ID',
    menuPath       VARCHAR(500)   NOT NULL COMMENT '菜单路径，如：/system/user',
    component      VARCHAR(500)   NULL     COMMENT '前端组件路径',
    icon           VARCHAR(255)   NULL     COMMENT '菜单图标',
    sortOrder      INT            NOT NULL DEFAULT 0 COMMENT '排序',
    menuType       INT            NOT NULL DEFAULT 1 COMMENT '菜单类型：1-目录，2-菜单，3-按钮/权限点',
    permCode       VARCHAR(100)   NULL     COMMENT '权限标识，如：system:user:add',
    visibleFlag    VARCHAR(1)     NOT NULL DEFAULT 'Y' COMMENT '是否可见：Y-可见，N-隐藏',
    statusFlag     VARCHAR(1)     NOT NULL DEFAULT 'Y' COMMENT '状态：Y-启用，N-禁用',
    sysMenuFlag    VARCHAR(1)     NOT NULL DEFAULT 'N' COMMENT '是否系统菜单：Y-是，N-否',
    addTime        DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho         VARCHAR(32)    NOT NULL COMMENT '创建人',
    editTime       DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    editWho        VARCHAR(32)    NOT NULL COMMENT '修改人',
    oprSeqFlag     VARCHAR(32)    NOT NULL COMMENT '操作序列标识',
    currentVersion INT            NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag     VARCHAR(1)     NOT NULL DEFAULT 'Y' COMMENT '活动状态标记：Y-活动，N-非活动',
    languageId     VARCHAR(20)    NOT NULL COMMENT '语言，联合主键',
    noteText       TEXT           NULL     COMMENT '备注信息',
    PRIMARY KEY (menuId, languageId),
    INDEX IDX_MENU_PARENT (parentId),
    INDEX IDX_MENU_TYPE (menuType),
    INDEX IDX_MENU_PERM (permCode),
    INDEX IDX_MENU_STATUS (statusFlag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='菜单表';
```

### 8.2 创建权限表

```sql
CREATE TABLE HUB_PERMISSION (
    permCode       VARCHAR(100)   NOT NULL COMMENT '权限标识，如：system:user:add，联合主键',
    permName       VARCHAR(100)   NOT NULL COMMENT '权限名称',
    menuId         VARCHAR(32)    NOT NULL COMMENT '所属菜单ID，联合主键',
    permType       INT            NOT NULL DEFAULT 1 COMMENT '权限类型：1-菜单，2-按钮，3-数据',
    statusFlag     VARCHAR(1)     NOT NULL DEFAULT 'Y' COMMENT '状态：Y-启用，N-禁用',
    addTime        DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho         VARCHAR(32)    NOT NULL COMMENT '创建人',
    editTime       DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    editWho        VARCHAR(32)    NOT NULL COMMENT '修改人',
    oprSeqFlag     VARCHAR(32)    NOT NULL COMMENT '操作序列标识',
    currentVersion INT            NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag     VARCHAR(1)     NOT NULL DEFAULT 'Y' COMMENT '活动状态标记：Y-活动，N-非活动',
    noteText       TEXT           NULL     COMMENT '备注信息',
    PRIMARY KEY (permCode, menuId),
    INDEX IDX_PERM_CODE (permCode),
    INDEX IDX_PERM_MENU (menuId),
    INDEX IDX_PERM_TYPE (permType),
    INDEX IDX_PERM_STATUS (statusFlag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表';

CREATE TABLE HUB_ROLE_PERM (
    roleId         VARCHAR(32)    NOT NULL COMMENT '角色ID，联合主键',
    permCode       VARCHAR(100)   NOT NULL COMMENT '权限标识，联合主键',
    tenantId       VARCHAR(32)    NOT NULL COMMENT '租户ID，联合主键',
    addTime        DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho         VARCHAR(32)    NOT NULL COMMENT '创建人',
    editTime       DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    editWho        VARCHAR(32)    NOT NULL COMMENT '修改人',
    oprSeqFlag     VARCHAR(32)    NOT NULL COMMENT '操作序列标识',
    currentVersion INT            NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag     VARCHAR(1)     NOT NULL DEFAULT 'Y' COMMENT '活动状态标记：Y-活动，N-非活动',
    noteText       TEXT           NULL     COMMENT '备注信息',
    PRIMARY KEY (roleId, permCode, tenantId),
    INDEX IDX_ROLE_PERM_ROLE (roleId),
    INDEX IDX_ROLE_PERM_PERM (permCode),
    INDEX IDX_ROLE_PERM_TENANT (tenantId)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表';
```
