package model

import "time"

// ConfigDraft 是未发布内容，存在 HUB_SERVICE_CONFIG_HISTORY（changeType=DRAFT）。
// 任意网关可读可发；不推给 SDK Watch。
type ConfigDraft struct {
	TenantID           string
	CenterInstanceName string
	NamespaceID        string
	GroupName          string
	DataID             string
	Content            string
	ContentType        string // 如 text/plain、application/json
	Description        string
}

// ConfigRelease 是一次已发布快照。Version 单调递增，Rollback 会生成新版本。
type ConfigRelease struct {
	TenantID           string
	CenterInstanceName string
	NamespaceID        string
	GroupName          string
	DataID             string
	Content            string
	ContentType        string
	Description        string
	Version            int64
	MD5                string    // 内容摘要，供客户端判断是否变化
	Reason             string    // 发布或回滚原因
	PublishedBy        string    // 操作者
	PublishedAt        time.Time // 发布时间
}

// ConfigDiff 是两个已发布版本的内容对比。
type ConfigDiff struct {
	DataID      string
	FromVersion int64
	ToVersion   int64
	FromContent string
	ToContent   string
}
