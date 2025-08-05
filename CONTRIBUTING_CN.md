# Gateway 贡献指南

首先，感谢您对 Gateway 项目的关注和贡献意愿！本文档将帮助您了解如何高效、规范地为 Gateway 做出贡献。

## 📖 目录

1. [快速开始](#快速开始)
2. [开发规范](#开发规范)
3. [提交代码](#提交代码)
4. [测试要求](#测试要求)
5. [问题反馈](#问题反馈)
6. [文档贡献](#文档贡献)


## 🚀 快速开始

### 1. 准备代码

```bash
git clone https://github.com/your-username/gateway.git
cd gateway
git remote add upstream https://github.com/your-org/gateway.git
```

### 2. 开发环境

```bash
# 安装依赖
go mod download

# 启动开发模式
make dev
```

### 3. 提交代码

```bash
git commit -m "feat: 添加新功能"  # 遵循提交规范
git push origin feature/your-feature
```

### 4. 创建 Pull Request

在 GitHub 上创建 Pull Request，详细描述您的更改。

## 📝 开发规范

#### 代码风格

- **使用 `gofmt` 格式化代码**

- **遵循标准 Go 命名规范（如 `NewGateway` 而非 `new_gateway`）**

- **错误处理必须包含上下文信息**


#### 目录结构规范

```
gateway/
├── cmd/                   # 入口代码
├── internal/              # 核心实现
├── pkg/                   # 公共库
└── test/                  # 测试文件
```

## 💾 提交代码

### 提交消息格式

```
类型(范围): 描述

示例:
feat(gateway): 添加 HTTP 代理支持
fix(auth): 修复登录失败问题
```

### 提交类型：

| 类型 | 描述 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 | 
| `test` | 测试相关 |


## 🧪 测试规范

### 必须包含：

1. **单元测试（核心逻辑）**
2. **集成测试（关键流程）**

### 运行测试:

```bash
make test       # 运行全部测试
make test-unit  # 仅单元测试
```

## 🐛 问题反馈

### 提交 Issue 时请注明：

1. **环境信息（Go 版本等）**
2. **重现步骤**
3. **预期与实际行为**

## 📄 文档贡献

- **API 文档**
- **使用教程**
- **配置说明**

使用 Markdown 格式，中英文均可

---

## 💬 需要帮助？

如果您有任何问题或需要帮助，可以通过以下方式联系我们：

- 💬 GitHub Issues: [提交问题](https://github.com/your-org/gateway/issues)
- 📧 邮件联系: [project-email@example.com](mailto:project-email@example.com)
- 📱 微信群: [加入微信群聊]
- 🐧 QQ群: [加入QQ群聊]

再次感谢您的贡献！🎉 