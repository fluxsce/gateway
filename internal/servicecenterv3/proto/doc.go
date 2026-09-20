// Package proto 是 Service Center 3.0 数据面线协议。
//
// 唯一 RPC：ServiceCenterStream.Connect。Java / 其他语言 SDK v3
// 必须按下列契约实现，不要再兼容 Unary 或「保存即发布」。
//
// # 会话
//
//  1. 建连后第一条业务消息必须是 CLIENT_HANDSHAKE，否则回 PROTOCOL_HANDSHAKE_REQUIRED。
//  2. protocolVersion 填 "v3"；服务端回 "v3" 以及 version / apiLevel。
//  3. TenantID 只出现在 ServerHandshake，来自鉴权；请求里的操作人字段无效。
//  4. requestId 用于请求-响应配对；SERVER_SERVICE_CHANGE / SERVER_CONFIG_CHANGE 的 requestId 为空。
//  5. 订阅与 Watch 成功必须回 SERVER_ACK，失败回 SERVER_ERROR。业务失败也只走 SERVER_ERROR。
//  6. SERVER_CLOSE 表示中心实例正在关闭；SDK 应结束本流并按重连策略换地址或重连。
//
// # 命名
//
//	Node 即服务下的实例，必须带 serviceName，不存在孤立节点。
//	CLIENT_REGISTER_SERVICE 写服务定义，可选附带一个 Node。
//	CLIENT_REGISTER_NODE 只往该服务下加/更新节点；服务不存在时服务端先补一条服务定义。
//	注销服务不得夹带实例 ID（ServiceKey.nodeId 已 reserved）；注销节点走 NodeKey。
//	临时实例：Node.ephemeral 未设置视为临时；断开连接后剔除。
//	推送事件使用 NamingEventType 的 NODE_* / SERVICE_*，时间为 Unix 毫秒。
//
// # 配置
//
//	CLIENT_SAVE_CONFIG 只写草稿。CLIENT_PUBLISH_CONFIG 才落库并推送。
//	CLIENT_GET_CONFIG / WATCH 只看见已发布版本。
//	ConfigData 不含审计字段；回滚用 configVersion。
//
// 重新生成：
//
//	protoc --go_out=. --go_opt=paths=source_relative \
//	       --go-grpc_out=. --go-grpc_opt=paths=source_relative scv3_*.proto
//
// 文件名与 protobuf package 带 servicecenterv3 前缀，避免与历史 stream/registry/config
// 全局描述符重名。数据面仍按 /stream.ServiceCenterStream/Connect 注册，见 access/stream.Server。
package proto
