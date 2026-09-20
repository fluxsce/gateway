package store

import "gateway/pkg/database"

// Store 聚合本中心用到的持久化访问，不是稳定面。
type Store struct {
	db database.Database
	// Center 中心实例定义。
	Center *CenterStore
	// Namespace 命名空间。
	Namespace *NamespaceStore
	// Service 服务定义。
	Service *ServiceStore
	// Node 服务节点（临时和持久都在 HUB_SERVICE_NODE）。
	Node *NodeStore
	// Config 已发布配置与历史。
	Config *ConfigStore
	// Auth 用户与令牌校验。
	Auth *AuthStore
}

// New 按同一 Database 构造各子存储。
func New(db database.Database) *Store {
	s := &Store{db: db}
	s.Center = &CenterStore{db: db}
	s.Namespace = &NamespaceStore{db: db}
	s.Service = &ServiceStore{db: db}
	s.Node = &NodeStore{db: db}
	s.Config = &ConfigStore{db: db}
	s.Auth = &AuthStore{db: db}
	return s
}
