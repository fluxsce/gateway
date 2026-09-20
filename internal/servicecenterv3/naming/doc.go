// Package naming 实现 contract.Naming：服务注册、发现、心跳、订阅与过期剔除。
// 集群分层与读写路径图见 ../ARCHITECTURE.md。
//
// 发现必须叠活视图：GetService 读该服务 Hash；对账沿中心→命名空间→服务走目录。对账失败才退回 Cache。
// L1 服务键与 Redis 目录一样用冒号拼接（namespace:group:name），不是一份 Redis 副本。
// 心跳改 Cache、只续 TTL；约 10s 整包回写 JSON。灌视图不得用更旧 LastBeatTime 回退；Status 仍取活视图。
// SDK 重连走心跳带快照：活视图还在则 Expire（认领/换连接才整包 Put）；键没了才按快照 RegisterNode。
// 3s 新鲜窗口只给 Evictor 免重复全量拉，不用于对外发现。
// 注册/注销/更新/驱逐写活视图，并经 NamingHook 写入 HUB_CLUSTER_EVENT；其它网关 ApplyRemote 只改 Cache。
// 启动从库灌入仍存活的节点后再叠活视图。未配 Redis 时行为与只走 Cache + 库相同。
// OnClientLost 区分 SDK 断流（驱逐临时节点并删活视图）与网关停机（交还 Owner，节点保留、TTL 续期）。
// 当前 Owner 才 removeNode 广播；副本 / 空 Owner 只 dropLocalNode。空 Owner 事件不回退新认领。
// 数据面 DeregisterService 只摘本连接节点；管理面 ConnectionID 为空才清整服务。
// 活视图对账补推本机订阅；删除重试失败则注销失败。客户端自带 nodeId 须合法且不得占活着的别人。
package naming
