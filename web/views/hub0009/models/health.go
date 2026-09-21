package models

// HealthStoreIndex 系统日志集合上一条期望索引的对照结果。
type HealthStoreIndex struct {
	Name    string `json:"name"`
	Keys    string `json:"keys"`
	Present bool   `json:"present"`
}

// HealthStoreObject 一张表或一个集合的存在性与索引对照。
type HealthStoreObject struct {
	Name    string             `json:"name"`
	Kind    string             `json:"kind"`
	Exists  bool               `json:"exists"`
	Indexes []HealthStoreIndex `json:"indexes,omitempty"`
}

// HealthItem 单个依赖的探活结果。不回传连接串、主机或口令。
type HealthItem struct {
	Name      string              `json:"name"`
	Kind      string              `json:"kind"`
	Driver    string              `json:"driver"`
	Status    string              `json:"status"`
	LatencyMs int64               `json:"latencyMs"`
	Message   string              `json:"message"`
	Enabled   bool                `json:"enabled"`
	Database  string              `json:"database"`
	Objects   []HealthStoreObject `json:"objects,omitempty"`
}

// HealthResponse 环境设置「系统健康」页一次探活的汇总。
// status 为 ok 或 degraded：已配置组件全部可达为 ok，任一失败为 degraded。
type HealthResponse struct {
	Status    string       `json:"status"`
	CheckedAt int64        `json:"checkedAt"`
	Items     []HealthItem `json:"items"`
}
