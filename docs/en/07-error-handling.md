# FLUX Gateway - Error Handling

This document introduces FLUX Gateway's error handling mechanisms and best practices to help developers write robust code.

---

## 📋 Table of Contents

- [Error Handling Mechanism](#-error-handling-mechanism)
- [HubError Usage](#-huberror-usage)
- [HTTP Error Responses](#-http-error-responses)
- [Best Practices](#-best-practices)
- [Common Error Scenarios](#-common-error-scenarios)

---

## 🔧 Error Handling Mechanism

### Standard Error Handling

Gateway uses Go's standard error handling approach:

```go
import (
    "errors"
    "fmt"
)

// Create simple error
err := errors.New("config file not found")

// Format error
err := fmt.Errorf("cannot connect to database: %s", dbHost)

// Wrap error
if err != nil {
    return fmt.Errorf("initialization failed: %w", err)
}
```

### Error Checking

```go
import "errors"

// Check error type
if errors.Is(err, os.ErrNotExist) {
    // File does not exist
}

// Check error value
var pathError *os.PathError
if errors.As(err, &pathError) {
    // Handle path error
}
```

---

## 🎯 HubError Usage

Gateway provides an enhanced error handling mechanism `huberrors` that creates errors with precise location information.

### Features

1. ✅ Automatically captures **exact file name, line number, and function name** where error occurred
2. ✅ Deep integration with logging system, provides complete error stack
3. ✅ Supports wrapping original errors, preserves complete error chain
4. ✅ Error message formatting support, compatible with `fmt.Errorf` parameter format

### Import Package

```go
import (
    "github.com/fluxsce/gateway/pkg/utils/huberrors"
)
```

### Create New Error

```go
// Create error with location info (similar to errors.New)
err := huberrors.NewError("an error occurred")

// Support formatting (similar to fmt.Errorf)
err := huberrors.NewError("cannot parse config: %s", configName)

// Error output example:
// cannot parse config: database.yaml (at /path/to/file.go:25 in package.function)
```

### Wrap Existing Error

```go
// Wrap existing error and add context
originalErr := someFunction()
wrappedErr := huberrors.WrapError(originalErr, "operation failed")

// Support formatting
wrappedErr := huberrors.WrapError(originalErr, "failed processing %s", resourceName)

// Error output example:
// failed processing users (at /path/to/file.go:42 in package.function): original error message
```

### Integration with Logging System

```go
err := huberrors.NewError("database connection failed")
logger.Error("system startup failed", err)
```

**Log output will include:**
- Error message
- Error type
- Complete file path, line number, and function name
- Stack trace

---

## 🌐 HTTP Error Responses

### Standard Error Response Format

```go
type ErrorResponse struct {
    Code    int    `json:"code"`    // Business error code
    Message string `json:"message"` // Error description
    Details string `json:"details"` // Detailed information (optional)
}
```

### Common HTTP Status Codes

| Status Code | Description | Use Case |
|-------------|-------------|----------|
| **400** | Bad Request | Request parameter error |
| **401** | Unauthorized | Not authenticated |
| **403** | Forbidden | No permission |
| **404** | Not Found | Resource not found |
| **409** | Conflict | Resource conflict |
| **429** | Too Many Requests | Too frequent requests |
| **500** | Internal Server Error | Server internal error |
| **502** | Bad Gateway | Backend service error |
| **503** | Service Unavailable | Service unavailable |
| **504** | Gateway Timeout | Gateway timeout |

### Error Response Examples

```go
// Parameter error
c.JSON(http.StatusBadRequest, ErrorResponse{
    Code:    40001,
    Message: "Parameter error",
    Details: "userId cannot be empty",
})

// Not authenticated
c.JSON(http.StatusUnauthorized, ErrorResponse{
    Code:    40101,
    Message: "Not authenticated",
    Details: "Token has expired",
})

// Resource not found
c.JSON(http.StatusNotFound, ErrorResponse{
    Code:    40401,
    Message: "Resource not found",
    Details: "User ID 123 does not exist",
})

// Server error
c.JSON(http.StatusInternalServerError, ErrorResponse{
    Code:    50001,
    Message: "Server internal error",
    Details: "Database connection failed",
})
```

---

## ✅ Best Practices

### 1. Use NewError at Error Origin

Use `NewError` where the error first occurs to ensure capturing the most precise location information.

```go
func readConfig(path string) (Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        // ✅ Wrap error at origin
        return nil, huberrors.WrapError(err, "failed to read config file: %s", path)
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        // ✅ Wrap error at origin
        return nil, huberrors.WrapError(err, "failed to parse config file: %s", path)
    }
    
    if !config.IsValid() {
        // ✅ Create new error
        return nil, huberrors.NewError("invalid config: %s", path)
    }
    
    return config, nil
}
```

### 2. Use WrapError in Middle Layers

When passing errors up, use `WrapError` to add context information.

```go
func initDatabase() error {
    config, err := readConfig("database.yaml")
    if err != nil {
        // ✅ Wrap error, add context
        return huberrors.WrapError(err, "database initialization failed")
    }
    
    db, err := connectDB(config)
    if err != nil {
        // ✅ Wrap error, add context
        return huberrors.WrapError(err, "database connection failed")
    }
    
    return nil
}
```

### 3. Keep Error Messages Concise and Clear

Error messages should clearly describe what happened, without too many technical details.

```go
// ✅ Good error messages
err := huberrors.NewError("config file not found")
err := huberrors.NewError("database connection failed")
err := huberrors.NewError("username or password incorrect")

// ❌ Bad error messages
err := huberrors.NewError("os.Open() returned error: file not found at line 42")
err := huberrors.NewError("SQL error 1045: Access denied for user 'root'@'localhost'")
```

### 4. Avoid Repeatedly Creating Errors

If you already have a `HubError` with location info, use `WrapError` instead of creating a new error.

```go
// ✅ Correct approach
func processUser(id string) error {
    user, err := getUser(id)
    if err != nil {
        return huberrors.WrapError(err, "failed to process user")
    }
    // ...
}

// ❌ Wrong approach
func processUser(id string) error {
    user, err := getUser(id)
    if err != nil {
        return huberrors.NewError("failed to process user: %v", err) // Loses original error location
    }
    // ...
}
```

---

## 🐛 Common Error Scenarios

### 1. Configuration File Errors

```go
func loadConfig(path string) (*Config, error) {
    // Check if file exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, huberrors.NewError("config file does not exist: %s", path)
    }
    
    // Read file
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, huberrors.WrapError(err, "failed to read config file: %s", path)
    }
    
    // Parse YAML
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, huberrors.WrapError(err, "failed to parse config file: %s", path)
    }
    
    // Validate config
    if err := config.Validate(); err != nil {
        return nil, huberrors.WrapError(err, "config validation failed: %s", path)
    }
    
    return &config, nil
}
```

### 2. Database Errors

```go
func getUser(id string) (*User, error) {
    var user User
    
    // Query database
    err := db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user)
    
    // Handle different types of errors
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, huberrors.NewError("user does not exist: %s", id)
        }
        return nil, huberrors.WrapError(err, "failed to query user: %s", id)
    }
    
    return &user, nil
}
```

### 3. HTTP Request Errors

```go
func callExternalAPI(url string) ([]byte, error) {
    // Make request
    resp, err := http.Get(url)
    if err != nil {
        return nil, huberrors.WrapError(err, "HTTP request failed: %s", url)
    }
    defer resp.Body.Close()
    
    // Check status code
    if resp.StatusCode != http.StatusOK {
        return nil, huberrors.NewError("HTTP request failed: %s, status code: %d", url, resp.StatusCode)
    }
    
    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, huberrors.WrapError(err, "failed to read response: %s", url)
    }
    
    return body, nil
}
```

---

## 📖 Next Steps

After mastering error handling, we recommend continuing with:

- [Project Introduction](./01-introduction.md) - Understand project architecture and core capabilities
- [Development Guide](./02-quick-start.md) - Development environment setup and configuration
- [Debugging Guide](./06-debugging.md) - Debugging techniques and methods

---

## 🤝 Get Help

If you encounter error handling related issues, you can get help through:

- 📧 Email: fluxopensource@flux.com.cn
- 💬 GitHub Issues: [Submit Issue](https://github.com/fluxsce/gateway/issues)

---

**[Back to Directory](./README.md) • [Previous: Debugging Guide](./06-debugging.md)**

---

<div align="center">

Made with ❤️ by FLUX Gateway Team

</div>

