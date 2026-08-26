# Gateway 构建脚本

本地打包用本目录脚本。GitHub Actions 发版**不调用**这些脚本，产物说明见 [.github/CI.md](../../.github/CI.md)。

## 脚本

| 脚本 | 平台 | 默认 |
|------|------|------|
| `build-win10.cmd` | Windows 10/11 amd64 | **无 Oracle**（`--oracle` 开启） |
| `build-centos7.sh` | Linux amd64，glibc 兼容 CentOS 7 | **无 Oracle**（`--oracle` 开启） |
| `build-win2008-oracle.cmd` | Windows Server 2008 兼容 | Oracle |
| `setup-oracle-env.cmd` | Windows | 配置 Instant Client 环境变量 |

`--version` 必填，未传时会交互询问。

```cmd
.\scripts\build\build-win10.cmd --version=3.3.2
.\scripts\build\build-win10.cmd --oracle --version=3.3.2
```

```bash
./scripts/build/build-centos7.sh --version=3.3.2
./scripts/build/build-centos7.sh --oracle --version=3.3.2
```

输出：`dist/gateway/`（可执行文件、`configs/`、`web/`、`scripts/db` 等）。标准构建含 MySQL / SQLite / ClickHouse（CGO）。Oracle 构建需要本机 Instant Client（Basic + SDK）和 `ORACLE_HOME`。
