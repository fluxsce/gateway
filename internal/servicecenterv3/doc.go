// Package servicecenterv3 is the Gateway-embedded Service Center 3.0
// (registry + naming + config). It is not a standalone cluster product
// (no Distro/Raft) and is not a drop-in replacement for Nacos or Apollo.
//
// # Open-source and review boundary
//
// This tree lives under gateway/internal, so Go modules outside this
// repository cannot import it. That is intentional: the first-class
// public surface for SDKs and other languages is the bidirectional
// gRPC stream (ServiceCenterStream.Connect) and its message types.
// The in-process Go surface is contract (Admin / Naming / Config) plus
// model. A future extract to a standalone module should keep those two
// layers and the stream protocol unchanged.
//
// License: this package inherits the Gateway repository Apache License
// 2.0 (see LICENSE at the module root). Do not add a second license
// file in this tree. Third-party dependencies (gRPC) keep their
// own notices in the module vendor/go.mod graph.
//
// # Stability
//
// APILevel is experimental. This version boots servicecenterv3 only and
// no longer starts internal/servicecenter. Compatibility promises
// for this version:
//
//   - Sentinel errors in contract are stable for errors.Is.
//   - CallContext fields and the three L3 interfaces are additive-only.
//   - Stream opcode names in proto stay compatible with existing Java
//     clients; new opcodes must be optional. The IDL lives in this
//     tree (proto/), not in internal/servicecenter.
//   - Persistence still uses HUB_SERVICE* tables; column additions are
//     additive. Do not rename or drop columns in this series.
//
// Breaking changes belong in a new major (servicecenterv4 or a
// standalone module), not silent edits to these contracts.
//
// # Isolation and tenancy
//
// A CenterInstance is one listen address plus one isolated runtime
// (cache, evictor, session hub, stream server). Instances do not share
// a global cache. The runtime key is InstanceName + Environment so
// DEV/STAGING/PRODUCTION can coexist. TenantID is written only by the
// auth interceptor (or the instance's own TenantID when auth is off)
// and must match the instance tenant. Opaque tokens may bind
// instanceName / environment / namespaceIds in extProperty. Clients
// must not supply tenant identity. Namespace is the post-auth workspace,
// not the environment boundary.
//
// Gateway production is a process cluster. CenterInstance Start / Stop /
// Reload / Unload is published as SERVICE_CENTER_INSTANCE on
// HUB_CLUSTER_EVENT so every node applies the same lifecycle.
// Naming and config mutations fan out over HUB_CLUSTER_EVENT
// (SERVICE_CENTER_INSTANCE NAMING/CONFIG). Heartbeats do not publish events. Ephemeral and
// persistent nodes both live in HUB_SERVICE_NODE plus the local Cache.
// Published config is database-only: Get/Watch read HUB_SERVICE_CONFIG_DATA.
// Subscribe naming pushes the Cache snapshot on the connected pod, then
// live-view reconcile (when Redis is on) backfills peers without waiting
// for HUB_CLUSTER_EVENT. The node that accepted the SDK stream is Owner
// for eviction. This is not Distro/Raft: every gateway is a primary for
// register/discover. Admin HTTP is hub0040 (center lifecycle), hub0041
// (namespace / sessions), hub0042 (service/node inspect and UpdateNode),
// hub0043 (config draft/publish). Data-plane unregister is per connection;
// admin CallContext has no ConnectionID and must not be used as “one SDK left”.
//
// # Data plane vs management plane
//
// Data plane: gRPC stream only, on each CenterInstance listen port.
// Management plane: Gateway HTTP on the web port, calling contract.Admin
// in-process. There is no public Unary gRPC ServiceRegistry/ConfigCenter
// and no OpenAPI data-plane in this version.
//
// # Concurrency
//
// Init / GetPool / InProcess / HasRunning are safe for concurrent use.
// Each CenterInstance owns its own mutexes; do not share a Cache across
// instances.
//
// # 中文：架构与实现原理
//
// 分层图、集群部署、Redis 目录与读写路径见同目录 ARCHITECTURE.md，给使用方和文档直接看。
//
// 本包是网关内嵌的注册 / 发现 / 配置中心。客户端不直连数据库，只连中心实例
// 的 gRPC 双向流 ServiceCenterStream.Connect。一条流上用 opcode 复用：
// 握手、注册、心跳、发现、订阅、Watch 配置。管理面走网关 HTTP，进程内调
// contract.Admin。
//
// 运行时按「中心实例名 + 环境」隔离。每个中心实例有自己的 Cache、SessionHub、
// Evictor、流服务器，不共用全局注册表。进程内分层：
//
//	SDK ──gRPC Connect──► access/stream.Handler
//	                         │
//	                         ├─ naming.App   注册发现订阅（读本机 Cache，叠 Redis 活视图）
//	                         ├─ config.App   配置发布 Watch（读库，不缓存正文）
//	                         ├─ Cache        本进程服务/节点内存表（L1）
//	                         ├─ Live         集群节点活视图（默认 Redis：心跳 / Owner / 健康）
//	                         ├─ Store        服务目录与节点审计 + 已发布配置
//	                         ├─ ClusterEvent 注册发现配置变更通知（HUB_CLUSTER_EVENT）
//	                         └─ SessionHub   数据面会话
//
// # 中文：服务如何注册
//
// 注册分两层，不要和「中心实例」卡片搞混：
//
//   - 服务名（Service）：逻辑名称，如 order。RegisterService 写入 Cache；
//     非临时定义会写 HUB_SERVICE。
//   - 业务节点（Node）：某个 ip:port。RegisterService 可附带 Node，
//     也可随后 RegisterNode。节点打上当前连接 ID，供断连清理。
//
// 临时和持久节点都写 HUB_SERVICE_NODE。心跳改本机 Cache、续 Redis 活视图 TTL，并节流回写 LastBeatTime，
// 不发集群事件。配置发布走 HUB_SERVICE_CONFIG_DATA，Get/Watch 直接读库。
// 每台网关都是对等主节点：发现读本机 Cache 再叠活视图；SDK 连到哪台都能看到同一份节点名单。
// 启动从库加载仍存活的节点后再叠活视图。未配 Redis 时仍只走 Cache + 库 + 集群事件。
// 注册/注销/更新/驱逐/配置发布写入 HUB_CLUSTER_EVENT，其它网关 ApplyRemote
// 只改本机 Cache（配置再读库推 Watch）。接到注册的网关进程是该节点当前 Owner，
// 负责断连清理与超时驱逐（删库并广播）。任意网关都能认领；空 Owner 或副本只按
// 库心跳清本机 Cache，不广播 Evict。空 Owner 的集群更新不得回退已经更新的认领。
// REGISTER_NODE 隐含服务只在缓存，不落 HUB_SERVICE；最后一临时节点离开后
// 从缓存删除，不向 SDK 推 SERVICE_DELETED。
// 同进程网关转发（discoveryType=INTERNAL）走 inproc。库里 LastBeatTime 仍
// 在心跳宽限内的持久节点加载后保持 HEALTHY，供任意网关立刻转发。
// 中心实例 Stop 会先下发 SERVER_CLOSE（reason=server_shutdown），SDK 收到后
// 立刻走断连重连，不必等 gRPC 断流。A 调 B 仍是直连 ip:port，中心停不会
// 掐断业务 TCP。网关停机的 OnClientLost 只交还 Owner、不驱逐节点，发现表保持到
// SDK 重注册或心跳超时。SDK 应用断流才立刻摘掉临时节点。旧连接若发现节点已绑到
// 新 ConnectionID 则不再处理，避免重连竞态。
//
// # 中文：A / B 双节点集群（最常见问法）
//
// 先把角色分开。服务中心不是反向代理，不替 A 转发业务请求。
// A 调 B 始终是 A 进程直连 B 的业务 ip:port。中心只告诉 A「B 现在有哪些节点」。
//
//	A1、A2  各自向中心注册服务名 A + 自己的节点（两条 gRPC 会话）
//	B1、B2  各自向中心注册服务名 B + 自己的节点
//	A1、A2  再各自 subscribeService(B)（订阅是连在中心的那条流上的，不是「A 服务」整体代订）
//
// 订阅当下：中心把 B 的当前节点当快照推给 A（B1、B2 各一条 NODE_REGISTERED）。
// 事件里 allNodes / Service.Nodes 是「此刻完整名单」，changedNode 是「这一条变更」。
// Java SDK 不会在内部替你维护一份权威地址表，只把事件交给 listener。
// A 要做负载均衡，必须在 listener 里用 allNodes 覆盖本地名单，或收到 NODE_REMOVED
// 后删掉 changedNode。
//
// 然后 B1 下线（进程停、流断、心跳超时，或显式注销），B2 还活着：
//
//	中心 Cache：临时节点 B1 删掉，B 还在，节点只剩 B2。
//	            持久节点 B1 还在表里但标 Unhealthy，健康发现只返回 B2。
//	推给 A：    NODE_REMOVED（DEREGISTERED / EVICTED）或 NODE_UPDATED（变不健康）。
//	            allNodes = [B2]，changedNode = B1。
//	A 调 B：    应按 allNodes 只打 B2。B 服务没挂，只少了一个副本。
//
// 「还能不能访问到已下线的 B1？」——中心已经不会再把它当健康节点给你。
// 若 A 把第一次快照里的 B1 地址缓存着、又没处理后续事件，A 仍可能自己去连 B1，
// 那是业务侧过期名单，连不上是 B1 进程没了，不是中心还在路由到 B1。
// 正确消费：listener 里直接换成 event.getAllNodes()（可再过滤健康），不要只追加不删除。
//
// A1 挂了不影响 A2：每条连接自己的临时节点和订阅。B 订 A 时同理，只会少 A1。
//
// 临时 vs 持久：SDK 默认临时。临时下线 = 从注册表抹掉；持久下线 = 留着但 Unhealthy，
// 健康发现 / 按健康过滤的 allNodes 都不应再打它。
//
// 取消订阅按服务名（CLIENT_UNSUBSCRIBE_SERVICES），不断其它订阅。
// 断连 OnClientLost 清掉该连接的订阅。SDK 离开才删临时节点；网关停机只交还 Owner。
// 重连后 SDK 按剩余订阅 redo。
//
// # 中文：配置 Watch
//
// 与命名不对称：配置正文只在库里。Watch 当下从库读已发布快照推给客户端，草稿不推。
// 控制台添加 / 点发布 / 删除写库后推；集群其它网关收到通知后自己读库，再推给连在那台上的 SDK。
// 保存草稿不推 SDK。按 dataId 取消走 CLIENT_UNWATCH_CONFIGS；清空全部 Watch 走短名 CLIENT_UNWATCH_CONFIG。
//
// # 中文：网关集群怎么部署
//
// 每台网关进程都对外提供同一套中心实例（同 InstanceName + Environment + 监听约定）。
// SDK 用 serverAddress=gw1:port,gw2:port,gw3:port 连其中一台；注册写在接到流的那台，
// 节点活视图进 Redis：中心 Hash 绑命名空间，命名空间 Hash 索引服务，服务 Hash 挂 nodeId，scv3:n 存正文+TTL。
// 发现任意一台 GetService 都按服务叠活视图（服务 Hash + 节点 GET），不以本机 Cache 代替集群真相。
// 订阅先靠本机快照 + HUB_CLUSTER_EVENT，副本每 2s 沿目录对账补推。心跳只续 TTL，不改 Status。
// SDK 重连 restoreState 发心跳（带原 nodeId 与快照），不再整包 REGISTER_NODE。
//
// 部署清单：
//   - 各节点同一套 HUB_SERVICE* MySQL，同一默认缓存（cache.default 指向 Redis，
//     GetDefaultCache().GetCacheType()=="redis" 且 Ping 通才挂活视图）
//   - LocalGatewayID 在集群初始化时写成 svc.GetNodeId()，进程级，不要多进程共用
//   - 中心实例卡片在管理面 Start 后经 SERVICE_CENTER_INSTANCE 在各节点拉起同一
//     InstanceName+Environment；SDK 连的是卡片上的数据面端口，不是网关 Web 口
//   - 未配 Redis 或 Ping 失败时退回 Cache+库+事件，及时一致性变差，单机仍可用
//
// 网关停机：先 SERVER_CLOSE(server_shutdown)，OnClientLost 只交还 Owner、续 TTL，
// 节点留在发现表。SDK 会换地址重连并带原 nodeId 续上。SDK 进程退出或断流才摘临时节点。
//
// # 中文：SDK 应用怎么用（一条连接、多个服务名）
//
// 一个 ServiceCenterClient = 一条 Connect 流。上面可以 registerNode 很多个不同服务名。
// 心跳按 nodeId 各打各的；Ping 和心跳共用 heartbeatInterval，但 Ping 只证明流还在。
// checkHealth：先 Ping，若本地有登记节点再打一次心跳。
//
//	registerService / registerNode   空 nodeId 由服务端发号，记在 RegisterNodeResult
//	unregisterNode(nodeId)           只摘这一个，不断流
//	unregisterService(ns,g,name,null) 只摘本客户端该服务名下的节点，其它服务名不动、不断流
//	close()                          注销本客户端全部节点并关流
//
// 不要用 unregisterService 当「下整个集群服务」。集群里一台下线用 unregisterNode 或 close。
// 自带 nodeId 须符合 [A-Za-z0-9._:/-]{1,64}，且不能占别人还活着的不同 ip:port/服务。
//
// 订阅：listener 用 event.getAllNodes() 覆盖本地名单，不要只追加。Java SDK 不维护权威地址表。
// 重连后 SDK 按剩余节点打心跳续上，再 redo 订阅和 Watch。
//
// # 中文：管理端能做什么、缺什么
//
//	hub0040  中心实例卡片：创建/改配置、Start Stop Reload、令牌、Overview
//	         （节点数/连接数/Owner/副本）。Start 经集群事件在各网关拉起同一实例。
//	hub0041  命名空间用量（叠运行时）、数据面会话。连接来自本进程 SessionHub，
//	         不是全集群连接总表；连在别的网关上的 SDK 这里看不到。
//	hub0042  服务目录仍走 HUB_SERVICE DAO。名单/详情叠 v3 ListNodes（含活视图，
//	         含 ownerGatewayId）。EditNode / OfflineNode 优先 Naming.UpdateNode，
//	         失败再退旧缓存。AddService / EditService / DeleteService 仍只写库表
//	         和旧缓存，没有接 v3 RegisterService / DeregisterService。
//	hub0043  配置草稿、发布、历史、回滚（写库后推 Watch）
//
// 管理面 CallContext.ConnectionID 为空。若接上 Naming.DeregisterService，空连接
// 会清该服务全部节点，那是「下整个服务」而不是「某一台 SDK 离开」。现在控制台
// 删服务只动目录，运行时节点可能还在；改权重/下线用 EditNode / OfflineNode。
// 没有「按 nodeId 踢掉一个实例」：摘实例应让 SDK unregisterNode / close，或等心跳超时。
package servicecenterv3
