# 开发指南

本文档提供了 Gateway 项目的开发指南，重点介绍项目结构和开发规范。其他详细内容请参考相关文档。

## 📋 目录

- [项目结构说明](#项目结构说明)
- [开发规范](#开发规范)
- [相关资源](#相关资源)

## 📁 项目结构说明

### 目录结构

```
gateway/
├── api/                          # API相关文件
│   ├── proto/                    # Protocol Buffers定义
│   └── swagger/                  # Swagger API文档
├── bin/                          # 编译后的可执行文件
│   ├── gateway                   # 网关服务程序
│   └── test                      # 测试程序
├── cmd/                          # 应用程序入口
│   ├── app/                      # 主应用程序
│   ├── cache/                    # 缓存服务
│   ├── gateway/                  # 网关服务
│   └── web/                      # Web管理界面
├── configs/                      # 配置文件
│   ├── database.yaml             # 数据库配置
│   ├── gateway.yaml              # 网关配置
│   ├── logger.yaml               # 日志配置
│   └── web.yaml                  # Web服务配置
├── internal/                     # 内部包（不对外暴露）
│   ├── auth/                     # 认证相关
│   ├── cache/                    # 缓存实现
│   ├── common/                   # 通用组件
│   ├── discovery/                # 服务发现
│   ├── gateway/                  # 网关核心实现
│   │   ├── bootstrap/            # 网关启动器
│   │   ├── config/               # 网关配置
│   │   ├── core/                 # 核心引擎
│   │   ├── handler/              # 处理器集合
│   │   └── loader/               # 配置加载器
│   ├── metrics/                  # 监控指标
│   ├── model/                    # 数据模型
│   ├── service/                  # 业务服务
│   └── storage/                  # 存储层
├── pkg/                          # 公共包（可对外暴露）
│   ├── cache/                    # 缓存工具包
│   ├── config/                   # 配置工具包
│   ├── database/                 # 数据库工具包
│   ├── logger/                   # 日志工具包
│   └── utils/                    # 通用工具
├── test/                         # 测试文件
│   ├── cache/                    # 缓存测试
│   ├── database/                 # 数据库测试
│   ├── e2e/                      # 端到端测试
│   ├── gateway/                  # 网关测试
│   └── integration/              # 集成测试
├── web/                          # Web管理平台
│   ├── globalmodels/             # 全局数据模型
│   ├── middleware/               # Web中间件
│   ├── routes/                   # 全局路由配置
│   ├── utils/                    # Web工具包
│   └── views/                    # 业务模块视图
│       ├── hub0001/              # 认证授权模块
│       │   ├── controllers/      # 控制器层
│       │   ├── dao/              # 数据访问层
│       │   ├── models/           # 数据模型
│       │   └── routes/           # 路由配置
│       ├── hub0002/              # 用户管理模块
│       ├── hub0020/              # 网关实例管理模块
│       └── hubcommon002/         # 通用配置管理模块
├── go.mod                        # Go模块定义
├── go.sum                        # Go模块校验和
├── Makefile                      # 构建脚本
└── README.md                     # 项目说明文档
```

### 核心组件

#### 网关引擎 (`internal/gateway/core/`)
- **engine.go**: 主网关引擎
- **context.go**: 请求上下文管理
- **handler.go**: 请求处理器接口

#### 处理器 (`internal/gateway/handler/`)
- **auth/**: 认证处理器
- **proxy/**: 代理处理器
- **router/**: 路由处理器
- **filter/**: 请求/响应过滤器
- **cors/**: 跨域处理器
- **limiter/**: 限流处理器
- **circuitbreaker/**: 熔断处理器
- **assertion/**: 断言处理器

#### 配置管理 (`internal/gateway/config/`)
- **config.go**: 配置结构定义
- **loader/**: 配置加载器

#### Web 界面 (`web/`)
- **views/hub*/**: 业务模块
- **routes/**: 路由定义
- **middleware/**: Web 中间件

## 📝 开发规范

### 添加新的业务模块

1. **创建模块目录结构**
```bash
mkdir -p web/views/hubXXXX/{controllers,dao,models,routes}
```

2. **实现模块组件**
- **数据模型** (`models/`): 定义数据结构和表映射
- **数据访问层** (`dao/`): 实现CRUD操作
- **控制器** (`controllers/`): 实现RESTful API
- **路由配置** (`routes/`): 注册API路由

3. **注册模块**
```go
// 在 web/moduleimports/imports.go 中导入
import _ "gateway/web/views/hubXXXX/routes"
```

### 添加新的配置管理模块

1. **在通用配置模块中添加新配置类型**
```bash
# 创建配置文件
touch web/views/hubcommon002/models/new_config.go
touch web/views/hubcommon002/dao/new_config_dao.go
touch web/views/hubcommon002/controllers/new_config_controller.go
```

2. **遵循统一API规范**
```go
POST /gateway/hubcommon002/{config-type}/add     # 添加配置
POST /gateway/hubcommon002/{config-type}/get     # 获取配置
POST /gateway/hubcommon002/{config-type}/update  # 更新配置
POST /gateway/hubcommon002/{config-type}/delete  # 删除配置
POST /gateway/hubcommon002/{config-type}/query   # 查询配置列表
```

### 数据库表设计规范

#### 表命名规范
- 所有表名以 `HUB_` 开头
- 使用大写字母和下划线分隔
- 表名应具有描述性

#### 必需字段
所有配置表都应包含以下系统字段:
```sql
tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
addWho VARCHAR(32) NOT NULL COMMENT '创建人ID',
editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
editWho VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
oprSeqFlag VARCHAR(32) NOT NULL COMMENT '操作序列标识',
currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息'
```

### Go 代码风格

#### 命名规范
```go
// 包名：小写，简短，描述性
package router

// 接口名：以 'er' 结尾或具有描述性
type Handler interface {
    Handle(ctx *Context) error
}

// 结构体名：大驼峰命名
type GatewayConfig struct {
    Listen string `yaml:"listen"`
    Name   string `yaml:"name"`
}

// 函数名：公开函数大驼峰，私有函数小驼峰
func NewGateway(config *GatewayConfig) *Gateway {
    return &Gateway{config: config}
}

func (g *Gateway) handleRequest(ctx *Context) error {
    // 实现
}
```

#### 错误处理
```go
// 使用描述性错误消息
func validateConfig(config *Config) error {
    if config.Listen == "" {
        return fmt.Errorf("监听地址不能为空")
    }
    return nil
}

// 包装错误并添加上下文
func loadConfigFile(path string) (*Config, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
    }
    // 处理数据...
}

// 使用预定义错误
var (
    ErrConfigNotFound = errors.New("未找到配置")
    ErrInvalidRoute   = errors.New("无效的路由配置")
)
```

### 代码注释规范

```go
// Package router 提供 Gateway API 网关的 HTTP 请求路由功能
package router

// Router 根据配置规则处理 HTTP 请求路由
type Router struct {
    routes map[string]*Route
    config *RouterConfig
}

// NewRouter 创建一个新的路由器实例
// 如果配置无效则返回错误
func NewRouter(config *RouterConfig) (*Router, error) {
    // 实现
}

// 复杂逻辑注释
func (r *Router) matchRoute(path string) *Route {
    // 首先尝试精确匹配以提高性能
    if route, exists := r.routes[path]; exists {
        return route
    }
    
    // 回退到模式匹配
    for pattern, route := range r.routes {
        if matched, _ := filepath.Match(pattern, path); matched {
            return route
        }
    }
    
    return nil
}
```

## 🔗 相关资源

- [架构指南](architecture.md) - 系统架构概览
- [配置参考](configuration.md) - 完整配置指南
- [API 参考](../api/README.md) - API 文档
- [部署指南](deployment.md) - 部署策略
- [贡献指南](../../CONTRIBUTING.md) - 贡献流程 