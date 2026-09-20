// Package cache 提供按 CenterInstanceName 隔离的内存索引。
//
// 一个 Cache 只服务一个中心实例。读写并发安全。读路径返回防御性拷贝，
// 避免调用方改到内部 map。不要在包级保存 Cache 单例。
package cache
