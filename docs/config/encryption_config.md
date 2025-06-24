# 加密配置文档

## 概述

本文档说明如何在不同环境下配置数据加密功能，支持开发环境禁用加密，生产环境启用加密。

## 配置项说明

### app.encryption_enabled
- **类型**: `boolean`
- **默认值**: `true`
- **说明**: 全局加密开关，控制是否启用加密功能

### app.encryption_key
- **类型**: `string`
- **默认值**: 空（使用默认密钥）
- **说明**: AES加密密钥，建议32字符以上，生产环境必须配置

### app.env
- **类型**: `string`
- **默认值**: `development`
- **可选值**: `development`, `dev`, `production`, `prod`, `testing`, `test`
- **说明**: 运行环境标识

### app.dev_disable_encryption
- **类型**: `boolean`
- **默认值**: `false`
- **说明**: 开发环境下是否禁用加密（仅在development环境生效）

## 环境配置示例

### 开发环境配置 (configs/development.yaml)

```yaml
app:
  env: development
  # 开发环境禁用加密，便于调试
  dev_disable_encryption: true
  # 或者启用加密但使用默认密钥
  # encryption_enabled: true
  # encryption_key: "dev-encryption-key-for-testing-32chars"
```

### 测试环境配置 (configs/testing.yaml)

```yaml
app:
  env: testing
  # 测试环境启用加密
  encryption_enabled: true
  encryption_key: "test-environment-encryption-key-32chars"
```

### 生产环境配置 (configs/production.yaml)

```yaml
app:
  env: production
  # 生产环境必须启用加密
  encryption_enabled: true
  # 生产环境必须配置强密钥
  encryption_key: "your-production-encryption-key-32chars-or-more"
```

## 使用方式

### 前端发送加密请求

当加密功能启用时，前端可以发送不同类型的加密请求：

#### 1. JSON格式 (application/json)
```javascript
// 请求头
headers: {
  'Content-Type': 'application/json',
  'X-Encrypted': 'true'  // 标记为加密请求
}

// 请求体
{
  "data": "base64加密数据",
  "iv": "base64初始化向量"
}
```

#### 2. Form URL编码格式 (application/x-www-form-urlencoded)
```javascript
// 请求头
headers: {
  'Content-Type': 'application/x-www-form-urlencoded',
  'X-Encrypted': 'true'  // 标记为加密请求
}

// 请求体 (URL编码格式)
"data=base64加密数据&iv=base64初始化向量"
```

#### 3. Multipart表单格式 (multipart/form-data)
```javascript
// 请求头
headers: {
  'Content-Type': 'multipart/form-data; boundary=----WebKitFormBoundary...',
  'X-Encrypted': 'true'  // 标记为加密请求
}

// 请求体 (multipart格式)
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="data"

base64加密数据
------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="iv"

base64初始化向量
------WebKitFormBoundary7MA4YWxkTrZu0gW--
```

### 前端接收加密响应

```javascript
// 请求头
headers: {
  'X-Encrypt-Response': 'true'  // 请求加密响应
}

// 响应体
{
  "encrypted": true,
  "data": "base64加密数据", 
  "iv": "base64初始化向量"
}
```

### 开发环境下的便利性

当 `app.dev_disable_encryption: true` 时：
- 前端可以直接发送明文JSON数据
- 后端直接返回明文JSON响应
- 无需处理加密/解密逻辑
- 便于调试和开发

## 安全建议

1. **生产环境**：
   - 必须设置 `encryption_enabled: true`
   - 必须配置强加密密钥 `encryption_key`
   - 密钥长度建议32字符以上
   - 定期更换加密密钥

2. **开发环境**：
   - 可以设置 `dev_disable_encryption: true` 便于调试
   - 如果启用加密，可以使用相对简单的密钥

3. **测试环境**：
   - 建议启用加密功能
   - 使用与生产环境相同的加密配置

## 环境变量配置

也可以通过环境变量进行配置：

```bash
# 设置环境
export APP_ENV=development

# 开发环境禁用加密
export APP_DEV_DISABLE_ENCRYPTION=true

# 设置加密密钥
export APP_ENCRYPTION_KEY=your-encryption-key
```

## 中间件工作流程

1. **请求处理**：
   - 检查 `isEncryptionEnabled()`
   - 如果禁用，直接跳过解密
   - 检查Content-Type是否支持 (JSON/form-urlencoded/multipart)
   - 如果启用且请求标记为加密，根据Content-Type进行解密处理

2. **响应处理**：
   - 检查 `isEncryptionEnabled()`
   - 如果禁用，直接跳过加密
   - 如果启用且请求要求加密响应，进行加密处理

## 支持的Content-Type

| Content-Type | 说明 | 加密数据格式 |
|-------------|------|------------|
| application/json | JSON格式 | `{"data": "...", "iv": "..."}` |
| application/x-www-form-urlencoded | 表单URL编码 | `data=...&iv=...` |
| multipart/form-data | 多部分表单 | 两个form字段: data和iv |

## 前端JavaScript示例

### 使用Fetch API发送加密请求

```javascript
// JSON格式
async function sendEncryptedJSON(url, plainData) {
  // 假设有加密函数 encryptData(data) 返回 {data, iv}
  const encrypted = await encryptData(JSON.stringify(plainData));
  
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Encrypted': 'true'
    },
    body: JSON.stringify(encrypted)
  });
  
  return response.json();
}

// Form URL编码格式
async function sendEncryptedForm(url, plainData) {
  const encrypted = await encryptData(JSON.stringify(plainData));
  
  const formData = new URLSearchParams();
  formData.append('data', encrypted.data);
  formData.append('iv', encrypted.iv);
  
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      'X-Encrypted': 'true'
    },
    body: formData.toString()
  });
  
  return response.json();
}

// Multipart表单格式
async function sendEncryptedMultipart(url, plainData) {
  const encrypted = await encryptData(JSON.stringify(plainData));
  
  const formData = new FormData();
  formData.append('data', encrypted.data);
  formData.append('iv', encrypted.iv);
  
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'X-Encrypted': 'true'
      // Content-Type会自动设置为multipart/form-data
    },
    body: formData
  });
  
  return response.json();
}
```

## 日志说明

- **开发环境禁用加密**：`DEBUG` 级别日志
- **使用默认密钥**：
  - 开发环境：`INFO` 级别
  - 生产环境：`ERROR` 级别
  - 其他环境：`WARN` 级别 