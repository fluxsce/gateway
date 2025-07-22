# Gateway 贡献指南

首先，感谢您对 Gateway 项目的关注和贡献意愿！本文档将帮助您了解如何为 Gateway 项目做出贡献。

## 📖 目录

- [🚀 快速开始](#快速开始)
- [🏗️ 开发环境设置](#开发环境设置)
- [📝 开发规范](#开发规范)
- [🎯 Git 提交规范](#git-提交规范)
- [🔀 分支管理策略](#分支管理策略)
- [🧪 测试规范](#测试规范)
- [📋 代码审查流程](#代码审查流程)
- [🐛 Bug 报告](#bug-报告)
- [💡 功能请求](#功能请求)
- [📄 文档贡献](#文档贡献)
- [🏆 贡献者认可](#贡献者认可)

## 🚀 快速开始

### 1. Fork 项目

1. 访问 [Gateway 项目主页](https://github.com/your-org/gateway)
2. 点击右上角的 "Fork" 按钮
3. 克隆你 Fork 的仓库到本地

```bash
git clone https://github.com/your-username/gateway.git
cd gateway
git remote add upstream https://github.com/your-org/gateway.git
```

### 2. 创建开发分支

```bash
# 确保主分支是最新的
git checkout main
git pull upstream main

# 创建功能分支
git checkout -b feature/your-feature-name
```

### 3. 进行开发并提交

```bash
# 进行代码修改后
git add .
git commit -m "feat: add your feature description"
git push origin feature/your-feature-name
```

### 4. 创建 Pull Request

在 GitHub 上创建 Pull Request，详细描述您的更改。

## 🏗️ 开发环境设置

### 必需软件

- **Go**: 1.19+ (推荐最新稳定版本)
- **Git**: 2.20+
- **Make**: 构建工具
- **Docker**: 可选，用于容器化部署
- **MySQL**: 8.0+ (开发和测试)
- **Redis**: 6.0+ (可选)

### IDE 推荐

- **VS Code** + Go 扩展
- **GoLand** (JetBrains)
- **Vim/Neovim** + vim-go

### 开发环境配置

```bash
# 1. 安装依赖
go mod download

# 2. 设置开发环境变量
export GATEWAY_ENV=development

# 3. 初始化数据库
make db-init

# 4. 运行项目
make dev

# 5. 运行测试
make test
```

## 📝 开发规范

### 代码风格

#### Go 代码规范

我们遵循标准的 Go 代码风格指南：

1. **使用 gofmt 格式化代码**
```bash
go fmt ./...
```

2. **使用 goimports 管理导入**
```bash
goimports -w .
```

3. **使用 golangci-lint 进行静态检查**
```bash
golangci-lint run
```

#### 命名规范

**包名（Package）**
```go
// ✅ 推荐
package gateway
package handler
package config

// ❌ 不推荐
package gatewayPackage
package handlerUtils
```

**接口名（Interface）**
```go
// ✅ 推荐
type Reader interface {}
type Writer interface {}
type ConfigLoader interface {}

// ❌ 不推荐  
type IReader interface {}
type ReaderInterface interface {}
```

**结构体名（Struct）**
```go
// ✅ 推荐
type GatewayConfig struct {}
type HTTPProxy struct {}
type WebSocketConnection struct {}

// ❌ 不推荐
type gatewayConfig struct {}
type Http_Proxy struct {}
```

**函数名（Function）**
```go
// ✅ 推荐
func LoadConfig() {}
func NewHTTPProxy() {}
func (h *HTTPProxy) HandleRequest() {}

// ❌ 不推荐
func loadconfig() {}
func new_http_proxy() {}
func (h *HTTPProxy) handle_request() {}
```

**常量名（Constant）**
```go
// ✅ 推荐
const (
    DefaultTimeout = 30 * time.Second
    MaxRetryCount  = 3
    APIVersion     = "v1"
)

// ❌ 不推荐
const (
    default_timeout = 30 * time.Second
    MAX_RETRY_COUNT = 3
    apiversion      = "v1"
)
```

#### 目录结构规范

```
gateway/
├── cmd/                    # 应用程序入口点
│   ├── gateway/           # 网关服务
│   └── web/               # Web 管理界面
├── internal/              # 内部包（不可导入）
│   ├── gateway/           # 网关核心实现
│   │   ├── core/         # 核心引擎
│   │   ├── handler/      # 处理器
│   │   └── config/       # 配置管理
│   ├── auth/             # 认证授权
│   ├── storage/          # 存储层
│   └── service/          # 业务服务层
├── pkg/                   # 公共包（可导入）
│   ├── logger/           # 日志工具
│   ├── database/         # 数据库工具
│   └── utils/            # 通用工具
├── api/                   # API 定义
├── web/                   # Web 应用
├── docs/                  # 文档
├── scripts/               # 脚本
├── configs/               # 配置文件
└── test/                  # 测试文件
```

#### 错误处理规范

```go
// ✅ 推荐：包装错误并添加上下文
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("读取配置文件失败 %s: %w", path, err)
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("解析配置文件失败: %w", err)
    }
    
    return &config, nil
}

// ❌ 不推荐：忽略或简单返回错误
func LoadConfig(path string) (*Config, error) {
    data, _ := os.ReadFile(path) // 忽略错误
    var config Config
    yaml.Unmarshal(data, &config)
    return &config, nil
}
```

#### 日志记录规范

```go
// ✅ 推荐：结构化日志
logger.Info("处理请求",
    zap.String("method", req.Method),
    zap.String("path", req.URL.Path),
    zap.Duration("duration", duration),
    zap.String("remote_addr", req.RemoteAddr),
)

// ❌ 不推荐：格式化字符串
logger.Infof("处理请求 %s %s 耗时 %v 来自 %s", 
    req.Method, req.URL.Path, duration, req.RemoteAddr)
```

#### 配置管理规范

```go
// ✅ 推荐：使用结构体标签进行配置映射
type ServerConfig struct {
    Host         string        `yaml:"host" json:"host" mapstructure:"host"`
    Port         int           `yaml:"port" json:"port" mapstructure:"port"`
    ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout" mapstructure:"read_timeout"`
    WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout" mapstructure:"write_timeout"`
}

// 提供默认值
func DefaultServerConfig() ServerConfig {
    return ServerConfig{
        Host:         "0.0.0.0",
        Port:         8080,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    }
}
```

### 注释规范

```go
// Package gateway 实现高性能的微服务网关
// 
// 该包提供了完整的网关功能，包括：
//   - HTTP/HTTPS 代理
//   - WebSocket 代理
//   - 负载均衡
//   - 限流
//   - 认证授权
//
// 基本使用方法：
//   gateway := gateway.New(config)
//   gateway.Start()
package gateway

// GatewayConfig 网关配置
//
// 包含网关运行所需的所有配置项，支持从 YAML 文件或环境变量加载。
type GatewayConfig struct {
    // Server 服务器配置
    Server ServerConfig `yaml:"server" json:"server"`
    
    // Router 路由配置
    Router RouterConfig `yaml:"router" json:"router"`
}

// NewGateway 创建新的网关实例
//
// 参数：
//   - config: 网关配置，不能为 nil
//
// 返回值：
//   - *Gateway: 网关实例
//   - error: 创建过程中的错误
//
// 示例：
//   config := LoadConfig("gateway.yaml")
//   gateway, err := NewGateway(config)
//   if err != nil {
//       log.Fatal(err)
//   }
func NewGateway(config *GatewayConfig) (*Gateway, error) {
    // 实现...
}
```

## 🎯 Git 提交规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范。

### 提交消息格式

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### 提交类型（Type）

| 类型 | 描述 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat: 添加 WebSocket 代理支持` |
| `fix` | Bug 修复 | `fix: 修复内存泄漏问题` |
| `docs` | 文档更新 | `docs: 更新 API 文档` |
| `style` | 代码格式调整（不影响功能） | `style: 格式化代码` |
| `refactor` | 重构（不是新功能，也不是修复bug） | `refactor: 重构配置加载逻辑` |
| `test` | 添加或修改测试 | `test: 添加网关代理测试` |
| `chore` | 构建过程或辅助工具的变动 | `chore: 更新依赖版本` |
| `perf` | 性能优化 | `perf: 优化路由匹配算法` |
| `ci` | CI/CD 相关更改 | `ci: 添加 GitHub Actions 工作流` |
| `build` | 构建系统或外部依赖更改 | `build: 更新 Makefile` |
| `revert` | 回滚之前的提交 | `revert: 回滚 feat: 添加缓存功能` |

### 作用域（Scope）

建议的作用域：

- `gateway` - 网关核心
- `auth` - 认证授权
- `config` - 配置管理
- `handler` - 处理器
- `middleware` - 中间件
- `storage` - 存储层
- `web` - Web 界面
- `api` - API 接口
- `docs` - 文档
- `test` - 测试

### 提交消息示例

```bash
# 添加新功能
git commit -m "feat(gateway): 添加 WebSocket 代理支持"

# 修复 Bug
git commit -m "fix(auth): 修复 JWT 令牌验证失败问题"

# 文档更新
git commit -m "docs(api): 更新认证 API 文档"

# 重构代码
git commit -m "refactor(config): 使用 Viper 重构配置加载"

# 性能优化
git commit -m "perf(gateway): 优化请求路由匹配性能"

# 添加测试
git commit -m "test(handler): 添加限流处理器单元测试"
```

### 详细提交消息示例

```
feat(gateway): 添加 WebSocket 代理支持

- 实现 WebSocket 连接升级处理
- 添加双向消息代理转发
- 支持子协议和扩展协商
- 添加连接池管理和统计

Closes #123
```

### 破坏性变更（Breaking Changes）

```
feat(api)!: 重构认证 API 接口

BREAKING CHANGE: 认证接口响应格式发生变化
- 移除 `user` 字段
- 添加 `userInfo` 嵌套对象
- `token` 字段重命名为 `accessToken`

迁移指南请参考: docs/migration/v2.0.md
```

## 🔀 分支管理策略

我们使用 **Git Flow** 分支模型：

### 主要分支

- **`main`**: 主分支，始终保持可发布状态
- **`develop`**: 开发分支，包含最新的开发功能

### 辅助分支

- **`feature/*`**: 功能分支
  - 从 `develop` 分支创建
  - 完成后合并回 `develop`
  - 命名格式：`feature/feature-name`

- **`release/*`**: 发布分支
  - 从 `develop` 分支创建
  - 完成后合并到 `main` 和 `develop`
  - 命名格式：`release/v1.2.0`

- **`hotfix/*`**: 热修复分支
  - 从 `main` 分支创建
  - 完成后合并到 `main` 和 `develop`
  - 命名格式：`hotfix/fix-critical-bug`

### 分支操作示例

```bash
# 创建功能分支
git checkout develop
git pull origin develop
git checkout -b feature/websocket-proxy

# 开发完成后
git checkout develop
git pull origin develop
git merge --no-ff feature/websocket-proxy
git push origin develop
git branch -d feature/websocket-proxy

# 创建发布分支
git checkout develop
git pull origin develop
git checkout -b release/v1.2.0

# 发布完成后
git checkout main
git merge --no-ff release/v1.2.0
git tag v1.2.0
git checkout develop
git merge --no-ff release/v1.2.0
git branch -d release/v1.2.0
```

## 🧪 测试规范

### 测试类型

1. **单元测试** - 测试单个函数或方法
2. **集成测试** - 测试组件间的交互
3. **端到端测试** - 测试完整的用户场景

### 测试文件结构

```
test/
├── unit/                  # 单元测试
│   ├── gateway/
│   ├── auth/
│   └── config/
├── integration/           # 集成测试
│   ├── api/
│   └── database/
└── e2e/                   # 端到端测试
    └── scenarios/
```

### 测试命名规范

```go
// 测试函数命名: Test + 被测试的函数名 + 测试场景
func TestLoadConfig_ValidFile_Success(t *testing.T) {}
func TestLoadConfig_InvalidPath_ReturnsError(t *testing.T) {}
func TestLoadConfig_MalformedYAML_ReturnsError(t *testing.T) {}

// 基准测试命名: Benchmark + 被测试的函数名
func BenchmarkRouteMatch(b *testing.B) {}

// 示例测试命名: Example + 被测试的函数名
func ExampleNewGateway() {}
```

### 测试用例编写规范

```go
func TestHTTPProxy_ProxyRequest_Success(t *testing.T) {
    // Arrange - 准备测试数据
    config := DefaultHTTPProxyConfig()
    proxy := NewHTTPProxy(config, mockServiceManager)
    
    req := httptest.NewRequest("GET", "/api/users", nil)
    rec := httptest.NewRecorder()
    ctx := core.NewContext(rec, req)
    
    // Act - 执行被测试的操作
    err := proxy.ProxyRequest(ctx, "http://backend-service")
    
    // Assert - 验证结果
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
    assert.Contains(t, rec.Body.String(), "expected-content")
}
```

### 运行测试

```bash
# 运行所有测试
make test

# 运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 运行端到端测试
make test-e2e

# 生成覆盖率报告
make test-coverage

# 运行基准测试
make benchmark
```

## 📋 代码审查流程

### Pull Request 要求

1. **PR 标题**：遵循提交消息规范
2. **描述清晰**：说明改动的内容和原因
3. **测试充分**：包含相应的测试用例
4. **文档更新**：如果需要，更新相关文档
5. **无冲突**：确保与目标分支无冲突

### PR 模板

```markdown
## 🎯 变更类型
- [ ] Bug 修复
- [ ] 新功能
- [ ] 破坏性变更
- [ ] 文档更新
- [ ] 代码重构
- [ ] 性能优化
- [ ] 测试添加

## 📝 变更描述
简要描述本次变更的内容和原因。

## 🧪 测试情况
- [ ] 添加了单元测试
- [ ] 添加了集成测试
- [ ] 手动测试通过
- [ ] 所有现有测试通过

## 📚 相关文档
如果有相关的 Issue、文档或其他 PR，请在此列出。

## 🔍 检查清单
- [ ] 代码遵循项目规范
- [ ] 提交消息遵循约定式提交规范
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 通过了所有 CI 检查
```

### 审查标准

代码审查重点关注：

1. **功能正确性**：代码是否实现了预期功能
2. **代码质量**：可读性、可维护性、性能
3. **安全性**：是否存在安全漏洞
4. **测试覆盖**：是否有足够的测试
5. **文档完整**：是否需要更新文档
6. **向后兼容**：是否破坏现有 API

## 🐛 Bug 报告

请使用 [Issue 模板](https://github.com/your-org/gateway/issues/new?template=bug_report.md) 报告 Bug。

### Bug 报告应包含

1. **环境信息**：操作系统、Go 版本、Gateway 版本
2. **重现步骤**：详细的重现步骤
3. **期望行为**：描述期望的正确行为
4. **实际行为**：描述实际发生的错误行为
5. **错误日志**：相关的错误日志或堆栈跟踪
6. **附加信息**：配置文件、屏幕截图等

## 💡 功能请求

请使用 [Feature Request 模板](https://github.com/your-org/gateway/issues/new?template=feature_request.md) 提出新功能请求。

### 功能请求应包含

1. **功能概述**：简要描述建议的功能
2. **使用场景**：描述什么情况下需要这个功能
3. **解决方案**：建议的实现方案
4. **替代方案**：其他可能的解决方案
5. **附加信息**：相关的参考资料、竞品分析等

## 📄 文档贡献

文档同样重要！我们欢迎以下类型的文档贡献：

1. **API 文档**：API 接口文档和示例
2. **用户指南**：安装、配置、使用指南
3. **开发文档**：架构设计、开发指南
4. **教程**：最佳实践、常见场景教程
5. **翻译**：文档的多语言版本

### 文档规范

- 使用 **Markdown** 格式
- 遵循 **中文文案排版指北**
- 添加适当的**目录结构**
- 包含**代码示例**和**屏幕截图**
- 保持文档的**时效性**

## 🏆 贡献者认可

我们重视每一位贡献者的付出！

### 贡献类型

- 💻 代码贡献
- 📝 文档改进
- 🐛 Bug 报告
- 💡 功能建议
- 🤔 问题解答
- 🌍 翻译贡献
- 🎨 设计贡献

### 认可方式

- **Contributors 页面**：在项目主页展示所有贡献者
- **Release Notes**：在版本发布说明中感谢贡献者
- **社区荣誉**：定期评选优秀贡献者

---

## 📞 联系我们

如果您有任何问题或需要帮助，可以通过以下方式联系我们：

- 📧 Email: [project-email@example.com](mailto:project-email@example.com)
- 💬 GitHub Issues: [项目 Issues](https://github.com/your-org/gateway/issues)
- 📱 微信群: [加入微信群聊]
- 🐧 QQ群: [加入QQ群聊]

感谢您对 Gateway 项目的贡献！🎉 