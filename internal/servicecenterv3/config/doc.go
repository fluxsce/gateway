// Package config 实现 contract.Config：草稿、发布、已发布读、历史与 Watch。
//
// 所有读写都打数据库，网关不缓存配置正文。SaveDraft 只写草稿行，不推 SDK。
// Publish / 回滚 / 删除写 HUB_SERVICE_CONFIG_DATA 后推本机 Watch，再经 ConfigHook
// 通知其它网关自己读库推给它们的 SDK。配置是文件、内容大、变更不频繁。
package config
