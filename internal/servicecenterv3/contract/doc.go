// Package contract 定义 Service Center 3.0 的进程内稳定面。
//
// 三个接口是调用方应依赖的全部能力：
//
//   - Admin：中心实例生命周期与只读概览（管理面）。
//   - Naming：服务注册、发现、心跳与订阅（数据面命名）。
//   - Config：草稿、发布、已发布读、历史、回滚与 Watch（数据面配置）。
//
// 观察能力不单独成包：连接列表与 Overview 在 Admin，推送在 Naming/Config
// 的订阅方法上。实现可以换，接口与哨兵错误不得静默改语义。
//
// 每次调用必须带 CallContext。TenantID 只允许鉴权写入；CenterInstanceName
// 选择隔离运行时；NamespaceID 是命名/配置的工作空间。
//
// 错误用 errors.Is 与本包哨兵比较，不要匹配中文/英文文案。
package contract
