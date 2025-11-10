# FLUX Gateway - Debugging Guide

This document provides debugging methods and techniques for FLUX Gateway to help developers quickly locate and resolve issues.

---

## 📋 Table of Contents

- [Development Environment Debugging](#-development-environment-debugging)
- [Performance Analysis](#-performance-analysis)
- [Common Issues Troubleshooting](#-common-issues-troubleshooting)

---

## 🔧 Development Environment Debugging

### VSCode Debug Configuration

The project has pre-configured `.vscode/launch.json` with the following debug modes:

**1. Run Main Application**
```json
{
  "name": "Run Main Application",
  "type": "go",
  "request": "launch",
  "mode": "debug",
  "program": "${workspaceFolder}/cmd/app/main.go"
}
```

**2. Run with Debug Info**
```json
{
  "name": "Run with Debug Info",
  "type": "go",
  "request": "launch",
  "mode": "debug",
  "program": "${workspaceFolder}/cmd/app/main.go",
  "env": {
    "DEBUG": "true"
  }
}
```

### Using Breakpoints

**Set Breakpoint:**
- Click on the left margin of the code line number (red dot appears)
- Shortcut: `F9`

**Conditional Breakpoint:**
- Right-click breakpoint → Edit Breakpoint → Add condition expression
- Example: `userId == "123"` or `count > 10`

**Log Point:**
- Right-click breakpoint → Edit Breakpoint → Add message
- Won't pause program, only outputs logs

### Debug Shortcuts

| Action | Windows/Linux | Mac |
|--------|---------------|-----|
| Start Debugging | `F5` | `F5` |
| Continue | `F5` | `F5` |
| Step Over | `F10` | `F10` |
| Step Into | `F11` | `F11` |
| Step Out | `Shift+F11` | `Shift+F11` |
| Restart | `Ctrl+Shift+F5` | `Cmd+Shift+F5` |
| Stop | `Shift+F5` | `Shift+F5` |

---

## 📊 Performance Analysis

### Enable pprof

Gateway has built-in pprof performance analysis tool.

**Configuration** (`configs/app.yaml`):
```yaml
pprof:
  enabled: true
  port: 6060
```

**Access after startup:**
```
http://localhost:6060/debug/pprof/
```

### CPU Performance Analysis

```bash
# Collect 30 seconds of CPU performance data
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Generate flame graph
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
```

### Memory Analysis

```bash
# View current memory usage
go tool pprof http://localhost:6060/debug/pprof/heap

# Generate memory analysis report
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

### Goroutine Analysis

```bash
# View all goroutines
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# Analyze goroutine leaks
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

---

## 🐛 Common Issues Troubleshooting

### 1. Application Won't Start

**Problem:**
```
Error: failed to connect to database
```

**Troubleshooting Steps:**

1. **Check Database Configuration**
   ```bash
   cat configs/database.yaml
   ```

2. **Test Database Connection**
   ```bash
   # MySQL
   mysql -h localhost -P 3306 -u gateway -p

   # Or use telnet
   telnet localhost 3306
   ```

3. **Check Database Service Status**
   ```bash
   # Check MySQL process
   ps aux | grep mysql
   
   # Check port listening
   netstat -tulpn | grep 3306
   ```

---

### 2. Memory Leak

**Symptoms:**
- Memory continuously growing
- OOM (Out of Memory) errors

**Troubleshooting Steps:**

1. **View Memory Usage**
   ```bash
   # View process memory usage
   ps aux | grep gateway
   
   # Use top command
   top -p $(pgrep gateway)
   ```

2. **Analyze Memory Snapshot**
   ```bash
   go tool pprof http://localhost:6060/debug/pprof/heap
   ```

3. **View Goroutine Count**
   ```bash
   curl http://localhost:6060/debug/pprof/goroutine?debug=1 | grep goroutine
   ```

---

### 3. Performance Issues

**Symptoms:**
- Slow response time
- High CPU usage

**Troubleshooting Steps:**

1. **View CPU Usage**
   ```bash
   # View process CPU usage
   ps aux | grep gateway
   
   # Use top command
   top -p $(pgrep gateway)
   ```

2. **CPU Performance Analysis**
   ```bash
   go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
   ```

3. **View Slow Query Logs**
   ```bash
   grep "slow query" logs/debug.log
   ```

---

### 4. Network Connection Issues

**Problem:**
```
Error: dial tcp: connect: connection refused
```

**Troubleshooting Steps:**

1. **Check Port Listening**
   ```bash
   # Check port listening status
   netstat -tulpn | grep 8080
   
   # Or use ss command
   ss -tulpn | grep 8080
   ```

2. **Test Port Connectivity**
   ```bash
   telnet localhost 8080
   nc -zv localhost 8080
   ```

3. **Check Firewall Rules**
   ```bash
   # Linux
   iptables -L -n

   # Check SELinux
   getenforce
   ```

---

### 5. Configuration File Issues

**Problem:**
```
Error: failed to load config file
```

**Troubleshooting Steps:**

1. **Check if Configuration Files Exist**
   ```bash
   ls -la configs/
   ```

2. **Validate YAML Syntax**
   ```bash
   # Use yamllint
   yamllint configs/database.yaml

   # Or use Python
   python -c "import yaml; yaml.safe_load(open('configs/database.yaml'))"
   ```

3. **Check File Permissions**
   ```bash
   ls -l configs/
   chmod 644 configs/*.yaml
   ```

---

## 📖 Next Steps

After mastering debugging techniques, we recommend continuing with:

- [Project Introduction](./01-introduction.md) - Understand project architecture and core capabilities
- [Development Guide](./02-quick-start.md) - Development environment setup and configuration
- [Error Handling](./07-error-handling.md) - Error handling best practices

---

## 🤝 Get Help

If you encounter debugging issues, you can get help through:

- 📧 Email: fluxopensource@flux.com.cn
- 💬 GitHub Issues: [Submit Issue](https://github.com/fluxsce/gateway/issues)

---

**[Back to Directory](./README.md) • [Previous: Database Specifications](./05-database-specs.md) • [Next: Error Handling](./07-error-handling.md)**

---

<div align="center">

Made with ❤️ by FLUX Gateway Team

</div>

