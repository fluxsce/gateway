# 更新日志

本文件记录了本项目的所有重要变更。

格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，并遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范。

## [Unreleased]

## [3.2.7] - 2026-08-21

### 新增
- **企业级改密与重置**：本人改密只认 session，服务端校验 8-20 位复杂度（大小写、数字、`@$!%*?&`），失败按账号冷却。管理员重置生成一次性临时密码、踢会话，并置 `mustChangePwd=Y`，登录后只允许改密/登出。新建用户同样须首次登录改密。已有库由启动脚本对 `HUB_USER.sql` 末尾 `ALTER` 增量执行补列，不必另跑 patch。

### 修复
- **归档分批删除**：生命周期清理用 `sqlutils.BuildLimitedDeleteQuery` 按批 DELETE（默认每批 2000，MySQL `LIMIT` / Oracle `ROWNUM`），不用主键 IN 列表；单轮有批次上限，剩余下轮再删。
- **集群 ACK 与事件天数对齐**：`HUB_CLUSTER_EVENT_ACK` 不再读 yaml `ack_retention_hours`，与环境设置「集群事件」保留天数相同。

### 变更
- **用户口令改为 bcrypt 哈希**：登录、建用户、改密统一走 `pkg/security.HashPassword` / `VerifyPassword`。兼容历史明文和 `ENCY_` 密文，登录成功后静默升级为当前成本哈希。查询接口不再回传密码。默认管理员种子改为 `123456` 的 bcrypt 哈希。数据库、Redis 等需还原的密钥仍用 `ENCY_` 可逆加密。
- **管理端 HTTP 超时与 axios 对齐**：`http.Server` 的 Read/Write 超时跟随环境设置 `requestTimeoutSeconds`，保存后同步；未落库时仍回落 `web.read_timeout`。

### 新增
- **OTLP 链路追踪**：进程启动时由 `cmd/init` 调用 `pkg/tracing.Open` 创建全局 Tracer（网关数据面、管理端与其它模块共用）。配置写在 `configs/app.yaml` 的 `app.tracing`。默认关闭；开启后识别/生成 W3C `traceparent` 并在转发上游时注入，内部 `trace_id` 仍写访问日志。配置了 Collector endpoint 时异步导出 OTLP。改配置需重启进程。说明见 [链路追踪](docs/zh-CN/10-链路追踪.md)。
- **环境设置与归档（hub0009）**：系统设置可配归档天数、归档任务与 Web 超时。审计、任务、预警、集群、指标由生命周期单节点清理；网关访问日志仍按实例保留。
- **归档任务（retentionJob）**：独立分组控制统一清理 Job 的启停、分钟级间隔与每天开始时刻。关闭启用后停止清理。填写开始时间则每个自然日该时刻执行一轮；留空则按间隔反复执行（最短 1 分钟，默认 60 分钟）。
- **环境设置集群同步**：保存 hub0009 分组后发布 `ENV_SETTING/RELOAD`，其它节点从库重载该分组（含 HTTP 超时）。发布节点已写入本机缓存，轮询会排除自身。集群未就绪时跳过发布。

## [3.2.6] - 2026-08-20

### 新增
- **Linux arm64 发布包**：CI 增加 `gateway-linux-arm64-{version}.tar.gz`（`ubuntu-24.04-arm` + manylinux2014 aarch64）。标准 Docker 在对应架构 runner 上原生构建，并推 `{version}-amd64` / `{version}-arm64`，再合成多架构 `{version}` / `latest`。Oracle 包、Oracle 镜像与 Windows 仍为 amd64（Instant Client 21.18 仅 x64）。
- **审计日志查看（hub0004）**：系统设置下查询 `HUB_AUTH_AUDIT_LOG`（操作人、动作、模块、目标、结果、时间范围），只读详情。已有库执行 `patch_auth_resource_20260820_hub0004.sql`（并确认已有审计表）。
- **审计日志加固**：业务失败与无权限拒绝记 `result=N`；未声明后缀的写按钮默认入账；登录成功与踢会话写入审计；登录失败仅在账号进入冷却时记一笔（验证码/单次密码错误走 `HUB_LOGIN_LOG`，避免刷爆）。导出按当前筛选条件用数据库游标逐行写出 CSV，不限制条数。已有库另执行 `patch_auth_resource_20260820_hub0004_export.sql`。写入失败日志关键字 `AUDIT_WRITE_FAILED`。

### 修复
- **标准 Docker arm64 镜像 CGO 编译失败**：`ARG TARGETARCH=amd64` 不会被 `--platform linux/arm64` 覆盖，ARM runner 上 Go 把 `-m64` 传给本机 gcc。改为无默认值的 `TARGETARCH`，`go build` 显式传 `GOARCH`，并校验与 `GOHOSTARCH` 一致。

### 变更
- **标准 Docker 运行时升级 Alpine 3.19 → 3.23**：构建阶段钉 `golang:1.24-alpine3.23`，与仍在支持期内的 Alpine 对齐。
- **登录页首包减负**：Vite 不再把 Monaco / CodeMirror / xterm 打进登录图；boot 预加载运行时单独成块。`@/ui` 不再 re-export 代码编辑器（改从 `@/ui/code-editor` 引入）。主布局按需加载。CI 钉到 npm `niuma-ui@1.1.6`（`sideEffects` 便于摇掉未使用的编辑器）。本地开发仍用 package.json 的 link。
- **审计与鉴权分离**：权限中间件只做校验。审计由 `ApplyGlobalMiddleware` 挂载的 `AuditMiddleware` 统一落库；业务模块 `SetEvent` 计入对象，登录 `WriteDirect`，踢会话当场写。未 SetEvent 的请求不记审计。网关实例/路由/代理、安全子配置、调度任务、服务中心、命名空间、配置中心、隧道、告警配置、网关日志重发与审计导出已按此写入。

## [3.2.5] - 2026-08-20

### 新增
- **管理端按钮级鉴权**：路由组先校验登录与 MODULE；写接口在路由上声明 `RequireButton`（服务端写死按钮码，不信任客户端 `buttonCode`）。查询/详情未声明按钮时有模块即放行。租户管理员跳过校验。目录里没有该 BUTTON 时拒绝，避免拼错码被当成未配置而放行。`hubplugin` HTTP 代发映射到 `hub0023`，需 `hub0023:reset` 或 `hub0023:batchReset`。
- **前端按资源码藏按钮**：侧栏 `hasModule`、路由守卫、工具栏/搜索 `{moduleId}:{key}`、表单写权限 `hasFormWritePermission`。子弹窗沿用入口码（如 `hub0021:corsConfig`），无子码时入口码即写权限。
- **权限变更即时失效**：改角色授权或用户角色后踢掉相关会话。角色授权时勾选按钮会带上父级，勾选 MODULE/GROUP 会带上子孙。
- **写操作审计**：`HUB_AUTH_AUDIT_LOG` 记录 CREATE/UPDATE/DELETE/ROLLBACK/GRANT。用户/角色/资源/安全配置/预警日志/HTTP 代发在成功后写目标 ID 与名称；其余声明了写按钮的接口由中间件兜底（含请求里的主键）。已有库需执行对应建表脚本。
- **权限目录补齐写按钮**：`hub0006` 补 add/edit/delete；新增 `hub0003` 定时任务 MODULE 及调度器/任务写按钮；补 `hub0020/21/22:securityConfig`、`hub0082:edit`。sqlite / mysql / oracle 三份 `HUB_AUTH_RESOURCE.sql` 同步。已有库执行 `patch_auth_resource_20260820.sql`（并确认已有 `HUB_AUTH_AUDIT_LOG`）。超级管理员初始化仍全量授予目录资源。
- **登录图形验证码防刷**：验证码改为 HMAC 签名票 + 服务端 PNG，答案不出网、不写缓存。登录强制校验验证码；凭据失败按账号渐进冷却（第 5 次 30 秒，第 6 次 1 分钟，之后 2 分钟），登录页显示剩余秒数。验证码填错不计入。

### 变更
- **登录权限不再自动展开**：角色只拿到直接授予的资源码，有 MODULE 不等于有该模块全部按钮。只读角色仍只授 GROUP/MODULE 与 view/search/reset 等查询类按钮。
- **鉴权直接查库**：多实例下缓存失效不一致，权限服务暂不走进程内缓存。
- **发版流水线 niuma-ui**：CI 从前端 link 替换为 npm `niuma-ui@1.1.5`（含 1.1.4 的 Select/Table/Terminal/Dialog 与 1.1.5 的命令式弹窗销毁修复）。本地开发仍用 package.json 的 link。

## [3.2.4] - 2026-08-16

### 新增
- **路由重定向后端**：路由后端可选「重定向」，同页填写 301/302/307/308 与 Location。命中后按 RFC 9110 写出跳转并终止链路：`Location`、非 HEAD 的短 HTML 链接正文；301/308 不强制禁缓存，302/307 带 `Cache-Control: no-store`。目标支持站点内绝对路径（`/#/datahublogin`）、http(s) 绝对地址，以及 `{scheme}`/`{host}`（scheme 优先 TLS，其次单一合法的 `X-Forwarded-Proto`；host 只认请求 Host）；拒绝协议相对地址、用户信息和换行。已有库需为 `HUB_GW_ROUTE_CONFIG` 增加 `redirectStatus`、`redirectLocation`。
### 修复
- **正则路由路径不再自动补 `/`**：匹配类型为正则时，路径按输入原样保存（如 `^/datahub01webVue/(d10|d12)`），不再默认加上前导斜杠。精确匹配与前缀匹配仍自动补 `/`。

## [3.2.3] - 2026-08-16

### 新增
- **路由静态资源托管**：路由后端可选服务代理或本机目录。命中后从本机目录出文件，不再转发上游。配置挂在 `HUB_GW_STATIC_HOST_CONFIG`，`HUB_GW_ROUTE_CONFIG.backendType` 区分 `proxy`/`static`（已有库需执行对应 ALTER）。支持剥离前缀、路径重写（prefix/exact/regex）、目录索引、SPA 回退、缓存分层、预压缩、扩展名白名单、文件大小上限、自定义 404/403。查找顺序为过滤器改写路径 → 剥前缀 → 重写规则 → 文件/索引/SPA；静态映射不改请求 URI。成功命中不写访问日志、4xx 不告警，仅异常落库、5xx 告警。管理端在路由编辑选「静态资源」后维护，纯静态路由不展示转发路径项。
- **静态根目录允许名单**：根目录可写 `{v1,v2}`，剥前缀后用第一段路径匹配名单并展开；展开不得跳出固定父目录。

## [3.2.2] - 2026-08-15

### 新增
- **节点熔断接入数据面**：`enableCircuitBreaker=Y` 时按上游实例统计失败/慢调用；开闸后负载均衡跳过该节点，不按整服务直接 503。状态经 `pkg/cache` 读写，默认进程内 memory。
- **熔断阈值使用配置表**：服务定义只保留启用开关；阈值在列表右键「熔断配置」中维护，写入 `HUB_GW_CIRCUIT_BREAKER_CONFIG`。无行时运行时回退默认值。
- **熔断与失败重试分层**：失败重试只换节点；熔断按每次尝试记到对应实例，二者互不替代。
- **熔断键只按上游实例**：固定 `cb_node:{serviceId}:{nodeId}`。调用方 IP 属于限流，不再提供 service/route/ip 等统计维度。
- **全部开闸回退健康列表**：由 `nodesEligibleForSelect` 实现，不会把整服务熔死。
- **熔断使用时间滑动窗口**：按 `windowSizeSeconds` 分桶统计，过期失败不再拖住开闸判断。
- **上游 5xx 计入节点熔断**：连接失败/超时与 5xx 摘除该实例；4xx 不摘除，也不因此改重试。
- **半开探测占坑**：`Check` 占用 `halfOpenMaxRequests` 名额，避免恢复期并发打满刚复活的节点。

### 修复
- **SSE/WebSocket 正常结束被标成处理异常**：流式诊断改为写入访问日志 `noteText`，不再占用 `errorMessage`；控制台处理状态只认 `errorCode`/`errorMessage`。
- **熔断默认值不一致**：工厂回退值与配置表/表单对齐（开闸 60s、半开 3 次、慢调用 60000ms/50%）。
- **慢调用默认过短**：默认阈值从 1 秒改为 60 秒，避免把正常的长处理接口当成故障摘除。
- **告警日志当天查不到已发送记录**：查询把本地「今天」转成 UTC，东八区 16:00 之后的记录会被滤掉；已改为按本地墙钟时间查询。

## [3.2.1] - 2026-08-15

### 修复
- **Linux 发布包 glibc 过新**：流水线不再在 Ubuntu 24.04 上直接 CGO 编译（会链到 `GLIBC_2.38` 等，CentOS 7 / RHEL 7 启动即失败）；改为在 `manylinux2014`（glibc 2.17）容器内打包。动态链接器报错发生在进程进入 `main` 之前，不会写入网关 `service.log`，只会出现在 systemd journal / `/var/log/messages`。
- **Linux 打包容器 SIGPIPE 退出 141**：`set -o pipefail` 下 `ldd --version | head` 读完即关管道，改为直接打印 `ldd --version`。

### 变更
- **Linux 发布包权限与属主**：归档固定 `root:root`（`--owner=0 --group=0 --numeric-owner`），目录 `0755`、普通文件 `0644`、二进制与 `*.sh` `0755`，与 Reproducible Builds / GoReleaser 惯例一致。

## [3.2.0] - 2026-08-14

### 修复
- **Windows 发版流水线 MinGW 安装失败**：`setup-mingw@v2` 在 Chocolatey mingw 16 上删除已不存在的 `libpthread.dll.a` 导致任务退出；改为 `egor-tensin/setup-mingw@v3`。
- **标准版 Docker 镜像 apk 参数无效**：`apk add` 不支持 `--timeout` / `--tries`（那是 apt 选项），GitHub Actions 构建直接失败；已去掉。CI 构建改用官方 Alpine/Debian 源，避免默认阿里云镜像在海外 runner 上不可达。
- **Oracle Docker 构建找不到 libaio1**：`golang:1.24` 基础镜像已切到 Debian Trixie，包名变为 `libaio1t64`；构建阶段固定为 `golang:1.24-bookworm`，与运行时 `debian:bookworm-slim` 对齐。
- **阿里云 ACR 拒绝 OCI provenance**：`build-push-action` 默认附带 `application/vnd.oci.empty.v1+json` 证明，个人版 ACR 不识别；流水线关闭 provenance/SBOM 后再推送。
- **本地 Docker 推送脚本去掉明文 ACR 口令**：`scripts/docker/push.sh` 改为读取 `ALIYUN_USERNAME` / `ALIYUN_PASSWORD`（及可选 `ALIYUN_REGISTRY` / `ALIYUN_NAMESPACE`），未设置则拒绝推送。
- **前端生产构建缺少 Tailwind v4**：`niuma-ui/styles.css` 的 `@import "tailwindcss"` 在 npm 安装（无本地 link）时无法解析；宿主增加 `tailwindcss` / `@tailwindcss/vite` 并接入 Vite 插件。
- **CI 打包后控制台图标全空**：钉到 `niuma-ui@1.1.3`（含 1.1.2 的 `RsIcon` 导入修复，以及抽屉拖拽缩放与贴边定位）。
- **服务中心双向流命名空间订阅无推送**：`StreamHandler.handleSubscribeNamespace` 此前只写入连接本地订阅列表，未注册到 `ServiceSubscriber`，客户端收不到变更。现与 `SubscribeServices` 对齐，接入 `SubscribeNamespace` 并转发 `SERVER_SERVICE_CHANGE`，断连时 `UnsubscribeNamespace` 清理。

### 变更
- **Docker 镜像标签**：`{version}` / `latest` 为标准版（Alpine，MySQL/SQLite）；Oracle 版为 `{version}-oracle`。不再使用 `{version}-standard`。流水线与 `scripts/docker/build.sh`、`push.sh` 一致。
- **发版流水线推送阿里云 ACR**：与 GHCR 同一构建打相同标签；凭据使用 Secrets `ALIYUN_USERNAME` / `ALIYUN_PASSWORD`，未配置则只推 GHCR。

### 新增
- **GitHub Actions 多平台发布包**：tag `v*` 或手动触发时并行打 Linux/Windows 二进制（MySQL/SQLite 与 Oracle 两套）及 standard/oracle Docker 镜像；打包逻辑在 `.github/` 内维护，推 GHCR 并挂到 GitHub Release。
- **服务中心 SQLite E2E 测试环境**：`internal/servicecenter/testutil` 用临时 SQLite 建表/种子并启动 gRPC 实例；`internal/servicecenter/e2e` 覆盖注册发现、命名空间推送、配置 Watch、注销；`cmd/servicecenter-testd` 供 Java SDK 拉起真实服务端。
- **服务中心 API 缺口 E2E**：`api_coverage_test.go` 覆盖 RegisterNode、UnregisterService（整服务/带 nodeId）、SubscribeServices、ListConfigs/GetConfigHistory/RollbackConfig/DeleteConfig；`stream_control_test.go` 覆盖 Stream `CLIENT_SUBSCRIBE_NAMESPACE` 与 `SERVER_ERROR`；`ConnectionManager.Close` 单元测试覆盖 `SERVER_CLOSE` 广播。
- **服务中心认证落地与 E2E**：Bearer 占位改为真实校验——不透明 API Token（`HUB_SERVICE_AUTH_TOKEN`）与 HS256 JWT（密钥/issuer 来自实例 `extProperty`）；`testutil` 支持 `EnableAuth` 种子用户/令牌；`auth_scenario_test.go` 覆盖无凭证/错凭证拒绝与 Basic、API Token、JWT 成功路径及 Stream 认证；`servicecenter-testd` 支持 `SC_E2E_ENABLE_AUTH`。
- **服务中心 TLS/mTLS E2E**：`testutil` 自动生成自签 CA/服务端/客户端证书（含 127.0.0.1 SAN）；`EnableTLS`/`EnableMTLS` 启动选项；修复 mTLS 下 `ClientCAs` 从 `certChainContent`（CA PEM 或路径）加载；`tls_scenario_test.go` 覆盖明文拒绝、TLS 成功、mTLS 无客户端证书拒绝与成功；`servicecenter-testd` 支持 `SC_E2E_ENABLE_TLS` / `SC_E2E_ENABLE_MTLS`。

### 优化
- **表单日期绑定 RFC3339**：`RsDatePicker` 支持 `valueFormat=iso`（展示仍为墙钟）；数据表单 date/datetime 默认 `iso`，提交仍兜底转 Go `time.Time`。

## [3.1.9] - 2026-07-15

### 新增
- **网关日志参数/报文体关键字查询**：日志查询支持按请求参数、报文体关键字子串检索（关系库 / ClickHouse / Mongo）。
- **网关代际热重载**：新连接立即使用新配置，旧连接按优雅关闭超时排空，避免重载中断在途请求。
- **网关容量保护**：接入最大 TCP 连接数和最大在途 HTTP 请求数限制；请求过载时返回包含 `trace_id` 的统一 503 响应。
- **统一流式代理**：SSE 与两类 WebSocket 入口统一使用请求取消、滚动读写超时和代际排空生命周期；WebSocket 会话使用串行写泵和幂等资源回收。

### 优化
- **请求耗时与配置界面**：统一请求起止时间和异常访问日志，补全容量及优雅排空配置；统一管理前端浮层层级，避免对话框内提示被遮挡。
- **流式配置语义**：补齐路由超时、重试、WebSocket 和路径重写映射；明确 SSE 单服务、`timeout=0` 及长连接占用 `MaxWorkers` 的配置约束。
- **SSE/WebSocket 访问与后端日志**：按日志配置采样请求/响应报文前缀；写入转发字节数、断开/关闭原因；WebSocket 补充后端握手追踪。
- **网关对外错误文案英文**：对访问端返回的 `error` 统一为英文短句；`AddError`/运维日志仍可使用中文便于排障。
- **实例状态说明与时间**：列表展示 `reserved1` 最近异常说明；启动/停止同步刷新 `editTime`，重载成功/失败也会更新实例表时间戳与状态说明。

### 修复
- **路由关联服务重新选择未回显勾选**：已选服务点击「重新选择」时，选择弹窗内不再丢失原勾选状态。
- **网关重载稳定性**：修复连接分发、实例池锁顺序和旧代际资源释放问题，并增加热重载压力及竞态测试。
- **HTTP 访问下日志重发弹窗报错**：`crypto.randomUUID` 在非安全上下文不可用，改为兼容回退，避免 Gateway 日志重发页初始化失败。

## [3.1.8] - 2026-06-25

### 新增
- **Bearer Token 认证**：网关新增 `bearer-token` 策略，校验 `Authorization: Bearer <token>` 与配置值是否一致；认证配置页支持选择「Bearer Token认证」并填写 `token`。

### 优化
- **API Key 认证配置精简**：对齐 `APIKeyConfig` 定义，仅保留 `param_name`、`in`、`key` 三个字段；移除 `APIKeyItem`、多 Key 列表与前缀匹配；前端改为参数名 + 单个密钥值配置。
- **认证配置页**：按认证类型维护字段白名单，切换类型时清理无关参数；JWT 按算法显示 `secret`/`publicKey`；API Key / Bearer Token / Basic 独立配置区与提交校验。

### 修复
- **JWT 认证未真正验签**：接入 `golang-jwt` 实现签名、过期与可选 issuer 校验（HS/RS 系列算法）。
- **Basic 认证未生效**：修复 `BasicAuthFromConfig` 误返回 `BaseAuthenticator` 导致认证被绕过的问题；凭据改为恒定时间比较。
- **OAuth2 认证误放行**：修复工厂返回基础认证器导致静默通过；远端 Token 内省暂未实现时明确拒绝请求。
- **API Key 空密钥绕过**：未配置有效 Key 时不再放行；仅填参数名无法通过认证。

## [3.1.7] - 2026-06-18

### 新增
- **构建脚本支持直接打包**：构建后生成发布压缩包（Linux `tar.gz` / Windows `zip`），版本号改为 `--version` 指定。

### 优化
- **Oracle 客户端库不随包分发**：因 OTN 许可限制，打包不再包含 Oracle 客户端运行库，需在目标机器自行安装。

### 修复
- **服务节点地址编辑丢参**：编辑/保存时保留用户填写的完整 URL（路径与查询参数）。
- **网关转发参数覆盖**：转发时节点地址中配置的参数覆盖前台携带的同名参数。
- **服务定义改为静态配置不生效**：修复由"服务发现"改为"静态配置"（serviceType=0 零值被跳过）无法保存的问题。
- **服务发现改为静态配置残留元数据**：由"服务发现"改为"静态配置"时一并清空 `serviceMetadata`/`discoveryType`/`discoveryConfig`（字符串零值会被 skipZero 跳过），避免网关仍按服务发现路由。

## [3.1.6] - 2026-05-14

### 修复（Web 控制台）

- **主布局页签菜单切换后之前的状态丢失修复**：

## [3.1.5] - 2026-04-23

### 优化
- **弱网首屏**：静态 Loading + `boot.ts` 延后加载 `main.ts`，减少空白等待。

## [3.1.4] - 2026-04-21

### 新增
- **帮助手册使用说明**：基于 VitePress 的模块文档（与菜单分组对齐）；生产打包写入 `web/frontend/dist/docs/` 并由网关静态挂载至与控制台同源的 `/docs` 路径；头部帮助入口侧栏预览与「新窗口打开」优化。

### 修复
- **网关日志详情查看**：修复详情查看失败问题。

## [3.1.3] - 2026-04-11

### 新增
- **网关日志请求重发**：在网关日志查询中支持选中条目打开重发弹窗

### 修复
- **全屏缩放**：

## [3.1.2] - 2026-04-02

### 新增
- **服务发现路径负载均衡**：从服务中心（INTERNAL）拉取多实例时，与静态节点一致，按服务定义中的负载均衡策略（轮询、随机、IP 哈希、最少连接、加权轮询、一致性哈希等）选择目标实例；不再固定使用首个健康实例。

### 修复
- **网关日志「仅错误」查询**：修正仅查询错误日志时的条件与前后端参数语义（如 `errorOnly` 与网关状态码筛选），避免误查或漏查。
- **指标采集 server_info 生命周期**：停止指标采集时按当前 `metricServerId` 物理删除本机服务器信息行，避免容器等环境反复重启导致表中记录无限增长

## [3.1.1] - 2026-03-24

### 修复
- **Oracle 部分模块 SQL 语法兼容问题**：修复多处使用 MySQL 专有 `LIMIT` 语法的查询，在 Oracle 环境下报 `SQL command not properly ended` 的错误

## [3.1.0] - 2026-03-21

### 新增
- **主布局多页签**：已移除全局面包屑状态与路由守卫中的面包屑写入。
- **网关实例导出/导入（Excel）**：按 `gatewayInstanceId` 导出多 Sheet 配置并支持导入 Upsert。

### 优化
- **容器内应用目录**：由 `/opt/gateway` 改为 `/home/gateway`（与运行用户主目录一致，减轻部分环境挂载 `/opt` 的权限问题）；Oracle 版仍使用 `/opt/oracle`；`docker-compose` / K8s 示例与文档已同步，升级需将卷挂载目标改为新路径。
- **路由多服务名称显示**：选择服务时将服务名称映射（`serviceNameMap`）存入 `routeMetadata`，列表展示时优先显示服务名称而非 ID，解决多服务场景下仅显示 ID 的问题。

### 修复
- 新增用户时密码框误隐藏：改为按弹窗模式（新增/编辑）控制密码框显示，避免先填用户ID导致密码框消失。
- Docker 国内镜像与构建失败：Oracle 版默认阿里云 Debian 源，标准版默认国内 Alpine 源；build.sh 支持 -m/DEBIAN_MIRROR；修复镜像源替换 RUN 退出码导致构建失败。

## [3.0.9] - 2026-03-09

### 新增
- 🗄️ **Oracle 数据库支持完善**
  - Linux 安装脚本（`install-service-linux.sh`）增强 Oracle 环境变量检查
  - 安装时自动加载 `/etc/profile`、`~/.bash_profile`、`~/.profile` 以继承 `ORACLE_HOME`、`LD_LIBRARY_PATH`
  - 支持 `ORACLE_HOME` 下库文件在根目录或 `lib` 子目录两种解压结构
  - 当环境变量存在但 glob 检测失败时，直接使用 `ORACLE_HOME` 作为回退

### 修复
- 🐛 **修复 Oracle 表名超出 30 字符限制导致 ORA-00972 错误**
  - `HUB_STATEMENT_EXECUTION_HISTORY` 改为 `HUB_STMT_EXECUTION_HIST`（Oracle）
  - `HUB_SCRIPT_EXECUTION_HISTORY` 改为 `HUB_SCRIPT_EXEC_HIST`（Oracle）
  - MySQL、SQLite、ClickHouse 保持原表名不变

## [3.0.8] - 2026-02-28

### 新增
- 🚀 **网关日志自动清理功能**
  - 新增 `LogCleaner` 组件，支持自动清理过期的访问日志和后端追踪日志

### 修复
- 🐛 **[严重] 修复日志批量写入失败导致的死循环问题**
  - 修复 `Flush` 方法在批量写入失败后不清空缓冲区，导致每次定时刷新都重复尝试写入相同的失败数据
- 🐛 **修复日志写入时 HTTP 方法和路径过长导致数据库插入失败问题**

## [3.0.7] - 2026-02-19

### 新增
- 🚀 **服务注册发现集群支持（基于 Redis）**
  - 新增 `RedisServiceCache` 实现，支持服务注册中心集群部署

- 🔔 **服务注册中心告警支持**
  - 前端新增告警配置选项卡，支持 9 种告警场景配置
  - 后端集成启动/停止、节点注册/注销/驱逐、缓存同步、配置变更、连接断开等告警
  - 新增 `ConfigProvider` 接口统一配置访问

- 🔒 **Redis 配置支持加密密码**
  - Redis 配置的密码字段支持使用 `password_plugin` 加密

- 🔒 **数据库配置支持加密密码**
  - 数据库连接配置的密码字段支持使用 `password_plugin` 加密

### 修复
- 🐛 **[严重] 修复 Redis 缓存潜在的 Panic 和并发安全问题**
  - 修复 `getUniversalClient()` 返回 nil 导致所有 Redis 操作 panic 的问题
  - 修复 `Close()` 后继续使用导致 panic 的问题
  - 完整实现 TLS CA 证书加载功能
  - 将 `redis.go`（1338行）拆分为 8 个模块化文件，提高可维护性

- 🐛 **修复 Docker 容器中文乱码和日志识别问题**
  - 在 Alpine 镜像（Dockerfile）中添加 `musl-locales` 和 `musl-locales-lang` 包支持
  - 在 Debian 镜像（Dockerfile.oracle）中添加 `locales` 包并生成 `zh_CN.UTF-8` locale

- 🐛 **修复集群服务重启后回溯处理历史事件问题**
  - 修复集群服务重启时 `lastEventTime` 初始化为1小时前导致的问题

## [3.0.6] - 2026-01-30

### 新增
- 🚀 **服务注册发现功能增强**
  - 基于 gRPC 双向流的高性能服务注册与发现
  - 支持服务节点自动注册、注销与心跳检测
  - 支持服务订阅与实时变更推送（SubscribeServices、SubscribeNamespace）
  - 服务健康检查与自动故障剔除
  - 支持临时实例与永久实例
  - 支持服务元数据与标签管理
  - 多命名空间与分组隔离

- 🚀 **配置中心管理功能**
  - 支持配置的创建、修改、删除与查询
  - 支持配置变更实时推送
  - 支持配置版本管理与历史记录
  - 支持配置导入导出

## [3.0.5] - 2026-01-22

### 新增
- 🚀 **网关预警功能**
  - 新增网关日志告警功能，支持基于日志写入路径的自动告警
  - 支持状态码告警：可配置多个HTTP状态码（默认502），当请求返回指定状态码时自动触发告警
  - 支持超时告警：可配置超时阈值（默认120秒），当请求总耗时超过阈值时自动触发告警

- 🚀 **密码加密工具**
  - 新增独立的密码加密命令行工具（`password_plugin`），支持交互式和命令行参数两种模式
  - 支持 AES-256-CBC 加密算法，系统自动识别并处理已加密的密码字符串
  - 支持使用默认密钥（从配置文件 `app.encryption_key` 或环境变量 `GATEWAY_APP_ENCRYPTION_KEY` 读取）

## [3.0.4] - 2026-01-15

### 新增
- 🚀 **集群部署节点自动通知功能**
  - 新增集群事件发布/订阅机制，支持多节点自动同步
  - 网关实例启动、停止、重载配置时自动通知所有集群节点

- 🚀 **后端代理重试机制**
  - 新增 HTTP 代理请求重试功能，避免集群环境后端某个节点突然挂掉发现不及时
  - 支持配置重试次数（`RetryCount`）和重试超时时间（`RetryTimeout`），默认重试次数为 0（不重试）
  - 每次重试自动重新选择节点，避免一直重试同一台异常节点，提高请求成功率
  - 重试场景下累加所有重试的耗时，准确记录后端服务总耗时
  - 每次后端调用（包括重试）都会记录后端追踪日志，日志中包含重试次数信息
  - 支持单服务和多服务转发场景的重试机制
  - 使用原子操作保证多服务并发转发场景下的线程安全

### 变更
- ⚡ **集群功能数据库要求**
  - 集群功能依赖数据库作为事件存储和同步介质
  - **要求使用物理数据库**（MySQL、Oracle、PostgreSQL、SQL Server 等）

## [3.0.3] - 2026-01-12

### 变更
- 🔧 **配置工具包迁移**
  - 将 `cmd/common/utils/config.go` 迁移至 `pkg/config/path.go`
  - 统一配置路径管理，包括 `GetConfigDir()`、`GetConfigPath()`、`ResolvePath()` 等函数
  - 新增 `GetDuration()` 函数，支持时间间隔配置读取
  - 更新所有相关模块的导入路径

- 🔧 **构建脚本优化**
  - 修改 `build-win10.cmd` 启用 CGO 以支持 SQLite 数据库
  - 添加 GCC 环境检查，构建前提示安装 MinGW/TDM-GCC

- ⚡ **网关日志 IP 查询支持模糊匹配**
  - 关系数据库、MongoDB、ClickHouse 三种存储的日志查询均支持客户端 IP 模糊查询
  - 支持输入部分 IP 进行模糊搜索，如输入 `192.168` 可匹配 `192.168.1.1`、`192.168.0.100` 等

- 🔧 **网关日志和监控查询时间范围必填**
  - 网关日志查询和监控查询的时间范围字段标记为必填，显示红色星号提示

### 修复
- 🐛 **修复前端组件 HTML id 包含特殊字符问题**
  - 修复 Vue 组件中 HTML id 属性包含冒号（:）等特殊字符导致的问题

- 🐛 **修复日志记录客户端 IP 带端口号问题**

- 🐛 **修复右键菜单二级菜单被遮挡问题**

## [3.0.2] - 2026-01-07

### 新增
- 🚀 **路由包问题断言功能**
  - 新增请求体内容断言器（BodyContentAsserter），支持对 HTTP 请求体进行断言
  - 支持多种比较操作符：等于、不等于、包含、不包含、以...开头、以...结尾、正则匹配、存在、不存在
  - 支持大小写敏感/不敏感的比较选项
  - 优先使用上下文中已缓存的请求体，避免重复读取，提升性能
  - 正确处理请求体资源管理，确保无资源泄露

- 🚀 **Web 应用 HTTPS 支持**
  - 新增 Web 应用的 HTTPS 协议支持

### 修复
- 🐛 **修复断言比较操作符逻辑错误**
  - 修复 `Contains` 和 `NotContains` 操作符的参数顺序错误
  - 修复前：检查期望值是否包含实际值（逻辑错误）
  - 修复后：检查实际值是否包含期望值（逻辑正确）
  - 确保所有断言器（Header、QueryParam、Cookie、BodyContent 等）的比较逻辑一致

- 🐛 **修复多服务转发响应体乱码问题**
  - 修复多服务转发场景下，其中一个服务的响应体出现乱码的问题
  - 正确处理 Go `http.Client` 自动解压后的响应体，移除 `Content-Encoding` 头避免客户端二次解压
  - 确保日志记录中的响应体内容正确显示

- 🐛 **优化日志配置获取性能**
  - 在请求上下文中直接存储日志配置，避免重复获取
  - 根据日志配置（`RecordRequestBody`、`RecordResponseBody`）控制是否缓存和记录请求体/响应体
  - 仅在需要记录时才缓存请求体和响应体，减少内存占用
  - 使用 `MaxBodySizeBytes` 限制记录的大小，避免记录过大的内容

- 🐛 **修复网关功能权限初始化缺失问题**
  - 完善 hub0020、hub0021、hub0022 等模块的权限资源初始化脚本
  - 补充全局配置、路由过滤器、节点管理等子功能的增删改查重置子权限
  - 确保前端模块 `moduleId` 与数据库中权限资源一一对应，避免权限校验异常

- 🐛 **修复日志写入到操作系统日志问题**
  - 修复通过 systemd 启动服务时，应用日志和 GIN 访问日志被重定向到 `/var/log/messages` 的问题
  - 将应用日志默认输出从 `stdout` 改为文件 `./logs/gateway.log`，支持日志轮转
  - 将 GIN 框架访问日志输出到文件 `./logs/web.log`，避免输出到 stdout
  - 确保所有日志独立存储，不再混入系统日志，便于日志管理和排查

- 🐛 **修复分辨率导致页面布局显示问题**
  - 修复不同分辨率下页面布局显示异常的问题
  - 优化响应式布局，确保在不同屏幕分辨率下页面元素正确显示
  - 修复高分辨率或低分辨率显示器上的布局错乱问题

- 🐛 **修复 1024 以下端口绑定权限问题**
  - 修复 Linux systemd 服务无法绑定特权端口（如 443）的问题
  - 在 systemd 服务文件中添加 `AmbientCapabilities=CAP_NET_BIND_SERVICE` 和 `CapabilityBoundingSet=CAP_NET_BIND_SERVICE`
  - 在 Docker Compose 配置中添加 `cap_add: [NET_BIND_SERVICE]` 支持
  - 在 Kubernetes Deployment 配置中添加 `NET_BIND_SERVICE` capability
  - 改进网关启动代码的端口绑定错误检测，当绑定特权端口失败时提供明确的错误提示和解决方案
  - 支持非 root 用户绑定特权端口，提升安全性

## [3.0.1] - 2026-01-05

### 新增
- 🚀 **前端样式统一**
  - 统一前端 UI 组件样式规范
  - 优化用户体验一致性
  - 路由名称和网关日志中的路由名称使用主题主色突出显示

- 🚀 **网关后端服务群发功能**
  - 支持向多个网关后端服务批量发送消息

- 🚀 **权限资源管理**
  - 完善权限资源管理功能
  - 优化权限资源结构

### 变更
- ⚡ **网络隧道 SSE 传输优化**
  - 使用 `io.CopyBuffer` 替代 `io.Copy`，避免 Linux `splice` 系统调用导致的连接超时问题

- 🔧 **网关安全设置支持列表配置**
  - 网关安全设置部分支持列表配置方式
  - 提升配置灵活性和可维护性

- 🔧 **构建脚本优化**
  - 统一 Windows 和 Linux 构建脚本的打包路径和目录结构
  - 修复构建标签使用问题（Oracle 支持使用 `!no_oracle` 标签，而非 `ora` 标签）
  - 增强 Oracle 环境检查，环境不完整时停止构建并提示
  - 优化可执行文件与目录冲突处理逻辑
  - 修复 Windows 构建脚本中的链接器参数问题

### 修复
- 🐛 **修复 WebSocket 连接 hijack 后的响应处理问题**
  - 修复 WebSocket 升级成功后连接被 hijack 时，后续错误处理尝试写入 HTTP 响应导致的错误

- 🐛 **修复高并发场景下数据连接建立超时问题**
  - 修复客户端频繁出现 `dial tcp ... i/o timeout` 错误的问题
  - 添加重试机制，提高临时网络问题下的连接成功率

- 🐛 **修复隧道客户端服务注册后代理未启动问题**
  - 修复服务注册成功后未调用 `StartProxy` 导致 "proxy for service not found" 错误的问题
  - 确保服务注册成功后自动启动代理，服务注销时自动停止代理

## [2.0.4] - 2025-11-28

### 修复
- 🐛 **修复指标采集服务器信息管理器初始化问题**
  - 修复服务器信息插入时的唯一约束冲突问题
  - 优化查询逻辑，优先使用唯一约束字段（tenantId, metricServerId）查询
  - 确保服务器信息存在时执行更新而非插入操作

- 🐛 **修复限流器IP策略的IP提取问题**
  - 修复 ExtractIPKey 函数从 RemoteAddr 提取IP时未去除端口号的问题
  - 正确处理 X-Forwarded-For 头部包含多个IP的情况（取第一个）
  - 使用 net.SplitHostPort 正确分离IP和端口号

- 🐛 **修复语句执行历史记录ID生成可能重复的问题**
  - 修复 generateStatementId 和 generateExecutionId 函数使用时间戳可能生成重复ID的问题
  - 改用 random.Generate32BitRandomString() 生成唯一ID，确保分布式环境下也不重复
  - 使用时间戳+进程标识+原子计数器+强随机数组合，避免唯一约束冲突

## [2.0.3] - 2025-11-20

### 修复
- 🐛 **[关键] 修复异步日志记录的并发安全问题**
  - 修复日志记录时的竞态条件
  - 增强并发安全性保护

- 🐛 **[关键] 修复隧道客户端自动重连机制**

- 🐛 **内网穿透隧道功能超时问题修复**
  - 修复网络延迟较高情况下客户端数据连接建立超时问题
  - 修复带宽较低情况下连接超时问题

- 🐛 **修复 TLS 端口绑定异常问题**

### 变更
- ⚡ **隧道重连机制性能优化**
## [2.0.2] - 2025-11-11

### 新增
- 🚀 **内网穿透（隧道）功能**
  - 基于 FRP 架构的企业级内网穿透解决方案
  - 支持 TCP、UDP、HTTP、HTTPS 等多种协议代理
  - 反向代理：将内网服务暴露到公网
  - 客户端-服务器架构，支持多客户端连接
  - 动态服务注册与端口分配
  - 静态端口映射与负载均衡
  - 连接池技术，显著提升性能
  - 双向数据连接池（客户端到本地服务、客户端到服务器）
  - 池化连接预建立机制，降低连接延迟
  - 连接有效性检查，自动清理失效连接
  - 支持长连接（SSH、数据库等）与短连接（HTTP）
  - TCP 半关闭机制，优雅处理连接终止
  - 服务注册持久化，应用重启自动恢复

- 📊 **隧道管理界面**
  - 隧道服务器管理（hub0060）
  - 服务器节点管理（hub0061）
  - 隧道客户端管理（hub0062）
  - 隧道服务管理（hub0063）

### 变更
- ⚡ **性能提升**
  - 双重连接池机制，大幅降低连接建立开销
  - 连接复用率提升 80%+
  - 高并发场景下性能提升 3-5 倍
  - 减少 TCP TIME_WAIT 状态堆积
  - 优化内存使用，避免连接泄漏

### 修复

## [2.0.1] - 2025-10-24

### 新增
- 🚀 **服务注册中心功能**
  - 支持内部服务注册与服务发现机制
  - 支持 Nacos、Consul、Eureka、ETCD、ZooKeeper 等外部注册中心代理模式
  - 服务分组与授权管理
  - 服务实例自动注册、注销与心跳检测
  - 服务事件日志记录与追踪
  - 完整的服务健康检查机制
  - 支持临时实例与永久实例

- 🚀 **JVM 监控系统**
  - JVM 资源全方位监控（内存、线程、GC、类加载）
  - 堆内存与非堆内存详细监控
  - 内存池监控（Eden、Survivor、Old Gen、Metaspace 等）
  - GC 快照采集（jstat -gc 风格）
  - 线程监控与线程状态分布统计
  - 死锁检测与告警
  - 类加载器统计与性能分析
  - 应用级监控数据采集（线程池、连接池、自定义指标等）
  - 健康状态评级与智能告警

- 📊 **数据库支持增强**
  - 新增服务注册中心数据库表（4 张表）
  - 新增 JVM 监控系统数据库表（9 张表）
  - 完善 MySQL、Oracle、SQLite 初始化脚本
  - 优化索引策略

### 修复
- 🐛 修复服务以 service 方式启动时日志文件不滚动的问题

## [2.0.0] - 2025-08-05

### 新增
- **核心网关功能**
  - 支持基于路径、方法、Header 的智能路由
  - 轮询、加权轮询、最少连接等多种负载均衡算法
  - 服务发现与健康检查
  - 高性能 HTTP/HTTPS 代理转发

- **安全特性**
  - JWT、OAuth2、API Key 认证
  - IP 白名单/黑名单、User-Agent 过滤
  - 完善的 CORS 支持
  - 自动注入安全响应头

- **性能优化**
  - 令牌桶、滑动窗口限流
  - 熔断器，自动故障检测与降级
  - 支持 Redis 与内存缓存
  - 连接池高效复用

- **监控与运维**
  - 实时系统指标与 API 调用统计
  - 多格式结构化日志
  - 服务健康监控
  - 集成 pprof 性能分析

- **管理界面**
  - 直观的 Web 管理控制台
  - 可视化配置管理与实时更新
  - 实时性能与流量监控面板
  - 在线日志查看与检索

- **数据库支持**
  - MySQL 8.0+ 支持
  - MongoDB 4.4+ 支持
  - ClickHouse 分析型数据库支持
  - Oracle 数据库支持
  - SQLite 轻量级部署支持
  - Redis 缓存集成

- **其他特性**
  - 多租户架构支持
  - 配置热加载，无需重启服务
  - 优雅的关闭与重启
  - SFTP、SSH 工具集成
  - 定时任务与 Cron 管理
  - 指标采集与系统监控
  - 插件架构，便于扩展

### 技术细节
- 基于 Go 1.24+ 开发
- 使用 Gin 框架处理 HTTP
- 实现现代微服务架构模式
- 云原生设计，支持 Kubernetes
- 完善的测试覆盖
- 支持 Docker 容器化部署

### 配置
- 基于 YAML 的配置系统
- 支持环境变量
- 配置校验与热加载
- 提供丰富的示例与模板

### 性能
- 支持 10,000+ 并发连接
- 支持 50,000+ QPS
- P99 延迟低于 1ms
- 内存占用低于 100MB
- 空闲时 CPU 占用低于 5%

## [1.x.x] - 历史版本

### 说明
1.x.x 版本为内部发布版本，未在本日志中详细记录。
2.0.0 为首个公开发布版本，包含完整文档与 Apache 2.0 协议。

## 版本对比

| 版本    | 发布时间    | 主要特性                                       |
|---------|-------------|------------------------------------------------|
| 3.0.6   | 2026-01-28  | 服务注册发现增强、配置中心管理、节点重连复用修复 |
| 3.0.5   | 2026-01-22  | 网关预警功能、告警管理模块、密码加密工具、数据库初始化脚本完善 |
| 3.0.4   | 2026-01-15  | 集群部署节点自动通知、后端代理重试机制、事件发布/订阅机制        |
| 3.0.3   | 2026-01-12  | 配置工具包迁移、构建脚本优化、日志IP模糊查询、前端修复 |
| 3.0.2   | 2026-01-07  | 路由断言功能、断言逻辑修复、多服务转发乱码修复、日志优化、权限初始化修复 |
| 3.0.1   | 2026-01-05  | 前端UI统一、权限资源结构优化、隧道客户端修复、构建脚本优化 |
| 2.0.4   | 2025-11-28  | 指标采集初始化修复、限流器IP提取修复           |
| 2.0.3   | 2025-11-20  | 自动重连修复、心跳机制增强、服务注册重试       |
| 2.0.2   | 2025-11-11  | 内网穿透、双向连接池、死锁修复、性能优化       |
| 2.0.1   | 2025-10-24  | 服务注册中心、JVM 监控、日志滚动修复           |
| 2.0.0   | 2025-08-05  | 全面重构，Apache 2.0，完整文档                 |
| 1.x.x   | 内部版本    | 旧版内部版本                                   |

## 升级指南

### 从 2.0.x 升级到 3.0.1

本次升级包含网关安全管理配置的不兼容变更，需要重新配置安全设置：

1. **网关安全设置配置变更**
   - 安全设置部分支持列表配置方式
   - 原有安全设置配置格式已变更
   - **重要**：需要重新配置网关的安全设置（IP 白名单/黑名单、User-Agent 过滤、CORS 等）

**迁移步骤：**
1. 备份现有配置与数据
2. 备份当前网关安全设置配置
3. 使用安装脚本重新安装服务
   - **Linux**: 执行 `scripts/deploy/install-service-linux.sh` 重新安装服务
   - **Windows**: 执行 `scripts/deploy/install-service.cmd` 重新安装服务
4. 重新配置网关安全设置（IP 白名单/黑名单、认证策略等）
5. 在开发环境测试新配置
6. 验证安全策略是否正常工作
7. 采用滚动升级策略部署到生产环境

**注意事项：**
- 升级前请务必备份安全设置配置
- 升级后需要手动重新配置所有安全设置
- 建议在升级前记录当前的安全策略，以便迁移时参考

### 从 1.x.x 升级到 2.0.0
本次为重大版本升级，存在不兼容变更：

1. **配置格式**：YAML 配置全新设计
2. **API 变更**：采用全新 RESTful API 结构
3. **数据库结构**：表结构与命名规范更新
4. **依赖升级**：升级至 Go 1.24+ 及现代依赖

**迁移步骤：**
1. 备份现有配置与数据
2. 参考 `/configs/gateway.yaml` 新配置格式
3. 使用迁移脚本更新数据库结构
4. 在开发环境测试新配置
5. 采用滚动升级策略部署到生产环境

### 版本兼容性承诺
自 2.0.0 起：
- **遵循语义化版本规范**
  - 主版本号 (X.0.0)：不兼容的 API 变更
  - 次版本号 (X.Y.0)：向下兼容的功能新增
  - 修订号 (X.Y.Z)：向下兼容的问题修复
- 次版本和修订版完全向后兼容，无需升级指南
- 重大变更将在主版本中进行，并提前公告

**注意：** 3.0.1 版本虽然为修订版本，但包含网关安全管理配置的不兼容变更，需要重新设置。这是由于权限系统优化的必要性，我们将在后续版本中保持更好的向后兼容性。

## 贡献日志说明

贡献代码时请：

1. 将变更添加到 `[Unreleased]` 部分
2. 使用如下分类：
   - `新增`：新特性
   - `变更`：已有功能变更
   - `弃用`：即将移除的功能
   - `移除`：已移除的功能
   - `修复`：Bug 修复
   - `安全`：安全漏洞修复

3. 格式示例：
   ```markdown
   - **类别**：简要描述 [#PR编号](PR链接)
   ```

4. 链接相关的 Pull Request 和 Issue

## 支持与反馈

如需了解特定版本详情或升级支持：
- 查阅[项目文档](docs/)
- 提交 [Issue](https://github.com/fluxsce/gateway/issues)
- 参与[社区讨论](https://github.com/fluxsce/gateway/discussions)

---

**图例：**
- 🚀 新特性
- 🔧 改进
- 🐛 修复

- ⚡ 性能
- 🛡️ 安全 