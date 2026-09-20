// Package inproc 提供网关进程内调用 Naming 的适配器，不经过 gRPC 流。
//
// 用于同进程代理选路等场景。调用方仍须传入租户与中心实例名；本包不绕过
// 命名空间校验。外部 SDK 不得依赖本包。
package inproc
