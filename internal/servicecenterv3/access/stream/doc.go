// Package stream 是数据面唯一入口：在 CenterInstance 监听端口上提供
// ServiceCenterStream.Connect 双向流。
//
// 会话必须先 HANDSHAKE。SAVE_CONFIG 只写草稿，发布走 PUBLISH_CONFIG。
// 订阅与 Watch 回 SERVER_ACK，并按连接绑定，断连先取消再关通道。
// TenantID 只来自鉴权。后续 SDK v3 按 proto 包文档实现即可。
package stream
