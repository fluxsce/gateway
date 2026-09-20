// Package center 装配并隔离多个 CenterInstance 运行时。
// 进程内分层图见 ../ARCHITECTURE.md。
//
// Pool 是进程内实例池，实现 contract.Admin。每个 Instance 持有独立的
// cache、naming、config、evictor、session 与 stream Server。
// 进程内键为 InstanceName + Environment，避免多套环境同名互相覆盖。
// Reload 在监听/TLS/鉴权/缓冲变化时 Stop+Start；仅告警开关变化不打断会话。
// 禁止把 Cache 提升为包级全局变量。
package center
