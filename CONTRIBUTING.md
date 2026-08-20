# Gateway Contribution Guide

First of all, thank you for your interest in the Gateway project and your willingness to contribute!
This document will help you understand how to contribute to Gateway efficiently and in a standardized manner.

## 📖 Table of Contents

- [🚀 Quick Start](#Quick Start)
- [🏗️ Development Environment Setup](#Development Environment Setup)
- [📝 Development Standards](#Development Standards)
- [🎯 Git Commit Standards](#git-Commit Standards)
- [🔀 Branch Management Strategy](#Branch Management Strategy)
- [🧪 Testing Standards](#Testing Standards)
- [📋 Code Review Process](#Code Review Process)
- [🐛 Bug Reports](#bug-Reports)
- [💡 Feature Requests](#Feature Requests)
- [📄 Documentation Contributions](#Documentation Contributions)
- [🏆 Contributor Recognition](#Contributor Recognition)

## 🚀 Quick Start

### 1. Fork and Clone the Project

1. Visit the [Gateway project homepage](https://github.com/fluxsce/gateway)
2. Click the "Fork" button in the top right corner
3. Clone your forked repository locally

```bash
git clone https://github.com/your-username/gateway.git
cd gateway
git remote add upstream https://github.com/fluxsce/gateway.git
```

### 2. Create a Development Branch

```bash
# Ensure the main branch is up-to-date
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name
```

### 3. Develop and Commit

```bash
# After making code changes
git add .
git commit -m "feat: add your feature description"
git push origin feature/your-feature-name
```

### 4. Create a Pull Request

Create a Pull Request on GitHub, describing your changes in detail.

## 🏗️ Development Environment Setup

### Required Software

- **Go**: 1.24+
- **Git**: 2.20+
- **C compiler**: required (CGO / SQLite)
- **Docker**: optional
- **Database**: SQLite by default; MySQL / Oracle optional in production

### Recommended IDEs

- **VS Code** + Go extension
- **GoLand** (JetBrains)
- **Vim/Neovim** + vim-go

### Development Environment Configuration

```bash
# 1. Install dependencies
go mod download

# 2. Development environment (optional)
export GATEWAY_ENV=development

# 3. Database: default is SQLite. First start runs scripts/db automatically.
#    File: ./scripts/data/gateway.db  (see configs/database.yaml)

# 4. Run from the repo root
go run cmd/app/main.go --config ./configs

# 5. Tests
go test ./...
```

Console: `http://localhost:12003/gatewayweb`. Health: `http://localhost:12003/health`. Gateway listen: `:8080`. Default login `admin` / `123456`. Source builds need a C compiler (CGO / SQLite) and a frontend `dist` or `pnpm run dev:vite`; see [docs/en/02-quick-start.md](docs/en/02-quick-start.md).

## 📝 Development Standards

### Code Style

#### Go Code Standards

We follow standard Go code style guidelines:：

1. **Format code with gofmt**
```bash
go fmt ./...
```

2. **Manage imports with goimports**
```bash
goimports -w .
```

3. **Static analysis with golangci-lint**
```bash
golangci-lint run
```

#### Naming Conventions

**Package Names**
```go
// ✅ Recommended
package gateway
package handler
package config

// ❌ Not recommended
package gatewayPackage
package handlerUtils
```

**Interface Names**
```go
// ✅ Recommended
type Reader interface {}
type Writer interface {}
type ConfigLoader interface {}

// ❌ Not recommended  
type IReader interface {}
type ReaderInterface interface {}
```

**Struct Names**
```go
// ✅ Recommended
type GatewayConfig struct {}
type HTTPProxy struct {}
type WebSocketConnection struct {}

// ❌ Not recommended
type gatewayConfig struct {}
type Http_Proxy struct {}
```

**Function Names**
```go
// ✅ Recommended
func LoadConfig() {}
func NewHTTPProxy() {}
func (h *HTTPProxy) HandleRequest() {}

// ❌ Not recommended
func loadconfig() {}
func new_http_proxy() {}
func (h *HTTPProxy) handle_request() {}
```

**Constant Names**
```go
// ✅ Recommended
const (
    DefaultTimeout = 30 * time.Second
    MaxRetryCount  = 3
    APIVersion     = "v1"
)

// ❌ Not recommended
const (
    default_timeout = 30 * time.Second
    MAX_RETRY_COUNT = 3
    apiversion      = "v1"
)
```

#### Directory Structure Standards

```
gateway/
├── cmd/                    # Application entry points
│   ├── gateway/           # Gateway service
│   └── web/               # Web management interface
├── internal/              # Internal packages (not importable)
│   ├── gateway/           # Gateway core implementation
│   │   ├── core/         # Core engine
│   │   ├── handler/      # Handlers
│   │   └── config/       # Configuration management
│   ├── auth/             # Authentication and authorization
│   ├── storage/          # Storage layer
│   └── service/          # Business service layer
├── pkg/                   # Public packages (importable)
│   ├── logger/           # Logging utilities
│   ├── database/         # Database utilities
│   └── utils/            # General utilities
├── api/                   # API definitions
├── web/                   # Web application
├── docs/                  # Documentation
├── scripts/               # Scripts
├── configs/               # Configuration files
└── test/                  # Test files
```

#### Error Handling Standards

```go
// ✅ Recommended：Wrap errors and add context
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %w", err)
    }
    
    return &config, nil
}

// ❌ Not recommended：Ignore or simply return errors
func LoadConfig(path string) (*Config, error) {
    data, _ := os.ReadFile(path) // Ignore error
    var config Config
    yaml.Unmarshal(data, &config)
    return &config, nil
}
```

#### Logging Standards

```go
// ✅ Recommended：Structured logging
logger.Info("Handling request",
    zap.String("method", req.Method),
    zap.String("path", req.URL.Path),
    zap.Duration("duration", duration),
    zap.String("remote_addr", req.RemoteAddr),
)

// ❌ Not recommended：Formatted strings
logger.Infof("Handling request %s %s took %v from %s", 
    req.Method, req.URL.Path, duration, req.RemoteAddr)
```

#### Configuration Management Standards

```go
// ✅ Recommended：Use struct tags for configuration mapping
type ServerConfig struct {
    Host         string        `yaml:"host" json:"host" mapstructure:"host"`
    Port         int           `yaml:"port" json:"port" mapstructure:"port"`
    ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout" mapstructure:"read_timeout"`
    WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout" mapstructure:"write_timeout"`
}

// Provide default values
func DefaultServerConfig() ServerConfig {
    return ServerConfig{
        Host:         "0.0.0.0",
        Port:         8080,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    }
}
```

### Commenting Standards

```go
// Package gateway implements a high-performance microservice gateway
// 
// This package provides complete gateway functionality including：
//   - HTTP/HTTPS proxy
//   - WebSocket proxy
//   - Load balancing
//   - Rate limiting
//   - Authentication and authorization
//
// Basic usage：
//   gateway := gateway.New(config)
//   gateway.Start()
package gateway

// GatewayConfig gateway configuration
//
// Contains all configuration items required for gateway operation,
// supports loading from YAML files or environment variables.
type GatewayConfig struct {
    // Server server configuration
    Server ServerConfig `yaml:"server" json:"server"`
    
    // Router routing configuration
    Router RouterConfig `yaml:"router" json:"router"`
}

// NewGateway creates a new gateway instance
//
// Parameters：
//   - config: Gateway configuration, cannot be nil
//
// Returns：
//   - *Gateway: Gateway instance
//   - error: Error during creation
//
// Example：
//   config := LoadConfig("gateway.yaml")
//   gateway, err := NewGateway(config)
//   if err != nil {
//       log.Fatal(err)
//   }
func NewGateway(config *GatewayConfig) (*Gateway, error) {
    // Implementation...
}
```

## 🎯 Git Commit Standards

We use [Conventional Commits](https://www.conventionalcommits.org/) standards.

### Commit Message Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Commit Types

| Type | Description | Example |
|------|------|------|
| `feat` | New feature | `feat: add WebSocket proxy support` |
| `fix` | Bug fix | `fix: fix memory leak issue` |
| `docs` | Documentation updates | `docs: update API documentation` |
| `style` | Code formatting changes（no functional impact） | `style: format code` |
| `refactor` | Refactoring（not a new feature or bug fix） | `refactor: refactor config loading logic` |
| `test` | 	Add or modify tests | `test: add gateway proxy tests` |
| `chore` | Build process or tooling changes | `chore: update dependency versions` |
| `perf` | Performance optimization | `perf: optimize route matching algorithm` |
| `ci` | CI/CD related changes | `ci: add GitHub Actions workflow` |
| `build` | Build system or external dependency changes | `build: update Makefile` |
| `revert` | Revert a previous commit | `revert: revert feat: add caching feature` |

### Scope

Suggested scopes：

- `gateway` - Gateway core
- `auth` - Authentication and authorization
- `config` - Configuration management
- `handler` - Handlers
- `middleware` - Middleware
- `storage` - Storage layer
- `web` - Web interface
- `api` - API interfaces
- `docs` - Documentation
- `test` - Tests

### Commit Message Examples

```bash
# Add new feature
git commit -m "feat(gateway): add WebSocket proxy support"

# Fix Bug
git commit -m "fix(auth): fix JWT token validation failure"

# Documentation update
git commit -m "docs(api): update authentication API documentation"

# Refactor code
git commit -m "refactor(config): refactor config loading using Viper"

# Performance optimization
git commit -m "perf(gateway): optimize request route matching performance"

# Add tests
git commit -m "test(handler): add rate limiting handler unit tests"
```

### Detailed Commit Message Example

```
feat(gateway): add WebSocket proxy support

- Implement WebSocket connection upgrade handling
- Add bidirectional message proxy forwarding
- Support subprotocol and extension negotiation
- Add connection pool management and statistics

Closes #123
```

### Breaking Changes

```
feat(api)!: refactor authentication API interface

BREAKING CHANGE: Authentication interface response format changed
- Remove `user` field
- Add `userInfo` nested object
- Rename `token` field to `accessToken`

Migration guide: docs/migration/v2.0.md
```

## 🔀 Branch Management Strategy

We use the **Git Flow** branching model：

### Main Branches

- **`main`**: Main branch, always in a releasable state
- **`develop`**: Development branch, contains latest development features

### Supporting Branches

- **`feature/*`**: Feature branches
  - Created from `develop` when complete
  - Merged back into `develop`
  - Naming format：`feature/feature-name`

- **`release/*`**: Release branches
  - Created from `develop`
  - Merged into `main` and `develop` when complete
  - Naming format：`release/v1.2.0`

- **`hotfix/*`**: Hotfix branches
  - Created from `main`
  - Merged into `main` and `develop` when complete
  - Naming format：`hotfix/fix-critical-bug`

### Branch Operation Examples

```bash
# Create feature branch
git checkout develop
git pull origin develop
git checkout -b feature/websocket-proxy

# After development
git checkout develop
git pull origin develop
git merge --no-ff feature/websocket-proxy
git push origin develop
git branch -d feature/websocket-proxy

# Create release branch
git checkout develop
git pull origin develop
git checkout -b release/v1.2.0

# After release
git checkout main
git merge --no-ff release/v1.2.0
git tag v1.2.0
git checkout develop
git merge --no-ff release/v1.2.0
git branch -d release/v1.2.0
```

## 🧪 Testing Standards

### Test Types

1. **Unit Tests** - Test individual functions or methods
2. **Integration Tests** - Test interactions between components
3. **End-to-End Tests** - Test complete user scenarios

### Test File Structure

```
test/
├── unit/                  # Unit tests
│   ├── gateway/
│   ├── auth/
│   └── config/
├── integration/           # Integration tests
│   ├── api/
│   └── database/
└── e2e/                   # End-to-end tests
    └── scenarios/
```

### Test Naming Standards

```go
// Test function naming: Test + Function being tested + Test scenario
func TestLoadConfig_ValidFile_Success(t *testing.T) {}
func TestLoadConfig_InvalidPath_ReturnsError(t *testing.T) {}
func TestLoadConfig_MalformedYAML_ReturnsError(t *testing.T) {}

// Benchmark naming: Benchmark + Function being tested
func BenchmarkRouteMatch(b *testing.B) {}

// Example test naming: Example + Function being tested
func ExampleNewGateway() {}
```

### Test Case Writing Standards

```go
func TestHTTPProxy_ProxyRequest_Success(t *testing.T) {
    // Arrange - Prepare test data
    config := DefaultHTTPProxyConfig()
    proxy := NewHTTPProxy(config, mockServiceManager)
    
    req := httptest.NewRequest("GET", "/api/users", nil)
    rec := httptest.NewRecorder()
    ctx := core.NewContext(rec, req)
    
    // Act - Execute the operation being tested
    err := proxy.ProxyRequest(ctx, "http://backend-service")
    
    // Assert - Verify results
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
    assert.Contains(t, rec.Body.String(), "expected-content")
}
```

### Running Tests

```bash
# All packages
go test ./...

# One package, verbose
go test -v ./pkg/utils/huberrors

# Coverage (example)
go test ./... -coverprofile=coverage.out
```

There is no `make test-unit` / `make test-e2e` target in the repo Makefile. Use `go test` on the package you changed.

## 📋 Code Review Process

### Pull Request Requirements

1. **PR Title**：Follow commit message standards
2. **Clear Description**：Explain what was changed and why
3. **Adequate Testing**：Include corresponding test cases
4. **Documentation Updates**：Update relevant documentation if needed
5. **No Conflicts**：Ensure no conflicts with target branch

### PR Template

```markdown
## 🎯 Change Type
- [ ] Bug Fix
- [ ] New Feature
- [ ] Breaking Change
- [ ] Documentation Update
- [ ] Code Refactor
- [ ] Performance Optimization
- [ ] Test Addition

## 📝 Change Description
Briefly describe the content and reason for this change.

## 🧪 Testing Status
- [ ] Added unit tests
- [ ] Added integration tests
- [ ] Manual testing passed
- [ ] All existing tests passed

## 📚 Related Documentation
List any related Issues, documentation, or other PRs here.

## 🔍 Checklist
- [ ] Code follows project standards
- [ ] Commit messages follow conventional commit standards
- [ ] Added necessary tests
- [ ] Updated relevant documentation
- [ ] Passed all CI checks
```

### Review Criteria

Code reviews focus on：

1. **Functional Correctness**：Does the code implement the expected functionality
2. **Code Quality**：Readability, maintainability, performance
3. **Security**：Are there any security vulnerabilities
4. **Test Coverage**：Are there sufficient tests
5. **Documentation Completeness**：Does documentation need updating
6. **Backward Compatibility**：Does it break existing APIs

## 🐛 Bug Reports

Please use the [Issue template](https://github.com/fluxsce/gateway/issues/new?template=bug_report.md) to report bugs.

### Bug Reports Should Include

1. **Environment Information**：OS, Go version, Gateway version
2. **Reproduction Steps**：Detailed steps to reproduce
3. **Expected Behavior**：Description of expected correct behavior
4. **Actual Behavior**：Description of actual erroneous behavior
5. **Error Logs**：Relevant error logs or stack traces
6. **Additional Information**：Configuration files, screenshots, etc.

## 💡 Feature Requests

Please use the [Feature Request template](https://github.com/fluxsce/gateway/issues/new?template=feature_request.md) to propose new features.

### Feature Requests Should Include

1. **Feature Overview**：Brief description of the proposed feature
2. **Use Case**：Describe when this feature would be needed
3. **Solution**：Suggested implementation approach
4. **Alternatives**：Other possible solutions
5. **Additional Information**：Relevant references, competitive analysis, etc.

## 📄 Documentation Contributions

Documentation is equally important! We welcome the following types of documentation contributions：

1. **API Documentation**：API interface documentation and examples
2. **User Guides**：Installation, configuration, usage guides
3. **Development Documentation**：Architecture design, development guides
4. **Tutorials**：Best practices, common scenario tutorials
5. **Translations**：Multilingual versions of documentation

### Documentation Standards

- Use **Markdown** format
- Follow **Chinese Copywriting Guidelines**
- Add appropriate **table of contents**
- Include**code examples**and**screenshots**
- Maintain documentation**timeliness**

## 🏆 Contributor Recognition

We value every contributor's efforts!

### Contribution Types

- 💻 Code development
- 📝 Documentation maintenance
- 🐛 Issue reporting
- 💡 New ideas
- 🌍 Translation support


### Recognition Methods

- **Contributors Page**：Show all contributors on the project homepage
- **Release Notes**：Thank contributors in version release notes
- **Community Honors**：Regularly recognize outstanding contributors

---

## 📞 Contact Us

If you have any questions or need help, you can contact us through：

- 📧 Email: [project-email@example.com](mailto:project-email@example.com)
- 💬 GitHub Issues: [Project Issues](https://github.com/fluxsce/gateway/issues)
- 📱 WeChat Group: [Join WeChat group]
- 🐧 QQ Group: [Join QQ group]

Thank you again for your contributions！🎉 